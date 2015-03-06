package git

import (
	"crypto/sha256"
	"fmt"

	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

// Token represents a access token retrieved in the oauth process.
type Token struct {
	accessToken string
}

// NewToken returns a new token created from a oauth token.
func NewToken(oauthtoken string) Token {
	return Token{oauthtoken}
}

// HasTokenInStore checks if the token is in storage.
func (m *Token) HasTokenInStore() bool {
	hash := sha256.Sum256([]byte(m.accessToken))
	return getTokenStore().Has(fmt.Sprintf("%x", hash))
}

// GetUsernameFromTokenInStore gets the username associated with the token.
func (m *Token) GetUsernameFromTokenInStore() (user string, err error) {
	hash := sha256.Sum256([]byte(m.accessToken))
	err = getTokenStore().ReadGob(fmt.Sprintf("%x", hash), &user, false)
	return user, err
}

// SetUsernameToTokenInStore sets the username associated with the token.
func (m *Token) SetUsernameToTokenInStore(username string) (err error) {
	hash := sha256.Sum256([]byte(m.accessToken))
	err = getTokenStore().WriteGob(fmt.Sprintf("%x", hash), username)
	return
}

// RemoveTokenInStore removed the token from storage.
func (m *Token) RemoveTokenInStore() (err error) {
	return getTokenStore().Erase(m.accessToken)
}

// HasToken checks if the token is set.
func (t *Token) HasToken() bool {
	return t.accessToken != ""
}

// GetToken returns the plain token string.
func (t *Token) GetToken() string {
	return t.accessToken
}

var tokenstore *diskv.Diskv

func getTokenStore() *diskv.Diskv {
	if tokenstore == nil {
		tokenstore = diskv.New(diskv.Options{
			BasePath:     global.Basepath + "diskv/tokens",
			CacheSizeMax: 1024 * 1024 * 64,
		})
	}
	return tokenstore
}
