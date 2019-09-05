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
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestDomains_Index(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domains": [{"name":"foo.com"},{"name":"bar.com"}]}`)
	})

	domains, _, err := client.Domains.Index(ctx, nil)
	if err != nil {
		t.Errorf("Domains.Index returned error: %v", err)
	}

	expected := []Domain{{Name: "foo.com"}, {Name: "bar.com"}}
	if !reflect.DeepEqual(domains, expected) {
		t.Errorf("Domains.Index returned %+v, expected %+v", domains, expected)
	}
}

func TestDomains_Index_MultiplePages(t *testing.T) {
	setup()
	defer teardown()

	responses := []string{
		`{"offset": 0, "size": 1, "total": 2, "domains": [{"name":"foo.com"}]}`,
		`{"offset": 1, "size": 1, "total": 2, "domains": [{"name":"bar.com"}]}`,
	}
	index := 0

	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, responses[index])
		index++
	})

	domains, _, err := client.Domains.Index(ctx, &PageOptions{Size: 1})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Domain{{Name: "foo.com"}, {Name: "bar.com"}}
	if !reflect.DeepEqual(domains, expected) {
		t.Errorf("Domains.Index returned %+v, expected %+v", domains, expected)
	}
}

func TestDomains_Show_NoName(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Domains.Show(ctx, "")
	if err == nil {
		t.Errorf("Domains.Show should have returned an error for an empty domain")
	}
}

func TestDomains_Show(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains/foo.com", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domain": {"name":"foo.com", "accountNumber":"1234", "serviceType":"rsemail"}}`)
	})

	domains, _, err := client.Domains.Show(ctx, "foo.com")
	if err != nil {
		t.Errorf("Domains.Show returned error: %v", err)
	}

	expected := &Domain{
		Name:          "foo.com",
		AccountNumber: "1234",
		ServiceType:   "rsemail",
	}
	if !reflect.DeepEqual(domains, expected) {
		t.Errorf("Domains.Show returned %+v, expected %+v", domains, expected)
	}
}
