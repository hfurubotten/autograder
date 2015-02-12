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

var htmlBase string = global.Basepath + "web/html/"

type Webserver struct {
	Port int
}

func NewWebServer(port int) Webserver {
	return Webserver{port}
}

// func FakeServer(port int, stopchan <-chan int) {
// 	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	listener, err := net.ListenTCP("tcp", tcpAddr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

// 	})
// 	http.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {

// 	})

// 	err = http.Serve(listener, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	<-stopchan

// 	listener.Close()
// }

func (ws Webserver) Start() {

	// OAuth process
	http.Handle("/login", http.RedirectHandler(global.OAuth_RedirectURL+"?client_id="+global.OAuth_ClientID, 307))
	http.HandleFunc("/oauth", global.OAuth_Handler)
	http.HandleFunc(pages.SIGNOUT, auth.RemoveApprovalHandler)

	// Page handlers
	http.HandleFunc("/home", homehandler)
	http.HandleFunc(pages.REGISTER_REDIRECT, profilehandler)
	http.HandleFunc("/course/new", newcoursehandler)
	http.HandleFunc("/course/new/org", newcoursehandler)
	http.HandleFunc("/course/new/org/", selectorghandler)
	http.HandleFunc("/course/create", saveorghandler)
	http.HandleFunc("/course/register", newcoursememberhandler)
	http.HandleFunc("/course/register/", registercoursememberhandler)
	http.HandleFunc("/course/teacher/", teacherspanelhandler)
	http.HandleFunc("/admin", adminhandler)
	http.HandleFunc("/course/", maincoursepagehandler)
	http.HandleFunc("/course/result/", showresulthandler)
	http.HandleFunc("/help/", helphandler)

	// proccessing handlers
	http.HandleFunc("/updatemember", updatememberhandler)
	http.HandleFunc("/admin/teacher", setteacherhandler)
	http.HandleFunc("/admin/user", setadminhandler)
	http.HandleFunc("/course/approvemember/", approvecoursemembershiphandler)
	http.HandleFunc("/course/approvelab", approvelabhandler)
	http.HandleFunc("/course/ciresutls", ciresulthandler)
	http.HandleFunc("/course/cisummary", ciresultsummaryhandler)
	http.HandleFunc("/course/update", updatecoursehandler)
	http.HandleFunc("/course/newgroup", newgrouphandler)
	http.HandleFunc("/course/requestrandomgroup", requestrandomgrouphandler)
	http.HandleFunc("/course/removegroup", removependinggrouphandler)
	http.HandleFunc("/course/approvegroup", approvegrouphandler)
	http.HandleFunc("/course/addassistant", addassistanthandler)
	http.HandleFunc("/course/removepending", removependinguserhandler)
	http.HandleFunc("/event/hook", webhookeventhandler)
	http.HandleFunc("/event/manualbuild", manualcihandler)

	// static files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("web/fonts/"))))

	// catch all not matched wth other patterns
	http.HandleFunc("/", catchallhandler)

	// start the server
	log.Println("Starts listening")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(ws.Port), nil))
}

//var indextemplate = template.Must(template.New("index").ParseFiles("web/html/index.html"))

func catchallhandler(w http.ResponseWriter, r *http.Request) {
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

type homeview struct {
	Member    *git.Member
	Teaching  map[string]git.Organization
	Assisting map[string]git.Organization
	Courses   map[string]git.Organization
}

func homehandler(w http.ResponseWriter, r *http.Request) {
	member, err := checkMemberApproval(w, r, true)
	if err != nil {
		return
	}

	view := homeview{
		Member:    &member,
		Teaching:  make(map[string]git.Organization),
		Assisting: make(map[string]git.Organization),
		Courses:   make(map[string]git.Organization),
	}

	for key, _ := range member.Teaching {
		view.Teaching[key] = git.NewOrganization(key)
	}
	for key, _ := range member.AssistantCourses {
		view.Assisting[key] = git.NewOrganization(key)
	}
	for key, _ := range member.Courses {
		view.Courses[key] = git.NewOrganization(key)
	}

	if !member.IsComplete() {
		http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	execTemplate("home.html", w, view)
}

func checkMemberApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member git.Member, err error) {
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

	member = git.NewMember(value.(string))

	if !member.IsComplete() {
		if redirect {
			http.Redirect(w, r, pages.REGISTER_REDIRECT, 307)
		}
		err = errors.New("Member got an uncomplete profile, redirecting.")
		return
	}

	return
}

func checkTeacherApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member git.Member, err error) {
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

func checkAdminApproval(w http.ResponseWriter, r *http.Request, redirect bool) (member git.Member, err error) {
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
