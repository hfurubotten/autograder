package sessions

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func init() {
	store.Options = &sessions.Options{
		MaxAge: 86400,
	}
}

// SetSessionsAndRedirect will set a session and redirects.
func SetSessionsAndRedirect(w http.ResponseWriter, r *http.Request, sessionsname string, key interface{}, value interface{}, redirecturl string) (err error) {
	session, _ := store.Get(r, sessionsname)

	session.Values[key] = value
	err = session.Save(r, w)

	handler := http.RedirectHandler(redirecturl, 307)
	handler.ServeHTTP(w, r)

	return
}

// SetSessions sets a session.
func SetSessions(w http.ResponseWriter, r *http.Request, sessionsname string, key interface{}, value interface{}) (err error) {
	session, _ := store.Get(r, sessionsname)

	session.Values[key] = value
	err = session.Save(r, w)

	return
}

// GetSessions gets a certain session.
func GetSessions(r *http.Request, sessionsname string, key interface{}) (interface{}, error) {
	session, _ := store.Get(r, sessionsname)

	if val, ok := session.Values[key]; ok {
		return val, nil
	}

	return nil, errors.New("Couldn't find the key in that session.")
}
