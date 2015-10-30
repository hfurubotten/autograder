package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelpStatusOK(t *testing.T) {
	req, err := http.NewRequest("GET", "http://autograder.ux.uis.no/help/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	HelpHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("http get returned status code: %d, expected: %d",
			w.Code, http.StatusOK)
	}
}
