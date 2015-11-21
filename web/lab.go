package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hfurubotten/autograder/ci"
	git "github.com/hfurubotten/autograder/entities"
)

// ApproveLabURL is the URL used to call ApproveLabHandler.
var ApproveLabURL = "/course/approvelab"

// ApproveLabHandler is a http handler used by teachers to approve a lab.
func ApproveLabHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkTeacherApproval(w, r, true)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if approve != "true" {
		log.Println("Missing approval")
		http.Error(w, "Not approved", http.StatusNotFound)
		return
	}

	if !git.HasOrganization(course) || username == "" {
		log.Println("Missing username or uncorrect course")
		http.Error(w, "Unknown Organization", http.StatusNotFound)
		return
	}

	org, err := git.NewOrganization(course, true)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !org.IsTeacher(member) {
		log.Println(member.Name + " is not a teacher of " + org.Name)
		http.Error(w, "Not a teacher of this course.", http.StatusNotFound)
		return
	}

	var isgroup bool
	if git.HasMember(username) {
		isgroup = false
	} else {
		isgroup = strings.Contains(username, "group")
		if !isgroup {
			log.Println("No user found")
			http.Error(w, "Unknown User", http.StatusNotFound)
			return
		}
	}

	var latestbuild int
	var res *ci.BuildResult
	if isgroup {
		group, err := git.GetGroup(username) // username==groupname TODO consider changing this
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		defer func() {
			if err := group.Save(); err != nil {
				group.Unlock()
				log.Println(err)
			}
		}()

		latestbuild = group.GetLastBuildID(labnum)
		if latestbuild < 0 {
			http.Error(w, "No build registered on lab.", http.StatusInternalServerError)
			return
		}

		res, err = ci.GetBuildResult(latestbuild)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		group.SetApprovedBuild(res.Labnum, res.ID, res.PushTime)

		if org.Slipdays {
			for username := range group.Members {
				user, err := git.GetMember(username)
				if err != nil {
					log.Println(err)
					continue
				}

				copt := user.Courses[org.Name]
				err = copt.RecalculateSlipDays()
				if err != nil {
					log.Println(err)
				}
				user.Courses[org.Name] = copt
			}
		}
	} else {
		user, err := git.GetMember(username)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() {
			if err := user.Save(); err != nil {
				user.Unlock()
				log.Println(err)
			}
		}()

		latestbuild = user.GetLastBuildID(course, labnum)
		if latestbuild < 0 {
			http.Error(w, "No build registered on lab.", http.StatusInternalServerError)
			return
		}

		res, err = ci.GetBuildResult(latestbuild)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user.SetApprovedBuild(org.Name, res.Labnum, res.ID, res.PushTime)

		if org.Slipdays {
			copt := user.Courses[org.Name]
			err = copt.RecalculateSlipDays()
			if err != nil {
				log.Println(err)
			}
			user.Courses[org.Name] = copt
		}
	}

	res.Status = "Approved"

	if err := res.Save(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NotesURL is the url used to call AddNotesHandler.
var NotesURL = "/course/notes"

// NotesView is the object which is returned when NotesHandler is called with
// POST header.
type NotesView struct {
	Course   string
	Username string
	Group    int
	Labnum   int
	Notes    string
}

// NotesHandler will add a note to a lab for a given user.
// Page requested with method GET will return latest note and POST will store a
// new note to the user or group.
// required input:
// - Course
// - Username //or
// - Group
// - labnum
// - Notes
func NotesHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	teacher, err := checkTeacherApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	course := r.FormValue("Course")
	username := r.FormValue("Username")
	notes := r.FormValue("Notes")
	groupid, _ := strconv.Atoi(r.FormValue("Group"))
	labnum, err := strconv.Atoi(r.FormValue("Labnum"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	org, err := git.NewOrganization(course, true)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !org.IsTeacher(teacher) {
		log.Println(err)
		http.Error(w, "Not a teacher of this course", http.StatusNotFound)
		return
	}

	if groupid > 0 {
		groupName := git.GroupRepoPrefix + strconv.Itoa(groupid)
		group, err := git.GetGroup(groupName)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if group.Course != org.Name {
			log.Println(err)
			http.Error(w, "Not a group in this course", http.StatusNotFound)
			return
		}

		if r.Method == "POST" {
			group.AddNotes(labnum, notes)
		} else {
			view := &NotesView{
				Course: course,
				Group:  groupid,
				Labnum: labnum,
				Notes:  group.GetNotes(labnum),
			}

			enc := json.NewEncoder(w)
			if err = enc.Encode(view); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err = group.Save(); err != nil {
			group.Unlock()
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		user, err := git.GetMember(username)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !org.IsMember(user) {
			log.Println(err)
			http.Error(w, "Not a member of this course", http.StatusNotFound)
			return
		}

		if r.Method == "POST" {
			user.AddNotes(org.Name, labnum, notes)
		} else {
			view := &NotesView{
				Course:   course,
				Username: username,
				Labnum:   labnum,
				Notes:    user.GetNotes(course, labnum),
			}

			enc := json.NewEncoder(w)
			if err = enc.Encode(view); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err = user.Save(); err != nil {
			user.Unlock()
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// SlipdaysView is the structure returned when
type SlipdaysView struct {
	Course       string
	Username     string
	UsedSlipdays int
	MaxSlipdays  int
}

// SlipdaysURL is the url used to call SlipdaysHandler.
var SlipdaysURL = "/course/slipdays"

// SlipdaysHandler is used to get used slipdays for a user in a course.
func SlipdaysHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is signed in and a teacher.
	member, err := checkMemberApproval(w, r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	orgname := r.FormValue("Course")

	org, err := git.NewOrganization(orgname, true)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := SlipdaysView{
		Course:      orgname,
		MaxSlipdays: org.SlipdaysMax,
	}

	if org.IsTeacher(member) {
		username := r.FormValue("Username")

		user, err := git.GetMember(username)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if !org.IsMember(member) {
			http.Error(w, "Unknown member of course.", http.StatusNotFound)
			return
		}

		courseopt := user.Courses[org.Name]

		view.UsedSlipdays = courseopt.UsedSlipDays
		view.Username = user.Username
	} else if org.IsMember(member) {
		courseopt := member.Courses[org.Name]

		view.UsedSlipdays = courseopt.UsedSlipDays
		view.Username = member.Username
	} else {
		http.Error(w, "Unknown member of course.", http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(view)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
