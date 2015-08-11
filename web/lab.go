package web

import (
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
		http.Error(w, "Unknown Organization", 404)
		return
	}

	org, err := git.NewOrganization(course, true)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}

	if !org.IsTeacher(member) {
		log.Println(member.Name + " is not a teacher of " + org.Name)
		http.Error(w, "Not a teacher of this course.", 404)
		return
	}

	var isgroup bool
	if git.HasMember(username) {
		isgroup = false
	} else {
		isgroup = strings.Contains(username, "group")
		if !isgroup {
			log.Println("No user found")
			http.Error(w, "Unknown User", 404)
			return
		}
	}

	var labfolder string
	if isgroup {
		gnum, err := strconv.Atoi(username[len("group"):])
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
		group, err := git.NewGroup(course, gnum, false)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}

		defer func() {
			if err := group.Save(); err != nil {
				group.Unlock()
				log.Println(err)
			}
		}()

		if group.CurrentLabNum <= labnum {
			group.CurrentLabNum = labnum + 1
		}

		labfolder = org.GroupLabFolders[labnum]
	} else {
		user, err := git.NewMemberFromUsername(username, true)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		copt := user.Courses[course]
		if copt.CurrentLabNum <= labnum {
			user.Lock()
			copt.CurrentLabNum = labnum + 1
			user.Courses[course] = copt
			user.Save()
		}

		labfolder = org.IndividualLabFolders[labnum]
	}

	teststore := ci.GetCIStorage(org.Name, username)
	// TODO: Fix new storage for builds
	res := ci.Result{}
	err = teststore.ReadGob(labfolder, &res, false)
	if err != nil {
		log.Println(err)
		return
	}
	res.Status = "Approved"

	err = teststore.WriteGob(labfolder, res)
	if err != nil {
		log.Println(err)
	}
}
