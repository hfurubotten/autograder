package git

import (
	"crypto/sha256"
	"fmt"

	"github.com/hfurubotten/autograder/database"
)

// TokenBucketName is the bucket/table name for tokens in the database.
var TokenBucketName = "tokens"

func init() {
	database.RegisterBucket(TokenBucketName)
}

// Token represents a access token retrieved in the oauth process.
type Token struct {
	accessToken string
}

// NewToken returns a new token created from a oauth token.
func NewToken(oauthtoken string) Token {
	return Token{oauthtoken}
}

// HasTokenInStore checks if the token is in storage.
func (t *Token) HasTokenInStore() bool {
	hash := sha256.Sum256([]byte(t.accessToken))
	val := ""
	err := database.Get(TokenBucketName, fmt.Sprintf("%x", hash), val, true)
	return err == nil && val != ""
}

// GetUsernameFromTokenInStore gets the username associated with the token.
func (t *Token) GetUsernameFromTokenInStore() (user string, err error) {
	hash := sha256.Sum256([]byte(t.accessToken))
	err = database.Get(TokenBucketName, fmt.Sprintf("%x", hash), &user, true)

	return user, err
}

// SetUsernameToTokenInStore sets the username associated with the token.
func (t *Token) SetUsernameToTokenInStore(username string) (err error) {
	hash := sha256.Sum256([]byte(t.accessToken))
	err = database.Store(TokenBucketName, fmt.Sprintf("%x", hash), username)
	return
}

// RemoveTokenInStore removed the token from storage.
func (t *Token) RemoveTokenInStore() (err error) {
	return database.Remove(TokenBucketName, t.accessToken)
}

// HasToken checks if the token is set.
func (t *Token) HasToken() bool {
	return t.accessToken != ""
}

// GetToken returns the plain token string.
func (t *Token) GetToken() string {
	return t.accessToken
}
