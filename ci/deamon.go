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

	. "github.com/hfurubotten/autograder/ci/score"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Result{})
}

type DaemonOptions struct {
	Org          string
	User         string
	Repo         string
	BaseFolder   string
	LabFolder    string
	AdminToken   string
	MimicLabRepo bool
	Secret       string
	IsPush       bool
}

func StartTesterDaemon(opt DaemonOptions) {
	// safeguard
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()

	logarray := make([]string, 0)
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

	var destfolder string
	if opt.MimicLabRepo {
		destfolder = git.STANDARD_REPO_NAME
	} else {
		destfolder = opt.Repo
	}

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
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + opt.Repo + ".git" + " " + opt.BaseFolder + destfolder + "/", true},
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + git.TEST_REPO_NAME + ".git" + " " + opt.BaseFolder + git.TEST_REPO_NAME + "/", true},
		{"/bin/bash -c \"cp -rf \"" + opt.BaseFolder + git.TEST_REPO_NAME + "/*\" \"" + opt.BaseFolder + destfolder + "/\" \"", true},

		{"chmod 777 " + opt.BaseFolder + destfolder + "/dependencies.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + destfolder + "/\" && ./dependencies.sh)\"", true},
		{"chmod 777 " + opt.BaseFolder + destfolder + "/" + opt.LabFolder + "/test.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + destfolder + "/" + opt.LabFolder + "/\" && ./test.sh)\"", false},
	}

	r := Result{
		Log:        logarray,
		Course:     opt.Org,
		Timestamp:  time.Now(),
		PushTime:   time.Now(),
		User:       opt.User,
		Status:     "Active lab assignment",
		TestScores: make([]Score, 0),
	}

	for _, cmd := range cmds {
		err = execute(&env, cmd.Cmd, &r, opt)
		if err != nil && cmd.Breakable {
			logOutput("Unexpected end of integration.", &r, opt)
			log.Println(err)
			break
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
	log.Println(r)
	err = teststore.WriteGob(opt.LabFolder, r)
	if err != nil {
		panic(err)
	}

}

// Uses a array of Score objects to calculate a total score between 0 and 100.
func CalculateTestScore(s []Score) (total int) {
	total_weight := float32(0)
	weight := make([]float32, 0)
	score := make([]float32, 0)
	max := make([]float32, 0)
	for _, ts := range s {
		total_weight += float32(ts.Weight)
		weight = append(weight, float32(ts.Weight))
		score = append(score, float32(ts.Score))
		max = append(max, float32(ts.MaxScore))
	}

	tmp_total := float32(0)
	for i := 0; i < len(s); i = i + 1 {
		if score[i] > max[i] {
			score[i] = max[i]
		}
		tmp_total += ((score[i] / max[i]) * (weight[i] / total_weight))
	}

	return int(tmp_total * 100)
}

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

	var score Score
	// TODO: must be a better way of detecting JSON data!
	err := json.Unmarshal([]byte(s), &score)
	if err == nil {
		if score.Secret == opt.Secret {
			score.Secret = "Sanitized"
			l.TestScores = append(l.TestScores, score)
		}
		return
	}

	l.Log = append(l.Log, strings.TrimSpace(s))
	fmt.Println(s)
}

func GetIntegationResults(org, user, lab string) (logs Result, err error) {
	teststore := GetCIStorage(org, user)

	if !teststore.Has(lab) {
		err = errors.New("Doesn't have any CI logs yet.")
		return
	}

	err = teststore.ReadGob(lab, &logs, false)
	return
}

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
