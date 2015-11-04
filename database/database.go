package database

import (
	"errors"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var (
	db                    *bolt.DB
	registeredBucketNames = make([]string, 0)
	databaseFileName      = "database"
)

// Start will start the database using the provided dbloc path.
// If the database does not exist, a new database will be created.
func Start(dbpath string) (err error) {
	db, err = bolt.Open(filepath.Join(dbpath, databaseFileName), 0666, nil)
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
		data, err := Marshal(value)
		// data, err := json.Marshal(value)
		return b.Put([]byte(key), data)
	})
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
			return KeyNotFoundError{key, bucket}
		}
		return Unmarshal(data, val)
		// return json.Unmarshal(data, val)
	})
}

// ForEach iterates over all keys in bucket, evaluating the function fn.
func ForEach(bucket string, fn func(k, v []byte) error) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("unknown bucket: " + bucket)
		}
		return b.ForEach(fn)
	})
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
