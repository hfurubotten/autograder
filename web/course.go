package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
)

// NewCourseInfoURL is the URL used to call the information page of NewCourseHandler.
var NewCourseInfoURL = "/course/new"

// NewCourseURL is the URL used to call the organization selection page of NewCourseHandler.
var NewCourseURL = "/course/new/org"

// CourseView is the struct sent to the html template compiler.
type CourseView struct {
	StdTemplate
	Org  string
	Orgs []string
}

// NewCourseHandler is a http hander giving a information page for
// teachers when they want to create a new course in autograder.
func NewCourseHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := CourseView{}

	view.Member = member

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

// SelectOrgURL is the URL used to call SelectOrgHandler.
var SelectOrgURL = "/course/new/org/"

// SelectOrgHandler is a http hander giving a page for selecting
// the organization to use for the new course.
func SelectOrgHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := CourseView{}

	if path := strings.Split(r.URL.Path, "/"); len(path) == 5 {
		view.Org = path[4]
	} else {
		http.Redirect(w, r, "/course/new", 307)
		return
	}

	view.Member = member
	view.Orgs, err = member.ListOrgs()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, pages.SIGNOUT, 307)
		return
	}

	execTemplate("newcourse-register.html", w, view)
}

// CreateOrgURL is the URL used to call CreateOrgHandler.
var CreateOrgURL = "/course/create"

