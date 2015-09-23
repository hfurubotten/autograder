package githubobjects

import (
	"encoding/json"
)

func UnmarshalPullRequestComments(b []byte) (payload PullRequestCommentPayload, err error) {
	err = json.Unmarshal(b, &payload)
	return
}

func UnmarshalIssueComment(b []byte) (payload IssueCommentPayload, err error) {
	err = json.Unmarshal(b, &payload)
	return
}

func UnmarshalCommitComment(b []byte) (payload CommitCommentPayload, err error) {
	err = json.Unmarshal(b, &payload)
	return
}

func UnmarshalIssues(b []byte) (payload IssuesPayload, err error) {
	err = json.Unmarshal(b, &payload)
	return
}

func UnmarshalPush(b []byte) (payload PushPayload, err error) {
	err = json.Unmarshal(b, &payload)
	return
}
