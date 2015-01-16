package ci

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

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

	for _, cmd := range cmds {
		err = execute(&env, cmd.Cmd, &logarray)
		if err != nil && cmd.Breakable {
			logOutput("Unexpected end of integration.", &logarray)
			log.Println(err)
			break
		}
	}

	// parsing the results
	r := Result{
		Log:       logarray,
		Course:    opt.Org,
		Timestamp: time.Now(),
		User:      opt.User,
		Status:    "Active lab assignment",
	}

	parseResults(&r)

	teststore := GetCIStorage(opt.Org, opt.User)

	if teststore.Has(opt.LabFolder) {
		oldr := Result{}
		err = teststore.ReadGob(opt.LabFolder, &oldr, false)
		r.Status = oldr.Status
	}

	err = teststore.WriteGob(opt.LabFolder, r)
	if err != nil {
		panic(err)
	}

}

func parseResults(r *Result) {
	key := "--- PASS"
	for _, l := range r.Log {
		r.NumPasses = r.NumPasses + strings.Count(l, key)
	}

	log.Println("Found ", r.NumPasses, " passed tests.")
}

func GetCIStorage(course, user string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:     global.Basepath + "diskv/CI/" + course + "/" + user,
		CacheSizeMax: 1024 * 1024 * 256,
	})
}

func execute(v *Virtual, cmd string, l *[]string) (err error) {

	buf := bytes.NewBuffer(make([]byte, 0))
	bufw := bufio.NewWriter(buf)

	fmt.Println("$", cmd)

	err = v.ExecuteCommand(cmd, nil, bufw, bufw)

	s := bufio.NewScanner(buf)

	for s.Scan() {
		text := s.Text()
		logOutput(text, l)
	}

	return
}

func logOutput(s string, l *[]string) {
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

	*l = append(*l, strings.TrimSpace(s))
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
