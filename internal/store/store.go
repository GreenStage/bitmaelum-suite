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
	"encoding/json"
	"fmt"

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// StoreEntryType is the structure that defines a key entry
type StoreEntryType struct {
	Key            hash.Hash
	Parent         hash.Hash
	IsCollection   bool
	Data           []byte
	Timestamp      uint64
	Entries        []hash.Hash
	SubCollections []hash.Hash
}

// MarshalBinary converts a storeentrytype to binary format so it can be stored in Redis
func (e *StoreEntryType) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

// UnmarshalBinary converts binary to a ticket so it can be fetched from Redis
func (e *StoreEntryType) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

// Repository is a store repository to fetch and store tickets
type Repository interface {
	HasKey(account hash.Hash, key hash.Hash) bool
	RemoveKey(account hash.Hash, key hash.Hash) error
	GetKey(account hash.Hash, key hash.Hash) (*StoreEntryType, error)

	OpenDb(account hash.Hash) error
	CloseDb(account hash.Hash) error
}

func createStoreKey(id string) string {
	return fmt.Sprintf("store-%s", id)
}