// CreateOrgHandler is a http handler which will link a new course
// to a github organization. This function will make a new course
// in autograder and then create all needed repositories on github.
//
// Expected input: org, desc, groups, indv
// Optional input: private, template
func CreateOrgHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	member.Lock()

	org, err := git.NewOrganization(r.FormValue("org"), false)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		// Saved the new organization info
		err = org.Save()
		if err != nil {
			org.Unlock()
			member.Unlock()
			log.Println(err)
			return
		}

		err = member.Save()
		if err != nil {
			member.Unlock()
			log.Println(err)
			return
		}
	}()

	log.Println("Creating Course")

	org.AdminToken = member.GetToken()
	org.Private = r.FormValue("private") == "on"
	org.ScreenName = org.Name
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

	currepos, err := org.ListRepos()
	if err != nil {
		log.Println("Problem listing repos in the new course organization: ", err)
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	templaterepos := make(map[string]git.Repo)
	if r.FormValue("template") != "" {
		templateorg, _ := git.NewOrganization(r.FormValue("template"), true)
		templaterepos, err = templateorg.ListRepos()
		if err != nil {
			log.Println("Problem listing repos in the template organization: ", err)
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}
	}

	// creates the course info repo
	if _, ok := currepos[git.CourseInfoName]; !ok {
		if _, ok = templaterepos[git.CourseInfoName]; ok {
			err = org.Fork(r.FormValue("template"), git.CourseInfoName)
			if err != nil {
				log.Println("Couldn't fork the course info repo: ", err)
				http.Redirect(w, r, pages.HOMEPAGE, 307)
				return
			}
		} else {
			repo := git.RepositoryOptions{
				Name:     git.CourseInfoName,
				Private:  false,
				AutoInit: true,
			}
			err = org.CreateRepo(repo)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	log.Println("Created ", org.Name, "/", git.CourseInfoName)

	// creates the lab assignment repo
	labsl := make(chan int, 1)
	if _, ok := currepos[git.StandardRepoName]; !ok {
		go func(l chan int) {
			defer func() {
				log.Println("Created ", org.Name, "/", git.StandardRepoName)
				l <- 1
			}()
			if _, ok = templaterepos[git.StandardRepoName]; ok {
				err = org.Fork(r.FormValue("template"), git.StandardRepoName)
				if err != nil {
					log.Println("Couldn't fork the individual assignment repo: ", err)
					http.Redirect(w, r, pages.HOMEPAGE, 307)
					return
				}
			} else {
				repo := git.RepositoryOptions{
					Name:     git.StandardRepoName,
					Private:  org.Private,
					AutoInit: true,
					Issues:   true,
				}
				err = org.CreateRepo(repo)
				if err != nil {
					log.Println(err)
					return
				}

				_, err = org.CreateFile(git.StandardRepoName, ".gitignore", git.IgnoreFileContent, "Standard .gitignore file")
				if err != nil {
					log.Println(err)
				}

				for i := 0; i < org.IndividualAssignments; i++ {
					path := "lab" + strconv.Itoa(i+1) + "/README.md"
					commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
					content := "# Lab assignment " + strconv.Itoa(i+1)
					_, err = org.CreateFile(git.StandardRepoName, path, content, commitmessage)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}(labsl)
	} else {
		labsl <- 1
	}

	// creates test repo
	testl := make(chan int, 1)
	if _, ok := currepos[git.TestRepoName]; !ok {
		go func(l chan int) {
			defer func() {
				log.Println("Created ", org.Name, "/", git.TestRepoName)
				l <- 1
			}()
			if _, ok = templaterepos[git.TestRepoName]; ok {
				err = org.Fork(r.FormValue("template"), git.TestRepoName)
				if err != nil {
					log.Println("Couldn't fork the test repo: ", err)
					http.Redirect(w, r, pages.HOMEPAGE, 307)
					return
				}
			} else {
				repo := git.RepositoryOptions{
					Name:     git.TestRepoName,
					Private:  org.Private,
					AutoInit: true,
					Issues:   true,
					//Hook:     "push", // TODO: uncomment when CI rebuilds all on new test.
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
					_, err = org.CreateFile(git.TestRepoName, path, content, commitmessage)
					content = "# Lab assignment " + strconv.Itoa(i+1) + " test"
					if err != nil {
						log.Println(err)
					}
				}

				for i := 0; i < org.GroupAssignments; i++ {
					path := "grouplab" + strconv.Itoa(i+1) + "/README.md"
					commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
					content := "# Lab assignment " + strconv.Itoa(i+1)
					_, err = org.CreateFile(git.TestRepoName, path, content, commitmessage)
					content = "# Lab assignment " + strconv.Itoa(i+1) + " test"
					if err != nil {
						log.Println(err)
					}
				}
			}
		}(testl)
	} else {
		testl <- 1
	}

	// creates the group assignment repo, if number of assignments are larger than 0.
	glabsl := make(chan int)
	if org.GroupAssignments > 0 {
		if _, ok := currepos[git.GroupsRepoName]; !ok {
			go func(l chan int) {
				defer func() {
					log.Println("Created ", org.Name, "/", git.GroupsRepoName)
					l <- 1
				}()
				if _, ok = templaterepos[git.GroupsRepoName]; ok {
					err = org.Fork(r.FormValue("template"), git.GroupsRepoName)
					if err != nil {
						log.Println("Couldn't fork the group assignment repo: ", err)
						http.Redirect(w, r, pages.HOMEPAGE, 307)
						return
					}
				} else {
					repo := git.RepositoryOptions{
						Name:     git.GroupsRepoName,
						Private:  org.Private,
						AutoInit: true,
						Issues:   true,
					}
					err = org.CreateRepo(repo)
					if err != nil {
						log.Println(err)
						return
					}

					_, err = org.CreateFile(git.GroupsRepoName, ".gitignore", git.IgnoreFileContent, "Standard .gitignore file")
					if err != nil {
						log.Println(err)
					}

					for i := 0; i < org.GroupAssignments; i++ {
						path := "grouplab" + strconv.Itoa(i+1) + "/README.md"
						commitmessage := "Adding readme file for lab assignment " + strconv.Itoa(i+1)
						content := "# Lab assignment " + strconv.Itoa(i+1)
						_, err = org.CreateFile(git.GroupsRepoName, path, content, commitmessage)
						content = "# Lab assignment " + strconv.Itoa(i+1) + " test"
						if err != nil {
							log.Println(err)
						}
					}
				}
			}(glabsl)
		} else {
			glabsl <- 1
		}
	} else {
		glabsl <- 1
	}

	// wait on github completion of repos
	// TODO: fix correct channel use further up.
	<-labsl
	<-testl
	<-glabsl

	// Creates the student team
	// TODO: put this in a seperate go rutine and check if the team exsists already.
	var repos []string
	repos = append(repos, git.StandardRepoName, git.CourseInfoName)
	if org.GroupAssignments > 0 {
		repos = append(repos, git.GroupsRepoName)
	}

	team := git.TeamOptions{
		Name:       "students",
		Permission: git.PullPermission,
		RepoNames:  repos,
	}
	org.StudentTeamID, err = org.CreateTeam(team)
	if err != nil {
		log.Println(err)
	}

	log.Println("Created team ", org.Name, "/", "students")

	org.AddTeacher(member)

	member.AddTeachingOrganization(org)

	http.Redirect(w, r, pages.FRONTPAGE, 307)
}

// NewCourseMemberURL is the URL used to call NewCourseMemberHandler.
var NewCourseMemberURL = "/course/register"

// NewMemberView is the struct passed to the html template compiler in NewCourseMemberHandler and RegisterCourseMemberHandler.
type NewMemberView struct {
	StdTemplate
	Orgs []*git.Organization
	Org  string
}

// NewCourseMemberHandler is a http handler which gives a page where
// students can sign up for a course in autograder.
func NewCourseMemberHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in.
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}

	view := NewMemberView{
		StdTemplate: StdTemplate{
			Member: member,
		},
		Orgs: git.ListRegisteredOrganizations(),
	}
	execTemplate("course-registermember.html", w, view)
}

// RegisterCourseMemberURL is the URL used to call RegisterCourseMemberURL.
var RegisterCourseMemberURL = "/course/register/"

// RegisterCourseMemberHandler is a http handler which register new students
// signing up for a course. After registering the student, this handler
// gives back a informal page about how to accept the invitation to the
// organization on github.
func RegisterCourseMemberHandler(w http.ResponseWriter, r *http.Request) {
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

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		http.Redirect(w, r, "/course/register", 307)
		return
	}

	defer func() {
		err = org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	if _, ok := org.Members[member.Username]; ok {
		http.Redirect(w, r, "/course/"+orgname, 307)
		return
	}

	err = org.AddMembership(member)
	if err != nil {
		log.Println("Error adding the student to course. Error msg:", err)
	}

	view := NewMemberView{
		StdTemplate: StdTemplate{
			Member: member,
		},
		Org: orgname,
	}
	execTemplate("course-registeredmemberinfo.html", w, view)
}

// ApproveCourseMembershipURL is the URL used to call ApproveCourseMembershipHandler.
var ApproveCourseMembershipURL = "/course/approvemember/"

// ApproveMembershipView represents the view sent back in the JSON reply in ApproveCourseMembershipHandler.
type ApproveMembershipView struct {
	Error    bool
	ErrorMsg string
	Approved bool
	User     string
}

// ApproveCourseMembershipHandler is a http handler used when a teacher wants
// to accept a student for a course in autograder. This handler will link the
// student to the course organization on github and also create all the needed
// repositories on github.
func ApproveCourseMembershipHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	view := ApproveMembershipView{}
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

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		view.ErrorMsg = "Could not retrieve the stored organization."
		enc.Encode(view)
		return
	}
	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	teams, err := org.ListTeams()
	if err != nil {
		log.Println(err)
		view.ErrorMsg = "Error communicating with Github. Can't get list teams."
		enc.Encode(view)
		return
	}

	if org.IndividualAssignments > 0 {
		repo := git.RepositoryOptions{
			Name:     username + "-" + git.StandardRepoName,
			Private:  org.Private,
			AutoInit: true,
			Issues:   true,
			Hook:     "*",
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
				Permission: git.PushPermission,
				RepoNames:  []string{username + "-" + git.StandardRepoName},
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
			err = org.LinkRepoToTeam(t.ID, username+"-"+git.StandardRepoName)
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

	member, err := git.NewMemberFromUsername(username, false)
	if err != nil {
		view.ErrorMsg = "Could not retrieve the stored user."
		enc.Encode(view)
		return
	}
	defer func() {
		err = member.Save()
		if err != nil {
			member.Unlock()
			log.Println(err)
		}
	}()

	member.AddOrganization(org)

	view.Error = false // it wasn't an error after all
	view.Approved = true
	view.User = username
	enc.Encode(view)
}

// UserCoursePageURL is the URL used to call UserCoursePageHandler
var UserCoursePageURL = "/course/"

// MainCourseView is the struct sent to the html template compiler in UserCoursePageHandler.
type MainCourseView struct {
	StdTemplate
	Group       *git.Group
	Labnum      int
	GroupLabnum int
	Org         *git.Organization
}

// UserCoursePageHandler is a http handler giving back the main user
// page for a course. This page gived information about all the labs
// and results for a user. A user can also submit code reviews from
// this page.
func UserCoursePageHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in.
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

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		http.Redirect(w, r, pages.HOMEPAGE, 307)
		return
	}

	view := MainCourseView{
		StdTemplate: StdTemplate{
			Member: member,
		},
		Org: org,
	}

	nr := member.Courses[org.Name].CurrentLabNum
	if nr >= org.IndividualAssignments {
		view.Labnum = org.IndividualAssignments - 1
	} else {
		view.Labnum = nr - 1
	}

	if member.Courses[orgname].IsGroupMember {
		group, err := git.NewGroup(orgname, member.Courses[orgname].GroupNum, true)
		if err != nil {
			log.Println(err)
			return
		}
		view.Group = group
		if group.CurrentLabNum >= org.GroupAssignments {
			view.GroupLabnum = org.GroupAssignments - 1
		} else {
			view.GroupLabnum = group.CurrentLabNum - 1
		}
	}

	execTemplate("maincoursepage.html", w, view)
}

