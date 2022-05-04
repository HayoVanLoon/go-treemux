// Copyright 2022 Hayo van Loon. All rights reserved.
// Use of this source code is governed by an Apache
// license that can be found in the LICENSE file.

package treemux

import (
	"fmt"
	"testing"
)

func TestWildcardTrie_Equals(t *testing.T) {
	cases := []struct {
		name  string
		left  wildcardTrie
		right wildcardTrie
		want  bool
	}{
		{
			"empty",
			wildcardTrie{},
			wildcardTrie{},
			true,
		},
		{
			"simple equals no value",
			wildcardTrie{key: "foo"},
			wildcardTrie{key: "foo"},
			true,
		},
		{
			"simple equals",
			wildcardTrie{key: "foo", value: 1},
			wildcardTrie{key: "foo", value: 1},
			true,
		},
		{
			"equals with children",
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}}},
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}}},
			true,
		},
		{
			"unequal key",
			wildcardTrie{key: "foo"},
			wildcardTrie{key: "moo"},
			false,
		},
		{
			"unequal value",
			wildcardTrie{key: "foo", value: 1},
			wildcardTrie{key: "foo", value: 2},
			false,
		},
		{
			"with and without value",
			wildcardTrie{key: "foo"},
			wildcardTrie{key: "foo", value: 1},
			false,
		},
		{
			"unequal child value",
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}}},
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 2}}},
			false,
		},
		{
			"no child value",
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}}},
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar"}}},
			false,
		},
		{
			"different number of children",
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}}},
			wildcardTrie{key: "foo", value: 1,
				children: []wildcardTrie{{key: "bar", value: 1}, {key: "bla", value: 1}}},
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.left.equals(c.right) != c.want {
				t.Errorf("expected left == right")
			}
			if c.right.equals(c.left) != c.want {
				t.Errorf("expected right == left")
			}
		})
	}
}

func TestWildcardTrie_Get(t *testing.T) {
	basicTrie := wildcardTrie{
		separator: "/",
		value:     -1,
		children: []wildcardTrie{
			{"/", "moo", "/moo", 1, []wildcardTrie{{"/", "cow", "/moo/cow", 14, nil}}},
			{"/", "foo", "/foo", 2, []wildcardTrie{
				{"/", "bar", "/foo/bar", 3, nil},
				{"/", "*", "/foo/*", 99, nil},
				{"/", "bla", "/foo/bla", 5, []wildcardTrie{{"/", "*", "/foo/bla/*", 6, nil}}}}}},
	}
	cases := []struct {
		name        string
		tr          wildcardTrie
		input       string
		want        interface{}
		wantPattern string
	}{
		{"empty", basicTrie, "", -1, ""},
		// TODO(hvl): this needs to be explained in documentation
		{"only separator", basicTrie, "/", nil, ""},
		{"single key no separators", basicTrie, "foo", 2, "/foo"},
		{"single key with separator", basicTrie, "foo/", 99, "/foo/*"},
		{"leading separator single key", basicTrie, "/foo", 2, "/foo"},
		{"double keys", basicTrie, "/foo/bar", 3, "/foo/bar"},
		{"node is not a leaf", basicTrie, "foo/bar/", nil, ""},
		{"node with lower prio than wildcard", basicTrie, "foo/bla", 99, "/foo/*"},
		{"wildcard leaf", basicTrie, "foo/meow", 99, "/foo/*"},
		{"sub-node (with separators) with children", basicTrie, "foo/bla/", 6, "/foo/bla/*"},
		{"unknown", basicTrie, "moo/woof", nil, ""},
		{"unknown leaf", basicTrie, "moo/cowpie", nil, ""},
		{
			"unsupported partial wildcard",
			wildcardTrie{
				separator: "/", key: "", value: "", children: []wildcardTrie{
					{separator: "/", key: "foo*", value: 42},
				}},
			"foobar",
			nil,
			""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, pattern := c.tr.Get(c.input)
			if actual != c.want {
				t.Errorf("expected %v, got %v", c.want, actual)
			}
			if pattern != c.wantPattern {
				t.Errorf("expected (ok) %v, got %v", c.wantPattern, pattern)
			}
		})
	}
}

func TestWildcardTrie_Add(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	cases := []struct {
		name  string
		tr    wildcardTrie
		args  args
		want  *wildcardTrie
		panic string
	}{
		{
			"add first node",
			wildcardTrie{"/", "", "/", nil, nil},
			args{"foo", 1},
			&wildcardTrie{"/", "", "/", nil, []wildcardTrie{{"/", "foo", "/foo", 1, nil}}},
			"",
		},
		{
			"add to existing node",
			wildcardTrie{"/", "", "", nil, []wildcardTrie{{"/", "foo", "/foo", 1, nil}}},
			args{"foo/bar", 2},
			&wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{{"/", "bar", "/foo/bar", 2, nil}}}}},
			"",
		},
		{
			"add wildcard node to existing node",
			wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{{"/", "bar", "/foo/bar", 2, nil}}}}},
			args{"foo/*", 99},
			&wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil}}}}},
			"",
		},
		{
			"add wildcard node to existing sub-node",
			wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil}}}}},
			args{"foo/bla/*", 6},
			&wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil},
						{"/", "bla", "/foo/bla", nil, []wildcardTrie{
							{"/", "*", "/foo/bla/*", 6, nil}}}}}}},
			"",
		},
		{
			"set value on valueless existing sub-node",
			wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil},
						{"/", "bla", "/foo/bla", nil, []wildcardTrie{
							{"/", "*", "/foo/bla/*", 6, nil}}}}}}},
			args{"foo/bla", 5},
			&wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil},
						{"/", "bla", "/foo/bla", 5, []wildcardTrie{
							{"/", "*", "/foo/bla/*", 6, nil}}}}}}},
			"",
		},
		{
			"update value on existing sub-node",
			wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 2, nil},
						{"/", "*", "/foo/*", 99, nil},
						{"/", "bla", "/foo/bla", 5, []wildcardTrie{
							{"/", "*", "/foo/bla/*", 6, nil}}}}}}},
			args{"/foo/bar", 666},
			&wildcardTrie{
				"/", "", "", nil, []wildcardTrie{
					{"/", "foo", "/foo", 1, []wildcardTrie{
						{"/", "bar", "/foo/bar", 666, nil},
						{"/", "*", "/foo/*", 99, nil},
						{"/", "bla", "/foo/bla", 5, []wildcardTrie{
							{"/", "*", "/foo/bla/*", 6, nil}}}}}}},
			"",
		},
		{
			"! end in slash",
			wildcardTrie{separator: "/", key: ""},
			args{"foo/bar/", 1},
			nil,
			"path cannot end with slash",
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf(c.name), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if r != c.panic {
						t.Errorf("expected panic %s, got %s", c.panic, r)
					}
				}
			}()
			c.tr.Add(c.args.key, c.args.value)
			if !c.tr.equals(*c.want) {
				t.Errorf("\nexpected: %s,\ngot:      %s", c.want, c.tr)
			}
		})
	}
}
