package git

type TeamOptions struct {
	Name       string
	Permission string
	RepoNames  []string
}

type Team struct {
	ID          int
	Name        string
	Permission  string
	Repocount   int
	MemberCount int
}
