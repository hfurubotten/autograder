package entities

import (
	"log"
	"os"
	"testing"

	"github.com/hfurubotten/autograder/database"
)

func TestMain(m *testing.M) {
	err := database.Start("/tmp/ag")
	if err != nil {
		log.Println("Unable to start database:", err)
		return
	}
	m.Run()
	//TODO Add panic handler to ensure deletion of test output
	err = database.Close()
	if err != nil {
		log.Println("Unable to close the database: ", err)
	}
	err = os.RemoveAll("/tmp/ag")
	if err != nil {
		log.Println("Unable to remove database from file system: ", err)
	}
}
