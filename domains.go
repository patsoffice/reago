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
	"fmt"
	"net/http"
)

const domainsBasePath = "v1/domains"

// DomainsService is an interface for managing DNS with the Rackspace Email
// API.
//
// See: http://api-wiki.apps.rackspace.com/api-wiki/index.php?title=Domain_(Rest_API)
type DomainsService interface {
	Index(context.Context, *PageOptions) ([]Domain, *Response, error)
	Show(context.Context, string) (*Domain, *Response, error)
}

// DomainsServiceOp handles communication with the domain related methods of
// the Rackspace Email API.
type DomainsServiceOp struct {
	client *Client
}

var _ DomainsService = DomainsServiceOp{}
var _ DomainsService = &DomainsServiceOp{}

// Domain represents a Rackspace Email API domain
type Domain struct {
	Name                           string `json:"name"`
	AccountNumber                  string `json:"accountNumber"`
	ServiceType                    string `json:"serviceType"`
	ActiveSyncLicenses             int    `json:"activeSyncLicenses"`
	ActiveSyncMobileServiceEnabled bool   `json:"activeSyncMobileServiceEnabled"`
	ArchivingServiceEnabled        bool   `json:"archivingServiceEnabled"`
	BlackBerryLicenses             int    `json:"blackBerryLicenses"`
	BlackBerryMobileServiceEnabled bool   `json:"blackBerryMobileServiceEnabled"`
	ExchangeExtraStorage           int    `json:"exchangeExtraStorage"`
	ExchangeMaxNumMailboxes        int    `json:"exchangeMaxNumMailboxes"`
	ExchangeUsedStorage            int    `json:"exchangeUsedStorage"`
	RSEmailBaseMailboxSize         int    `json:"rsEmailBaseMailboxSize"`
	RSEmailExtraStorage            int    `json:"rsEmailExtraStorage"`
	RSEmailMaxNumberMailboxes      int    `json:"rsEmailMaxNumberMailboxes"`
	RSEmailUsedStorage             int    `json:"rsEmailUsedStorage"`
}

type domainRoot struct {
	Domain *Domain `json:"domain"`
}

type domainsRoot struct {
	Offset  int      `struct:"offset"`
	Size    int      `struct:"size"`
	Total   int      `struct:"total"`
	Domains []Domain `json:"domains"`
}

// Index lists all domains
func (s DomainsServiceOp) Index(ctx context.Context, opt *PageOptions) ([]Domain, *Response, error) {
	var domains []Domain
	var resp *Response
	var err error

	if opt == nil {
		opt = &PageOptions{Size: defaultPageSize}
	}

	for {
		path := domainsBasePath
		path, err := addOptions(path, opt)
		if err != nil {
			return nil, nil, err
		}

		req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, nil, err
		}

		root := new(domainsRoot)
		resp, err = s.client.Do(ctx, req, root)
		if err != nil {
			return nil, resp, err
		}
		domains = append(domains, root.Domains...)

		if root.Total <= root.Size+root.Offset {
			break
		}
		opt.Offset = root.Size + root.Offset
	}

	return domains, resp, err
}

// Show gets details of a domain and requires a non-empty domain name
func (s DomainsServiceOp) Show(ctx context.Context, name string) (*Domain, *Response, error) {
	if len(name) < 1 {
		return nil, nil, NewArgError("name", "cannot be an empty string")
	}

	path := fmt.Sprintf("%s/%s", domainsBasePath, name)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(domainRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Domain, resp, err
}
