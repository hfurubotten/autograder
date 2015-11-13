package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hfurubotten/autograder/config"
)

func TestOAuthScopeRedirectURL(t *testing.T) {
	c, err := config.NewConfig(
		"http://thunder.ux.uis.no",
		"3993eoijW-randomID",
		"38938djbd39cn93n3gdsierjds-secret-ndkj231h",
		"/tmp/ag",
	)
	if err != nil {
		t.Error("Failed to create new configuration object")
	}
	c.SetCurrent()
	// e := "https://github.com/login/oauth/authorize?client_id=3993eoijW-randomID&scope=admin:org,repo,admin:repo_hook"
	e := "https://github.com/login/oauth/authorize?client_id=3993eoijW-randomID&scope=admin%3Aorg%2Crepo%2Cadmin%3Arepo_hook"
	s := OAuthScopeRedirectURL()
	if e != s {
		t.Errorf("\nexpected %s\ngot      %s\n", e, s)
	}
	// doRedirect(s, t)
}

func TestOAuthRedirectURL(t *testing.T) {
	e := "https://github.com/login/oauth/authorize?client_id=3993eoijW-randomID"
	s := OAuthRedirectURL()
	if e != s {
		t.Errorf("\nexpected %s\ngot      %s\n", e, s)
	}
}

func doRedirect(s string, t *testing.T) {
	rdir := http.RedirectHandler(s, http.StatusTemporaryRedirect)
	ts := httptest.NewServer(rdir)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("http get returned status code: %d, expected: %d",
			res.StatusCode, http.StatusOK)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", body)
}
