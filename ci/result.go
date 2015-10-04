package ci

import (
	"strconv"
	"time"

	"github.com/autograde/kit/score"
	"github.com/hfurubotten/autograder/database"
)

func init() {
	// gob.Register(BuildResult{})
	//
	database.RegisterBucket(BuildBucketName)
}

// BuildBucketName is the bucket/table name used in the database.
var BuildBucketName = "buildresults"

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
}

// NewBuildResult will create a new build result object.
func NewBuildResult() (*BuildResult, error) {
	nextid, err := database.NextID(BuildBucketName)
	if err != nil {
		return nil, err
	}
	return &BuildResult{
		ID:         int(nextid),
		TestScores: make([]score.Score, 0),
		Log:        make([]string, 0),
	}, nil
}

// GetBuildResult will find a build result on its ID.
func GetBuildResult(buildID int) (br *BuildResult, err error) {
	key := strconv.Itoa(buildID)
	err = database.Get(BuildBucketName, key, &br)
	return br, err
}

// Save stores the build results in the database.
func (br *BuildResult) Save() error {
	key := strconv.Itoa(br.ID)
	return database.Put(BuildBucketName, key, br)
}
