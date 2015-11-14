package auth

import (
	"errors"
	"net/mail"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/entities"
	"golang.org/x/oauth2"
)

var (
	// ErrNoAccessToken indicates that a empty access token was provided
	ErrNoAccessToken = errors.New("non-empty OAuth access token required")
)

// Connect returns a github client object.
func connect(token string) (*github.Client, error) {
	if token == "" {
		return nil, ErrNoAccessToken
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc), nil
}

// githubUserProfile returns a new UserProfile populated with data from github.
func githubUserProfile(token, scope string) (*entities.UserProfile, error) {
	c, err := connect(token)
	if err != nil {
		return nil, err
	}
	gu, _, err := c.Users.Get("")
	if err != nil {
		return nil, err
	}
	if gu.Login == nil {
		return nil, errors.New("missing login name for github account")
	}
	u := entities.NeUserProfile(token, *gu.Login, scope)
	if gu.Name != nil {
		u.Name = *gu.Name
	}
	if gu.AvatarURL != nil {
		u.AvatarURL = *gu.AvatarURL
	}
	if gu.HTMLURL != nil {
		u.ProfileURL = *gu.HTMLURL
	}
	if gu.Location != nil {
		u.Location = *gu.Location
	}
	if gu.Email != nil {
		m, err := mail.ParseAddress(*gu.Email)
		if err == nil {
			u.Email = m
		}
	}
	return u, nil
}
