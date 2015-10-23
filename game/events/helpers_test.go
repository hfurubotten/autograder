package events

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

var userList = make(map[string]*git.UserProfile)

// Does not test if stored correctly, only the point calculation.
func TestDistributeScores(t *testing.T) {
	org, err := git.NewOrganizationX("testorg")
	if err != nil {
		t.Error("Failed to open new org:", err)
		return
	}

	for _, dst := range distributeScoresTest {
		user, ok := userList[dst.inUser]
		if !ok {
			user, err = git.NewUser(dst.inUser)
			if err != nil {
				t.Error("Failed to open new user:", err)
				continue
			}
			userList[dst.inUser] = user
		}

		err = DistributeScores(dst.inScore, user, org)
		if err != nil {
			t.Error(err)
		}

		if user.TotalScore != dst.wantScore {
			t.Errorf("Want score %d for %s, but got %d.", dst.wantScore, dst.inUser, user.TotalScore)
		}

		if org.GetUserScore(dst.inUser) != dst.wantScore {
			t.Errorf("Want score %d for %s in testorg, but got %d.", dst.wantScore, dst.inUser, org.GetUserScore(dst.inUser))
		}
	}

	// // Cleans up the saved obj
	// entities.GetUserStore().Erase("user1")
	// entities.GetRepoStore("testorg").Erase("testrepo")
	// entities.GetOrganizationStore().Erase("testorg")

	// checks panic on nil user value
	defer PanicHandler(false)

	DistributeScores(0, nil, nil)
}

func TestPanicHandler(t *testing.T) {
	defer PanicHandler(false)

	panic("This is the test. Fails if this panic goes through...")
}
