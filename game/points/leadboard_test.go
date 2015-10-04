package points

import (
	"sync"
	"testing"
	"time"
)

var checkIntergrityTest = []Leaderboard{
	{},
	{
		TotalScore: make(map[string]int64),
	},
	{
		TotalScore:  make(map[string]int64),
		WeeklyScore: make(map[int]map[string]int64),
	},
	{
		TotalScore:   make(map[string]int64),
		WeeklyScore:  make(map[int]map[string]int64),
		MonthlyScore: make(map[time.Month]map[string]int64),
	},
}

func TestCheckIntegrity(t *testing.T) {
	for _, in := range checkIntergrityTest {
		in.checkIntegrity()
		if in.TotalScore == nil {
			t.Error("Wanted map structure in TotalScore field, got nil.")
		}
		if in.WeeklyScore == nil {
			t.Error("Wanted map structure in WeeklyScore field, got nil.")
		}
		if in.MonthlyScore == nil {
			t.Error("Wanted map structure in MonthlyScore field, got nil.")
		}
	}
}

var incScoreByTest = []struct {
	inUser    string
	inScore   int
	wantScore int64
}{
	{"user1", 30, 30},
	{"user2", 100, 100},
	{"user3", 25, 25},
	{"user1", 356, 386},
	{"user3", 231, 256},
	{"user2", 515, 615},
	{"user2", 850, 1465},
	{"user3", 848, 1104},
	{"user1", 552, 938},
}

var decScoreByTest = []struct {
	inUser    string
	inScore   int
	wantScore int64
}{
	{"user1", 40, 898},
	{"user2", 110, 1355},
	{"user3", 104, 1000},
	{"user2", 250, 1105},
	{"user1", 58, 840},
	{"user3", 250, 750},
}

