// +build github

// Leave an empty line above this comment.
// To test with github enabled, make sure to set the variable 'mytoken'
// to a personal access token obtained from the github settings page.
//
//   cp mytoken_test.go mytoken_personal_test.go
//    << edit mytoken_personal_test.go adding your personal token >>
//
// To run the github dependent tests use the following:
//   go test -v -tags github
// Or:
//   go test -v -tags github -run TestGetGithubMember
package entities

import "testing"

func TestGetGithubMember(t *testing.T) {
	gc, err := connect(mytoken)
	if err != nil {
		t.Errorf("Could not connect github user %v", err)
	}
	gu, xu, err := gc.Users.Get("")
	if err != nil {
		t.Errorf("Could not get github user %v", err)
	}
	t.Logf("g: %v\nx: %v\n", gu, xu)

	m, err := GetMember(*gu.Login)
	if err != nil {
		t.Errorf("Could not create member: %v", err)
	}
	if m.Username != *gu.Login {
		t.Errorf("expected: %s, but got: %s", *gu.Login, m.Username)
	}
}

func TestLookupAndNewMember(t *testing.T) {
	m, err := LookupMember(mytoken)
	if err == nil || m != nil {
		t.Errorf("Expected error, but found member: %v", m)
	}
	m, err = NewMember(mytoken)
	if err != nil || m == nil {
		t.Errorf("Expected member, but got error: %v", err)
	}
	if m.accessToken != mytoken {
		t.Errorf("Expected member with access token: %s, but got: %s", mytoken, m.accessToken)
	}

	// we should already be connected by previous call to connect() from NewMember
	gu, err := getGithubUser(m.githubclient)
	if err != nil {
		t.Errorf("Expected github user, but got error: %v", err)
	}
	if m.Username != *gu.Login {
		t.Errorf("Expected member with Username: %s, but got: %s", m.Username, *gu.Login)
	}

	// remove member inserted into database; it won't be needed in other tests
	err = m.RemoveMember()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
