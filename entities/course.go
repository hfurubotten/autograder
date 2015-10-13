package entities

import (
	"encoding/gob"
	"math"
)

func init() {
	gob.Register(Course{})
}

// CourseOptions represent the course options a user need when signed up for a course.
type Course struct {
	CourseName    string
	CurrentLabNum int
	Assignments   map[int]*Assignment
	UsedSlipDays  int

	// Group link
	IsGroupMember bool
	GroupNum      int
}

// NewCourseOptions will create a new course option object.
func NewCourse(course string) Course {
	return Course{
		CourseName:    course,
		CurrentLabNum: 1,
		Assignments:   make(map[int]*Assignment),
	}
}

// RecalculateSlipDays will calculate and set the number of slipdays used on the
// specified course.
func (co *Course) RecalculateSlipDays() error {
	org, err := NewOrganization(co.CourseName, true)
	if err != nil {
		return err
	}

	days := 0

	for i, lab := range co.Assignments {
		if _, ok := org.IndividualDeadlines[i]; !ok {
			continue
		}

		if lab.ApproveDate.After(org.IndividualDeadlines[i]) {
			days += int(math.Floor((lab.ApproveDate.Sub(org.IndividualDeadlines[i]).Hours() - 3) / 24))
		}
	}

	if co.IsGroupMember {
		group, err := NewGroup(co.CourseName, co.GroupNum, true)
		if err != nil {
			return err
		}

		for i, lab := range group.Assignments {
			if _, ok := org.GroupDeadlines[i]; !ok {
				continue
			}

			if lab.ApproveDate.After(org.GroupDeadlines[i]) {
				days += int(math.Floor((lab.ApproveDate.Sub(org.GroupDeadlines[i]).Hours() - 3) / 24))
			}
		}
	}

	co.UsedSlipDays = days

	return nil
}
