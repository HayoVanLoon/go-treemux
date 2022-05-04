// Copyright 2022 Hayo van Loon. All rights reserved.
// Use of this source code is governed by an Apache
// license that can be found in the LICENSE file.

package treemux

import (
	"fmt"
	"reflect"
	"strings"
)

type WildcardTrie interface {
	Get(s string) (interface{}, string)
	Add(s string, v interface{})
}

type wildcardTrie struct {
	separator string
	key       string
	pattern   string
	value     interface{}
	children  []wildcardTrie
}

func newWildcardTrie(separator string) WildcardTrie {
	return &wildcardTrie{separator: separator, key: ""}
}

// Add breaks up a string using the specified separator and adds the data to the
// trie. When a path already exists in the trie, the old data is overwritten.
//
// The first element is always expected to be empty. Therefore the following
// statements are idempotent.
//   trie.Add("/foo/bar", "/", 1)
//   trie.Add("foo/bar", "/", 1)
//
// The wildcard is a flexible, retrieval-time parameter. It plays no role
// whatsoever at construction-time. One could even apply different wildcard
// schemes for different purposes on the same trie.
// See Get for more details on wildcard behaviour.
func (t *wildcardTrie) Add(s string, v interface{}) {
	xs := strings.Split(s, t.separator)
	idx := 0
	if xs[0] == "" {
		// skip empty root
		idx = 1
	}
	if len(xs) > 1 && xs[len(xs)-1] == "" {
		panic("path cannot end with slash")
	}
	t.grow(idx, xs, v)
}

func (t *wildcardTrie) grow(idx int, xs []string, v interface{}) {
	if len(xs) == idx {
		t.value = v
		return
	}
	for i := range t.children {
		if t.children[i].key == xs[idx] {
			t.children[i].grow(idx+1, xs, v)
			return
		}
	}
	if len(xs) > idx {
		c := newTrie(t.separator, xs[idx], xs[:idx+1])
		if len(xs) == idx+1 {
			c.value = v
		} else {
			c.grow(idx+1, xs, v)
		}
		t.children = append(t.children, c)
	}
}

func newTrie(sep, key string, path []string) wildcardTrie {
	return wildcardTrie{separator: sep, key: key, pattern: "/" + strings.Join(path, sep)}
}

const wildcard = "*"

// Get attempts to retrieve the data from the specified path, split up by the
// specified separator using the default wildcard "*".
//
// Wildcard elements hold no special status over other elements. When, due to a
// wildcard, a path has two valid end points, the one inserted earliest wins.
func (t *wildcardTrie) Get(s string) (interface{}, string) {
	// TODO(hvl): input validation
	xs := strings.Split(s, t.separator)
	if xs[0] == "" {
		return t.get(0, xs, wildcard)
	}
	for _, c := range t.children {
		if v, pattern := c.get(0, xs, wildcard); pattern != "" {
			return v, pattern
		}
	}
	return nil, ""
}

func (t *wildcardTrie) get(idx int, xs []string, wildcard string) (interface{}, string) {
	if xs[idx] != t.key && t.key != wildcard {
		if t.key == "" && len(t.children) == 0 {
			return t.value, t.pattern
		}
		return nil, ""
	}
	if len(xs)-idx == 1 {
		return t.value, t.pattern
	}
	for _, c := range t.children {
		if v, pattern := c.get(idx+1, xs, wildcard); pattern != "" {
			return v, pattern
		}
	}
	return nil, ""
}

func (t *wildcardTrie) equals(other wildcardTrie) bool {
	if t.separator != other.separator {
		return false
	}
	if t.key != other.key {
		return false
	}
	if t.pattern != other.pattern {
		return false
	}
	if !reflect.DeepEqual(t.value, other.value) {
		return false
	}
	if len(t.children) != len(other.children) {
		return false
	}
	for i, c := range t.children {
		if !c.equals(other.children[i]) {
			return false
		}
	}
	return true
}

func (t wildcardTrie) String() string {
	b := &strings.Builder{}
	b.WriteString("WildcardTrie(")
	b.WriteString(t.separator)
	b.WriteRune(')')
	t.string(b)
	return b.String()
}

func (t *wildcardTrie) string(b *strings.Builder) {
	b.WriteString("{\"")
	b.WriteString(t.pattern)
	b.WriteString(fmt.Sprintf("\"=%v", t.value))
	if len(t.children) > 0 {
		b.WriteString(",[")
		t.children[0].string(b)
		for i := 1; i < len(t.children); i += 1 {
			b.WriteRune(',')
			t.children[i].string(b)
		}
		b.WriteRune(']')
	}
	b.WriteRune('}')
}
