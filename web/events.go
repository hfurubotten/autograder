package web

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	ci "github.com/hfurubotten/autograder/ci"
	git "github.com/hfurubotten/autograder/entities"
	github "github.com/hfurubotten/autograder/game/githubobjects"
	"github.com/hfurubotten/autograder/game/points"
	"github.com/hfurubotten/autograder/game/trophies"
)

var (
	// ScoreDistributionErrorMsg is the error message sent back to github when the scores wont update.
	ScoreDistributionErrorMsg = "Could not distribute the scores, try to resend the payload."

	// RegisterActionErrorMsg is the error message sent back to github when the trophy action wont update.
	RegisterActionErrorMsg = "Could not register the action made."

	// TeacherActionMsg is the error message sent back to github when the action comes from a teacher.
	TeacherActionMsg = "Event triggered by teacher. No action done in autograder."

	// DecodeGithubPayloadErrorMsg is the error message sent back to github when a issue payload wont decode.
	DecodeGithubPayloadErrorMsg = "Issue not decoded correctly."
)

// WebhookEventURL is the URL used to call WebhookEventHandler
var WebhookEventURL = "/event/hook"

// WebhookEventHandler is a http handler used to recieve webhooks
// from github. Upon recieving a payload it will find out if there
// is a action done on github or a push that triggered the webhook.
// On push a build will be done.
// On Github actions the user will be rawarded points.
func WebhookEventHandler(w http.ResponseWriter, r *http.Request) {
	defer PanicHandler(true)

	event := GetPayloadType(r)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Bytes received could not be decoded."))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var statusCode = http.StatusServiceUnavailable
	var body = "Wow, you actually got to see this msg. That shouldn't have happened."

	switch event {
	case COMMMIT_COMMENT:
		body, statusCode = handleCommitComment(b)

	case ISSUE_COMMENT:
		body = "Comment rewarded."
		statusCode = http.StatusOK

		payload, err := github.UnmarshalIssueComment(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statusCode = http.StatusInternalServerError
			break
		}

		user, err := git.GetMember(*payload.Comment.User.Login)
		if err != nil {
			log.Println("Error in member lookup: ", err)
			body = "Unknown GitHub User"
			statusCode = http.StatusInternalServerError
			break
		}
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(points.COMMENT)
		} else {
			err = DistributeScores(points.COMMENT, user, org)
			if err != nil {
				statusCode = http.StatusInternalServerError
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = RegisterAction(trophies.TALKACTION, user)
		if err != nil {
			statusCode = http.StatusInternalServerError
			body = RegisterActionErrorMsg
		}

	case ISSUES:
		body = "Issue action rewarded."
		statusCode = http.StatusOK

		payload, err := github.UnmarshalIssues(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statusCode = http.StatusInternalServerError
			break
		}

		user, err := git.GetMember(*payload.Sender.Login)
		if err != nil {
			log.Println("Error in member lookup: ", err)
			body = "Unknown GitHub User"
			statusCode = http.StatusInternalServerError
			break
		}
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		p, ta, err := findIssuesPointsAndTrophyAction(payload)
		if err != nil {
			log.Println("Issue event error:", err)
			statusCode = http.StatusInternalServerError
			body = "Could not calculate what score to give for the event."
			break
		}

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(p)
		} else {
			err = DistributeScores(p, user, org)
			if err != nil {
				statusCode = http.StatusInternalServerError
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = RegisterAction(ta, user)
		if err != nil {
			statusCode = http.StatusInternalServerError
			body = RegisterActionErrorMsg
		}

	case PING:
		body = "Pong"
		statusCode = http.StatusOK

	case PUSH:
		body = "Test build started"
		statusCode = http.StatusOK

		payload, err := github.UnmarshalPush(b)
		if err != nil {
			log.Println("Error decoding Push payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statusCode = http.StatusInternalServerError
			break
		}

		err = StartTestBuildProcess(payload)
		if err != nil {
			log.Println("Error starting test: ", err)

			//TODO: what? no error handling?
		}

	case PULL_REQUEST:
		body = "Be patient, users will soon get points for push requests also."
		statusCode = http.StatusNotImplemented

	case PULL_REQUEST_COMMENT:
		body = "Comment on push request rewarded."
		statusCode = http.StatusOK

		payload, err := github.UnmarshalPullRequestComments(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statusCode = http.StatusInternalServerError
			break
		}

		user, err := git.GetMember(*payload.Comment.User.Login)
		if err != nil {
			log.Println("Error in member lookup: ", err)
			body = "Unknown GitHub User"
			statusCode = http.StatusInternalServerError
			break
		}
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(points.COMMENT)
		} else {
			err = DistributeScores(points.COMMENT, user, org)
			if err != nil {
				statusCode = http.StatusInternalServerError
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = RegisterAction(trophies.TALKACTION, user)
		if err != nil {
			statusCode = http.StatusInternalServerError
			body = RegisterActionErrorMsg
		}

	case STATUS:
		body = "Be patient, Status update is soon also processed."
		statusCode = http.StatusNotImplemented

	case WIKI:
		body = "Be patient, wiki updates will be rewarded in time."
		statusCode = http.StatusNotImplemented

	case REPO_CREATE:
		body = "New Repos will be added at some point, but not at this time."
		statusCode = http.StatusNotImplemented

	default:
		body = "Unknown payload, thus not processed."
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(body))
}

// StartTestBuildProcess will use the payload from github to start the ci build.
func StartTestBuildProcess(load github.PushPayload) (err error) {
	userLogin := *load.Pusher.Name
	repoName := *load.Repo.Name
	orgName := *load.Organization.Login

	if !git.HasMember(userLogin) {
		return errors.New("invalid user login: " + userLogin)
	}
	if !git.HasOrganization(orgName) {
		return errors.New("invalid organization name: " + orgName)
	}
	org, err := git.NewOrganization(orgName, true)
	user, err := git.GetMember(userLogin)
	//TODO these erros will be returned at the end. Is that intentional??
	// Shouldn't they be handled here?

	var labfolder string
	var destfolder string
	var labnum int
	var username string
	var groupName = ""

	// TODO: Clean up this logic:
	// Make func: isGroupRepository()
	// Can this function instead check the number of student members of the repo
	// to determine if it's a group repo? Then we can avoid complicated logic.
	// isgroup := !strings.Contains(repoName, "-"+git.StandardRepoName)
	// if isgroup {
	if strings.HasPrefix(repoName, git.GroupRepoPrefix) {
		groupName = strings.Split(repoName, "-")[0]
		group, err := git.GetGroup(groupName)
		if err != nil {
			return err
		}

		labnum = group.CurrentLabNum
		if labnum > org.GroupAssignments {
			labnum = org.GroupAssignments
		}
		labfolder = org.GroupLabFolders[labnum]
		username = repoName
		destfolder = git.GroupsRepoName
	} else {
		labnum = user.Courses[org.Name].CurrentLabNum
		if labnum > org.IndividualAssignments {
			labnum = org.IndividualAssignments
		}
		labfolder = org.IndividualLabFolders[labnum]
		username = strings.TrimRight(repoName, "-"+git.StandardRepoName)
		destfolder = git.StandardRepoName
	}

	opt := ci.DaemonOptions{
		Org:        org.Name,
		User:       username,
		GroupName:  groupName,
		UserRepo:   repoName,
		TestRepo:   git.TestRepoName,
		BaseFolder: org.CI.Basepath,
		LabFolder:  labfolder,
		LabNumber:  labnum,
		AdminToken: org.AdminToken,
		DestFolder: destfolder,
		IsPush:     true,
		Secret:     org.CI.Secret,
	}

	go ci.StartTesterDaemon(opt)

	return //TODO This will return err; is that intentional??
}
