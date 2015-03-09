package web

// JSONErrorMsg represents the error part of a JSON reply.
type JSONErrorMsg struct {
	Error    bool   `json:"Error"`
	ErrorMsg string `json:"ErrorMsg"`
}

var (
	// ErrSignIn is a standard JSON error reply when a user is not signed in.
	ErrSignIn = &JSONErrorMsg{true, "You are not signed in. Please sign in to preform actions."}

	// ErrAccessToken is a standard JSON error reply when the access token cant be found.
	ErrAccessToken = &JSONErrorMsg{true, "Couldn't get your access token. Try to sign in again."}

	// ErrNotMember is a standard JSON error reply when a user is not a member of a course.
	ErrNotMember = &JSONErrorMsg{true, "You are not a member of this course."}

	// ErrNotMember is a standard JSON error reply when a user is not a member of a course.
	ErrUnknownMember = &JSONErrorMsg{true, "Unknown student."}

	// ErrNotAdmin is a standard JSON error reply when a user is not a admin.
	ErrNotAdmin = &JSONErrorMsg{true, "You are not a administrator. If you infact are an administrator, try to sign in again."} // TODO Fix text

	// ErrNotAdmin is a standard JSON error reply when a user is not a admin.
	ErrNotTeacher = &JSONErrorMsg{true, "You are not teaching this course."} // TODO Fix text

	// ErrMissingField is a standard JSON error reply when a request is midding fields.
	ErrMissingField = &JSONErrorMsg{true, "Missing required parameters."} // TODO Fix text

	// ErrInvalidAdminField is a standard JSON error reply when a invalid admin field is recieved.
	ErrInvalidAdminField = &JSONErrorMsg{true, "Can't use admin parameters."} // TODO Fix text

	// ErrInvalidTeacherField is a standard JSON error reply when a invalid teacher field is recieved.
	ErrInvalidTeacherField = &JSONErrorMsg{true, "Can't use teacher parameters."} // TODO Fix text

	// ErrNotStored is a standard JSON error reply when a storage error occures.
	ErrNotStored = &JSONErrorMsg{true, "Edit not stored in system."}

	// ErrUnknownCourse is a standard JSON error reply when the course is unknown.
	ErrUnknownCourse = &JSONErrorMsg{true, "Unknown course."}

	// ErrUnknownCourse is a standard JSON error reply when the course is unknown.
	ErrUnknownGroup = &JSONErrorMsg{true, "Unknown group."}
)
