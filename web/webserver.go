package web

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
	"github.com/hfurubotten/autograder/web/staticfiles"
)

// We use go generate to build in all static files into a go file. Each time the
// static files are changed/removed/added the go generate command need to be
// executed. And need to be added together with the commit for the changed
// static files. The go-bindata program can be obtained by running:
//
//   go get -u github.com/jteeuwen/go-bindata/go-bindata
//
//go:generate $GOPATH/bin/go-bindata -o=staticfiles/staticfiles.go -pkg=staticfiles css/ fonts/ html/... img/... js/

// the default html base path; used to access assets
var htmlBase = "html/"

// SetHandlers sets up http handler functions for the Autograder web server.
func SetHandlers() {
	// OAuth handlers
	http.Handle(pages.Signin, http.RedirectHandler(auth.OAuthRedirectURL(), http.StatusTemporaryRedirect))
	http.HandleFunc(pages.OAuth, auth.OAuthHandler)
	http.HandleFunc(pages.Signout, auth.RemoveApprovalHandler)

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
	http.HandleFunc(RemoveUserURL, RemoveUserHandler)
	http.HandleFunc(ApproveGroupURL, ApproveGroupHandler)
	http.HandleFunc(AddAssistantURL, AddAssistantHandler)
	http.HandleFunc(RemoveAssistantURL, RemoveAssistantHandler)
	http.HandleFunc(RemovePendingUserURL, RemovePendingUserHandler)
	http.HandleFunc(WebhookEventURL, WebhookEventHandler)
	http.HandleFunc(ManualCITriggerURL, ManualCITriggerHandler)
	http.HandleFunc(PublishReviewURL, PublishReviewHandler)
	http.HandleFunc(ListReviewsURL, ListReviewsHandler)
	http.HandleFunc(LeaderboardDataURL, LeaderboardDataHandler)
	http.HandleFunc(CIResultListURL, CIResultListHandler)
	http.HandleFunc(NotesURL, NotesHandler)
	http.HandleFunc(SlipdaysURL, SlipdaysHandler)

	// static files
	// http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js/"))))
	// http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))
	// http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	// http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("web/fonts/"))))
	http.HandleFunc("/js/", StaticfilesHandler)
	http.HandleFunc("/css/", StaticfilesHandler)
	http.HandleFunc("/img/", StaticfilesHandler)
	http.HandleFunc("/fonts/", StaticfilesHandler)

	// catch all URLs not matched by any other patterns
	http.HandleFunc(CatchAllURL, CatchAllHandler)
}

// CatchAllURL is the URL used to call CatchAllHandler.
var CatchAllURL = "/"

// CatchAllHandler is a http handler which is meant to catch empty and non existing pages.
func CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if auth.IsApprovedUser(r) {
			http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
			return
		}

		data, err := staticfiles.Asset(htmlBase + "index.html")
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		if _, err = w.Write(data); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
		}

		// index, err := os.Open(htmlBase + "index.html")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// //err :=indextemplate.Execute(w, nil)
		// _, err = io.Copy(w, index)
		// if err != nil {
		// 	log.Println("Error sending frontpage:", err)
		// }

	} else {
		http.Error(w, "This is not the page you are looking for!\n", http.StatusNotFound)
	}
}

// StaticfilesHandler handles finding static files for the server.
func StaticfilesHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	if strings.HasPrefix(uri, "/") {
		uri = uri[1:]
	}

	data, err := staticfiles.Asset(uri)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if strings.HasSuffix(uri, ".css") {
		w.Header().Add("content-type", "text/css")
	} else if strings.HasSuffix(uri, ".js") {
		w.Header().Add("content-type", "text/javascript")
	} else if strings.HasSuffix(uri, ".woff") {
		w.Header().Add("content-type", "application/font-woff")
	}

	if _, err = w.Write(data); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HomeView is the view passed to the html template compiler in HomeHandler.
