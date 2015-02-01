package web

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
)

type teacherspanelview struct {
	Member       git.Member
	Org          git.Organization
	PendingUser  map[string]interface{}
	PendingGroup map[int]git.Group
}

func teacherspanelhandler(w http.ResponseWriter, r *http.Request) {
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

	org := git.NewOrganization(orgname)

	if _, ok := org.Teachers[member.Username]; !ok {
		// migrate from bug where org does not contain teacher names.
		if _, ok := member.Teaching[org.Name]; ok {
			org.AddTeacher(member)
			org.StickToSystem()
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
		users[username] = git.NewMemberFromUsername(username)
		status, err = org.GetMembership(users[username].(git.Member))
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
		org.Members[username] = git.NewMemberFromUsername(username)
	}

	// get pending groups
	var group git.Group
	var groupmember git.Member
	pendinggroups := make(map[int]git.Group)
	for groupID, _ := range org.PendingGroup {
		group, err = git.NewGroup(org.Name, groupID)
		if err != nil {
			log.Println(err)
		}
		for key, _ := range group.Members {
			groupmember = git.NewMemberFromUsername(key)
			group.Members[key] = groupmember
		}
		pendinggroups[groupID] = group
	}

	// get groups
	for groupname, _ := range org.Groups {
		groupID, _ := strconv.Atoi(groupname[5:])
		group, _ = git.NewGroup(org.Name, groupID)
		for key, _ := range group.Members {
			groupmember = git.NewMemberFromUsername(key)
			group.Members[key] = groupmember
		}
		org.Groups[groupname] = group
	}

	view := teacherspanelview{
		Member:       member,
		PendingUser:  users,
		Org:          org,
		PendingGroup: pendinggroups,
	}
	execTemplate("teacherspanel.html", w, view)
}

type showresultview struct {
	Member   git.Member
	Org      git.Organization
	Username string
	Labnum   int
	IsGroup  bool
}

func showresulthandler(w http.ResponseWriter, r *http.Request) {
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

	org := git.NewOrganization(orgname)

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
		user := git.NewMemberFromUsername(username)
		nr := user.Courses[org.Name].CurrentLabNum
		if nr >= org.IndividualAssignments {
			labnum = org.IndividualAssignments - 1
		} else {
			labnum = nr - 1
		}
	}

	view := showresultview{
		Member:   member,
		Org:      org,
		Username: username,
		Labnum:   labnum,
		IsGroup:  isgroup,
	}
	execTemplate("teacherresultpage.html", w, view)
}

func addassistanthandler(w http.ResponseWriter, r *http.Request) {
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

	if _, ok := member.Teaching[course]; !ok {
		http.Error(w, "User is not the teacher for this course.", 404)
		return
	}

	if username == member.Username {
		return
	}

	assistant := git.NewMemberFromUsername(username)
	org := git.NewOrganization(course)

	assistant.AddAssistingOrganization(org)
	assistant.StickToSystem()

	org.AddTeacher(assistant)
	if _, ok := org.PendingUser[username]; ok {
		delete(org.PendingUser, username)
	}
	org.StickToSystem()

}
