package middleware

import (
	"context"
	"errors"
	"github.com/bitmaelum/bitmaelum-suite/core/container"
	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/internal/encrypt"
	"github.com/bitmaelum/bitmaelum-suite/pkg/address"
	"github.com/gorilla/mux"
	"github.com/vtolstov/jwt-go"
	"net/http"
	"strings"
)

// JwtToken is a middleware that automatically verifies given JWT token
type JwtToken struct{}

type claimsContext string
type addressContext string

// ErrTokenNotValidated is returend when the token could not be validated (for any reason)
var ErrTokenNotValidated = errors.New("token could not be validated")

// Middleware JWT token authentication
func (*JwtToken) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		haddr, err := address.NewHashFromHash(mux.Vars(req)["addr"])
		if err != nil {
			http.Error(w, "Cannot authorize without address", http.StatusUnauthorized)
			return
		}

		as := container.GetAccountService()
		if !as.AccountExists(*haddr) {
			http.Error(w, "Address not found", http.StatusUnauthorized)
			return
		}

		token, err := checkToken(req.Header.Get("Authorization"), *haddr)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, claimsContext("claims"), token.Claims)
		ctx = context.WithValue(ctx, addressContext("address"), token.Claims.(*jwt.StandardClaims).Subject)

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// Check if the authorization contains a valid JWT token for the given address
func checkToken(auth string, addr address.HashAddress) (*jwt.Token, error) {
	if auth == "" {
		return nil, ErrTokenNotValidated
	}

	if len(auth) <= 6 || strings.ToUpper(auth[0:7]) != "BEARER " {
		return nil, ErrTokenNotValidated
	}
	tokenString := auth[7:]

	as := container.GetAccountService()
	keys := as.GetPublicKeys(addr)
	for _, key := range keys {
		pubKey, err := encrypt.PEMToPubKey([]byte(key))
		if err != nil {
			continue
		}

		token, err := internal.ValidateJWTToken(tokenString, addr, pubKey)
		if err == nil {
			return token, nil
		}
	}

	return nil, ErrTokenNotValidated
}
