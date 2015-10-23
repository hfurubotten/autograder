package entities

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/database"
	"github.com/hfurubotten/autograder/global"
	"golang.org/x/oauth2"
)

// OrganizationBucketName is the bucket/table name for organizations in the DB.
var OrganizationBucketName = "orgs"

// InMemoryOrgs is a mapper where pointers to all the Organization are kept in memory.
var InMemoryOrgs = make(map[string]*Organization)

// InMemoryOrgsLock is the locking for the org mapper.
var InMemoryOrgsLock sync.Mutex

func init() {
	gob.Register(Organization{})

	database.RegisterBucket(OrganizationBucketName)
}

// Organization represent a course and a organization on github.
type Organization struct {
	OrganizationX

	GroupAssignments      int
	IndividualAssignments int

	// Lab assignment info. TODO: collect this into one struct!
	IndividualLabFolders map[int]string
	GroupLabFolders      map[int]string
	IndividualDeadlines  map[int]time.Time
	GroupDeadlines       map[int]time.Time

	StudentTeamID int
	TeacherTeamID int
	Private       bool

	//GroupCount         int
	PendingGroup       map[int]interface{}
	PendingRandomGroup map[string]interface{}
	// TODO: change the groups field to ActiveGroups map[int]interface{}
	// will need change in logic several places in the web package!
	Groups      map[string]interface{}
	PendingUser map[string]interface{}
	Members     map[string]interface{}
	Teachers    map[string]interface{}

	CodeReview     bool
	CodeReviewlist []codeReviewID

	Slipdays    bool
	SlipdaysMax int

	AdminToken  string
	githubadmin *github.Client

	CI CIOptions
}

// NewOrganization tries to fetch a organization from storage on disk or memory.
// If non exists with given name, it creates a new organization.
func NewOrganization(name string, readonly bool) (org *Organization, err error) {
	InMemoryOrgsLock.Lock()
	defer InMemoryOrgsLock.Unlock()

	if _, ok := InMemoryOrgs[name]; ok {
		org = InMemoryOrgs[name]
		if !readonly {
			org.Lock()
		}
		return org, nil
	}

	o, err := NewOrganizationX(name)
	if err != nil {
		return nil, err
	}

	org = &Organization{
		OrganizationX:        *o,
		IndividualLabFolders: make(map[int]string),
		GroupLabFolders:      make(map[int]string),
		PendingGroup:         make(map[int]interface{}),
		PendingRandomGroup:   make(map[string]interface{}),
		Groups:               make(map[string]interface{}),
		PendingUser:          make(map[string]interface{}),
		Members:              make(map[string]interface{}),
		Teachers:             make(map[string]interface{}),
		IndividualDeadlines:  make(map[int]time.Time),
		GroupDeadlines:       make(map[int]time.Time),
		CodeReviewlist:       make([]codeReviewID, 0),
		CI: CIOptions{
			Basepath: "/testground/src/github.com/" + name + "/",
			Secret:   fmt.Sprintf("%x", md5.Sum([]byte(name+time.Now().String()))),
		},
	}

	err = org.LoadStoredData(!readonly)
	if err != nil {
		if err.Error() != "No data in database" {
			return nil, err
		}
	}

	// Add the org to in memory mapper.
	InMemoryOrgs[org.Name] = org

	return org, nil
}

// NewOrganizationWithGithubData will create a new organization object
// from a github organization data object. It will first attempt to
// load it from storage, if not found it creates a new one.
func NewOrganizationWithGithubData(gorg *github.Organization, readonly bool) (org *Organization, err error) {
	if gorg == nil {
		return nil, errors.New("Cannot use nil github.Organization object")
	}

	org, err = NewOrganization(*gorg.Login, readonly)
	if err != nil {
		return nil, err
	}

	org.ImportGithubDataX(gorg)
	return
}

// connectAdminToGithub will create a github client. This client will be used to talk with githubs api.
func (o *Organization) connectAdminToGithub() error {
	if o.githubadmin != nil {
		return nil
	}

	if o.AdminToken == "" {
		return errors.New("Missing AccessToken to the memeber. Can't contact github.")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: o.AdminToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	o.githubadmin = github.NewClient(tc)
	return nil
}

// LoadStoredData fetches the organization data stored on disk or in cached memory.
func (o *Organization) LoadStoredData(lock bool) (err error) {
	database.GetPureDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OrganizationBucketName))
		if b == nil {
			return errors.New("Bucket not found. Are you sure the bucket was registered correctly?")
		}

		data := b.Get([]byte(o.Name))
		if data == nil {
			return errors.New("No data in database")
		}

		buf := &bytes.Buffer{}
		decoder := gob.NewDecoder(buf)

		n, _ := buf.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		err = decoder.Decode(o)
		if err != nil {
			return err
		}

		return nil
	})

	// locks the object directly in order to ensure consistent info from DB.
	if lock {
		o.Lock()
	}

	return
}

