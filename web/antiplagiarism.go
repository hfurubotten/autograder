package web

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	pb "github.com/autograde/antiplagiarism/proto"
	git "github.com/hfurubotten/autograder/entities"

	//"golang.org/x/net/context"
	//"google.golang.org/grpc"
)

// TODO: CHANGE THIS
var resultsBaseDir = "/home/ericfree/results"

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
	request := pb.ApRequest{GithubOrg: org.Name,
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
	/*endpoint := "localhost:11111"
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
	}*/

	checkResults(org)
	clearPreviousResults(org, isGroup)
	saveNewResults(org, isGroup)
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
func saveNewResults(org *git.Organization, isGroup bool) {
	// Make a slice of out tools
	tools := []string{"dupl", "jplag", "moss"}

	if isGroup {
		// For each lab
		for labIndex := 0; labIndex < org.GroupAssignments; labIndex++ {
			lab := org.GroupLabFolders[labIndex]
		
			// For each tool
			for _, tool := range tools {
				resultsDir := filepath.Join(resultsBaseDir, org.Name, lab, tool)
				resultsFile := filepath.Join(resultsDir, "percentage.txt")
			
				// Get the tool's results for that lab
				success := getFileResults(resultsFile, labIndex, tool, org, isGroup)
				if success {
					// Delete this file.
				}
			}
		}
	} else {
		// For each lab
		for labIndex := 0; labIndex < org.IndividualAssignments; labIndex++ {
			lab := org.IndividualLabFolders[labIndex]
		
			// For each tool
			for _, tool := range tools {
				resultsDir := filepath.Join(resultsBaseDir, org.Name, lab, tool)
				resultsFile := filepath.Join(resultsDir, "percentage.txt")
			
				// Get the tool's results for that lab
				success := getFileResults(resultsFile, labIndex, tool, org, isGroup)
				if success {
					// Delete this file.
				}
			}
		}
	}
}

func getFileResults(resultsFile string, labIndex int, tool string, org *git.Organization, isGroup bool) bool {
	file, err := os.Open(resultsFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		username := ""	
		tempString := scanner.Text()
		values := strings.Split(tempString, "|")
		
		// Format each value from that line
		if !isGroup {
			length := len(values[0])
			username = values[0][:length - 5]
		} else {
			username = values[0]
		}
		percent64, err := strconv.ParseFloat(values[1], 32)
		if err != nil {
			fmt.Printf("Error converting %s to a float while reading file %s.\n", values[1], resultsFile)
			continue
		}
		percent32 := float32(percent64)
		url := values[2]
		
		if isGroup {
			// Get the Group ID
			groupId, err := strconv.Atoi(username[len(git.GroupRepoPrefix):])
			if err != nil {
				fmt.Printf("Could not get group number from %s. %s\n", username, err)
				continue
			}
		
			// Get the database record
			group, _ := git.NewGroup(org.Name, groupId, false)
			
			// Update the results
			results := group.GetAntiPlagiarismResults(org.Name, labIndex)
			switch tool {
			case "dupl":
				results.DuplPct = percent32
				results.DuplUrl = url			
			case "jplag":
				results.JplagPct = percent32
				results.JplagUrl = url
			case "moss":
				results.MossPct = percent32
				results.MossUrl = url
			}
			group.AddAntiPlagiarismResults(org.Name, labIndex, results)
				
			// Save the database record
			group.Save()
		} else {
			// Get the database record
			student, _ := git.NewMemberFromUsername(username, false)
			
			// Update the results
			results := student.GetAntiPlagiarismResults(org.Name, labIndex)
			switch tool {
			case "dupl":
				results.DuplPct = percent32
				results.DuplUrl = url			
			case "jplag":
				results.JplagPct = percent32
				results.JplagUrl = url
			case "moss":
				results.MossPct = percent32
				results.MossUrl = url
			}
			student.AddAntiPlagiarismResults(org.Name, labIndex, results)
				
			// Save the database record
			student.Save()
		}
	}
	
	return true
}
