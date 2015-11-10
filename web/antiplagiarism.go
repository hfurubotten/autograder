package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	pb "github.com/autograde/antiplagiarism/proto"
	git "github.com/hfurubotten/autograder/entities"

	//"golang.org/x/net/context"
	//"google.golang.org/grpc"
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
	var languages []int32
	var repos []string
	isGroup := false

	if strings.Contains(r.FormValue("labs"), "group") {
		// Get the information for groups
		isGroup = true

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
		isGroup = false

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
			repos = append(repos, indvName+"-labs")
		}
	}

	// Create request
	request := pb.ApRequest{GithubOrg: r.FormValue("course"),
		GithubToken:  org.AdminToken,
		StudentRepos: repos,
		LabNames:     labs,
		LabLanguages: languages}

	go callAntiplagiarism(request, org, isGroup)

	fmt.Printf("%v\n", request)
}

// callAntiplagiarism sends a request to the anti-plagiarism software.
// It takes as input request, an ApRequest (anti-plagiarism request),
// org, a database record for the class, and isGroup, whether or not
// the request if for individual or group assignments.
func callAntiplagiarism(request pb.ApRequest, org *git.Organization, isGroup bool) {
	// Currently just on localhost.
	endpoint := "localhost:11111"
	var opts []grpc.DialOption
	// TODO: Add transport security.
	opts = append(opts, grpc.WithInsecure())

	// Create connection
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		fmt.Printf("Error while connecting to server: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("Connected to server on %v\n", endpoint)

	// Create client
	client := pb.NewApClient(conn)

	// Send request and get response
	response, err := client.CheckPlagiarism(context.Background(), &request)

	// Check response
	if err != nil {
		fmt.Printf("gRPC error: %s\n", err)
		return
	} else if response.Success == false {
		fmt.Printf("Anti-plagiarism error: %s\n", response.Err)
		return
	} else {
		fmt.Printf("Anti-plagiarism application ran successfully.\n")
	}

	checkResults()
	clearPreviousResults(ord, isGroup)
	saveNewResults()
}

// checkResults checks that there are results in the results directory.
// It takes as input org, a database record for the class.
func checkResults(org *git.Organization) {

}

// clearPreviousResults clears the previous anti-plagiarim results,
// because the specific urls can change. It takes as input
// org, a database record for the class, and isGroup, whether or not
// the request if for individual or group assignments.
func clearPreviousResults(org *git.Organization, isGroup bool) {
	if isGroup {
		// Clear old group results
		// For each group
		for groupName, _ := range org.Groups {
			// Get the Group ID
			groupId, err := strconv.Atoi(groupName[len(git.GroupRepoPrefix):])
			if err != nil {
				fmt.Printf("Could not get group number from %s. %s\n", groupName, err)
				continue
			}

			// Get the database record
			group, _ := git.NewGroup(org.Name, groupId, false)
			// For each lab
			for labIndex := 0; labIndex < org.GroupAssignments; labIndex++ {
				// Clear the specific lab results
				results := git.AntiPlagiarismResults{MossPct: 0.0,
					MossUrl:  "",
					DuplPct:  0.0,
					DuplUrl:  "",
					JplagPct: 0.0,
					JplagUrl: ""}
				group.AddAntiPlagiarismResults(org.Name, labIndex, &results)
			}
			// Save the database record
			group.Save()
		}
	} else {
		// Clear old individual results
		// For each student
		for username, _ := range org.Members {
			// Get the database record
			student, _ := git.NewMemberFromUsername(username, false)
			// For each lab
			for labIndex := 0; labIndex < org.IndividualAssignments; labIndex++ {
				// Clear the specific lab results
				results := git.AntiPlagiarismResults{MossPct: 0,
					MossUrl:  "",
					DuplPct:  0,
					DuplUrl:  "",
					JplagPct: 0,
					JplagUrl: ""}
				student.AddAntiPlagiarismResults(org.Name, labIndex, &results)
			}
			// Save the database record
			student.Save()
		}
	}

}

// saveNewResults saves the results in the results directory to the database.
// It takes as input org, a database record for the class.
func saveNewResults(org *git.Organization) {

}
