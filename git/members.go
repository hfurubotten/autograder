package git

import (
	"encoding/gob"
	"errors"
	"log"

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

func (m *Member) loadData() (err error) {
	if getUserstore().Has(m.Username) {
		var tmp Member

		err = getUserstore().ReadGob(m.Username, &tmp, false)
		if err != nil {
			return
		}
		m.Copy(tmp)

		if !m.accessToken.HasTokenInStore() {
			m.accessToken.SetUsernameToTokenInStore(m.Username)
		}
	}

	return
}

func (m Member) StickToSystem() (err error) {
	return getUserstore().WriteGob(m.Username, m)
}

func (m *Member) Copy(tmp Member) {
	m.Username = tmp.Username
	m.Name = tmp.Name
	m.StudentID = tmp.StudentID
	m.IsTeacher = tmp.IsTeacher
	m.IsAdmin = tmp.IsAdmin
	m.Teaching = tmp.Teaching
	m.Courses = tmp.Courses
	m.AssistantCourses = tmp.AssistantCourses
}

func (m Member) IsComplete() bool {
	if m.Name == "" || m.StudentID == 0 || m.Username == "" {
		return false
	}

	return true
}

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

func (m *Member) AddTeachingOrganization(org Organization) (err error) {
	if m.Teaching == nil {
		m.Teaching = make(map[string]interface{})
	}

	m.IsTeacher = true
	if _, ok := m.Teaching[org.Name]; !ok {
		m.Teaching[org.Name] = nil
	}

	return
}

func (m *Member) AddAssistingOrganization(org Organization) (err error) {
	if m.AssistantCourses == nil {
		m.AssistantCourses = make(map[string]interface{})
	}

	if _, ok := m.AssistantCourses[org.Name]; !ok {
		m.AssistantCourses[org.Name] = nil
	}

	return
}

func (m Member) GetToken() (token string) {
	return m.accessToken.GetToken()
}

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

func HasMember(username string) bool {
	return getUserstore().Has(username)
}

var userstore *diskv.Diskv

func getUserstore() *diskv.Diskv {
	if userstore == nil {
		userstore = diskv.New(diskv.Options{
			BasePath:     global.Basepath + "diskv/users/",
			CacheSizeMax: 1024 * 1024 * 256,
		})
	}

	return userstore
}
