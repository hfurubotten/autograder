package web

import "testing"

func TestPanicHandler(t *testing.T) {
	defer PanicHandler(false)
	panic("Catch this panic.")
}
