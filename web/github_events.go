package web

import (
	"net/http"
	"strings"
)

//TODO gorename these constants to be github agnostic
const (
	GITHUB_COMMMIT_COMMENT             string = "commit_comment"
	GITHUB_ISSUE_COMMENT               string = "issue_comment"
	GITHUB_ISSUES                      string = "issues"
	GITHUB_PUSH                        string = "push"
	GITHUB_PULL_REQUEST                string = "pull_request"
	GITHUB_PULL_REQUEST_REVIEW_COMMENT string = "pull_request_review_comment"
	GITHUB_STATUS                      string = "status"
	GITHUB_WIKI                        string = "gollum"
	GITHUB_PING                        string = "ping"
	GITHUB_REPO                        string = "repository"

	UNKOWN          int = -1
	COMMMIT_COMMENT int = iota
	ISSUE_COMMENT
	ISSUES
	PING
	PUSH
	PULL_REQUEST
	PULL_REQUEST_COMMENT
	STATUS
	WIKI
	REPO_CREATE
)

// Maps the github type string to a internal type int.
var GITHUB_PAYLOADS map[string]int = map[string]int{
	GITHUB_COMMMIT_COMMENT:             COMMMIT_COMMENT,
	GITHUB_ISSUE_COMMENT:               ISSUE_COMMENT,
	GITHUB_ISSUES:                      ISSUES,
	GITHUB_PUSH:                        PUSH,
	GITHUB_PULL_REQUEST:                PULL_REQUEST,
	GITHUB_PULL_REQUEST_REVIEW_COMMENT: PULL_REQUEST_COMMENT,
	GITHUB_STATUS:                      STATUS,
	GITHUB_WIKI:                        WIKI,
	GITHUB_PING:                        PING,
	GITHUB_REPO:                        REPO_CREATE,
}

// GetPayloadType finds the event type of a payload request.
func GetPayloadType(r *http.Request) int {
	agent := r.Header.Get("User-Agent")
	switch {
	case strings.Contains(agent, "GitHub-Hookshot"):
		eventstring := r.Header.Get("X-Github-Event")
		if event, ok := GITHUB_PAYLOADS[eventstring]; ok {
			return event
		} else {
			return UNKOWN
		}
	default:
		return UNKOWN
	}
}
