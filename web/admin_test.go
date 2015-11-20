package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"testing"

	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web/pages"
	"github.com/hfurubotten/autograder/web/sessions"
)

func TestAdminSetterNotAdmin(t *testing.T) {
	admin := "meling"
	pv := url.Values{"user": {admin}, "admin": {"true"}}
	r, err := http.NewRequest("POST", "http://example.com/", bytes.NewBuffer([]byte(pv.Encode())))
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	SetAdminHandler(w, r)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(ErrNotAdmin)
	expected := buf.String()
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected: %d, got: %d", http.StatusOK, w.Code)
	}
}

func TestAdminSetterIncompleteMember(t *testing.T) {
	admin := "meling"
	accessToken := "dummy-token1"
	scope := "scope"

	u := entities.NewUserProfile(accessToken, admin, scope)
	m := entities.NewMember(u)
	err := entities.PutMember(accessToken, m)
	if err != nil {
		t.Error(err)
	}

	m.IsAdmin = true
	err = m.Save()
	if err != nil {
		t.Error(err)
	}

	pv := url.Values{"user": {admin}, "admin": {"true"}}
	r, err := http.NewRequest("POST", "http://example.com/", bytes.NewBuffer([]byte(pv.Encode())))
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// mark auth session as approved
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
	// save the access token for this session
	sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.Home)

	SetAdminHandler(w, r)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(ErrNotAdmin)
	expected := buf.String()
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected: %d, got: %d", http.StatusTemporaryRedirect, w.Code)
	}
}

func TestAdminSetterIsAdmin(t *testing.T) {
	admin := "meling"
	accessToken := "dummy-token2"
	scope := "scope"

	u := entities.NewUserProfile(accessToken, admin, scope)
	m := entities.NewMember(u)
	err := entities.PutMember(accessToken, m)
	if err != nil {
		t.Error(err)
	}

	m.IsAdmin = true
	m.Name = "Hein Meling"
	m.StudentID = 33847
	m.Email, err = mail.ParseAddress("James <alice@example.com>")
	err = m.Save()
	if err != nil {
		t.Error(err)
	}

	pv := url.Values{"user": {admin}, "admin": {"true"}}
	r, err := http.NewRequest("POST", "http://example.com/", bytes.NewBuffer([]byte(pv.Encode())))
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// mark auth session as approved
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
	// save the access token for this session
	sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.Home)

	SetAdminHandler(w, r)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	msg := SetAdminView{
		User:  m.Username,
		Admin: m.IsAdmin,
	}
	err = enc.Encode(msg)
	expected := buf.String()
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected: %d, got: %d", http.StatusTemporaryRedirect, w.Code)
	}
}

func TestAdminSetterMissingField1(t *testing.T) {
	admin := "meling"
	accessToken := "dummy-token3"
	scope := "scope"

	u := entities.NewUserProfile(accessToken, admin, scope)
	m := entities.NewMember(u)
	err := entities.PutMember(accessToken, m)
	if err != nil {
		t.Error(err)
	}

	m.IsAdmin = true
	m.Name = "Hein Meling"
	m.StudentID = 33847
	m.Email, err = mail.ParseAddress("James <alice@example.com>")
	err = m.Save()
	if err != nil {
		t.Error(err)
	}

	pv := url.Values{"user": {admin}}
	r, err := http.NewRequest("POST", "http://example.com/", bytes.NewBuffer([]byte(pv.Encode())))
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// mark auth session as approved
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
	// save the access token for this session
	sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.Home)

	SetAdminHandler(w, r)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(ErrMissingField)
	expected := buf.String()
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected: %d, got: %d", http.StatusTemporaryRedirect, w.Code)
	}
}

func TestAdminSetterMissingField2(t *testing.T) {
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
	m.Name = "Hein Meling"
	m.StudentID = 33847
	m.Email, err = mail.ParseAddress("James <alice@example.com>")
	err = m.Save()
	if err != nil {
		t.Error(err)
	}

	pv := url.Values{"fakeuser": {admin}}
	r, err := http.NewRequest("POST", "http://example.com/", bytes.NewBuffer([]byte(pv.Encode())))
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	// mark auth session as approved
	sessions.SetSessions(w, r, sessions.AuthSession, sessions.ApprovedSessionKey, true)
	// save the access token for this session
	sessions.SetSessionsAndRedirect(w, r, sessions.AuthSession, sessions.AccessTokenSessionKey, accessToken, pages.Home)

	SetAdminHandler(w, r)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(ErrMissingField)
	expected := buf.String()
	got := w.Body.String()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected: %d, got: %d", http.StatusTemporaryRedirect, w.Code)
	}
}
