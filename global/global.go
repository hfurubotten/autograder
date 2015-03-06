package global

import (
	"net/http"
)

var (
	// Hostname globally stores the hostname autograder is running under.
	Hostname string

	// OAuthClientID globally stores the oauth client id used up to github.
	OAuthClientID string

	// OAuthClientSecret globally stores the secret oauth code used up to github.
	OAuthClientSecret string

	// OAuthScope globally stores the oauth scope needed to use autograder against github.
	OAuthScope string

	// OAuthState globally stores the oauth state.
	OAuthState string

	// OAuthRedirectURL globally stores the url the oauth process should redirect back to.
	OAuthRedirectURL string

	// OAuthHandler globally stores a http handler for processing a oauth request.
	OAuthHandler = func(w http.ResponseWriter, r *http.Request) {
		// Empty placeholder
	}

	// Basepath globally stores the basepath for the code directory.
	Basepath string
)
