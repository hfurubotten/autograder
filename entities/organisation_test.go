package git

import (
	"strconv"
	"testing"
)

var testNewOrganizationAndSaveInput = []string{
	"course1",
	"course2",
	"uis-test",
	"course3",
	"course15-algorithms-fall2015",
}

func TestNewOrganizationAndSave(t *testing.T) {
	for _, orgid := range testNewOrganizationAndSaveInput {
		org, err := NewOrganization(orgid, false)
		if err != nil {
			t.Error("Error while creating a new Organization object:", err)
		}
		checkForOrganizationNilValues(org, t)

		org.AdminToken = "abcdef0123456789"
		org.Private = true
		org.GroupAssignments = 4
		org.IndividualAssignments = 2

		org.Save()
		registeredOrganizationList = append(registeredOrganizationList, orgid)

		org2, err := NewOrganization(orgid, true)
		if err != nil {
			t.Error("Error while creating a new Organization object:", err)
		}

		checkForOrganizationNilValues(org2, t)
		compareOrganizations(org, org2, t)
	}

	// checks again after clearing memory cache.
	for _, orgid := range testNewOrganizationAndSaveInput {
		org, err := NewOrganization(orgid, false)
		if err != nil {
			t.Error("Error while creating a new Organization object:", err)
		}
		checkForOrganizationNilValues(org, t)

		org.AdminToken = "0123456789abcdef"
		org.Private = false
		org.GroupAssignments = 5
		org.IndividualAssignments = 6

		org.Save()

		delete(InMemoryOrgs, orgid)

		org2, err := NewOrganization(orgid, true)
		if err != nil {
			t.Error("Error while creating a new Organization object:", err)
		}

		checkForOrganizationNilValues(org2, t)
		compareOrganizations(org, org2, t)
	}
}

var testAddGroupInput = []struct {
	Course  string
	GroupID int
}{
	{"course1", 1},
	{"course1", 2},
	{"course1", 3},
	{"course1", 4},
	{"course2", 5},
	{"course3", 6},
	{"course3", 7},
}

func TestAddGroup(t *testing.T) {
	for _, in := range testAddGroupInput {
		org, err := NewOrganization(in.Course, false)
		if err != nil {
			t.Error("Could not create a new course:", err)
		}
		g, err := NewGroup(in.Course, in.GroupID, true)
		if err != nil {
			t.Error("Could not create a new group:", err)
		}

		org.PendingGroup[in.GroupID] = nil

		org.AddGroup(g)
		org.Save()

		org2, err := NewOrganization(in.Course, true)
		if err != nil {
			t.Error("Could not create a new course:", err)
		}

		if _, ok := org2.PendingGroup[in.GroupID]; ok {
			t.Errorf("Cound find group with ID %d in the PendingGroup map", in.GroupID)
		}
		if _, ok := org2.Groups["group"+strconv.Itoa(in.GroupID)]; !ok {
			t.Errorf("Cound not find group with ID %d in the Group map", in.GroupID)
		}
	}
}

var registeredOrganizationList = []string{}

func TestHasOrganization(t *testing.T) {
	for _, orgid := range registeredOrganizationList {
		if !HasOrganization(orgid) {
			t.Errorf("Cant find the organization %v in database", orgid)
		}
	}
}

func TestListRegisteredOrganizations(t *testing.T) {
	list := ListRegisteredOrganizations()
	for _, orgid := range registeredOrganizationList {
		found := false
		for _, org := range list {
			if org.Name == orgid {
				found = true
			}
		}
		if !found {
			t.Errorf("Cant find the organization %v in organization list.", orgid)
		}
	}
}

