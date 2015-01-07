package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/hfurubotten/autograder/global"
)

var htmlBase string = global.Basepath + "web/html/"

func execTemplate(page string, w http.ResponseWriter, view interface{}) {
	t, err := template.ParseFiles(htmlBase+page, htmlBase+"template.html")
	if err != nil {
		log.Printf("Error parsing %s: %v", page, err)
		return
	}
	err = t.ExecuteTemplate(w, "template", view)
	if err != nil {
		log.Printf("Error executing template for %s: %v", page, err)
	}
}
