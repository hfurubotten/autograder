package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/hfurubotten/autograder/git"
)

type adminview struct {
	Member  *git.Member
	Members []*git.Member
}

var AdminURL string = "/admin"

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkAdminApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := adminview{}
	view.Member = member
	view.Members = git.ListAllMembers()
	execTemplate("admin.html", w, view)
}

type SetAdminView struct {
	JSONErrorMsg
	User  string `json:User`
	Admin bool   `json:Admin`
}

var SetAdminURL string = "/admin/user"

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

	m, err := git.NewMemberFromUsername(r.FormValue("user"))
	if err != nil {

	}

	m.Lock()
	defer m.Unlock()

	m.IsAdmin, err = strconv.ParseBool(r.FormValue("admin"))
	if err != nil {
		err = enc.Encode(ErrInvalidAdminField)
		return
	}

	err = m.Save()
	if err != nil {
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

type SetTeacherView struct {
	JSONErrorMsg
	User    string `json:User`
	Teacher bool   `json:Teacher`
}

var SetTeacherURL string = "/admin/teacher"

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

	m, err := git.NewMemberFromUsername(r.FormValue("user"))
	if err != nil {
		log.Println("Unautorized request of admin page.") // TODO replace this with more appropiate msg.
		err = enc.Encode(ErrNotAdmin)                     // TODO replace this with more appropiate msg.
		return
	}

	m.Lock()
	defer m.Unlock()

	m.IsTeacher, err = strconv.ParseBool(r.FormValue("teacher"))
	if err != nil {
		err = enc.Encode(ErrInvalidTeacherField)
		return
	}

	err = m.Save()
	if err != nil {
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