// UpdateCourseURL is the URL used to call UpdateCourseHandler.
var UpdateCourseURL = "/course/update"

// UpdateCourseHandler is a http handler used to update course information.
func UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
	if err != nil {
		http.Redirect(w, r, pages.FRONTPAGE, 307)
		log.Println(err)
		return
	}

	r.ParseForm()
	orgname := r.FormValue("org")

	org, err := git.NewOrganization(orgname, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, "Not valid organization.", 404)
		return
	}

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

	if r.FormValue("screenname") == "" {
		org.ScreenName = org.Name
	} else {
		org.ScreenName = r.FormValue("screenname")
	}

	org.Description = r.FormValue("desc")
	org.Private = r.FormValue("private") == "on"
	org.CodeReview = r.FormValue("codereview") == "on"

	org.Slipdays = r.FormValue("slipdays") == "on"
	maxslipdays, err := strconv.Atoi(r.FormValue("maxslipdays"))
	if err != nil {
		http.Error(w, "Cant use the max slip days format.", 415)
		return
	}
	org.SlipdaysMax = maxslipdays

	basepath := r.FormValue("basepath")
	if basepath != "" {
		org.CI.Basepath = basepath
	}

	indvfolders := r.PostForm["lab"]
	for i := 1; i <= indv; i = i + 1 {
		if len(indvfolders) <= i-1 {
			org.IndividualLabFolders[i] = "lab" + strconv.Itoa(i)
			continue
		}

		if fname := indvfolders[i-1]; fname != "" {
			org.IndividualLabFolders[i] = fname
		} else {
			org.IndividualLabFolders[i] = "lab" + strconv.Itoa(i)
		}
	}

	groupfolders := r.PostForm["group"]
	for i := 1; i <= groups; i = i + 1 {
		if len(groupfolders) <= i-1 {
			org.GroupLabFolders[i] = "grouplab" + strconv.Itoa(i)
			continue
		}

		if fname := groupfolders[i-1]; fname != "" {
			org.GroupLabFolders[i] = fname
		} else {
			org.GroupLabFolders[i] = "grouplab" + strconv.Itoa(i)
		}
	}

	timelayout := "02/01/2006 15:04"
	indvdeadlines := r.PostForm["indvdeadline"]
	for i := 1; i <= indv; i = i + 1 {
		if len(indvdeadlines) <= i-1 {
			org.SetIndividualDeadline(i, time.Now())
			continue
		}

		if timestring := indvdeadlines[i-1]; timestring != "" {
			t, err := time.Parse(timelayout, timestring)
			if err != nil {
				org.SetIndividualDeadline(i, time.Now())
			} else {
				org.SetIndividualDeadline(i, t)
			}
		}
	}

	groupdeadlines := r.PostForm["groupdeadline"]
	for i := 1; i <= groups; i = i + 1 {
		if len(groupdeadlines) <= i-1 {
			org.SetGroupDeadline(i, time.Now())
			continue
		}

		if timestring := groupdeadlines[i-1]; timestring != "" {
			t, err := time.Parse(timelayout, timestring)
			if err != nil {
				org.SetGroupDeadline(i, time.Now())
			} else {
				org.SetGroupDeadline(i, t)
			}
		}
	}

	http.Redirect(w, r, "/course/teacher/"+org.Name, 307)
}

