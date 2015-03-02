package web

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
)

type TeachersPanelView struct {
	Member *git.Member
	Org    *git.Organization

	HasPendingUsers bool
	PendingUser     map[string]interface{}

	HasPendingRandomGroupAssignees bool
	HasPendingGroups               bool
	PendingGroup                   map[int]*git.Group
}

// TeachersPanelURL is the URL used to call TeachersPanelHandler.
var TeachersPanelURL string = "/course/teacher/"

// TeachersPanelHandler is a http handler serving the Teacher panel.
// This page shows a summary of all the students and groups.
func TeachersPanelHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
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

	org, err := git.NewOrganization(orgname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	org.Lock()
	defer org.Unlock()

	if _, ok := org.Teachers[member.Username]; !ok {
		// migrate from bug where org does not contain teacher names.
		if _, ok := member.Teaching[org.Name]; ok {
			org.AddTeacher(member)
			org.Save()
		} else {
			log.Println("User is not a teacher for this course.")
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}

	}

	// gets pending users
	users := org.PendingUser
	var status string
	for username, _ := range users {
		// check status up against Github
		users[username], err = git.NewMemberFromUsername(username)
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

	// gets users
	for username, _ := range org.Members {
		org.Members[username], _ = git.NewMemberFromUsername(username)
	}

	// get pending groups
	pendinggroups := make(map[int]*git.Group)
	for groupID, _ := range org.PendingGroup {
		group, err := git.NewGroup(org.Name, groupID)
		if err != nil {
			log.Println(err)
		}
		for key, _ := range group.Members {
			groupmember, _ := git.NewMemberFromUsername(key)
			group.Members[key] = groupmember
		}
		pendinggroups[groupID] = group
	}

	// get groups
	for groupname, _ := range org.Groups {
		groupID, _ := strconv.Atoi(groupname[5:])
		group, _ := git.NewGroup(org.Name, groupID)
		for key, _ := range group.Members {
			groupmember, _ := git.NewMemberFromUsername(key)
			group.Members[key] = groupmember
		}
		org.Groups[groupname] = group
	}

	view := TeachersPanelView{
		Member:       member,
		PendingUser:  users,
		Org:          org,
		PendingGroup: pendinggroups,
	}
	execTemplate("teacherspanel.html", w, view)
}

type ShowResultView struct {
	Member   *git.Member
	Org      *git.Organization
	Username string
	Labnum   int
	IsGroup  bool
}

// ShowResultURL is the URL used to call ShowResultHandler.
var ShowResultURL string = "/course/result/"

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

	org, err := git.NewOrganization(orgname)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	isgroup := false
	labnum := 0
	if !git.HasMember(username) {
		groupnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}
		if git.HasGroup(org.Name, groupnum) {
			isgroup = true
			group, err := git.NewGroup(org.Name, groupnum)
			if err != nil {
				http.Redirect(w, r, pages.HOMEPAGE, 307)
				return
			}
			if group.CurrentLabNum >= org.GroupAssignments {
				labnum = org.GroupAssignments - 1
			} else {
				labnum = group.CurrentLabNum - 1
			}
		} else {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}
	} else {
		user, err := git.NewMemberFromUsername(username)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		nr := user.Courses[org.Name].CurrentLabNum
		if nr >= org.IndividualAssignments {
			labnum = org.IndividualAssignments - 1
		} else {
			labnum = nr - 1
		}
	}

	view := ShowResultView{
		Member:   member,
		Org:      org,
		Username: username,
		Labnum:   labnum,
		IsGroup:  isgroup,
	}
	execTemplate("teacherresultpage.html", w, view)
}

// AddAssistantURL is the URL used to call AddAssistantHandler.
var AddAssistantURL string = "/course/addassistant"

// AddAssistantHandler is a http handler used to add users as assistants on a course.
func AddAssistantHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
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

	assistant, err := git.NewMemberFromUsername(username)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	assistant.Lock()
	defer assistant.Unlock()

	org, err := git.NewOrganization(course)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	org.Lock()
	defer org.Unlock()

	if !org.IsTeacher(member) {
		http.Error(w, "User is not the teacher for this course.", 404)
		return
	}

	assistant.AddAssistingOrganization(org)
	assistant.Save()

	org.AddTeacher(assistant)
	if _, ok := org.PendingUser[username]; ok {
		delete(org.PendingUser, username)
	}
	org.Save()

}
