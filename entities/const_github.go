package entities

const (
	// Github repository types
	usrType string = "User"
	orgType string = "Organization"

	// Standard github team names
	studentsTeam = "students"
	teachersTeam = "Owners" //TODO Should be changed to 'teachers' and correspondingly we need to create this team at the start of the course.

	// Team permission names on github
	AdminPermission = "admin"
	PullPermission  = "pull"
	PushPermission  = "push"
)
