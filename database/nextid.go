package database

import (
	"errors"

	"github.com/boltdb/bolt"
)

// NextID returns the next ID from the given bucket.
func NextID(bucket string) (id uint64, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("unknown bucket: " + bucket)
		}
		id, err = b.NextSequence()
		return err
	})

	return id, err
}
