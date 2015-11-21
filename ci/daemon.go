package ci

import (
	"bufio"
	"bytes"
	// "errors"
	"fmt"
	"log"

	git "github.com/hfurubotten/autograder/entities"
)

//TODO Rename: DaemonOptions -> TestParameters

// DaemonOptions represent the options needed to start the testing daemon.
type DaemonOptions struct {
	Org       string
	User      string
	GroupName string

	UserRepo   string
	TestRepo   string
	BaseFolder string
	LabFolder  string
	LabNumber  int
	DestFolder string
	IsPush     bool

	AdminToken string
	Secret     string
}

// StartTesterDaemon will start a new test build in the background.
//TODO this functions is too long. needs to be split into shorter functions.
func StartTesterDaemon(opt DaemonOptions) {
	// safeguard
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()

	// Test execution
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

	r, err := NewBuildResult(opt)
	if err != nil {
		log.Println(err)
		return
	}

	startMsg := fmt.Sprintf("Running tests for: %s/%s", opt.Org, opt.UserRepo)
	log.Println(startMsg)
	r.Add(startMsg, opt)

	runCommands(env, r, opt)
	r.Done()

	defer func() {
		// saves the build results
		if err := r.Save(); err != nil {
			log.Println("Error saving build results:", err)
			return
		}
	}()

	// Build for group assignment. Stores build ID in group.
	if len(opt.GroupName) > 0 {
		//TODO move to groups.go
		group, err := git.GetGroup(opt.GroupName)
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
		//TODO move to members.go
		user, err := git.GetMember(opt.User)
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

func runCommands(env Virtual, r *BuildResult, opt DaemonOptions) {
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
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + opt.UserRepo + ".git" + " " + opt.BaseFolder + opt.DestFolder + "/", true},
		{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + git.TestRepoName + ".git" + " " + opt.BaseFolder + git.TestRepoName + "/", true},
		{"/bin/bash -c \"cp -rf \"" + opt.BaseFolder + git.TestRepoName + "/*\" \"" + opt.BaseFolder + opt.DestFolder + "/\" \"", true},

		{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/dependencies.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/\" && ./dependencies.sh)\"", true},
		{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/test.sh", true},
		{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/\" && ./test.sh)\"", false},
	}

	// executes build commands
	for _, cmd := range cmds {
		err := execute(&env, cmd.Cmd, r, opt)
		if err != nil {
			r.Add(err.Error(), opt)
			log.Println(err)
			if cmd.Breakable {
				r.Add("Unexpected end of integration.", opt)
				break
			}
		}
	}
}

func execute(v *Virtual, cmd string, l *BuildResult, opt DaemonOptions) error {
	buf := bytes.NewBuffer(make([]byte, 0))
	bufw := bufio.NewWriter(buf)

	//TODO fmt?
	fmt.Println("$", cmd)

	err := v.Execute(cmd, nil, bufw, bufw)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(buf)
	for s.Scan() {
		text := s.Text()
		l.Add(text, opt)
	}
	return nil
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
