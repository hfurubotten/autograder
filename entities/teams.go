package git

// TeamOptions represents the options needed when creating a team within a organization.
type TeamOptions struct {
	Name       string
	Permission string
	RepoNames  []string
}

// Team represents a existing team.
type Team struct {
	ID          int
	Name        string
	Permission  string
	Repocount   int
	MemberCount int
}