func TestIncDecScoreBy(t *testing.T) {
	lb := Leaderboard{scorelock: &sync.RWMutex{}}

	_, week := time.Now().ISOWeek()
	month := time.Now().Month()

	for _, isb := range incScoreByTest {
		lb.IncScoreBy(isb.inUser, isb.inScore)

		if lb.TotalScore[isb.inUser] != isb.wantScore {
			t.Errorf("Wrong total score added to %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.TotalScore[isb.inUser], isb.wantScore)
		}

		if lb.MonthlyScore[month][isb.inUser] != isb.wantScore {
			t.Errorf("Wrong month score added to %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.MonthlyScore[month][isb.inUser], isb.wantScore)
		}

		if lb.WeeklyScore[week][isb.inUser] != isb.wantScore {
			t.Errorf("Wrong week score added to %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.WeeklyScore[week][isb.inUser], isb.wantScore)
		}
	}

	for _, isb := range decScoreByTest {
		lb.DecScoreBy(isb.inUser, isb.inScore)

		if lb.TotalScore[isb.inUser] != isb.wantScore {
			t.Errorf("Wrong total score subtracted from %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.TotalScore[isb.inUser], isb.wantScore)
		}

		if lb.MonthlyScore[month][isb.inUser] != isb.wantScore {
			t.Errorf("Wrong month score subtracted from to %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.MonthlyScore[month][isb.inUser], isb.wantScore)
		}

		if lb.WeeklyScore[week][isb.inUser] != isb.wantScore {
			t.Errorf("Wrong week score subtracted from to %s with %d, got %d, want %d.", isb.inUser, isb.inScore, lb.WeeklyScore[week][isb.inUser], isb.wantScore)
		}
	}
}

var getUnknownUserTest = []string{
	"user4",
	"user5",
}

func TestGetUserScores(t *testing.T) {
	lb := Leaderboard{scorelock: &sync.RWMutex{}}
	_, week := time.Now().ISOWeek()
	month := time.Now().Month()

	// check on empty leaderboard object.
	if lb.GetUserScore("unknown") != 0 {
		t.Errorf("Got %d on empty leaderboard object, want 0.", lb.GetUserScore("unknown"))
	}

	if lb.GetMonthlyUserScore(month, "unknown") != 0 {
		t.Errorf("Got %d on empty leaderboard object, want 0.", lb.GetUserScore("unknown"))
	}

	if lb.GetWeeklyUserScore(week, "unknown") != 0 {
		t.Errorf("Got %d on empty leaderboard object, want 0.", lb.GetUserScore("unknown"))
	}

	// checks with values for users.
	for _, isb := range incScoreByTest {
		lb.IncScoreBy(isb.inUser, isb.inScore)

		if lb.GetUserScore(isb.inUser) != isb.wantScore {
			t.Errorf("Wrong total score given back for %s, got %d, want %d.", isb.inUser, lb.GetUserScore(isb.inUser), isb.wantScore)
		}

		if lb.GetMonthlyUserScore(month, isb.inUser) != isb.wantScore {
			t.Errorf("Wrong total score given back for %s, got %d, want %d.", isb.inUser, lb.GetMonthlyUserScore(month, isb.inUser), isb.wantScore)
		}

		if lb.GetWeeklyUserScore(week, isb.inUser) != isb.wantScore {
			t.Errorf("Wrong total score given back for %s, got %d, want %d.", isb.inUser, lb.GetWeeklyUserScore(week, isb.inUser), isb.wantScore)
		}
	}

	// checks with unknown month and week.
	if lb.GetMonthlyUserScore(month-1, "unknown") != 0 {
		t.Errorf("Got %d on empty leaderboard object, want 0.", lb.GetUserScore("unknown"))
	}

	if lb.GetWeeklyUserScore(week-1, "unknown") != 0 {
		t.Errorf("Got %d on empty leaderboard object, want 0.", lb.GetUserScore("unknown"))
	}

	// checks with unknown user.
	for _, u := range getUnknownUserTest {

		if lb.GetUserScore(u) != 0 {
			t.Errorf("Wrong total score given back for %s, got %d, want 0.", u, lb.GetUserScore(u))
		}

		if lb.GetMonthlyUserScore(month, u) != 0 {
			t.Errorf("Wrong total score given back for %s, got %d, want 0.", u, lb.GetMonthlyUserScore(month, u))
		}

		if lb.GetWeeklyUserScore(week, u) != 0 {
			t.Errorf("Wrong total score given back for %s, got %d, want 0.", u, lb.GetWeeklyUserScore(week, u))
		}
	}
}

var leaderboardInTest = []struct {
	user  string
	score int
}{
	{"user1", 400},
	{"user2", 1000},
	{"user3", 40},
}

var leaderboardOutTest = []string{
	"user2",
	"user1",
	"user3",
}

func TestGetTotalLeadBoard(t *testing.T) {
	lb := Leaderboard{scorelock: &sync.RWMutex{}}

	for _, in := range leaderboardInTest {
		lb.IncScoreBy(in.user, in.score)
	}

	tl := lb.GetTotalLeaderboard()

	for i, s := range tl {
		if s != leaderboardOutTest[i] {
			t.Errorf("Got %s on place %d on total leadboard, want %s.", s, i, leaderboardOutTest[i])
		}
	}
}

func TestGetMonthlyLeadBoard(t *testing.T) {
	lb := Leaderboard{scorelock: &sync.RWMutex{}}
	month := time.Now().Month()

	for _, in := range leaderboardInTest {
		lb.IncScoreBy(in.user, in.score)
	}

	tl := lb.GetMonthlyLeaderboard(month)

	for i, s := range tl {
		if s != leaderboardOutTest[i] {
			t.Errorf("Got %s on place %d on total leadboard, want %s.", s, i, leaderboardOutTest[i])
		}
	}
}

func TestGetWeeklyLeadBoard(t *testing.T) {
	lb := Leaderboard{scorelock: &sync.RWMutex{}}
	_, week := time.Now().ISOWeek()

	for _, in := range leaderboardInTest {
		lb.IncScoreBy(in.user, in.score)
	}

	tl := lb.GetWeeklyLeaderboard(week)

	for i, s := range tl {
		if s != leaderboardOutTest[i] {
			t.Errorf("Got %s on place %d on total leadboard, want %s.", s, i, leaderboardOutTest[i])
		}
	}
}
