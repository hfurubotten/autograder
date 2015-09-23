package web

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	ci "github.com/hfurubotten/autograder/ci"
	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/game/events"
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
	defer events.PanicHandler(true)

	event := events.GetPayloadType(r)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Bytes received could not be decoded."))
		w.WriteHeader(503)
		return
	}

	var statuscode = 503
	var body = "Wow, you actually got to see this msg. That shouldn't have happened."

	switch event {
	case events.COMMMIT_COMMENT:
		body = "Comment rewarded."
		statuscode = 200

		payload, err := github.UnmarshalCommitComment(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statuscode = 500
			break
		}

		user, _ := git.NewUserWithGithubData(payload.Comment.User, true)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(points.COMMENT)
		} else {
			err = events.DistributeScores(points.COMMENT, user, nil, org)
			if err != nil {
				statuscode = 500
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = events.RegisterAction(trophies.TALKACTION, user)
		if err != nil {
			statuscode = 500
			body = RegisterActionErrorMsg
		}

	case events.ISSUE_COMMENT:
		body = "Comment rewarded."
		statuscode = 200

		payload, err := github.UnmarshalIssueComment(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statuscode = 500
			break
		}

		user, _ := git.NewUserWithGithubData(payload.Comment.User, true)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(points.COMMENT)
		} else {
			err = events.DistributeScores(points.COMMENT, user, nil, org)
			if err != nil {
				statuscode = 500
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = events.RegisterAction(trophies.TALKACTION, user)
		if err != nil {
			statuscode = 500
			body = RegisterActionErrorMsg
		}

	case events.ISSUES:
		body = "Issue action rewarded."
		statuscode = 200

		payload, err := github.UnmarshalIssues(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statuscode = 500
			break
		}

		user, _ := git.NewUserWithGithubData(payload.Sender, true)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		p, ta, err := events.FindIssuesPointsAndTrophyAction(payload)
		if err != nil {
			log.Println("Issue event error:", err)
			statuscode = 500
			body = "Could not calculate what score to give for the event."
			break
		}

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(p)
		} else {
			err = events.DistributeScores(p, user, nil, org)
			if err != nil {
				statuscode = 500
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = events.RegisterAction(ta, user)
		if err != nil {
			statuscode = 500
			body = RegisterActionErrorMsg
		}

	case events.PING:
		body = "Pong"
		statuscode = 200

	case events.PUSH:
		// go events.HandlePush(b)
		body = "Test build started"
		statuscode = 200

		payload, err := github.UnmarshalPush(b)
		if err != nil {
			log.Println("Error decoding Push payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statuscode = 500
			break
		}

		err = StartTestBuildProcess(payload)
		if err != nil {

		}

	case events.PULL_REQUEST:
		// go events.HandlePullRequest(b)
		body = "Be patient, users will soon get points for push requests also."
		statuscode = 501

	case events.PULL_REQUEST_COMMENT:
		// go events.HandlePullRequestComments(b)
		body = "Comment on push request rewarded."
		statuscode = 200

		payload, err := github.UnmarshalPullRequestComments(b)
		if err != nil {
			log.Println("Error decoding Commit Comment payload:", err)
			body = DecodeGithubPayloadErrorMsg
			statuscode = 500
			break
		}

		user, _ := git.NewUserWithGithubData(payload.Comment.User, true)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization, true)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			user.IncScoreBy(points.COMMENT)
		} else {
			err = events.DistributeScores(points.COMMENT, user, nil, org)
			if err != nil {
				statuscode = 500
				body = ScoreDistributionErrorMsg
				break
			}
		}
		err = events.RegisterAction(trophies.TALKACTION, user)
		if err != nil {
			statuscode = 500
			body = RegisterActionErrorMsg
		}

	case events.STATUS:
		// go events.HandleStatusUpdate(b)
		body = "Be patient, Status update is soon also processed."
		statuscode = 501

	case events.WIKI:
		// go events.HandleWikiUpdate(b)
		body = "Be patient, wiki updates will be rewarded in time."
		statuscode = 501

	case events.REPO_CREATE:
		// go events.HandleNewRepo(b)
		body = "New Repos will be added at some point, but not at this time."
		statuscode = 501

	default:
		body = "Unknown payload, thus not processed."
		statuscode = 503
	}

	w.WriteHeader(statuscode)
	w.Write([]byte(body))
}

// StartTestBuildProcess will use the payload from github to start the ci build.
func StartTestBuildProcess(load github.PushPayload) (err error) {
	userlogin := *load.Pusher.Name
	reponame := *load.Repo.Name
	orgname := *load.Organization.Login

	if !git.HasMember(userlogin) {
		log.Println("Not a valid user: ", userlogin)
		return
	}

	if !git.HasOrganization(orgname) {
		log.Println("Not a valid org: ", orgname)
		return errors.New("Not a valid org: " + orgname)
	}

	org, err := git.NewOrganization(orgname, true)
	user, err := git.NewMemberFromUsername(userlogin, true)

	isgroup := !strings.Contains(reponame, "-"+git.StandardRepoName)

	var labfolder string
	var destfolder string
	var labnum int
	var username string
	var gnum = -1
	if isgroup {
		gnum, err = strconv.Atoi(reponame[len("group"):])
		if err != nil {
			log.Println(err)
			return err
		}

		group, err := git.NewGroup(org.Name, gnum, true)
		if err != nil {
			log.Println(err)
			return err
		}

		labnum = group.CurrentLabNum
		if labnum > org.GroupAssignments {
			labnum = org.GroupAssignments
		}
		labfolder = org.GroupLabFolders[labnum]
		username = reponame
		destfolder = git.GroupsRepoName
	} else {
		labnum = user.Courses[org.Name].CurrentLabNum
		if labnum > org.IndividualAssignments {
			labnum = org.IndividualAssignments
		}
		labfolder = org.IndividualLabFolders[labnum]
		username = strings.TrimRight(reponame, "-"+git.StandardRepoName)
		destfolder = git.StandardRepoName
	}

	opt := ci.DaemonOptions{
		Org:        org.Name,
		User:       username,
		Group:      gnum,
		Repo:       reponame,
		BaseFolder: org.CI.Basepath,
		LabFolder:  labfolder,
		LabNumber:  labnum,
		AdminToken: org.AdminToken,
		DestFolder: destfolder,
		IsPush:     true,
		Secret:     org.CI.Secret,
	}

	go ci.StartTesterDaemon(opt)

	return
}
