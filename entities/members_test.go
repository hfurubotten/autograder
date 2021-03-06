package git

import (
	"net/mail"
	"testing"
	"time"

	"github.com/hfurubotten/autograder/game/entities"
)

var testNewMemberInput = []struct {
	token    string
	username string
	studid   int
}{
	{
		"123456789abcdef",
		"user100",
		123456789,
	},
	{
		"12345674489abcdef",
		"user101",
		987654321,
	},
	{
		"12345655789abcdef",
		"user102",
		156789,
	},
}

func TestNewMember(t *testing.T) {
	for _, in := range testNewMemberInput {
		tok := NewToken(in.token)
		if err := tok.SetUsernameToTokenInStore(in.username); err != nil {
			t.Error("Error storing tokens with username:", err)
			continue
		}

		m, err := NewMember(in.token, false)
		if err != nil {
			t.Error("Error creating new member:", err)
			continue
		}

		if m.Username != in.username {
			t.Errorf("Username does not match. %v != %v", m.Username, in.username)
			continue
		}

		m.StudentID = in.studid

		if err = m.Save(); err != nil {
			t.Error("Could not save user:", err)
		}

		testListAllMembersInput = append(testListAllMembersInput, in.username)

		m2, err := NewMember(in.token, false)
		if err != nil {
			t.Error("Error creating new member:", err)
			continue
		}

		if m2.Username != in.username {
			t.Errorf("Username does not match. %v != %v", m.Username, in.username)
			continue
		}

		if m2.StudentID != in.studid {
			t.Errorf("StudentID does not match. %v != %v", m.StudentID, in.studid)
			continue
		}

		m.StudentID = in.studid + 1

		if err = m.Save(); err != nil {
			t.Error("Could not save user:", err)
		}

		m3, err := NewMember(in.token, true)
		if err != nil {
			t.Error("Error creating new member:", err)
			continue
		}

		if m3.Username != in.username {
			t.Errorf("Username does not match. %v != %v", m.Username, in.username)
			continue
		}

		if m3.StudentID != in.studid+1 {
			t.Errorf("StudentID does not match. %v != %v", m.StudentID, in.studid)
			continue
		}

		// checks again with loading from DB and not just memory caches
		delete(InMemoryMembers, in.username)
		m4, err := NewMember(in.token, true)
		if err != nil {
			t.Error("Error creating new member:", err)
			continue
		}

		if m4.Username != in.username {
			t.Errorf("Username does not match. %v != %v", m.Username, in.username)
			continue
		}

		if m4.StudentID != in.studid+1 {
			t.Errorf("StudentID does not match. %v != %v", m.StudentID, in.studid)
			continue
		}
	}
}

