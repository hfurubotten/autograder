package entities

import (
	"net/mail"
	"testing"
	"time"
)

var testNewMemberInput = []struct {
	token    string
	username string
	scope    string
	studid   int
}{
	{
		"123456789abcdef",
		"user100",
		"admin:org,repo,admin:repo_hook",
		123456789,
	},
	{
		"12345674489abcdef",
		"user101",
		"admin:org,repo,admin:repo_hook",
		987654321,
	},
	{
		"12345655789abcdef",
		"user102",
		"admin%3Aorg%2Crepo%2Cadmin%3Arepo_hook",
		156789,
	},
}

func TestNewMember(t *testing.T) {
	for _, in := range testNewMemberInput {
		u := NewUserProfile(in.token, in.username, in.scope)
		m := NewMember(u)
		err := PutMember(in.token, m)
		if err != nil {
			t.Errorf("Failed to create member with token (%s): %v", in.token, err)
		}
		if m.accessToken != in.token {
			t.Errorf("Access token mismatch: %s, got: %s", in.token, m.accessToken)
		}
		if m.Username != in.username {
			t.Errorf("Username mismatch: %v, got: %v", in.username, m.Username)
		}
		if m.Scope != in.scope {
			t.Errorf("Scope mismatch: %v, got: %v", in.scope, m.Scope)
		}
		// clean up database
		err = m.RemoveMember()
		if err != nil {
			t.Errorf("Failed to remove member: %v", err)
		}
	}
}

func TestNewMemberAlreadyInDatabase(t *testing.T) {
	for _, in := range testNewMemberInput {
		// tweak to ensure that user is already believed to be in database
		if err := putToken(in.token, in.username); err != nil {
			t.Errorf("Failed to store token for user (%s): %v", in.username, err)
			continue
		}
		u := NewUserProfile(in.token, in.username, in.scope)
		m := NewMember(u)
		err := PutMember(in.token, m)
		if err == nil {
			t.Errorf("Unexpected member creation with token (%s): %v", in.token, err)
		}
		// clean up database
		err = removeToken(in.token)
		if err != nil {
			t.Errorf("Failed to remove token: %v", err)
		}
	}
}

