package web

// TODO Clean up: some of these error messages seem to be less useful;
// Or rather it should not be possible to click on things that only teachers
// have access to when you are a student. Maybe I'm misunderstanding their use.
// Are these errors ever shown to users in the Web UI?? Or is this just for
// logging.

// JSONErrorMsg represents the error part of a JSON reply.
type JSONErrorMsg struct {
	Error    bool   `json:"Error"` //TODO Why have a boolean here?
	ErrorMsg string `json:"ErrorMsg"`
}

var (
	// ErrSignIn is a standard JSON error reply when a user is not signed in.
	ErrSignIn = &JSONErrorMsg{true, "You are not signed in. Please sign in to preform actions."}

	// ErrAccessToken is a standard JSON error reply when the access token cant be found.
	ErrAccessToken = &JSONErrorMsg{true, "Couldn't get your access token. Try to sign in again."}

	// ErrNotMember is a standard JSON error reply when a user is not a member of a course.
	ErrNotMember = &JSONErrorMsg{true, "You are not a member of this course."}

	// ErrUnknownMember is a standard JSON error reply when a user is not a member of a course.
	ErrUnknownMember = &JSONErrorMsg{true, "Unknown student."}

	// ErrNotAdmin is a standard JSON error reply when a user is not a admin.
	ErrNotAdmin = &JSONErrorMsg{true, "You don't have administrator privileges. If you are an administrator, try to sign in again."}

	// ErrNotTeacher is a standard JSON error reply when a user is not a teacher.
	ErrNotTeacher = &JSONErrorMsg{true, "You are not a teacher for this course."}

	// ErrMissingField is a standard JSON error reply when a request is missing fields.
	ErrMissingField = &JSONErrorMsg{true, "Missing required parameters."} // TODO Fix text

	// ErrInvalidAdminField is a standard JSON error reply when a invalid admin field is recieved.
	ErrInvalidAdminField = &JSONErrorMsg{true, "Can't use admin parameters."} // TODO Fix text

	// ErrInvalidTeacherField is a standard JSON error reply when a invalid teacher field is recieved.
	ErrInvalidTeacherField = &JSONErrorMsg{true, "Can't use teacher parameters."} // TODO Fix text

	// ErrNotStored is a standard JSON error reply when a storage error occures.
	ErrNotStored = &JSONErrorMsg{true, "Changes have not been stored."} // TODO should we provide guidelines

	// ErrUnknownCourse is a standard JSON error reply when the course is unknown.
	ErrUnknownCourse = &JSONErrorMsg{true, "Unknown course."}

	// ErrUnknownGroup is a standard JSON error reply when the course is unknown.
	ErrUnknownGroup = &JSONErrorMsg{true, "Unknown group."}
)
