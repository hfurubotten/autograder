package git

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/mail"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Member{})
	gob.Register(CourseOptions{})
}

type CourseOptions struct {
	Course        string
	CurrentLabNum int
	IsGroupMember bool
	GroupNum      int
}

type Member struct {
	githubclient *github.Client
	Username     string
	Name         string
	Email        *mail.Address
	StudentID    int
	IsTeacher    bool
	IsAssistant  bool
	IsAdmin      bool

	Teaching         map[string]interface{}
	Courses          map[string]CourseOptions
	AssistantCourses map[string]interface{}

	accessToken token
	Scope       string
}

// NewMember tries to use the given oauth token to find the
// user stored on disk/memory. If not found it will load user
// data from github and make a new user.
func NewMember(oauthtoken string) (m Member) {
	m = Member{
		accessToken:      NewToken(oauthtoken),
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]CourseOptions),
		AssistantCourses: make(map[string]interface{}),
	}

	var err error
	if m.accessToken.HasTokenInStore() {
		m.Username, err = m.accessToken.GetUsernameFromTokenInStore()
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		err = m.loadDataFromGithub()
		if err != nil {
			log.Println(err)
			return
		}
	}

	err = m.loadData()
	if err != nil {
		log.Println(err)
		return
	}

	if m.IsTeacher {
		var org Organization
		for k, _ := range m.Teaching {
			org = NewOrganization(k)
			if org.AdminToken != oauthtoken {
				org.AdminToken = oauthtoken
				org.StickToSystem()
			}
		}
	}

	return
}

// NewMemberFromUsername loads a user from storage with the given username.
func NewMemberFromUsername(username string) (m Member) {
	m = Member{}
	m.Username = username

	err := m.loadData()
	if err != nil {
		log.Println(err)
	}

	return
}

func (m *Member) loadDataFromGithub() (err error) {
	err = m.connectToGithub()
	if err != nil {
		return
	}

	user, _, err := m.githubclient.Users.Get("")
	if err != nil {
		return
	}

	if user.Login != nil {
		m.Username = *user.Login
	}

	if user.Name != nil {
		m.Name = *user.Name
	}

	return
}

// loadData loads data from storage if it exists.
func (m *Member) loadData() (err error) {
	if getUserstore().Has(m.Username) {

		err = getUserstore().ReadGob(m.Username, m, false)
		if err != nil {
			return
		}

		if !m.accessToken.HasTokenInStore() {
			m.accessToken.SetUsernameToTokenInStore(m.Username)
		}
	}

	return
}

// StickToSystem stores the user to disk and caches it in memory.
func (m Member) StickToSystem() (err error) {
	return getUserstore().WriteGob(m.Username, m)
}

func (m Member) IsComplete() bool {
	if m.Name == "" || m.StudentID == 0 || m.Username == "" || m.Email == nil {
		return false
	}

	return true
}

// connectToGithub creates a new github client.
func (m *Member) connectToGithub() error {
	if m.githubclient != nil {
		return nil
	}

	if !m.accessToken.HasToken() {
		return errors.New("Missing AccessToken to the memeber. Can't contact github.")
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: m.accessToken.GetToken()},
	}
	m.githubclient = github.NewClient(t.Client())
	return nil
}

// ListOrgs will list all organisations the user is a member of.
func (m *Member) ListOrgs() (ls []string, err error) {
	err = m.connectToGithub()
	if err != nil {
		return
	}

	orgs, _, err := m.githubclient.Organizations.List("", nil)

	ls = make([]string, len(orgs))

	for i, org := range orgs {
		ls[i] = *org.Login
	}

	return
}

// AddOrganization will add a new github organization to attending courses.
func (m *Member) AddOrganization(org Organization) (err error) {
	if m.Courses == nil {
		m.Courses = make(map[string]CourseOptions)
	}

	if _, ok := m.Courses[org.Name]; !ok {
		m.Courses[org.Name] = CourseOptions{
			Course:        org.Name,
			CurrentLabNum: 1,
		}
	}

	return
}

// AddTeachingOrganization will add a new github organization to courses the user are teaching.
func (m *Member) AddTeachingOrganization(org Organization) (err error) {
	if m.Teaching == nil {
		m.Teaching = make(map[string]interface{})
	}

	m.IsTeacher = true
	m.Teaching[org.Name] = nil

	return
}

// AddAssistingOrganization will add a new github organization to courses the user are teaching assistant of.
func (m *Member) AddAssistingOrganization(org Organization) (err error) {
	if m.AssistantCourses == nil {
		m.AssistantCourses = make(map[string]interface{})
	}

	m.IsAssistant = true
	m.AssistantCourses[org.Name] = nil

	return
}

// GetToken returns the users github token.
func (m Member) GetToken() (token string) {
	return m.accessToken.GetToken()
}

// String will make a stringify the member.
func (m Member) String() string {
	return fmt.Sprintf("Student: %s <%s>, Student ID: %d, Github: %s", m.Name, m.Email, m.StudentID, m.Username)
}

// ListAllMembers lists all members stored in the system.
func ListAllMembers() (out []Member) {
	out = make([]Member, 0)
	keys := getUserstore().Keys()
	var m Member

	for key := range keys {
		m = NewMemberFromUsername(key)
		out = append(out, m)
	}

	return
}

// HasMember checks if the user is stored in the system.
func HasMember(username string) bool {
	return getUserstore().Has(username)
}

var userstore *diskv.Diskv

// getUserstore will return the diskv object to access users stored in memory and on disk.
func getUserstore() *diskv.Diskv {
	if userstore == nil {
		userstore = diskv.New(diskv.Options{
			BasePath:     global.Basepath + "diskv/users/",
			CacheSizeMax: 1024 * 1024 * 256,
		})
	}

	return userstore
}
