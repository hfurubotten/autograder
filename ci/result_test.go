package ci

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/autograde/kit/score"
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
	opt := DaemonOptions{}
	first, err := NewBuildResult(opt)
	if err != nil {
		t.Error(err)
	}
	for i := first.ID + 1; i <= iter; i++ {
		build, err := NewBuildResult(opt)
		if err != nil {
			t.Error("Error creating a new build result object:", err)
		}
		if build.ID != i {
			t.Errorf("Error generating build ID. Got %d, wanted %d.", build.ID, i)
		}
		if build.TestScores == nil {
			t.Error("Field TestScores cannot be nil")
		}
		if build.log == nil {
			t.Error("Field log cannot be nil")
		}
	}
}

func TestConcurrentNewBuildResult(t *testing.T) {
	opt := DaemonOptions{}
	first, err := NewBuildResult(opt)
	if err != nil {
		t.Error(err)
	}
	for i := first.ID + 1; i <= iter; i++ {
		go func() {
			build, err := NewBuildResult(opt)
			if err != nil {
				t.Error("Error creating a new build result object:", err)
			}
			if build.ID != i {
				t.Errorf("Error generating build ID. Got %d, wanted %d.", build.ID, i)
			}
			if build.TestScores == nil {
				t.Error("Field TestScores cannot be nil")
			}
			if build.log == nil {
				t.Error("Field log cannot be nil")
			}
		}()
	}
}

var buildResults = []*BuildResult{
	&BuildResult{
		ID:     8,
		Course: "coursex",
		User:   "user1",
	},
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
	for _, br := range buildResults {
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
	if br1.numBuildFailure != br2.numBuildFailure {
		t.Errorf("Field values for numBuildFailure does not match. %v != %v.", br1.numBuildFailure, br2.numBuildFailure)
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

var logOutput = `
{"Secret":"my secret code","TestName":"TestErrorsAG","Score":16,"MaxScore":16,"Weight":20}
TestErrorsAG: 16/16 cases passed

{"Secret":"my secret code","TestName":"TestFibonacciAG","Score":14,"MaxScore":14,"Weight":20}
TestFibonacciAG: 14/14 cases passed

{"Secret":"my secret code","TestName":"TestFormatAG","Score":3,"MaxScore":3,"Weight":20}
TestFormatAG: 3/3 cases passed

{"Secret":"my secret code","TestName":"TestGoCommandsAG","Score":0,"MaxScore":5,"Weight":20}
TestGoCommandsAG: 0/5 cases passed
--- FAIL: TestGoCommandsAG (0.00s)
	multiple_choice.go:50: TestGoCommandsAG 1: '\n' is incorrect.
	multiple_choice.go:50: TestGoCommandsAG 2: '\n' is incorrect.
	multiple_choice.go:50: TestGoCommandsAG 3: '\n' is incorrect.
	multiple_choice.go:50: TestGoCommandsAG 4: '\n' is incorrect.
	multiple_choice.go:50: TestGoCommandsAG 5: '\n' is incorrect.

{"Secret":"my secret code","TestName":"TestWritersAG","Score":12,"MaxScore":12,"Weight":20}
TestWritersAG: 12/12 cases passed

{"Secret":"my secret code","TestName":"TestRot13AG","Score":6,"MaxScore":6,"Weight":20}
TestRot13AG: 6/6 cases passed

{"Secret":"my secret code","TestName":"TestStringerAG","Score":3,"MaxScore":3,"Weight":20}
TestStringerAG: 3/3 cases passed
FAIL
exit status 1
FAIL	github.com/uis-dat320/labassignments-2015/lab2-go	0.108s
`

func TestBuildResultLog(t *testing.T) {
	opt := DaemonOptions{Secret: "my secret code", AdminToken: "my token"}
	br := buildResults[0]
	if err := br.Save(); err != nil {
		t.Error("Failed to save build result: ", err)
	}
	r, err := GetBuildResult(br.ID)
	if err != nil {
		t.Error("Failed to get a build result from DB: ", err)
	}
	lines := strings.Split(logOutput, "\n")
	for _, l := range lines {
		r.Add(l, opt)
	}
	tot := score.Total(r.TestScores)
	if tot != 85 { // 6 exercises with all pass of 7 = 6/7 = 85.71
		t.Errorf("Got: %d, Want: %d", tot, 86)
	}
	for _, l := range r.log {
		if strings.Contains(l, opt.Secret) || strings.Contains(l, opt.AdminToken) {
			t.Errorf("Security issue (log contains secret keyword):\n%v", l)
		}
	}
}
