package entities

import (
	"encoding/gob"
	"time"

	"github.com/autograde/kit/score"
)

func init() {
	gob.Register(Assignment{})
}

// Assignment holds data related to results of a lab assignment.
type Assignment struct {
	Notes         string      // teachers notes on assignment
	ExtraCredit   score.Score // extra credit awarded by teacher
	ApproveDate   time.Time   // when a lab was approved
	ApprovedBuild int         // which build approved the lab
	Builds        []int
}

// NewAssignment creates a new Assignment object.
func NewAssignment() *Assignment {
	return &Assignment{
		ApprovedBuild: -1,
		Builds:        []int{},
	}
}

// AddBuildResult will add build ID to the assignment options.
func (l *Assignment) AddBuildResult(buildid int) {
	if l.Builds == nil {
		l.Builds = []int{}
	}
	l.Builds = append(l.Builds, buildid)
}
