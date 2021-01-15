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

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// EntryType is the structure that defines a key entry
type EntryType struct {
	Key       hash.Hash   `json:"key"`       // Key of the entry
	Parent    *hash.Hash  `json:"parent"`    // key of the parent, or nil when it's the root
	Data      []byte      `json:"data"`      // actual (encrypted) data
	Timestamp int64       `json:"timestamp"` // Timestamp of this entry, or the highest timestamp of any entry below
	Entries   []hash.Hash `json:"entries"`   // Keys to entries below this
	Signature []byte      `json:"signature"` // Signature for this entry
}

// NewEntry creates a new entry
func NewEntry(data []byte) EntryType {
	return EntryType{
		Data:      data,
		Timestamp: internal.TimeNow().Unix(),
	}
}

// MarshalBinary converts a storeentrytype to binary format so it can be stored in Redis
func (e *EntryType) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

// UnmarshalBinary converts binary to a ticket so it can be fetched from Redis
func (e *EntryType) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

// Repository is a store repository to fetch and store tickets
type Repository interface {
	HasEntry(account, key hash.Hash) bool
	RemoveEntry(account, key hash.Hash, recursive bool) error
	GetEntry(account, key hash.Hash) (*EntryType, error)
	SetEntry(account, key hash.Hash, parent *hash.Hash, entry EntryType) error

	OpenDb(account hash.Hash) error
	CloseDb(account hash.Hash) error
}
