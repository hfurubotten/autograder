package main

import (
	"flag"
	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web"
	"github.com/hfurubotten/diskv"
	"log"
)

var (
	admin    = flag.String("admin", "", "Sets up a admin user up agains the system. The value has to be a valid Github username.")
	hostname = flag.String("hostname", "", "Give the hostname for the autogradersystem.")
	help     = flag.Bool("help", false, "List the startup options for the autograder.")
)

var optionstore = diskv.New(diskv.Options{
	BasePath:     "diskv/options/",
	CacheSizeMax: 1024 * 1024 * 256,
})

func main() {
	var err error
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *hostname != "" {
		optionstore.WriteGob("hostname", *hostname)
		global.Hostname = *hostname
	} else {
		if !optionstore.Has("hostname") {
			log.Fatal("Missing hostname, set this the first time you start the system.")
		}

		var hname string
		err = optionstore.ReadGob("hostname", &hname, false)
		if err != nil {
			log.Fatal(err)
		}

		global.Hostname = hname
	}

	if *admin != "" {
		log.Println("New admin added to the system: ", *admin)
		m := git.NewMemberFromUsername(*admin)
		m.IsAdmin = true
		err = m.StickToSystem()
		if err != nil {
			log.Println("Couldn't store admin user in system:", err)
		}
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Server starting")

	server := web.NewWebServer(80)
	server.Start()
}
