package sessions

import(
	"net/http"
	"errors"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func init() {
	store.Options = &sessions.Options{
		Path: "sessions/",
		MaxAge: 86400,
	}
}

func SetSessionsAndRedirect(w http.ResponseWriter, r *http.Request, sessionsname string, key interface{}, value interface{}, redirecturl string) (err error) {
	session, _ := store.Get(r, sessionsname)

	session.Values[key] = value
	err = session.Save(r, w)

	handler := http.RedirectHandler(redirecturl, 307)
	handler.ServeHTTP(w, r)

	return
}

func SetSessions(w http.ResponseWriter, r *http.Request, sessionsname string, key interface{}, value interface{}) (err error) {
	session, _ := store.Get(r, sessionsname)

	session.Values[key] = value
	err = session.Save(r, w)

	return 
}

func GetSessions(r *http.Request, sessionsname string, key interface{}) (interface{}, error) {
	session, _ := store.Get(r, sessionsname)

	if val, ok := session.Values[key]; ok {
		return val, nil
	} else {
		return nil, errors.New("Couldn't find the key in that session.")
	}
}
