package git

const (
	StandardRepoName   string = "labs"
	GroupsRepoName     string = "glabs"
	CourseInfoName     string = "course-info"
	TestRepoName       string = "labs-test"
	GrouptestRepoName  string = "groups-test" // Deprecated. Use only the TEST_REPO_NAME.
	CodeReviewRepoName string = "code-reviews"

	AdminPermission string = "admin"
	PullPermission  string = "pull"
	PushPermission  string = "push"

	IndividualType int = iota
	GroupType
)
