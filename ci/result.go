package ci

import (
	"time"
)

type Result struct {
	Course    string
	User      string
	Log       []string
	NumPasses int
	NumFails  int
	Status    string
	Labnum    int
	Timestamp time.Time
}
