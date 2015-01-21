package ci

import (
	"time"

	. "github.com/hfurubotten/autograder/ci/score"
)

type Result struct {
	Course     string
	User       string
	Log        []string
	NumPasses  int
	NumFails   int
	Status     string
	Labnum     int
	Timestamp  time.Time
	PushTime   time.Time
	TestScores []Score
	TotalScore int
}
