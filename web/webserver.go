package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	//"html/template"
	"io"
	"os"

	"github.com/hfurubotten/autograder/auth"
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

	// Page handlers
	http.HandleFunc("/session", sessionshandler)

	// static files
	http.Handle("/js/*", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js/"))))
	http.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))
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

func sessionshandler(w http.ResponseWriter, r *http.Request) {
	val, err := sessions.GetSessions(r, sessions.AUTHSESSION, "user")
	fmt.Fprintln(w, val)
	fmt.Fprintln(w, err)
}