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

package message

import (
	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// Compose will create a new message and places it inside an envelope. This can be used for actual sending the message
func Compose(addressing Addressing, subject string, b, a []string) (*Envelope, error) {
	var (
		senderPrivKey   *bmcrypto.PrivKey
		recipientPubKey *bmcrypto.PubKey
		cat             *Catalog
	)

	switch addressing.Type {
	case SignedByTypeOrigin:
		senderPrivKey = addressing.Sender.PrivKey
		recipientPubKey = addressing.Recipient.PubKey
		cat = NewCatalog(addressing, subject)

	case SignedByTypeServer:
		senderPrivKey = addressing.Sender.PrivKey
		recipientPubKey = addressing.Recipient.PubKey
		cat = NewCatalog(addressing, subject)
		cat.AddFlags("postmaster")

	case SignedByTypeAuthorized:
		senderPrivKey = addressing.Sender.PrivKey
		recipientPubKey = addressing.Recipient.PubKey
		cat = NewCatalog(addressing, subject)
		cat.AddFlags("onbehalf")
	}

	// Create envelope
	envelope, err := NewEnvelope()
	if err != nil {
		return nil, err
	}

	// Add catalog
	cat, err = finishCatalog(cat, b, a)
	if err != nil {
		return nil, err
	}

	err = envelope.AddCatalog(cat)
	if err != nil {
		return nil, err
	}

	// Add header
	header, err := generateHeader(addressing.Sender.Address.Hash(), addressing.Recipient.Address.Hash(), addressing.Type)
	if err != nil {
		return nil, err
	}

	err = envelope.AddHeader(header)
	if err != nil {
		return nil, err
	}

	// Close the envelope for sending
	err = envelope.CloseAndEncrypt(senderPrivKey, recipientPubKey)
	if err != nil {
		return nil, err
	}

	return envelope, nil
}

// Generate a header file based on the info provided
func generateHeader(sender, recipient hash.Hash, origin SignedByType) (*Header, error) {
	header := &Header{}

	header.From.Addr = sender
	header.To.Addr = recipient
	header.From.SignedBy = origin

	return header, nil
}

func finishCatalog(cat *Catalog, b, a []string) (*Catalog, error) {
	// Add blocks to catalog
	blocks, err := GenerateBlocks(b)
	if err != nil {
		return nil, err
	}
	for _, block := range blocks {
		err := cat.AddBlock(block)
		if err != nil {
			return nil, err
		}
	}

	// Add attachments to catalog
	attachments, err := GenerateAttachments(a)
	if err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		err := cat.AddAttachment(attachment)
		if err != nil {
			return nil, err
		}
	}

	return cat, nil
}
