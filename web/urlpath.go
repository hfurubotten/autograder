package web

import (
	"net/http"
	"strings"
)

func lastPathElem(r *http.Request) string {
	pathElms := strings.Split(r.URL.Path, "/")
	return pathElms[len(pathElms)-1]
}
