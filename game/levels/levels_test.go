package levels

import (
	"math/rand"
	"testing"
)

var findLevelTest = []struct {
	in  int64
	out int
}{
	{0, 1},
	{72, 1},
}

func TestFindLevel(t *testing.T) {
	// dynamically adds some test cases.
	tmp := findLevelTest[0]
	for i, a := range LEVELS {
		tmp.in = a
		tmp.out = i + 1
		findLevelTest = append(findLevelTest, tmp)

		var diff int64
		if i+1 == len(LEVELS) {
			diff = 1000
		} else {
			diff = LEVELS[i+1] - LEVELS[i]
		}

		tmp.in = a + rand.Int63n(diff)
		tmp.out = i + 1
		findLevelTest = append(findLevelTest, tmp)
	}

	// does tests
	for _, flt := range findLevelTest {
		if FindLevel(flt.in) != flt.out {
			t.Errorf("Wrong level found for score %d, got %d, but want %d.", flt.in, FindLevel(flt.in), flt.out)
		}
	}
}