// Save will store the organization to cached memory and disk. On successfull
// save the object will be automatically unlocked.
// NB: If error occure the unlocking of the object need to be done manually.
func (o *Organization) Save() (err error) {
	return database.GetPureDB().Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(OrganizationBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		buf := &bytes.Buffer{}
		encoder := gob.NewEncoder(buf)

		if err = encoder.Encode(o); err != nil {
			return
		}

		data, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		err = b.Put([]byte(o.Name), data)
		if err != nil {
			return err
		}

		o.Unlock()
		return nil
	})
}

// AddCodeReview will add a new code review. This method will
// upload the codereview to github and append it to the list over
// code reviews in this organization.
//
// Filename format committed to github: 'CR-ID'-'Title'-'Username'.'file_ext'
// Commit message: 'CR-ID' 'Username': 'Title'
//
// This method needs locking
func (o *Organization) AddCodeReview(cr *CodeReview) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, labname, _ := o.FindCurrentLab()

	var path string
	if labname != "" {
		path = fmt.Sprintf("%s/%d-%s-%s.%s", labname, cr.ID, strings.Replace(cr.Title, " ", "", -1), cr.User, cr.Ext)
	} else {
		path = fmt.Sprintf("%d-%s-%s.%s", cr.ID, strings.Replace(cr.Title, " ", "", -1), cr.User, cr.Ext)
	}
	commitmsg := fmt.Sprintf("%d %s: %s", cr.ID, cr.User, cr.Title)

	// Creates the review file
	SHA, err := o.CreateFile(CodeReviewRepoName, path, cr.Code+"\n", commitmsg)
	if err != nil {
		return
	}

	commentmsg := fmt.Sprintf("Code Review %d: %s\n\n%s\n\nHey, could someone look through this and give me some feedback conserning this? \nSincerely @%s\n\n---------\n@%s, follow this tread for feedback.\n",
		cr.ID, cr.Title, cr.Desc, cr.User, cr.User)

	// Makes a comment on the commit.
	comment := new(github.RepositoryComment)
	comment.Body = github.String(commentmsg)

	_, _, err = o.githubadmin.Repositories.CreateComment(o.Name, CodeReviewRepoName, SHA, comment)
	if err != nil {
		return
	}

	cr.URL = fmt.Sprintf("https://github.com/%s/%s/commit/%s", o.Name, CodeReviewRepoName, SHA)

	o.CodeReviewlist = append(o.CodeReviewlist, cr.ID)
	return nil
}

// FindCurrentLab will find out which lab has the nearest deadline.
// If no lab has been found the labnum return value will be zero.
func (o *Organization) FindCurrentLab() (labnum int, labname string, labtype int) {
	var lowesttimediff int64 = math.MaxInt64
	for i, t := range o.IndividualDeadlines {
		if time.Now().After(t) {
			continue
		}

		diff := t.Unix() - time.Now().Unix()
		if diff < lowesttimediff {
			labnum = i
			labname = o.IndividualLabFolders[i]
			labtype = IndividualType
			lowesttimediff = diff
		}
	}

	for i, t := range o.GroupDeadlines {
		if time.Now().After(t) {
			continue
		}

		diff := t.Unix() - time.Now().Unix()
		if diff < lowesttimediff {
			labnum = i
			labname = o.GroupLabFolders[i]
			labtype = GroupType
			lowesttimediff = diff
		}
	}
	return
}

// AddMembership will add a user as a pending student in this
// organization. A pending student is a student which still has
// to be approved by the teaching staff. This method will also
// add the user to the student team on github.
//
// This method needs locking
func (o *Organization) AddMembership(member *Member) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	var teams map[string]Team
	if o.StudentTeamID == 0 {
		teams, err = o.ListTeams()
		students, ok := teams[studentsTeam]
		if !ok {
			return errors.New("Couldn't find the students team.")
		}
		o.StudentTeamID = students.ID
	}

	_, _, err = o.githubadmin.Organizations.AddTeamMembership(o.StudentTeamID,
		member.Username,
		&github.OrganizationAddTeamMembershipOptions{
			Role: "member",
		})
	if err != nil {
		return
	}

	if o.PendingUser == nil {
		o.PendingUser = make(map[string]interface{})
	}

	o.PendingUser[member.Username] = nil

	return
}

