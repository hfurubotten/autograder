package main

import (
	"github.com/hfurubotten/autograder/web"
	"log"
)

func main() {

	log.Println("Server starting")

	server := web.NewWebServer(80)
	server.Start()
}
