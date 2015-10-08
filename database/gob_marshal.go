package database

import (
	"bytes"
	"encoding/gob"
)

// Marshal encodes the val object into a []byte.
func Marshal(val interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes data into the object val.
func Unmarshal(data []byte, val interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(val)
}
