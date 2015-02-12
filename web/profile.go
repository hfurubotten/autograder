package web

import (
	"log"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

type profileview struct {
	Member git.Member
}

var ProfileURL string = "/profile"

// profilehandler is a http handler which writes back a page about the
// users profile settings. The page can also be used to edit profile data.
func profilehandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		return
	}

	view := profileview{}
	view.Member = git.NewMember(value.(string))
	execTemplate("profile.html", w, view)
}

var UpdateMemberURL string = "/updatemember"

//  updatememberhandler is a http handler for updating a users profile data.
func updatememberhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("name") == "" || r.FormValue("studentid") == "" || r.FormValue("email") == "" {
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		if !auth.IsApprovedUser(r) {
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
		if err != nil {
			log.Println("Error getting access token from sessions: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		member := git.NewMember(value.(string))
		member.Name = r.FormValue("name")
		studentid, err := strconv.Atoi(r.FormValue("studentid"))
		if err != nil {
			log.Println("studentid atoi error: ", err)
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		member.StudentID = studentid

		email, err := mail.ParseAddress(r.FormValue("email"))
		if err != nil {
			log.Println("Parsing email error: ", err)
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}
		member.Email = email

		member.StickToSystem()

		http.Redirect(w, r, pages.HOMEPAGE, 307)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}
