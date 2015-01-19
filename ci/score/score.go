package score

// This struct is used to decode a score from a test or tests. When a test
// is passed or a calculation of partial passed test is found, output a JSON
// object representing this struct.
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
// student_score   = sum( task_score[i] * (Weight[i]/x) ), gives {0 < student_score < 1}
type Score struct {
	Secret   string // the unique identifier for your course
	TestName string // Name of the tests that is covered
	Score    int    // The score the student has accomplished
	MaxScore int    // Max score possible to get on this specific test(s)
	Weight   int    // The weight of this test(s)
}
