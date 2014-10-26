package pages

import "net/http"

const (
	OAUTH_REDIRECT = "/oauth"
	FRONTPAGE = "/"
	REGISTER_REDIRECT = "/profile"
	HOMEPAGE = "/home"
	SIGNOUT = "/logout"
)

func RedirectTo(w http.ResponseWriter, r *http.Request, page string, status int) {
	handler := http.RedirectHandler(page, status)
	handler.ServeHTTP(w, r)
}