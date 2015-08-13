package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/ci"
	git "github.com/hfurubotten/autograder/entities"
)

// ManualCITriggerURL is the URL used to call ManualCITriggerHandler.
var ManualCITriggerURL = "/event/manualbuild"

// ManualCITriggerHandler is a http handler for manually triggering test builds.
func ManualCITriggerHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	course := r.FormValue("course")
	user := r.FormValue("user")
	lab := r.FormValue("lab")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown organization", 404)
		return
	}

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, "Organization Error", 404)
		return
	}

	// Defaults back to username or group name for the user if not a teacher.
	if !org.IsTeacher(member) {
		if org.IsMember(member) {
			if strings.Contains(user, "group") {
				if member.Courses[org.Name].IsGroupMember {
					user = "group" + strconv.Itoa(member.Courses[org.Name].GroupNum)
				} else {
					http.Error(w, "Not a group member", 404)
					return
				}
			} else {
				user = member.Username
			}
		} else {
			http.Error(w, "Not a member of the course", 404)
			return
		}
	}

	var repo string
	var destfolder string
	if _, ok := org.Members[user]; ok {
		repo = user + "-" + git.StandardRepoName
		destfolder = git.StandardRepoName
	} else if _, ok := org.Groups[user]; ok {
		repo = user
		destfolder = git.GroupsRepoName
	} else {
		http.Error(w, "Unknown user", 404)
		return
	}

	opt := ci.DaemonOptions{
		Org:        org.Name,
		User:       user,
		Repo:       repo,
		BaseFolder: org.CI.Basepath,
		LabFolder:  lab,
		AdminToken: org.AdminToken,
		DestFolder: destfolder,
		Secret:     org.CI.Secret,
		IsPush:     false,
	}

	log.Println(opt)

	ci.StartTesterDaemon(opt)
}

// CIResultURL is the URL used to call CIResultURL.
var CIResultURL = "/course/ciresutls"

// CIResultHandler is a http handeler for getting results from
// a build. This handler writes back the results as JSON data.
func CIResultHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	// TODO: add more security
	orgname := r.FormValue("Course")
	username := r.FormValue("Username")
	labname := r.FormValue("Labname")

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !org.IsMember(member) {
		http.Error(w, "Not a member for this course.", 404)
		return
	}

	var res *ci.BuildResult

	if strings.HasPrefix(username, git.GroupRepoPrefix) {
		labnum := -1
		for i, name := range org.GroupLabFolders {
			if name == labname {
				labnum = i
				break
			}
		}

		if labnum < 0 {
			http.Error(w, "No lab with that name found.", 404)
			return
		}

		groupid, err := strconv.Atoi(username[len(git.GroupRepoPrefix):])
		if err != nil {
			http.Error(w, "Could not convert the group ID.", 404)
			return
		}

		group, err := git.NewGroup(orgname, groupid, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		buildid := group.GetLastBuildID(labnum)
		if buildid < 0 {
			http.Error(w, "Could not find the build.", 404)
			return
		}

		res, err = ci.GetBuildResult(buildid)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
	} else {
		labnum := -1
		for i, name := range org.IndividualLabFolders {
			if name == labname {
				labnum = i
				break
			}
		}

		if labnum < 0 {
			http.Error(w, "No lab with that name found.", 404)
			return
		}

		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		buildid := user.GetLastBuildID(orgname, labnum)
		if buildid < 0 {
			http.Error(w, "Could not find the build.", 404)
			return
		}

		res, err = ci.GetBuildResult(buildid)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
	}

	enc := json.NewEncoder(w)

	err = enc.Encode(res)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}

}

// SummaryView is the struct used to store date for JSON writeback in CIResultSummaryHandler.
type SummaryView struct {
	Course  string
	User    string
	Summary map[string]*ci.BuildResult
}

// CIResultSummaryURL is the URL used to call CIResultSummaryURL.
var CIResultSummaryURL = "/course/cisummary"

// CIResultSummaryHandler is a http handler used to get a build summary
// of the build for a user or group. This handler writes back the summary
// as JSON data.
func CIResultSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	teacher, err := checkTeacherApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	orgname := r.FormValue("Course")
	username := r.FormValue("Username")

	if orgname == "" || username == "" {
		http.Error(w, "Empty request.", 404)
		return
	}

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !org.IsTeacher(teacher) {
		http.Error(w, "Not a teacher for this course.", 404)
		return
	}

	res := make(map[string]*ci.BuildResult)
	//if group ...
	if strings.HasPrefix(username, git.GroupRepoPrefix) {
		groupid, err := strconv.Atoi(username[len(git.GroupRepoPrefix):])
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		group, err := git.NewGroup(orgname, groupid, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		for labnum := range group.Assignments {
			labname := org.GroupLabFolders[labnum]
			buildid := group.GetLastBuildID(labnum)
			if buildid < 0 {
				continue
			}

			build, err := ci.GetBuildResult(buildid)
			if err != nil {
				log.Println(err)
				continue
			}
			res[labname] = build
		}
	} else {
		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		courseopt, ok := user.Courses[orgname]
		if ok {
			for labnum := range courseopt.Assignments {
				labname := org.IndividualLabFolders[labnum]
				buildid := user.GetLastBuildID(orgname, labnum)
				if buildid < 0 {
					continue
				}

				build, err := ci.GetBuildResult(buildid)
				if err != nil {
					log.Println(err)
					continue
				}
				res[labname] = build
			}
		}
	}

	view := SummaryView{
		Course:  orgname,
		User:    username,
		Summary: res,
	}

	enc := json.NewEncoder(w)

	err = enc.Encode(view)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
}
