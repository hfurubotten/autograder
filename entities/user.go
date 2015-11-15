package entities

import (
	"net/mail"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/game/levels"
	"github.com/hfurubotten/autograder/game/trophies"
)

// UserProfile contains information about a user. Often this information will
// be gleaned from external sources such as GitHub.
type UserProfile struct {
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
	accessToken  string // will not be stored in DB (since package private access)
	Scope        string
}

// CreateUserProfile returns a new UserProfile populated with data from github.
func CreateUserProfile(userName string) (u *UserProfile, err error) {
	u = &UserProfile{
		Username:     userName,
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
	return u, nil
}

// NewUserProfile returns a new UserProfile.
func NewUserProfile(token, user, scope string) *UserProfile {
	return &UserProfile{
		Username:     user,
		Scope:        scope,
		accessToken:  token,
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
}

func (u *UserProfile) hasAccessToken() bool {
	return u.accessToken != "" && len(u.accessToken) > 0
}

// IncScoreBy increases the total score with given amount.
func (u *UserProfile) IncScoreBy(score int) {
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
func (u *UserProfile) DecScoreBy(score int) {
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
func (u *UserProfile) IncLevel() {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.Level++
}

// DecLevel decreases the level with one until it equals zero.
func (u *UserProfile) DecLevel() {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.Level > 0 {
		u.Level--
	}
}

// GetTrophyChest return the users ThropyChest.
func (u *UserProfile) GetTrophyChest() *trophies.TrophyChest {
	if u.Trophies == nil {
		u.Trophies = trophies.NewTrophyChest()
	}

	return u.Trophies
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
