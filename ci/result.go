package ci

import (
	"time"

	"github.com/hfurubotten/ag-scoring/score"
)

// Result represent a result from a test build.
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
	TestScores      []score.Score
	TotalScore      int
}
