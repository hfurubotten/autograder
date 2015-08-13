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

func TestNewBuildResult(t *testing.T) {
	startpoint := GetNextBuildID()
	for i := 1 + startpoint; i <= testGetNextBuildIDIterations; i++ {
		build, err := NewBuildResult()
		if err != nil {
			t.Error("Error creating a new build result object:", err)
		}

		if build.ID != i {
			t.Errorf("Error while generating build ID. Got %d, want %d.", build.ID, i)
		}

		if build.TestScores == nil {
			t.Error("Field TestScores cannot be nil")
		}

		if build.Log == nil {
			t.Error("Field Log cannot be nil")
		}
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
			t.Error("Failed to save the build result:", err)
			continue
		}

		br2, err := GetBuildResult(br.ID)
		if err != nil {
			t.Error("Failed to get a build result from DB:", err)
		}
		compareBuildResults(br, br2, t)
	}
}

var testGetNextBuildIDIterations = 100

func TestGetNextBuildID(t *testing.T) {
	startpoint := GetNextBuildID()
	for i := 1 + startpoint; i <= testGetNextBuildIDIterations; i++ {
		nextid := GetNextBuildID()
		if nextid != i {
			t.Errorf("Error while couting up to next build ID. Got %d, want %d.", nextid, i)
		}
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
	if br1.CommitID != br2.CommitID {
		t.Errorf("Field values for CommitID does not match. %v != %v.", br1.CommitID, br2.CommitID)
	}
	if br1.CommitText != br2.CommitText {
		t.Errorf("Field values for CommitText does not match. %v != %v.", br1.CommitText, br2.CommitText)
	}
}
