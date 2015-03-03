package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hfurubotten/autograder/git"
)

// ScoreboardView is the stuct used to pass data to the html template compiler.
type ScoreboardView struct {
	Member *git.Member
	Org    *git.Organization
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
		return
	}

	if !org.IsMember(member) {
		http.Redirect(w, r, HomeURL, 307)
		return
	}

	view := ScoreboardView{
		Member: member,
		Org:    org,
	}

	execTemplate("scoreboard.html", w, view)
}

type LeaderboardDataView struct {
	JSONErrorMsg
	Scores      map[string]int64
	Leaderboard []string
}

var LeaderboardDataURL string = "/leaderboard"

const (
	TOTALSCORE int = iota
	MONTHLYSCORE
	WEEKLYSCORE
)

func LeaderboardDataHandler(w http.ResponseWriter, r *http.Request) {
	view := LeaderboardDataView{}
	view.Error = true
	enc := json.NewEncoder(w)

	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		enc.Encode(ErrAccessToken)
		return
	}

	orgname := r.FormValue("course")
	period, err := strconv.Atoi(r.FormValue("period"))
	if err != nil {
		enc.Encode(ErrMissingField)
		return
	}

	if !git.HasOrganization(orgname) {
		enc.Encode(ErrUnknownCourse)
		return
	}

	org, err := git.NewOrganization(orgname)
	if err != nil {
		view.ErrorMsg = err.Error()
		enc.Encode(view)
		return
	}

	if !org.IsMember(member) {
		enc.Encode(ErrNotMember)
		return
	}

	var t time.Time
	if period == TOTALSCORE {
		view.Error = false
		view.Leaderboard = org.GetTotalLeaderboard()
		view.Scores = org.TotalScore
	} else if period == MONTHLYSCORE {
		t, err = time.Parse("1", r.FormValue("month"))
		if err != nil {
			t = time.Now()
		}
		month := t.Month()

		view.Error = false
		view.Leaderboard = org.GetMonthlyLeaderboard(month)
		view.Scores = org.MonthlyScore[month]
	} else if period == WEEKLYSCORE {
		week, err := strconv.Atoi(r.FormValue("week"))
		if err != nil {
			_, week = time.Now().ISOWeek()
		}

		if week < 1 || week > 53 {
			view.ErrorMsg = "Week need to be between 1 and 53."
			enc.Encode(view)
			return
		}

		view.Error = false
		view.Leaderboard = org.GetWeeklyLeaderboard(week)
		view.Scores = org.WeeklyScore[week]
	}

	enc.Encode(view)
}

var UserScoreDataURL string = "/score"

func UserScoreDataHandler(rw http.ResponseWriter, req *http.Request) {

}
