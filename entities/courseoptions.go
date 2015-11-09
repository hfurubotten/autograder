package git

import (
	"encoding/gob"
	"math"
	"time"

	"github.com/autograde/kit/score"
)

func init() {
	gob.Register(CourseOptions{})
}

// LabAssignmentOptions represents a lab assignments results.
type LabAssignmentOptions struct {
	Notes         string      // Teachers notes on a lab.
	ExtraCredit   score.Score // extra credit from the teacher.
	ApproveDate   time.Time   // When a lab was approved.
	ApprovedBuild int         // Which build approved the lab.
	Builds        []int
}

// NewLabAssignmentOptions will create a new LabAssignmentOptions object.
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
	ApResults     map[int]*AntiPlagiarismResults
	UsedSlipDays  int

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
		ApResults:     make(map[int]*AntiPlagiarismResults),
	}
}

// RecalculateSlipDays will calculate and set the number of slipdays used on the
// specified course.
func (co *CourseOptions) RecalculateSlipDays() error {
	org, err := NewOrganization(co.Course, true)
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
		group, err := NewGroup(co.Course, co.GroupNum, true)
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

// AntiPlagiarismResults holds the results from the anti-plagiarism application.
type AntiPlagiarismResults struct {
	MossPct float32
	MossUrl string
	DuplPct float32
	DuplUrl string
	JplagPct float32
	JplagUrl string
}
