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
	StdTemplate
	Org *git.Organization

	PendingUser  map[string]interface{}
	PendingGroup map[int]*git.Group

	CurrentLabType int
}

// TeachersPanelURL is the URL used to call TeachersPanelHandler.
var TeachersPanelURL = "/course/teacher/"

// TeachersPanelHandler is a http handler serving the Teacher panel.
// This page shows a summary of all the students and groups.
func TeachersPanelHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	log.Println(r)
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 4 {
		if !git.HasOrganization(path[3]) {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[3]
	} else {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !org.IsTeacher(member) {
		log.Println("User is not a teacher for this course.")
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	// gets pending users
	users := org.PendingUser
	var status string
	for username := range users {
		// check status up against Github
		users[username], err = git.NewMemberFromUsername(username, true)
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
		org.Teachers[username], _ = git.NewMemberFromUsername(username, true)
	}

	// gets users
	for username := range org.Members {
		org.Members[username], _ = git.NewMemberFromUsername(username, true)
	}

	// get pending groups
	pendinggroups := make(map[int]*git.Group)
	for groupID := range org.PendingGroup {
		group, err := git.NewGroup(org.Name, groupID, true)
		if err != nil {
			log.Println(err)
		}

		if group.Course != org.Name {
			org.Lock()
			delete(org.PendingGroup, groupID)
			err := org.Save()
			if err != nil {
				log.Println(err)
				org.Unlock()
			}
			continue
		}

		for key := range group.Members {
			groupmember, _ := git.NewMemberFromUsername(key, true)
			group.Members[key] = groupmember
		}

		pendinggroups[groupID] = group
	}

	// get groups
	for groupname := range org.Groups {
		groupID, _ := strconv.Atoi(groupname[5:])
		group, _ := git.NewGroup(org.Name, groupID, true)
		for key := range group.Members {
			groupmember, _ := git.NewMemberFromUsername(key, true)
			group.Members[key] = groupmember
		}
		org.Groups[groupname] = group
	}

	_, _, labtype := org.FindCurrentLab()

	view := TeachersPanelView{
		StdTemplate: StdTemplate{
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
	StdTemplate
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
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[3]
	} else {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	username := r.FormValue("user")
	if username == "" {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	if !git.HasOrganization(orgname) {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	isgroup := false
	groupid := -1
	labnum := 0
	if !git.HasMember(username) {
		groupnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}
		if git.HasGroup(groupnum) {
			isgroup = true
			group, err := git.NewGroup(org.Name, groupnum, true)
			if err != nil {
				http.Redirect(w, r, pages.HOMEPAGE, 307)
				return
			}

			groupid = group.ID

			if group.CurrentLabNum >= org.GroupAssignments {
				labnum = org.GroupAssignments
			} else {
				labnum = group.CurrentLabNum
			}
		} else {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}
	} else {
		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		nr := user.Courses[org.Name].CurrentLabNum
		if nr >= org.IndividualAssignments {
			labnum = org.IndividualAssignments
		} else {
			labnum = nr
		}
	}

	view := ShowResultView{
		StdTemplate: StdTemplate{
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
		http.Error(w, "Unknown course.", 404)
		return
	}

	if username == member.Username {
		return
	}

	assistant, err := git.NewMemberFromUsername(username, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, err.Error(), 500)
		return
	}

	defer func() {
		if err := org.Save(); err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		http.Error(w, "User is not the teacher for this course.", 404)
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
		http.Error(w, "Unknown course.", 404)
		return
	}

	if username == member.Username {
		return
	}

	assistant, err := git.NewMemberFromUsername(username, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, err.Error(), 500)
		return
	}

	defer func() {
		if err := org.Save(); err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		http.Error(w, "User is not the teacher for this course.", 404)
		return
	}

	assistant.RemoveAssistingOrganization(org)

	org.RemoveTeacher(assistant)
}
