package database

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
)

// Marshal encodes the val object into a []byte.
func Marshal(val interface{}) ([]byte, error) {
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

// Unmarshal decodes data into the object val.
func Unmarshal(data []byte, val interface{}) error {
	buf := &bytes.Buffer{}
	decoder := gob.NewDecoder(buf)
	// Write to buf will write all data and return err=nil
	buf.Write(data)
	return decoder.Decode(val)
}
