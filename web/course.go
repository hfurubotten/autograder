package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	ci "github.com/hfurubotten/autograder/ci"
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
			pages.RedirectTo(w, r, pages.SIGNOUT, 307)
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
		pages.RedirectTo(w, r, "/course/new", 307)
		return
	}

	view.Member = &member
	view.Orgs, err = member.ListOrgs()
	if err != nil {
		log.Println(err)
		pages.RedirectTo(w, r, pages.SIGNOUT, 307)
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
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
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

	pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
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
			pages.RedirectTo(w, r, "/course/register", 307)
			return
		}

		orgname = path[3]
	} else {
		pages.RedirectTo(w, r, "/course/register", 307)
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

type teacherspanelview struct {
	Member      git.Member
	Org         git.Organization
	PendingUser map[string]interface{}
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

	view := teacherspanelview{
		Member:      member,
		PendingUser: users,
		Org:         org,
	}
	execTemplate("teacherspanel.html", w, view)
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

	// Checks if the user is signed in and a teacher.
	/*member*/ _, err := checkTeacherApproval(w, r, false)
	if err != nil {
		log.Println(err)

		view.Error = true
		view.ErrorMsg = "You are not singed in or not a teacher."

		enc.Encode(view)
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
		view.Error = true
		view.ErrorMsg = "Username was not set in the request."

		enc.Encode(view)
		return
	}

	org := git.NewOrganization(orgname)

	teams, err := org.ListTeams()
	if err != nil {
		log.Println(err)

		view.Error = true
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

			view.Error = true
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

				view.Error = true
				view.ErrorMsg = "Error communicating with Github. Can't create team."

				enc.Encode(view)
				return
			}

			err = org.AddMemberToTeam(teamID, username)
			if err != nil {
				log.Println(err)

				view.Error = true
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."

				enc.Encode(view)
				return
			}
		} else {
			err = org.LinkRepoToTeam(t.ID, username+"-"+git.STANDARD_REPO_NAME)
			if err != nil {
				log.Println(err)

				view.Error = true
				view.ErrorMsg = "Error communicating with Github. Can't link repo to team."

				enc.Encode(view)
				return
			}

			err = org.AddMemberToTeam(t.ID, username)
			if err != nil {
				log.Println(err)

				view.Error = true
				view.ErrorMsg = "Error communicating with Github. Can't add member to team."

				enc.Encode(view)
				return
			}
		}
	}

	delete(org.PendingUser, username)
	org.Members[username] = nil
	org.StickToSystem()

	view.Error = false
	view.Approved = true
	view.User = username
	enc.Encode(view)
}

type maincourseview struct {
	Member git.Member
	Labnum int
	Org    git.Organization
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
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}

		orgname = path[2]
	} else {
		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
		return
	}

	org := git.NewOrganization(orgname)
	view := maincourseview{
		Member: member,
		Org:    org,
		Labnum: member.Courses[org.Name].CurrentLabNum - 1,
	}
	execTemplate("maincoursepage.html", w, view)
}

type showresultview struct {
	Member   git.Member
	Org      git.Organization
	Username string
	Labnum   int
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

	org := git.NewOrganization(orgname)
	view := showresultview{
		Member:   member,
		Org:      org,
		Username: username,
		Labnum:   member.Courses[org.Name].CurrentLabNum - 1,
	}
	execTemplate("teacherresultpage.html", w, view)
}

func approvelabhandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	_, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	course := r.FormValue("Course")
	username := r.FormValue("User")
	approve := r.FormValue("Approve")
	labnum, err := strconv.Atoi(r.FormValue("Labnum"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}

	if approve != "true" {
		log.Println("Missing approval")
		http.Error(w, "Not approved", 404)
		return
	}

	if !git.HasOrganization(course) || username == "" {
		log.Println("Missing username or uncorrect course")
		http.Error(w, "Unknown", 404)
		return
	}

	var isgroup bool
	if git.HasMember(username) {
		isgroup = false
	} else {
		isgroup = strings.Contains(username, "group")
		if !isgroup {
			log.Println("No user found")
			http.Error(w, "Unknown 2", 404)
			return
		}
	}

	if isgroup {
		gnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
		group, err := git.NewGroup(course, gnum)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
		group.CurrentLabNum = labnum + 1
		group.StickToSystem()
	} else {
		user := git.NewMemberFromUsername(username)
		copt := user.Courses[course]
		copt.CurrentLabNum = labnum + 1
		user.Courses[course] = copt
		user.StickToSystem()
	}
}

func ciresulthandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	_, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), 404)
		log.Println(err)
		return
	}

	// TODO: add more security
	orgname := r.FormValue("Course")
	username := r.FormValue("Username")
	labname := r.FormValue("Labname")

	res, err := ci.GetIntegationResults(orgname, username, labname)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	enc := json.NewEncoder(w)

	err = enc.Encode(res)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}

}
