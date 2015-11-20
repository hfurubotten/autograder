package web

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func logServerError(w http.ResponseWriter, err error) {
	errCode := rand.Uint32()
	errMsg := fmt.Sprintf("InternalServerError(%d)", errCode)
	http.Error(w, errMsg, http.StatusInternalServerError)
	log.Printf("%s: %s\n", errMsg, err)
}

func logNotFoundError(w http.ResponseWriter, err error) {
	errCode := rand.Uint32()
	errMsg := fmt.Sprintf("NotFound(%d)", errCode)
	http.Error(w, errMsg, http.StatusNotFound)
	log.Printf("%s: %s\n", errMsg, err)
}

func logErrorAndRedirect(w http.ResponseWriter, r *http.Request, redirectTo string, err error) {
	log.Printf("Redirecting due to: %v\n", err)
	http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
}

func logAndRedirect(w http.ResponseWriter, r *http.Request, redirectTo, msg string) {
	log.Println(msg)
	http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
}
