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
//   go test -v -tags github -run TestRepo
package entities

import "testing"

func TestRepo(t *testing.T) {
	gc, err := connect(mytoken)
	if err != nil {
		t.Errorf("Could not connect github user %v", err)
	}

	repolist, _, err := gc.Repositories.ListByOrg("uis-dat320", nil)
	for _, r := range repolist {
		repo, err := GetRepo(&r)
		if err != nil {
			t.Error(err)
		}
		t.Logf("r: %v\n", repo)
	}
}
