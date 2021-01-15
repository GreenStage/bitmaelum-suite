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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitmaelum/bitmaelum-suite/internal/store"
	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// StoreGetKey will fetch an entry
func (api *API) StoreGetKey(addr hash.Hash, key string) (*store.EntryType, error) {
	keyHash := hash.New(addr.String() + key)

	body, statusCode, err := api.Get(fmt.Sprintf("/account/%s/store/%s", addr.String(), keyHash.String()))
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode > 299 {
		return nil, errNoSuccess
	}

	entry := &store.EntryType{}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// StorePutValue will store an value to a key
func (api *API) StorePutValue(kp bmcrypto.KeyPair, addr hash.Hash, key string, value string) error {
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
	var ph *hash.Hash = &parentHash

	var parent interface{} = parentHash.String()
	if key == "/" {
		parent = nil
		ph = nil
	}

	sig, err := generateSignature(kp.PrivKey, keyHash, ph, value)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(jsonOut{
		"key":        keyHash,
		"parent":     parent,
		"value":      []byte(value),
		"signature":  sig,
		"public_key": kp.PubKey.String(),
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

func generateSignature(privKey bmcrypto.PrivKey, keyHash hash.Hash, parentHash *hash.Hash, value string) ([]byte, error) {
	sha := sha256.New()
	sha.Write(keyHash.Byte())
	if parentHash != nil {
		sha.Write(parentHash.Byte())
	}
	sha.Write([]byte(value))
	out := sha.Sum(nil)

	return bmcrypto.Sign(privKey, out)
}
