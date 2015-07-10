package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

var encoder *gob.Encoder
var encbuffer bytes.Buffer
var encbufferlock sync.Mutex

var decoder *gob.Decoder
var decbuffer bytes.Buffer
var decbufferlock sync.Mutex

// Start will start up the database. If the database does not already exist, a new one will be created.
func Start(dbloc string) (err error) {
	db, err = bolt.Open(dbloc, 0666, nil)
	if err != nil {
		return err
	}

	encoder = gob.NewEncoder(&encbuffer)
	decoder = gob.NewDecoder(&decbuffer)

	return
}

// Store will put a new value in the assigned bucket(e.g. table) with given key.
//
// Key variable can be integer or string type.
// Value variable can be any type.
func Store(bucket string, key []byte, value interface{}) (err error) {
	err = db.Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(bucket))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			// Create a bucket.
			b, err = tx.CreateBucket([]byte(bucket))
			if err != nil {
				return err
			}
		}

		encbufferlock.Lock()
		defer encbufferlock.Unlock()

		if err = encoder.Encode(value); err != nil {
			return
		}

		data, err := ioutil.ReadAll(&encbuffer)
		if err != nil {
			return err
		}

		return b.Put(key, data)
	})

	return err
}

// Get will get a value for the given key in a bucket(e.g. table).
func Get(bucket string, key []byte, val interface{}) (err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Trying to access a nonexisting bucket. ")
		}

		decbufferlock.Lock()
		defer decbufferlock.Unlock()

		data := b.Get(key)
		if data == nil {
			return errors.New("No data in database.")
		}

		n, _ := decbuffer.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		return decoder.Decode(val)
	})
	return err
}

// Close will shut down the database in a safe mather.
func Close() (err error) {
	err = db.Close()

	encoder = nil
	decoder = nil

	return
}
