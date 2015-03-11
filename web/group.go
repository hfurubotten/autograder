package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/hfurubotten/autograder/git"
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
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	orgname := r.FormValue("course")
	if !git.HasOrganization(orgname) {
		http.Error(w, "Does not have organization.", 404)
	}

	org, err := git.NewOrganization(orgname)
	if err != nil {
		http.Error(w, "Does not have organization.", 404)
	}

	org.Lock()
	defer org.Unlock()

	org.PendingRandomGroup[member.Username] = nil
	org.Save()
}

// NewGroupURL is the URL used to call NewGroupHandler.
var NewGroupURL = "/course/newgroup"

// NewGroupHandler is a http handler used when submitting a new group for approval.
func NewGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	newgrouplock.Lock()
	defer newgrouplock.Unlock()

	course := r.FormValue("course")

	if _, ok := member.Courses[course]; !ok {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		log.Println("Unknown course.")
		return
	}

	org, err := git.NewOrganization(course)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err)
		return
	}

	org.Lock()
	defer org.Unlock()

	org.GroupCount = org.GroupCount + 1

	group, err := git.NewGroup(course, org.GroupCount)
	if err != nil {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		log.Println("Couldn't make new group object.")
		return
	}

	group.Lock()
	defer group.Unlock()

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

	var opt git.CourseOptions
	for _, username := range members {
		user, err := git.NewMemberFromUsername(username)
		if err != nil {
			continue
		}

		user.Lock()
		defer user.Unlock()

		opt = user.Courses[course]
		if !opt.IsGroupMember {
			opt.IsGroupMember = true
			opt.GroupNum = org.GroupCount
			user.Courses[course] = opt
			user.Save()
			group.AddMember(username)
		}

		delete(org.PendingRandomGroup, username)
	}

	org.PendingGroup[org.GroupCount] = nil
	org.Save()
	group.Save()

	if member.IsTeacher {
		http.Redirect(w, r, "/course/teacher/"+org.Name+"#groups", 307)
	} else {
		http.Redirect(w, r, "/course/"+org.Name+"#groups", 307)
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
		http.Error(w, err.Error(), 404)
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

	group, err := git.NewGroup(orgname, groupID)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	group.Lock()
	defer group.Unlock()

	if group.Active {
		view.ErrorMsg = "This group is already active."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if len(group.Members) <= 1 {
		view.ErrorMsg = "No members in this group."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	org, err := git.NewOrganization(orgname)
	if err != nil {
		view.ErrorMsg = "Could not retrieve stored organization."
		err = enc.Encode(view)
		if err != nil {
			log.Println(err)
		}
		return
	}

	org.Lock()
	defer org.Unlock()

	if !org.IsTeacher(member) {
		err = enc.Encode(ErrNotTeacher)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if org.GroupAssignments > 0 {
		repo := git.RepositoryOptions{
			Name:     git.GroupRepoPrefix + r.FormValue("groupid"),
			Private:  org.Private,
			AutoInit: true,
			Hook:     true,
		}
		err = org.CreateRepo(repo)
		if err != nil {
			log.Println(err)
			view.ErrorMsg = "Error communicating with Github. Couldn't create repository."
			enc.Encode(view)
			return
		}

		newteam := git.TeamOptions{
			Name:       git.GroupRepoPrefix + r.FormValue("groupid"),
			Permission: git.PushPermission,
			RepoNames:  []string{git.GroupRepoPrefix + r.FormValue("groupid")},
		}

		teamID, err := org.CreateTeam(newteam)
		if err != nil {
			log.Println(err)
			view.ErrorMsg = "Error communicating with Github. Can't create team."
			enc.Encode(view)
			return
		}

		group.TeamID = teamID

		for username := range group.Members {
			err = org.AddMemberToTeam(teamID, username)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."
				enc.Encode(view)
				return
			}
		}
	}

	org.AddGroup(group)
	org.Save()

	group.Activate()
	group.Save()

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
		http.Redirect(w, r, "/", 307)
		log.Println(err)
		return
	}

	groupid, err := strconv.Atoi(r.FormValue("groupid"))
	if err != nil {
		http.Error(w, "Group ID is not a number: "+err.Error(), 404)
		return
	}
	course := r.FormValue("course")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown course.", 404)
		return
	}

	org, err := git.NewOrganization(course)
	if err != nil {
		http.Error(w, "Unknown course.", 404)
		return
	}

	org.Lock()
	defer org.Unlock()

	if !org.IsTeacher(member) {
		http.Error(w, "Is not a teacher or assistant for this course.", 404)
		return
	}

	if !git.HasGroup(org.Name, groupid) {
		groupname := git.GroupRepoPrefix + strconv.Itoa(groupid)
		if _, ok := org.Groups[groupname]; ok {
			delete(org.Groups, groupname)
			org.Save()
			return
		}

		http.Error(w, "Unknown group.", 404)
		return
	}

	if _, ok := org.PendingGroup[groupid]; ok {
		delete(org.PendingGroup, groupid)
	}

	groupname := git.GroupRepoPrefix + strconv.Itoa(groupid)
	if _, ok := org.Groups[groupname]; ok {
		delete(org.Groups, groupname)
	}

	group, err := git.NewGroup(org.Name, groupid)
	if err != nil {
		http.Error(w, "Could not get the group: "+err.Error(), 404)
		return
	}

	group.Delete()
	org.Save()
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

	if !git.HasGroup(orgname, groupid) {
		err = enc.Encode(ErrUnknownGroup)
		return
	}

	org, err := git.NewOrganization(orgname)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		return
	}

	org.Lock()
	defer org.Unlock()

	if !org.IsTeacher(member) {
		err = enc.Encode(ErrNotTeacher)
		return
	}

	group, err := git.NewGroup(orgname, groupid)
	if err != nil {
		view.ErrorMsg = err.Error()
		err = enc.Encode(view)
		return
	}

	group.Lock()
	defer group.Unlock()

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
	group.Save()
	org.Save()

	view.Added = true
	view.Error = false
	enc.Encode(view)
}
