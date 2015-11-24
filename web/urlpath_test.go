package web

import (
	"net/http"
	"testing"
)

func TestLastPathElem(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	lelm := lastPathElem(r)
	if lelm != "" {
		t.Error("expected empty string, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "what", nil)
	lelm = lastPathElem(r)
	if lelm != "what" {
		t.Error("expected: what, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "/what", nil)
	lelm = lastPathElem(r)
	if lelm != "what" {
		t.Error("expected: what, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "where/and/what", nil)
	lelm = lastPathElem(r)
	if lelm != "what" {
		t.Error("expected: what, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "/where/and/what/", nil)
	lelm = lastPathElem(r)
	if lelm != "" {
		t.Error("expected empty string, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "http://what.com", nil)
	lelm = lastPathElem(r)
	if lelm != "" {
		t.Error("expected empty string, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "http://what.com/", nil)
	lelm = lastPathElem(r)
	if lelm != "" {
		t.Error("expected empty string, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "http://what.com/where", nil)
	lelm = lastPathElem(r)
	if lelm != "where" {
		t.Error("expected: where, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "http://what.com/what/is/where", nil)
	lelm = lastPathElem(r)
	if lelm != "where" {
		t.Error("expected: where, got: " + lelm)
	}
	r, _ = http.NewRequest("GET", "http://what.com/what/is/where/", nil)
	lelm = lastPathElem(r)
	if lelm != "" {
		t.Error("expected empty string, got: " + lelm)
	}
}