// RemoveMembership will remove a user from this organization
// on github and from the course in Autograder.
//
// This method needs locking
func (o *Organization) RemoveMembership(member *Member) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, err = o.githubadmin.Organizations.RemoveOrgMembership(member.Username, o.Name)
	if err != nil {
		return
	}

	if o.PendingUser == nil {
		o.PendingUser = make(map[string]interface{})
	}

	if _, ok := o.PendingUser[member.Username]; ok {
		delete(o.PendingUser, member.Username)
	}

	if _, ok := o.Members[member.Username]; ok {
		delete(o.Members, member.Username)
	}

	return
}

// AddTeacher will add a teacher to the teaching staff. This
// method also adds the user to the owners team on github.
//
// TODO: Owners team is no longer a spesial admin team over the
// organization on github. This method needs to be rewritten
// to suppert the new admin API.
//
// This method needs locking
func (o *Organization) AddTeacher(member *Member) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	o.Teachers[member.Username] = nil

	var teams map[string]Team
	if o.TeacherTeamID == 0 {
		teams, err = o.ListTeams()
		owners, ok := teams[teachersTeam]
		if !ok {
			return errors.New("Couldn't find the owners team.")
		}
		o.TeacherTeamID = owners.ID
	}

	_, _, err = o.githubadmin.Organizations.AddTeamMembership(o.TeacherTeamID, member.Username,
		&github.OrganizationAddTeamMembershipOptions{
			Role: "member",
		})
	return
}

// RemoveTeacher will remove a teacher from the teaching staff. This
// method also removes the user to the owners team on github.
//
// TODO: Owners team is no longer a spesial admin team over the
// organization on github. This method needs to be rewritten
// to suppert the new admin API.
//
// This method needs locking
func (o *Organization) RemoveTeacher(member *Member) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	delete(o.Teachers, member.Username)

	var teams map[string]Team
	if o.TeacherTeamID == 0 {
		teams, err = o.ListTeams()
		owners, ok := teams[teachersTeam]
		if !ok {
			return errors.New("Couldn't find the owners team.")
		}
		o.TeacherTeamID = owners.ID
	}

	_, err = o.githubadmin.Organizations.RemoveTeamMembership(o.TeacherTeamID, member.Username)
	return
}

// IsTeacher returns whether if a user is a teacher or not.
func (o *Organization) IsTeacher(member *Member) bool {
	_, orgok := o.Teachers[member.Username]

	// Clean up if any sync problems occures
	var mok bool
	if member.IsTeacher {
		_, mok = member.Teaching[o.Name]
		_, aok := member.AssistantCourses[o.Name]
		if orgok && (!mok && !aok) {
			member.Teaching[o.Name] = nil
			member.Save() // This line is not tread safe!
		}

		if mok && aok {
			delete(member.AssistantCourses, o.Name)
			member.Save()
		}
	} else {
		var ok bool
		if _, ok = member.Teaching[o.Name]; orgok && ok {
			delete(member.Teaching, o.Name)
		}

		_, mok = member.AssistantCourses[o.Name]
		if orgok && !mok {
			member.AssistantCourses[o.Name] = nil
			member.Save() // This line is not tread safe!
		} else if ok {
			member.Save() // This line is not tread safe!
		}
	}

	if !orgok && mok {
		o.Teachers[member.Username] = nil
	}

	return orgok || mok
}

// IsMember return whether if the user is a member or not.
func (o *Organization) IsMember(member *Member) bool {
	if o.IsTeacher(member) {
		return true
	}

	_, orgok := o.Members[member.Username]
	_, mok := member.Courses[o.Name]

	if orgok && !mok {
		member.Courses[o.Name] = NewCourse(o.Name)
	} else if !orgok && mok {
		o.Members[member.Username] = nil
	}

	return orgok || mok
}

// SetIndividualDeadline will set the deadline of one lab assignment.
//
// This method needs locking
func (o *Organization) SetIndividualDeadline(lab int, t time.Time) {
	if o.IndividualDeadlines == nil {
		o.IndividualDeadlines = make(map[int]time.Time)
	}

	o.IndividualDeadlines[lab] = t
}

// SetGroupDeadline will set the deadline of one lab assignment.
//
// This method needs locking
func (o *Organization) SetGroupDeadline(lab int, t time.Time) {
	if o.GroupDeadlines == nil {
		o.GroupDeadlines = make(map[int]time.Time)
	}

	o.GroupDeadlines[lab] = t
}

