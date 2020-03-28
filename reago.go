// Copyright © 2019 Patrick Lawrence <patrick.lawrence@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package reago

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/time/rate"
)

const (
	libraryVersion            = "1.0"
	apiVersion                = "v1"
	defaultBaseURL            = "https://api.emailsrvr.com/"
	userAgent                 = "reago/" + libraryVersion
	mediaType                 = "application/json"
	defaultPageSize           = 50
	defaultGetLimit           = 1.9
	defaultGetBurst           = 1
	defaultPutPostDeleteLimit = 1.4
	defaultPutPostDeleteBurst = 1
)

// Client manages communication with Rackspace Email v1 API
type Client struct {
	// HTTP client used to communicate with the Rackspace Email API.
	client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	// Auth
	userKey   string
	secretKey string

	RackspaceEmailAliases RackspaceEmailAliasesService
	Domains               DomainsService

	debugHTTP bool

	getLimiter           *rate.Limiter
	putPostDeleteLimiter *rate.Limiter
}

// PageOptions specifies the request pagination options
type PageOptions struct {
	Offset int `url:"offset,omitempty"`
	Size   int `url:"size,omitempty"`
}

// NewClient returns a Rackspace Email API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.RackspaceEmailAliases = &RackspaceEmailAliasesServiceOp{client: c}
	c.Domains = &DomainsServiceOp{client: c}

	c.getLimiter = rate.NewLimiter(rate.Limit(defaultGetLimit), defaultGetBurst)
	c.putPostDeleteLimiter = rate.NewLimiter(rate.Limit(defaultPutPostDeleteLimit), defaultPutPostDeleteBurst)

	return c
}

// Response is a Rackspace Email API response. This wraps the standard
// http.Response returned from Rackspace.
type Response struct {
	*http.Response
}

// ErrorResponse returns the information from an API error
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// RequestID returned from the API, useful to contact support.
	RequestID string `json:"request_id"`
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

// New returns a new API client instance.
func New(httpClient *http.Client, options ...func(*Client) error) (*Client, error) {
	c := NewClient(httpClient)
	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// SetBaseURL is a client option for setting the base URL.
func SetBaseURL(bu string) func(*Client) error {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}

		c.BaseURL = u
		return nil
	}
}

// SetUserAgent is a client option for setting the user agent.
func SetUserAgent(ua string) func(*Client) error {
	return func(c *Client) error {
		c.UserAgent = fmt.Sprintf("%s", ua)
		return nil
	}
}

// SetUserKey is a client option for setting the user key.
func SetUserKey(uk string) func(*Client) error {
	return func(c *Client) error {
		c.userKey = uk
		return nil
	}
}

// SetSecretKey is a client option for setting the secret key.
func SetSecretKey(sk string) func(*Client) error {
	return func(c *Client) error {
		c.secretKey = sk
		return nil
	}
}

// SetDebugHTTP is a client option for setting debugging for HTTP calls.
func SetDebugHTTP() func(*Client) error {
	return func(c *Client) error {
		c.debugHTTP = true
		return nil
	}
}

// SetGetLimiter is a client option for setting the ratelimiter for GET
// requests. rps is the requests per second and burst is the number of
// burst requests allowed.
func SetGetLimiter(rps float64, burst int) func(*Client) error {
	return func(c *Client) error {
		c.getLimiter = rate.NewLimiter(rate.Limit(rps), burst)
		return nil
	}
}

// SetPostLimiter is a client option for setting the ratelimiter for POST
// requests. rps is the requests per second and burst is the number of
// burst requests allowed.
func SetPostLimiter(rps float64, burst int) func(*Client) error {
	return func(c *Client) error {
		c.putPostDeleteLimiter = rate.NewLimiter(rate.Limit(rps), burst)
		return nil
	}
}

// NewRequest creates an API request. A relative URL can be provided in
// urlStr, which will be resolved to the BaseURL of the Client. Relative URLs
// should always be specified without a preceding slash. If specified, the
// map body is rendered as application/x-www-form-urlencoded.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body map[string]string) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	data := url.Values{}
	if body != nil {
		for k, v := range body {
			data.Add(k, v)
		}
	}

	req, err := http.NewRequest(method, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Add("Content-Type", mediaType)
	}
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)

	c.sign(req)

	return req, nil
}

func (c *Client) sign(req *http.Request) {
	ua := req.Header.Get("User-Agent")
	ts := time.Now().Format("20060102150405")

	hasher := sha1.New()
	io.WriteString(hasher, fmt.Sprintf("%s%s%s%s", c.userKey, ua, ts, c.secretKey))

	b64 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	sig := fmt.Sprintf("%s:%s:%s", c.userKey, ts, b64)

	req.Header.Add("X-Api-Signature", sig)
}

func newResponse(r *http.Response) *Response {
	response := Response{Response: r}

	return &response
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	dump, err := httputil.DumpRequest(req, true)
	if c.debugHTTP {
		fmt.Fprintf(os.Stderr, "Req: %s\n", string(dump))
	}

	// Rate limiting
	switch req.Method {
	case "GET":
		if err := c.getLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	default:
		if err := c.putPostDeleteLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	resp, err := DoRequestWithClient(ctx, c.client, req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	resDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}

	if c.debugHTTP {
		fmt.Fprintf(os.Stderr, "Resp: %s\n", resDump)
	}

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, err
}

// DoRequest submits an HTTP request.
func DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return DoRequestWithClient(ctx, http.DefaultClient, req)
}

// DoRequestWithClient submits an HTTP request using the specified client.
func DoRequestWithClient(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	return errorResponse
}

// Error returns a string representation of an API error
func (r *ErrorResponse) Error() string {
	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}
