package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web/sessions"
)

type adminview struct {
	Member  *git.Member
	Members []git.Member
}

func adminhandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkAdminApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := adminview{}
	view.Member = &member
	view.Members = git.ListAllMembers()

	t, err := template.ParseFiles(global.Basepath+"web/html/admin.html", global.Basepath+"web/html/template.html")
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
	enc := json.NewEncoder(w)
	if !auth.IsApprovedUser(r) {
		err = enc.Encode(ErrSignIn)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		err = enc.Encode(ErrAccessToken)
		return
	}

	member := git.NewMember(value.(string))

	if !member.IsAdmin {
		log.Println("Unautorized request of admin page.")
		err = enc.Encode(ErrNotAdmin)
		return
	}

	//TODO Check logic of the following two errors
	if r.FormValue("user") == "" || r.FormValue("admin") == "" {
		err = enc.Encode(ErrMissingField)
		return
	}

	m := git.NewMemberFromUsername(r.FormValue("user"))
	m.IsAdmin, err = strconv.ParseBool(r.FormValue("admin"))
	if err != nil {
		err = enc.Encode(ErrInvalidAdminField)
		return
	}

	err = m.StickToSystem()
	if err != nil {
		err = enc.Encode(ErrNotStored)
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
	enc := json.NewEncoder(w)
	if !auth.IsApprovedUser(r) {
		err = enc.Encode(ErrSignIn)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		err = enc.Encode(ErrAccessToken)
		return
	}

	member := git.NewMember(value.(string))

	if !member.IsAdmin {
		log.Println("Unautorized request of admin page.")
		err = enc.Encode(ErrNotAdmin)
		return
	}

	//TODO Check logic
	if r.FormValue("user") == "" || r.FormValue("teacher") == "" {
		err = enc.Encode(ErrMissingField)
		return
	}

	m := git.NewMemberFromUsername(r.FormValue("user"))
	m.IsTeacher, err = strconv.ParseBool(r.FormValue("teacher"))
	if err != nil {
		err = enc.Encode(ErrInvalidTeacherField)
		return
	}

	err = m.StickToSystem()
	if err != nil {
		err = enc.Encode(ErrNotStored)
		return
	}

	//TODO should this struct also have json tags?
	msg := struct {
		User    string
		Teacher bool
	}{
		User:    m.Username,
		Teacher: m.IsTeacher,
	}
	err = enc.Encode(msg)
	return
}
