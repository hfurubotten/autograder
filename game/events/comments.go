package events

import (
	"log"

	git "github.com/hfurubotten/autograder/entities"
	. "github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
)

func HandlePullRequestComments(b []byte) {
	defer PanicHandler(true)
	payload, err := UnmarshalPullRequestComments(b)
	if err != nil {
		log.Println("Error decoding Commit Comment payload:", err)
		return
	}

	gu := payload.Comment.User
	gr := payload.Repo
	o := payload.Organization

	user, _ := git.NewUserWithGithubData(gu)
	repo, _ := git.GetRepo(gr)
	org, _ := git.NewOrganizationWithGithubDataX(o)

	err = DistributeScores(points.COMMENT, user, repo, org)
	if err != nil {
		log.Println("Error distributing scores:", err)
	}
	RegisterAction(PULL_REQUEST_COMMENT, user)
}

func HandleIssueComment(b []byte) {
	defer PanicHandler(true)
	payload, err := UnmarshalIssueComment(b)
	if err != nil {
		log.Println("Error decoding Commit Comment payload:", err)
		return
	}

	gu := payload.Comment.User
	gr := payload.Repo
	o := payload.Organization

	user, _ := git.NewUserWithGithubData(gu)
	repo, _ := git.GetRepo(gr)
	org, _ := git.NewOrganizationWithGithubDataX(o)

	err = DistributeScores(points.COMMENT, user, repo, org)
	if err != nil {
		log.Println("Error distributing scores:", err)
	}
	RegisterAction(ISSUE_COMMENT, user)
}

func HandleCommitComment(b []byte) {
	defer PanicHandler(true)
	payload, err := UnmarshalCommitComment(b)
	if err != nil {
		log.Println("Error decoding Commit Comment payload:", err)
		return
	}

	gu := payload.Comment.User
	gr := payload.Repo
	o := payload.Organization

	user, _ := git.NewUserWithGithubData(gu)
	repo, _ := git.GetRepo(gr)
	org, _ := git.NewOrganizationWithGithubDataX(o)

	err = DistributeScores(points.COMMENT, user, repo, org)
	if err != nil {
		log.Println("Error distributing scores:", err)
	}
	RegisterAction(COMMMIT_COMMENT, user)
}
