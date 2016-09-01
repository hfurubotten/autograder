package web

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/hfurubotten/autograder/entities"
)

func xTestExecTemplate(t *testing.T) {
	// page := "newcourse-info.html"
	page := "newcourse-orgselect.html"
	w := httptest.NewRecorder()
	view := CourseView{}
	var err error
	view.Member, err = entities.CreateMember("bond")
	if err != nil {
		t.Error(err)
	}
	view.Orgs = []string{"finale", "john", "jarle"}
	view.Org = "halloen"
	execTemplate(page, w, view)
	fmt.Printf("%d - %s", w.Code, w.Body.String())
}
