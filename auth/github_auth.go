package auth

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hfurubotten/autograder/config"
	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

//TODO do this in the config start up procedure ?? These are github consts??
// Sets up github as the OAuth provider.
// To get the variables and functions loaded into the standard that is used,
// use the init method. This will set this as soon as the package is loaded
// the first time. Replace or comment out the init method to use another OAuth provider.
func init() {
	global.OAuthScope = "admin:org,repo,admin:repo_hook"
	global.OAuthRedirectURL = "https://github.com/login/oauth/authorize"

	global.OAuthHandler = githubOauthHandler
}

func githubOauthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		clientID := config.Get().OAuthClientID
		clientSecret := config.Get().OAuthClientSecret

		getvalues := r.URL.Query()

		code := getvalues.Get("code")
		errstr := getvalues.Get("error")

		if len(errstr) > 0 {
			log.Println("OAuth error: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		postdata := []byte("client_id=" + clientID + "&client_secret=" + clientSecret + "&code=" + code)
		//TODO github const?
		requrl := "https://github.com/login/oauth/access_token"
		req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(postdata))
		if err != nil {
			log.Println("Exchange error with github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Exchange error with github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read error: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		q, err := url.ParseQuery(string(data))
		if err != nil {
			log.Println("Data error from github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		accessToken := q.Get(sessions.AccessTokenSessionKey)
		errstr = q.Get("error")
		approved := false

		if len(errstr) > 0 {
			log.Println("Access token error: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		approved = true

		scope := q.Get("scope")

		if scope != "" {
			//TODO This should probably be LookupMember() (but in a Update() transaction, since we are updating the scope.)
			m, err := git.NewMember(accessToken)
			if err != nil {
				log.Println("Could not open Member object:", err)
				http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
				return
			}

			m.Scope = scope
			err = m.Save()
			if err != nil {
				m.Unlock()
			}
		}

		sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, approved)
		sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.HOMEPAGE)
	} else {
		//TODO Check and understand if this should be StatusTemporaryRedirect??
		http.Redirect(w, r, pages.FRONTPAGE, http.StatusBadRequest)
	}
}
