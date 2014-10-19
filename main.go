package main

import (
	"github.com/hfurubotten/autograder/web"
)

func main() {

	server := web.NewWebServer(8082)
	server.Start()
}
