package auth

import (
	"net/http"

	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// IsApprovedUser checks if the user is logged in and approved.
func IsApprovedUser(r *http.Request) (approved bool) {
	val, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY)
	if err != nil {
		return false
	}

	switch val.(type) {
	case bool:
		approved = val.(bool)
	default:
		return false
	}

	return
}

// RemoveApproval will revoke the login approval in the sessions of a user.
func RemoveApproval(w http.ResponseWriter, r *http.Request) {
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, false)
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, "")
}

// RemoveApprovalHandler is a http handler which will revoke the login
// approval in the session of the user and then redirect to the front page.
func RemoveApprovalHandler(w http.ResponseWriter, r *http.Request) {
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, false)
	sessions.SetSessionsAndRedirect(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, "", pages.FRONTPAGE)
}
