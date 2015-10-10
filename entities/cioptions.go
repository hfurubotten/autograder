package entities

import (
	"encoding/gob"
)

func init() {
	gob.Register(CIOptions{})
}

// CIOptions represents the possible CI options for a course.
type CIOptions struct {
	Basepath string
	Secret   string
}
