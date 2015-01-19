package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
)

type courseview struct {
	Member *git.Member
	Org    string
	Orgs   []string
}

func newcoursehandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := courseview{}

	view.Member = &member

	var page string
	switch r.URL.Path {
	case "/course/new":
		page = "newcourse-info.html"
	case "/course/new/org":
		page = "newcourse-orgselect.html"

		view.Orgs, err = member.ListOrgs()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, pages.SIGNOUT, 307)
			return
		}
	}
	execTemplate(page, w, view)
}

func selectorghandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := courseview{}

	if path := strings.Split(r.URL.Path, "/"); len(path) == 5 {
		view.Org = path[4]
	} else {
		http.Redirect(w, r, "/course/new", 307)
		return
	}

	view.Member = &member
	view.Orgs, err = member.ListOrgs()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, pages.SIGNOUT, 307)
		return
	}

	execTemplate("newcourse-register.html", w, view)
}

func saveorghandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	org := git.NewOrganization(r.FormValue("org"))
	org.AdminToken = member.GetToken()
	org.Private = r.FormValue("private") == "on"
	org.Description = r.FormValue("desc")
	groups, err := strconv.Atoi(r.FormValue("groups"))
	if err != nil {
		log.Println("Cannot convert number of groups assignments from string to int: ", err)
		groups = 0
	}
	org.GroupAssignments = groups
	indv, err := strconv.Atoi(r.FormValue("indv"))
	if err != nil {
		log.Println("Cannot convert number of individual assignments from string to int: ", err)
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}
	org.IndividualAssignments = indv

	if r.FormValue("template") == "" {
		repo := git.RepositoryOptions{
			Name:     git.COURSE_INFO_NAME,
			Private:  false,
			AutoInit: true,
			Hook:     false,
		}
		err = org.CreateRepo(repo)
		if err != nil {
			log.Println(err)
			return
		}

		repo = git.RepositoryOptions{
			Name:     git.STANDARD_REPO_NAME,
			Private:  org.Private,
			AutoInit: true,
			Hook:     false,
		}
		err = org.CreateRepo(repo)
		if err != nil {
			log.Println(err)
			return
		}

		repo = git.RepositoryOptions{
			Name:     git.TEST_REPO_NAME,
			Private:  org.Private,
			AutoInit: true,
			Hook:     false,
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
			}
			err = org.CreateFile(git.TEST_REPO_NAME, path, content, commitmessage)
			content = "# Lab assignment " + strconv.Itoa(i+1) + " test"
			if err != nil {
				log.Println(err)
			}
		}

		if org.GroupAssignments > 0 {
			repo.Name = git.GROUPS_REPO_NAME
			err = org.CreateRepo(repo)
			if err != nil {
				log.Println(err)
				return
			}
			repo = git.RepositoryOptions{
				Name:     git.GROUPTEST_REPO_NAME,
				Private:  org.Private,
				AutoInit: true,
			}
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
				content = "# Group assignment " + strconv.Itoa(i+1) + " tests"
				err = org.CreateFile(git.GROUPTEST_REPO_NAME, path, content, commitmessage)
				if err != nil {
					log.Println(err)
				}
			}
		}

	} else {
		var repo git.RepositoryOptions

		// Tries to fork the course-info repo, if it fails it will create a blank one.
		err = org.Fork(r.FormValue("template"), git.COURSE_INFO_NAME)
		if err != nil {
			repo = git.RepositoryOptions{
				Name:     git.COURSE_INFO_NAME,
				Private:  org.Private,
				AutoInit: true,
				Hook:     false,
			}
			err = org.CreateRepo(repo)
			if err != nil {
				log.Println(err)
				return
			}
		}

		// Tries to fork the labs repo, if it fails it will create a blank one.
		err = org.Fork(r.FormValue("template"), git.STANDARD_REPO_NAME)
		if err != nil {
			repo = git.RepositoryOptions{
				Name:     git.STANDARD_REPO_NAME,
				Private:  org.Private,
				AutoInit: true,
				Hook:     false,
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
				}
			}
		}

		// Tries to fork the test-labs repo, if it fails it will create a blank one.
		err = org.Fork(r.FormValue("template"), git.TEST_REPO_NAME)
		if err != nil {
			repo = git.RepositoryOptions{
				Name:     git.TEST_REPO_NAME,
				Private:  org.Private,
				AutoInit: true,
				Hook:     false,
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
				err = org.CreateFile(git.TEST_REPO_NAME, path, content, commitmessage)
				if err != nil {
					log.Println(err)
				}
			}
		}

		if org.GroupAssignments > 0 {
			err = org.Fork(r.FormValue("template"), git.GROUPS_REPO_NAME)
			if err != nil {
				repo = git.RepositoryOptions{
					Name:     git.GROUPS_REPO_NAME,
					Private:  org.Private,
					AutoInit: true,
					Hook:     false,
				}
				err = org.CreateRepo(repo)
				if err != nil {
					log.Println(err)
					return
				}

				for i := 0; i < org.IndividualAssignments; i++ {
					path := "lab" + strconv.Itoa(i+1) + "/README.md"
					commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
					content := "# Group lab assignment " + strconv.Itoa(i+1)
					err = org.CreateFile(git.GROUPS_REPO_NAME, path, content, commitmessage)
					if err != nil {
						log.Println(err)
					}
				}
			}

			err = org.Fork(r.FormValue("template"), git.GROUPTEST_REPO_NAME)
			if err != nil {
				repo = git.RepositoryOptions{
					Name:     git.GROUPTEST_REPO_NAME,
					Private:  org.Private,
					AutoInit: true,
					Hook:     false,
				}
				err = org.CreateRepo(repo)
				if err != nil {
					log.Println(err)
					return
				}

				for i := 0; i < org.IndividualAssignments; i++ {
					path := "lab" + strconv.Itoa(i+1) + "/README.md"
					commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
					content := "# Group lab assignment " + strconv.Itoa(i+1)
					err = org.CreateFile(git.GROUPTEST_REPO_NAME, path, content, commitmessage)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	// Creates the student team
	repos := make([]string, 0)
	repos = append(repos, git.STANDARD_REPO_NAME, git.COURSE_INFO_NAME)
	if org.GroupAssignments > 0 {
		repos = append(repos, git.GROUPS_REPO_NAME)
	}

	team := git.TeamOptions{
		Name:       "students",
		Permission: git.PERMISSION_PULL,
		RepoNames:  repos,
	}
	org.StudentTeamID, err = org.CreateTeam(team)
	if err != nil {
		log.Println(err)
	}

	org.AddTeacher(member)

	// Saved the new organization info
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

	http.Redirect(w, r, pages.FRONTPAGE, 307)
}

type newmemberview struct {
	Member *git.Member
	Orgs   []git.Organization
	Org    string
}

func newcoursememberhandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := newmemberview{
		Member: &member,
		Orgs:   git.ListRegisteredOrganizations(),
	}
	execTemplate("course-registermember.html", w, view)
}

func registercoursememberhandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 4 {
		if !git.HasOrganization(path[3]) {
			http.Redirect(w, r, "/course/register", 307)
			return
		}

		orgname = path[3]
	} else {
		http.Redirect(w, r, "/course/register", 307)
		return
	}

	org := git.NewOrganization(orgname)

	err = org.AddMembership(member)
	if err != nil {
		log.Println(err)
	}

	err = org.StickToSystem()
	if err != nil {
		log.Println(err)
	}

	view := newmemberview{
		Member: &member,
		Org:    orgname,
	}
	execTemplate("course-registeredmemberinfo.html", w, view)
}

