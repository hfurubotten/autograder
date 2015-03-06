package git

import (
	"encoding/gob"
	"time"

	"github.com/hfurubotten/ag-scoring/score"
)

func init() {
	gob.Register(CourseOptions{})
}

// CourseOptions represent the course options a user need when signed up for a course.
type CourseOptions struct {
	Course        string
	CurrentLabNum int
	Notes         map[int]string      // Teachers notes on a lab.
	ExtraCredit   map[int]score.Score // extra credit from the teacher.
	ApproveDate   map[int]time.Time   // When a lab was approved.

	// Group link
	IsGroupMember bool
	GroupNum      int
}

// NewCourseOptions will create a new course option object.
func NewCourseOptions(course string) CourseOptions {
	return CourseOptions{
		Course:        course,
		CurrentLabNum: 1,
		Notes:         make(map[int]string),
		ExtraCredit:   make(map[int]score.Score),
		ApproveDate:   make(map[int]time.Time),
	}
}
