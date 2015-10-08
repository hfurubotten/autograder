package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

var tmploc = "test.db"
var tmpbucket = "test"
var agentBucket = "agent"

// TestStart test the start function in the database package.
func TestStart(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function. " + err.Error())
	}
	removeDB()
}

// testStringValues has string values for the TestPutGet function.
var testStringValues = []struct {
	key   string
	value interface{}
}{
	{"hei", "sann"},
	{"key", "value"},
	{"sponge", "bob"},
	{"square", "pants"},
}

type agent struct {
	Name    string
	DoubleO int
}

// testAgentValues has the agent object values for the TestPutGet function.
var testAgentValues = []struct {
	key   string
	value agent
}{
	{"agent", agent{"James Bond", 7}},
	{"mi", agent{"Ethan Hunt", 1}},
}

// TestPutGet will test the Put and Get functions.
func TestPutGet(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function." + err.Error())
	}

	for _, v := range testStringValues {
		err = Put(tmpbucket, v.key, v.value)
		if err != nil {
			t.Error(err)
		}
		var got string
		err = Get(tmpbucket, v.key, &got)
		if err != nil {
			t.Error(err)
		}
		if got != v.value {
			t.Errorf("Got %s wanted %s", got, v.key)
		}
	}

	for _, v := range testAgentValues {
		err = Put(agentBucket, v.key, v.value)
		if err != nil {
			t.Error(err)
		}
		var got agent
		err = Get(agentBucket, v.key, &got)
		if err != nil {
			t.Error(err)
		}
		if got.DoubleO != v.value.DoubleO || got.Name != v.value.Name {
			t.Errorf("Got %v wanted %v", got, v.key)
		}
	}

	cleanUpBucket()
	removeDB()
}

// BenchmarkPutGetString will benchmark the Put and Get functions.
func BenchmarkPutGetString(b *testing.B) {
	// we ignore errors in this benchmark test
	Start(tmploc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range testStringValues {
			Put(tmpbucket, v.key, v.value)
			var got string
			Get(tmpbucket, v.key, &got)
		}
	}
	removeDB()
}

// BenchmarkPutGetObject will benchmark the Put and Get functions.
func BenchmarkPutGetObject(b *testing.B) {
	// we ignore errors in this benchmark test
	Start(tmploc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range testAgentValues {
			Put(agentBucket, v.key, v.value)
			var got agent
			Get(agentBucket, v.key, &got)
		}
	}
	removeDB()
}

// BenchmarkPutGetDiffKey will benchmark the Put and Get functions.
func BenchmarkPutGetDiffKey(b *testing.B) {
	// we ignore errors in this benchmark test
	Start(tmploc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		Put(tmpbucket, key, "v.value")
		var got string
		Get(tmpbucket, key, &got)
	}
	removeDB()
}

// BenchmarkPutKey will benchmark the Put function.
func BenchmarkPutKey(b *testing.B) {
	// we ignore errors in this benchmark test
	Start(tmploc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		Put(tmpbucket, key, "v.value")
	}
	removeDB()
}

// BenchmarkGetKey will benchmark the Get function.
func BenchmarkGetKey(b *testing.B) {
	// we ignore errors in this benchmark test
	Start(tmploc)
	Put(tmpbucket, "key", "v.value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got string
		Get(tmpbucket, "key", &got)
	}
	removeDB()
}

func TestHas(t *testing.T) {
	err := Start(tmploc)
	if err != nil {
		t.Error("Got error while executing start function." + err.Error())
	}

	for _, v := range testStringValues {
		err = Put(tmpbucket, v.key, v.value)
		if err != nil {
			t.Error(err)
		}
		if !Has(tmpbucket, v.key) {
			t.Errorf("Key %s not found in bucket %s", v.key, tmpbucket)
		}
	}
	// test also that it doesn't always return true
	if Has(tmpbucket, "this key shouldn't be in there") {
		t.Errorf("Key was unexpectedly found in bucket %s", tmpbucket)
	}

	cleanUpBucket()
	removeDB()
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

	removeDB()
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
func removeDB() {
	db.Close()
	if err := os.Remove(tmploc); err != nil {
		panic("Couldn't remove database." + err.Error())
	}
}
