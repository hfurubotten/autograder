package web

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/hfurubotten/autograder/auth"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

type Webserver struct {
	Port int
}

func NewWebServer(port int) Webserver {
	return Webserver{port}
}

func (ws Webserver) Start() {

	// OAuth process
	http.Handle("/login", http.RedirectHandler(auth.RedirectURL+"?client_id="+auth.Clientid+"&scope="+auth.Scope, 307))
	http.HandleFunc("/oauth", auth.Handler)
	http.HandleFunc(pages.SIGNOUT, auth.RemoveApprovalHandler)

	// Page handlers
	http.HandleFunc("/home", homehandler)
	http.HandleFunc(pages.REGISTER_REDIRECT, profilehandler)
	http.HandleFunc("/course/new", newcoursehandler)
	http.HandleFunc("/course/new/org", newcoursehandler)
	http.HandleFunc("/course/new/org/", selectorghandler)
	http.HandleFunc("/course/create", saveorghandler)
	http.HandleFunc("/course/register", newcoursememberhandler)
	http.HandleFunc("/course/register/", newcoursememberhandler)
	http.HandleFunc("/admin", adminhandler)

	// proccessing handlers
	http.HandleFunc("/updatemember", updatememberhandler)
	http.HandleFunc("/admin/teacher", setteacherhandler)
	http.HandleFunc("/admin/user", setadminhandler)

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
			pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
			return
		}

		index, err := os.Open("web/html/index.html")
		if err != nil {
			log.Fatal(err)
		}
		//err :=indextemplate.Execute(w, nil)
		_, err = io.Copy(w, index)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}

func homehandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions: ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	type homeview struct {
		Member *git.Member
		Org    []string
	}

	view := homeview{}

	member := git.NewMember(value.(string))

	view.Member = &member
	view.Org, err = member.ListOrgs()
	if err != nil {
		log.Println(err)
		pages.RedirectTo(w, r, pages.SIGNOUT, 307)
		return
	}

	if !member.IsComplete() {
		pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
		return
	}

	t, err := template.ParseFiles("web/html/home.html", "web/html/template.html")
	if err != nil {
		log.Println("Error parsing register html: ", err)
		return
	}

	err = t.ExecuteTemplate(w, "template", view)
	if err != nil {
		log.Println("Error execute register html: ", err)
		return
	}
}
