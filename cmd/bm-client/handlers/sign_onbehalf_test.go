// Copyright (c) 2020 BitMaelum Authors
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

package handlers

import (
	"testing"

	testing2 "github.com/bitmaelum/bitmaelum-suite/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestSignOnbehalf(t *testing.T) {
	sourcePrivkey, _, _ := testing2.ReadTestKey("../../../testdata/key-ed25519-1.json")
	_, targetPubkey, _ := testing2.ReadTestKey("../../../testdata/key-ed25519-2.json")

	res, err := SignOnbehalf(*sourcePrivkey, *targetPubkey)
	assert.NoError(t, err)
	assert.Equal(t, "GRlSsdrkcciQSOiZNPHO0e0emRAv9x0rhB+eARdarkHeCJL9YOIHghen5C8+8IAaNGe0qHVPal+EcdXCKjQVBg==", res)
}
