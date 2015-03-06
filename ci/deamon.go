package ci

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hfurubotten/ag-scoring/score"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Result{})
}

// DaemonOptions represent the options needed to start the testing daemon.
type DaemonOptions struct {
	Org        string
	User       string
	Repo       string
	BaseFolder string
	LabFolder  string
	AdminToken string
	DestFolder string
	Secret     string
	IsPush     bool
}

// StartTesterDaemon will start a new test build in the background.
func StartTesterDaemon(opt DaemonOptions) {
	// safeguard
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()

	var logarray []string
	logarray = append(logarray, "CI starting up on repo "+opt.Org+"/"+opt.Repo)

	// Test execution
	log.Println("CI starting up on repo", opt.Org, "/", opt.Repo)

	env, err := NewVirtual()
	if err != nil {
		panic(err)
	}

	err = env.NewContainer("autograder")
	if err != nil {
		panic(err)
	}

	// cleanup
	defer env.RemoveContainer()

	// mkdir /testground/github.com/
	// git clone user-labs
	// git clone test-labs
	// cp test-labs user-labs
	// /bin/sh dependecies.sh
	// /bin/sh test.sh

	cmds := []struct {
		Cmd       string
		Breakable bool
	}{
		{"mkdir -p " + opt.BaseFolder, true},
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + opt.Repo + ".git" + " " + opt.BaseFolder + opt.DestFolder + "/", true},
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + git.TestRepoName + ".git" + " " + opt.BaseFolder + git.TestRepoName + "/", true},
		{"/bin/bash -c \"cp -rf \"" + opt.BaseFolder + git.TestRepoName + "/*\" \"" + opt.BaseFolder + opt.DestFolder + "/\" \"", true},

		{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/dependencies.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/\" && ./dependencies.sh)\"", true},
		{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/test.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/\" && ./test.sh)\"", false},
	}

	r := Result{
		Log:        logarray,
		Course:     opt.Org,
		Timestamp:  time.Now(),
		PushTime:   time.Now(),
		User:       opt.User,
		Status:     "Active lab assignment",
		TestScores: make([]score.Score, 0),
	}

	for _, cmd := range cmds {
		err = execute(&env, cmd.Cmd, &r, opt)
		if err != nil {
			logOutput(err.Error(), &r, opt)
			log.Println(err)
			if cmd.Breakable {
				logOutput("Unexpected end of integration.", &r, opt)
				break
			}
		}
	}

	// parsing the results
	SimpleParsing(&r)
	if len(r.TestScores) > 0 {
		r.TotalScore = CalculateTestScore(r.TestScores)
	} else {
		if r.NumPasses+r.NumFails != 0 {
			r.TotalScore = int((float64(r.NumPasses) / float64(r.NumPasses+r.NumFails)) * 100.0)
		}
	}

	if r.NumBuildFailure > 0 {
		r.TotalScore = 0
	}

	teststore := GetCIStorage(opt.Org, opt.User)

	if teststore.Has(opt.LabFolder) {
		oldr := Result{}
		err = teststore.ReadGob(opt.LabFolder, &oldr, false)
		r.Status = oldr.Status
		if !opt.IsPush {
			r.PushTime = oldr.PushTime
		}
	}

	err = teststore.WriteGob(opt.LabFolder, r)
	if err != nil {
		panic(err)
	}

}

// CalculateTestScore uses a array of Score objects to calculate a total score between 0 and 100.
func CalculateTestScore(s []score.Score) (total int) {
	totalWeight := float32(0)
	var weight []float32
	var score []float32
	var max []float32
	for _, ts := range s {
		totalWeight += float32(ts.Weight)
		weight = append(weight, float32(ts.Weight))
		score = append(score, float32(ts.Score))
		max = append(max, float32(ts.MaxScore))
	}

	tmpTotal := float32(0)
	for i := 0; i < len(s); i = i + 1 {
		if score[i] > max[i] {
			score[i] = max[i]
		}
		tmpTotal += ((score[i] / max[i]) * (weight[i] / totalWeight))
	}

	return int(tmpTotal * 100)
}

// SimpleParsing will do a simple parsing of the test results. It looks for the strings "--- PASS", "--- FAIL" and "build failed".
func SimpleParsing(r *Result) {
	key := "--- PASS"
	negkey := "--- FAIL"
	bfkey := "build failed"
	for _, l := range r.Log {
		r.NumPasses = r.NumPasses + strings.Count(l, key)
		r.NumFails = r.NumFails + strings.Count(l, negkey)
		r.NumBuildFailure = r.NumBuildFailure + strings.Count(l, bfkey)
	}

	log.Println("Found ", r.NumPasses, " passed tests.")
}

// GetCIStorage will create a Diskv object used to store the test results.
func GetCIStorage(course, user string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:     global.Basepath + "diskv/CI/" + course + "/" + user,
		CacheSizeMax: 1024 * 1024 * 256,
	})
}

func execute(v *Virtual, cmd string, l *Result, opt DaemonOptions) (err error) {

	buf := bytes.NewBuffer(make([]byte, 0))
	bufw := bufio.NewWriter(buf)

	fmt.Println("$", cmd)

	err = v.ExecuteCommand(cmd, nil, bufw, bufw)

	s := bufio.NewScanner(buf)

	for s.Scan() {
		text := s.Text()
		logOutput(text, l, opt)
	}

	return
}

func logOutput(s string, l *Result, opt DaemonOptions) {
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

	var testscore score.Score
	if strings.Contains(s, opt.Secret) {
		// TODO: must be a better way of detecting JSON data!
		err := json.Unmarshal([]byte(s), &testscore)
		if err == nil {
			if testscore.Secret == opt.Secret {
				testscore.Secret = "Sanitized"
				l.TestScores = append(l.TestScores, testscore)
			}
			return
		}

		s = strings.Replace(s, opt.Secret, "Sanitized", -1)
	}

	if strings.Contains(s, opt.AdminToken) {
		s = strings.Replace(s, opt.AdminToken, "Sanitized", -1)
	}

	l.Log = append(l.Log, strings.TrimSpace(s))
	fmt.Println(s)
}

// GetIntegationResults will find a test result for a user or group.
func GetIntegationResults(org, user, lab string) (logs Result, err error) {
	teststore := GetCIStorage(org, user)

	if !teststore.Has(lab) {
		err = errors.New("Doesn't have any CI logs yet.")
		return
	}

	err = teststore.ReadGob(lab, &logs, false)
	return
}

// GetIntegationResultSummary will return a summary of the test results for a user or a group.
func GetIntegationResultSummary(org, user string) (summary map[string]Result, err error) {
	summary = make(map[string]Result)
	teststore := GetCIStorage(org, user)
	keys := teststore.Keys()
	for key := range keys {
		var res Result
		err = teststore.ReadGob(key, &res, false)
		if err != nil {
			return
		}
		res.Log = make([]string, 0)
		summary[key] = res
	}
	return
}
