package ci

import (
	"time"

	. "github.com/hfurubotten/ag-scoring/score"
)

type Result struct {
	Course          string
	User            string
	Log             []string
	NumPasses       int
	NumFails        int
	NumBuildFailure int
	Status          string
	Labnum          int
	Timestamp       time.Time
	PushTime        time.Time
	TestScores      []Score
	TotalScore      int
}
