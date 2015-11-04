package web

import (
	"fmt"
	"net/http"
	"strings"

	git "github.com/hfurubotten/autograder/entities"
)

// ManualTestPlagiarismURL is the URL used to call ManualTestPlagiarismHandler.
var ManualTestPlagiarismURL = "/event/manualtestplagiarism"

// ManualTestPlagiarismHandler is a http handler for manually triggering test builds.
func ManualTestPlagiarismHandler(w http.ResponseWriter, r *http.Request) {
	if !git.HasOrganization(r.FormValue("course")) {
		http.Error(w, "Unknown organization", 404)
		return
	}

	org, err := git.NewOrganization(r.FormValue("course"), true)
	if err != nil {
		http.Error(w, "Organization Error", 404)
		return
	}

	var labs []string
	var languages []int
	var repos []string

	if strings.Contains(r.FormValue("labs"), "group") {
		// Get the information for groups

		// Order of labs and languages matters. They must match.
		length := len(org.GroupLabFolders)
		for i := 1; i <= length; i++ {
			if org.GroupLabFolders[i] != "" {
				labs = append(labs, org.GroupLabFolders[i])
				languages = append(languages, org.GroupLanguages[i])
			}
		}

		// The order of the repos does not matter.
		for groupName, _ := range org.Groups {
			repos = append(repos, groupName)
		}
	} else {
		// Get the information for individuals

		// Order of labs and languages matters. They must match.
		length := len(org.IndividualLabFolders)
		for i := 1; i <= length; i++ {
			if org.IndividualLabFolders[i] != "" {
				labs = append(labs, org.IndividualLabFolders[i])
				languages = append(languages, org.IndividualLanguages[i])
			}
		}

		// The order of the repos does not matter.
		for indvName, _ := range org.Members {
			repos = append(repos, indvName + "-labs")
		}
	}

	fmt.Printf("Org:       %v\n", r.FormValue("course"))
	fmt.Printf("Token:     %v\n", org.AdminToken)
	fmt.Printf("Labs:      %v\n", labs)
	fmt.Printf("Languages: %v\n", languages)
	fmt.Printf("Repos:     %v\n", repos)
}
