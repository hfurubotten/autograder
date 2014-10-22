package auth

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/sessions"
)

// sets up the github as the oauth provider. To get the variables and functions loaded into the standard that is used, use the init method. This will set this as soon as the package is loaded the first time. Replace or comment out the init method to use another oath provider. 
func init(){
	Clientid = "2e2c5b20f954de037b8f"
	clientsecret = "f69a12873ea33f365523b3b5adb040e443df48ae"
	Scope = "user"
	RedirectURL = "https://github.com/login/oauth/authorize"

	Handler = github_oauthhandler
}

func github_oauthhandler(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET" {
		getvalues := r.URL.Query()

		code := getvalues.Get("code")

		postdata := []byte("client_id=" + Clientid + "&client_secret=" + clientsecret + "&code=" + code)
		requrl := "https://github.com/login/oauth/access_token"
		req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(postdata))
		if err != nil {
			log.Println(err)
			// Do something to redirect or tell user of error
			return
		}
    	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			// Do something to redirect or tell user of error
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			// Do something to redirect or tell user of error
			return
		}

		q, err := url.ParseQuery(string(data))
		if err != nil {
			log.Println(err)
			// Do something to redirect or tell user of error
			return
		}

		member := git.NewMember(q.Get("access_token"))
		member.LoadRemoteData()

		log.Println("Logged in:", member.Username)

		sessions.SetSessionsAndRedirect(w, r, sessions.AUTHSESSION, "user", member, "http://tussi.hf-data.no/session")
	} else {
		redirect := http.RedirectHandler("/", 400)
		redirect.ServeHTTP(w, r)
	}
}