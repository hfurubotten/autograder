package global

import (
	"net/http"
)

var (
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
)
