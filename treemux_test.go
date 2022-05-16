// Copyright 2022 Hayo van Loon. All rights reserved.
// Use of this source code is governed by an Apache
// license that can be found in the LICENSE file.

package treemux

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHandler struct {
}

func (testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Handler " + r.URL.Path + "!"))
}

func TestTreeMux_Handle(t *testing.T) {
	handleFunc := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("HandleFunc!"))
	}
	notFound := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not!found!"))
	}

	cases := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{"double elements", "/foo/bar", 200, "HandleFunc!"},
		{"not found", "/foo/meow", 404, "not!found!"},
		{"simple", "/moo", 200, "HandleFunc!"},
		{"handler", "/moo/meh", 200, "Handler /moo/meh!"},
		{"handler with minimal path element", "/moo/", 200, "Handler /moo/!"},
		{"wildcard has no meaning in Get", "/foo/*", 404, "not!found!"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tr := NewTreeMux(OptionNotFound(notFound), OptionDebug())
			tr.HandleFunc("foo/bar", handleFunc)
			tr.HandleFunc("/moo", handleFunc)
			tr.Handle("/moo/*", testHandler{})

			s := httptest.NewServer(tr)
			defer s.Close()

			resp, err := s.Client().Get(s.URL + c.path)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if resp.StatusCode != c.wantCode {
				t.Errorf("expected %v, got %v", c.wantCode, resp.StatusCode)
			}
			bs, _ := ioutil.ReadAll(resp.Body)
			if string(bs) != c.wantBody {
				t.Errorf("expected %s, got %v", c.wantBody, string(bs))
			}
		})
	}
}
