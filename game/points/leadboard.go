package points

import (
	"sync"
	"time"
)

// Leaderboard holds the score for the leaders.
type Leaderboard struct {
	scorelock *sync.RWMutex

	// Scores. Holds up to 1 year with information.
	TotalScore   map[string]int64
	WeeklyScore  map[int]map[string]int64
	MonthlyScore map[time.Month]map[string]int64

	// Leadboard
	TotalLeaderboard []string
	WeeklyLeadboard  []string
	MonthlyLeadboard []string

	// Trackers
	TrackingWeek  int
	TrackingMonth time.Month
}

// checkIntegrity is a internal method to check if any of the fields are nil and create the values if they are nil.
func (l *Leaderboard) checkIntegrity() {
	if l.TotalScore == nil {
		l.TotalScore = make(map[string]int64)
	}

	if l.WeeklyScore == nil {
		l.WeeklyScore = make(map[int]map[string]int64)
	}

	if l.MonthlyScore == nil {
		l.MonthlyScore = make(map[time.Month]map[string]int64)
	}
}

// IncScoreBy will increment the score a user has earned.
// This also updates the weekly and montly scores.
func (l *Leaderboard) IncScoreBy(user string, score int) {
	l.scorelock.Lock()
	defer l.scorelock.Unlock()

	l.checkIntegrity()

	l.TotalScore[user] += int64(score)

	_, week := time.Now().ISOWeek()
	month := time.Now().Month()

	if l.TrackingWeek != week {
		if l.TrackingWeek == 52 && week == 1 {
			delete(l.WeeklyScore, 53)
		}
		l.TrackingWeek = week
		delete(l.WeeklyScore, week)
	}

	if l.TrackingMonth != month {
		l.TrackingMonth = month
		delete(l.MonthlyScore, month)
	}

	// updates weekly
	if _, ok := l.WeeklyScore[week]; !ok {
		l.WeeklyScore[week] = make(map[string]int64)
	}

	l.WeeklyScore[week][user] += int64(score)
	// updated monthly
	if _, ok := l.MonthlyScore[month]; !ok {
		l.MonthlyScore[month] = make(map[string]int64)
	}

	l.MonthlyScore[month][user] += int64(score)

	// updated the leadboard
	l.updateCurrentLeaderboard(week, month)
}

// DecScoreBy will decrement the score a user has earned.
// This will also update the weekly and monthly scores.
func (l *Leaderboard) DecScoreBy(user string, score int) {
	// the negative of adding a score.
	l.IncScoreBy(user, -score)
}

// updateCurrentLeaderboard is a internal method to sort
// the leadboard when new scores has come in. This method
// is not locked, so make sure to have locking in place
// where this gets called.
func (l *Leaderboard) updateCurrentLeaderboard(week int, month time.Month) {
	// total score
	sorter := NewScoreSorter(l.TotalScore)
	l.TotalLeaderboard = sorter.Sorted()

	// monthly score
	sorter = NewScoreSorter(l.MonthlyScore[month])
	l.MonthlyLeadboard = sorter.Sorted()

	// weekly score
	sorter = NewScoreSorter(l.WeeklyScore[week])
	l.WeeklyLeadboard = sorter.Sorted()
}

// GetTotalLeaderboard gives the total score for a given user.
func (l *Leaderboard) GetTotalLeaderboard() []string {
	return l.TotalLeaderboard
}

// GetWeeklyLeaderboard gives the ranking list of users on the given weekly leadboard.
func (l *Leaderboard) GetWeeklyLeaderboard(week int) []string {
	sorter := NewScoreSorter(l.WeeklyScore[week])
	return sorter.Sorted()
}

// GetMonthlyLeaderboard gives the ranking list of users on the given monthly leadboard.
func (l *Leaderboard) GetMonthlyLeaderboard(month time.Month) []string {
	sorter := NewScoreSorter(l.MonthlyScore[month])
	return sorter.Sorted()
}

// GetUserScore gives the score for a user in the total score.
func (l *Leaderboard) GetUserScore(user string) int64 {
	if l.TotalScore == nil {
		return 0
	}

	if s, ok := l.TotalScore[user]; ok {
		return s
	}

	return 0
}

// GetWeeklyUserScore gives the score for a user a given week.
func (l *Leaderboard) GetWeeklyUserScore(week int, user string) int64 {
	if l.WeeklyScore == nil {
		return 0
	}

	if _, ok := l.WeeklyScore[week]; !ok {
		return 0
	}

	if s, ok := l.WeeklyScore[week][user]; ok {
		return s
	}

	return 0
}

// GetMonthlyUserScore gives the score for a user a given month.
func (l *Leaderboard) GetMonthlyUserScore(month time.Month, user string) int64 {
	if l.MonthlyScore == nil {
		return 0
	}

	if _, ok := l.MonthlyScore[month]; !ok {
		return 0
	}

	if s, ok := l.MonthlyScore[month][user]; ok {
		return s
	}

	return 0
}
