package web

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	ci "github.com/hfurubotten/autograder/ci"
	"github.com/hfurubotten/autograder/git"
)

func webhookeventhandler(w http.ResponseWriter, r *http.Request) {
	load, err := git.DecodeHookPayload(r.Body)
	if err != nil {
		log.Println("Error: ", err)
	}

	if !git.HasMember(load.User) {
		log.Println("Not a valid user: ", load.User)
		return
	}

	if !git.HasOrganization(load.Organization) {
		log.Println("Not a valid org: ", load.Organization)
		return
	}

	org := git.NewOrganization(load.Organization)
	user := git.NewMemberFromUsername(load.User)

	isgroup := !strings.Contains(load.Repo, "-"+git.STANDARD_REPO_NAME)

	var labfolder string
	var labnum int
	var username string
	if isgroup {
		gnum, err := strconv.Atoi(load.Repo[len("group"):])
		if err != nil {
			log.Println(err)
			return
		}

		group, err := git.NewGroup(org.Name, gnum)
		if err != nil {
			log.Println(err)
			return
		}

		labnum = group.CurrentLabNum
		labfolder = org.GroupLabFolders[labnum]
		username = load.Repo
	} else {
		labnum = user.Courses[org.Name].CurrentLabNum
		labfolder = org.IndividualLabFolders[labnum]
		username = strings.TrimRight(load.Repo, "-"+git.STANDARD_REPO_NAME)
	}

	opt := ci.DaemonOptions{
		Org:          org.Name,
		User:         username,
		Repo:         load.Repo,
		BaseFolder:   org.CI.Basepath,
		LabFolder:    labfolder,
		AdminToken:   org.AdminToken,
		MimicLabRepo: true,
	}

	go ci.StartTesterDaemon(opt)
}
