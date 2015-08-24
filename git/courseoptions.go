package git

import (
	"encoding/gob"
	"time"

	"github.com/autograde/kit/score"
)

func init() {
	gob.Register(CourseOptions{})
}

// LabAssignmentOptions represents a lab assignments teacher set results.
type LabAssignmentOptions struct {
	Notes       string      // Teachers notes on a lab.
	ExtraCredit score.Score // extra credit from the teacher.
	ApproveDate time.Time   // When a lab was approved.
}

// CourseOptions represent the course options a user need when signed up for a course.
type CourseOptions struct {
	Course        string
	CurrentLabNum int
	Assignments   map[int]LabAssignmentOptions

	// Group link
	IsGroupMember bool
	GroupNum      int
}

// NewCourseOptions will create a new course option object.
func NewCourseOptions(course string) CourseOptions {
	return CourseOptions{
		Course:        course,
		CurrentLabNum: 1,
		Assignments:   make(map[int]LabAssignmentOptions),
	}
}
