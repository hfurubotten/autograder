package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
)

var (
	newgrouplock sync.Mutex
)

// RequestRandomGroupURL is the URL used to call RequestRandomGroupHandler.
var RequestRandomGroupURL = "/course/requestrandomgroup"

// RequestRandomGroupHandler is a http handler used by a student to request a random group assignment.
func RequestRandomGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	orgname := r.FormValue("course")
	if !git.HasOrganization(orgname) {
		http.Error(w, "Does not have organization.", http.StatusNotFound)
	}

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		http.Error(w, "Does not have organization.", http.StatusNotFound)
	}
	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	org.PendingRandomGroup[member.Username] = nil
}

// NewGroupURL is the URL used to call NewGroupHandler.
var NewGroupURL = "/course/newgroup"

// NewGroupHandler is a http handler used when submitting a new group for approval.
func NewGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	newgrouplock.Lock()
	defer newgrouplock.Unlock()

	course := r.FormValue("course")

	if _, ok := member.Courses[course]; !ok {
		http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
		log.Println("Unknown course.")
		return
	}

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	//org.GroupCount = org.GroupCount + 1

	gid := git.GetNextGroupID()
	if gid < 0 {
		http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
		log.Println("Error while getting next group ID.")
		return
	}

	group, err := git.NewGroup(course, gid, false)
	if err != nil {
		// couldn't make new group object
		logErrorAndRedirect(w, r, pages.Front, err)
		return
	}

	defer func() {
		err := group.Save()
		if err != nil {
			group.Unlock()
			log.Println(err)
		}
	}()

	r.ParseForm()
	members := r.PostForm["member"]

	if !org.IsTeacher(member) {
		var found bool
		for _, u := range members {
			if u == member.Username {
				found = true
			}
		}
		if !found {
			members = append(members, member.Username)
		}
	}

	var opt git.Course
	for _, username := range members {
		user, err := git.GetMember(username)
		if err != nil {
			log.Println(err)
			continue
		}

		opt = user.Courses[course]
		if !opt.IsGroupMember {
			user.Lock()
			opt.IsGroupMember = true
			opt.GroupNum = group.ID
			user.Courses[course] = opt
			err := user.Save()
			if err != nil {
				user.Unlock()
				log.Println(err)
			}
			group.AddMember(username)
		}

		delete(org.PendingRandomGroup, username)
	}

	org.PendingGroup[group.ID] = nil

	if member.IsTeacher {
		http.Redirect(w, r, "/course/teacher/"+org.Name+"#groups", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/course/"+org.Name+"#groups", http.StatusTemporaryRedirect)
	}
}

// ApproveGroupView is the view used to reply JSON data back when using ApproveGroupHandler.
type ApproveGroupView struct {
	JSONErrorMsg
	Approved bool
	ID       int
}

// ApproveGroupUrl is the URL used to call ApproveGroupHandler.
var ApproveGroupURL = "/course/approvegroup"

// ApproveGroupHandler is a http handler used by teachers to approve a group and activate it.
func ApproveGroupHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	view := ApproveGroupView{}
	view.Error = true
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	groupID, err := strconv.Atoi(r.FormValue("groupid"))
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	orgname := r.FormValue("course")

	group, err := git.NewGroup(orgname, groupID, false)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	defer func() {
		err := group.Save()
		if err != nil {
			group.Unlock()
			log.Println(err)
		}
	}()

	if group.Active {
		view.ErrorMsg = "This group is already active."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if len(group.Members) < 1 {
		view.ErrorMsg = "No members in this group."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		view.ErrorMsg = "Could not retrieve stored organization."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		err = enc.Encode(ErrNotTeacher)
		if err != nil {
			log.Println(err)
		}
		return
	}

	orgrepos, err := org.ListRepos()
	if err != nil {
		log.Println(err)
		view.ErrorMsg = err.Error()
		enc.Encode(view)
	}

	orgteams, err := org.ListTeams()
	if err != nil {
		log.Println(err)
		view.ErrorMsg = err.Error()
		enc.Encode(view)
	}

	if org.GroupAssignments > 0 {
		repo := git.RepositoryOptions{
			Name:     git.GroupRepoPrefix + r.FormValue("groupid"),
			Private:  org.Private,
			AutoInit: true,
			Issues:   true,
			Hook:     "*",
		}
		if _, ok := orgrepos[repo.Name]; !ok {
			err = org.CreateRepo(repo)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Couldn't create repository."
				enc.Encode(view)
				return
			}
		}

		newteam := git.TeamOptions{
			Name:       git.GroupRepoPrefix + r.FormValue("groupid"),
			Permission: git.PushPermission,
			RepoNames:  []string{git.GroupRepoPrefix + r.FormValue("groupid")},
		}

		var teamID int
		if team, ok := orgteams[newteam.Name]; ok {
			teamID = team.ID
		} else {
			teamID, err = org.CreateTeam(newteam)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't create team."
				enc.Encode(view)
				return
			}
		}

		group.TeamID = teamID

		for username := range group.Members {
			err = org.AddMemberToTeam(teamID, username)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."
				enc.Encode(view)
				continue
			}
		}
	}

	org.AddGroup(group)

	group.Activate()

	view.Error = false
	view.Approved = true
	view.ID = groupID
	err = enc.Encode(view)
	if err != nil {
		log.Println(err)
	}
}

