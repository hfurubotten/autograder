package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetEventType(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header != nil {
		req.Header["User-Agent"] = []string{"GitHub-Hookshot"}
	}
	values := []struct {
		inType  string
		outType int
	}{
		{GITHUB_COMMMIT_COMMENT, COMMMIT_COMMENT},
		{GITHUB_ISSUE_COMMENT, ISSUE_COMMENT},
		{GITHUB_ISSUES, ISSUES},
		{GITHUB_PING, PING},
		{GITHUB_PUSH, PUSH},
		{GITHUB_PULL_REQUEST, PULL_REQUEST},
		{GITHUB_PULL_REQUEST_REVIEW_COMMENT, PULL_REQUEST_COMMENT},
		{GITHUB_STATUS, STATUS},
	}
	for _, v := range values {
		req.Header["X-Github-Event"] = []string{v.inType}
		eventType := GetPayloadType(req)
		fmt.Println(eventType)
		if eventType != v.outType {
			t.Errorf("Expected event type %d, got %d", v.outType, eventType)
		}
	}
}
