package githubobjects

import (
	"time"

	. "github.com/google/go-github/github"
)

type CommitCommentPayload struct {
	Comment      *Comment      `json:"comment,omitempty"`
	Repo         *Repository   `json:"repository,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
	Sender       *User         `json:"sender,omitempty"`
}

type Comment struct {
	ID        *int       `json:"id,omitempty"`
	User      *User      `json:"user,omitempty"`
	Position  *int       `json:"position,omitempty"`
	Path      *string    `json:"path,omitempty"`
	CommitID  *string    `json:"commit_id,omitempty"`
	Body      *string    `json:"body,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	URL            *string `json:"url,omitempty"`
	HTMLURL        *string `json:"html_url,omitempty"`
	IssueURL       *string `json:"issue_url,omitempty"`
	PullRequestURL *string `json:"pull_request_url,omitempty"`
}

type PullRequestCommentPayload struct {
	Action       *string       `json:"action,omitempty"`
	Comment      *Comment      `json:"comment,omitempty"`
	Repo         *Repository   `json:"repository,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
	Sender       *User         `json:"sender,omitempty"`
	PullRequest  *PullRequest  `json:"pull_request,omitempty"`
}

type IssuesPayload struct {
	Action       *string       `json:"action,omitempty"`
	Issue        *Issue        `json:"issue,omitempty"`
	Repo         *Repository   `json:"repository,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
	Sender       *User         `json:"sender,omitempty"`
	Label        *Label        `json:"label,omitempty"`
	Assignee     *User         `json:"assignee,omitempty"`
}

type IssueCommentPayload struct {
	IssueCommentEvent
	Organization *Organization `json:"organization,omitempty"`
}

//
type PushPayload struct {
	After        *string         `json:"after,omitempty"`
	Before       *string         `json:"before,omitempty"`
	Commits      []WebHookCommit `json:"commits,omitempty"`
	Compare      *string         `json:"compare,omitempty"`
	Created      *bool           `json:"created,omitempty"`
	Deleted      *bool           `json:"deleted,omitempty"`
	Forced       *bool           `json:"forced,omitempty"`
	HeadCommit   *WebHookCommit  `json:"head_commit,omitempty"`
	Pusher       *User           `json:"pusher,omitempty"`
	Ref          *string         `json:"ref,omitempty"`
	Repo         *PushRepository `json:"repository,omitempty"`
	Organization *Organization   `json:"organization,omitempty"`
}

// PushRepository is a minor rewrite of go-githubs
// implementation. This rewrite is due to inconsistency
// in GitHub API when it comes to webhook push data.
//
// DO NOT USE THIS STRUCT UP AGAINST go-github!
type PushRepository struct {
	ID               *int             `json:"id,omitempty"`
	Owner            *User            `json:"owner,omitempty"`
	Name             *string          `json:"name,omitempty"`
	FullName         *string          `json:"full_name,omitempty"`
	Description      *string          `json:"description,omitempty"`
	Homepage         *string          `json:"homepage,omitempty"`
	DefaultBranch    *string          `json:"default_branch,omitempty"`
	MasterBranch     *string          `json:"master_branch,omitempty"`
	CreatedAt        *Timestamp       `json:"created_at,omitempty"`
	PushedAt         *Timestamp       `json:"pushed_at,omitempty"`
	UpdatedAt        *Timestamp       `json:"updated_at,omitempty"`
	HTMLURL          *string          `json:"html_url,omitempty"`
	CloneURL         *string          `json:"clone_url,omitempty"`
	GitURL           *string          `json:"git_url,omitempty"`
	MirrorURL        *string          `json:"mirror_url,omitempty"`
	SSHURL           *string          `json:"ssh_url,omitempty"`
	SVNURL           *string          `json:"svn_url,omitempty"`
	Language         *string          `json:"language,omitempty"`
	Fork             *bool            `json:"fork"`
	ForksCount       *int             `json:"forks_count,omitempty"`
	NetworkCount     *int             `json:"network_count,omitempty"`
	OpenIssuesCount  *int             `json:"open_issues_count,omitempty"`
	StargazersCount  *int             `json:"stargazers_count,omitempty"`
	SubscribersCount *int             `json:"subscribers_count,omitempty"`
	WatchersCount    *int             `json:"watchers_count,omitempty"`
	Size             *int             `json:"size,omitempty"`
	AutoInit         *bool            `json:"auto_init,omitempty"`
	Parent           *Repository      `json:"parent,omitempty"`
	Source           *Repository      `json:"source,omitempty"`
	Organization     *Organization    `json:"organization,omitempty"`
	OrganizationName *string          `json:"organization,omitempty"` // Changed
	Permissions      *map[string]bool `json:"permissions,omitempty"`
}
