package ci

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/autograde/kit/score"
	"github.com/hfurubotten/autograder/database"
)

const hidden = "Sanitized"

// BuildResult represent a result from a test build.
type BuildResult struct {
	ID     int
	Course string
	User   string
	Group  int

	log []string
	//TODO unexport / lower case these:
	NumPasses       int
	NumFails        int
	numBuildFailure int
	Status          string
	Labnum          int

	Timestamp time.Time //TODO This is never used elsewhere. What is it meant to represent?
	PushTime  time.Time
	BuildTime time.Duration

	TestScores []*score.Score
	TotalScore int

	HeadCommitID   string
	HeadCommitText string
	CommitIDs      []string
	CommitTexts    []string

	Contributions map[string]int
}

var buildBucketName = "buildresults"

func init() {
	database.RegisterBucket(buildBucketName)
}

// NewBuildResult will create a new build result object.
func NewBuildResult(opt DaemonOptions) (*BuildResult, error) {
	nextid, err := database.NextID(buildBucketName)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()
	return &BuildResult{
		ID:         int(nextid),
		Course:     opt.Org,
		User:       opt.User,
		Status:     "Active lab assignment",
		Labnum:     opt.LabNumber,
		Timestamp:  startTime,
		PushTime:   startTime,
		TestScores: make([]*score.Score, 0),
		log:        make([]string, 0),
	}, nil
}

// Done records the build time and computes the test score.
func (br *BuildResult) Done() {
	br.BuildTime = time.Since(br.PushTime)

	if len(br.TestScores) > 0 {
		br.TotalScore = score.Total(br.TestScores)
	} else {
		if br.NumPasses+br.NumFails != 0 {
			br.TotalScore = int((float64(br.NumPasses) / float64(br.NumPasses+br.NumFails)) * 100.0)
		}
	}
	if br.numBuildFailure > 0 {
		br.TotalScore = 0
	}
}

// GetBuildResult returns the build result for the provided buildID.
func GetBuildResult(buildID int) (br *BuildResult, err error) {
	key := strconv.Itoa(buildID)
	err = database.Get(buildBucketName, key, &br)
	return br, err
}

// Save stores the build results in the database.
func (br *BuildResult) Save() error {
	key := strconv.Itoa(br.ID)
	return database.Put(buildBucketName, key, br)
}

// Add adds a line to the build results log and updates the test scores if
// any JSON score object are found.
func (br *BuildResult) Add(s string, opt DaemonOptions) {
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		s = string(v)
	}
	s = strings.Trim(s, string(0))
	s = strings.TrimSpace(s)

	// check for and parse JSON Score string
	if score.HasPrefix(s) {
		sc, err := score.Parse(s, opt.Secret)
		if err != nil {
			return
		}
		br.TestScores = append(br.TestScores, sc)
	}

	// remove any accidental secret output
	s = strings.Replace(s, opt.Secret, hidden, -1)
	s = strings.Replace(s, opt.AdminToken, hidden, -1)
	s = strings.TrimSpace(s)

	// append sanitized strong to log
	br.log = append(br.log, s)
	br.updateResultCount(s)
}

// TODO: Not sure if the following should be used. They are probably specific to
// Go and there is nothing that prevents students from inserting --- PASS.
var passStrings = []string{"--- PASS"}
var testFailStrings = []string{"--- FAIL"}
var buildFailStrings = []string{"build failed"}

//TODO Rename to updateBuildFailCount()
// updateResultCount searches the provided line for tests passed, failed, and
// build failures.
func (br *BuildResult) updateResultCount(line string) {
	for _, pass := range passStrings {
		br.NumPasses = br.NumPasses + strings.Count(line, pass)
	}
	for _, fail := range testFailStrings {
		br.NumFails = br.NumFails + strings.Count(line, fail)
	}
	for _, bfail := range buildFailStrings {
		br.numBuildFailure = br.numBuildFailure + strings.Count(line, bfail)
	}
}
