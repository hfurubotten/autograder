package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hfurubotten/autograder/ci"
	"github.com/hfurubotten/autograder/git"
)

func manualcihandler(w http.ResponseWriter, r *http.Request) {
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

	org := git.NewOrganization(course)

	var repo string
	if _, ok := org.Members[user]; ok {
		repo = user + "-" + git.STANDARD_REPO_NAME
	} else if _, ok := org.Groups[user]; ok {
		repo = user
	} else {
		http.Error(w, "Unknown user", 404)
		return
	}

	_, ok1 := member.Teaching[course]
	_, ok2 := member.AssistantCourses[course]

	if !ok1 && !ok2 {
		if _, ok := org.Members[member.Username]; ok {
			user = member.Username
		} else {
			http.Error(w, "Not a member of the course", 404)
			return
		}
	}

	opt := ci.DaemonOptions{
		Org:          org.Name,
		User:         user,
		Repo:         repo,
		BaseFolder:   org.CI.Basepath,
		LabFolder:    lab,
		AdminToken:   org.AdminToken,
		MimicLabRepo: true,
	}

	log.Println(opt)

	ci.StartTesterDaemon(opt)
}

func ciresulthandler(w http.ResponseWriter, r *http.Request) {
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
