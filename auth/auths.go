package auth

import (
	"log"
	"net/http"

	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// IsApprovedUser checks if the user is logged in and approved.
func IsApprovedUser(r *http.Request) bool {
	// val, err := sessions.GetSessions(r, sessions.AuthSession, sessions.ApprovedSessionKey)
	val, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
	if err != nil {
		return false
	}
	log.Println("checking for approved user session: ", val)
	switch val.(type) {
	case string:
		token := val.(string)
		return entities.HasToken(token)
	default:
		return false
	}

	// switch val.(type) {
	// case bool:
	// 	approved = val.(bool)
	// default:
	// 	return false
	// }

	// return
}

// RemoveApprovalHandler is a http handler which will revoke the login
// approval in the session of the user and then redirect to the front page.
func RemoveApprovalHandler(w http.ResponseWriter, r *http.Request) {
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, false)
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, "")
	http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
}
