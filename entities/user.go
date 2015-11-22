package entities

import (
	"net/mail"
	"sync"
)

// UserProfile contains information about a user. Often this information will
// be gleaned from external sources such as GitHub.
type UserProfile struct {
	mainlock sync.Mutex

	*GitHubConn

	Name          string
	Username      string
	Email         *mail.Address
	Location      string
	Active        bool
	PublicProfile bool
	AvatarURL     string
	ProfileURL    string
}

// CreateUserProfile returns a new UserProfile populated with data from github.
func CreateUserProfile(userName string) *UserProfile {
	return &UserProfile{
		Username: userName,
		GitHubConn: &GitHubConn{
			accessToken: "",
		},
	}
}

// NewUserProfile returns a new UserProfile.
func NewUserProfile(token, user, scope string) *UserProfile {
	return &UserProfile{
		Username: user,
		GitHubConn: &GitHubConn{
			Scope:       scope,
			accessToken: token,
		},
	}
}

// Dial connects to GitHub.
func (u *UserProfile) Dial() (err error) {
	if u.client == nil {
		u.client, err = connect(u.accessToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UserProfile) hasAccessToken() bool {
	return u.accessToken != "" && len(u.accessToken) > 0
}

// GetUsername will return the users unique username.
func (u *UserProfile) GetUsername() string {
	return u.Username
}

// Activate sets the user as active.
func (u *UserProfile) Activate() {
	u.Active = true
}

// IsActive returns whether or not the user is active.
func (u *UserProfile) IsActive() bool {
	return u.Active
}

// Deactivate sets the user as deactivated.
func (u *UserProfile) Deactivate() {
	u.Active = false
}

// SetPublicProfile sets if the profile should be open
// to thepublic to search through.
func (u *UserProfile) SetPublicProfile(public bool) {
	u.PublicProfile = public
}

// Lock will lock the user name from being written to by
// other instances of the same organization. This has to be used
// when new info is written, to prevent race conditions. Unlock
// occures when data is finished written to storage.
func (u *UserProfile) Lock() {
	u.mainlock.Lock()
}

// Unlock will unlock the writers block on the user.
func (u *UserProfile) Unlock() {
	u.mainlock.Unlock()
}
