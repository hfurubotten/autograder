package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	apCommon "github.com/autograde/antiplagiarism/common"
	apProto "github.com/autograde/antiplagiarism/proto"
	git "github.com/hfurubotten/autograder/entities"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// TODO: CHANGE THIS
var resultsBaseDir = "/home/ericfree/results"

// callAntiplagiarism sends a request to the anti-plagiarism software.
// It takes as input request, an ApRequest (anti-plagiarism request),
// org, a database record for the class, and isGroup, whether or not
// the request if for individual or group assignments.
func callAntiplagiarism(request apProto.ApRequest, org *git.Organization, isGroup bool) {
	// Currently just on localhost.
	endpoint := "localhost:11111"
	var opts []grpc.DialOption
	// TODO: Add transport security.
	opts = append(opts, grpc.WithInsecure())

	// Create connection
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		fmt.Printf("callAntiplagiarism: Error while connecting to server: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("Connected to server on %v\n", endpoint)

	// Create client
	client := apProto.NewApClient(conn)

	// Send request and get response
	response, err := client.CheckPlagiarism(context.Background(), &request)

	// Check response
	if err != nil {
		fmt.Printf("callAntiplagiarism: gRPC error: %s\n", err)
		return
	} else if response.Success == false {
		fmt.Printf("callAntiplagiarism: Anti-plagiarism error: %s\n", response.Err)
		return
	} else {
		fmt.Printf("Anti-plagiarism application ran successfully.\n")
	}

	clearPreviousResults(org, isGroup)
	saveNewResults(org, isGroup)
	showAllResults(org, isGroup)
}

// clearPreviousResults clears the previous anti-plagiarim results,
// because the specific urls can change. It takes as input
// org, a database record for the class, and isGroup, whether or not
// the request if for individual or group assignments.
func clearPreviousResults(org *git.Organization, isGroup bool) {
	if isGroup {
		// Clear old group results
		// For each group
		for groupName := range org.Groups {
			// Get the Group ID
			groupID, err := strconv.Atoi(groupName[len(git.GroupRepoPrefix):])
			if err != nil {
				fmt.Printf("clearPreviousResults: Could not get group number from %s. %s\n", groupName, err)
				continue
			}

			// Get the database record
			group, _ := git.NewGroup(org.Name, groupID, false)
			// For each lab
			for labIndex := 1; labIndex <= org.GroupAssignments; labIndex++ {
				// Clear the specific lab results
				results := git.AntiPlagiarismResults{MossPct: 0.0,
					MossURL:  "",
					DuplPct:  0.0,
					DuplURL:  "",
					JplagPct: 0.0,
					JplagURL: ""}
				group.AddAntiPlagiarismResults(org.Name, labIndex, &results)
			}
			// Save the database record
			group.Save()
		}
	} else {
		// Clear old individual results
		// For each student
		for username := range org.Members {
			// Get the database record
			student, _ := git.NewMemberFromUsername(username, false)
			// For each lab
			for labIndex := 1; labIndex <= org.IndividualAssignments; labIndex++ {
				// Clear the specific lab results
				results := git.AntiPlagiarismResults{MossPct: 0,
					MossURL:  "",
					DuplPct:  0,
					DuplURL:  "",
					JplagPct: 0,
					JplagURL: ""}
				student.AddAntiPlagiarismResults(org.Name, labIndex, &results)
			}
			// Save the database record
			student.Save()
		}
	}

}

// saveNewResults saves the results in the results directory to the database.
// It takes as input org, a database record for the class, and isGroup,
// whether or not the request is for individual or group assignments.
func saveNewResults(org *git.Organization, isGroup bool) {
	// Make a slice of out tools
	tools := []string{"dupl", "jplag", "moss"}

	if isGroup {
		// For each lab
		for labIndex := 1; labIndex <= org.GroupAssignments; labIndex++ {
			lab := org.GroupLabFolders[labIndex]

			// For each tool
			for _, tool := range tools {
				resultsDir := filepath.Join(resultsBaseDir, org.Name, lab, tool)
				resultsFile := filepath.Join(resultsDir, apCommon.ResultsFileName)

				// Get the tool's results for that lab
				success := getFileResults(resultsFile, labIndex, tool, org, isGroup)
				if success {
					// TODO: Delete this file.
				}
			}
		}
	} else {
		// For each lab
		for labIndex := 1; labIndex <= org.IndividualAssignments; labIndex++ {
			lab := org.IndividualLabFolders[labIndex]

			// For each tool
			for _, tool := range tools {
				resultsDir := filepath.Join(resultsBaseDir, org.Name, lab, tool)
				resultsFile := filepath.Join(resultsDir, apCommon.ResultsFileName)

				// Get the tool's results for that lab
				success := getFileResults(resultsFile, labIndex, tool, org, isGroup)
				if success {
					// TODO: Delete this file.
				}
			}
		}
	}
}

// getFileResults will read a results file and store the data in the database.
// It returns a bool value indicating if the function was successful or not.
// It takes as input resultsFile, the name of the results file, labIndex,
// the index of the lab, tool, which antiplagiarism tool made the file,
// org, a database record for the class, and isGroup,
// whether or not the request is for individual or group assignments.
func getFileResults(resultsFile string, labIndex int, tool string, org *git.Organization, isGroup bool) bool {

	buf, err := ioutil.ReadFile(resultsFile)
	if err != nil {
		fmt.Printf("Error reading file. %s\n", err)
		return false
	}

	var fileResults apCommon.ResultEntries

	err = json.Unmarshal(buf, &fileResults)
	if err != nil {
		fmt.Printf("Error unmarshalling results from JSON format. File: %s. %s\n", resultsFile, err)
		return false
	}

	if fileResults == nil {
		return false
	}

	for _, fileResult := range fileResults {

		if isGroup {
			// Make sure that this is a group
			if !strings.HasPrefix(fileResult.Repo, "group") {
				fmt.Printf("JPlag might be returning matching individual labs from previous sessions.\n")
				fmt.Printf("If that is the case, there is a group with code matching an individuals code\n")
				fmt.Printf("from another lab.\n")
				continue
			}

			// Get the Group ID
			groupID, err := strconv.Atoi(fileResult.Repo[len(git.GroupRepoPrefix):])
			if err != nil {
				fmt.Printf("getFileResults: Could not get group number from %s. %s\n", fileResult.Repo, err)
				continue
			}

			// Get the database record
			group, _ := git.NewGroup(org.Name, groupID, false)

			// Update the results
			results := group.GetAntiPlagiarismResults(org.Name, labIndex)
			if results != nil {
				switch tool {
				case "dupl":
					results.DuplPct = fileResult.Percent
					results.DuplURL = fileResult.URL
				case "jplag":
					results.JplagPct = fileResult.Percent
					results.JplagURL = fileResult.URL
				case "moss":
					results.MossPct = fileResult.Percent
					results.MossURL = fileResult.URL
				}
				group.AddAntiPlagiarismResults(org.Name, labIndex, results)

				// Save the database record
				group.Save()
			}
		} else {
			// Make sure that this is a group
			if strings.HasPrefix(fileResult.Repo, "group") {
				fmt.Printf("JPlag might be returning matching group labs from previous sessions.\n")
				fmt.Printf("If that is the case, there is a group with code matching an individuals code\n")
				fmt.Printf("from another lab.\n")
				continue
			}

			length := len(fileResult.Repo)
			username := fileResult.Repo[:length-5]

			// Get the database record
			student, _ := git.NewMemberFromUsername(username, false)
			// Update the results
			results := student.GetAntiPlagiarismResults(org.Name, labIndex)
			if results != nil {
				switch tool {
				case "dupl":
					results.DuplPct = fileResult.Percent
					results.DuplURL = fileResult.URL
				case "jplag":
					results.JplagPct = fileResult.Percent
					results.JplagURL = fileResult.URL
				case "moss":
					results.MossPct = fileResult.Percent
					results.MossURL = fileResult.URL
				}
				student.AddAntiPlagiarismResults(org.Name, labIndex, results)

				// Save the database record
				student.Save()
			}
		}
	}

	return true
}

// TODO: FOR DEBUGGING ONLY
// showAllResults shows the database records after the results have been saves.
func showAllResults(org *git.Organization, isGroup bool) {
	if isGroup {
		// For each group
		for groupName := range org.Groups {
			// Get the Group ID
			groupID, err := strconv.Atoi(groupName[len(git.GroupRepoPrefix):])
			if err != nil {
				fmt.Printf("showAllResults: Could not get group number from %s. %s\n", groupName, err)
				continue
			}

			// Get the database record
			group, _ := git.NewGroup(org.Name, groupID, false)
			// For each lab
			for labIndex := 1; labIndex <= org.GroupAssignments; labIndex++ {
				results := group.GetAntiPlagiarismResults(org.Name, labIndex)
				fmt.Printf("%v\n", results)
			}
		}
	} else {
		// For each student
		for username := range org.Members {
			// Get the database record
			student, _ := git.NewMemberFromUsername(username, false)
			// For each lab
			for labIndex := 1; labIndex <= org.IndividualAssignments; labIndex++ {
				results := student.GetAntiPlagiarismResults(org.Name, labIndex)
				fmt.Printf("%v\n", results)
			}
		}
	}
}
