package logging

import (
	"log"

	"github.com/hfurubotten/autograder/maintenance"
)

// LoggingFolder is the folder where the log files should be stored.
var LoggingFolder = "/logs/"
var logHandler *LoggerHandler

// Start will start up nessasery processes for logging porpuses.
func Start() error {
	// log print appearance
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// registers logging maintenance handler
	logHandler = NewLoggerHandler()
	logHandler.Execute()

	maintenance.Register(logHandler)

	return nil
}

// Stop will close current file used to store logging.
func Stop() {
	if logHandler.LogFile != nil {
		logHandler.LogFile.Close()
	}
}
