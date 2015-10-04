package database

import (
	"errors"

	"github.com/boltdb/bolt"
)

// NextID returns the next ID from the given bucket and key.
func NextID(bucket, key string) (id int, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("unknown bucket: " + bucket)
		}

		var er error
		data := b.Get([]byte(key))
		if data != nil {
			er = GobDecode(data, &id)
			if er != nil {
				return er
			}
		}

		id++
		data, er = GobEncode(id)
		if er != nil {
			return er
		}
		return b.Put([]byte(key), data)
	})

	return id, err
}
