package entities

import "github.com/google/go-github/github"

// RepositoryOptions represent the option when needed to create a repository within a organization.
type RepositoryOptions struct {
	Name     string
	Private  bool
	TeamID   int
	AutoInit bool
	Issues   bool
	Hook     string
}

// Repo represent a existing repository.
type Repo struct {
	Name     string
	HTMLURL  string
	CloneURL string
	Private  bool
	TeamID   int
}

func NewRepo(r *github.Repository) Repo {
	repo := Repo{}
	if r.Name != nil {
		repo.Name = *r.Name
	}
	if r.HTMLURL != nil {
		repo.HTMLURL = *r.HTMLURL
	}
	if r.CloneURL != nil {
		repo.CloneURL = *r.CloneURL
	}
	if r.Private != nil {
		repo.Private = *r.Private
	}
	if r.TeamID != nil {
		repo.TeamID = *r.TeamID
	}
	return repo
}
