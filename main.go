package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/hfurubotten/autograder/config"
	"github.com/hfurubotten/autograder/database"
	git "github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web"
)

var (
	admin        = flag.String("admin", "", "Sets up a admin user up agains the system. The value has to be a valid Github username.")
	hostname     = flag.String("domain", "", "Give the domain name for the autogradersystem.")
	clientID     = flag.String("clientid", "", "The application ID used in the OAuth process against Github. This can be generated at your settings page at Github.")
	clientSecret = flag.String("secret", "", "The secret application code used in the OAuth process against Github. This can be generated at your settings page at Github.")
	help         = flag.Bool("help", false, "List the startup options for the autograder.")
	configfile   = flag.String("configfile", "", "Path to a custom config file location. Used when a config file not stored in the standard file location is prefered.")
	basepath     = flag.String("basepath", "", "A custom file path for storing autograder files.")
)

func main() {
	// enables multi core use.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse flags
	flag.Parse()

	// prints the available flags to use on start
	if *help {
		flag.Usage()
		fmt.Println("First start up details:")
		fmt.Println("First time you start the system you need to supply OAuth details, domain name and an admin.")
		fmt.Println("To register a new application at GitHub, go to this address to generate OAuth tokens: https://github.com/settings/applications/new")
		fmt.Println("If you already have OAuth codes, you can find then on this address: https://github.com/settings/applications")
		fmt.Println("The Homepage URL is the domain name you are using to serve the system.")
		fmt.Println("The Authorization callback URL is your domainname with the path /oauth. (http://example.com/oauth)")
		return
	}

	// loads config file either from custom path or standard file path and validates.
	var conf *config.Configuration
	var err error
	if *configfile != "" {
		conf, err = config.LoadConfigFile(*configfile)
		if err != nil {
			log.Fatal(err)
		}
	} else if *basepath != "" {
		conf, err = config.LoadConfigFile(*basepath + config.ConfigFileName)
		if err != nil {
			log.Fatal(err)
		}

		conf.BasePath = *basepath
	} else {
		conf, err = config.LoadStandardConfigFile()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Updates config with evt. new information

	// checks for a domain name
	if *hostname != "" {
		conf.Hostname = *hostname
	}

	// checks for the application codes to GitHub
	if *clientID != "" && *clientSecret != "" {
		conf.OAuthID = *clientID
		conf.OAuthSecret = *clientSecret
	}

	// validates the configurations
	if conf.Validate() != nil {
		if err := conf.QuickFix(); err != nil {
			log.Fatal(err)
		}
	}

	conf.ExportToGlobalVars()

	// saves configurations
	if err := conf.Save(); err != nil {
		log.Fatal(err)
	}

	// starting database
	database.Start(conf.BasePath + "autograder.db")
	defer database.Close()

	// checks for an admin username
	if *admin != "" {
		log.Println("New admin added to the system: ", *admin)
		m, err := git.GetMember(*admin)
		if err != nil {
			log.Fatal(err)
		}

		m.IsAdmin = true
		err = m.Save()
		if err != nil {
			m.Unlock()
			log.Println("Couldn't store admin user in system:", err)
		}
	}

	// TODO: checks if the system should be set up as a deamon that starts on system startup.

	// TODO: checks for docker installation
	// TODO: install on supported systems
	// TODO: give notice for those systems not supported

	// log print appearance
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// starts up the webserver
	log.Println("Server starting")

	server := web.NewServer(80)
	server.Start()

	// Prevent main from returning immediately. Wait for interrupt.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	<-signalChan
	log.Println("Application closed by user.")
}
