package git

import (
	"testing"
)

var testNewGroup = []struct {
	inCourse string
	inGID    int
	want     *Group
}{
	{
		"course1",
		1,
		&Group{
			ID:            1,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		"course1",
		2,
		&Group{
			ID:            2,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		"course1",
		3,
		&Group{
			ID:            3,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		"course2",
		4,
		&Group{
			ID:            4,
			Course:        "course2",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		"course2",
		5,
		&Group{
			ID:            5,
			Course:        "course2",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		"course2",
		6,
		&Group{
			ID:            6,
			Course:        "course2",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
}

func TestNewGroup(t *testing.T) {
	for _, tcase := range testNewGroup {
		// test a not known group
		g1, err := NewGroup(tcase.inCourse, tcase.inGID, false)
		if err != nil {
			t.Errorf("Error creating group: %v", err)
			continue
		}

		compareGroups(g1, tcase.want, t)

		err = g1.Save()
		if err != nil {
			t.Errorf("Error saving group: %v", err)
			continue
		}

		// test when known
		g2, err := NewGroup(tcase.inCourse, tcase.inGID, true)
		if err != nil {
			t.Errorf("Error creating group: %v", err)
		}

		compareGroups(g1, g2, t)
	}
}

var testActivate = []struct {
	in   *Group
	want *Group
}{
	{
		&Group{
			ID:            7,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
		&Group{
			ID:            7,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        true,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            8,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user1": nil,
				"user2": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
		&Group{
			ID:            8,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user1": nil,
				"user2": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            9,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user3": nil,
				"user4": nil,
				"user5": nil,
				"user6": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
		&Group{
			ID:            9,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user3": nil,
				"user4": nil,
				"user5": nil,
				"user6": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
}

func TestActivate(t *testing.T) {
	for _, tcase := range testActivate {
		for username := range tcase.want.Members {
			u, err := NewMemberFromUsername(username, false)
			if err != nil {
				t.Errorf("Error getting user: %v", err)
				continue
			}

			org, err := NewOrganization(tcase.want.Course, true)
			if err != nil {
				t.Errorf("Error creating org: %v", err)
				continue
			}

			u.AddOrganization(org)
			u.Save()
			testListAllMembersInput = append(testListAllMembersInput, username)
		}

		tcase.in.Activate()

		compareGroups(tcase.in, tcase.want, t)

		for username := range tcase.want.Members {
			u, err := NewMemberFromUsername(username, true)
			if err != nil {
				t.Errorf("Error getting user: %v", err)
			}

			if u.Courses[tcase.want.Course].GroupNum != tcase.want.ID {
				t.Errorf("User not updated with correct group ID. Got %d, want %d.", u.Courses[tcase.want.Course].GroupNum, tcase.want.ID)
			}

			if !u.Courses[tcase.want.Course].IsGroupMember {
				t.Errorf("User not updated with group membership. Got %t for IsGroupMember field, want true.", u.Courses[tcase.want.Course].IsGroupMember)
			}
		}
	}
}

var testAddMember = []struct {
	in      *Group
	inUsers []string
	want    *Group
}{
	{
		&Group{
			ID:            10,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
		[]string{"user7"},
		&Group{
			ID:            10,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user7": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            12,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
		[]string{"user8", "user9"},
		&Group{
			ID:            12,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user8": nil,
				"user9": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            13,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user10": nil,
				"user11": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
		[]string{"user12", "user13"},
		&Group{
			ID:            13,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user10": nil,
				"user11": nil,
				"user12": nil,
				"user13": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
}

func TestAddMember(t *testing.T) {
	for _, tcase := range testAddMember {
		for _, username := range tcase.inUsers {
			tcase.in.AddMember(username)
		}

		compareGroups(tcase.in, tcase.want, t)
	}
}

var testSaveHasAndDelete = []struct {
	in *Group
}{
	{
		&Group{
			ID:            21,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            22,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user14": nil,
				"user15": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
	{
		&Group{
			ID:            23,
			Course:        "course1",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user16": nil,
				"user17": nil,
				"user18": nil,
				"user19": nil,
			},
			Assignments: make(map[int]*LabAssignmentOptions),
		},
	},
}

func TestSaveHasAndDelete(t *testing.T) {
	for _, tcase := range testSaveHasAndDelete {
		for username := range tcase.in.Members {
			u, err := NewMemberFromUsername(username, false)
			if err != nil {
				t.Errorf("Error getting user: %v", err)
			}

			org, err := NewOrganization(tcase.in.Course, true)
			if err != nil {
				t.Errorf("Error creating org: %v", err)
			}

			u.AddOrganization(org)
			u.Save()
			testListAllMembersInput = append(testListAllMembersInput, username)
		}

		tcase.in.Lock()
		tcase.in.Activate()
		err := tcase.in.Save()
		if err != nil {
			t.Errorf("Error saving the group: %v", err)
		}

		if !HasGroup(tcase.in.ID) {
			t.Error("Couldnt find the group after save.")
		}

		err = tcase.in.Delete()
		if err != nil {
			t.Errorf("Error deleting the group: %v", err)
		}

		if HasGroup(tcase.in.ID) {
			t.Error("Found the group after save.")
		}
	}
}

var testGetNextGroupIDIterations = 100

func TestGetNextGroupID(t *testing.T) {
	for i := 1; i <= testGetNextGroupIDIterations; i++ {
		nextID := GetNextGroupID()
		if nextID != i {
			t.Errorf("Error with counting in getting next group ID. Got %d, want %d.", nextID, i)
		}
	}
}

func compareGroups(in, want *Group, t *testing.T) {
	if in.Active != want.Active {
		t.Errorf("Error while comparing groups: got %t for active field value, want %t", in.Active, want.Active)
	}

	if in.Course != want.Course {
		t.Errorf("Error while comparing groups: got %s for active field value, want %s", in.Course, want.Course)
	}

	if in.CurrentLabNum != want.CurrentLabNum {
		t.Errorf("Error while comparing groups: got %d for active field value, want %d", in.CurrentLabNum, want.CurrentLabNum)
	}

	if in.ID != want.ID {
		t.Errorf("Error while comparing groups: got %d for active field value, want %d", in.ID, want.ID)
	}

	if in.TeamID != want.TeamID {
		t.Errorf("Error while comparing groups: got %d for active field value, want %d", in.TeamID, want.TeamID)
	}

	if in.Members == nil || want.Members == nil {
		t.Error("store field cannot be nil.")
	}

	if in.Members == nil || want.Members == nil {
		t.Error("Members field cannot be nil.")
		return
	}

	for username := range in.Members {
		if _, ok := want.Members[username]; !ok {
			t.Errorf("Unwanted member %s in member list.", username)
		}
	}

	for username := range want.Members {
		if _, ok := in.Members[username]; !ok {
			t.Errorf("Missing member %s in member list.", username)
		}
	}

	if in.Assignments == nil || want.Assignments == nil {
		t.Error("Assignments field cannot be nil.")
		return
	}

	if len(in.Assignments) != len(want.Assignments) {
		t.Errorf("Not enough assignments in the group, got length %d, want %d", len(in.Assignments), len(want.Assignments))
	}
}
