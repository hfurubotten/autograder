package web

import (
	"errors"
	"io"
	"log"
	//"net"
	"net/http"
	"os"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

var htmlBase string

type Webserver struct {
	Port int
}

// NewWebServer will return a new Webserver object with possibility to listen to given port.
func NewWebServer(port int) Webserver {
	return Webserver{port}
}

// Start will start up a new server listening on ws.Port. This
// method blocks, and will call os.Exit(1) if server error occures.
func (ws Webserver) Start() {
	// setting html base path
	htmlBase = global.Basepath + "web/html/"

	// OAuth process
	http.Handle("/login", http.RedirectHandler(global.OAuth_RedirectURL+"?client_id="+global.OAuth_ClientID, 307))
	http.HandleFunc("/oauth", global.OAuth_Handler)
	http.HandleFunc(pages.SIGNOUT, auth.RemoveApprovalHandler)

	// Page handlers
	http.HandleFunc(HomeURL, HomeHandler)
	http.HandleFunc(ProfileURL, ProfileHandler)
	http.HandleFunc(NewCourseInfoURL, NewCourseHandler)
	http.HandleFunc(NewCourseURL, NewCourseHandler)
	http.HandleFunc(SelectOrgURL, SelectOrgHandler)
	http.HandleFunc(CreateOrgURL, CreateOrgHandler)
	http.HandleFunc(NewCourseMemberURL, NewCourseMemberHandler)
	http.HandleFunc(RegisterCourseMemberURL, RegisterCourseMemberHandler)
	http.HandleFunc(TeachersPanelURL, TeachersPanelHandler)
	http.HandleFunc(AdminURL, AdminHandler)
	http.HandleFunc(UserCoursePageURL, UserCoursePageHandler)
	http.HandleFunc(ShowResultURL, ShowResultHandler)
	http.HandleFunc(HelpURL, HelpHandler)
	http.HandleFunc(ScoreboardURL, ScoreboardHandler)

	// proccessing handlers
	http.HandleFunc(UpdateMemberURL, UpdateMemberHandler)
	http.HandleFunc(SetTeacherURL, SetTeacherHandler)
	http.HandleFunc(SetAdminURL, SetAdminHandler)
	http.HandleFunc(ApproveCourseMembershipURL, ApproveCourseMembershipHandler)
	http.HandleFunc(ApproveLabURL, ApproveLabHandler)
	http.HandleFunc(CIResultURL, CIResultHandler)
	http.HandleFunc(CIResultSummaryURL, CIResultSummaryHandler)
	http.HandleFunc(UpdateCourseURL, UpdateCourseHandler)
	http.HandleFunc(NewGroupURL, NewGroupHandler)
	http.HandleFunc(RequestRandomGroupURL, RequestRandomGroupHandler)
	http.HandleFunc(RemovePendingGroupURL, RemovePendingGroupHandler)
	http.HandleFunc(ApproveGroupUrl, ApproveGroupHandler)
	http.HandleFunc(AddAssistantURL, AddAssistantHandler)
	http.HandleFunc(RemovePendingUserURL, RemovePendingUserHandler)
	http.HandleFunc(WebhookEventURL, WebhookEventHandler)
	http.HandleFunc(ManualCITriggerURL, ManualCITriggerHandler)
	http.HandleFunc(PublishReviewURL, PublishReviewHandler)
	http.HandleFunc(ListReviewsURL, ListReviewsHandler)

	// static files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("web/fonts/"))))

	// catch all not matched wth other patterns
	http.HandleFunc(CatchAllURL, CatchAllHandler)

	// start the server
	log.Println("Starts listening")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(ws.Port), nil))
}

var CatchAllURL string = "/"

// CatchAllHandler is a http handler which is meant to catch empty and non existing pages.
func CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if auth.IsApprovedUser(r) {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
			return
		}

		index, err := os.Open(htmlBase + "index.html")
		if err != nil {
			log.Fatal(err)
		}
		//err :=indextemplate.Execute(w, nil)
		_, err = io.Copy(w, index)
		if err != nil {
			log.Println("Error sending frontpage:", err)
		}

	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}

type HomeView struct {
	Member    *git.Member
	Teaching  map[string]*git.Organization
	Assisting map[string]*git.Organization
	Courses   map[string]*git.Organization
}

// HomeURL is the URL used to call HomeHandler.
var HomeURL string = "/home"

// homehandler is a http handler for the home page for logged in users.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		return
	}

	view := HomeView{
		Member:    member,
		Teaching:  make(map[string]*git.Organization),
		Assisting: make(map[string]*git.Organization),
		Courses:   make(map[string]*git.Organization),
	}

	for key, _ := range member.Teaching {
		view.Teaching[key], _ = git.NewOrganization(key)
	}
	for key, _ := range member.AssistantCourses {
		view.Assisting[key], _ = git.NewOrganization(key)
	}
	for key, _ := range member.Courses {
		view.Courses[key], _ = git.NewOrganization(key)
	}

	if !member.IsComplete() {
		http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	execTemplate("home.html", w, view)
}

// checkAdminApproval will check the sessions of the user and see if the user is logged in.
// If the user is not logged in the function will return error. If the redirect is true
// the function also writes a redirect to the response headers.
func checkMemberApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member *git.Member, err error) {
	if !auth.IsApprovedUser(r) {
		if redirect {
			http.Redirect(w, r, pages.FRONTPAGE, 307)
		}
		err = errors.New("The user is not logged in")
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		err = errors.New("Error getting access token from sessions")
		if redirect {
			http.Redirect(w, r, pages.FRONTPAGE, 307)
		}
		return
	}

	member, err = git.NewMember(value.(string))
	if err != nil {
		return nil, err
	}

	if !member.IsComplete() {
		if redirect {
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
		}
		err = errors.New("Member got an uncomplete profile, redirecting.")
		return
	}

	return
}

// checkAdminApproval will check the sessions of the user and see if the user is a teacher.
// If the user is not a teacher or logged in the function will return error. If the redirect is true
// the function also writes a redirect to the response headers.
func checkTeacherApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member *git.Member, err error) {
	member, err = checkMemberApproval(w, r, redirect)
	if err != nil {
		return
	}

	if !member.IsTeacher && !member.IsAssistant {
		err = errors.New("The user is not a teacher.")
		if redirect {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
		}
		return
	}

	if member.Scope == "" && member.IsTeacher {
		err = errors.New("Teacher need to renew scope.")
		if redirect {
			http.Redirect(w, r, global.OAuth_RedirectURL+"?client_id="+global.OAuth_ClientID+"&scope="+global.OAuth_Scope, 307)
		}
		return
	}

	return
}

// checkAdminApproval will check the sessions of the user and see if the user is a system admin.
// If the user is not an admin or a user the function will return error. If the redirect is true
// the function also writes a redirect to the response headers.
func checkAdminApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member *git.Member, err error) {
	member, err = checkMemberApproval(w, r, redirect)
	if err != nil {
		return
	}

	if !member.IsAdmin {
		err = errors.New("The user is not a teacher.")
		if redirect {
			http.Redirect(w, r, pages.HOMEPAGE, 307)
		}
		return
	}

	return
}
