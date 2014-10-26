package main

import (
	"github.com/hfurubotten/autograder/web"
	"log"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Server starting")

	server := web.NewWebServer(80)
	server.Start()
}