// RemovePendingUserURL is the URL used to call RemovePendingUserHandler.
var RemovePendingUserURL = "/course/removepending"

// RemovePendingUserHandler is http handler used to remove users from the list of pending students on a course.
func RemovePendingUserHandler(w http.ResponseWriter, r *http.Request) {
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

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, "Not valid organization.", 404)
		return
	}
	defer func() {
		err := org.Save()
		if err != nil {
			org.Unlock()
			log.Println(err)
		}
	}()

	org.Lock()
	defer org.Unlock()

	if !org.IsTeacher(member) {
		http.Error(w, "Is not a teacher or assistant for this course.", 404)
		return
	}

	if _, ok := org.PendingUser[username]; ok {
		delete(org.PendingUser, username)
	}
}

// RemovePendingUserURL is the URL used to call RemovePendingUserHandler.
var RemoveUserURL = "/course/removemember"

// RemoveUserHandler is http handler used to remove users from the list of students on a course.
func RemoveUserHandler(w http.ResponseWriter, r *http.Request) {
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

	org, err := git.NewOrganization(course, false)
	if err != nil {
		http.Error(w, "Not valid organization.", 404)
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
		http.Error(w, "Is not a teacher or assistant for this course.", 404)
		return
	}

	user, err := git.NewMemberFromUsername(username, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer func() {
		err := user.Save()
		if err != nil {
			user.Unlock()
			log.Println(err)
		}
	}()

	if org.IsMember(user) {
		org.RemoveMembership(user)
		user.RemoveOrganization(org)
	} else {
		http.Error(w, "Couldn't find this user in this course. ", 404)
		return
	}
}

// ListStudentsView is the struct used to return values from ListStudentsHandler.
type ListStudentsView struct {
	course string

	students []*git.Member
}

// ListStudentsURL is the url used to call ListStudentsHandler.
var ListStudentsURL = "/course/students"

// ListStudentsHandler will get all the members of a course.
func ListStudentsHandler(w http.ResponseWriter, r *http.Request) {

}
