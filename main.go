package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hfurubotten/autograder/git"
	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/autograder/web"
	"github.com/hfurubotten/diskv"
)

var (
	admin         = flag.String("admin", "", "Sets up a admin user up agains the system. The value has to be a valid Github username.")
	hostname      = flag.String("domain", "", "Give the domain name for the autogradersystem.")
	client_ID     = flag.String("clientid", "", "The application ID used in the OAuth process against Github. This can be generated at your settings page at Github.")
	client_secret = flag.String("secret", "", "The secret application code used in the OAuth process against Github. This can be generated at your settings page at Github.")
	help          = flag.Bool("help", false, "List the startup options for the autograder.")
)

var optionstore = diskv.New(diskv.Options{
	BasePath:     "diskv/options/",
	CacheSizeMax: 1024 * 1024 * 256,
})

func main() {
	// enables multi core use.
	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	flag.Parse()

	// prints the available flags to use on start
	if *help {
		flag.Usage()
		return
	}

	// checks for a domain name
	if *hostname != "" {
		if !strings.HasPrefix(*hostname, "http://") && !strings.HasPrefix(*hostname, "https://") {
			log.Fatal("The domain url is not a valid url.")
		}

		domain := *hostname

		if strings.HasSuffix(domain, "/") {
			domain = domain[:len(domain)-1]
		}

		optionstore.WriteGob("hostname", domain)
		global.Hostname = domain
	} else {
		if !optionstore.Has("hostname") {
			log.Fatal("Missing domain name, set this the first time you start the system.")
		}

		var hname string
		err = optionstore.ReadGob("hostname", &hname, false)
		if err != nil {
			log.Fatal(err)
		}

		global.Hostname = hname
	}

	// checks for the application codes to GitHub
	if *client_ID != "" && *client_secret != "" {
		optionstore.WriteGob("OAuthID", *client_ID)
		optionstore.WriteGob("OAuthSecret", *client_secret)
		global.OAuth_ClientID = *client_ID
		global.OAuth_ClientSecret = *client_secret
	} else {
		if !optionstore.Has("OAuthID") && !optionstore.Has("OAuthSecret") {
			log.Println("Missing OAuth details, set this the first time you start the system.")
			log.Println("To register a new application at GitHub, go to this address to generate OAuth tokens: https://github.com/settings/applications/new")
			log.Println("If you already have OAuth codes, you can find then on this address: https://github.com/settings/applications")
			log.Println("The Homepage URL is the domain name you are using to serve the system.")
			log.Fatal("The Authorization callback URL is your domainname with the path /oauth. (http://example.com/oauth)")

			// stop := make(chan int)

			// go web.FakeServer(80, stop)

			// fmt.Print("OAuth ID: ")
			// scanner := bufio.NewScanner(os.Stdin)
			// scanner.Scan()
			// id = strings.TrimSpace(scanner.Text())

			// fmt.Print("OAuth secret: ")
			// scanner = bufio.NewScanner(os.Stdin)
			// scanner.Scan()
			// secret = strings.TrimSpace(scanner.Text())

			// stop <- 1

			// //store them
			// optionstore.WriteGob("OAuthID", id)
			// optionstore.WriteGob("OAuthSecret", secret)
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

	// checks for an admin username
	if *admin != "" {
		log.Println("New admin added to the system: ", *admin)
		m, err := git.NewMemberFromUsername(*admin)
		if err != nil {
			log.Fatal(err)
		}

		m.Lock()
		defer m.Unlock()

		m.IsAdmin = true
		err = m.Save()
		if err != nil {
			log.Println("Couldn't store admin user in system:", err)
		}
	}

	// checks if the system should be set up as a deamon that starts on system startup.

	// checks for docker installation
	// install on supported systems
	// give notice for those systems not supported

	// determins the path to additional files
	execdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println("Couldn't determin the path to the executable.")
		log.Fatal(err)
	}
	dir_one := filepath.Join(execdir, "../src/github.com/hfurubotten/autograder/")
	dir_two := filepath.Join(execdir, "/")
	if info, err := os.Stat(dir_one); err == nil {
		if info.Mode().IsDir() {
			global.Basepath = dir_one + "/"
		} else {
			log.Fatal("Path found to source files is not a directory.")
		}

	} else if info, err := os.Stat(dir_two); err == nil {
		if info.Mode().IsDir() {
			global.Basepath = dir_two + "/"
		} else {
			log.Fatal("Path found to source files is not a directory.")
		}
	} else {
		log.Println("Couldn't determin the path ")
		log.Fatal("")
	}

	// log print appearance
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// starts up the webserver
	log.Println("Server starting")

	server := web.NewWebServer(80)
	server.Start()
}
