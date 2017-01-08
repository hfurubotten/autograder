package ci

import (
	"bufio"
	"bytes"
	"encoding/json"
	// "errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/autograde/kit/score"
	git "github.com/hfurubotten/autograder/entities"
)

// DaemonOptions represent the options needed to start the testing daemon.
type DaemonOptions struct {
	Org   string
	User  string
	Group int

	Repo       string
	BaseFolder string
	LabFolder  string
	LabNumber  int
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

	r, err := NewBuildResult()
	if err != nil {
		log.Println(err)
		return
	}

	r.Log = logarray
	r.Course = opt.Org
	r.Timestamp = time.Now()
	r.PushTime = time.Now()
	r.User = opt.User
	r.Status = "Active lab assignment"
	r.Labnum = opt.LabNumber

	starttime := time.Now()

	// executes build commands
	for _, cmd := range cmds {
		err = execute(&env, cmd.Cmd, r, opt)
		if err != nil {
			logOutput(err.Error(), r, opt)
			log.Println(err)
			if cmd.Breakable {
				logOutput("Unexpected end of integration.", r, opt)
				break
			}
		}
	}

	r.BuildTime = time.Since(starttime)

	// parsing the results
	SimpleParsing(r)
	if len(r.TestScores) > 0 {
		r.TotalScore = score.Total(r.TestScores)
	} else {
		// TODO move this computation to kit/score package
		if r.NumPasses+r.NumFails != 0 {
			r.TotalScore = int((float64(r.NumPasses) / float64(r.NumPasses+r.NumFails)) * 100.0)
		}
	}

	// check for stack overflow in log
	for _, l := range r.Log {
		if strings.Contains(l, "fatal error: stack overflow") {
			r.TotalScore = 0
			break
		}
	}

	if r.NumBuildFailure > 0 {
		r.TotalScore = 0
	}

	defer func() {
		// saves the build results
		if err := r.Save(); err != nil {
			log.Println("Error saving build results:", err)
			return
		}
	}()

	// Build for group assignment. Stores build ID in group.
	if opt.Group > 0 {
		group, err := git.NewGroup(opt.Org, opt.Group, false)
		if err != nil {
			log.Println(err)
			return
		}

		oldbuildID := group.GetLastBuildID(opt.LabNumber)
		if oldbuildID > 0 {
			oldr, err := GetBuildResult(oldbuildID)
			if err != nil {
				log.Println(err)
				return
			}
			r.Status = oldr.Status
			if !opt.IsPush {
				r.PushTime = oldr.PushTime
			}
		}

		group.AddBuildResult(opt.LabNumber, r.ID)

		if err := group.Save(); err != nil {
			group.Unlock()
			log.Println(err)
		}
		// build for single user. Stores build ID to user.
	} else {
		user, err := git.NewMemberFromUsername(opt.User, false)
		if err != nil {
			log.Println(err)
			return
		}

		oldbuildID := user.GetLastBuildID(opt.Org, opt.LabNumber)
		if oldbuildID > 0 {
			oldr, err := GetBuildResult(oldbuildID)
			if err != nil {
				log.Println(err)
				return
			}
			r.Status = oldr.Status
			if !opt.IsPush {
				r.PushTime = oldr.PushTime
			}
		}

		user.AddBuildResult(opt.Org, opt.LabNumber, r.ID)

		if err := user.Save(); err != nil {
			user.Unlock()
			log.Println(err)
		}
	}
}

// SimpleParsing will do a simple parsing of the test results. It looks for the strings "--- PASS", "--- FAIL" and "build failed".
func SimpleParsing(r *BuildResult) {
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

func execute(v *Virtual, cmd string, l *BuildResult, opt DaemonOptions) (err error) {

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

func logOutput(s string, l *BuildResult, opt DaemonOptions) {
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

	//TODO: Move this code to a new function in kit/score package? Reason: easier to test.
	if strings.Contains(s, opt.Secret) {
		// TODO: must be a better way of detecting JSON data!  TODO: Hein@Heine: Why?
		var testscore score.Score
		err := json.Unmarshal([]byte(s), &testscore)
		if err == nil {
			if testscore.Secret == opt.Secret {
				testscore.Secret = "Sanitized"
				l.TestScores = append(l.TestScores, &testscore)
			}
			return
		}
		// ensure that the error message does not reveal the secret token
		es := strings.Replace(err.Error(), opt.Secret, "Sanitized", -1)
		log.Printf("Parse error: %s\n", es)
	}
	s = strings.Replace(s, opt.Secret, "Sanitized", -1)
	s = strings.Replace(s, opt.AdminToken, "Sanitized", -1)

	l.Log = append(l.Log, strings.TrimSpace(s))
	fmt.Println(s)
}

// // GetIntegationResults will find a test result for a user or group.
// func GetIntegationResults(org, user, lab string) (logs Result, err error) {
// 	teststore := GetCIStorage(org, user)
//
// 	if !teststore.Has(lab) {
// 		err = errors.New("Doesn't have any CI logs yet.")
// 		return
// 	}
//
// 	err = teststore.ReadGob(lab, &logs, false)
// 	return
// }
//
// // GetIntegationResultSummary will return a summary of the test results for a user or a group.
// func GetIntegationResultSummary(org, user string) (summary map[string]Result, err error) {
// 	summary = make(map[string]Result)
// 	teststore := GetCIStorage(org, user)
// 	keys := teststore.Keys()
// 	for key := range keys {
// 		var res Result
// 		err = teststore.ReadGob(key, &res, false)
// 		if err != nil {
// 			return
// 		}
// 		res.Log = make([]string, 0)
// 		summary[key] = res
// 	}
// 	return
// }
