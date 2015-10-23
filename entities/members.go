package entities

import (
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/database"
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
	UserProfile //TODO clean up; make clearer separation between User and Member and the game stuff

	StudentID   int
	IsTeacher   bool
	IsAssistant bool
	IsAdmin     bool

	Teaching         map[string]interface{}
	Courses          map[string]Course
	AssistantCourses map[string]interface{}

	accessToken  string
	githubclient *github.Client
}

// NewMember tries to use the given oauth token to find the
// user stored on disk/memory. If not found it will load user
// data from github and make a new user.
func NewMember(token string) (m *Member, err error) {
	if token == "" {
		return nil, errors.New("non-empty OAuth token is required")
	}
	var user string
	if hasToken(token) {
		user, err = getToken(token)
		if err != nil {
			return nil, err
		}
		m, err = GetMember(user) //TODO need to pass in token also
		if err != nil {
			return nil, err
		}
		// m.accessToken = token //TODO or just set it here!!
	} else {
		//TODO clean up this code later
		//TODO This code branch is probably not being tested; it should be
		u := UserProfile{
			Username:     user,
			WeeklyScore:  make(map[int]int64),
			MonthlyScore: make(map[time.Month]int64),
		}
		m = &Member{
			UserProfile:      u,
			accessToken:      token,
			Teaching:         make(map[string]interface{}),
			Courses:          make(map[string]Course),
			AssistantCourses: make(map[string]interface{}),
		}
		err = m.loadDataFromGithub()
		if err != nil {
			return nil, err
		}
	}

	//TODO Refactor: This code should be moved elsewhere
	if m.IsTeacher {
		var org *Organization
		for k := range m.Teaching {
			org, err = NewOrganization(k, true)
			if err != nil {
				continue
			}

			if org.AdminToken != token {
				org.Lock()
				org.AdminToken = token
				org.Save()
			}
		}
	}

	return
}

// GetMember returns the member associated with the given userName.
//TODO should this also take token?
func GetMember(userName string) (m *Member, err error) {
	err = database.Get(MemberBucketName, userName, &m)
	if err == nil {
		// TODO We should make sure that the access token is always saved whenever
		// we put a member into the MemberBucketName. So that we can remove this code.
		// unless token for userName is already stored
		if !hasToken(m.accessToken) {
			// make reverse lookup association from token to userName.
			putToken(m.accessToken, m.Username)
		}
		// userName found in database; return early
		return m, nil
	}

	// userName not found in database; create new member object
	u := UserProfile{
		Username:     userName,
		WeeklyScore:  make(map[int]int64),
		MonthlyScore: make(map[time.Month]int64),
	}
	m = &Member{
		// accessToken:      token,
		UserProfile:      u,
		Teaching:         make(map[string]interface{}),
		Courses:          make(map[string]Course),
		AssistantCourses: make(map[string]interface{}),
	}
	return m, nil
}

// Update database under a lock regime to ensure safety.
func (m *Member) Update(fn func() error) (err error) {
	//TODO implement safe locking
	return nil
}

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
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
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
		opt.Assignments[labnum] = NewAssignment()
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
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
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
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
		m.Courses[course] = g
	}

	return g.Assignments[lab].Notes
}

// AddOrganization will add a new github organization to attending courses.
func (m *Member) AddOrganization(org *Organization) (err error) {
	if m.Courses == nil {
		m.Courses = make(map[string]Course)
	}

	if _, ok := m.Courses[org.Name]; !ok {
		m.Courses[org.Name] = NewCourse(org.Name)
	}

	return
}

// RemoveOrganization will remove a github organization from attending courses.
func (m *Member) RemoveOrganization(org *Organization) (err error) {
	if m.Courses == nil {
		m.Courses = make(map[string]Course)
	}

	if _, ok := m.Courses[org.Name]; ok {
		c := m.Courses[org.Name]

		if c.IsGroupMember {
			g, err := NewGroup(c.CourseName, c.GroupNum, false)
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
			m, err := GetMember(string(k))
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
