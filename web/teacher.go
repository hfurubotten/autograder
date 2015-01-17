package web

import (
	"net/http"
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
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 4 {
		if !git.HasOrganization(path[3]) {
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[3]
	} else {
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	org := git.NewOrganization(orgname)

	users := org.PendingUser

	repos, err := org.ListRepos()
	if err != nil {
		log.Println("Couldn't get all the repos in the organization. ")
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	// gets pending users
	var status string
	for username, _ := range users {
		// TODO: check status up against Github
		users[username] = git.NewMemberFromUsername(username)
		status, err = org.GetMembership(users[username].(git.Member))

		if err != nil {
			log.Println(err)
			continue
		}

		if status == "active" {
			if _, ok := repos[username+"-"+git.STANDARD_REPO_NAME]; !ok && org.IndividualAssignments > 0 {
				continue
			} else {
				delete(users, username)
			}
			// TODO: what about group assignments?
		} else if status == "pending" {
			delete(users, username)
		} else {
			delete(users, username)
			log.Println("Got a unexpected status back from Github regarding Membership")
		}
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

	view := teacherspanelview{
		Member:       member,
		PendingUser:  users,
		Org:          org,
		PendingGroup: pendinggroups,
	}
	execTemplate("teacherspanel.html", w, view)
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
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[3]
	} else {
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	username := r.FormValue("user")
	if username == "" {
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	if !git.HasOrganization(orgname) {
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	org := git.NewOrganization(orgname)

	isgroup := false
	labnum := 0
	if !git.HasMember(username) {
		groupnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}
		if git.HasGroup(org.Name, groupnum) {
			isgroup = true
			group, err := git.NewGroup(org.Name, groupnum)
			if err != nil {
				pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
				return
			}
			if group.CurrentLabNum >= org.GroupAssignments {
				labnum = org.GroupAssignments - 1
			} else {
				labnum = group.CurrentLabNum - 1
			}
		} else {
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}
	} else {
		nr := member.Courses[org.Name].CurrentLabNum
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
