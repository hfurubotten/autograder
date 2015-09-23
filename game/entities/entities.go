package entities

import (
	"sync"
)

type Saver interface {
	sync.Locker
	Save() (err error) // Save method must call the unlock method
}
