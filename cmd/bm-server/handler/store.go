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

package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-server/internal/httputils"
	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/sirupsen/logrus"

	"github.com/bitmaelum/bitmaelum-suite/internal/container"

	"github.com/bitmaelum/bitmaelum-suite/internal/userstore"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/gorilla/mux"
)

var (
	errKeyNotFound = errors.New("store: key not found")
)

type inputStoreEntry struct {
	Data         string `json:"data"`
	Parent       string `json:"parent"`
	IsCollection bool   `json:"is_collection"`
}


// StoreGetRoot will get the root of the store, which is a collection of all other collections and keys
func StoreGetRoot(w http.ResponseWriter, req *http.Request) {
	getKey(w, hash.New("/"))
}

// StoreGet will retrieve a key or collection
func StoreGet(w http.ResponseWriter, req *http.Request) {
	keyHash, err := hash.NewFromHash(mux.Vars(req)["key"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	getKey(w, *keyHash)
}

func getKey(w http.ResponseWriter, keyHash hash.Hash) {
	// Check if key exists in database
	storesvc := container.Instance.GetStoreService()
	if !storesvc.HasKey(keyHash) {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	// Check if the key is root.
	if (keyHash == hash.New("/")) {
		_ = httputils.JSONOut(w, http.StatusOK, jsonOut{
			"parent": nil,
			"collection": true,
			"value": "This is the root collection. It cannot be deleted",
			"timestamp": internal.TimeNow(),
			"entries": []string{
				hash.New("boxes").String(),
				hash.New("contacts").String(),
				hash.New("bm-client-gui").String(),
			},
		})
	}

	entry, err := storesvc.GetKey(keyHash)
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	_ = httputils.JSONOut(w, http.StatusOK, jsonOut{
		"parent": entry.Parent,
		"collection": entry.IsCollection,
		"value": entry.Value,
		"timestamp": entry.TimeStamp,
		"entries": entry.Entries,
		"subcollections": 
	})


	var sinceTimestamp int64
	onlyIndex := true

	dump := req.URL.Query().Get("dump")
	if strings.ToLower(dump) == "true" {
		onlyIndex = false
	}

	since := req.URL.Query().Get("since")
	if since != "" {
		sinceTimestamp, err = strconv.ParseInt(since, 10, 64)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, "since param error")
			return
		}
	}

	logrus.Trace("RetrieveStore called for addr ", h, " and key ", k)

	repo := container.GetUserStoreRepo()
	entry, _ := repo.Fetch(h, k)
	var entries interface{}

	if entry.IsCollection || k == "" {
		entries, err = dumpStore(onlyIndex, h, k, sinceTimestamp)
		if err != nil {
			msg := fmt.Sprintf("error while retrieving store: %s", err)
			ErrorOut(w, http.StatusInternalServerError, msg)
			return
		}
	} else {
		n := make(map[string]interface{})
		n[entry.ID] = entry.Data
		entries = interface{}(n)
	}

	// Output entries
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(entries)

}

// StorePost will create/update a key or collection
func StorePost(w http.ResponseWriter, req *http.Request) {
	// Get variables from request
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	// Decode post body
	var input inputStoreEntry
	err = DecodeBody(w, req.Body, &input)
	if err != nil {
		ErrorOut(w, http.StatusBadRequest, "incorrect body")
		return
	}

	data, err := base64.StdEncoding.DecodeString(input.Data)
	if err != nil {
		ErrorOut(w, http.StatusBadRequest, "incorrect data")
		return
	}

	repo := container.GetUserStoreRepo()

	// Check if belongs to a collection
	if input.Parent != "" {
		parentEntry, err := repo.Fetch(h, input.Parent)
		if err != nil {
			ErrorOut(w, http.StatusBadRequest, "parent not found")
			return
		}

		// Check if parent is actually a collection
		if !parentEntry.IsCollection {
			ErrorOut(w, http.StatusBadRequest, "parent is not a collection")
			return
		}
	}

	// Generate new entry and store
	entry := userstore.NewEntry(k, data, input.Parent, input.IsCollection)
	err = repo.Store(h, entry)
	if err != nil {
		ErrorOut(w, http.StatusInternalServerError, "unable to store the data")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StoreRemove will remove a key or collection
func StoreRemove(w http.ResponseWriter, req *http.Request) {
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	repo := container.GetUserStoreRepo()
	if k == "" {
		// Delete the whole store
		entries, err := repo.Dump(h, k, 0)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, accountNotFound)
			return
		}

		for _, entry := range *entries {
			if entry.IsCollection {
				deleteCollection(repo, &h, &entry)
			}

			entry.IsCollection = false
			entry.Data = nil
			repo.Store(h, entry)
		}

	} else {
		key, _ := hash.NewFromHash(k)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, errkeyIsMissing.String)
			return
		}

		entry, err := repo.Fetch(h, key.String())
		if err != nil {
			msg := fmt.Sprintf("error while fetching key: %s", err)
			ErrorOut(w, http.StatusInternalServerError, msg)
			return
		}

		if entry.IsCollection {
			deleteCollection(repo, &h, entry)
		}

		entry.IsCollection = false
		entry.Data = nil
		repo.Store(h, *entry)
		//repo.Remove(h, entry.ID)

	}

	// All ok
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}




func getVariables(req *http.Request) (hash.Hash, string, error) {
	k := mux.Vars(req)["key"]
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		return "", "", errors.New("accountNotFound")
	}

	return *h, k, nil
}

func deleteCollection(repo userstore.Repository, addr *hash.Hash, entry *userstore.StoreEntry) {
	for _, k := range entry.Entries {
		newEntry, err := repo.Fetch(*addr, k)
		if err != nil {
			continue
		}

		if newEntry.IsCollection {
			deleteCollection(repo, addr, newEntry)
		} else {
			newEntry.Data = nil
			repo.Store(*addr, *newEntry)
			//repo.Remove(*addr, k)
		}
	}

	entry.Data = nil
	entry.IsCollection = false
	repo.Store(*addr, *entry)
	//repo.Remove(*addr, entry.ID)
}

func dumpStore(onlyIndex bool, addr hash.Hash, key string, sinceTimestamp int64) (interface{}, error) {
	repo := container.GetUserStoreRepo()

	var (
		entries *[]userstore.StoreEntry
		err     error
	)

	if onlyIndex {
		entries, err = repo.DumpIndex(addr, key, sinceTimestamp)
	} else {
		entries, err = repo.Dump(addr, key, sinceTimestamp)
	}

	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for _, entry := range *entries {
		logrus.Trace("entry -> ", entry.Parent, " ", entry.ID)

		if m[entry.Parent] == nil {
			m[entry.Parent] = make(map[string]interface{})
		}

		if m[entry.ID] == nil {
			if entry.IsCollection {
				m[entry.ID] = make(map[string]interface{})
			} else {
				m[entry.ID] = entry.Data
			}
		} else {
			// This check is because the key may have been changed from collection to deleted
			switch m[entry.ID].(type) {
			case map[string]interface{}:
				if !entry.IsCollection {
					m[entry.ID] = entry.Data
				}
			}
		}

		// We need to check if this key is still an interface type because if it's been removed it has changed to nil
		switch m[entry.Parent].(type) {
		case map[string]interface{}:
			m[entry.Parent].(map[string]interface{})[entry.ID] = m[entry.ID]
		}
	}
	if m[key] == nil {
		// Return empty interface
		return make(map[string]interface{}), nil
	}

	return m[key], nil
}
