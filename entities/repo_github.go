package entities

import (
	"errors"

	"github.com/google/go-github/github"
)

var repoCache map[string]*Repo

// GetRepo returns a Repo object based on the provided github repoistory.
// It is an error to pass nil to this function.
func GetRepo(gr *github.Repository) (repo *Repo, err error) {
	if repoCache == nil {
		repoCache = make(map[string]*Repo)
	}
	if gr == nil {
		return nil, errors.New("no repository provided")
	}
	// check cache if we have the repoistory
	if repo, ok := repoCache[*gr.FullName]; ok {
		return repo, nil
	}
	// create new repo object and cache it for next time
	repo = newRepo(gr)
	repoCache[*gr.FullName] = repo
	return
}

// newRepo creates a new Repo object from the provided github repoistory.
// It is an error to pass nil to this function.
func newRepo(gr *github.Repository) (r *Repo) {
	r = &Repo{}
	if gr.Name != nil {
		r.Name = *gr.Name
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

	// Owner information
	switch *gr.Owner.Type {
	case usrType:
		r.OwnerType = usrOwner
		r.Admins[*gr.Owner.Login] = nil
	case orgType:
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
