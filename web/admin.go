package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/hfurubotten/autograder/config"
	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
)

// AdminURL is used to call AdminHandler.
var AdminURL = "/admin"

// AdminHandler is the http handler for serving the admin panel.
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkAdminApproval(w, r, true)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}
	adminView := struct {
		SysName          string
		OptionalHeadline bool
		Member           *entities.Member
		Members          []*entities.Member
	}{
		config.SysName, false, member, entities.ListAllMembers(),
	}
	execTemplate("admin.html", w, adminView)
}

// SetAdminView represents the view sent back in the JSON reply from SetAdminHandler.
type SetAdminView struct {
	JSONErrorMsg
	User  string `json:"User"`
	Admin bool   `json:"Admin"`
}

// SetAdminURL is used to call SetAdminHandler.
var SetAdminURL = "/admin/user"

// SetAdminHandler is a http handler to toggle the admin privileges of a user.
func SetAdminHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	// check if the user toggling the admin button has admin privileges
	_, err := checkAdminApproval(w, r, false)
	if err != nil {
		log.Println("Unauthorized request for admin page:", err)
		err = enc.Encode(ErrNotAdmin)
		return
	}

	user, admin := r.FormValue("user"), r.FormValue("admin")
	if user == "" || admin == "" {
		log.Printf("Missing required field: user=%s ; admin=%s", user, admin)
		err = enc.Encode(ErrMissingField)
		return
	}

	// get the selected member's details
	m, err := entities.GetMember(user)
	if err != nil {
		log.Println("Member not found:", err)
		err = enc.Encode(ErrUnknownMember)
		return
	}

	m.IsAdmin, err = strconv.ParseBool(admin)
	if err != nil {
		log.Println("Invalid admin field:", err)
		err = enc.Encode(ErrInvalidAdminField)
		return
	}

	err = m.Save()
	if err != nil {
		log.Println("Failed to update member to admin status:", err)
		err = enc.Encode(ErrNotStored)
		return
	}

	msg := SetAdminView{
		User:  m.Username,
		Admin: m.IsAdmin,
	}
	err = enc.Encode(msg)
}

// SetTeacherView represents the view sent back the JSON reply in SetTeacherHandler.
type SetTeacherView struct {
	JSONErrorMsg
	User    string `json:"User"`
	Teacher bool   `json:"Teacher"`
}

// SetTeacherURL is the URL used to call SetTeacherHandler.
var SetTeacherURL = "/admin/teacher"

// SetTeacherHandler is a http handler to toggle the teacher privileges of a user.
func SetTeacherHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	// check if the user toggling the teacher button has admin privileges
	_, err := checkAdminApproval(w, r, false)
	if err != nil {
		log.Println("Unauthorized request for admin page:", err)
		err = enc.Encode(ErrNotAdmin)
		return
	}

	user, teacher := r.FormValue("user"), r.FormValue("teacher")
	if user == "" || teacher == "" {
		log.Printf("Missing required field: user=%s ; teacher=%s", user, teacher)
		err = enc.Encode(ErrMissingField)
		return
	}

	m, err := entities.GetMember(user)
	if err != nil {
		log.Println("Member not found:", err)
		err = enc.Encode(ErrUnknownMember)
		return
	}

	m.IsTeacher, err = strconv.ParseBool(teacher)
	if err != nil {
		log.Println("Invalid teacher field:", err)
		err = enc.Encode(ErrInvalidTeacherField)
		return
	}

	err = m.Save()
	if err != nil {
		log.Println("Failed to update member to teacher status:", err)
		err = enc.Encode(ErrNotStored)
		return
	}

	msg := SetTeacherView{
		User:    m.Username,
		Teacher: m.IsTeacher,
	}
	err = enc.Encode(msg)
}
