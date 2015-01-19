package web

import (
	"log"
	"net/http"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

type profileview struct {
	Member git.Member
}

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

func updatememberhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("name") == "" || r.FormValue("studentid") == "" {
			//pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
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
		member.StickToSystem()

		http.Redirect(w, r, pages.HOMEPAGE, 307)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}
