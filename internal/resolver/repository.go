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

package resolver

import (
	"errors"

	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/bitmaelum/bitmaelum-suite/pkg/proofofwork"
)

var errKeyNotFound = errors.New("hash not found")

// Repository is a complete key resolver repository with the different parts
type Repository interface {
	AddressRepository
	RoutingRepository
	OrganisationRepository
}

// AddressRepository is the interface to manage address resolving
type AddressRepository interface {
	ResolveAddress(addr hash.Hash) (*AddressInfo, error)
	UploadAddress(info *AddressInfo, privKey bmcrypto.PrivKey, pow proofofwork.ProofOfWork) error
	DeleteAddress(info *AddressInfo, privKey bmcrypto.PrivKey) error
}

// RoutingRepository is the interface to manage route resolving
type RoutingRepository interface {
	ResolveRouting(routingID string) (*RoutingInfo, error)
	UploadRouting(info *RoutingInfo, privKey bmcrypto.PrivKey) error
	DeleteRouting(info *RoutingInfo, privKey bmcrypto.PrivKey) error
}

// OrganisationRepository is the interface to manage organisation resolving
type OrganisationRepository interface {
	ResolveOrganisation(orgHash hash.Hash) (*OrganisationInfo, error)
	UploadOrganisation(info *OrganisationInfo, privKey bmcrypto.PrivKey, pow proofofwork.ProofOfWork) error
	DeleteOrganisation(info *OrganisationInfo, privKey bmcrypto.PrivKey) error
}
