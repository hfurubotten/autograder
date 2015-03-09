package git

const (
	// Standard repository names
	StandardRepoName   = "labs"
	GroupsRepoName     = "glabs"
	CourseInfoName     = "course-info"
	TestRepoName       = "labs-test"
	GrouptestRepoName  = "groups-test" // Deprecated. Use only the TEST_REPO_NAME.
	CodeReviewRepoName = "code-reviews"
	GroupRepoPrefix    = "group"

	// Team premission names on github
	AdminPermission = "admin"
	PullPermission  = "pull"
	PushPermission  = "push"

	// Assignment types
	IndividualType int = iota
	GroupType
)
