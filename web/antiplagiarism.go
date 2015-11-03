package web

import (
	"fmt"
	"net/http"
)

// ManualTestPlagiarismURL is the URL used to call ManualTestPlagiarismHandler.
var ManualTestPlagiarismURL = "/event/manualtestplagiarism"

// ManualTestPlagiarismHandler is a http handler for manually triggering test builds.
func ManualTestPlagiarismHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Button push resulted in ManualTestPlagiarismHandler() call.\n")
	fmt.Printf("Course: %s, Labs: %s\n", r.FormValue("course"), r.FormValue("labs"))
}
