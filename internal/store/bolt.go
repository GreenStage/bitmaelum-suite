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
	"errors"
	"path/filepath"
	"strings"

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	errKeyNotFound   = errors.New("store: key not found")
	errNoEmptyParent = errors.New("store: no empty parent allowed")
)

type boltRepo struct {
	Clients map[string]*bolt.DB
	path    string
}

// BucketName is the bucket name to save the sote
const BucketName = "store"

// BoltDBFile is the filename to store the boltdb database
const BoltDBFile = "store.db"

// NewBoltRepository initializes a new repository
func NewBoltRepository(accountsPath string) Repository {
	return &boltRepo{
		Clients: make(map[string]*bolt.DB),
		path:    accountsPath,
	}
}

// OpenDB will try and open the store database
func (b boltRepo) OpenDb(account hash.Hash) error {
	// Open file
	p := filepath.Join(b.path, BoltDBFile)
	logrus.Trace("opening boltdb file: ", p)

	db, err := bolt.Open(p, 0600, nil)
	if err != nil {
		logrus.Trace("error while opening boltdb: ", err)
		return err
	}

	// Store in cache
	b.Clients[account.String()] = db

	// Check if root exists
	if !b.HasEntry(account, "/") {
		entry := &StoreEntryType{
			Timestamp: internal.TimeNow().Unix(),
		}
		err := b.SetEntry(account, "/", *entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// CloseDb will close the store database - if openened
func (b boltRepo) CloseDb(account hash.Hash) error {
	// check if db exists
	db, ok := b.Clients[account.String()]
	if !ok {
		return nil
	}

	delete(b.Clients, account.String())
	return db.Close()
}

// HasEntry will return true when the database has the specific key present
func (b boltRepo) HasEntry(account hash.Hash, key string) bool {
	client, err := b.getClientDb(account)
	if err != nil {
		return false
	}

	err = client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			return errKeyNotFound
		}

		keyHash := hash.New(account.String() + key)
		data := bucket.Get(keyHash.Byte())
		if data == nil {
			return errKeyNotFound
		}

		return nil
	})

	return err == nil
}

// GetEntry will return the given entry
func (b boltRepo) GetEntry(account hash.Hash, key string) (*StoreEntryType, error) {
	client, err := b.getClientDb(account)
	if err != nil {
		return nil, err
	}

	entry := &StoreEntryType{}

	err = client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			return errKeyNotFound
		}

		keyHash := hash.New(account.String() + key)
		data := bucket.Get(keyHash.Byte())
		if data == nil {
			return errKeyNotFound
		}

		err := json.Unmarshal(data, &entry)
		if err != nil {
			return errKeyNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (b boltRepo) SetEntry(account hash.Hash, key string, entry StoreEntryType) error {
	client, err := b.getClientDb(account)
	if err != nil {
		return err
	}

	// Update entry and tree back to root with this timestamp
	lastUpdateTimestamp := internal.TimeNow().Unix()

	parents := getParentHashes(account, key)

	// Populate key and parent hash in our store entry
	keyHash := hash.New(account.String() + key)
	entry.Key = keyHash
	if len(parents) > 0 {
		entry.Parent = &parents[len(parents)-1].Hash
	}

	return client.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", BucketName, err)
			return err
		}

		// Store entry
		buf, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		err = bucket.Put(entry.Key.Byte(), buf)
		if err != nil {
			return err
		}

		// Iterate parents back to root and update
		for _, parent := range parents {
			// Get parent entry
			data := bucket.Get(parent.Hash.Byte())
			if data == nil {
				// Parent not found. Add a new one
				nextEntry := &StoreEntryType{
					Key:       parent.Hash,
					Timestamp: 0,
					Entries:   []hash.Hash{entry.Key},
				}
				entry = *nextEntry
			} else {
				err = json.Unmarshal(data, &entry)
				if err != nil {
					return err
				}
			}

			// Store current entry
			entry.Timestamp = lastUpdateTimestamp

			buf, err := json.Marshal(entry)
			if err != nil {
				return err
			}

			err = bucket.Put(entry.Key.Byte(), buf)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveEntry will remove the key from the database, and update the collection tree
func (b boltRepo) RemoveEntry(account hash.Hash, key string) error {
	panic("implement me")
}

// getClientDB will open or create the account's store database
func (b boltRepo) getClientDb(account hash.Hash) (*bolt.DB, error) {
	// Fetch db file from cache
	db, ok := b.Clients[account.String()]
	if ok {
		return db, nil
	}

	// Open/create if not found in cache
	err := b.OpenDb(account)
	if err != nil {
		return nil, err
	}

	return b.Clients[account.String()], nil
}

type ParentMapType struct {
	Key  string
	Hash hash.Hash
}

func getParentHashes(addr hash.Hash, key string) []ParentMapType {
	// Assume always absolute from root. Remove the root if present
	if len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}

	// Root key, does not have a parent
	if key == "" {
		return nil
	}

	var parents []ParentMapType

	parts := strings.Split(key, "/")
	for len(parts) > 0 {
		parts = parts[:len(parts)-1]

		parentKey := strings.Join(parts, "/")

		parents = append(parents, ParentMapType{
			Key:  "/" + parentKey,
			Hash: hash.New(addr.String() + "/" + parentKey),
		})
	}

	return parents
}
