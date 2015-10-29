package entities

import (
	"crypto/sha256"
	"fmt"

	"github.com/hfurubotten/autograder/database"
)

// TokenBucketName is the bucket name for tokens in the database.
var TokenBucketName = "tokens"

func init() {
	database.RegisterBucket(TokenBucketName)
}

// hasToken returns true if the given token is in the database.
// Tokens are hashed before they are stored in the database to protect the on
// disk version of the tokens.
func hasToken(token string) bool {
	hash := sha256.Sum256([]byte(token))
	return database.Has(TokenBucketName, fmt.Sprintf("%x", hash))
}

// getToken returns the user name associated with the given token, if the given
// token exists in the database. An error is returned if the token is not
// present in the database, or the database operation failed.
// Tokens are hashed before they are stored in the database to protect the on
// disk version of the tokens.
func getToken(token string) (user string, err error) {
	if token == "" {
		// empty tokens are not allowed
		return "", ErrNoAccessToken
	}
	hash := sha256.Sum256([]byte(token))
	err = database.Get(TokenBucketName, fmt.Sprintf("%x", hash), &user)
	return user, err
}

// putToken stores the association between the given token and user name in the
// database. An error is returned if the database operation failed.
// Tokens are hashed before they are stored in the database to protect the on
// disk version of the tokens.
func putToken(token, user string) error {
	if token == "" {
		// empty tokens are not allowed
		return ErrNoAccessToken
	}
	hash := sha256.Sum256([]byte(token))
	return database.Put(TokenBucketName, fmt.Sprintf("%x", hash), user)
}

// removeToken removes the given token from the database. An error is returned
// if the database operation failed.
// Tokens are hashed before they are stored in the database to protect the on
// disk version of the tokens.
func removeToken(token string) error {
	if token == "" {
		// empty tokens are not allowed
		return ErrNoAccessToken
	}
	hash := sha256.Sum256([]byte(token))
	return database.Remove(TokenBucketName, fmt.Sprintf("%x", hash))
}
