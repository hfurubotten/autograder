package git

import (
	"crypto/sha256"
	"fmt"

	"github.com/hfurubotten/diskv"
)

var tokenstore = diskv.New(diskv.Options{
	BasePath:     "diskv/tokens",
	CacheSizeMax: 1024 * 1024 * 256,
})

type token struct {
	accessToken string
}

func NewToken(oauthtoken string) token {
	return token{oauthtoken}
}

func (m *token) HasTokenInStore() bool {
	hash := sha256.Sum256([]byte(m.accessToken))
	return tokenstore.Has(fmt.Sprintf("%x", hash))
}

func (m *token) GetUsernameFromTokenInStore() (user string, err error) {
	hash := sha256.Sum256([]byte(m.accessToken))
	err = tokenstore.ReadGob(fmt.Sprintf("%x", hash), &user, false)
	return user, err
}

func (m *token) SetUsernameToTokenInStore(username string) (err error) {
	hash := sha256.Sum256([]byte(m.accessToken))
	err = tokenstore.WriteGob(fmt.Sprintf("%x", hash), username)
	return
}

func (m *token) RemoveTokenInStore() (err error) {
	return tokenstore.Erase(m.accessToken)
}

func (t token) HasToken() bool {
	return t.accessToken != ""
}

func (t token) GetToken() string {
	return t.accessToken
}
