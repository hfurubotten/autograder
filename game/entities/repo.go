package entities

import (
	"encoding/gob"
	"errors"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
)

func init() {
	gob.Register(Repo{})
}

const (
	orgOwner int = iota
	userOwner
)

// Repo represent the repository of TODO.
type Repo struct {
	points.Leaderboard
	// lock sync.Mutex

	Name        string
	Fullname    string
	Description string
	Language    string

	// Owners
	OwnerType int //TODO should be separate type instead of int
	Owner     string
	Admins    map[string]interface{}

	// URLs
	HTMLURL  string
	CloneURL string
	Homepage string
}

// NewRepo will try to load repository info from storage. If
// nothing is found a new empty repo object is returend back.
func NewRepo(owner, name string) (repo *Repo, err error) {
	repo = new(Repo)
	repo.Owner = owner
	repo.Name = name
	repo.Fullname = owner + "/" + name

	err = repo.loadStoredData()
	if err != nil {
		return nil, err
	}

	return
}

// NewRepoWithGithubData will use github data to load repository
// info from storage. The information will also be updated with
// latest from github.
func NewRepoWithGithubData(gr *github.Repository) (repo *Repo, err error) {
	if gr == nil {
		return nil, errors.New("Cannot parse nil object.")
	}

	repo, err = NewRepo(*gr.Owner.Login, *gr.Name)
	if err != nil {
		repo = new(Repo)
		err = nil
	}

	repo.ImportGithubData(gr)

	return
}

// ImportGithubData imports data from the given github
// data object and stores it in the given Repo object.
func (r *Repo) ImportGithubData(gr *github.Repository) {
	if gr == nil {
		return
	}

	if gr.FullName != nil {
		r.Fullname = *gr.FullName
	}

	if gr.Description != nil {
		r.Description = *gr.Description
	}

	if gr.Language != nil {
		r.Language = *gr.Language
	}

	if gr.Language != nil {
		r.Language = *gr.Language
	}

	// Owner information
	if *gr.Owner.Type == githubobjects.USERTYPE {
		r.OwnerType = userOwner
		r.Admins[*gr.Owner.Login] = nil
	} else if *gr.Owner.Type == githubobjects.ORGANIZATIONTYPE {
		r.OwnerType = orgOwner
	}

	r.Owner = *gr.Owner.Login

	// URLs
	if gr.HTMLURL != nil {
		r.HTMLURL = *gr.HTMLURL
	}

	if gr.CloneURL != nil {
		r.CloneURL = *gr.CloneURL
	}

	if gr.Homepage != nil {
		r.Homepage = *gr.Homepage
	}
}

// loadStoredData fetches the repository data stored on disk or in cached memory.
// ATM a NO-OP
func (r *Repo) loadStoredData() (err error) {
	return nil
}

// Lock will lock the user name from being written to by
// other instances of the same organization. This has to be used
// when new info is written, to prevent race conditions. Unlock
// occures when data is finished written to storage.
// func (r *Repo) Lock() {
// 	r.lock.Lock()
// }

// Unlock will unlock the writers block on the user.
// func (r *Repo) Unlock() {
// 	r.lock.Unlock()
// }

// Save stores the repo object to memory cache and disk.
// ATM a NO-OP
func (r *Repo) Save() (err error) {
	// r.Unlock() //TODO why Unlock here??
	return nil
}

// HasRepo checks if there is registered a repo with the given login name.
// ATM a NO-OP
func HasRepo(owner, repo string) bool {
	return false
}
