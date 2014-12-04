package auth

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// sets up the github as the oauth provider. To get the variables and functions loaded into the standard that is used, use the init method. This will set this as soon as the package is loaded the first time. Replace or comment out the init method to use another oath provider.
func init() {
	Clientid = global.OAuth_ClientID         //"2e2c5b20f954de037b8f"
	clientsecret = global.OAuth_ClientSecret //"f69a12873ea33f365523b3b5adb040e443df48ae"
	Scope = "user,admin:org,repo,admin:repo_hook"
	RedirectURL = "https://github.com/login/oauth/authorize"

	Handler = github_oauthhandler
}

func github_oauthhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getvalues := r.URL.Query()

		code := getvalues.Get("code")
		errstr := getvalues.Get("error")

		if len(errstr) > 0 {
			log.Println("OAuth error: " + errstr)
			// redirect to home page
			return
		}

		postdata := []byte("client_id=" + Clientid + "&client_secret=" + clientsecret + "&code=" + code)
		requrl := "https://github.com/login/oauth/access_token"
		req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(postdata))
		if err != nil {
			log.Println("Echange error with github: ", err)
			// Do something to redirect or tell user of error
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Echange error with github: ", err)
			// Do something to redirect or tell user of error
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read error: ", err)
			// Do something to redirect or tell user of error
			return
		}

		q, err := url.ParseQuery(string(data))
		if err != nil {
			log.Println("Data error from github: ", err)
			// Do something to redirect or tell user of error
			return
		}

		access_token := q.Get("access_token")
		errstr = q.Get("error")
		approved := false

		if len(errstr) > 0 {
			log.Println("Access token error: " + errstr)
			// redirect to home page
			return
		} else {
			approved = true
		}

		sessions.SetSessions(w, r, sessions.AUTHSESSION, sessions.APPROVEDSESSIONKEY, approved)
		sessions.SetSessionsAndRedirect(w, r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY, access_token, pages.HOMEPAGE)
	} else {
		redirect := http.RedirectHandler(pages.FRONTPAGE, 400)
		redirect.ServeHTTP(w, r)
	}
}
