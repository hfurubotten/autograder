package git

import (
	"log"
	"encoding/gob"

	"github.com/google/go-github/github"
	"code.google.com/p/goauth2/oauth"
)

func init() {
	gob.Register(member{})
}

type member struct {
	githubclient 	*github.Client
	Username string
	Name string
	AccessToken string
}

func NewMember(oauthtoken string) member {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: oauthtoken},
	}
	g := github.NewClient(t.Client())
	return member{githubclient: g, AccessToken: oauthtoken,}
}

func (m *member) LoadRemoteData() (err error) {
	user, _, err := m.githubclient.Users.Get("")
	if err != nil {
		log.Println(err)
		return
	}

	if user.Login != nil {
		m.Username = *user.Login
	}

	if user.Name != nil {
		m.Name = *user.Name
	}

	return
}