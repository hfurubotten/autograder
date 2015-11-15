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
package auth

import (
	"testing"

	"github.com/hfurubotten/autograder/entities"
)

func TestGetGitHubUser(t *testing.T) {
	gc, err := connect(mytoken)
	if err != nil {
		t.Errorf("Could not connect github user %v", err)
	}
	gu, xu, err := gc.Users.Get("")
	if err != nil {
		t.Errorf("Could not get github user %v", err)
	}
	t.Logf("g: %v\nx: %v\n", gu, xu)
}

func TestLookupAndNewMember(t *testing.T) {
	m, err := entities.LookupMember(mytoken)
	if err == nil || m != nil {
		t.Errorf("Expected error, but found member: %v", m)
	}

	scope := "admin:org,repo,admin:repo_hook"
	u, err := githubUserProfile(mytoken, scope)
	if err != nil {
		t.Errorf("Failed to get user from GitHub: %v", err)
	}

	m = entities.NewMember(u)
	err = entities.PutMember(mytoken, m)
	if err != nil || m == nil {
		t.Errorf("Expected member, but got error: %v", err)
	}

	m, err = entities.LookupMember(mytoken)
	if err != nil {
		t.Errorf("Expected member with access token: %s, but got: %s", mytoken, err)
	}
	if m.GetToken() != mytoken {
		t.Errorf("Expected token: %s, but got: %s", mytoken, m.GetToken())
	}
	t.Logf("m: %v\n", m)

	// remove member inserted into database; it won't be needed in other tests
	err = m.RemoveMember()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
