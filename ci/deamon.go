package grader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/diskv"
)

var teststore = diskv.New(diskv.Options{
	BasePath:     "diskv/CI/",
	CacheSizeMax: 1024 * 1024 * 256,
	//Transform: func(s string) []string {
	//	path := strings.Split(s, ",")
	//	path = path[:len(path)-1]
	//	return path
	//},
})

func StartTesterDeamon(load git.HookPayload) {
	// safeguard
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()

	if !git.HasMember(load.User) {
		return
	}

	env, err := NewVirtual()
	if err != nil {
		panic(err)
	}

	err = env.NewContainer("autograder")
	if err != nil {
		panic(err)
	}

	// mkdir /testground/github.com/
	// git clone user-labs
	// git clone test-labs
	// cp test-labs user-labs
	// /bin/sh dependecies.sh
	// /bin/sh test.sh

	log.Println("CI starting up on repo", load.Fullname)

	logarray := make([]string, 0)
	logarray = append(logarray, "CI starting up on repo "+load.Fullname)
	org := git.NewOrganization(load.Organization)

	basefolder := "/testground/src/github.com/" + org.Name + "/"

	cmds := []string{
		"mkdir -p " + basefolder,
		"git clone http://" + org.AdminToken + ":x-oauth-basic@github.com/" + org.Name + "/" + load.Repo + ".git" + " " + basefolder + load.Repo + "/",
		"git clone http://" + org.AdminToken + ":x-oauth-basic@github.com/" + org.Name + "/" + git.TEST_REPO_NAME + ".git" + " " + basefolder + git.TEST_REPO_NAME + "/",
		"/bin/bash -c \"cp -rf \"" + basefolder + git.TEST_REPO_NAME + "/*\" \"" + basefolder + load.Repo + "/\" \"",

		"chmod 777 " + basefolder + load.Repo + "/dependencies.sh",
		"/bin/sh -c \"(cd \"" + basefolder + load.Repo + "/\" && ./dependencies.sh)\"",
		"chmod 777 " + basefolder + load.Repo + "/test.sh",
		"/bin/sh -c \"(cd \"" + basefolder + load.Repo + "/\" && ./test.sh)\"",
	}

	for _, cmd := range cmds {
		err = execute(&env, cmd, &logarray)
		if err != nil {
			log.Println(err)
			break
		}
	}

	err = teststore.WriteGob(org.Name+"/"+load.User+"/result.log", logarray)
	if err != nil {
		log.Println(err)
		return
	}
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

func GetIntegationResults(org, user string) (logs []string, err error) {
	if !teststore.Has(org + "/" + user + "/result.log") {
		return nil, errors.New("Doesn't have any CI logs yet.")
	}

	err = teststore.ReadGob(org+"/"+user+"/result.log", &logs, false)
	return
}
