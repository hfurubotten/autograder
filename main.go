package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"text/template"

	"github.com/hfurubotten/autograder/config"
	"github.com/hfurubotten/autograder/database"
	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/web"
)

const instructions = `
The first time you start {{.SystemName}} you will need to supply a few details
about your host environment, the administrator and the git repository hosting
environment. Currently, we only support GitHub for hosting git repositories.

{{.SystemName}} can either read a configuration file with the necessary
information (see the example below), or you can provide these details as
command line arguments (also shown below).

Here is an example {{.ConfigFileName}} file:

{
  "HomepageURL": "http://example.com/",
  "ClientID": "123456789",
  "ClientSecret": "123456789abcdef",
  "BasePath": "/usr/share/{{.SystemNameLC}}/"
}

Before you can start you will need to register the {{.SystemName}} application
at GitHub; you will need to do this from the administrator account.

1. Go to https://github.com/settings/applications/new
2. Enter the information requested.
   - Application name: e.g. "{{.SystemName}} at University of Stavanger"
   - Homepage URL: e.g. "http://{{.SystemNameLC}}.ux.uis.no"
   - Authorization callback URL: e.g. "http://{{.SystemNameLC}}.ux.uis.no/oauth"

Note that, the Homepage URL must be a fully qualified URL, including http://.
This must be the hostname (or an alias) of server running the '{{.SystemNameLC}}'
program. This server must have a public IP address, since GitHub will make calls
to this server to support {{.SystemName}}'s functionality. Further, {{.SystemName}}
requires that the Authorization callback URL is the same as the Homepage URL
with the added "/oauth" path.

Once you have completed the above steps, the Client ID and Client Secret will be
available from the GitHub web interface. Simply copy each of these OAuth tokens
and paste them into the configuration file, or on the command line when starting
{{.SystemName}} for the first time. You will not need to repeat this process
when starting {{.SystemName}} in the future.

If you need to obtain the OAuth tokens at a later time, e.g. if you have deleted
the configuration file, go to: https://github.com/settings/developers and
select your Application to be able to view the OAuth tokens again.

`

var (
	admin        = flag.String("admin", "", "Admin must be a valid GitHub username")
	hostname     = flag.String("url", "", "Homepage URL for "+config.SystemName)
	clientID     = flag.String("id", "", "Client ID for OAuth with Github")
	clientSecret = flag.String("secret", "", "Client Secret for OAuth with Github")
	help         = flag.Bool("help", false, "Helpful instructions")
	path         = flag.String("basepath", config.StandardBasePath, "Path for data storage for "+config.SystemName)
)

func main() {
	flag.Parse()

	// print instructions and command usage
	if *help {
		data := struct {
			SystemName, SystemNameLC, ConfigFileName string
		}{
			config.SystemName, config.SystemNameLC, config.ConfigFileName,
		}
		t := template.Must(template.New("instructions").Parse(instructions))
		err := t.Execute(os.Stdout, data)
		if err != nil {
			log.Fatalln(err)
		}
		flag.Usage()
		return
	}

	// load configuration data from the provided base path
	var conf *config.Configuration
	var err error
	if *path != "" {
		conf, err = config.Load(*path)
		if err != nil {
			log.Println(err)
			conf, err = config.NewConfig(*hostname, *clientID, *clientSecret, *path)
			if err != nil {

			}
		}
	}

	// Updates config with new information, if available

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

	// save configuration
	if err := conf.Save(); err != nil {
		log.Fatal(err)
	}

	// start database
	if err := database.Start(conf.BasePath); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// checks for an admin username
	if *admin != "" {
		log.Println("New admin added to the system: ", *admin)
		m, err := entities.GetMember(*admin)
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

	log.Println("Starting webserver")
	server := web.NewServer(80)
	server.Start()

	// Prevent main from returning immediately. Wait for interrupt.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	<-signalChan
	log.Println("Application shutdown by user.")
}
