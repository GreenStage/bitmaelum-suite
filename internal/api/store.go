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

package api

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitmaelum/bitmaelum-suite/internal/store"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

func (api *API) StoreGetKey(addr hash.Hash, key string) (*store.StoreEntryType, error) {
	keyHash := hash.New(addr.String() + key)

	body, statusCode, err := api.Get(fmt.Sprintf("/account/%s/store/%s", addr.String(), keyHash.String()))
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode > 299 {
		return nil, errNoSuccess
	}

	entry := &store.StoreEntryType{}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (api *API) StorePutValue(addr hash.Hash, key string, value string) error {
	// Calc parentKey
	parentKey, _ := filepath.Split(key)
	// correct "/foo/" to "/foo" from "/foo/bar"
	parentKey = strings.TrimRight(parentKey, "/")
	// correct "" to "/" from "/foo"
	if parentKey == "" {
		parentKey = "/"
	}

	keyHash := hash.New(addr.String() + key)
	parentHash := hash.New(addr.String() + parentKey)

	var parent interface{} = parentHash.String()
	if key == "/" {
		parent = nil
	}

	data, err := json.MarshalIndent(jsonOut{
		"parent": parent,
		"value":  []byte(value),
	}, "", "  ")
	if err != nil {
		return err
	}

	_, statusCode, err := api.Post(fmt.Sprintf("/account/%s/store/%s", addr.String(), keyHash.String()), data)
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode > 299 {
		return errNoSuccess
	}

	return nil

}
