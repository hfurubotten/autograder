package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

type courseview struct {
	Member *git.Member
	Org    string
	Orgs   []string
}

func newcoursehandler(w http.ResponseWriter, r *http.Request) {
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

	view := courseview{}

	member := git.NewMember(value.(string))
	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	view.Member = &member

	var page string
	switch r.URL.Path {
	case "/course/new":
		page = "web/html/newcourse-info.html"
	case "/course/new/org":
		page = "web/html/newcourse-orgselect.html"

		view.Orgs, err = member.ListOrgs()
		if err != nil {
			log.Println(err)
			pages.RedirectTo(w, r, pages.SIGNOUT, 307)
			return
		}
	}

	t, err := template.ParseFiles(page)
	if err != nil {
		log.Println("Error parsing register html: ", err)
		return
	}

	err = t.Execute(w, view)
	if err != nil {
		log.Println("Error execute register html: ", err)
		return
	}
}

func selectorghandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	token, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	view := courseview{}

	if path := strings.Split(r.URL.Path, "/"); len(path) == 5 {
		view.Org = path[4]
	} else {
		pages.RedirectTo(w, r, "/course/new", 307)
		return
	}

	member := git.NewMember(token.(string))
	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	view.Member = &member
	view.Orgs, err = member.ListOrgs()
	if err != nil {
		log.Println(err)
		pages.RedirectTo(w, r, pages.SIGNOUT, 307)
		return
	}

	page := "web/html/newcourse-register.html"
	t, err := template.ParseFiles(page)
	if err != nil {
		log.Println("Error parsing register html: ", err)
		return
	}

	err = t.Execute(w, view)
	if err != nil {
		log.Println("Error execute register html: ", err)
		return
	}
}

func saveorghandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	if r.FormValue("org") == "" || r.FormValue("indv") == "" || r.FormValue("groups") == "" {
		log.Println("Missing POST elements request for new course creation. ")
		pages.RedirectTo(w, r, "/course/new/org", 307)
		return
	}

	token, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	member := git.NewMember(token.(string))
	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	org := git.NewOrganization(r.FormValue("org"))
	org.AdminToken = token.(string)
	org.Private = r.FormValue("private") == "on"
	org.Description = r.FormValue("desc")
	groups, err := strconv.Atoi(r.FormValue("groups"))
	if err != nil {
		log.Println("Cannot convert number of groups assignments from string to int: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}
	org.GroupAssignments = groups
	indv, err := strconv.Atoi(r.FormValue("indv"))
	if err != nil {
		log.Println("Cannot convert number of individual assignments from string to int: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}
	org.IndividualAssignments = indv

	if r.FormValue("template") == "" {
		repo := git.RepositoryOptions{
			Name:     git.STANDARD_REPO_NAME,
			Private:  org.Private,
			AutoInit: true,
		}
		err = org.CreateRepo(repo)
		if err != nil {
			log.Println(err)
			return
		}

		for i := 0; i < org.IndividualAssignments; i++ {
			path := "lab" + strconv.Itoa(i+1) + "/README.md"
			commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
			content := "# Lab assignment " + strconv.Itoa(i+1)
			err = org.CreateFile(git.STANDARD_REPO_NAME, path, content, commitmessage)
			if err != nil {
				log.Println(err)
				return
			}
		}

		if org.GroupAssignments > 0 {
			repo.Name = git.GROUPS_REPO_NAME
			err = org.CreateRepo(repo)
			if err != nil {
				log.Println(err)
				return
			}

			for i := 0; i < org.GroupAssignments; i++ {
				path := "lab" + strconv.Itoa(i+1) + "/README.md"
				commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
				content := "# Group assignment " + strconv.Itoa(i+1)
				err = org.CreateFile(git.GROUPS_REPO_NAME, path, content, commitmessage)
				if err != nil {
					log.Println(err)
				}
			}
		}

		team := git.TeamOptions{
			Name:       "students",
			Permission: git.PERMISSION_PULL,
			RepoNames:  []string{git.STANDARD_REPO_NAME, git.GROUPS_REPO_NAME},
		}
		org.StudentTeamID, err = org.CreateTeam(team)
		if err != nil {
			log.Println(err)
			return
		}

	} else {
		org.Fork(r.FormValue("template"), git.STANDARD_REPO_NAME)
		if org.GroupAssignments > 0 {
			org.Fork(r.FormValue("template"), git.STANDARD_REPO_NAME)
		}
	}

	err = org.StickToSystem()
	if err != nil {
		log.Println(err)
		return
	}

	member.AddTeachingOrganization(org)
	err = member.StickToSystem()
	if err != nil {
		log.Println(err)
		return
	}

	pages.RedirectTo(w, r, pages.FRONTPAGE, 307)

}

type newmemberview struct {
	Member *git.Member
	Orgs   []git.Organization
}

func newcoursememberhandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	token, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	member := git.NewMember(token.(string))
	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	view := newmemberview{}
	view.Member = &member
	view.Orgs = git.ListRegisteredOrganizations()

	page := "web/html/course-registermember.html"
	t, err := template.ParseFiles(page, "web/html/template.html")
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

func registercoursememberhandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	token, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	member := git.NewMember(token.(string))
	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	view := newmemberview{}
	view.Member = &member
	view.Orgs = git.ListRegisteredOrganizations()

	page := "web/html/course-registermember.html"
	t, err := template.ParseFiles(page)
	if err != nil {
		log.Println("Error parsing register html: ", err)
		return
	}

	err = t.Execute(w, view)
	if err != nil {
		log.Println("Error execute register html: ", err)
		return
	}
}
