package auth

import "net/http"

var (
	Clientid = ""
	clientsecret = ""

	Scope = ""
	State = ""

	RedirectURL = ""

	Handler = func (w http.ResponseWriter, r *http.Request) {
		// Empty placeholder
	}
)