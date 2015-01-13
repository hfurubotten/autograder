package ci

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"strconv"
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

func StartTesterDeamon(load git.HookPayload) {
	// safeguard
	/*defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()*/

	if !git.HasMember(load.User) {
		log.Println("Not a valid user: ", load.User)
		return
	}

	if !git.HasOrganization(load.Organization) {
		log.Println("Not a valid org: ", load.Organization)
		return
	}

	logarray := make([]string, 0)
	logarray = append(logarray, "CI starting up on repo "+load.Fullname)
	org := git.NewOrganization(load.Organization)
	user := git.NewMemberFromUsername(load.User)

	isgroup := !strings.Contains(load.Repo, "-labs")

	var labfolder string
	var labnum int
	if isgroup {
		gnum, err := strconv.Atoi(load.Repo[len("group"):])
		if err != nil {
			panic(err)
		}

		group, err := git.NewGroup(org.Name, gnum)
		if err != nil {
			panic(err)
		}

		labnum = group.CurrentLabNum
		labfolder = org.GroupLabFolders[labnum]
	} else {
		labnum = user.Courses[org.Name].CurrentLabNum
		labfolder = org.IndividualLabFolders[labnum]
	}

	basefolder := org.CI.Basepath

	// Test execution
	log.Println("CI starting up on repo", load.Fullname)

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

	cmds := []string{
		"mkdir -p " + basefolder,
		"git clone https://" + org.AdminToken + ":x-oauth-basic@github.com/" + org.Name + "/" + load.Repo + ".git" + " " + basefolder + load.Repo + "/",
		"git clone https://" + org.AdminToken + ":x-oauth-basic@github.com/" + org.Name + "/" + git.TEST_REPO_NAME + ".git" + " " + basefolder + git.TEST_REPO_NAME + "/",
		"/bin/bash -c \"cp -rf \"" + basefolder + git.TEST_REPO_NAME + "/*\" \"" + basefolder + load.Repo + "/\" \"",

		"chmod 777 " + basefolder + load.Repo + "/dependencies.sh",
		"/bin/sh -c \"(cd \"" + basefolder + load.Repo + "/\" && ./dependencies.sh)\"",
		"chmod 777 " + basefolder + load.Repo + "/" + labfolder + "/test.sh",
		"/bin/sh -c \"(cd \"" + basefolder + load.Repo + "/" + labfolder + "/\" && ./test.sh)\"",
	}

	for _, cmd := range cmds {
		err = execute(&env, cmd, &logarray)
		if err != nil {
			logOutput("Unexpected end of integration.", &logarray)
			log.Println(err)
			break
		}
	}

	// parsing the results
	r := Result{
		Log:       logarray,
		Course:    org.Name,
		Labnum:    labnum,
		Timestamp: time.Now(),
	}

	r.User = user.Username

	var name string
	if isgroup {
		name = load.Repo
	} else {
		name = user.Username
	}

	parseResults(&r)

	teststore := getCIStorage(org.Name, name)

	err = teststore.WriteGob(labfolder, r)
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

func getCIStorage(course, user string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:     global.Basepath + "diskv/CI/" + course + "/" + user,
		CacheSizeMax: 1024 * 1024 * 256,
	})
}

func execute(v *Virtual, cmd string, l *[]string) (err error) {

	read := bytes.NewBuffer(make([]byte, 0))

	fmt.Println("$", cmd)

	err = v.ExecuteCommand(cmd, nil, bufio.NewWriter(read), bufio.NewWriter(read))

	s := bufio.NewScanner(read)

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
	teststore := getCIStorage(org, user)

	if !teststore.Has(lab) {
		err = errors.New("Doesn't have any CI logs yet.")
		return
	}

	err = teststore.ReadGob(lab, &logs, false)
	return
}