func TestLookupMemberBasic(t *testing.T) {
	m, err := LookupMember("some unexpected token")
	if err == nil || m != nil {
		t.Errorf("expected error, but got member: %v", m)
	}
	m, err = LookupMember("")
	if err == nil || m != nil {
		t.Errorf("expected error, but got member: %v", m)
	}

	mytoken := "mytoken"
	userName := "jamesbond"
	m, err = CreateMember(userName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if err = putToken(mytoken, userName); err != nil {
		t.Errorf("Error storing token for '%s': %v", userName, err)
	}
	u, err := getToken(mytoken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if u != userName {
		t.Errorf("expected user: %s, got: %s", userName, u)
	}

	m, err = LookupMember(mytoken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	// tweak to emulate github lookup
	m.accessToken = mytoken

	// remove member inserted into database; it won't be needed in other tests
	err = m.RemoveMember()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLookupMember(t *testing.T) {
	for _, in := range testNewMemberInput {
		u := NewUserProfile(in.token, in.username, in.scope)
		m := NewMember(u)
		err := PutMember(in.token, m)
		if err != nil {
			t.Error("Error storing new member: ", err)
			continue
		}

		m, err = LookupMember(in.token)
		if err != nil {
			t.Error("Error looking up member using token: ", err)
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

		m2, err := LookupMember(in.token)
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

		m3, err := LookupMember(in.token)
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

		// checks again with loading from DB
		m4, err := LookupMember(in.token)
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

func BenchmarkNewMember(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, in := range testNewMemberInput {
			removeToken(in.token)
			u := NewUserProfile(in.token, in.username, in.scope)
			m := NewMember(u)
			err := PutMember(in.token, m)
			if err != nil {
				b.Error(err)
				continue
			}
		}
	}
}

func BenchmarkLookupMember(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, in := range testNewMemberInput {
			LookupMember(in.token)
		}
	}
}

func BenchmarkLookupMemberAndSave(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, in := range testNewMemberInput {
			m, err := LookupMember(in.token)
			if err != nil {
				b.Error(err)
				continue
			}
			m.StudentID = in.studid
			m.Save()
		}
	}
}

var testIsComplete = []struct {
	in   *Member
	want bool
}{
	{
		&Member{
			UserProfile: &UserProfile{
				Name: "Ola Normann",
			},
		},
		false,
	},
	{
		&Member{
			UserProfile: &UserProfile{
				Name: "Ola Normann",
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			UserProfile: &UserProfile{
				Name:     "Ola Normann",
				Username: "olanormann",
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			UserProfile: &UserProfile{
				Name:     "Ola Normann",
				Username: "olanormann",
			},
		},
		false,
	},
	{
		&Member{
			UserProfile: &UserProfile{
				Username: "olanormann",
				Email:    &mail.Address{},
			},
			StudentID: 112222,
		},
		false,
	},
	{
		&Member{
			UserProfile: &UserProfile{
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
				if opt.CourseName != cname {
					t.Errorf("In course options for %s got wrong course name. Got %s", cname, opt.CourseName)
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
		user, err := GetMember(username)
		if err != nil {
			t.Error(err)
		}
		user.Save()
	}

	list, err := ListAllMembers()
	if err != nil {
		t.Error(err)
	}
	for _, user := range list {
		found := false
		for _, username := range testListAllMembersInput {
			if username == user.Username {
				found = true
			}
		}
		if !found {
			t.Errorf("Found unexpected member: %s", user.Username)
		}
	}
	if len(testListAllMembersInput) != len(list) {
		t.Errorf("Expected: %d, got: %d members", len(testListAllMembersInput), len(list))
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
		_, err := CreateMember(in.username)
		if err != nil {
			t.Error(err)
			continue
		}
	}

	for _, in := range testAddMemberBuildResultInput {
		user, err := GetMember(in.username)
		if err != nil {
			t.Error(err)
			continue
		}
		org, err := NewOrganization(in.course, true)
		if err != nil {
			t.Error(err)
			continue
		}
		user.AddOrganization(org)

		for labnum, buildids := range in.builds {
			if user.GetLastBuildID(in.course, labnum) > 0 {
				t.Error("Found a build id before adding one")
				continue
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
				continue
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
		_, err := CreateMember(in.username)
		if err != nil {
			t.Error(err)
			continue
		}
	}
	for _, in := range testAddAndGetMemberNotesInput {
		user, err := GetMember(in.username)
		if err != nil {
			t.Error(err)
			continue
		}
		org, err := NewOrganization(in.course, true)
		if err != nil {
			t.Error(err)
			continue
		}
		user.AddOrganization(org)

		for labnum, notes := range in.notes {
			if user.GetNotes(in.course, labnum) != "" {
				t.Error("Found a note before adding one")
				continue
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
		_, err := CreateMember(in.User)
		if err != nil {
			t.Error(err)
			continue
		}
	}
	for _, in := range testMemberSetApprovedBuildInput {
		user, err := GetMember(in.User)
		if err != nil {
			t.Error(err)
			continue
		}
		user.Courses[in.Course] = NewCourse(in.Course)
		user.SetApprovedBuild(in.Course, in.Labnum, in.BuildID, in.Date)
		opt := user.Courses[in.Course]

		if opt.Assignments[in.Labnum].ApproveDate != in.Date {
			t.Errorf("Approved date not set correctly. want %s, got %s for user %s",
				in.Date,
				opt.Assignments[in.Labnum].ApproveDate,
				in.User)
			continue
		}
		if opt.Assignments[in.Labnum].ApprovedBuild != in.BuildID {
			t.Errorf("Approved date not set correctly. want %d, got %d for user %s",
				in.BuildID,
				opt.Assignments[in.Labnum].ApprovedBuild,
				in.User)
		}
	}
}
