package database

import "fmt"

// KeyNotFoundError reports that a key was not found in bucket.
type KeyNotFoundError struct {
	key, bucket string
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("key '%s' not found in bucket: %s", e.key, e.bucket)
}
