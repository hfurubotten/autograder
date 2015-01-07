package ci

import (
	"testing"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
)

func init() {
	global.Basepath = "/home/heinef/Dropbox/Mitt_Arbeid/Universitetet/Master_Thesis/Workspace/src/github.com/hfurubotten/autograder/"
}

func TestStartTesterDeamon(t *testing.T) {
	load := git.HookPayload{
		User:         "hfurubotten",
		Repo:         "testuser-labs",
		Fullname:     "uis-autograder-test/testuser-labs",
		Organization: "uis-autograder-test",
	}

	StartTesterDeamon(load)
}
