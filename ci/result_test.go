package ci

import (
	"log"
	"os"
	"testing"

	"github.com/hfurubotten/autograder/database"
)

func TestMain(m *testing.M) {
	if err := database.Start("test.db"); err != nil {
		log.Println("Failed to start testing, error while setting up the database:", err)
		return
	}

	m.Run()

	if err := database.Close(); err != nil {
		log.Println("Failed to close database after testing:", err)
	}
	if err := os.RemoveAll("test.db"); err != nil {
		log.Println("Unable to clean up database file from filesystem")
	}
}

var iter = 100

func TestNewBuildResult(t *testing.T) {
	first, err := NewBuildResult()
	if err != nil {
		t.Error(err)
	}
	for i := first.ID + 1; i <= iter; i++ {
		build, err := NewBuildResult()
		if err != nil {
			t.Error("Error creating a new build result object:", err)
		}
		if build.ID != i {
			t.Errorf("Error generating build ID. Got %d, wanted %d.", build.ID, i)
		}
		if build.TestScores == nil {
			t.Error("Field TestScores cannot be nil")
		}
		if build.Log == nil {
			t.Error("Field Log cannot be nil")
		}
	}
}

func TestConcurrentNewBuildResult(t *testing.T) {
	first, err := NewBuildResult()
	if err != nil {
		t.Error(err)
	}
	for i := first.ID + 1; i <= iter; i++ {
		go func() {
			build, err := NewBuildResult()
			if err != nil {
				t.Error("Error creating a new build result object:", err)
			}
			if build.ID != i {
				t.Errorf("Error generating build ID. Got %d, wanted %d.", build.ID, i)
			}
			if build.TestScores == nil {
				t.Error("Field TestScores cannot be nil")
			}
			if build.Log == nil {
				t.Error("Field Log cannot be nil")
			}
		}()
	}
}

var testGetAndSaveBuildResultInput = []*BuildResult{
	&BuildResult{
		ID:        1,
		Course:    "course1",
		User:      "user1",
		Group:     -1,
		NumPasses: 20,
		NumFails:  10,
		Labnum:    2,
	},
	&BuildResult{
		ID:        2,
		Course:    "course2",
		User:      "user1",
		Group:     -1,
		NumPasses: 20,
		NumFails:  10,
		Labnum:    2,
	},
	&BuildResult{
		ID:        3,
		Course:    "course1",
		User:      "user2",
		Group:     -1,
		NumPasses: 20,
		NumFails:  10,
		Labnum:    2,
	},
	&BuildResult{
		ID:        4,
		Course:    "course1",
		User:      "user1",
		Group:     -1,
		NumPasses: 23,
		NumFails:  1,
		Labnum:    3,
	},
	&BuildResult{
		ID:        5,
		Course:    "course1",
		Group:     22,
		NumPasses: 20,
		NumFails:  10,
		Labnum:    2,
	},
	&BuildResult{
		ID:        6,
		Course:    "course1",
		Group:     65,
		NumPasses: 20,
		NumFails:  52,
		Labnum:    2,
	},
	&BuildResult{
		ID:        7,
		Course:    "course4",
		User:      "user1",
		Group:     -1,
		NumPasses: 21,
		NumFails:  10,
		Labnum:    5,
	},
}

func TestGetAndSaveBuildResult(t *testing.T) {
	for _, br := range testGetAndSaveBuildResultInput {
		if err := br.Save(); err != nil {
			t.Error("Failed to save build result: ", err)
			continue
		}

		br2, err := GetBuildResult(br.ID)
		if err != nil {
			t.Error("Failed to get a build result from DB: ", err)
			continue
		}
		compareBuildResults(br, br2, t)
	}
}

func compareBuildResults(br1, br2 *BuildResult, t *testing.T) {
	if br1.ID != br2.ID {
		t.Errorf("Field values for ID does not match. %v != %v.", br1.ID, br2.ID)
	}
	if br1.Course != br2.Course {
		t.Errorf("Field values for Course does not match. %v != %v.", br1.Course, br2.Course)
	}
	if br1.User != br2.User {
		t.Errorf("Field values for User does not match. %v != %v.", br1.User, br2.User)
	}
	if br1.NumPasses != br2.NumPasses {
		t.Errorf("Field values for NumPasses does not match. %v != %v.", br1.NumPasses, br2.NumPasses)
	}
	if br1.NumFails != br2.NumFails {
		t.Errorf("Field values for NumFails does not match. %v != %v.", br1.NumFails, br2.NumFails)
	}
	if br1.NumBuildFailure != br2.NumBuildFailure {
		t.Errorf("Field values for NumBuildFailure does not match. %v != %v.", br1.NumBuildFailure, br2.NumBuildFailure)
	}
	if br1.Status != br2.Status {
		t.Errorf("Field values for Status does not match. %v != %v.", br1.Status, br2.Status)
	}
	if br1.Labnum != br2.Labnum {
		t.Errorf("Field values for Labnum does not match. %v != %v.", br1.Labnum, br2.Labnum)
	}
	if br1.Timestamp != br2.Timestamp {
		t.Errorf("Field values for Timestamp does not match. %v != %v.", br1.Timestamp, br2.Timestamp)
	}
	if br1.PushTime != br2.PushTime {
		t.Errorf("Field values for PushTime does not match. %v != %v.", br1.PushTime, br2.PushTime)
	}
	if br1.TotalScore != br2.TotalScore {
		t.Errorf("Field values for TotalScore does not match. %v != %v.", br1.TotalScore, br2.TotalScore)
	}
	if br1.HeadCommitID != br2.HeadCommitID {
		t.Errorf("Field values for HeadCommitID does not match. %v != %v.", br1.HeadCommitID, br2.HeadCommitID)
	}
	if br1.HeadCommitText != br2.HeadCommitText {
		t.Errorf("Field values for HeadCommitText does not match. %v != %v.", br1.HeadCommitText, br2.HeadCommitText)
	}
	if br1.BuildTime != br2.BuildTime {
		t.Errorf("Field values for BuildTime does not match. %v != %v.", br1.BuildTime, br2.BuildTime)
	}
}
