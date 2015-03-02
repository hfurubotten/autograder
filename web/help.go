package web

import (
	"net/http"
	"strings"

	"github.com/hfurubotten/autograder/git"
)

type helpview struct {
	Member *git.Member
}

var HelpURL string = "/help/"

func HelpHandler(w http.ResponseWriter, r *http.Request) {
	addr := strings.TrimPrefix(r.URL.String(), "/")
	addr = strings.TrimSuffix(addr, "/")
	if addr == "help" {
		addr = "help/index"
	}

	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, false)
	if err == nil {
		view := helpview{
			Member: member,
		}

		execTemplate(addr+".html", w, view)
	} else {
		execTemplate(addr+".html", w, nil)
	}

}
