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

func TestGetGithubMember(t *testing.T) {
	// gu := &github.User{}
	gc, err := connect(mytoken)
	if err != nil {
		t.Errorf("Could not connect github user %v", err)
	}
	gu, xu, err := gc.Users.Get("")
	if err != nil {
		t.Errorf("Could not get github user %v", err)
	}
	t.Logf("g: %v, x: %v\n", gu, xu)

	// m, err := NewUserWithGithubData(gu)
	// if err != nil {
	// 	t.Errorf("Could not create github user %v", err)
	// }
	// t.Logf("m: %v\n", m)
}