type approvemembershipview struct {
	Error    bool
	ErrorMsg string
	Approved bool
	User     string
}

func approvecoursemembershiphandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	view := approvemembershipview{}
	view.Error = true // default is an error; if its not we anyway set it to false before encoding

	// Checks if the user is signed in and a teacher.
	/*member*/ _, err := checkTeacherApproval(w, r, false)
	if err != nil {
		log.Println(err)
		view.ErrorMsg = "You are not singed in or not a teacher."
		enc.Encode(view)
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
		view.ErrorMsg = "Username was not set in the request."
		enc.Encode(view)
		return
	}

	org := git.NewOrganization(orgname)
	teams, err := org.ListTeams()
	if err != nil {
		log.Println(err)
		view.ErrorMsg = "Error communicating with Github. Can't get list teams."
		enc.Encode(view)
		return
	}

	if org.IndividualAssignments > 0 {
		repo := git.RepositoryOptions{
			Name:     username + "-" + git.STANDARD_REPO_NAME,
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

		if t, ok := teams[username]; !ok {
			newteam := git.TeamOptions{
				Name:       username,
				Permission: git.PERMISSION_PUSH,
				RepoNames:  []string{username + "-" + git.STANDARD_REPO_NAME},
			}

			teamID, err := org.CreateTeam(newteam)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't create team."
				enc.Encode(view)
				return
			}

			err = org.AddMemberToTeam(teamID, username)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."
				enc.Encode(view)
				return
			}
		} else {
			err = org.LinkRepoToTeam(t.ID, username+"-"+git.STANDARD_REPO_NAME)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't link repo to team."
				enc.Encode(view)
				return
			}

			err = org.AddMemberToTeam(t.ID, username)
			if err != nil {
				log.Println(err)
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."
				enc.Encode(view)
				return
			}
		}
	}

	delete(org.PendingUser, username)
	org.Members[username] = nil
	org.StickToSystem()

	member := git.NewMemberFromUsername(username)
	member.AddOrganization(org)
	err = member.StickToSystem()

	view.Error = false // it wasn't an error after all
	view.Approved = true
	view.User = username
	enc.Encode(view)
}

