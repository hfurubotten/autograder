package web

import (
	"testing"

	git "github.com/hfurubotten/autograder/entities"
)

var distributeScoresTest = []struct {
	inUser    string
	inScore   int
	wantScore int64
}{
	{"user1", 20, 20},
	{"user1", 21, 41},
	{"user1", 20, 61},
}

// Does not test if stored correctly, only the point calculation.
func TestDistributeScores(t *testing.T) {
	org, err := git.NewOrganizationX("testorg")
	if err != nil {
		t.Error(err)
		return
	}

	for _, dst := range distributeScoresTest {
		if !git.HasMember(dst.inUser) {
			_, err := git.CreateMember(dst.inUser)
			if err != nil {
				t.Error(err)
				continue
			}
		}
	}

	for _, dst := range distributeScoresTest {
		user, err := git.GetMember(dst.inUser)
		if err != nil {
			t.Error(err)
			continue
		}

		err = DistributeScores(dst.inScore, user, org)
		if err != nil {
			t.Error(err)
			continue
		}
		if user.TotalScore != dst.wantScore {
			t.Errorf("Want score %d for %s, but got %d.", dst.wantScore, dst.inUser, user.TotalScore)
		}
		if org.GetUserScore(dst.inUser) != dst.wantScore {
			t.Errorf("Want score %d for %s in testorg, but got %d.", dst.wantScore, dst.inUser, org.GetUserScore(dst.inUser))
		}
	}
}

func TestNilScore(t *testing.T) {
	err := DistributeScores(0, nil, nil)
	if err == nil {
		t.Error("Expected error due to nil input")
	}
}
