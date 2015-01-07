package global

import (
	"net/http"
)

var (
	Hostname           string = ""
	OAuth_ClientID     string = ""
	OAuth_ClientSecret string = ""
	OAuth_Scope        string = ""
	OAuth_State        string = ""
	OAuth_RedirectURL  string = ""
	OAuth_Handler             = func(w http.ResponseWriter, r *http.Request) {
		// Empty placeholder
	}
	Basepath string = ""
)
