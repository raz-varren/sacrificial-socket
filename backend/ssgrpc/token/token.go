/*
Package token is used by package ssgrpc for generating and validating HMAC-SHA256 tokens sent as "per RPC credentials".
*/
package token

import (
	"encoding/json"
	"errors"
	"github.com/dvsekhvalnov/jose2go"
	"time"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrBadPayload   = errors.New("payload is missing critical values")

	ErrNoToken        = errors.New("user did not provide a token")
	ErrBadBearerValue = errors.New("user provided an invalid Bearer value")
)

//UserToken represents an authenticated user
type UserToken struct {
	//authenticated user name
	IAM string `json:"iam"`

	//unix epoc expire time
	EXP int64 `json:"exp"`
}

//GenUserToken will generate an expiring HMAC-SHA256 signed token representing an authenticated user
func GenUserToken(user string, expires time.Duration, signingKey []byte) (UserToken, string, error) {
	var u UserToken
	u.IAM = user
	u.EXP = time.Now().Add(expires).Unix()
	data, _ := json.Marshal(u)
	tok, err := jose.Sign(string(data), jose.HS256, signingKey)
	return u, tok, err
}

//ValidateUserToken will validate the token and attempt to unmarshal it's payload into a UserToken.
//
//error is nil if validation succeeded.
func ValidateUserToken(token string, signingKey []byte) (UserToken, error) {
	var u UserToken
	payload, _, err := jose.Decode(token, signingKey)
	if err != nil {
		return u, err
	}

	err = json.Unmarshal([]byte(payload), &u)
	if err != nil {
		return u, err
	}

	if u.IAM == "" || u.EXP == 0 {
		return u, ErrBadPayload
	}

	if u.EXP < time.Now().Unix() {
		return u, ErrTokenExpired
	}

	return u, nil
}
