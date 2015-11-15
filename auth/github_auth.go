package auth

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hfurubotten/autograder/config"
	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// Constants and methods for using GitHub as OAuth provider for Autograder.

const (
	// tokenURL is used to fetch the access token for a user.
	tokenURL = "https://github.com/login/oauth/access_token"
	// oauthRedirectURL is used for redirecting users to the GitHub login page.
	oauthRedirectURL = "https://github.com/login/oauth/authorize"
	// oauthScope defines the scope necessary for teacher access to GitHub.
	oauthScope = "admin:org,repo,admin:repo_hook"
)

// OAuthScopeRedirectURL returns a URL for redirecting to obtain a new scope.
// This is used to authenticate as teacher.
func OAuthScopeRedirectURL() string {
	u, err := url.Parse(oauthRedirectURL)
	if err != nil {
		// the redirection URL must be a valid URL
		panic(err)
	}
	values := u.Query()
	values.Set("client_id", config.Get().OAuthClientID)
	values.Set("scope", oauthScope)
	u.RawQuery = values.Encode()
	return u.String()
	// return oauthRedirectURL + "?client_id=" + config.Get().OAuthClientID + "&scope=" + oauthScope
}

// OAuthRedirectURL returns a URL for redirecting to login page.
func OAuthRedirectURL() string {
	u, err := url.Parse(oauthRedirectURL)
	if err != nil {
		// the redirection URL must be a valid URL
		panic(err)
	}
	values := u.Query()
	values.Set("client_id", config.Get().OAuthClientID)
	u.RawQuery = values.Encode()
	return u.String()
	// return oauthRedirectURL+"?client_id="+config.Get().OAuthClientID
}

// OAuthHandler is the OAuth handler for the GitHub.
func OAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getValues := r.URL.Query()
		code := getValues.Get("code")
		errstr := getValues.Get("error")
		if len(errstr) > 0 {
			log.Println("Failed to obtain temporary OAuth code: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		postValues := url.Values{}
		postValues.Set("client_id", config.Get().OAuthClientID)
		postValues.Set("client_secret", config.Get().OAuthClientSecret)
		postValues.Set("code", code)
		s := postValues.Encode()
		req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer([]byte(s)))
		if err != nil {
			log.Println("Failed to create POST request: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Failed to issue POST request: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		q, err := url.ParseQuery(string(data))
		if err != nil {
			log.Println("Failed to parse query data: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		accessToken := q.Get(sessions.AccessTokenSessionKey)
		errstr = q.Get("error")
		if len(errstr) > 0 {
			log.Println("Failed to obtain access token: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
			return
		}

		scope := q.Get("scope")
		if scope != "" {
			// TODO Consider if updating scope needs to be in a transaction using the Update() function.
			// check if access token is associated with existing member
			m, err := entities.LookupMember(accessToken)
			if err != nil {
				// access token is not in the database; must be a new member
				u, err := githubUserProfile(accessToken, scope)
				if err != nil {
					log.Printf("Failed to get user from GitHub: %v", err)
					http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
					return
				}
				m = entities.NewMember(u)
				err = entities.PutMember(accessToken, m)
				if err != nil {
					log.Printf("Failed to create member with token: %s\n%v", accessToken, err)
					http.Redirect(w, r, pages.FRONTPAGE, http.StatusTemporaryRedirect)
					return
				}
				log.Printf("Created new member %s", m.Name)
			}
			log.Printf("Current scope (%s) for %s", m.Scope, m.Name)
			if m.Scope != scope {
				m.Scope = scope
				err = m.Save()
				if err != nil {
					log.Printf("Failed to update scope (%s) for %s:\n%v", scope, m.Name, err)
				} else {
					log.Printf("Successfully updated scope (%s) for %s", scope, m.Name)
				}
			}
		}

		// mark auth session as approved
		sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
		// save the access token for this session
		sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.HOMEPAGE)
	} else {
		// was not a GET request method; redirect with a bad request status.
		log.Println("Bad request: ", r)
		http.Redirect(w, r, pages.FRONTPAGE, http.StatusBadRequest)
	}
}
