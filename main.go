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
	admin         = flag.String("admin", "", "Sets up a admin user up agains the system. The value has to be a valid Github username.")
	hostname      = flag.String("hostname", "", "Give the hostname for the autogradersystem.")
	client_ID     = flag.String("clientid", "", "The application ID used in the OAuth process against Github. This can be generated at your settings page at Github.")
	client_secret = flag.String("secret", "", "The secret application code used in the OAuth process against Github. This can be generated at your settings page at Github.")
	help          = flag.Bool("help", false, "List the startup options for the autograder.")
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

	if *client_ID != "" && *client_secret != "" {
		optionstore.WriteGob("OAuthID", *client_ID)
		optionstore.WriteGob("OAuthSecret", *client_secret)
		global.OAuth_ClientID = *client_ID
		global.OAuth_ClientSecret = *client_secret
	} else {
		if !optionstore.Has("OAuthID") && !optionstore.Has("OAuthSecret") {
			log.Fatal("Missing OAuth details, set this the first time you start the system. See help pages on how to do this.")
		}

		var id string
		var secret string
		err = optionstore.ReadGob("OAuthID", &id, false)
		if err != nil {
			log.Fatal(err)
		}

		err = optionstore.ReadGob("OAuthSecret", &secret, false)
		if err != nil {
			log.Fatal(err)
		}

		global.OAuth_ClientID = id
		global.OAuth_ClientSecret = secret
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
