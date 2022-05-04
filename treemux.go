// Copyright 2022 Hayo van Loon. All rights reserved.
// Use of this source code is governed by an Apache
// license that can be found in the LICENSE file.

// TreeMux is an HTTP request multiplexer that routes using a tree structure.
//
// Wildcards ("*") are used to indicate flexible path elements in a resource
// URL, which can then be mapped to a single Handler (function).
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
	HandleFunc(path string, handler func(http.ResponseWriter, *http.Request))

	Handler(r *http.Request) (h http.Handler, pattern string)
}

type treeMux struct {
	trie     WildcardTrie
	notFound http.HandlerFunc
}

func (t *treeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, _ := t.Handler(r)
	h.ServeHTTP(w, r)
}

func (t *treeMux) Handle(path string, handler http.Handler) {
	t.trie.Add(path, handler)
}

func (t *treeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	t.Handle(pattern, http.HandlerFunc(handler))
}

func (t treeMux) Handler(r *http.Request) (http.Handler, string) {
	v, pattern := t.trie.Get(r.URL.Path)
	if pattern == "" {
		return t.notFound, pattern
	}
	return v.(http.Handler), r.URL.Path
}

// NewTreeMux creates a new tree-based request multiplexer. If a request path
// cannot be matched, the standard `http.NotFound` will be used unless
// OptionNotFound specifies a different one.
func NewTreeMux(options ...Option) TreeMux {
	t := &treeMux{
		trie:     newWildcardTrie("/"),
		notFound: http.NotFound,
	}
	for _, o := range options {
		o.Apply(t)
	}
	return t
}

type Option interface {
	Apply(mux *treeMux)
	private()
}

type optionNotFound struct {
	value http.HandlerFunc
}

func (o optionNotFound) Apply(mux *treeMux) {
	mux.notFound = o.value
}

func (o optionNotFound) private() {}

func OptionNotFound(handler http.HandlerFunc) Option {
	return optionNotFound{handler}
}
