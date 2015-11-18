package web

import (
	"html/template"
	"log"
	"net/http"
	"time"

	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/staticfiles"
)

// stdTemplate is the standard common view components
type stdTemplate struct {
	OptionalHeadline bool
	Member           *git.Member
}

var funcMap = template.FuncMap{
	"noescape": func(s string) template.HTML {
		return template.HTML(s)
	},
	"noescapetime": func(t time.Time) template.HTML {
		return template.HTML(t.String())
	},
}

func execTemplate(page string, w http.ResponseWriter, view interface{}) {
	pagedata, err := staticfiles.Asset(htmlBase + page)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	templatedata, err := staticfiles.Asset(htmlBase + "template.html")
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	//TODO shouldn't also the errors below be reported back to the web client?
	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(templatedata))
	if err != nil {
		log.Printf("Error parsing %s: %v", page, err)
		return
	}
	t, err = t.Parse(string(pagedata))
	if err != nil {
		log.Printf("Error parsing %s: %v", page, err)
		return
	}
	// t, err := t.ParseFiles(htmlBase+page, htmlBase+"template.html")
	// if err != nil {
	// 	log.Printf("Error parsing %s: %v", page, err)
	// 	return
	// }
	err = t.ExecuteTemplate(w, "template", view)
	if err != nil {
		log.Printf("Error executing template for %s: %v", page, err)
	}
}
