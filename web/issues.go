package web

import (
	"errors"

	"github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
	"github.com/hfurubotten/autograder/game/trophies"
)

func findIssuesPointsAndTrophyAction(payload githubobjects.IssuesPayload) (int, int, error) {
	if payload.Action == nil {
		return 0, 0, errors.New("Cant use empty Action on issues payload.")
	}

	var p int
	var ta int
	switch *payload.Action {
	case "assigned":
		p = points.ASSIGNMENT
		ta = trophies.ASSIGNACTION
	case "unassigned":
		p = points.UNASSIGNMENT
		ta = trophies.ASSIGNACTION
	case "labeled":
		p = points.LABEL
		ta = trophies.LABELACTION
	case "unlabeled":
		p = points.UNLABEL
		ta = trophies.LABELACTION
	case "opened":
		p = points.OPEN_ISSUE
		ta = trophies.ISSUEACTION
	case "closed":
		p = points.CLOSE_ISSUE
		ta = trophies.ISSUEACTION
	case "reopened":
		p = points.REOPEN_ISSUE
		ta = trophies.ISSUEACTION
	default:
		return 0, 0, errors.New("Issue action not known for " + *payload.Action)
	}

	return p, ta, nil
}
