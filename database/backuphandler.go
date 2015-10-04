package database

import (
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hfurubotten/autograder/global"
)

// DBBackupFolder is the folder used to store DB backups.
var DBBackupFolder = "/db_bak/"

// BackupHandler is a MaintenaceHandler which will set a new log file for every
// new day.
type BackupHandler struct{}

// NewBackupHandler will create a new handler for updating log files.
func NewBackupHandler() *BackupHandler {
	return &BackupHandler{}
}

// Execute will actually update the log file location to the log package, and
// close the old file.
func (bh *BackupHandler) Execute() error {
	return db.View(func(tx *bolt.Tx) error {
		filename := filepath.Join(global.Basepath, DBBackupFolder, time.Now().Format("2006-01-02")+".log")

		return tx.CopyFile(filename, 0666)
	})
}

// ExecutingTime should give at which time the handler should be fired, set to
// static zero for now.
func (bh BackupHandler) ExecutingTime() int {
	return 0
}

// RemoveAfterExecute tells if the handler should be removed after execution.
// Value: Static false.
func (bh BackupHandler) RemoveAfterExecute() bool {
	return false
}
