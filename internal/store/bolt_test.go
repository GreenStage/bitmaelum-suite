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

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/stretchr/testify/assert"
)

func TestBoltStorage(t *testing.T) {
	var (
		ok  bool
		err error
	)

	rand.Seed(time.Now().UnixNano())

	acc1 := hash.New("foo!")
	acc2 := hash.New("bar!")

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 01, 01, 12, 34, 56, 0, time.UTC)
	})

	// Unfortunately, boltdb cannot be used with afero
	path := filepath.Join(os.TempDir(), fmt.Sprintf("store-%d", rand.Int31()))
	_ = os.MkdirAll(filepath.Join(path, acc1.String()[:2], acc1.String()[2:]), 0755)

	defer func() {
		_ = os.RemoveAll(path)
	}()

	b := NewBoltRepository(path)
	assert.NotNil(t, b)

	err = b.OpenDb(acc1)
	assert.NoError(t, err)

	// Initially, only root should be present
	ok = b.HasEntry(acc1, makeHash(acc1, "/"))
	assert.True(t, ok)
	ok = b.HasEntry(acc1, makeHash(acc1, "/something"))
	assert.False(t, ok)

	// Incorrect hash
	ok = b.HasEntry(acc1, makeHash(acc2, "/"))
	assert.False(t, ok)

	// Get root entry
	entry, err := b.GetEntry(acc1, makeHash(acc1, "/"))
	assert.NoError(t, err)
	assert.NotNil(t, entry)

	// Add entry
	entry2 := NewEntry([]byte("foobar"))
	p := makeHash(acc1, "/")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts"), &p, entry2)
	assert.NoError(t, err)

	ok = b.HasEntry(acc1, makeHash(acc1, "/"))
	assert.True(t, ok)
	ok = b.HasEntry(acc1, makeHash(acc1, "/contacts"))
	assert.True(t, ok)

	entry, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "9f198242afd0a2660077b05c90c4aad8807b381f8e1af89e556c9a0e0e66331d", entry.Key.String())
	assert.Equal(t, "94723340d93b27ca21384fa64db760e10ee2382a3ded94f1e4243bacc24825e6", entry.Parent.String())
	assert.Equal(t, []byte("foobar"), entry.Data)

	// Create entry without correct path
	entry2 = NewEntry([]byte("foobar"))
	p = makeHash(acc1, "/path/not/exist")
	err = b.SetEntry(acc1, makeHash(acc1, "/path/not/exist/item"), &p, entry2)
	assert.Error(t, err)
}

func TestTimePropagation(t *testing.T) {
	var (
		err error
	)

	rand.Seed(time.Now().UnixNano())

	acc1 := hash.New("foo!")

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 01, 01, 12, 34, 56, 0, time.UTC)
	})

	// Unfortunately, boltdb cannot be used with afero
	path := filepath.Join(os.TempDir(), fmt.Sprintf("store-%d", rand.Int31()))
	_ = os.MkdirAll(filepath.Join(path, acc1.String()[:2], acc1.String()[2:]), 0755)

	defer func() {
		_ = os.RemoveAll(path)
	}()

	b := NewBoltRepository(path)
	assert.NotNil(t, b)

	err = b.OpenDb(acc1)
	assert.NoError(t, err)

	// Initially, only root should be present
	entry := NewEntry([]byte("contact list"))
	p := makeHash(acc1, "/")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("john doe"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/1"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("foo bar"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/2"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("jane austin"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/3"), &p, entry)
	assert.NoError(t, err)

	entry2, err := b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Len(t, entry2.Entries, 3)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/1"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/2"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	// New entry

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 05, 05, 12, 34, 56, 0, time.UTC)
	})

	entry = NewEntry([]byte("latest entry"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/7"), &p, entry)
	assert.NoError(t, err)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Len(t, entry2.Entries, 4)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/1"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/7"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1273062896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/2"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1273062896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1273062896), entry2.Timestamp)

	// Update entry

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 8, 8, 12, 34, 56, 0, time.UTC)
	})

	entry = NewEntry([]byte("update entry"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/2"), &p, entry)
	assert.NoError(t, err)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Len(t, entry2.Entries, 4)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/1"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/7"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1273062896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/2"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1281270896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1281270896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1281270896), entry2.Timestamp)
}

func TestRemoveEntries(t *testing.T) {
	var (
		err error
	)

	rand.Seed(time.Now().UnixNano())

	acc1 := hash.New("foo!")

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 01, 01, 12, 34, 56, 0, time.UTC)
	})

	// Unfortunately, boltdb cannot be used with afero
	path := filepath.Join(os.TempDir(), fmt.Sprintf("store-%d", rand.Int31()))
	_ = os.MkdirAll(filepath.Join(path, acc1.String()[:2], acc1.String()[2:]), 0755)

	defer func() {
		_ = os.RemoveAll(path)
	}()

	b := NewBoltRepository(path)
	assert.NotNil(t, b)

	err = b.OpenDb(acc1)
	assert.NoError(t, err)

	// Initially, only root should be present
	entry := NewEntry([]byte("contact list"))
	p := makeHash(acc1, "/")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("john doe"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/1"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("foo bar"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/2"), &p, entry)
	assert.NoError(t, err)

	entry = NewEntry([]byte("jane austin"))
	p = makeHash(acc1, "/contacts")
	err = b.SetEntry(acc1, makeHash(acc1, "/contacts/3"), &p, entry)
	assert.NoError(t, err)

	entry2, err := b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Len(t, entry2.Entries, 3)

	// Remove entry

	internal.SetMockTime(func() time.Time {
		return time.Date(2010, 8, 8, 12, 34, 56, 0, time.UTC)
	})

	// Cannot remove collections
	err = b.RemoveEntry(acc1, makeHash(acc1, "/contacts"), true)
	assert.Error(t, err)
	err = b.RemoveEntry(acc1, makeHash(acc1, "/"), true)
	assert.Error(t, err)

	// Remove second entry
	err = b.RemoveEntry(acc1, makeHash(acc1, "/contacts/2"), true)
	assert.NoError(t, err)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Len(t, entry2.Entries, 2)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/1"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	ok := b.HasEntry(acc1, makeHash(acc1, "/contacts/2"))
	assert.False(t, ok)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts/3"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1262349296), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/contacts"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1281270896), entry2.Timestamp)

	entry2, err = b.GetEntry(acc1, makeHash(acc1, "/"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1281270896), entry2.Timestamp)
}

func makeHash(account hash.Hash, key string) hash.Hash {
	return hash.New(account.String() + key)
}
