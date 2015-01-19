package auth

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// sets up the github as the oauth provider. To get the variables and functions loaded into the standard that is used, use the init method. This will set this as soon as the package is loaded the first time. Replace or comment out the init method to use another oath provider.
func init() {
	global.OAuth_Scope = "admin:org,repo,admin:repo_hook"
	global.OAuth_RedirectURL = "https://github.com/login/oauth/authorize"

	global.OAuth_Handler = github_oauthhandler
}

func github_oauthhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		Clientid := global.OAuth_ClientID
		clientsecret := global.OAuth_ClientSecret

		getvalues := r.URL.Query()

		code := getvalues.Get("code")
		errstr := getvalues.Get("error")

		if len(errstr) > 0 {
			log.Println("OAuth error: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		postdata := []byte("client_id=" + Clientid + "&client_secret=" + clientsecret + "&code=" + code)
		requrl := "https://github.com/login/oauth/access_token"
		req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(postdata))
		if err != nil {
			log.Println("Echange error with github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Echange error with github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read error: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		q, err := url.ParseQuery(string(data))
		if err != nil {
			log.Println("Data error from github: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		access_token := q.Get("access_token")
		errstr = q.Get("error")
		approved := false

		if len(errstr) > 0 {
			log.Println("Access token error: " + errstr)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		} else {
			approved = true
		}

		scope := q.Get("scope")

		if scope != "" {
			m := git.NewMember(access_token)
			m.Scope = scope
			m.StickToSystem()
		}

		sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, approved)
		sessions.SetSessionsAndRedirect(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, access_token, pages.HOMEPAGE)
	} else {
		redirect := http.RedirectHandler(pages.FRONTPAGE, 400)
		redirect.ServeHTTP(w, r)
	}
}
