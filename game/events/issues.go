package events

import (
	"errors"
	"log"

	git "github.com/hfurubotten/autograder/entities"
	. "github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
	"github.com/hfurubotten/autograder/game/trophies"
)

func FindIssuesPointsAndTrophyAction(payload IssuesPayload) (int, int, error) {
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

func HandleIssues(b []byte) {
	defer PanicHandler(true)
	payload, err := UnmarshalIssues(b)
	if err != nil {
		log.Println("Error decoding Commit Comment payload:", err)
		return
	}

	gu := payload.Sender
	o := payload.Organization

	p, ta, err := FindIssuesPointsAndTrophyAction(payload)
	if err != nil {
		log.Println("Issues payload error:", err)
		return
	}

	user, _ := git.GetMemberX(gu)
	org, _ := git.NewOrganizationWithGithubDataX(o)

	err = DistributeScores(p, user, org)
	if err != nil {
		log.Println("Error distributing scores:", err)
	}
	err = RegisterAction(ta, user)
	if err != nil {
		log.Println("Error registrating action:", err)
	}
}
