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
	stdTemplate

	PointsToNextLvl    int64
	PercentLvlComplete int

	MissingName      bool
	MissingStudentID bool
	MissingEmail     bool
}

// ProfileHandler is a http handler which writes back a page about the
// users profile settings. The page can also be used to edit profile data.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
	if err != nil {
		// error getting access token from session
		logErrorAndRedirect(w, r, pages.Front, err)
		return
	}

	m, err := git.LookupMember(value.(string))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO Move level calculation to user score object
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
		stdTemplate: stdTemplate{
			Member:           m,
			OptionalHeadline: true,
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
			http.Redirect(w, r, pages.Profile, http.StatusTemporaryRedirect)
			return
		}

		if !auth.IsApprovedUser(r) {
			http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
			return
		}

		value, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
		if err != nil {
			// error getting access token from session
			logErrorAndRedirect(w, r, pages.Front, err)
			return
		}

		//TODO Should this be replaced with a Update() transaction with LookupMember() and Put()??
		member, err := git.LookupMember(value.(string))
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			// failed to convert string to int for studentid
			logErrorAndRedirect(w, r, pages.Profile, err)
			return
		}

		member.StudentID = studentid

		email, err := mail.ParseAddress(r.FormValue("email"))
		if err != nil {
			// failed to parse email address
			logErrorAndRedirect(w, r, pages.Profile, err)
			return
		}
		member.Email = email

		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", http.StatusNotFound)
	}
}
