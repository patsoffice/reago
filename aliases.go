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
	"strings"
)

const rackspaceEmailAliasesBasePath = "v1/domains/%s/rs/aliases"

// RackspaceEmailAliasesService is an interface for managing Rackspace Email aliases with the Rackspace Email
// API.
//
// See: http://api-wiki.apps.rackspace.com/api-wiki/index.php?title=Rackspace_Alias(Rest_API)
type RackspaceEmailAliasesService interface {
	Add(context.Context, string, string, []string) (*Response, error)
	Delete(context.Context, string, string) (*Response, error)
	Show(context.Context, string, string) (*RackspaceEmailAliasShow, *Response, error)
	Index(context.Context, *PageOptions, string) ([]RackspaceEmailAlias, *Response, error)
}

// RackspaceEmailAliasesServiceOp handles communication with the rackspace
// email alias related methods of the Rackspace Email API.
type RackspaceEmailAliasesServiceOp struct {
	client *Client
}

var _ RackspaceEmailAliasesService = &RackspaceEmailAliasesServiceOp{}

// RackspaceEmailAlias represents a Rackspace Email API alias from the Index
// method.
type RackspaceEmailAlias struct {
	Name            string `json:"name"`
	NumberOfMembers int    `json:"numberOfMembers"`
}

// EmailAddress represents an array of email addresses that iare tied to a
// Rackspace Email alias.
type EmailAddress struct {
	Addresses []string `json:"emailAddress"`
}

// RackspaceEmailAliasShow represents the response from the Show method.
type RackspaceEmailAliasShow struct {
	Name             string       `json:"name"`
	EmailAddressList EmailAddress `json:"emailAddressList"`
}

// type rackspaceEmailAliasRoot struct {
// 	RackspaceEmailAlias *RackspaceEmailAlias `json:"alias"`
// }

type rackspaceEmailAliasesRoot struct {
	Offset                int                   `struct:"offset"`
	Size                  int                   `struct:"size"`
	Total                 int                   `struct:"total"`
	RackspaceEmailAliases []RackspaceEmailAlias `json:"aliases"`
}

type rackspaceEmailAliasAddRequest struct {
	RackspaceEmailAliasEmails string `json:"aliasEmails"`
}

// Index lists all Rackspace Email aliases
func (s RackspaceEmailAliasesServiceOp) Index(ctx context.Context, opt *PageOptions, domain string) ([]RackspaceEmailAlias, *Response, error) {
	var aliases []RackspaceEmailAlias
	var resp *Response
	var err error

	if len(domain) < 1 {
		return nil, nil, NewArgError("domain", "it cannot be an empty string")
	}

	if opt == nil {
		opt = &PageOptions{Size: defaultPageSize}
	}

	for {
		path := fmt.Sprintf(rackspaceEmailAliasesBasePath, domain)
		path, err = addOptions(path, opt)
		if err != nil {
			return nil, nil, err
		}

		req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, nil, err
		}

		root := new(rackspaceEmailAliasesRoot)
		resp, err := s.client.Do(ctx, req, root)
		if err != nil {
			return nil, resp, err
		}
		aliases = append(aliases, root.RackspaceEmailAliases...)

		if root.Total <= root.Size+root.Offset {
			break
		}
		opt.Offset = root.Size + root.Offset
	}

	return aliases, resp, err
}

// Show gets details of a Rackspace Email alias and requires a non-empty domain
// name and a non-empty alias.
func (s *RackspaceEmailAliasesServiceOp) Show(ctx context.Context, domain, alias string) (*RackspaceEmailAliasShow, *Response, error) {
	if len(domain) < 1 {
		return nil, nil, NewArgError("domain", "cannot be an empty string")
	}

	if len(alias) < 1 {
		return nil, nil, NewArgError("alias", "cannot be an empty string")
	}

	path := fmt.Sprintf(rackspaceEmailAliasesBasePath, domain)
	path = fmt.Sprintf("%s/%s", path, alias)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(RackspaceEmailAliasShow)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, err
}

// Add adds a new Rackspace Email alias and requires a non-empty domain name
// and a non-empty alias and a slice of email addresses.
func (s *RackspaceEmailAliasesServiceOp) Add(ctx context.Context, domain, alias string, emailAddresses []string) (*Response, error) {
	if len(domain) < 1 {
		return nil, NewArgError("domain", "cannot be an empty string")
	}
	if len(alias) < 1 {
		return nil, NewArgError("alias", "cannot be an empty string")
	}
	if len(emailAddresses) < 1 {
		return nil, NewArgError("emailAddresses", "cannot be an empty list of strings")
	}

	body := map[string]string{"aliasEmails": strings.Join(emailAddresses, ",")}

	path := fmt.Sprintf(rackspaceEmailAliasesBasePath, domain)
	path = fmt.Sprintf("%s/%s", path, alias)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, err
}

// Delete removes a Rackspace Email alias and requires a non-empty domain name
// and a non-empty alias.
func (s *RackspaceEmailAliasesServiceOp) Delete(ctx context.Context, domain, alias string) (*Response, error) {
	if len(domain) < 1 {
		return nil, NewArgError("domain", "cannot be an empty string")
	}
	if len(alias) < 1 {
		return nil, NewArgError("alias", "cannot be an empty string")
	}

	path := fmt.Sprintf(rackspaceEmailAliasesBasePath, domain)
	path = fmt.Sprintf("%s/%s", path, alias)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)

	return resp, err
}
