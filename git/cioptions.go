package git

import (
	"encoding/gob"
)

func init() {
	gob.Register(CIOptions{})
}

type CIOptions struct {
	Basepath string
}
