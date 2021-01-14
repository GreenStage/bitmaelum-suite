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
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-server/internal/httputils"
	"github.com/bitmaelum/bitmaelum-suite/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/internal/store"

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/gorilla/mux"
)

var (
	errKeyNotFound = errors.New("store: key not found")
)

// StoreGetRoot will get the root of the store, which is a collection of all other collections and keys
func StoreGetRoot(w http.ResponseWriter, req *http.Request) {
	haddr, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	getKey(w, *haddr, hash.New(haddr.String() + "/"))
}

// StoreGet will retrieve a key or collection
func StoreGet(w http.ResponseWriter, req *http.Request) {
	haddr, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	keyHash, err := hash.NewFromHash(mux.Vars(req)["key"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	getKey(w, *haddr, *keyHash)
}


// UpdateType is a request for a store entry
type UpdateType struct {
	Parent      *hash.Hash `json:"parent"`
	Value       []byte    `json:"value"`
}


// StoreUpdate will update a key or collection
func StoreUpdate(w http.ResponseWriter, req *http.Request) {
	updateRequest := &UpdateType{}
	err := json.NewDecoder(req.Body).Decode(updateRequest)
	if err != nil {
		httputils.ErrorOut(w, http.StatusBadRequest, "Malformed JSON: " + err.Error())
		return
	}

	haddr, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	keyHash, err := hash.NewFromHash(mux.Vars(req)["key"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	storeKey(w, *haddr, *keyHash, updateRequest.Parent, updateRequest.Value)
}

// StoreDelete will remove a key or collection
func StoreDelete(w http.ResponseWriter, req *http.Request) {
	haddr, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	keyHash, err := hash.NewFromHash(mux.Vars(req)["key"])
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	deleteKey(w, *haddr, *keyHash)
}


func storeKey(w http.ResponseWriter, addrHash, keyHash hash.Hash, parentHash *hash.Hash, value []byte) {
	err := openDb(w, addrHash)
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}
	defer closeDb(addrHash)


	// Add entry
	entry := &store.StoreEntryType{
		Data:           value,
	}
	storesvc := container.Instance.GetStoreRepo()
	err = storesvc.SetEntry(addrHash, keyHash, parentHash, *entry)
	if err != nil {
		httputils.ErrorOut(w, http.StatusInternalServerError, errKeyNotFound.Error())
		return
	}

	_ = httputils.JSONOut(w, http.StatusOK, nil)
}

func deleteKey(w http.ResponseWriter, addrHash hash.Hash, keyHash hash.Hash) {
	err := openDb(w, addrHash)
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}
	defer closeDb(addrHash)

	// Check if key exists in database
	storesvc := container.Instance.GetStoreRepo()
	if !storesvc.HasEntry(addrHash, keyHash) {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	err = storesvc.RemoveEntry(addrHash, keyHash, false)
	if err != nil {
		httputils.ErrorOut(w, http.StatusInternalServerError, errKeyNotFound.Error())
		return
	}

	_ = httputils.JSONOut(w, http.StatusNoContent, nil)
}

func getKey(w http.ResponseWriter, addrHash hash.Hash, keyHash hash.Hash) {
	err := openDb(w, addrHash)
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}
	defer closeDb(addrHash)

	// Check if key exists in database
	storesvc := container.Instance.GetStoreRepo()
	if !storesvc.HasEntry(addrHash, keyHash) {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	entry, err := storesvc.GetEntry(addrHash, keyHash)
	if err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return
	}

	_ = httputils.JSONOut(w, http.StatusOK, entry)
}

func openDb(w http.ResponseWriter, addrHash hash.Hash) error {
	// Open DB
	storesvc := container.Instance.GetStoreRepo()
	if err := storesvc.OpenDb(addrHash); err != nil {
		httputils.ErrorOut(w, http.StatusNotFound, errKeyNotFound.Error())
		return errors.New("cannot open db")
	}

	return nil
}

func closeDb(addrhash hash.Hash) {
	storesvc := container.Instance.GetStoreRepo()
	_ = storesvc.CloseDb(addrhash)
}
