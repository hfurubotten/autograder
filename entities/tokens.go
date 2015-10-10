package entities

import "github.com/hfurubotten/autograder/database"

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
	return database.Has(TokenBucketName, t.accessToken)
}

// GetUsernameFromTokenInStore gets the username associated with the token.
func (t *Token) GetUsernameFromTokenInStore() (user string, err error) {
	err = database.Get(TokenBucketName, t.accessToken, &user)
	return user, err
}

// SetUsernameToTokenInStore sets the username associated with the token.
func (t *Token) SetUsernameToTokenInStore(username string) (err error) {
	return database.Put(TokenBucketName, t.accessToken, username)
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
