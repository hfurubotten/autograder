package web

import (
	"net/http"
	"strings"
)

// HelpURL is the URL used to call HelpHandler.
var HelpURL = "/help/"

// HelpHandler is a http handler used to serve the help pages.
func HelpHandler(w http.ResponseWriter, r *http.Request) {
	addr := strings.TrimPrefix(r.URL.Path, "/")
	addr = strings.TrimSuffix(addr, "/")
	if addr == "help" {
		addr = "help/index"
	}

	var view stdTemplate
	// Checks if the user is signed in
	member, err := checkMemberApproval(w, r, false)
	if err == nil {
		view = stdTemplate{Member: member}
	}
	execTemplate(addr+".html", w, view)
}