// AddGroup will add a group to the list of groups in the
// organization and also in the pending group list. The
// pending group list will have to be approved by the teaching
// staff.
//
// This method needs locking
func (o *Organization) AddGroup(g *Group) {
	if o.Groups == nil {
		o.Groups = make(map[string]interface{})
	}

	if _, ok := o.PendingGroup[g.ID]; ok {
		delete(o.PendingGroup, g.ID)
	}
	o.Groups["group"+strconv.Itoa(g.ID)] = nil
}

// GetMembership will return the status of a membership to the
// student team on github. The states possible is active or
// pending. Returns error if user is never invited.
func (o *Organization) GetMembership(member *Member) (status string, err error) {
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

// Fork will fork a different repository into the organization on
// github. The fork call is async. Only communication errors will
// be reported back, no errors in the forking process.
func (o *Organization) Fork(owner, repo string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	forkopt := github.RepositoryCreateForkOptions{Organization: o.Name}
	_, _, err = o.githubadmin.Repositories.CreateFork(owner, repo, &forkopt)
	return
}

// CreateRepo will create a new repository in the organization on github.
//
// TODO: When the hook option is activated, it can only create a push hook.
// Extend this to include a optional event hook.
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
	repo.HasIssues = github.Bool(opt.Issues)
	if opt.TeamID != 0 {
		repo.TeamID = github.Int(opt.TeamID)
	}

	_, _, err = o.githubadmin.Repositories.Create(o.Name, repo)
	if err != nil {
		return
	}

	if opt.Hook != "" {
		config := make(map[string]interface{})
		config["url"] = global.Hostname + "/event/hook"
		config["content_type"] = "json"

		hook := github.Hook{
			Name:   github.String("web"),
			Config: config,
			Events: []string{
				opt.Hook,
			},
		}

		_, _, err = o.githubadmin.Repositories.CreateHook(o.Name, opt.Name, &hook)
	}
	return
}

// CreateTeam will create a new team in the organization on github.
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
			_, err = o.githubadmin.Organizations.AddTeamRepo(*team.ID, o.Name, repo, nil)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return *team.ID, nil
}

// LinkRepoToTeam will link a repo to a team on github.
func (o *Organization) LinkRepoToTeam(teamID int, repo string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, err = o.githubadmin.Organizations.AddTeamRepo(teamID, o.Name, repo, nil)
	return
}

// AddMemberToTeam will add a user to a team on github.
func (o *Organization) AddMemberToTeam(teamID int, user string) (err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	_, _, err = o.githubadmin.Organizations.AddTeamMembership(teamID, user,
		&github.OrganizationAddTeamMembershipOptions{
			Role: "member",
		})
	return
}

// ListTeams will list all the teams within the organization on github.
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

// ListRepos lists all the repositories in the organization on github.
func (o *Organization) ListRepos() (repos map[string]Repo, err error) {
	err = o.connectAdminToGithub()
	if err != nil {
		return nil, err
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

// CreateFile will commit a new file to a repository in the organization on github.
func (o *Organization) CreateFile(repo, path, content, commitmsg string) (commitcode string, err error) {
	if repo == "" || path == "" || content == "" || commitmsg == "" {
		err = errors.New("Missing one of the arguments to create a file.")
		return
	}

	err = o.connectAdminToGithub()
	if err != nil {
		return
	}

	contentopt := github.RepositoryContentFileOptions{
		Message: github.String(commitmsg),
		Content: []byte(content),
	}
	commit, _, err := o.githubadmin.Repositories.CreateFile(o.Name, repo, path, &contentopt)
	if err != nil {
		return
	}

	return *commit.SHA, nil
}

// ListRegisteredOrganizations returns the list of Autograder organizations.
func ListRegisteredOrganizations() (orgs []*Organization) {
	// iteration function called for each entry in the organization bucket
	fn := func(k, v []byte) error {
		org, err := NewOrganization(string(k), true)
		if err != nil {
			// this will terminate the iteration, even if other orgs could be created
			return err
		}
		orgs = append(orgs, org)
		return nil
	}

	err := database.ForEach(OrganizationBucketName, fn)
	if err != nil {
		log.Println(err)
	}
	return
}

// HasOrganization checks if the organization is already registered in autograder.
func HasOrganization(name string) bool {
	if _, ok := InMemoryOrgs[name]; ok {
		return true
	}

	return database.Has(OrganizationBucketName, name)
}
