package web

import (
	"html/template"
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
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	view := profileview{}

	view.Member = git.NewMember(value.(string))

	t, err := template.ParseFiles("web/html/profile.html", "web/html/template.html")
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

func updatememberhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("name") == "" || r.FormValue("studentid") == "" {
			//pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

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
		member.Name = r.FormValue("name")
		studentid, err := strconv.Atoi(r.FormValue("studentid"))
		if err != nil {
			log.Println("studentid atoi error: ", err)
			pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		member.StudentID = studentid
		member.StickToSystem()

		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}
