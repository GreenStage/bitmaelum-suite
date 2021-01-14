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

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	errKeyNotFound   = errors.New("store: key not found")
	errParentNotFound = errors.New("store: parent entry not found")
	errCannotRemoveCollection = errors.New("store: cannot remove collection")
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

	rootHash := hash.New(account.String() + "/")

	// Check if root exists
	if !b.HasEntry(account, rootHash) {
		entry := &StoreEntryType{
			Timestamp: internal.TimeNow().Unix(),
		}

		err := b.SetEntry(account, rootHash, nil, *entry)
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
func (b boltRepo) HasEntry(account, key hash.Hash) bool {
	client, err := b.getClientDb(account)
	if err != nil {
		return false
	}

	err = client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			return errKeyNotFound
		}

		data := bucket.Get(key.Byte())
		if data == nil {
			return errKeyNotFound
		}

		return nil
	})

	return err == nil
}

// GetEntry will return the given entry
func (b boltRepo) GetEntry(account, key hash.Hash) (*StoreEntryType, error) {
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

		entry = getFromBucket(bucket, key)
		if entry == nil {
			return errKeyNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (b boltRepo) SetEntry(account, key hash.Hash, parent *hash.Hash, entry StoreEntryType) error {
	client, err := b.getClientDb(account)
	if err != nil {
		return err
	}

	// Check if parent exists
	if parent != nil && ! b.HasEntry(account, *parent) {
		return errParentNotFound
	}

	// Update entry and tree back to root with this timestamp
	lastUpdateTimestamp := internal.TimeNow().Unix()

	// Update entry values
	entry.Timestamp = lastUpdateTimestamp
	entry.Key = key
	entry.Parent = parent

	return client.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", BucketName, err)
			return err
		}

		err = putInBucket(bucket, entry)
		if err != nil {
			return err
		}

		// Update all parents
		return updateParentEntries(bucket, entry, lastUpdateTimestamp)
	})
}

func getFromBucket(bucket *bolt.Bucket, key hash.Hash) *StoreEntryType {
	data := bucket.Get(key.Byte())
	if data == nil {
		return nil
	}

	entry := &StoreEntryType{}
	err := json.Unmarshal(data, &entry)
	if err != nil {
		return nil
	}

	return entry
}

func putInBucket(bucket *bolt.Bucket, entry StoreEntryType) error {
	buf, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return bucket.Put(entry.Key.Byte(), buf)
}

// RemoveEntry will remove the key from the database, and update the collection tree
func (b boltRepo) RemoveEntry(account, key hash.Hash, recursive bool) error {
	client, err := b.getClientDb(account)
	if err != nil {
		return err
	}

	entry, err := b.GetEntry(account, key)
	if err != nil {
		return errKeyNotFound
	}

	// @TODO: recursive deletion is not yet supported
	if len(entry.Entries) > 0 {
		return errCannotRemoveCollection
	}

	return client.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", BucketName, err)
			return err
		}

		// Remove actual entry
		err = bucket.Delete(entry.Key.Byte())
		if err != nil {
			return err
		}

		// Get parent entry and remove entry reference
		parentEntry := &StoreEntryType{}
		data := bucket.Get(entry.Parent.Byte())
		err = json.Unmarshal(data, &parentEntry)
		if err != nil {
			return err
		}

		// Update entry and tree back to root with this timestamp
		lastUpdateTimestamp := internal.TimeNow().Unix()

		parentEntry.Timestamp = lastUpdateTimestamp
		parentEntry.Entries = removeFromEntries(parentEntry.Entries, entry.Key)

		err = putInBucket(bucket, *parentEntry)
		if err != nil {
			return err
		}

		// Update all parents
		return updateParentEntries(bucket, *parentEntry, lastUpdateTimestamp)
	})
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


// updateEntries will add the key, but only when it's not yet present in the list
func updateEntries(entries []hash.Hash, key hash.Hash) []hash.Hash {
	for i := range entries {
		if entries[i].String() == key.String() {
			return entries
		}
	}

	entries = append(entries, key)
	return entries
}

// removeFromEntries will add the key, but only when it's not yet present in the list
func removeFromEntries(entries []hash.Hash, key hash.Hash) []hash.Hash {
	// Find element in list
	found := -1
	for i := range entries {
		if entries[i].String() == key.String() {
			found = i
		}
	}

	if found == -1 {
		return entries
	}

	return append(entries[:found], entries[found+1:]...)
}

func updateParentEntries(bucket *bolt.Bucket, entry StoreEntryType, ts int64) error {
	parent := entry.Parent

	for parent != nil {
		curEntryKey := entry.Key

		// Get parent entry
		entry := getFromBucket(bucket, *parent)
		if entry == nil {
			return errParentNotFound
		}

		// Update this parent entry
		entry.Timestamp = ts
		entry.Entries = updateEntries(entry.Entries, curEntryKey)

		// Save back
		err := putInBucket(bucket, *entry)
		if err != nil {
			return err
		}

		// Onto the next parent until we hit nil
		parent = entry.Parent
	}

	return nil
}
