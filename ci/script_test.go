package ci

import (
	"fmt"
	"testing"

	git "github.com/hfurubotten/autograder/entities"
)

// opt := ci.DaemonOptions{
//   Org:   org.Name,
//   User:  user,
//   Group: groupid,
//
//   Repo:       repo,
//   BaseFolder: org.CI.Basepath,
//   LabFolder:  lab,
//   LabNumber:  labnum,
//   AdminToken: org.AdminToken,
//   DestFolder: destfolder,
//   Secret:     org.CI.Secret,
//   IsPush:     false,
// }

func TestScript(t *testing.T) {
	opt := DaemonOptions{
		Org:        "uis-dat320",
		User:       "meling",
		GroupName:  "1",
		UserRepo:   "meling-" + git.StandardRepoName,
		TestRepo:   git.TestRepoName,
		BaseFolder: "testground/",
		DestFolder: git.StandardRepoName,
		// DestFolder: git.GroupsRepoName,
		LabFolder:  "lab1",
		LabNumber:  1,
		AdminToken: "h34jddzg29",
		Secret:     "a34h19hs3n",
		IsPush:     false,
	}
	script := Script(opt)
	for _, s := range script {
		if len(s) > 0 {
			fmt.Println(s)
		}
	}
}
