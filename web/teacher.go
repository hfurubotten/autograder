package web

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
)

// TeachersPanelView is the view passed to the html template compiler in TeachersPanelHandler.
type TeachersPanelView struct {
	stdTemplate
	Org *git.Organization

	PendingUser  map[string]interface{}
	PendingGroup map[string]*git.Group

	CurrentLabType int
}

// TeachersPanelURL is the URL used to call TeachersPanelHandler.
var TeachersPanelURL = "/course/teacher/"

// TeachersPanelHandler is a http handler serving the Teacher panel.
// This page shows a summary of all the students and groups.
func TeachersPanelHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		logErrorAndRedirect(w, r, pages.Home, err)
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 4 {
		if !git.HasOrganization(path[3]) {
			http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
			return
		}

		orgname = path[3]
	} else {
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !org.IsTeacher(member) {
		log.Println("User is not a teacher for this course.")
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}

	// gets pending users
	users := org.PendingUser
	var status string
	for username := range users {
		// check status up against Github
		//TODO Check for errors from GetMember()
		users[username], err = git.GetMember(username)
		if err != nil {
			continue
		}

		status, err = org.GetMembership(users[username].(*git.Member))
		if err != nil {
			log.Println(err)
			continue
		}

		if status == "active" {
			continue
			// TODO: what about group assignments?
		} else if status == "pending" {
			delete(users, username)
		} else {
			delete(users, username)
			log.Println("Got a unexpected status back from Github regarding Membership")
		}
	}

	// gets teachers
	for username := range org.Teachers {
		org.Teachers[username], _ = git.GetMember(username)
	}

	// gets users
	for username := range org.Members {
		org.Members[username], _ = git.GetMember(username)
	}

	// get pending groups
	pendinggroups := make(map[string]*git.Group)
	for groupName := range org.PendingGroup {
		group, err := git.GetGroup(groupName)
		if err != nil {
			log.Println(err)
		}

		if group.Course != org.Name {
			delete(org.PendingGroup, groupName)
			err := org.Save()
			if err != nil {
				log.Println(err)
			}
			continue
		}

		for key := range group.Members {
			groupmember, _ := git.GetMember(key)
			group.Members[key] = groupmember
		}

		pendinggroups[groupName] = group
	}

	// get groups
	for groupName := range org.Groups {
		group, err := git.GetGroup(groupName)
		if err != nil {
			log.Println(err)
		}
		for key := range group.Members {
			groupmember, err := git.GetMember(key)
			if err != nil {
				log.Println(err)
			}
			group.Members[key] = groupmember //TODO what does this do?
		}
		org.Groups[groupName] = group
	}

	_, _, labtype := org.FindCurrentLab()

	view := TeachersPanelView{
		stdTemplate: stdTemplate{
			Member: member,
		},
		PendingUser:    users,
		Org:            org,
		PendingGroup:   pendinggroups,
		CurrentLabType: labtype,
	}
	execTemplate("teacherspanel.html", w, view)
}

// ShowResultView is the view passed to the html template compiler in ShowResultHandler.
type ShowResultView struct {
	stdTemplate
	Org      *git.Organization
	Username string
	Labnum   int
	IsGroup  bool
	GroupID  int
}

// ShowResultURL is the URL used to call ShowResultHandler.
var ShowResultURL = "/course/result/"

// ShowResultHandler is a http handler for showing a page detailing
// lab resutls for a single user or group.
func ShowResultHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 4 {
		if !git.HasOrganization(path[3]) {
			http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
			return
		}

		orgname = path[3]
	} else {
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}

	username := r.FormValue("user")
	if username == "" {
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}

	if !git.HasOrganization(orgname) {
		http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		return
	}

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	isgroup := false
	groupid := -1
	labnum := 0
	if !git.HasMember(username) {
		groupnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			logErrorAndRedirect(w, r, pages.Home, err)
			return
		}
		if git.HasGroup(groupnum) {
			isgroup = true
			group, err := git.GetGroup(username) // username==groupname TODO consider changing this
			if err != nil {
				logErrorAndRedirect(w, r, pages.Home, err)
				return
			}

			groupid = group.ID

			if group.CurrentLabNum >= org.GroupAssignments {
				labnum = org.GroupAssignments
			} else {
				labnum = group.CurrentLabNum
			}
		} else {
			http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
			return
		}
	} else {
		user, err := git.GetMember(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		nr := user.Courses[org.Name].CurrentLabNum
		if nr >= org.IndividualAssignments {
			labnum = org.IndividualAssignments
		} else {
			labnum = nr
		}
	}

	view := ShowResultView{
		stdTemplate: stdTemplate{
			Member: member,
		},
		Org:      org,
		Username: username,
		Labnum:   labnum,
		IsGroup:  isgroup,
		GroupID:  groupid,
	}
	execTemplate("teacherresultpage.html", w, view)
}

// AddAssistantURL is the URL used to call AddAssistantHandler.
var AddAssistantURL = "/course/addassistant"

// AddAssistantHandler is a http handler used to add users as assistants on a course.
func AddAssistantHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, false)
	if err != nil {
		log.Println(err)
		return
	}

	username := r.FormValue("assistant")
	course := r.FormValue("course")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown course.", http.StatusNotFound)
		return
	}

	if username == member.Username {
		return
	}

	assistant, err := git.GetMember(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := assistant.Save(); err != nil {
			assistant.Unlock()
			log.Println(err)
		}
	}()

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := org.Save(); err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		http.Error(w, "User is not the teacher for this course.", http.StatusNotFound)
		return
	}

	assistant.AddAssistingOrganization(org)

	org.AddTeacher(assistant)
	if _, ok := org.PendingUser[username]; ok {
		delete(org.PendingUser, username)
	}
}

// RemoveAssistantURL is the URL used to call RemoveAssistantHandler.
var RemoveAssistantURL = "/course/removeassistant"

// RemoveAssistantHandler is a http handler used to remove users as assistants on a course.
func RemoveAssistantHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, false)
	if err != nil {
		log.Println(err)
		return
	}

	username := r.FormValue("assistant")
	course := r.FormValue("course")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown course.", http.StatusNotFound)
		return
	}

	if username == member.Username {
		return
	}

	assistant, err := git.GetMember(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := assistant.Save(); err != nil {
			assistant.Unlock()
			log.Println(err)
		}
	}()

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := org.Save(); err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		http.Error(w, "User is not the teacher for this course.", http.StatusNotFound)
		return
	}

	assistant.RemoveAssistingOrganization(org)

	org.RemoveTeacher(assistant)
}
