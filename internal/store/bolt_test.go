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
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/stretchr/testify/assert"
)

func TestParentHashes(t *testing.T) {
	acc1 := hash.New("foo!")

	parents := getParentHashes(acc1, "/foo/bar/baz")
	assert.Len(t, parents, 3)
	assert.Equal(t, "/foo/bar", parents[0].Key)
	assert.Equal(t, "79780c884b68f0bb259371679413fc3607c3c4bc9eef2d675ab5266e09f04bce", parents[0].Hash.String())
	assert.Equal(t, "/foo", parents[1].Key)
	assert.Equal(t, "f2f5d73819bf7302d137500293b85e5e13e8c2069e3f3ad85fa4ad8ea7ed1efe", parents[1].Hash.String())
	assert.Equal(t, "/", parents[2].Key)
	assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", parents[2].Hash.String())

	parents = getParentHashes(acc1, "")
	assert.Len(t, parents, 0)
	assert.Nil(t, parents)

	parents = getParentHashes(acc1, "/")
	assert.Len(t, parents, 0)
	assert.Nil(t, parents)

	parents = getParentHashes(acc1, "/foo/bar")
	assert.Len(t, parents, 2)
	assert.Equal(t, "/foo", parents[0].Key)
	assert.Equal(t, "f2f5d73819bf7302d137500293b85e5e13e8c2069e3f3ad85fa4ad8ea7ed1efe", parents[0].Hash.String())
	assert.Equal(t, "/", parents[1].Key)
	assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", parents[1].Hash.String())

	parents = getParentHashes(acc1, "/foo")
	assert.Len(t, parents, 1)
	assert.Equal(t, "/", parents[0].Key)
	assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", parents[0].Hash.String())
}

func TestBoltStorage(t *testing.T) {
	var (
		ok bool
		err error
	)

	rand.Seed(time.Now().UnixNano())

	acc1 := hash.New("foo!")
	// acc2 := hash.New("bar!")

	// Unfortunately, boltdb cannot be used with afero
	path := filepath.Join(os.TempDir(), fmt.Sprintf("store-%d", rand.Int31()))
	_ = os.MkdirAll(path, 0755)

	defer func() {
		_ = os.Remove(path)
	}()

	b := NewBoltRepository(path)
	assert.NotNil(t, b)

	err = b.OpenDb(acc1)
	assert.NoError(t, err)

	// Initially, only root should be present
	ok = b.HasEntry(acc1, "/")
	assert.True(t, ok)
	ok = b.HasEntry(acc1, "/something")
	assert.False(t, ok)

	// Get root entry
	entry, err := b.GetEntry(acc1, "/")
	assert.NoError(t, err)
	assert.NotNil(t, entry)

	// Add entry
	entry2 := NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/contacts", entry2)
	assert.NoError(t, err)

	ok = b.HasEntry(acc1, "/")
	assert.True(t, ok)
	ok = b.HasEntry(acc1, "/contacts")
	assert.True(t, ok)


	entry, err = b.GetEntry(acc1, "/contacts")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "9f198242afd0a2660077b05c90c4aad8807b381f8e1af89e556c9a0e0e66331d", entry.Key.String())
	assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", entry.Parent.String())
	assert.Equal(t, []byte("foobar"), entry.Data)


	// Create path
	entry2 = NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/foo", entry2)
	assert.NoError(t, err)
	entry2 = NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/foo/bar", entry2)
	assert.NoError(t, err)
	entry2 = NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/foo/bar/baz", entry2)
	assert.NoError(t, err)

	entry2 = NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/foo/bar/baz/baq", entry2)
	assert.NoError(t, err)

	entry, err = b.GetEntry(acc1, "/foo/bar/baz/baq")
	assert.NoError(t, err)
	assert.NotNil(t, entry)



	entry2 = NewEntry([]byte("foobar"))
	err = b.SetEntry(acc1, "/this/path/must/be/set", entry2)
	assert.NoError(t, err)

	ok = b.HasEntry(acc1, "/this/path")
	assert.True(t, ok)

	ok = b.HasEntry(acc1, "/this/path/does/not/exists")
	assert.False(t, ok)
}