func compareOrganizations(org1, org2 *Organization, t *testing.T) {
	if org1.Name != org2.Name {
		t.Errorf("Two organizations do not have equal Name field. %v != %v", org1.Name, org2.Name)
	}
	if org1.ScreenName != org2.ScreenName {
		t.Errorf("Two organizations do not have equal ScreenName field. %v != %v", org1.ScreenName, org2.ScreenName)
	}
	if org1.Description != org2.Description {
		t.Errorf("Two organizations do not have equal Description field. %v != %v", org1.Description, org2.Description)
	}
	if org1.Location != org2.Location {
		t.Errorf("Two organizations do not have equal Location field. %v != %v", org1.Location, org2.Location)
	}
	if org1.Company != org2.Company {
		t.Errorf("Two organizations do not have equal Company field. %v != %v", org1.Company, org2.Company)
	}
	if org1.HTMLURL != org2.HTMLURL {
		t.Errorf("Two organizations do not have equal HTMLURL field. %v != %v", org1.HTMLURL, org2.HTMLURL)
	}
	if org1.AvatarURL != org2.AvatarURL {
		t.Errorf("Two organizations do not have equal AvatarURL field. %v != %v", org1.AvatarURL, org2.AvatarURL)
	}
	if org1.GroupAssignments != org2.GroupAssignments {
		t.Errorf("Two organizations do not have equal GroupAssignments field. %v != %v", org1.GroupAssignments, org2.GroupAssignments)
	}
	if org1.IndividualAssignments != org2.IndividualAssignments {
		t.Errorf("Two organizations do not have equal IndividualAssignments field. %v != %v", org1.IndividualAssignments, org2.IndividualAssignments)
	}
	if org1.StudentTeamID != org2.StudentTeamID {
		t.Errorf("Two organizations do not have equal StudentTeamID field. %v != %v", org1.StudentTeamID, org2.StudentTeamID)
	}
	if org1.OwnerTeamID != org2.OwnerTeamID {
		t.Errorf("Two organizations do not have equal OwnerTeamID field. %v != %v", org1.OwnerTeamID, org2.OwnerTeamID)
	}
	if org1.Private != org2.Private {
		t.Errorf("Two organizations do not have equal Private field. %v != %v", org1.Private, org2.Private)
	}
	//if org1.GroupCount != org2.GroupCount {
	//	t.Errorf("Two organizations do not have equal GroupCount field. %v != %v", org1.GroupCount, org2.GroupCount)
	//}
	if org1.CodeReview != org2.CodeReview {
		t.Errorf("Two organizations do not have equal CodeReview field. %v != %v", org1.CodeReview, org2.CodeReview)
	}
	if org1.AdminToken != org2.AdminToken {
		t.Errorf("Two organizations do not have equal AdminToken field. %v != %v", org1.AdminToken, org2.AdminToken)
	}
	if org1.CI.Basepath != org2.CI.Basepath {
		t.Errorf("Two organizations do not have equal CI.Basepath field. %v != %v", org1.CI.Basepath, org2.CI.Basepath)
	}
	if org1.CI.Secret != org2.CI.Secret {
		t.Errorf("Two organizations do not have equal CI.Secret field. %v != %v", org1.CI.Secret, org2.CI.Secret)
	}

	// compares important mappers

}

func checkForOrganizationNilValues(org *Organization, t *testing.T) {
	if org.IndividualLabFolders == nil {
		t.Error("org.IndividualLabFolders cannot be nil after created.")
	}

	if org.GroupLabFolders == nil {
		t.Error("org.GroupLabFolders cannot be nil after created.")
	}

	if org.IndividualDeadlines == nil {
		t.Error("org.IndividualDeadlines cannot be nil after created.")
	}

	if org.GroupDeadlines == nil {
		t.Error("org.GroupDeadlines cannot be nil after created.")
	}

	if org.PendingGroup == nil {
		t.Error("org.PendingGroup cannot be nil after created.")
	}

	if org.PendingRandomGroup == nil {
		t.Error("org.PendingRandomGroup cannot be nil after created.")
	}

	if org.Groups == nil {
		t.Error("org.Groups cannot be nil after created.")
	}

	if org.PendingUser == nil {
		t.Error("org.PendingUser cannot be nil after created.")
	}

	if org.Members == nil {
		t.Error("org.Members cannot be nil after created.")
	}

	if org.Teachers == nil {
		t.Error("org.Teachers cannot be nil after created.")
	}

	if org.TotalScore == nil {
		t.Error("org.TotalScore cannot be nil after created.")
	}

	if org.WeeklyScore == nil {
		t.Error("org.WeeklyScore cannot be nil after created.")
	}

	if org.MonthlyScore == nil {
		t.Error("org.MonthlyScore cannot be nil after created.")
	}
}
