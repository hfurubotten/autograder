package web

import (
	"html/template"
	"log"
	"net/http"
	"time"

	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/staticfiles"
)

type StdTemplate struct {
	OptinalHeadline bool
	Member          *git.Member
}

var funcMap = template.FuncMap{
	"noescape": func(s string) template.HTML {
		return template.HTML(s)
	},
	"noescapetime": func(t time.Time) template.HTML {
		return template.HTML(t.String())
	},
	// selectitem adds the "selected" attribute to an option in
	// a drop down menu. It takes as input the current option value
	// and the option value that is to be selected. It returns
	// "selected" if they match, otherwise it returns and empty string.
	"selectitem": func(current int, selected int) template.HTMLAttr {
		if current == selected {
			return template.HTMLAttr("selected")
		}
		return template.HTMLAttr("")
	},
}

func execTemplate(page string, w http.ResponseWriter, view interface{}) {
	pagedata, err := staticfiles.Asset(htmlBase + page)
	if err != nil {
		http.Error(w, "Page not found", 404)
		return
	}

	templatedata, err := staticfiles.Asset(htmlBase + "template.html")
	if err != nil {
		http.Error(w, "Template not found", 404)
		return
	}

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
