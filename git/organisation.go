package git

import (
	"encoding/gob"
	"errors"
	"log"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Organization{})
}

var orgstore = diskv.New(diskv.Options{
	BasePath:     "diskv/orgs/",
	CacheSizeMax: 1024 * 1024 * 256,
})

type Organization struct {
	Name                  string
	Description           string
	GroupAssignments      int
	IndividualAssignments int

	StudentTeamID int
	Private       bool

	AdminToken  string
	githubadmin *github.Client
}

func NewOrganization(name string) Organization {
	if orgstore.Has(name) {
		var org Organization
		orgstore.ReadGob(name, &org, false)
		return org
	}
	return Organization{Name: name}
}

func (o *Organization) connectAdminToGithub() error {
	if o.githubadmin != nil {
		return nil
	}

	if o.AdminToken == "" {
		return errors.New("Missing AccessToken to the memeber. Can't contact github.")
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: o.AdminToken},
	}
	o.githubadmin = github.NewClient(t.Client())
	return nil
}

func (o *Organization) AddMembership(member Member) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, _, err = o.githubadmin.Organizations.AddTeamMembership(o.StudentTeamID, member.Username)
	return
}

func (o *Organization) StickToSystem() (err error) {
	return orgstore.WriteGob(o.Name, o)
}

func (o *Organization) Fork(owner, repo string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	forkopt := github.RepositoryCreateForkOptions{Organization: o.Name}
	_, _, err = o.githubadmin.Repositories.CreateFork(owner, repo, &forkopt)
	return
}

func (o *Organization) CreateRepo(opt RepositoryOptions) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	if opt.Name == "" {
		return errors.New("Missing required name field. ")
	}

	repo := &github.Repository{}
	repo.Name = github.String(opt.Name)
	repo.Private = github.Bool(opt.Private)
	repo.AutoInit = github.Bool(opt.AutoInit)
	if opt.TeamID != 0 {
		repo.TeamID = github.Int(opt.TeamID)
	}

	_, _, err = o.githubadmin.Repositories.Create(o.Name, repo)
	return
}

func (o *Organization) CreateTeam(opt TeamOptions) (teamID int, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	team := &github.Team{}
	team.Name = github.String("students")
	if opt.Permission != "" {
		team.Permission = github.String(opt.Permission)
	}
	team, _, err = o.githubadmin.Organizations.CreateTeam(o.Name, team)
	if err != nil {
		return
	}

	if opt.RepoNames != nil {
		for _, repo := range opt.RepoNames {
			_, err = o.githubadmin.Organizations.AddTeamRepo(*team.ID, o.Name, repo)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return *team.ID, nil
}

func (o *Organization) CreateFile(repo, path, content, commitmsg string) (err error) {
	if repo == "" || path == "" || content == "" || commitmsg == "" {
		return errors.New("Missing one of the arguments to create a file.")
	}

	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	contentopt := github.RepositoryContentFileOptions{
		Message: github.String(commitmsg),
		Content: []byte(content),
	}
	_, _, err = o.githubadmin.Repositories.CreateFile(o.Name, repo, path, &contentopt)
	return
}

func ListRegisteredOrganizations() (out []Organization) {
	out = make([]Organization, 0)
	keys := orgstore.Keys()
	var org Organization

	for key := range keys {
		org = NewOrganization(key)
		out = append(out, org)
	}

	return
}
