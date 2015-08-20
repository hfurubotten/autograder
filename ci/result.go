package ci

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hfurubotten/ag-scoring/score"
	"github.com/hfurubotten/autograder/database"
)

func init() {
	gob.Register(BuildResult{})

	database.RegisterBucket(BuildBucketName)
}

// BuildBucketName is the bucket/table name used in the database.
var BuildBucketName = "buildresults"

// BuildLengthKey is the key name used to increment build IDs.
var BuildLengthKey = "lenght"

// BuildResult represent a result from a test build.
type BuildResult struct {
	ID     int
	Course string
	User   string
	Group  int

	Log             []string
	NumPasses       int
	NumFails        int
	NumBuildFailure int
	Status          string
	Labnum          int

	Timestamp time.Time
	PushTime  time.Time

	TestScores []score.Score
	TotalScore int

	HeadCommitID   string
	HeadCommitText string
	CommitIDs      []string
	CommitTexts    []string

	BuildTime time.Duration

	Contributions map[string]int

	lock sync.Mutex
}

// NewBuildResult will create a new build result object.
func NewBuildResult() (*BuildResult, error) {
	nextid := GetNextBuildID()
	if nextid < 0 {
		return nil, errors.New("Error occured while generating Build ID")
	}
	return &BuildResult{
		ID:         nextid,
		TestScores: make([]score.Score, 0),
		Log:        make([]string, 0),
	}, nil
}

// GetBuildResult will find a build result on its ID.
func GetBuildResult(buildid int) (*BuildResult, error) {
	br := new(BuildResult)
	br.ID = buildid

	if err := br.loadStoredData(false); err != nil {
		return nil, err
	}

	return br, nil
}

func (br *BuildResult) loadStoredData(lock bool) error {
	return database.GetPureDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BuildBucketName))
		if b == nil {
			return errors.New("Bucket not found. Are you sure the bucket was registered correctly?")
		}

		data := b.Get([]byte(strconv.Itoa(br.ID)))
		if data == nil {
			return errors.New("No data in database")
		}

		buf := &bytes.Buffer{}
		decoder := gob.NewDecoder(buf)

		n, _ := buf.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		err := decoder.Decode(br)
		if err != nil {
			return err
		}

		return nil
	})
}

// Lock will put a writers lock on the build results.
func (br *BuildResult) Lock() {
	br.lock.Lock()
}

// Unlock will remove the writers lock on the build result.
func (br *BuildResult) Unlock() {
	br.lock.Unlock()
}

// Save will store the build results to the database.
func (br *BuildResult) Save() error {
	return database.GetPureDB().Update(func(tx *bolt.Tx) error {
		// open the bucket
		b := tx.Bucket([]byte(BuildBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		buf := &bytes.Buffer{}
		encoder := gob.NewEncoder(buf)

		if err := encoder.Encode(br); err != nil {
			return err
		}

		data, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		err = b.Put([]byte(strconv.Itoa(br.ID)), data)
		if err != nil {
			return err
		}

		return nil
	})
}

// GetNextBuildID will find the next available build ID.
// returns -1 on error
func GetNextBuildID() int {
	nextid := -1
	if err := database.GetPureDB().Update(func(tx *bolt.Tx) error {
		// open the bucket
		b := tx.Bucket([]byte(BuildBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		var err error
		data := b.Get([]byte(BuildLengthKey))
		if data == nil {
			nextid = 0
		} else {
			nextid, err = strconv.Atoi(string(data))
			if err != nil {
				return err
			}
		}

		nextid++

		data = []byte(strconv.Itoa(nextid))

		err = b.Put([]byte(BuildLengthKey), data)
		if err != nil {
			return err
		}

		return nil

	}); err != nil {
		return -1
	}

	return nextid
}
