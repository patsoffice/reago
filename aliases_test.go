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

func TestRackspaceEmailAliases_Index(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains/domain.com/rs/aliases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"aliases": [{"name":"foo"},{"name":"bar"}]}`)
	})

	aliases, _, err := client.RackspaceEmailAliases.Index(ctx, nil, "domain.com")
	if err != nil {
		t.Errorf("RackspaceEmailAliases.Index returned error: %v", err)
	}

	expected := []RackspaceEmailAlias{{Name: "foo"}, {Name: "bar"}}
	if !reflect.DeepEqual(aliases, expected) {
		t.Errorf("RackspaceEmailAlias.Index returned %+v, expected %+v", aliases, expected)
	}
}

func TestRackspaceEmailAliases_Index_DomainEmpty(t *testing.T) {
	_, _, err := client.RackspaceEmailAliases.Index(ctx, nil, "")
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Index should have returned an error for an empty domain")
	}
}

func TestRackspaceEmailAliases_Index_MultiplePages(t *testing.T) {
	setup()
	defer teardown()

	responses := []string{
		`{"offset": 0, "size": 1, "total": 2, "aliases": [{"name":"foo"}]}`,
		`{"offset": 1, "size": 1, "total": 2, "aliases": [{"name":"bar"}]}`,
	}
	index := 0

	mux.HandleFunc("/v1/domains/domain.com/rs/aliases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, responses[index])
		index++
	})

	aliases, _, err := client.RackspaceEmailAliases.Index(ctx, &PageOptions{Size: 1}, "domain.com")
	if err != nil {
		t.Fatal(err)
	}

	expected := []RackspaceEmailAlias{{Name: "foo"}, {Name: "bar"}}
	if !reflect.DeepEqual(aliases, expected) {
		t.Errorf("RackspaceEmailAliases.Index returned %+v, expected %+v", aliases, expected)
	}
}

func TestRackspaceEmailAliases_Show_NoDomain(t *testing.T) {
	_, _, err := client.RackspaceEmailAliases.Show(ctx, "", "foo")
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Show should have returned an error for an empty domain")
	}
}

func TestRackspaceEmailAliases_Show_NoAlias(t *testing.T) {
	_, _, err := client.RackspaceEmailAliases.Show(ctx, "domain.com", "")
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Show should have returned an error for an empty alias")
	}
}

func TestRackspaceEmailAliases_Show(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains/foo.com/rs/aliases/bar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"name": "bar", "emailAddressList": {"emailAddress": ["baz@bar.com", "qux@bar.com"]}}`)
	})

	aliases, _, err := client.RackspaceEmailAliases.Show(ctx, "foo.com", "bar")
	if err != nil {
		t.Errorf("RackspaceEmailAliases.Show returned error: %v", err)
	}

	expected := &RackspaceEmailAliasShow{
		Name: "bar",
		EmailAddressList: EmailAddress{
			Addresses: []string{
				"baz@bar.com",
				"qux@bar.com",
			},
		},
	}
	if !reflect.DeepEqual(aliases, expected) {
		t.Errorf("RackspaceEmailAliases.Show returned %+v, expected %+v", aliases, expected)
	}
}

func TestRackspaceEmailAliases_Add_NoDomain(t *testing.T) {
	_, err := client.RackspaceEmailAliases.Add(ctx, "", "foo", []string{"foo@bar.com"})
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Add should have returned an error for an empty domain")
	}
}

func TestRackspaceEmailAliases_Add_NoAlias(t *testing.T) {
	_, err := client.RackspaceEmailAliases.Add(ctx, "domain.com", "", []string{"foo@bar.com"})
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Add should have returned an error for an empty alias")
	}
}

func TestRackspaceEmailAliases_Add_NoAddresses(t *testing.T) {
	_, err := client.RackspaceEmailAliases.Add(ctx, "domain.com", "foo", nil)
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Add should have returned an error for an empty slice of addresses")
	}
}

func TestRackspaceEmailAliases_Add(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains/foo.com/rs/aliases/bar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	_, err := client.RackspaceEmailAliases.Add(ctx, "foo.com", "bar", []string{"foo@bar.com"})
	if err != nil {
		t.Errorf("RackspaceEmailAliases.Add returned error: %v", err)
	}
}

func TestRackspaceEmailAliases_Delete_NoDomain(t *testing.T) {
	_, err := client.RackspaceEmailAliases.Delete(ctx, "", "foo")
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Delete should have returned an error for an empty domain")
	}
}

func TestRackspaceEmailAliases_Delete_NoAlias(t *testing.T) {
	_, err := client.RackspaceEmailAliases.Delete(ctx, "domain.com", "")
	if err == nil {
		t.Errorf("RackspaceEmailAliases.Delete should have returned an error for an empty alias")
	}
}

func TestRackspaceEmailAliases_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/domains/foo.com/rs/aliases/bar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.RackspaceEmailAliases.Delete(ctx, "foo.com", "bar")
	if err != nil {
		t.Errorf("RackspaceEmailAliases.Delete returned error: %v", err)
	}
}
