package web

import (
	"log"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/game/levels"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

// ProfileView is the view passed to the html template compiler for ProfileHandler.
type ProfileView struct {
	StdTemplate

	PointsToNextLvl    int64
	PercentLvlComplete int

	MissingName      bool
	MissingStudentID bool
	MissingEmail     bool
}

// ProfileURL is the URL used to call ProfileHandler.
var ProfileURL = "/profile"

// ProfileHandler is a http handler which writes back a page about the
// users profile settings. The page can also be used to edit profile data.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		return
	}

	m, err := git.NewMember(value.(string))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// Level calculations
	lvlPoint := levels.RequiredForLevel(m.Level - 1)
	nextLvlPoint := levels.RequiredForLevel(m.Level)
	diffPointsNextLvl := nextLvlPoint - lvlPoint
	diffUser := diffPointsNextLvl - (m.TotalScore - lvlPoint)
	percentDone := 100 - int(float64(diffUser)/float64(diffPointsNextLvl)*100)

	if percentDone > 100 {
		percentDone = 100
	} else if percentDone < 0 {
		percentDone = 0
	}

	view := ProfileView{
		StdTemplate: StdTemplate{
			Member:          m,
			OptinalHeadline: true,
		},
		PointsToNextLvl:    diffUser,
		PercentLvlComplete: percentDone,
		MissingName:        m.Name == "",
		MissingStudentID:   m.StudentID == 0,
		MissingEmail:       m.Email == nil,
	}
	execTemplate("profile.html", w, view)
}

// UpdateMemberURL is the URL used to call UpdateMemberHandler.
var UpdateMemberURL = "/updatemember"

// UpdateMemberHandler is a http handler for updating a users profile data.
func UpdateMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("name") == "" || r.FormValue("studentid") == "" || r.FormValue("email") == "" {
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		if !auth.IsApprovedUser(r) {
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		value, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
		if err != nil {
			log.Println("Error getting access token from sessions: ", err)
			http.Redirect(w, r, pages.FRONTPAGE, 307)
			return
		}

		member, err := git.NewMember(value.(string))
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		defer func() {
			if err := member.Save(); err != nil {
				member.Unlock()
				log.Println("Error storing:", err)
			}
		}()

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

		http.Redirect(w, r, pages.HOMEPAGE, 307)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}
