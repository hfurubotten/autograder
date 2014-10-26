package web

import (
	"log"
	"net/http"
	"strconv"
	"html/template"
	"io"
	"os"	

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
	http.Handle("/login", http.RedirectHandler(auth.RedirectURL + "?client_id=" + auth.Clientid + "&scope=" + auth.Scope, 307))
	http.HandleFunc("/oauth", auth.Handler)
	http.HandleFunc(pages.SIGNOUT, auth.RemoveApprovalHandler)

	// Page handlers
	http.HandleFunc("/home", homehandler)
	http.HandleFunc(pages.REGISTER_REDIRECT, profilehandler)

	// proccessing handlers
	http.HandleFunc("/updatemember", updatememberhandler)

	// static files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))

	// catch all not matched wth other patterns
	http.HandleFunc("/", catchallhandler)

	// start the server
	log.Println("Starts listening")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(ws.Port), nil))
}

//var indextemplate = template.Must(template.New("index").ParseFiles("web/html/index.html"))

func catchallhandler(w http.ResponseWriter, r *http.Request){
	if r.URL.Path == "/" || r.URL.Path == ""{
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
		log.Println("Error getting access token from sessions(web/webserver): ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	type homeview struct {
		Member *git.Member
		Org []string
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

	t, err := template.ParseFiles("web/html/home.html")
	if err != nil {
		log.Println("Error parsing register html(web/webserver): ", err)
		return
	}

	err = t.Execute(w, view)
	if err != nil {
		log.Println("Error execute register html(web/webserver): ", err)
		return
	}
}

func profilehandler(w http.ResponseWriter, r *http.Request){
	if !auth.IsApprovedUser(r) {
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
	if err != nil {
		log.Println("Error getting access token from sessions(web/webserver): ", err)
		pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
		return
	}

	member := git.NewMember(value.(string))

	t, err := template.ParseFiles("web/html/register.html")
	if err != nil {
		log.Println("Error parsing register html(web/webserver): ", err)
		return
	}

	err = t.Execute(w, member)
	if err != nil {
		log.Println("Error execute register html(web/webserver): ", err)
		return
	}
}

func updatememberhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("name") == "" || r.FormValue("studentid") == "" {
			//pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		if !auth.IsApprovedUser(r) {
			pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
			return
		}

		value, err := sessions.GetSessions(r, sessions.AUTHSESSION, sessions.ACCESSTOKENSESSIONKEY)
		if err != nil {
			log.Println("Error getting access token from sessions(web/webserver): ", err)
			pages.RedirectTo(w, r, pages.FRONTPAGE, 307)
			return
		}

		member := git.NewMember(value.(string))
		member.Name = r.FormValue("name")
		studentid, err := strconv.Atoi(r.FormValue("studentid"))
		if err != nil {
			log.Println("studentid atoi error: ", err)
			pages.RedirectTo(w, r, pages.REGISTER_REDIRECT, 307)
			return
		}

		member.StudentID = studentid
		member.StickToSystem()


		pages.RedirectTo(w, r, pages.HOMEPAGE, 307)
	} else {
		http.Error(w, "This is not the page you are looking for!\n", 404)
	}
}