package git

import (
	"net/mail"
	"os"
	"testing"

	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/github-gamification/entities"
)

func cleanUpMemberStorage() error {
	if err := os.RemoveAll(global.Basepath + "diskv/users/"); err != nil {
		return err
	}
	return os.RemoveAll(global.Basepath + "diskv/tokens/")
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
				t.Errorf("Error creating org", err)
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
				t.Errorf("Error creating org", err)
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
				t.Errorf("Error creating org", err)
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
		t.Error("Missing members when listing the members. %d != %d", len(testListAllMembersInput), i)
	}

	cleanUpMemberStorage()
}
