package entities

import (
	"encoding/gob"
	"net/mail"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/game/levels"
	"github.com/hfurubotten/autograder/game/trophies"
)

func init() {
	gob.Register(User{})
}

type User struct {
	lock     sync.RWMutex
	mainlock sync.Mutex

	Name          string
	Username      string
	Email         *mail.Address
	Location      string
	Active        bool
	PublicProfile bool

	// Scores
	TotalScore   int64
	WeeklyScore  map[int]int64
	MonthlyScore map[time.Month]int64
	Level        int
	Trophies     *trophies.TrophyChest

	// URLs
	AvatarURL  string
	ProfileURL string

	// remote access
	githubclient *github.Client
	accessToken  string
	Scope        string
}

// NewUser tries to find the user in storage with the username
// and loads this on success. If no user with the given
// login name is found it will give a blank User object.
func NewUser(login string) (u *User, err error) {
	u = new(User)
	u.Username = login

	err = u.loadStoredData()
	if err != nil {
		return nil, err
	}

	if u.WeeklyScore == nil {
		u.WeeklyScore = make(map[int]int64)
	}

	if u.MonthlyScore == nil {
		u.MonthlyScore = make(map[time.Month]int64)
	}

	return
}

// NewUserWithGithubAccessToken will attemt to find the owner
// of the access token in storage. If the owner is found,
// it returns the User object which is owner. If not found,
// it loads user data from Github and makes a new User
// from this.
func NewUserWithGithubAccessToken(token string) (u *User, err error) {
	u = new(User)

	if hasToken(token) {
		u.Username, err = getToken(token)
		if err != nil {
			return nil, err
		}

		err = u.loadStoredData()
		if err != nil {
			return
		}
	} else {
		gu, err := u.loadDataFromGithub()
		if err != nil {
			return u, err
		}
		u.Username = *gu.Login

		err = u.loadStoredData()
		if err != nil {
			return u, err
		}

		u.ImportGithubData(gu)
	}

	if u.WeeklyScore == nil {
		u.WeeklyScore = make(map[int]int64)
	}

	if u.MonthlyScore == nil {
		u.MonthlyScore = make(map[time.Month]int64)
	}

	return
}

// SetName will set the name of the user.
func (u *User) SetName(name string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Name = name
}

// SetEmail will set the email of the user.
func (u *User) SetEmail(email *mail.Address) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Email = email
}

// SetLocation will set the location of the user.
func (u *User) SetLocation(location string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Location = location
}

// SetScope will set the scope of the user.
func (u *User) SetScope(scope string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Scope = scope
}

// IncScoreBy increases the total score with given amount.
func (u *User) IncScoreBy(score int) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.TotalScore += int64(score)
	u.Level = levels.FindLevel(u.TotalScore) // How to tackle level up notification?

	_, week := time.Now().ISOWeek()
	month := time.Now().Month()

	// updates weekly
	u.WeeklyScore[week] += int64(score)
	// updated monthly
	u.MonthlyScore[month] += int64(score)
}

// DecScoreBy descreases the total score with given amount.
func (u *User) DecScoreBy(score int) {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.TotalScore-int64(score) > 0 {
		u.TotalScore -= int64(score)
	} else {
		u.TotalScore = 0
	}

	u.Level = levels.FindLevel(u.TotalScore)

	_, week := time.Now().ISOWeek()
	month := time.Now().Month()

	// updates weekly
	u.WeeklyScore[week] -= int64(score)
	// updated monthly
	u.MonthlyScore[month] -= int64(score)
}

// IncLevel increases the level with one.
func (u *User) IncLevel() {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Level++
}

// DecLevel decreases the level with one until it equals zero.
func (u *User) DecLevel() {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.Level > 0 {
		u.Level--
	}
}

// GetTrophyChest return the users ThropyChest.
func (u *User) GetTrophyChest() *trophies.TrophyChest {
	if u.Trophies == nil {
		u.Trophies = trophies.NewTrophyChest()
	}

	return u.Trophies
}

// GetUsername will return the users unique username.
func (u *User) GetUsername() string {
	return u.Username
}

// Activate sets the user as active.
func (u *User) Activate() {
	u.Active = true
}

// IsActive returns whether or not the user is active.
func (u *User) IsActive() bool {
	return u.Active
}

// Deactivate sets the user as deactivated.
func (u *User) Deactivate() {
	u.Active = false
}

// SetPublicProfile sets if the profile should be open
// to thepublic to search through.
func (u *User) SetPublicProfile(public bool) {
	u.PublicProfile = public
}

// ImportGithubData imports data from the given github
// data object and stores it in the given User object.
func (u *User) ImportGithubData(gu *github.User) {
	if gu == nil {
		return
	}

	if gu.AvatarURL != nil {
		u.AvatarURL = *gu.AvatarURL
	}

	if gu.HTMLURL != nil {
		u.ProfileURL = *gu.HTMLURL
	}

	if gu.Name != nil {
		u.Name = *gu.Name
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

}

// loadStoredData fetches the user data stored on disk or in cached memory.
// ATM a NO-OP
func (u *User) loadStoredData() (err error) {
	return nil
}

// loadDataFromGithub attempts to load user data from
// github and sets data from there in the user object.
func (u *User) loadDataFromGithub() (user *github.User, err error) {
	err = u.connectToGithub()
	if err != nil {
		return
	}

	user, _, err = u.githubclient.Users.Get("")
	return
}

// Lock will lock the user name from being written to by
// other instances of the same organization. This has to be used
// when new info is written, to prevent race conditions. Unlock
// occures when data is finished written to storage.
func (u *User) Lock() {
	u.mainlock.Lock()
}

// Unlock will unlock the writers block on the user.
func (u *User) Unlock() {
	u.mainlock.Unlock()
}

// Save will store the Organization object to disk and be cached in
// memory. The save function will also unlock the organization for
// writing. If the org is not locked before saving, a runtime error
// will be called.
// ATM a NO-OP
func (u *User) Save() error {
	u.Unlock()
	return nil
}

// HasUser checks if there is registered a user with the given login name.
// ATM a NO-OP
func HasUser(login string) bool {
	return false
}
