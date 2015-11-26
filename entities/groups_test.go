package entities

import (
	"testing"
	"time"
)

var testNewGroup = []struct {
	course string
	name   string
	gid    int
	want   *Group
}{
	{
		"course1",
		"hein's group",
		1,
		&Group{
			ID:            1,
			Course:        "course1",
			Name:          "hein's group",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		"course1",
		"nuugers",
		2,
		&Group{
			ID:            2,
			Course:        "course1",
			Name:          "nuugers",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		"course1",
		"suckers",
		3,
		&Group{
			ID:            3,
			Course:        "course1",
			Name:          "suckers",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		"course2",
		"groupers",
		4,
		&Group{
			ID:            4,
			Course:        "course2",
			Name:          "groupers",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		"course2",
		"pekers",
		5,
		&Group{
			ID:            5,
			Course:        "course2",
			Name:          "pekers",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		"course2",
		"the giants",
		6,
		&Group{
			ID:            6,
			Course:        "course2",
			Name:          "the giants",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
}

func TestNewGroupX(t *testing.T) {
	for _, in := range testNewGroup {
		g := NewGroup(in.course, in.name)
		compareGroups(g, in.want, t)
		err := g.Save()
		if err != nil {
			t.Error(err)
			continue
		}
	}
	// check if newly created groups can be retrived from database
	for _, in := range testNewGroup {
		g, err := GetGroup(in.name)
		if err != nil {
			t.Error(err)
		}
		compareGroups(g, in.want, t)
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
			Name:          "7",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
		&Group{
			ID:            7,
			Course:        "course1",
			Name:          "7",
			CurrentLabNum: 1,
			Active:        true,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            8,
			Course:        "course1",
			Name:          "8",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user1": nil,
				"user2": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
		&Group{
			ID:            8,
			Course:        "course1",
			Name:          "8",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user1": nil,
				"user2": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            9,
			Course:        "course1",
			Name:          "9",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user3": nil,
				"user4": nil,
				"user5": nil,
				"user6": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
		&Group{
			ID:            9,
			Course:        "course1",
			Name:          "9",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user3": nil,
				"user4": nil,
				"user5": nil,
				"user6": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
}

func TestActivate(t *testing.T) {
	for _, tcase := range testActivate {
		for username := range tcase.want.Members {
			u, err := CreateMember(username)
			if err != nil {
				t.Errorf("Error creating member: %v", err)
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
			u, err := GetMember(username)
			if err != nil {
				t.Errorf("Error getting user: %v", err)
				continue
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
			Name:          "10",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
		[]string{"user7"},
		&Group{
			ID:            10,
			Course:        "course1",
			Name:          "10",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user7": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            12,
			Course:        "course1",
			Name:          "12",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
		[]string{"user8", "user9"},
		&Group{
			ID:            12,
			Course:        "course1",
			Name:          "12",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user8": nil,
				"user9": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            13,
			Course:        "course1",
			Name:          "13",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user10": nil,
				"user11": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
		[]string{"user12", "user13"},
		&Group{
			ID:            13,
			Course:        "course1",
			Name:          "13",
			CurrentLabNum: 1,
			Active:        false,
			Members: map[string]interface{}{
				"user10": nil,
				"user11": nil,
				"user12": nil,
				"user13": nil,
			},
			Assignments: make(map[int]*Assignment),
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
			Name:          "21",
			CurrentLabNum: 1,
			Active:        false,
			Members:       make(map[string]interface{}),
			Assignments:   make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            22,
			Course:        "course1",
			Name:          "22",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user14": nil,
				"user15": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
	{
		&Group{
			ID:            23,
			Course:        "course1",
			Name:          "23",
			CurrentLabNum: 1,
			Active:        true,
			Members: map[string]interface{}{
				"user16": nil,
				"user17": nil,
				"user18": nil,
				"user19": nil,
			},
			Assignments: make(map[int]*Assignment),
		},
	},
}

func TestSaveHasAndDelete(t *testing.T) {
	for _, tcase := range testSaveHasAndDelete {
		for username := range tcase.in.Members {
			u, err := CreateMember(username)
			if err != nil {
				t.Errorf("Error creating member: %v", err)
				continue
			}
			org, err := NewOrganization(tcase.in.Course, true)
			if err != nil {
				t.Errorf("Error creating org: %v", err)
				continue
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
		nextID, err := GetNextGroupID()
		if err != nil {
			t.Error(err)
		}
		if nextID != i {
			t.Errorf("Error with counting in getting next group ID. Got %d, want %d.", nextID, i)
		}
	}
}

var testAddGroupBuildResultInput = []struct {
	groupName string
	builds    [][]int
}{
	{
		groupName: "Wierd al Yankovic",
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
		groupName: "Adele",
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

func TestAddAndGetGroupBuildResult(t *testing.T) {
	for _, in := range testAddGroupBuildResultInput {
		group := NewGroup("", in.groupName)

		for labnum, buildids := range in.builds {
			if group.GetLastBuildID(labnum) > 0 {
				t.Error("Found a build id before adding one")
			}

			for _, buildid := range buildids {
				group.AddBuildResult(labnum, buildid)
			}

			if len(group.Assignments[labnum].Builds) != len(buildids) {
				t.Errorf("The number of build IDs in group does not match number added. %d != %d",
					len(group.Assignments[labnum].Builds),
					len(buildids))
			}

			if group.GetLastBuildID(labnum) != buildids[len(buildids)-1] {
				t.Errorf("Build ID does not match last one added in GetLastBuildID. %d != %d",
					group.GetLastBuildID(labnum),
					buildids[len(buildids)-1])
			}
		}

	}
}

var testAddAndGetGroupNotesInput = []struct {
	groupName string
	notes     [][]string
}{
	{
		groupName: "big 26",
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
		groupName: "small just now",
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

func TestAddAndGetGroupNotes(t *testing.T) {
	for _, in := range testAddAndGetGroupNotesInput {
		group := NewGroup("jungle course", in.groupName)

		for labnum, notes := range in.notes {
			if group.GetNotes(labnum) != "" {
				t.Error("Found a note before adding one")
			}

			for _, note := range notes {
				group.AddNotes(labnum, note)
			}

			if group.GetNotes(labnum) != notes[len(notes)-1] {
				t.Errorf("Build ID does not match last one added in GetLastBuildID. %s != %s",
					group.GetNotes(labnum),
					notes[len(notes)-1])
			}
		}
	}
}

var testGroupSetApprovedBuildInput = []struct {
	Course  string
	Name    string
	Group   int
	Labnum  int
	BuildID int
	Date    time.Time
}{
	{
		Course:  "approvecourse4",
		Name:    "1051",
		Group:   1051,
		Labnum:  1,
		BuildID: 2153,
		Date:    time.Date(2015, time.January, 12, 12, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
	{
		Course:  "approvecourse5",
		Name:    "5553",
		Group:   5553,
		Labnum:  2,
		BuildID: 2483,
		Date:    time.Date(2015, time.January, 2, 2, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
	{
		Course:  "approvecourse6",
		Name:    "4579",
		Group:   4579,
		Labnum:  4,
		BuildID: 21553,
		Date:    time.Date(2015, time.January, 1, 1, 12, 12, 0, time.FixedZone("unnamed", 1)),
	},
	{
		Course:  "approvecourse7",
		Name:    "579",
		Group:   579,
		Labnum:  6,
		BuildID: 2153,
		Date:    time.Date(2015, time.January, 10, 1, 1, 12, 0, time.FixedZone("unnamed", 1)),
	},
}

func TestGroupSetApprovedBuild(t *testing.T) {
	for _, in := range testGroupSetApprovedBuildInput {
		group := NewGroup(in.Course, in.Name)

		group.SetApprovedBuild(in.Labnum, in.BuildID, in.Date)
		if group.Assignments[in.Labnum].ApproveDate != in.Date {
			t.Errorf("Approved date not set correctly. want %s, got %s for user %d",
				in.Date,
				group.Assignments[in.Labnum].ApproveDate,
				in.Group)
		}
		if group.Assignments[in.Labnum].ApprovedBuild != in.BuildID {
			t.Errorf("Approved date not set correctly. want %d, got %d for user %d",
				in.BuildID,
				group.Assignments[in.Labnum].ApprovedBuild,
				in.Group)
		}
	}
}

func compareGroups(in, want *Group, t *testing.T) {
	if in.Active != want.Active {
		t.Errorf("Active field; got %t, want %t", in.Active, want.Active)
	}
	if in.Course != want.Course {
		t.Errorf("Course field; got %s, want %s", in.Course, want.Course)
	}
	if in.CurrentLabNum != want.CurrentLabNum {
		t.Errorf("CurrentLabNum field; got %d, want %d", in.CurrentLabNum, want.CurrentLabNum)
	}
	if in.Name != want.Name {
		t.Errorf("Name field; got %d, want %d", in.Name, want.Name)
	}
	// if in.ID != want.ID {
	// 	t.Errorf("ID field; got %d, want %d", in.ID, want.ID)
	// }
	if in.TeamID != want.TeamID {
		t.Errorf("TeamID field; got %d, want %d", in.TeamID, want.TeamID)
	}
	if in.Members == nil || want.Members == nil {
		t.Error("Members field cannot be nil")
		return
	}
	for username := range in.Members {
		if _, ok := want.Members[username]; !ok {
			t.Errorf("Unwanted member %s in member list", username)
		}
	}
	for username := range want.Members {
		if _, ok := in.Members[username]; !ok {
			t.Errorf("Missing member %s in member list", username)
		}
	}
	if in.Assignments == nil || want.Assignments == nil {
		t.Error("Assignments field cannot be nil")
		return
	}
	if len(in.Assignments) != len(want.Assignments) {
		t.Errorf("Missing assignments in the group, got %d, want %d", len(in.Assignments), len(want.Assignments))
	}
}
