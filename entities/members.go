package entities

import (
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/database"
	"github.com/hfurubotten/autograder/game/entities"
	"golang.org/x/oauth2"
)

// MemberBucketName is the bucket name for members in the DB.
var MemberBucketName = "members"

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

	accessToken  string
	githubclient *github.Client
}

// NewMember tries to use the given oauth token to find the
// user stored on disk/memory. If not found it will load user
// data from github and make a new user.
func NewMember(oauthtoken string) (m *Member, err error) {
	if oauthtoken == "" {
		return nil, errors.New("Cannot have empty oauth token")
	}

	u := entities.User{
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
	m = &Member{
		User:             u,
		accessToken:      oauthtoken,
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]CourseOptions),
		AssistantCourses: make(map[string]interface{}),
	}

	if has(m.accessToken) {
		m.Username, err = get(m.accessToken)
		if err != nil {
			return nil, err
		}
	} else {
		err = m.loadDataFromGithub()
		if err != nil {
			return nil, err
		}
	}
	user := m.Username
	//TODO THis is too messy; clean up
	mx, err := GetMember(user)
	if err != nil {
		if _, nokey := err.(database.KeyNotFoundError); nokey {
			err = nil
		}
	}
	if mx != nil {
		m = mx
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
func NewUserWithGithubData(gu *github.User) (u *Member, err error) {
	if gu == nil {
		return nil, errors.New("Cannot parse nil github.User object.")
	}

	u, err = NewMemberFromUsername(*gu.Login)
	if err != nil {
		return nil, err
	}

	u.ImportGithubData(gu)

	return
}

// NewMemberFromUsername loads a user from storage with the given username.
func NewMemberFromUsername(userName string) (m *Member, err error) {
	m, err = GetMember(userName)
	if err == nil {
		return m, nil
	}
	u := entities.User{
		Username:     userName,
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
	m = &Member{
		User:             u,
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]CourseOptions),
		AssistantCourses: make(map[string]interface{}),
	}
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

// GetMember returns the member data for the given user.
func GetMember(user string) (*Member, error) {
	var m *Member
	err := database.Get(MemberBucketName, user, &m)
	if err != nil {
		return nil, err
	}
	if !has(m.accessToken) {
		put(m.accessToken, m.Username)
	}
	return m, nil
}

// func (m *Member) loadStoredData() error {
// 	// var m *Member //TODO We should not create object first and then populate it.
// 	err := database.Get(MemberBucketName, m.Username, &m)
// 	if err != nil {
// 		// TODO Why is it ok that the key was not found?
// 		// This check is necessary since otherwise tests fail.
// 		if _, nokey := err.(database.KeyNotFoundError); nokey {
// 			err = nil
// 		}
// 		return err
// 		// return nil, err
// 	}
//
// 	if !m.accessToken.HasTokenInStore() {
// 		m.accessToken.SetUsernameToTokenInStore(m.Username)
// 	}
// 	return nil
// 	// return m, nil
// }

// Save stores the user to disk and caches it in memory.
// save the object will be automatically unlocked.
// NB: If error occure the unlocking of the object need to be done manually.
// Will panic if the member is not locked before saving.
func (m *Member) Save() (err error) {
	return database.Put(MemberBucketName, m.Username, m)
}

// IsComplete checks if all the required fields about the user has content.
func (m *Member) IsComplete() bool {
	if m.Name == "" || m.StudentID == 0 || m.Username == "" || m.Email == nil {
		return false
	}

	return true
}

func (m *Member) hasAccessToken() bool {
	return m.accessToken != "" && len(m.accessToken) > 0
}

// connectToGithub creates a new github client.
func (m *Member) connectToGithub() error {
	if m.githubclient != nil {
		return nil
	}
	if !m.hasAccessToken() {
		return errors.New("unable to connect to github; missing access token for " + m.Username)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: m.accessToken},
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

// SetApprovedBuild will put the approved build results in
func (m *Member) SetApprovedBuild(course string, labnum, buildid int, date time.Time) {
	if _, ok := m.Courses[course]; !ok {
		return
	}

	opt := m.Courses[course]
	if _, ok := opt.Assignments[labnum]; !ok {
		opt.Assignments[labnum] = NewLabAssignmentOptions()
	}

	opt.Assignments[labnum].ApproveDate = date
	opt.Assignments[labnum].ApprovedBuild = buildid
	if opt.CurrentLabNum <= labnum {
		opt.CurrentLabNum = labnum + 1
	}
	m.Courses[course] = opt
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
func (m *Member) GetToken() (token string) {
	return m.accessToken
}

// String will stringify the member.
func (m *Member) String() string {
	return fmt.Sprintf("Student: %s %s, Student ID: %d, Github: %s", m.Name, m.Email, m.StudentID, m.Username)
}

// ListAllMembers returns the list of all members stored in the system.
func ListAllMembers() (members []*Member) {
	database.GetPureDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MemberBucketName))
		if b == nil {
			return errors.New("unknown bucket: " + MemberBucketName)
		}

		b.ForEach(func(k, v []byte) error {
			m, err := NewMemberFromUsername(string(k))
			if err == nil {
				members = append(members, m)
			}
			// continue also if member couldn't be created
			return nil
		})

		return nil
	})

	return members
}

// HasMember checks if the user is stored in the system.
func HasMember(username string) bool {
	return database.Has(MemberBucketName, username)
}
