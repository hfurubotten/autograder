package ci

import (
	"time"
)

type Result struct {
	Course    string
	User      string
	Log       []string
	NumPasses int
	Status    string
	Labnum    int
	Timestamp time.Time
}
