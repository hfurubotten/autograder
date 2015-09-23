package levels

var LEVELS []int64 = []int64{
	0,
	100,
	500,
	1000,
	2500,
	6000,
	10000,
	14000,
	20000,
}

type Leveler interface {
	IncLevel()
	DecLevel()
	Level() int
}

func FindLevel(score int64) int {
	for i, s := range LEVELS {
		if s > score {
			return i
		}
	}

	return len(LEVELS)
}

func RequiredForLevel(lvl int) int64 {
	if lvl >= len(LEVELS) {
		return LEVELS[len(LEVELS)-1]
	} else if lvl < 0 {
		return 0
	}

	return LEVELS[lvl]
}
