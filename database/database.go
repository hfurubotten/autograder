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
var registeredBucketNames = make([]string, 0)

// Start will start up the database. If the database does not already exist, a new one will be created.
func Start(dbloc string) (err error) {
	db, err = bolt.Open(dbloc, 0666, nil)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) (err error) {
		// Create a buckets.
		for _, bucket := range registeredBucketNames {
			if _, err = tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	})
}

// Store will put a new value in the assigned bucket(e.g. table) with given key.
//
// Key variable can be integer or string type.
// Value variable can be any type.
func Store(bucket string, key string, value interface{}) (err error) {
	return db.Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(bucket))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			// Create a bucket.
			b, err = tx.CreateBucket([]byte(bucket))
			if err != nil {
				return err
			}

			if b == nil {
				return errors.New("Couldn't create bucket.")
			}
		}

		defer Unlock(bucket, key)

		buf := &bytes.Buffer{}
		encoder := gob.NewEncoder(buf)

		if err = encoder.Encode(value); err != nil {
			return
		}

		data, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), data)
	})
}

// Get will get a value for the given key in a bucket(e.g. table).
func Get(bucket string, key string, val interface{}, readonly bool) (err error) {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Trying to access a nonexisting bucket.")
		}

		if !readonly {
			Lock(bucket, key)
		}

		data := b.Get([]byte(key))
		if data == nil {
			return errors.New("No data in database.")
		}

		buf := &bytes.Buffer{}
		decoder := gob.NewDecoder(buf)

		n, _ := buf.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		return decoder.Decode(val)
	})
}

// Has will check if the key is pressent in the database.
func Has(bucket, key string) bool {
	found := false

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Unknown bucket")
		}
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if key == string(k) {
				found = true
				break
			}
		}

		return nil
	})

	if err != nil {
		return false
	}

	return found
}

// Remove will delete a key in specified bucket.
func Remove(bucket, key string) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(key))
	})
}

// RegisterBucket Will store all bucket names reserved by other packages. When
// the database is started these bucket names will be made sure exists in the DB.
func RegisterBucket(bucket string) (err error) {
	registeredBucketNames = append(registeredBucketNames, bucket)
	if db == nil {
		return
	}
	return db.Update(func(tx *bolt.Tx) (err error) {
		// Create a bucket.
		_, err = tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

// GetPureDB Returns the pure connection to the database. Can be used with more
// advanced DB interaction.
func GetPureDB() *bolt.DB {
	return db
}

// Close will shut down the database in a safe mather.
func Close() (err error) {
	return db.Close()
}

var writerslock sync.Mutex
var writerkeys = make(map[string]map[string]valueLocker)

type valueLocker struct {
	sync.Mutex
	islocked bool
}

// Lock will lock a specified key in a bucket for further use.
func Lock(bucket string, key string) {
	writerslock.Lock()
	defer writerslock.Unlock()

	if _, ok := writerkeys[bucket]; !ok {
		writerkeys[bucket] = make(map[string]valueLocker)
	}

	if _, ok := writerkeys[bucket][key]; !ok {
		writerkeys[bucket][key] = valueLocker{}
	}

	wkl := writerkeys[bucket][key]
	wkl.Lock()
	wkl.islocked = true
	writerkeys[bucket][string(key)] = wkl
}

// Unlock will unlock a specified key in a bucket and make it usable for other
// tasks running.
func Unlock(bucket string, key string) {
	writerslock.Lock()
	defer writerslock.Unlock()

	if _, ok := writerkeys[bucket]; !ok {
		return
	}

	if _, ok := writerkeys[bucket][key]; !ok {
		return
	}

	wkl := writerkeys[bucket][key]
	if wkl.islocked {
		wkl.Unlock()
		wkl.islocked = false
	}
	writerkeys[bucket][key] = wkl
}