// RemovePendingGroupURL is the URL used to call RemovePendingGroupHandler.
var RemovePendingGroupURL = "/course/removegroup"

// RemovePendingGroupHandler is used to remove a group.
func RemovePendingGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		logErrorAndRedirect(w, r, pages.Front, err)
		return
	}

	groupid, err := strconv.Atoi(r.FormValue("groupid"))
	if err != nil {
		http.Error(w, "Group ID is not a number: "+err.Error(), http.StatusNotFound)
		return
	}
	course := r.FormValue("course")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown course.", http.StatusNotFound)
		return
	}

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, "Unknown course.", http.StatusNotFound)
		return
	}

	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		http.Error(w, "Is not a teacher or assistant for this course.", http.StatusNotFound)
		return
	}

	if !git.HasGroup(groupid) {
		groupname := git.GroupRepoPrefix + strconv.Itoa(groupid)
		if _, ok := org.Groups[groupname]; ok {
			delete(org.Groups, groupname)
			return
		}

		http.Error(w, "Unknown group.", http.StatusNotFound)
		return
	}

	if _, ok := org.PendingGroup[groupid]; ok {
		delete(org.PendingGroup, groupid)
	}

	groupname := git.GroupRepoPrefix + strconv.Itoa(groupid)
	if _, ok := org.Groups[groupname]; ok {
		delete(org.Groups, groupname)
	}

	group, err := git.NewGroup(org.Name, groupid, false)
	if err != nil {
		http.Error(w, "Could not get the group: "+err.Error(), http.StatusNotFound)
		return
	}

	group.Delete()
}

// AddGroupMemberView is the view used to give a JSON reply to AddGroupMemberHandler.
type AddGroupMemberView struct {
	JSONErrorMsg
	Added bool `json:"Added"`
}

// AddGroupMemberURL is the URL used to call AddGroupMemberHandler.
var AddGroupMemberURL = "/group/addmember"

// AddGroupMemberHandler is a http handler adding an additional member to a active group.
func AddGroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	view := AddGroupMemberView{}
	view.Error = true
	enc := json.NewEncoder(w)

	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		err = enc.Encode(ErrSignIn)
		return
	}

	orgname := r.FormValue("course")
	if orgname == "" || !git.HasOrganization(orgname) {
		err = enc.Encode(ErrUnknownCourse)
		return
	}

	groupid, err := strconv.Atoi(r.FormValue("groupid"))
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		return
	}

	if !git.HasGroup(groupid) {
		err = enc.Encode(ErrUnknownGroup)
		return
	}

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		return
	}

	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if !org.IsTeacher(member) {
		err = enc.Encode(ErrNotTeacher)
		return
	}

	group, err := git.NewGroup(orgname, groupid, false)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		return
	}

	defer func() {
		err := group.Save()
		if err != nil {
			group.Unlock()
			log.Println(err)
		}
	}()

	if group.TeamID == 0 {
		teams, err := org.ListTeams()
		if err != nil {
			view.ErrorMsg = err.Error()
			err = enc.Encode(view)
			return
		}

		if team, ok := teams[git.GroupRepoPrefix+strconv.Itoa(groupid)]; ok {
			group.TeamID = team.ID
		} else {
			view.ErrorMsg = "Error finding team on GitHub."
			err = enc.Encode(view)
			return
		}
	}

	r.ParseForm()
	members := r.PostForm["member"]

	for _, username := range members {
		if username == "" || !git.HasMember(username) {
			continue
		}

		group.AddMember(username)

		org.AddMemberToTeam(group.TeamID, username)
		delete(org.PendingRandomGroup, username)
	}

	group.Activate()

	view.Added = true
	view.Error = false
	enc.Encode(view)
}
