package main

import (
	"flag"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/web"
	"log"
)

var (
	admin = flag.String("admin", "", "Sets up a admin user up agains the system. The value has to be a valid Github username.")
)

func main() {
	flag.Parse()

	if *admin != "" {
		log.Println("New admin added to the system: ", *admin)
		m := git.NewMemberFromUsername(*admin)
		m.IsAdmin = true
		err := m.StickToSystem()
		if err != nil {
			log.Println("Couldn't store admin user in system:", err)
		}
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Server starting")

	server := web.NewWebServer(80)
	server.Start()
}
