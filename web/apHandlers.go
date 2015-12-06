package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	apProto "github.com/autograde/antiplagiarism/proto"
	git "github.com/hfurubotten/autograder/entities"
)

// ApManualTestURL is the URL used to call ApManualTestHandler.
var ApManualTestURL = "/event/apmanualtest"

// ApLabResultsURL is the URL used to call ApLabResultsHandler
var ApLabResultsURL = "/course/aplabresults"

// ApUserResultsURL is the URL used to call ApUserResultsHandler
var ApUserResultsURL = "/course/apuserresults"

// ApManualTestHandler is a http handler for manually triggering
// anti-plagiarism tests.
func ApManualTestHandler(w http.ResponseWriter, r *http.Request) {

	if !git.HasOrganization(r.FormValue("course")) {
		http.Error(w, "Unknown organization", 404)
		log.Println("Unknown organization")
		return
	}

	org, err := git.NewOrganization(r.FormValue("course"), true)
	if err != nil {
		http.Error(w, "Organization Error", 404)
		log.Println(err)
		return
	}

	var labs []*apProto.ApRequestLab
	var repos []string
	isGroup := false

	if strings.Contains(r.FormValue("labs"), "group") {
		// Get the information for groups
		isGroup = true

		length := len(org.GroupLabFolders)
		for i := 1; i <= length; i++ {
			if org.GroupLabFolders[i] != "" {
				labs = append(labs, &apProto.ApRequestLab{Name: org.GroupLabFolders[i],
					Language: org.GroupLanguages[i]})
			}
		}

		// The order of the repos does not matter.
		for groupName := range org.Groups {
			repos = append(repos, groupName)
		}
	} else {
		// Get the information for individuals
		isGroup = false

		length := len(org.IndividualLabFolders)
		for i := 1; i <= length; i++ {
			if org.IndividualLabFolders[i] != "" {
				labs = append(labs, &apProto.ApRequestLab{Name: org.IndividualLabFolders[i],
					Language: org.IndividualLanguages[i]})
			}
		}

		// The order of the repos does not matter.
		for indvName := range org.Members {
			repos = append(repos, indvName+"-labs")
		}
	}

	// Create request
	request := apProto.ApRequest{GithubOrg: org.Name,
		GithubToken:  org.AdminToken,
		StudentRepos: repos,
		Labs:         labs}

	go callAntiplagiarism(request, org, isGroup)

	fmt.Printf("%v\n", request)
}

// ApLabResultsHandler is a http handeler for getting results for one lab of a user
// from the latest anti-plagiarism test. This handler writes back the results as JSON data.
func ApLabResultsHandler(w http.ResponseWriter, r *http.Request) {
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

	var results *git.AntiPlagiarismResults

	if strings.HasPrefix(username, git.GroupRepoPrefix) {
		labIndex := -1
		// Find the correct lab index
		for i, name := range org.GroupLabFolders {
			if name == labname {
				labIndex = i
				break
			}
		}

		if labIndex < 0 {
			http.Error(w, "No lab with that name found.", 404)
			return
		}

		// Get the group ID from the group name
		groupid, err := strconv.Atoi(username[len(git.GroupRepoPrefix):])
		if err != nil {
			http.Error(w, "Could not convert the group ID.", 404)
			return
		}

		// Get the group from the database
		group, err := git.NewGroup(orgname, groupid, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		// Get the results for the lab
		results = group.GetAntiPlagiarismResults(org.Name, labIndex)
	} else {
		labIndex := -1
		// Find the correct lab index
		for i, name := range org.IndividualLabFolders {
			if name == labname {
				labIndex = i
				break
			}
		}

		if labIndex < 0 {
			http.Error(w, "No lab with that name found.", 404)
			return
		}

		// Get the user from the database
		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		// Get the results for the lab
		results = user.GetAntiPlagiarismResults(org.Name, labIndex)
	}

	enc := json.NewEncoder(w)

	err = enc.Encode(results)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
}

// ApUserResultsHandler is a http handeler for getting all results for a user
// from the latest anti-plagiarism test. This handler writes back the results as JSON data.
func ApUserResultsHandler(w http.ResponseWriter, r *http.Request) {
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

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !org.IsMember(member) {
		http.Error(w, "Not a member for this course.", 404)
		return
	}

	results := make(map[string]git.AntiPlagiarismResults)

	if strings.HasPrefix(username, git.GroupRepoPrefix) {
		// Get the group ID from the group name
		groupid, err := strconv.Atoi(username[len(git.GroupRepoPrefix):])
		if err != nil {
			http.Error(w, "Could not convert the group ID.", 404)
			return
		}

		// Get the group from the database
		group, err := git.NewGroup(orgname, groupid, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		// For each lab
		for i, name := range org.GroupLabFolders {
			// Get the results for the lab
			temp := group.GetAntiPlagiarismResults(org.Name, i)
			results[name] = *temp
		}
	} else {
		// Get user from the database
		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		// For each lab
		for i, name := range org.IndividualLabFolders {
			// Get the results for the lab
			temp := user.GetAntiPlagiarismResults(org.Name, i)
			results[name] = *temp
		}
	}

	enc := json.NewEncoder(w)

	// Encode the results in JSON
	err = enc.Encode(results)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
}
