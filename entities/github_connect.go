package entities

import (
	"errors"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	// ErrNoAccessToken indicates that a empty access token was provided
	ErrNoAccessToken = errors.New("non-empty OAuth access token required")
	// ErrNotConnected indicates that the member object is not connected to github
	ErrNotConnected = errors.New("member object not connected to github")
)

// GitHubConn contains connection details for GitHub.
type GitHubConn struct {
	// remote access (private fields will not be stored in the database)
	client      *github.Client
	accessToken string
	Scope       string
}

func connect(token string) (*github.Client, error) {
	if token == "" {
		return nil, ErrNoAccessToken
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc), nil
}

// getGithubUser returns the github user associated with the provided client.
// The function assumes that the client was created with the appropriate OAuth
// token, for example using the connect function above.
func getGithubUser(client *github.Client) (user *github.User, err error) {
	if client == nil {
		return nil, ErrNotConnected
	}
	user, _, err = client.Users.Get("")
	return
}

// ListOrgs returns a list github organizations that the user is member of.
func (m *Member) ListOrgs() (ls []string, err error) {
	err = m.Dial()
	if err != nil {
		return nil, err
	}
	orgs, _, err := m.client.Organizations.List("", nil)
	ls = make([]string, len(orgs))
	for i, org := range orgs {
		ls[i] = *org.Login
	}
	return
}
