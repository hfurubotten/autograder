package web

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	ci "github.com/hfurubotten/autograder/ci"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/github-gamification/events"
	github "github.com/hfurubotten/github-gamification/githubobjects"
	"github.com/hfurubotten/github-gamification/points"
)

var (
	ScoreDistributionErrorMsg   string = "Could not distribute the scores, try to resend the payload."
	RegisterActionErrorMsg      string = "Could not register the action made."
	TeacherActionMsg            string = "Event triggered by teacher. No action done in autograder."
	DecodeGithubPayloadErrorMsg string = "Issue not decoded correctly."
)

func webhookeventhandler(w http.ResponseWriter, r *http.Request) {
	defer events.PanicHandler(true)

	event := events.GetPayloadType(r)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Bytes received could not be decoded."))
		w.WriteHeader(503)
		return
	}

	var statuscode int = 503
	var body string = "Wow, you actually got to see this msg. That shouldn't have happened."

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

		user, _ := git.NewUserWithGithubData(payload.Comment.User)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			break
		}

		err = events.DistributeScores(points.COMMENT, user, nil, org)
		if err != nil {
			statuscode = 500
			body = ScoreDistributionErrorMsg
			break
		}
		err = events.RegisterAction(events.COMMMIT_COMMENT, user)
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

		user, _ := git.NewUserWithGithubData(payload.Comment.User)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			break
		}

		err = events.DistributeScores(points.COMMENT, user, nil, org)
		if err != nil {
			statuscode = 500
			body = ScoreDistributionErrorMsg
			break
		}
		err = events.RegisterAction(events.ISSUE_COMMENT, user)
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

		user, _ := git.NewUserWithGithubData(payload.Sender)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			break
		}

		p, err := events.FindIssuesPoints(payload)
		if err != nil {
			log.Println("Issue event error:", err)
			statuscode = 500
			body = "Could not calculate what score to give for the event."
			break
		}

		err = events.DistributeScores(p, user, nil, org)
		if err != nil {
			statuscode = 500
			body = ScoreDistributionErrorMsg
			break
		}
		err = events.RegisterAction(events.ISSUES, user)
		if err != nil {
			statuscode = 500
			body = RegisterActionErrorMsg
		}

	case events.PING:
		// go events.HandlePing(b)
		body = "At some point this will add repos to the system."
		statuscode = 501

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

		user, _ := git.NewUserWithGithubData(payload.Comment.User)
		org, _ := git.NewOrganizationWithGithubData(payload.Organization)

		if org.IsTeacher(user) {
			body = TeacherActionMsg
			break
		}

		err = events.DistributeScores(points.COMMENT, user, nil, org)
		if err != nil {
			statuscode = 500
			body = ScoreDistributionErrorMsg
			break
		}
		err = events.RegisterAction(events.PULL_REQUEST_COMMENT, user)
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

	org, err := git.NewOrganization(orgname)
	user, err := git.NewMemberFromUsername(userlogin)

	isgroup := !strings.Contains(reponame, "-"+git.STANDARD_REPO_NAME)

	var labfolder string
	var destfolder string
	var labnum int
	var username string
	if isgroup {
		gnum, err := strconv.Atoi(reponame[len("group"):])
		if err != nil {
			log.Println(err)
			return err
		}

		group, err := git.NewGroup(org.Name, gnum)
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
		destfolder = git.GROUPS_REPO_NAME
	} else {
		labnum = user.Courses[org.Name].CurrentLabNum
		if labnum > org.IndividualAssignments {
			labnum = org.IndividualAssignments
		}
		labfolder = org.IndividualLabFolders[labnum]
		username = strings.TrimRight(reponame, "-"+git.STANDARD_REPO_NAME)
		destfolder = git.STANDARD_REPO_NAME
	}

	opt := ci.DaemonOptions{
		Org:        org.Name,
		User:       username,
		Repo:       reponame,
		BaseFolder: org.CI.Basepath,
		LabFolder:  labfolder,
		AdminToken: org.AdminToken,
		DestFolder: destfolder,
		IsPush:     true,
		Secret:     org.CI.Secret,
	}

	go ci.StartTesterDaemon(opt)

	return
}
