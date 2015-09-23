package points

import (
	"time"
)

const (
	COMMENT      int = 50
	OPEN_ISSUE   int = 70 // is a open and comment in the same action.
	CLOSE_ISSUE  int = 20
	REOPEN_ISSUE int = 20
	WIKI_UPDATE  int = 20
	ASSIGNMENT   int = 10
	UNASSIGNMENT int = 10
	LABEL        int = 10
	UNLABEL      int = 10
	ADD_LINE     int = 1
	REMOVE_LINE  int = 1
)

type SingleScorer interface {
	GetUsername() string
	DecScoreBy(score int)
	IncScoreBy(score int)
}

type MultiScorer interface {
	DecScoreBy(user string, score int)
	IncScoreBy(user string, score int)
	GetUserScore(user string) int64
	GetWeeklyUserScore(week int, user string) int64
	GetMonthlyUserScore(month time.Month, user string) int64
}
