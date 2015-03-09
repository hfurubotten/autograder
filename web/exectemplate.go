package web

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var funcMap = template.FuncMap{
	"noescape": func(s string) template.HTML {
		return template.HTML(s)
	},
	"noescapetime": func(t time.Time) template.HTML {
		return template.HTML(t.String())
	},
}

func execTemplate(page string, w http.ResponseWriter, view interface{}) {
	t := template.New("template").Funcs(funcMap)
	t, err := t.ParseFiles(htmlBase+page, htmlBase+"template.html")
	if err != nil {
		log.Printf("Error parsing %s: %v", page, err)
		return
	}
	err = t.ExecuteTemplate(w, "template", view)
	if err != nil {
		log.Printf("Error executing template for %s: %v", page, err)
	}
}
