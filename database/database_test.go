package database

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

var tmploc = "test.db"
var tmpbucket = "test"

// TestStart test the start function in the database package.
func TestStart(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function. " + err.Error())
	}

	if encoder == nil {
		t.Error("Encoder not initalized.")
	}

	if decoder == nil {
		t.Error("Decoder not initalized.")
	}

	cleanUpDB()
}

// testStoreValues is the test values for the TestStore and TestGet functions.
var testStoreValues = []struct {
	key, value string
}{
	{"hei", "sann"},
	{"key", "value"},
	{"sponge", "bob"},
	{"square", "pants"},
}

// TestStore will test the store function in the database package.
func TestStore(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function." + err.Error())
	}

	for _, v := range testStoreValues {
		Store(tmpbucket, []byte(v.key), v.value)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tmpbucket))
		if b == nil {
			t.Error("Couldn't open bucket: " + err.Error())
			t.FailNow()
		}

		for _, v := range testStoreValues {
			got := b.Get([]byte(v.key))

			var buf bytes.Buffer
			buf.Write(got)
			enc := gob.NewDecoder(&buf)

			var val string
			err := enc.Decode(&val)
			if err != nil {
				t.Error(err.Error())
			}

			if val != v.value {
				t.Errorf("Got %s want %s", val, v.value)
			}
		}
		return nil
	})

	cleanUpBucket()
	cleanUpDB()

	return
}

// TestGet will test the get function in the database package.
func TestGet(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function." + err.Error())
	}

	for _, v := range testStoreValues {
		Store(tmpbucket, []byte(v.key), v.value)

		var got string
		Get(tmpbucket, []byte(v.key), &got)

		if got != v.value {
			t.Errorf("Got %s want %s", got, v.key)
		}
	}

	cleanUpBucket()
	cleanUpDB()

	return
}

// TestClose will test the closing function of the database.
func TestClose(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function." + err.Error())
	}

	err = Close()
	if err != nil {
		t.Error("Error closing the database. " + err.Error())
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		return nil
	}); err == nil {
		t.Error("Database still open after close is called.")
	}

	cleanUpDB()
}

// cleanUpBucket will remove the test bucket from the database.
func cleanUpBucket() {
	if err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(tmpbucket))
	}); err != nil {
		panic("Couldn't clean up test bucket." + err.Error())
	}
}

// cleanUpDB will close and remove the database.
func cleanUpDB() {
	db.Close()
	if err := os.Remove(tmploc); err != nil {
		panic("Couldn't remove database." + err.Error())
	}
}
