package logging

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hfurubotten/autograder/global"
)

// LoggerHandler is a MaintenaceHandler which will set a new log file for every
// new day.
type LoggerHandler struct {
	LogFile *os.File
}

// NewLoggerHandler will create a new handler for updating log files.
func NewLoggerHandler() *LoggerHandler {
	return &LoggerHandler{}
}

// Execute will actually update the log file location to the log package, and
// close the old file.
func (lh *LoggerHandler) Execute() (err error) {
	old := lh.LogFile

	filename := filepath.Join(global.Basepath, LoggingFolder, time.Now().Format("2006-01-02")+".log")
	lh.LogFile, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	log.SetOutput(lh.LogFile)

	if old != nil {
		return old.Close()
	}
	return nil
}

// ExecutingTime should give at which time the handler should be fired, set to
// static zero for now.
func (lh LoggerHandler) ExecutingTime() int {
	return 0
}

// RemoveAfterExecute tells if the handler should be removed after execution.
// Value: Static false.
func (lh LoggerHandler) RemoveAfterExecute() bool {
	return false
}
