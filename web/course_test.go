package web

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"testing"

	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/sessions"
)

func xTestSelectOrg(t *testing.T) {
	admin := "meling"
	accessToken := "dummy-token4"
	scope := "scope"

	u := entities.NewUserProfile(accessToken, admin, scope)
	m := entities.NewMember(u)
	err := entities.PutMember(accessToken, m)
	if err != nil {
		t.Error(err)
	}

	m.IsAdmin = true
	m.IsTeacher = true
	m.Name = "Hein Meling"
	m.StudentID = 33847
	m.Email, err = mail.ParseAddress("James <alice@example.com>")
	err = m.Save()
	if err != nil {
		t.Error(err)
	}

	pv := url.Values{"user": {admin}, "teacher": {"true"}}
	b := bytes.NewBuffer([]byte(pv.Encode()))
	r, err := http.NewRequest("GET", "http://example.com"+SelectOrgURL+"mycourse", b)
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// mark auth session as approved
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
	// save the access token for this session
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken)

	SelectOrgHandler(w, r)

	// SetAdminHandler(w, r)
	// buf := new(bytes.Buffer)
	// enc := json.NewEncoder(buf)
	// err = enc.Encode(ErrMissingField)
	// expected := buf.String()
	expected := "fefe"
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected: %d, got: %d", http.StatusTemporaryRedirect, w.Code)
	}

	fmt.Printf("%d - %s", w.Code, w.Body.String())

}
