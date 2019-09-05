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
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	mux    *http.ServeMux
	ctx    = context.TODO()
	client *Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

func testClientServices(t *testing.T, c *Client) {
	services := []string{
		"RackspaceEmailAliases",
		"Domains",
	}

	cp := reflect.ValueOf(c)
	cv := reflect.Indirect(cp)

	for _, s := range services {
		if cv.FieldByName(s).IsNil() {
			t.Errorf("c.%s shouldn't be nil", s)
		}
	}
}

func testClientDefaultBaseURL(t *testing.T, c *Client) {
	if c.BaseURL == nil || c.BaseURL.String() != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, expected %v", c.BaseURL, defaultBaseURL)
	}
}

func testClientDefaultUserAgent(t *testing.T, c *Client) {
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

func testClientSetHTTPDebug(t *testing.T, c *Client) {
	if c.debugHTTP != false {
		t.Errorf("NewClient debugHTTP = %v, expected %v", c.debugHTTP, false)
	}
}

func testClientDefaults(t *testing.T, c *Client) {
	testClientDefaultBaseURL(t, c)
	testClientDefaultUserAgent(t, c)
	testClientServices(t, c)
}

func Test_NewClient(t *testing.T) {
	c := NewClient(nil)
	testClientDefaults(t, c)
}

func Test_New(t *testing.T) {
	c, err := New(nil)

	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	testClientDefaults(t, c)
}

func Test_New_OptionSetBaseURL(t *testing.T) {
	baseURL := "https://test.com/api"
	c, err := New(nil, SetBaseURL(baseURL))

	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if c.BaseURL.String() != baseURL {
		t.Errorf("NewClient BaseURL = %v, expected %v", c.BaseURL.String(), baseURL)
	}
}

func Test_New_OptionSetUserAgent(t *testing.T) {
	userAgent := "test_ua"
	c, err := New(nil, SetUserAgent(userAgent))

	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

func Test_New_OptionSetUserKey(t *testing.T) {
	userKey := "userid"
	c, err := New(nil, SetUserKey(userKey))

	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if c.userKey != userKey {
		t.Errorf("NewClient userKey = %v, expected %v", c.userKey, userKey)
	}
}

func Test_New_OptionSetSecretKey(t *testing.T) {
	secretKey := "hunter2"
	c, err := New(nil, SetSecretKey(secretKey))

	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if c.secretKey != secretKey {
		t.Errorf("NewClient secretKey = %v, expected %v", c.secretKey, secretKey)
	}
}

func Test_New_OptionDebug(t *testing.T) {
	c, err := New(nil, SetDebugHTTP())

	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if c.debugHTTP != true {
		t.Errorf("NewClient debugHTTP = %v, expected %v", c.debugHTTP, true)
	}
}
