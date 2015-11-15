package web

import (
	"log"
	"runtime/debug"
)

// PanicHandler is used to catch a panicing goroutine from crashing the process.
// If the provided printStack is true, a stack trace is printed to the log.
// The function must be called from a defer function.
func PanicHandler(printStack bool) {
	if r := recover(); r != nil {
		log.Println("Recovered from panic:", r)
		if printStack {
			log.Println(string(debug.Stack()))
		}
	}
}
