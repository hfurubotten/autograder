package entities

import (
	"testing"
	"time"
)

var newCourseOptionsInput = []string{
	"course1",
	"course2",
	"course3",
}

func TestNewCourseOptions(t *testing.T) {
	for _, inputname := range newCourseOptionsInput {
		opt := NewCourse(inputname)

		if opt.Assignments == nil {
			t.Error("NewCourseOptions created struct with Assignments as nil, want map[int]LabAssignmentOptions.")
		}

		if opt.CourseName != inputname {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %s, want %s.", opt.CourseName, inputname)
		}

		if opt.CurrentLabNum != 1 {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %d, want 1.", opt.CurrentLabNum)
		}

		if opt.GroupNum != 0 {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %d, want 0.", opt.GroupNum)
		}

		if opt.IsGroupMember {
			t.Errorf("NewCourseOptions created struct with IsGroupMember as %t, want false.", opt.IsGroupMember)
		}
	}
}

var zone = time.FixedZone("Europe/Berlin", 0)

var testRecalculateSlipDays = []struct {
	IndvApproveDate  []time.Time
	GroupApproveDate []time.Time
	IndvDeadline     []time.Time
	GroupDeadline    []time.Time
	ExpectedSlipdays int
	Course           string
	GroupID          int
	GroupName        string
}{
	{
		IndvApproveDate: []time.Time{
			time.Date(2015, time.January, 12, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 22, 12, 12, 12, 0, zone),
		},
		IndvDeadline: []time.Time{
			time.Date(2015, time.January, 10, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 20, 12, 12, 12, 0, zone),
		},
		ExpectedSlipdays: 2,
		Course:           "slipdaycourse1",
	},
	{
		IndvApproveDate: []time.Time{
			time.Date(2015, time.January, 2, 22, 12, 12, 0, zone),
			time.Date(2015, time.January, 4, 12, 12, 12, 0, zone),
		},
		IndvDeadline: []time.Time{
			time.Date(2015, time.January, 1, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 2, 12, 12, 12, 0, zone),
		},
		ExpectedSlipdays: 2,
		Course:           "slipdaycourse2",
	},
	{
		IndvApproveDate: []time.Time{
			time.Date(2015, time.January, 12, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 22, 12, 12, 12, 0, zone),
		},
		IndvDeadline: []time.Time{
			time.Date(2015, time.January, 10, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 20, 12, 12, 12, 0, zone),
		},
		GroupApproveDate: []time.Time{
			time.Date(2015, time.February, 12, 12, 12, 12, 0, zone),
			time.Date(2015, time.February, 22, 12, 12, 12, 0, zone),
		},
		GroupDeadline: []time.Time{
			time.Date(2015, time.February, 10, 12, 12, 12, 0, zone),
			time.Date(2015, time.February, 20, 12, 12, 12, 0, zone),
		},
		ExpectedSlipdays: 4,
		Course:           "slipdaycourse3",
		GroupID:          123,
		GroupName:        "123",
	},
	{
		IndvApproveDate: []time.Time{
			time.Date(2015, time.January, 12, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 22, 12, 12, 12, 0, zone),
			time.Date(2015, time.March, 29, 13, 12, 12, 0, zone),
		},
		IndvDeadline: []time.Time{
			time.Date(2015, time.January, 10, 12, 12, 12, 0, zone),
			time.Date(2015, time.January, 20, 12, 12, 12, 0, zone),
			time.Date(2015, time.March, 20, 12, 12, 12, 0, zone),
		},
		GroupApproveDate: []time.Time{
			time.Date(2015, time.February, 12, 12, 12, 12, 0, zone),
			time.Date(2015, time.February, 22, 12, 12, 12, 0, zone),
			time.Date(2015, time.April, 5, 22, 12, 12, 0, zone),
			time.Date(2015, time.April, 22, 12, 12, 12, 0, zone),
		},
		GroupDeadline: []time.Time{
			time.Date(2015, time.February, 10, 12, 12, 12, 0, zone),
			time.Date(2015, time.February, 20, 12, 12, 12, 0, zone),
			time.Date(2015, time.April, 1, 12, 12, 12, 0, zone),
			time.Date(2015, time.April, 16, 12, 12, 12, 0, zone),
		},
		ExpectedSlipdays: 21,
		Course:           "slipdaycourse3",
		GroupID:          123,
		GroupName:        "123",
	},
}

func TestRecalculateSlipDays(t *testing.T) {
	for _, in := range testRecalculateSlipDays {
		org, err := NewOrganization(in.Course, true)
		if err != nil {
			t.Error(err)
			continue
		}

		opt := NewCourse(in.Course)

		for i, date := range in.IndvApproveDate {
			org.IndividualDeadlines[i] = in.IndvDeadline[i]
			opt.Assignments[i] = NewAssignment()
			opt.Assignments[i].ApproveDate = date
		}

		if in.GroupID > 0 {
			group := NewGroupX(in.Course, in.GroupName)

			opt.IsGroupMember = true
			opt.GroupNum = in.GroupID
			opt.GroupName = in.GroupName

			for i, date := range in.GroupApproveDate {
				org.GroupDeadlines[i] = in.GroupDeadline[i]
				group.Assignments[i] = NewAssignment()
				group.Assignments[i].ApproveDate = date
			}

			err := group.Save()
			if err != nil {
				t.Error(err)
			}
		}

		if err := opt.RecalculateSlipDays(); err != nil {
			t.Error(err)
			continue
		}

		if opt.UsedSlipDays != in.ExpectedSlipdays {
			t.Errorf("Expected used slipdays not correct. want %d, got %d, for course %v",
				in.ExpectedSlipdays,
				opt.UsedSlipDays,
				in.Course)
		}
	}
}
