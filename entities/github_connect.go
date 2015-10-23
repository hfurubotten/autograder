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

// NewUserWithGithubData creates a new User object from a github User object.
// It will copy all information from the given GitHub data to the new User object.
func NewUserWithGithubData(gu *github.User) (u *Member, err error) {
	if gu == nil {
		return nil, errors.New("github user object is required")
	}
	u, err = GetMember(*gu.Login) //TODO Need to pass in token also??
	if err != nil {
		return nil, err
	}

	u.ImportGithubData(gu)

	return
}

func connect(token string) (*github.Client, error) {
	if token == "" {
		return nil, ErrNoAccessToken
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc), nil
}

// connectToGithub creates a new github client.
//TODO CUrrently not used
func (m *Member) xconnectToGithub() error {
	if m.githubclient != nil {
		return nil
	}
	if !m.hasAccessToken() {
		return errors.New("unable to connect to github; missing access token for " + m.Username)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: m.accessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	m.githubclient = github.NewClient(tc)
	return nil
}

func (m *Member) loadDataFromGithub() (err error) {
	if m.githubclient == nil {
		return ErrNotConnected
	}
	// err = m.connectToGithub()
	// if err != nil {
	// 	return
	// }

	user, _, err := m.githubclient.Users.Get("")
	if err != nil {
		return
	}
	if user.Login != nil {
		m.Username = *user.Login
	}
	m.ImportGithubData(user)
	return
}

// ListOrgs returns a list github organizations that the user is member of.
func (m *Member) ListOrgs() (ls []string, err error) {
	if m.githubclient == nil {
		return nil, ErrNotConnected
	}
	// err = m.connectToGithub()
	// if err != nil {
	// 	return
	// }

	orgs, _, err := m.githubclient.Organizations.List("", nil)
	ls = make([]string, len(orgs))
	for i, org := range orgs {
		ls[i] = *org.Login
	}
	return
}
