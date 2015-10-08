package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var registeredBucketNames = make([]string, 0)

// Start will start the database using the provided dbloc file.
// If the database does not exist, a new database will be created.
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

// Put associates the given key and value in the provided bucket.
// The value can be any type.
func Put(bucket string, key string, value interface{}) (err error) {
	return db.Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			// if bucket didn't exist, create the bucket.
			b, err = tx.CreateBucket([]byte(bucket))
			if err != nil {
				return err
			}
		}
		data, err := GobEncode(value)
		return b.Put([]byte(key), data)
	})
}

// GobEncode encodes the val object into a []byte.
func GobEncode(val interface{}) ([]byte, error) {
	//TODO I wonder if there are simpler ways to marshal the value
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(val); err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Get the value associated with the given key in the provided bucket.
// The provided value must be an address.
func Get(bucket string, key string, val interface{}) (err error) {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("unknown bucket: " + bucket)
		}
		data := b.Get([]byte(key))
		if data == nil {
			return errors.New("key '" + key + "' not found in bucket: " + bucket)
		}
		return GobDecode(data, val)
	})
}

// GobDecode decodes data into the object val.
func GobDecode(data []byte, val interface{}) error {
	buf := &bytes.Buffer{}
	decoder := gob.NewDecoder(buf)
	// Write to buf will write all data and return err=nil
	buf.Write(data)
	return decoder.Decode(val)
}

// Has returns true the key is present in the given bucket.
func Has(bucket, key string) bool {
	found := false
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("unknown bucket: " + bucket)
		}
		data := b.Get([]byte(key))
		if data != nil {
			found = true
		}
		return nil
	})
	// if an error was returned, found will still be false
	return found
}

// Remove will delete the given key in the specified bucket.
func Remove(bucket, key string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(key))
	})
}

// RegisterBucket will store all bucket names reserved by other packages. When
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

// GetPureDB returns the pure connection to the database. Can be used with more
// advanced DB interaction.
// TODO Avoid using this method
func GetPureDB() *bolt.DB {
	if db == nil {
		panic("Trying to obtain uninitalized database")
	}
	return db
}

// Close will shut down the database in a safe mather.
func Close() (err error) {
	return db.Close()
}
