// Copyright 2022 Hayo van Loon. All rights reserved.
// Use of this source code is governed by an Apache
// license that can be found in the LICENSE file.

// TreeMux is an HTTP request multiplexer that routes using a tree structure.
//
// Wildcards ("*") are used to indicate flexible path elements in a resource
// URL, which can then be mapped to a single Handler function.
//
// Example:
// With the following route:
//   t.Handle("/countries/*/cities", handleCities)
// Paths like these will be handled by `handleCities`:
//   "/countries/belgium/cities"
//   "/countries/france/cities"
//
// There is no support for elements with partial wildcards (i.e. `/foo*/bar`).

package treemux

import "net/http"

type TreeMux interface {
	http.Handler

	// Handle adds a new http.Handler for the given path. When a path already
	// exists in the tree, the old data is overwritten.
	//
	// The root element is always empty, so the following statements will have
	// the same result.
	//   t.Handle("/foo/bar", fn)
	//   t.Handle("foo/bar", fn)
	Handle(path string, handler http.Handler)

	// HandleFunc adds a new http.HandlerFunc for the given path. See Handle for
	// more details.
	HandleFunc(path string, handler http.HandlerFunc)
}

type treeMux struct {
	trie     WildcardTrie
	notFound http.HandlerFunc
}

func (t treeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v, found := t.trie.Get(r.URL.Path)
	if !found {
		t.notFound(w, r)
		return
	}
	switch h := v.(type) {
	case http.Handler:
		h.ServeHTTP(w, r)
	case http.HandlerFunc:
		h(w, r)
	}
}

func (t *treeMux) Handle(path string, handler http.Handler) {
	t.trie.Add(path, handler)
}

func (t *treeMux) HandleFunc(path string, handler http.HandlerFunc) {
	t.trie.Add(path, handler)
}

// NewTreeMux creates a new tree-based request multiplexer. If a request path
// cannot be matched the standard `http.NotFound` will be used.
func NewTreeMux() TreeMux {
	return NewTreeMuxWithNotFound(nil)
}

// NewTreeMuxWithNotFound creates a new tree-based request multiplexer. If a
// request path cannot be matched, the specified HandlerFunc will be used. If
// set to `nil`, the default `http.NotFound` will be used.
func NewTreeMuxWithNotFound(notFound http.HandlerFunc) TreeMux {
	if notFound == nil {
		notFound = http.NotFound
	}
	return &treeMux{
		trie:     newWildcardTrie("/"),
		notFound: notFound,
	}
}
