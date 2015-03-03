package web

type JSONErrorMsg struct {
	Error    bool   `json:Error`
	ErrorMsg string `json:ErrorMsg`
}

var (
	ErrSignIn              = &JSONErrorMsg{true, "You are not signed in. Please sign in to preform actions."}
	ErrAccessToken         = &JSONErrorMsg{true, "Couldn't get your access token. Try to sign in again."}
	ErrNotMember           = &JSONErrorMsg{true, "You are not a member of this course."}
	ErrNotAdmin            = &JSONErrorMsg{true, "You are not a administrator. If you infact are an administrator, try to sign in again."} // TODO Fix text
	ErrMissingField        = &JSONErrorMsg{true, "Missing required parameters."}                                                           // TODO Fix text
	ErrInvalidAdminField   = &JSONErrorMsg{true, "Can't use admin parameters."}                                                            // TODO Fix text
	ErrInvalidTeacherField = &JSONErrorMsg{true, "Can't use teacher parameters."}                                                          // TODO Fix text
	ErrNotStored           = &JSONErrorMsg{true, "Edit not stored in system."}
	ErrUnknownCourse       = &JSONErrorMsg{true, "Unknown course."}
)
