// Copyright (c) 2021 BitMaelum Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package store

import (
	"testing"
)

func TestNewEntry(t *testing.T) {
	// acc1 := hash.New("foo!")
	// acc2 := hash.New("bar!")
	//
	// e := NewEntry(acc1, []byte("foobar"), "/")
	// assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", e.Key.String())
	// assert.Nil(t, e.Parent)
	//
	// e = NewEntry(acc2, []byte("foobar"), "/")
	// assert.Equal(t, "1f1cce7df0c9a930ba4618be5025c2f25eb93ca3574ccc9b8197ba323399ddb1", e.Key.String())
	// assert.Nil(t, e.Parent)
	//
	// e = NewEntry(acc1, []byte("foobar"), "/foo")
	// assert.Equal(t, "f2f5d73819bf7302d137500293b85e5e13e8c2069e3f3ad85fa4ad8ea7ed1efe", e.Key.String())
	// assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", e.Parent.String())
	//
	// e = NewEntry(acc1, []byte("foobar"), "/foo/bar")
	// assert.Equal(t, "79780c884b68f0bb259371679413fc3607c3c4bc9eef2d675ab5266e09f04bce", e.Key.String())
	// assert.Equal(t, "f2f5d73819bf7302d137500293b85e5e13e8c2069e3f3ad85fa4ad8ea7ed1efe", e.Parent.String())
	//
	// e = NewEntry(acc1, []byte("foobar"), "/foo/bar/qux")
	// assert.Equal(t, "86eee6e98f7d4488325ee8cd22c4295bf2ba21927e07b1d5671b70c3cdfb6747", e.Key.String())
	// assert.Equal(t, "79780c884b68f0bb259371679413fc3607c3c4bc9eef2d675ab5266e09f04bce", e.Parent.String())
}
