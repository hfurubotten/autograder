package git

import (
	"encoding/gob"
	"errors"
	"log"
	"strconv"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Organization{})
}

type Organization struct {
	Name                  string
	Description           string
	GroupAssignments      int
	IndividualAssignments int

	IndividualLabFolders map[int]string
	GroupLabFolders      map[int]string

	StudentTeamID int
	Private       bool

	GroupCount         int
	PendingGroup       map[int]interface{}
	PendingRandomGroup map[string]interface{}
	Groups             map[string]interface{}
	PendingUser        map[string]interface{}
	Members            map[string]interface{}
	Teachers           map[string]interface{}

	AdminToken  string
	githubadmin *github.Client

	CI CIOptions
}

func NewOrganization(name string) Organization {
	if GetOrgstore().Has(name) {
		var org Organization
		GetOrgstore().ReadGob(name, &org, false)
		return org
	}
	return Organization{
		Name:                 name,
		IndividualLabFolders: make(map[int]string),
		GroupLabFolders:      make(map[int]string),
		PendingGroup:         make(map[int]interface{}),
		PendingRandomGroup:   make(map[string]interface{}),
		PendingUser:          make(map[string]interface{}),
		Members:              make(map[string]interface{}),
		Teachers:             make(map[string]interface{}),
		CI: CIOptions{
			Basepath: "/testground/src/github.com/" + name + "/",
		},
	}
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
	if err != nil {
		return
	}

	member.AddOrganization(*o)
	err = member.StickToSystem()
	if o.PendingUser == nil {
		o.PendingUser = make(map[string]interface{})
	}

	if _, ok := o.PendingUser[member.Username]; !ok {
		o.PendingUser[member.Username] = nil
	}

	return
}

func (o *Organization) GetMembership(member Member) (status string, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	memship, _, err := o.githubadmin.Organizations.GetTeamMembership(o.StudentTeamID, member.Username)
	if err != nil {
		return
	}

	if memship.State == nil {
		err = errors.New("Couldn't find any role on the username " + member.Username)
		return
	}

	status = *memship.State

	return

}

func (o *Organization) StickToSystem() (err error) {
	if o.IndividualLabFolders == nil {
		o.IndividualLabFolders = make(map[int]string)
	}

	var newfoldernames map[int]string
	if len(o.IndividualLabFolders) != o.IndividualAssignments {
		newfoldernames = make(map[int]string)
		for i := 1; i <= o.IndividualAssignments; i++ {
			if v, ok := o.IndividualLabFolders[i]; ok {
				newfoldernames[i] = v
			} else {
				newfoldernames[i] = "lab" + strconv.Itoa(i)
			}
		}
		o.IndividualLabFolders = newfoldernames
	}

	if o.GroupLabFolders == nil {
		o.GroupLabFolders = make(map[int]string)
	}

	if len(o.GroupLabFolders) != o.GroupAssignments {
		newfoldernames = make(map[int]string)
		for i := 1; i <= o.GroupAssignments; i++ {
			if v, ok := o.GroupLabFolders[i]; ok {
				newfoldernames[i] = v
			} else {
				newfoldernames[i] = "grouplab" + strconv.Itoa(i)
			}
		}
		o.GroupLabFolders = newfoldernames
	}

	return GetOrgstore().WriteGob(o.Name, o)
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
	if err != nil {
		return
	}

	if opt.Hook {
		config := make(map[string]interface{})
		config["url"] = global.Hostname + "/event/hook"
		config["content_type"] = "json"

		hook := github.Hook{
			Name:   github.String("web"),
			Config: config,
		}

		_, _, err = o.githubadmin.Repositories.CreateHook(o.Name, opt.Name, &hook)
	}
	return
}

func (o *Organization) CreateTeam(opt TeamOptions) (teamID int, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	team := &github.Team{}
	team.Name = github.String(opt.Name)
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

func (o *Organization) LinkRepoToTeam(teamID int, repo string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, err = o.githubadmin.Organizations.AddTeamRepo(teamID, o.Name, repo)
	return
}

func (o *Organization) AddMemberToTeam(teamID int, user string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, _, err = o.githubadmin.Organizations.AddTeamMembership(teamID, user)
	return
}

func (o *Organization) ListTeams() (teams map[string]Team, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	teams = make(map[string]Team)

	gitteams, _, err := o.githubadmin.Organizations.ListTeams(o.Name, nil)
	if err != nil {
		return
	}

	var team Team
	for _, t := range gitteams {
		team = Team{}
		if t.ID != nil {
			team.ID = *t.ID
		}
		if t.Name != nil {
			team.Name = *t.Name
		}
		if t.Permission != nil {
			team.Permission = *t.Permission
		}
		if t.MembersCount != nil {
			team.MemberCount = *t.MembersCount
		}
		if t.ReposCount != nil {
			team.Repocount = *t.ReposCount
		}

		teams[team.Name] = team
	}

	return
}

func (o *Organization) ListRepos() (repos map[string]Repo, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	repolist, _, err := o.githubadmin.Repositories.ListByOrg(o.Name, nil)

	repos = make(map[string]Repo)

	var repo Repo
	for _, r := range repolist {
		repo = Repo{}
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

		repos[repo.Name] = repo
	}

	return
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
	keys := GetOrgstore().Keys()
	var org Organization

	for key := range keys {
		org = NewOrganization(key)
		out = append(out, org)
	}

	return
}

func HasOrganization(name string) bool {
	return GetOrgstore().Has(name)
}

var orgstore *diskv.Diskv

func GetOrgstore() *diskv.Diskv {
	if orgstore == nil {
		orgstore = diskv.New(diskv.Options{
			BasePath:     global.Basepath + "diskv/orgs/",
			CacheSizeMax: 1024 * 1024 * 256,
		})
	}

	return orgstore
}
