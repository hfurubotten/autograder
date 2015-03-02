package web

import (
	"net/http"
	"strings"

	"github.com/hfurubotten/autograder/git"
)

// ScoreboardView is the stuct used to pass data to the html template compiler.
type ScoreboardView struct {
	Member      *git.Member
	Org         *git.Organization
	Leaderboard []string
	Scores      map[string]int64
}

// ScoreboardURL is the URL used to call ScoreboardHandler.
var ScoreboardURL string = "/scoreboard/"

// ScoreboardHandler is a http handler to give the user a page
// showing the scoreboard for a course
func ScoreboardHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 3 {
		if !git.HasOrganization(path[2]) {
			http.Redirect(w, r, HomeURL, 307)
			return
		}

		orgname = path[2]
	} else {
		http.Redirect(w, r, HomeURL, 307)
		return
	}

	org, err := git.NewOrganization(orgname)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	lb := org.GetTotalLeaderboard()
	sb := make(map[string]int64)

	for _, u := range lb {
		sb[u] = org.GetUserScore(u)
	}

	view := ScoreboardView{
		Member:      member,
		Org:         org,
		Leaderboard: lb,
		Scores:      sb,
	}

	execTemplate("scoreboard.html", w, view)
}
