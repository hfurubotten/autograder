package git

type RepositoryOptions struct {
	Name     string
	Private  bool
	TeamID   int
	AutoInit bool
	Hook     bool
}

type Repo struct {
	Name     string
	HTMLURL  string
	CloneURL string
	Private  bool
	TeamID   int
}
