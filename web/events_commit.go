package web

import (
	"log"
	"net/http"

	"github.com/hfurubotten/autograder/entities"
	github "github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
	"github.com/hfurubotten/autograder/game/trophies"
)

func handleCommitComment(b []byte) (body string, statusCode int) {
	body = "Comment rewarded."
	statusCode = http.StatusOK

	payload, err := github.UnmarshalCommitComment(b)
	if err != nil {
		log.Println("Error decoding Commit Comment payload:", err)
		body = DecodeGithubPayloadErrorMsg
		statusCode = http.StatusInternalServerError
		return
	}

	user, err := entities.GetMember(*payload.Comment.User.Login)
	if err != nil {
		log.Println("Error in member lookup: ", err)
		body = "Unknown GitHub User"
		statusCode = http.StatusInternalServerError
		return
	}

	org, _ := entities.NewOrganizationWithGithubData(payload.Organization, true)

	if org.IsTeacher(user) {
		body = TeacherActionMsg
		user.IncScoreBy(points.COMMENT)
	} else {
		err = DistributeScores(points.COMMENT, user, org)
		if err != nil {
			statusCode = http.StatusInternalServerError
			body = ScoreDistributionErrorMsg
			return
		}
	}
	err = RegisterAction(trophies.TALKACTION, user)
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = RegisterActionErrorMsg
	}
	return
}
