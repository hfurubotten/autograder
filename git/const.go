package git

const (
	STANDARD_REPO_NAME   string = "labs"
	GROUPS_REPO_NAME     string = "glabs"
	COURSE_INFO_NAME     string = "course-info"
	TEST_REPO_NAME       string = "labs-test"
	GROUPTEST_REPO_NAME  string = "groups-test" // Deprecated. Use only the TEST_REPO_NAME.
	CODEREVIEW_REPO_NAME string = "code-reviews"

	PERMISSION_ADMIN string = "admin"
	PERMISSION_PULL  string = "pull"
	PERMISSION_PUSH  string = "push"

	INDIVIDUAL int = iota
	GROUP
)
