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
	_, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	// TODO: add more security
	orgname := r.FormValue("Course")
	username := r.FormValue("Username")
	labname := r.FormValue("Labname")

	res, err := ci.GetIntegationResults(orgname, username, labname)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
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
	Summary map[string]ci.Result
}

// CIResultSummaryURL is the URL used to call CIResultSummaryURL.
var CIResultSummaryURL = "/course/cisummary"

// CIResultSummaryHandler is a http handler used to get a build summary
// of the build for a user or group. This handler writes back the summary
// as JSON data.
func CIResultSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	_, err := checkTeacherApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	// TODO: add more security
	orgname := r.FormValue("Course")
	username := r.FormValue("Username")

	if orgname == "" || username == "" {
		http.Error(w, "Empty request.", 404)
		return
	}

	res, err := ci.GetIntegationResultSummary(orgname, username)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
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
