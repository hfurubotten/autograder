package git

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/database"
	"github.com/hfurubotten/github-gamification/entities"
	"golang.org/x/oauth2"
)

// MemberBucketName is the bucket/table name for organizations in the DB.
var MemberBucketName = "members"

// InMemoryMembers is a mapper where pointers to all the Organization are kept in memory.
var InMemoryMembers = make(map[string]*Member)

// InMemoryMembersLock is the locking for the org mapper.
var InMemoryMembersLock sync.Mutex

func init() {
	gob.Register(Member{})

	database.RegisterBucket(MemberBucketName)
}

// Member represent a student in autograder.
type Member struct {
	entities.User

	StudentID   int
	IsTeacher   bool
	IsAssistant bool
	IsAdmin     bool

	Teaching         map[string]interface{}
	Courses          map[string]CourseOptions
	AssistantCourses map[string]interface{}

	accessToken  Token
	githubclient *github.Client
}

// NewMember tries to use the given oauth token to find the
// user stored on disk/memory. If not found it will load user
// data from github and make a new user.
func NewMember(oauthtoken string, readonly bool) (m *Member, err error) {
	if oauthtoken == "" {
		return nil, errors.New("Cannot have empty oauth token")
	}

	InMemoryMembersLock.Lock()
	defer InMemoryMembersLock.Unlock()

	u := entities.User{
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
	m = &Member{
		User:             u,
		accessToken:      NewToken(oauthtoken),
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]CourseOptions),
		AssistantCourses: make(map[string]interface{}),
	}

	if m.accessToken.HasTokenInStore() {
		m.Username, err = m.accessToken.GetUsernameFromTokenInStore()
		if err != nil {
			return nil, err
		}
	} else {
		err = m.loadDataFromGithub()
		if err != nil {
			return nil, err
		}
	}

	if _, ok := InMemoryMembers[m.Username]; ok {
		m = InMemoryMembers[m.Username]
		if !readonly {
			m.Lock()
		}
	} else {
		err = m.loadStoredData(!readonly)
		if err != nil {
			return nil, err
		}

		InMemoryMembers[m.Username] = m
	}

	if m.IsTeacher {
		var org *Organization
		for k := range m.Teaching {
			org, err = NewOrganization(k, true)
			if err != nil {
				continue
			}

			if org.AdminToken != oauthtoken {
				org.Lock()
				org.AdminToken = oauthtoken
				org.Save()
			}
		}
	}

	return
}

// NewUserWithGithubData creates a new User object from a github User object.
// It will copy all information from the given GitHub data to the new User object.
func NewUserWithGithubData(gu *github.User, readonly bool) (u *Member, err error) {
	if gu == nil {
		return nil, errors.New("Cannot parse nil github.User object.")
	}

	u, err = NewMemberFromUsername(*gu.Login, readonly)
	if err != nil {
		return nil, err
	}

	u.ImportGithubData(gu)

	return
}

// NewMemberFromUsername loads a user from storage with the given username.
func NewMemberFromUsername(username string, readonly bool) (m *Member, err error) {
	InMemoryMembersLock.Lock()
	defer InMemoryMembersLock.Unlock()
	if m, ok := InMemoryMembers[username]; ok {
		if !readonly {
			m.Lock()
		}
		return m, nil
	}

	u := entities.User{
		Username:     username,
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}

	m = &Member{
		User:             u,
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]CourseOptions),
		AssistantCourses: make(map[string]interface{}),
	}

	err = m.loadStoredData(!readonly)
	if err != nil {
		return nil, err
	}

	InMemoryMembers[m.Username] = m

	return m, nil
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

	m.ImportGithubData(user)

	return
}

// loadData loads data from storage if it exists.
func (m *Member) loadStoredData(lock bool) (err error) {
	err = database.GetPureDB().View(func(tx *bolt.Tx) error {
		// locks the object directly in order to ensure consistent info from DB.
		if lock {
			m.Lock()
		}

		b := tx.Bucket([]byte(MemberBucketName))
		if b == nil {
			return errors.New("Bucket not found. Are you sure the bucket was registered correctly?")
		}

		data := b.Get([]byte(m.Username))
		if data == nil {
			return errors.New("No data in database")
		}

		buf := &bytes.Buffer{}
		decoder := gob.NewDecoder(buf)

		n, _ := buf.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		err = decoder.Decode(m)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "No data in database" {
			err = nil
		}
	}

	if !m.accessToken.HasTokenInStore() {
		m.accessToken.SetUsernameToTokenInStore(m.Username)
	}

	return
}

// Save stores the user to disk and caches it in memory.
// save the object will be automatically unlocked.
// NB: If error occure the unlocking of the object need to be done manually.
// Will panic if the member is not locked before saving.
func (m *Member) Save() (err error) {
	return database.GetPureDB().Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(MemberBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		buf := &bytes.Buffer{}
		encoder := gob.NewEncoder(buf)

		if err = encoder.Encode(m); err != nil {
			return
		}

		data, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		err = b.Put([]byte(m.Username), data)
		if err != nil {
			return err
		}

		m.Unlock()
		return nil
	})
}

