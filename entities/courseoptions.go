package git

import (
	"encoding/gob"
	"time"

	"github.com/hfurubotten/ag-scoring/score"
)

func init() {
	gob.Register(CourseOptions{})
}

// LabAssignmentOptions represents a lab assignments teacher set results.
type LabAssignmentOptions struct {
	Notes         string      // Teachers notes on a lab.
	ExtraCredit   score.Score // extra credit from the teacher.
	ApproveDate   time.Time   // When a lab was approved.
	ApprovedBuild int         // Which build approved the lab.
	Builds        []int
}

func NewLabAssignmentOptions() *LabAssignmentOptions {
	return &LabAssignmentOptions{
		ApprovedBuild: -1,
		Builds:        []int{},
	}
}

// AddBuildResult will add build ID to the assignment options.
func (l *LabAssignmentOptions) AddBuildResult(buildid int) {
	if l.Builds == nil {
		l.Builds = []int{}
	}

	l.Builds = append(l.Builds, buildid)
}

// CourseOptions represent the course options a user need when signed up for a course.
type CourseOptions struct {
	Course        string
	CurrentLabNum int
	Assignments   map[int]*LabAssignmentOptions

	// Group link
	IsGroupMember bool
	GroupNum      int
}

// NewCourseOptions will create a new course option object.
func NewCourseOptions(course string) CourseOptions {
	return CourseOptions{
		Course:        course,
		CurrentLabNum: 1,
		Assignments:   make(map[int]*LabAssignmentOptions),
	}
}
