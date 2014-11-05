package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

type adminview struct {
	Member  *git.Member
	Members []git.Member
}

func adminhandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	member := git.NewMember(value.(string))

	if !member.IsAdmin {
		log.Println("Unautorized request of admin page.")
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}
	view := adminview{}
	view.Member = &member
	view.Members = git.ListAllMembers()

	t, err := template.ParseFiles("web/html/admin.html", "web/html/template.html")
	if err != nil {
		log.Println("Error parsing register html: ", err)
		return
	}

	err = t.ExecuteTemplate(w, "template", view)
	if err != nil {
		log.Println("Error execute register html: ", err)
		return
	}
}

func setadminhandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var errmsg struct {
		Error string
	}
	enc := json.NewEncoder(w)
	if !auth.IsApprovedUser(r) {
		errmsg = struct {
			Error string
		}{
			Error: "You are not signed in. Please sign in to preform actions.",
		}
		err = enc.Encode(errmsg)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		errmsg = struct {
			Error string
		}{
			Error: "Couldn't get you access token. Try to sign in again.",
		}
		err = enc.Encode(errmsg)
		return
	}

	member := git.NewMember(value.(string))

	if !member.IsAdmin {
		log.Println("Unautorized request of admin page.")
		errmsg = struct {
			Error string
		}{
			Error: "You are not a administrator.",
		}
		err = enc.Encode(errmsg)
		return
	}

	if r.FormValue("user") == "" || r.FormValue("admin") == "" {
		errmsg = struct {
			Error string
		}{
			Error: "Missing required parameters. ",
		}
		err = enc.Encode(errmsg)
		return
	}

	m := git.NewMemberFromUsername(r.FormValue("user"))
	m.IsAdmin, err = strconv.ParseBool(r.FormValue("admin"))
	if err != nil {
		errmsg = struct {
			Error string
		}{
			Error: "Can't use admin parameters.",
		}
		err = enc.Encode(errmsg)
		return
	}

	err = m.StickToSystem()
	if err != nil {
		errmsg = struct {
			Error string
		}{
			Error: "Edit not stored in system.",
		}
		err = enc.Encode(errmsg)
		return
	}

	msg := struct {
		User  string `json:User`
		Admin bool   `json:Admin`
	}{
		User:  m.Username,
		Admin: m.IsAdmin,
	}
	err = enc.Encode(msg)
	return
}

func setteacherhandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var errmsg struct {
		Error string
	}
	enc := json.NewEncoder(w)
	if !auth.IsApprovedUser(r) {
		errmsg = struct {
			Error string
		}{
			Error: "You are not signed in. Please sign in to preform actions.",
		}
		err = enc.Encode(errmsg)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		errmsg = struct {
			Error string
		}{
			Error: "Couldn't get you access token. Try to sign in again.",
		}
		err = enc.Encode(errmsg)
		return
	}

	member := git.NewMember(value.(string))

	if !member.IsAdmin {
		log.Println("Unautorized request of admin page.")
		errmsg = struct {
			Error string
		}{
			Error: "You are not a administrator.",
		}
		err = enc.Encode(errmsg)
		return
	}

	if r.FormValue("user") == "" || r.FormValue("teacher") == "" {
		errmsg = struct {
			Error string
		}{
			Error: "Missing required parameters.",
		}
		err = enc.Encode(errmsg)
		return
	}

	m := git.NewMemberFromUsername(r.FormValue("user"))
	m.IsTeacher, err = strconv.ParseBool(r.FormValue("teacher"))
	if err != nil {
		errmsg = struct {
			Error string
		}{
			Error: "Can't use teacher parameters.",
		}
		err = enc.Encode(errmsg)
		return
	}

	err = m.StickToSystem()
	if err != nil {
		errmsg = struct {
			Error string
		}{
			Error: "Edit not stored in system.",
		}
		err = enc.Encode(errmsg)
		return
	}

	msg := struct {
		User  string
		Admin bool
	}{
		User:  m.Username,
		Admin: m.IsAdmin,
	}
	err = enc.Encode(msg)
	return
}
