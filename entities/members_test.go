package git

import (
	"net/mail"
	"testing"

	"github.com/hfurubotten/github-gamification/entities"
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
