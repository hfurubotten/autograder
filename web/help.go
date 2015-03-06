package web

import (
	"net/http"
	"strings"

	"github.com/hfurubotten/autograder/git"
)

// HelpView is the struct given to the html template compilator in HelpHandler.
type HelpView struct {
	Member *git.Member
}

// HelpURL is the URL used to call HelpHandler.
var HelpURL = "/help/"

// HelpHandler is a http handler used to serve the help pages.
func HelpHandler(w http.ResponseWriter, r *http.Request) {
	addr := strings.TrimPrefix(r.URL.String(), "/")
	addr = strings.TrimSuffix(addr, "/")
	if addr == "help" {
		addr = "help/index"
	}

	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, false)
	if err == nil {
		view := HelpView{
			Member: member,
		}

		execTemplate(addr+".html", w, view)
	} else {
		execTemplate(addr+".html", w, nil)
	}

}
