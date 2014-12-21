package web

import (
	"log"
	"net/http"

	ci "github.com/hfurubotten/autograder/ci"
	"github.com/hfurubotten/autograder/git"
)

func webhookeventhandler(w http.ResponseWriter, r *http.Request) {
	payload, err := git.DecodeHookPayload(r.Body)
	if err != nil {
		log.Println("Error: ", err)
	}

	go ci.StartTesterDeamon(payload)

	log.Println(payload)
}