type maincourseview struct {
	Member      *git.Member
	Group       *git.Group
	Labnum      int
	GroupLabnum int
	Org         *git.Organization
}

func maincoursepagehandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	// Gets the org and check if valid
	orgname := ""
	if path := strings.Split(r.URL.Path, "/"); len(path) == 3 {
		if !git.HasOrganization(path[2]) {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[2]
	} else {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	org := git.NewOrganization(orgname)
	view := maincourseview{
		Member: &member,
		Org:    &org,
	}

	nr := member.Courses[org.Name].CurrentLabNum
	if nr >= org.IndividualAssignments {
		view.Labnum = org.IndividualAssignments - 1
	} else {
		view.Labnum = nr - 1
	}

	if member.Courses[orgname].IsGroupMember {
		group, err := git.NewGroup(orgname, member.Courses[orgname].GroupNum)
		if err != nil {
			log.Println(err)
			return
		}
		view.Group = &group
		if group.CurrentLabNum >= org.GroupAssignments {
			view.GroupLabnum = org.GroupAssignments - 1
		} else {
			view.GroupLabnum = group.CurrentLabNum - 1
		}
	}

	execTemplate("maincoursepage.html", w, view)
}

type showresultview struct {
	Member   git.Member
	Org      git.Organization
	Username string
	Labnum   int
	IsGroup  bool
}

func updatecoursehandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		log.Println(err)
		return
	}

	orgname := r.FormValue("org")

	if _, ok := member.Teaching[orgname]; !ok {
		http.Error(w, "Not valid organization.", 404)
		return
	}

	org := git.NewOrganization(orgname)

	indv, err := strconv.Atoi(r.FormValue("indv"))
	if err != nil {
		http.Error(w, "Cant use the individual assignment format.", 415)
		return
	}
	org.IndividualAssignments = indv

	groups, err := strconv.Atoi(r.FormValue("groups"))
	if err != nil {
		http.Error(w, "Cant use the group assignment format.", 415)
		return
	}
	org.GroupAssignments = groups

	org.Description = r.FormValue("desc")
	org.Private = r.FormValue("private") == "on"

	var fname string
	var fkey string
	for i := 1; i <= indv; i = i + 1 {
		fkey = "lab" + strconv.Itoa(i)
		fname = r.FormValue(fkey)
		if fname == "" {
			fname = fkey
		}

		org.IndividualLabFolders[i] = fname
	}

	for i := 1; i <= groups; i = i + 1 {
		fkey = "group" + strconv.Itoa(i)
		fname = r.FormValue(fkey)
		if fname == "" {
			fname = fkey
		}

		org.GroupLabFolders[i] = fname
	}

	org.StickToSystem()

	http.Redirect(w, r, "/course/teacher/"+org.Name, 307)
}

func removependinguserhandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		http.Redirect(w, r, "/", 307)
		log.Println(err)
		return
	}

	username := r.FormValue("user")
	course := r.FormValue("course")

	if !git.HasOrganization(course) {
		http.Error(w, "Unknown course.", 404)
		return
	}

	org := git.NewOrganization(course)

	if !org.IsTeacher(member) {
		http.Error(w, "Is not a teacher or assistant for this course.", 404)
		return
	}

	if _, ok := org.PendingUser[username]; ok {
		delete(org.PendingUser, username)
		org.StickToSystem()
	}
}