type HomeView struct {
	stdTemplate
	Teaching  map[string]*entities.Organization
	Assisting map[string]*entities.Organization
	Courses   map[string]*entities.Organization
}

// HomeURL is the URL used to call HomeHandler.
var HomeURL = "/home"

// HomeHandler is a http handler for the home page for logged in users.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Found member: %v", member)

	view := HomeView{
		stdTemplate: stdTemplate{
			Member: member,
		},
		Teaching:  make(map[string]*entities.Organization),
		Assisting: make(map[string]*entities.Organization),
		Courses:   make(map[string]*entities.Organization),
	}

	for key := range member.Teaching {
		view.Teaching[key], _ = entities.NewOrganization(key, true)
	}
	for key := range member.AssistantCourses {
		view.Assisting[key], _ = entities.NewOrganization(key, true)
	}
	for key := range member.Courses {
		view.Courses[key], _ = entities.NewOrganization(key, true)
	}

	//TODO: This redirect could be made obsolete if we can guarantee that nobody
	// gets to login before their member status is complete.
	if !member.IsComplete() {
		http.Redirect(w, r, pages.Profile, http.StatusTemporaryRedirect)
		return
	}

	execTemplate("home.html", w, view)
}

// checkAdminApproval will check the sessions of the user and see if the user is
// logged in. If the user is not logged in the function will return error. If the
// redirect is true the function also writes a redirect to the response headers.
//
// Member returned is standard read only. If written to, locking need to be done manually.
func checkMemberApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member *entities.Member, err error) {
	if !auth.IsApprovedUser(r) {
		if redirect {
			http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
		}
		err = errors.New("user is not logged in")
		return
	}

	value, err := sessions.GetSessions(r, sessions.AuthSession, sessions.AccessTokenSessionKey)
	if err != nil {
		//TODO: why overwrite the error from GetSessions??
		err = errors.New("failed to get access token from session")
		if redirect {
			http.Redirect(w, r, pages.Front, http.StatusTemporaryRedirect)
		}
		return
	}

	member, err = entities.LookupMember(value.(string))
	if err != nil {
		return nil, err
	}

	if !member.IsComplete() {
		if redirect {
			http.Redirect(w, r, pages.Profile, http.StatusTemporaryRedirect)
		}
		err = errors.New("member with incomplete profile, redirecting")
		return
	}

	return
}

// checkAdminApproval will check the sessions of the user and see if the user is
// a teacher. If the user is not a teacher or logged in the function will return
// error. If the redirect is true the function also writes a redirect to the
// response headers.
//
// Member returned is standard read only. If written to, locking need to be done manually.
func checkTeacherApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member *entities.Member, err error) {
	member, err = checkMemberApproval(w, r, redirect)
	if err != nil {
		return
	}

	if !member.IsTeacher && !member.IsAssistant {
		err = errors.New("user is not a teacher: " + member.Username)
		if redirect {
			http.Redirect(w, r, pages.Home, http.StatusTemporaryRedirect)
		}
		return
	}

	if member.Scope == "" && member.IsTeacher {
		err = errors.New("teacher must renew scope: " + member.Username)
		if redirect {
			http.Redirect(w, r, auth.OAuthScopeRedirectURL(), http.StatusTemporaryRedirect)
		}
		return
	}

	return
}

// checkAdminApproval checks the sessions of the user and see if the user is
// a system admin. If the user is not an admin or a user the function will
// return error. If the redirect is true the function also writes a redirect to
// the response headers.
//
// Member returned is standard read only. If written to, locking need to be done
// manually.
func checkAdminApproval(w http.ResponseWriter, r *http.Request, redirect bool) (*entities.Member, error) {
	member, err := checkMemberApproval(w, r, redirect)
	if err != nil {
		return nil, err
	}
	if !member.IsAdmin {
		return nil, errors.New("user is not admin: " + member.Username)
	}
	return member, nil
}
