package auth

import (
	"net/http"

	"github.com/hfurubotten/autograder/web/sessions"
	"github.com/hfurubotten/autograder/web/pages"
)

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

func RemoveApproval(w http.ResponseWriter, r *http.Request) {
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, false)
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, "")
}

func RemoveApprovalHandler(w http.ResponseWriter, r *http.Request) {
	sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, false)
	sessions.SetSessionsAndRedirect(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, "", pages.FRONTPAGE)
}