package entities

import (
	"encoding/gob"
	"errors"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
)

func init() {
	gob.Register(RepoX{})
}

const (
	orgOwner int = iota
	userOwner
)

// Repo represent the repository of TODO.
type RepoX struct {
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

//TODO Hide NewRepo functions

// NewRepo will try to load repository info from storage. If
// nothing is found a new empty repo object is returend back.
func NewRepoX2(owner, name string) (repo *RepoX, err error) {
	repo = new(RepoX)
	repo.Owner = owner
	repo.Name = name
	repo.Fullname = owner + "/" + name

	// err = repo.loadStoredData()
	// if err != nil {
	// 	return nil, err
	// }

	return
}

// NewRepo creates a new Repo object from the provided github repoistory.
// It is an error to pass nil to this function.
func NewRepoX(gr *github.Repository) (r *RepoX) {
	r = &RepoX{}
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

	return r
}

var repoCache map[string]*RepoX

// GetRepo returns a Repo object based on the provided github repoistory.
// It is an error to pass nil to this function.
func GetRepo(gr *github.Repository) (repo *RepoX, err error) {
	if repoCache == nil {
		repoCache = make(map[string]*RepoX)
	}
	if gr == nil {
		return nil, errors.New("no repository provided")
	}
	if repo, ok := repoCache[*gr.FullName]; ok {
		return repo, nil
	}
	repo = NewRepoX(gr)
	repoCache[*gr.FullName] = repo
	return
}