// IsComplete checks if all the required fields about the user has content.
func (m *Member) IsComplete() bool {
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

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: m.accessToken.GetToken()},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	m.githubclient = github.NewClient(tc)
	return nil
}

// AddBuildResult will add a build result to the group.
func (m *Member) AddBuildResult(course string, lab, buildid int) {
	if _, ok := m.Courses[course]; !ok {
		return
	}

	g := m.Courses[course]

	if g.Assignments == nil {
		g.Assignments = make(map[int]*LabAssignmentOptions)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewLabAssignmentOptions()
	}

	g.Assignments[lab].AddBuildResult(buildid)
}

// GetLastBuildID will get the last build ID added to a lab assignment.
func (m *Member) GetLastBuildID(course string, lab int) int {
	if _, ok := m.Courses[course]; !ok {
		return -1
	}

	g := m.Courses[course]

	if assignment, ok := g.Assignments[lab]; ok {
		if assignment.Builds == nil {
			return -1
		}
		if len(assignment.Builds) == 0 {
			return -1
		}

		return assignment.Builds[len(assignment.Builds)-1]
	}

	return -1
}

// AddNotes will add notes to a lab assignment.
func (m *Member) AddNotes(course string, lab int, notes string) {
	if _, ok := m.Courses[course]; !ok {
		return
	}

	g := m.Courses[course]

	if g.Assignments == nil {
		g.Assignments = make(map[int]*LabAssignmentOptions)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewLabAssignmentOptions()
		m.Courses[course] = g
	}

	g.Assignments[lab].Notes = notes
}

// GetNotes will get notes from a lab assignment.
func (m *Member) GetNotes(course string, lab int) string {
	if _, ok := m.Courses[course]; !ok {
		return ""
	}

	g := m.Courses[course]

	if g.Assignments == nil {
		g.Assignments = make(map[int]*LabAssignmentOptions)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewLabAssignmentOptions()
		m.Courses[course] = g
	}

	return g.Assignments[lab].Notes
}

// ListOrgs will list all organisations the user is a member of on github.
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
func (m *Member) AddOrganization(org *Organization) (err error) {
	if m.Courses == nil {
		m.Courses = make(map[string]CourseOptions)
	}

	if _, ok := m.Courses[org.Name]; !ok {
		m.Courses[org.Name] = NewCourseOptions(org.Name)
	}

	return
}

// RemoveOrganization will remove a github organization from attending courses.
func (m *Member) RemoveOrganization(org *Organization) (err error) {
	if m.Courses == nil {
		m.Courses = make(map[string]CourseOptions)
	}

	if _, ok := m.Courses[org.Name]; ok {
		c := m.Courses[org.Name]

		if c.IsGroupMember {
			g, err := NewGroup(c.Course, c.GroupNum, false)
			if err != nil {
				return err
			}

			g.RemoveMember(m.Username)
			g.Save()
		}

		delete(m.Courses, org.Name)
	}

	return
}

// AddTeachingOrganization will add a new github organization to courses the user are teaching.
func (m *Member) AddTeachingOrganization(org *Organization) (err error) {
	if m.Teaching == nil {
		m.Teaching = make(map[string]interface{})
	}

	m.IsTeacher = true
	m.Teaching[org.Name] = nil

	return
}

// AddAssistingOrganization will add a new github organization to courses the user are teaching assistant of.
func (m *Member) AddAssistingOrganization(org *Organization) (err error) {
	if m.AssistantCourses == nil {
		m.AssistantCourses = make(map[string]interface{})
	}

	m.IsAssistant = true
	m.AssistantCourses[org.Name] = nil

	return
}

// RemoveAssistingOrganization will add a new github organization to courses the user are teaching assistant of.
func (m *Member) RemoveAssistingOrganization(org *Organization) (err error) {
	if m.AssistantCourses == nil {
		m.AssistantCourses = make(map[string]interface{})
	}

	delete(m.AssistantCourses, org.Name)

	return
}

// GetToken returns the users github token.
func (m Member) GetToken() (token string) {
	return m.accessToken.GetToken()
}

// String will stringify the member.
func (m Member) String() string {
	return fmt.Sprintf("Student: %s %s, Student ID: %d, Github: %s", m.Name, m.Email, m.StudentID, m.Username)
}

// ListAllMembers lists all members stored in the system.
func ListAllMembers() (out []*Member) {
	out = make([]*Member, 0)
	keys := []string{}

	database.GetPureDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MemberBucketName))
		if b == nil {
			return errors.New("Unable to find bucket")
		}
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			keys = append(keys, string(k))
		}

		return nil
	})

	for _, key := range keys {
		m, err := NewMemberFromUsername(key, true)
		if err != nil {
			continue
		}

		out = append(out, m)
	}

	return
}

// HasMember checks if the user is stored in the system.
func HasMember(username string) bool {
	return database.Has(MemberBucketName, username)
}
