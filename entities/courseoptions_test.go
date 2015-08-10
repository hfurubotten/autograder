package git

import (
	"testing"
)

var newCourseOptionsInput = []string{
	"course1",
	"course2",
	"course3",
}

func TestNewCourseOptions(t *testing.T) {
	for _, inputname := range newCourseOptionsInput {
		opt := NewCourseOptions(inputname)

		if opt.Assignments == nil {
			t.Error("NewCourseOptions created struct with Assignments as nil, want map[int]LabAssignmentOptions.")
		}

		if opt.Course != inputname {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %s, want %s.", opt.Course, inputname)
		}

		if opt.CurrentLabNum != 1 {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %d, want 1.", opt.CurrentLabNum)
		}

		if opt.GroupNum != 0 {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %d, want 0.", opt.GroupNum)
		}

		if opt.IsGroupMember {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %b, want false.", opt.IsGroupMember)
		}
	}
}
