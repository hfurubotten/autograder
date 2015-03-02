package web

type errMsg struct {
	Error string
}

type JSONErrorMsg struct {
	Error    bool   `json:Error`
	ErrorMsg string `json:ErrorMsg`
}

var (
	ErrSignIn              = &errMsg{"xx You are not signed in. Please sign in to preform actions."}
	ErrAccessToken         = &errMsg{"Couldn't get your access token. Try to sign in again."}
	ErrNotAdmin            = &errMsg{"You are not a administrator. If you infact are an administrator, try to sign in again."} // TODO Fix text
	ErrMissingField        = &errMsg{"Missing required parameters."}                                                           // TODO Fix text
	ErrInvalidAdminField   = &errMsg{"Can't use admin parameters."}                                                            // TODO Fix text
	ErrInvalidTeacherField = &errMsg{"Can't use teacher parameters."}                                                          // TODO Fix text
	ErrNotStored           = &errMsg{"Edit not stored in system."}
)
