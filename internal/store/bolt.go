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

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	bolt "go.etcd.io/bbolt"
)

var errKeyNotFound = errors.New("store: key not found")

type boltRepo struct {
	Clients map[string]*bolt.DB
	path    string
}

//BucketName is the bucket name to store the tickets on the bolt db
const BucketName = "store"

//BoltDBFile is the filename to store the boltdb database
const BoltDBFile = "store.db"

// NewBoltRepository initializes a new repository
func NewBoltRepository(accountsPath string) Repository {
	// dbFile := filepath.Join(dbpath, BoltDBFile)
	// db, err := bolt.Open(dbFile, 0600, nil)
	// if err != nil {
	// 	logrus.Error("Unable to open filepath ", dbFile, err)
	// 	return nil
	// }

	return &boltRepo{
	 	Clients: make(map[string]*bolt.DB),
	 	path: accountsPath,
	}
}

// OpenDB will try and open the store database
func (b boltRepo) OpenDb(account hash.Hash) error {
	// Does nothing. Database will be opened when fetching a client
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

// HasKey will return true when the database has the specific key present
func (b boltRepo) HasKey(account hash.Hash, key hash.Hash) bool {
	client, err := b.getClientDb(account)
	if err != nil {
		return false
	}

	err = client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			return errKeyNotFound
		}

		data := bucket.Get([]byte(key.String()))
		if data == nil {
			return errKeyNotFound
		}

		return nil
	})

	return err == nil
}

// GetKey will return the given entry
func (b boltRepo) GetKey(account hash.Hash, key hash.Hash) (*StoreEntryType, error) {
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

		data := bucket.Get([]byte(key.String()))
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

// RemoveKey will remove the key from the database, and update the collection tree
func (b boltRepo) RemoveKey(account hash.Hash, key hash.Hash) error {
	panic("implement me")
}


// getClientDB will open or create the account's store database
func (b boltRepo) getClientDb(account hash.Hash) (*bolt.DB, error) {
	// Fetch db file from cache
	db, ok := b.Clients[account.String()]
	if ok {
		return db, nil
	}

	// Open file
	p := filepath.Join(b.path, BoltDBFile)
	db, err := bolt.Open(p, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Store in cache
	b.Clients[account.String()] = db
	return db, nil
}