var testIsComplete = []struct {
	in   *Member
	want bool
}{
	{
		&Member{
			User: entities.User{
				Name: "Ola Normann",
			},
		},
		false,
	},
	{
		&Member{
			User: entities.User{
				Name: "Ola Normann",
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			User: entities.User{
				Name:     "Ola Normann",
				Username: "olanormann",
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			User: entities.User{
				Name:     "Ola Normann",
				Username: "olanormann",
			},
		},
		false,
	},
	{
		&Member{
			User: entities.User{
				Username: "olanormann",
				Email:    &mail.Address{},
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			User: entities.User{
				Name:     "Ola Normann",
				Username: "olanormann",
				Email:    &mail.Address{},
			},

			StudentID: 112222,
		},
		true,
	},
}

func TestIsComplete(t *testing.T) {
	for _, tcase := range testIsComplete {
		if tcase.in.IsComplete() != tcase.want {
			t.Errorf("Error while checking if the member object is complete. Got %t for %v, want %t", tcase.in.IsComplete(), tcase.in, tcase.want)
		}
	}
}

var testAddOrganizations = []struct {
	in []string
}{
	{
		[]string{
			"course1",
		},
	},
	{
		[]string{
			"course1",
			"course2",
		},
	},
	{
		[]string{
			"course1",
			"course2",
			"course3",
		},
	},
	{
		[]string{
			"course1",
			"course2",
			"course3",
			"course4",
		},
	},
}

func TestAddOrganization(t *testing.T) {
	for _, tcase := range testAddOrganizations {
		user := &Member{}

		for _, cname := range tcase.in {
			org, err := NewOrganization(cname, true)
			if err != nil {
				t.Error("Error creating org", err)
			}
			user.AddOrganization(org)
		}

		for _, cname := range tcase.in {
			if opt, ok := user.Courses[cname]; ok {
				if opt.Course != cname {
					t.Errorf("In course options for %s got wrong course name. Got %s", cname, opt.Course)
				}
			} else {
				t.Errorf("%s is missing in course list", cname)
			}
		}
	}
}

func TestAddTeachingOrganization(t *testing.T) {
	for _, tcase := range testAddOrganizations {
		user := &Member{}

		for _, cname := range tcase.in {
			org, err := NewOrganization(cname, true)
			if err != nil {
				t.Error("Error creating org", err)
			}
			user.AddTeachingOrganization(org)
		}

		if !user.IsTeacher {
			t.Error("User did not get upgraded to teacher.")
		}

		for _, cname := range tcase.in {
			if _, ok := user.Teaching[cname]; !ok {
				t.Errorf("%s is missing in course list", cname)
			}
		}
	}
}

func TestAddAssistingOrganization(t *testing.T) {
	for _, tcase := range testAddOrganizations {
		user := &Member{}

		for _, cname := range tcase.in {
			org, err := NewOrganization(cname, true)
			if err != nil {
				t.Error("Error creating org", err)
			}
			user.AddAssistingOrganization(org)
		}

		if !user.IsAssistant {
			t.Error("User did not get upgraded to assistant.")
		}

		for _, cname := range tcase.in {
			if _, ok := user.AssistantCourses[cname]; !ok {
				t.Errorf("%s is missing in course list", cname)
			}
		}
	}
}

var testListAllMembersInput = []string{}

func TestListAllMembers(t *testing.T) {
	for _, username := range testListAllMembersInput {
		user, err := NewMemberFromUsername(username, false)
		if err != nil {
			t.Errorf("Error getting user: %v", err)
		}
		user.Save()
	}

	list := ListAllMembers()

	i := 0
	for _, user := range list {
		i++

		found := false
		for _, username := range testListAllMembersInput {
			if username == user.Username {
				found = true
			}
		}

		if !found {
			t.Errorf("Found member %s not requested to be stored.", user.Username)
		}
	}

	if len(testListAllMembersInput) != i {
		t.Errorf("Missing members when listing the members. %d != %d", len(testListAllMembersInput), i)
	}

}

var testAddMemberBuildResultInput = []struct {
	username string
	course   string
	builds   [][]int
}{
	{
		username: "resultuser1",
		course:   "resultscourse1",
		builds: [][]int{
			{
				1,
				2,
				3,
			},
			{
				4,
				5,
				6,
				7,
			},
			{
				8,
				9,
				10,
				11,
				12,
				13,
			},
			{
				14,
			},
		},
	},
	{
		username: "resultuser2",
		course:   "resultscourse2",
		builds: [][]int{
			{
				101,
				102,
				103,
			},
			{
				104,
				105,
				106,
				107,
			},
			{
				8,
				9,
				10,
				11,
				12,
				13,
			},
			{
				14,
			},
			{
				15,
				16,
				18,
				19,
				55,
				66,
				78,
			},
			{
				100,
				153,
				188,
			},
			{
				20000,
				22211,
			},
		},
	},
}

func TestAddAndGetMemberBuildResult(t *testing.T) {
	for _, in := range testAddMemberBuildResultInput {
		user, err := NewMemberFromUsername(in.username, true)
		if err != nil {
			t.Error(err)
			continue
		}

		org, err := NewOrganization(in.course, true)
		if err != nil {
			t.Error(err)
		}

		user.AddOrganization(org)

		for labnum, buildids := range in.builds {
			if user.GetLastBuildID(in.course, labnum) > 0 {
				t.Error("Found a build id before adding one")
			}

			for _, buildid := range buildids {
				user.AddBuildResult(in.course, labnum, buildid)
			}

			if _, ok := user.Courses[in.course]; !ok {
				t.Error("Missing course struct in user")
				continue
			}

			c := user.Courses[in.course]
			if len(c.Assignments[labnum].Builds) != len(buildids) {
				t.Errorf("The number of build IDs in group does not match number added. %d != %d",
					len(c.Assignments[labnum].Builds),
					len(buildids))
			}

			if user.GetLastBuildID(in.course, labnum) != buildids[len(buildids)-1] {
				t.Errorf("Build ID does not match last one added in GetLastBuildID. %d != %d",
					user.GetLastBuildID(in.course, labnum),
					buildids[len(buildids)-1])
			}
		}

	}
}

var testAddAndGetMemberNotesInput = []struct {
	username string
	course   string
	notes    [][]string
}{
	{
		username: "notesuser1",
		course:   "notescourse1",
		notes: [][]string{
			{
				"note 1",
				"note 2",
				"note 3",
			},
			{
				"note 4",
				"note 5",
				"note 6",
				"note 7 abcdefg",
			},
			{
				"note 8",
				"note 9",
				"note 10",
				"notes some thing something 11",
				"12",
				"note 13",
			},
			{
				"notes 14",
			},
		},
	},
	{
		username: "notesuser2",
		course:   "notescourse2",
		notes: [][]string{
			{
				"asvasfdasd asd",
				"aga gths t",
				"sdr gs dr dsr",
			},
			{
				"srd r ahrtsth",
				"dsfg y s sdf",
				"sdfg sdfsrtjhety",
				"note notes notes",
			},
			{
				"sdgfsgunlgrunlrueg",
				"arkønjaksjfnlakjsdmnrgu",
				"This is a real note",
				"akgrnøakgrøn",
				"æøåäè",
				"Good solution, but a bit bad implementation. Could have made the solution run faster",
			},
			{
				"heyhey",
			},
		},
	},
}

func TestAddAndGetMemberNotes(t *testing.T) {
	for _, in := range testAddAndGetMemberNotesInput {
		user, err := NewMemberFromUsername(in.username, true)
		if err != nil {
			t.Error(err)
			continue
		}

		org, err := NewOrganization(in.course, true)
		if err != nil {
			t.Error(err)
		}

		user.AddOrganization(org)

		for labnum, notes := range in.notes {
			if user.GetNotes(in.course, labnum) != "" {
				t.Error("Found a note before adding one")
			}

			for _, note := range notes {
				user.AddNotes(in.course, labnum, note)
			}

			if user.GetNotes(in.course, labnum) != notes[len(notes)-1] {
				t.Errorf("Build ID does not match last one added in GetLastBuildID. %s != %s",
					user.GetNotes(in.course, labnum),
					notes[len(notes)-1])
			}
		}
	}
}

var testMemberSetApprovedBuildInput = []struct {
	Course  string
	User    string
	Labnum  int
	BuildID int
	Date    time.Time
}{
	{
		Course:  "approvecourse1",
		User:    "approveuser1",
		Labnum:  1,
		BuildID: 2153,
		Date:    time.Date(2015, time.January, 12, 12, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
	{
		Course:  "approvecourse2",
		User:    "approveuser2",
		Labnum:  2,
		BuildID: 2483,
		Date:    time.Date(2015, time.January, 2, 2, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
	{
		Course:  "approvecourse3",
		User:    "approveuser3",
		Labnum:  4,
		BuildID: 21553,
		Date:    time.Date(2015, time.January, 1, 1, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
}

func TestMemberSetApprovedBuild(t *testing.T) {
	for _, in := range testMemberSetApprovedBuildInput {
		user, err := NewMemberFromUsername(in.User, true)
		if err != nil {
			t.Error(err)
			continue
		}

		user.Courses[in.Course] = NewCourseOptions(in.Course)

		user.SetApprovedBuild(in.Course, in.Labnum, in.BuildID, in.Date)

		opt := user.Courses[in.Course]

		if opt.Assignments[in.Labnum].ApproveDate != in.Date {
			t.Errorf("Approved date not set correctly. want %s, got %s for user %s",
				in.Date,
				opt.Assignments[in.Labnum].ApproveDate,
				in.User)
		}

		if opt.Assignments[in.Labnum].ApprovedBuild != in.BuildID {
			t.Errorf("Approved date not set correctly. want %d, got %d for user %s",
				in.BuildID,
				opt.Assignments[in.Labnum].ApprovedBuild,
				in.User)
		}
	}
}
