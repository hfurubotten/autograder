package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	git "github.com/hfurubotten/autograder/entities"
)

// AdminView is the struct passed to the html template compiler.
type AdminView struct {
	StdTemplate
	Members []*git.Member
}

// AdminURL is the URL used to call AdminHandler.
var AdminURL = "/admin"

// AdminHandler is a http handler which gives the administator page.
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkAdminApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := AdminView{}
	view.Member = member
	view.Members = git.ListAllMembers()
	execTemplate("admin.html", w, view)
}

// SetAdminView represents the view sendt back the JSON reply in SetAdminHandler.
type SetAdminView struct {
	JSONErrorMsg
	User  string `json:"User"`
	Admin bool   `json:"Admin"`
}

// SetAdminURL is the URL used to call SetAdminHandler.
var SetAdminURL = "/admin/user"

// SetAdminHandler is a http handler which can set or unset the admin property of a user.
func SetAdminHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	_, err := checkAdminApproval(w, r, false)
	if err != nil {
		log.Println("Unautorized request of admin page.")
		err = enc.Encode(ErrNotAdmin)
		return
	}

	//TODO Check logic of the following two errors
	if r.FormValue("user") == "" || r.FormValue("admin") == "" {
		err = enc.Encode(ErrMissingField)
		return
	}

	m, err := git.GetMember(r.FormValue("user"))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	m.IsAdmin, err = strconv.ParseBool(r.FormValue("admin"))
	if err != nil {
		m.Unlock()
		err = enc.Encode(ErrInvalidAdminField)
		return
	}

	err = m.Save()
	if err != nil {
		m.Unlock()
		err = enc.Encode(ErrNotStored)
		return
	}

	msg := SetAdminView{
		User:  m.Username,
		Admin: m.IsAdmin,
	}

	err = enc.Encode(msg)
	return
}

// SetTeacherView represents the view sendt back the JSON reply in SetTeacherHandler.
type SetTeacherView struct {
	JSONErrorMsg
	User    string `json:"User"`
	Teacher bool   `json:"Teacher"`
}

// SetTeacherURL is the URL used to call SetTeacherHandler.
var SetTeacherURL = "/admin/teacher"

// SetTeacherHandler is a http handler which can set or unset the teacher property of a user.
func SetTeacherHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	_, err := checkAdminApproval(w, r, false)
	if err != nil {
		log.Println("Unautorized request of admin page.")
		err = enc.Encode(ErrNotAdmin)
		return
	}

	//TODO Check logic
	if r.FormValue("user") == "" || r.FormValue("teacher") == "" {
		err = enc.Encode(ErrMissingField)
		return
	}

	m, err := git.GetMember(r.FormValue("user"))
	if err != nil {
		log.Println("Unautorized request of admin page.") // TODO replace this with more appropiate msg.
		err = enc.Encode(ErrNotAdmin)                     // TODO replace this with more appropiate msg.
		return
	}

	m.IsTeacher, err = strconv.ParseBool(r.FormValue("teacher"))
	if err != nil {
		m.Unlock()
		err = enc.Encode(ErrInvalidTeacherField)
		return
	}

	err = m.Save()
	if err != nil {
		m.Unlock()
		err = enc.Encode(ErrNotStored)
		return
	}

	msg := SetTeacherView{
		User:    m.Username,
		Teacher: m.IsTeacher,
	}
	err = enc.Encode(msg)
	return
}
