package score

import (
	"encoding/json"
	"testing"
)

// Unique secret for every course.
const (
	secretDAT520 = "ab553ad7c8a21"
	secretDATFOO = "abcdefghijklm"
)

// Score is a struct used to encode/decode a score from a test or tests. When a
// test is passed or a calculation of partial passed test is found, output a
// JSON object representing this struct.
//
// Secret read from the output steam need to correspond to the course identifier
// given on the teachers panel. All other output will be ignored.
//
// With the formula in the Autograder CI the score percentage is calculated
// automatically. Give any max score, then pass on a given score the student
// gets for passed sub test within this the max score. Finally, set a weight
// it should have on the total. The weight does not need to within 100 or a
// percentage. If you want to only give a score for completing a test, then
// MaxScore == Score.
//
// Calculations in the CI follows this formula:
// total_weight    = sum(Weight)
// task_score[0:n] = Score[i] / MaxScore[i], gives {0 < task_score < 1}
// student_score   = sum( task_score[i] * (Weight[i]/total_weight) ), gives {0 < student_score < 1}
type Score struct {
	Secret   string // the unique identifier for your course
	TestName string // Name of the tests that is covered
	Score    int    // The score the student has accomplished
	MaxScore int    // Max score possible to get on this specific test(s)
	Weight   int    // The weight of this test(s)
}

// DumpAsJSON encodes s as JSON and prints the result to testing context t.
func (s *Score) DumpAsJSON(t *testing.T) {
	b, err := json.Marshal(s)
	if err != nil {
		t.Logf("error dumping score to json: %v\n", err)
	}
	t.Logf("%s\n", b)
}

// DumpScoreToStudent prints score s to testing context t as a string using the
// format: "TestName: 2/10 cases passed".
func (s *Score) DumpScoreToStudent(t *testing.T) {
	t.Logf("%s: %d/%d cases passed", s.TestName, s.Score, s.MaxScore)
}

// NewScoreDAT520 returns a new Score with the given arguments and the secret
// for course DAT520 set.
func NewScoreDAT520(testName string, max, weight int) *Score {
	return &Score{
		Secret:   secretDAT520,
		TestName: testName,
		MaxScore: max,
		Weight:   weight,
	}
}
