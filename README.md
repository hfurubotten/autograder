Autograder
==========
Autograder is an automatic feedback system for the students. It is integrated with GitHub and manages courses and students within GitHubs git management system. When students push code to their repositories, the push triggers a continuous integration process. Results form this integration process is then given to the students on their personal web pages. Teachers can then access this integration log, thus have a valuabel tool in the grading of lab assignments. 

##Features##
Listed below is some of the features in autograder

###Training in industrial grade tools###
The teaching enviorment of autograder is infact GitHub itself, thus training the students in using version controll systems to have controll over their code and assignments. Integrated in autograder is also a custom made continous integration tools. Versjon controlling and continous integration is tools widely used by the industry. Training the students in tools like git, GitHub and CI makes the students more equipt when making the transition to working life.

###Automatic assignment testing###

###Code reviewing###

###Awardsystem for online discussions###

##Installation##

While setting up this system everything is contained within the binary file compiled from this source code. Fastest way to get the source code is to run this command: 

	go get github.com/hfurubotten/autograder

This will also download all dependent libraries mentioned below.  

###Supported Operating systems###

Autograder has been tested on and support following operating systems:

- Ubuntu

###Configuration###

The binary file can take a number of flags to configure its behavior. These configurations only need to be set at first start up, as the system remember last configuration given. 

	-admin="": Sets up an admin user in the system. The value has to be a valid Github username.
	-clientid="": The application ID used in the OAuth process against Github. This can be generated at your settings page at Github.
	-domain="": Give the domain name for the autogradersystem.
	-help=false: List the startup options for the autograder.
	-secret="": The secret application code used in the OAuth process against Github. This can be generated at your settings page at Github.

###Github Application codes###

To register a new application at GitHub, go to this address to generate OAuth tokens: [https://github.com/settings/applications/new]
If you already have OAuth codes, you can find then on this address: [https://github.com/settings/applications]
The Homepage URL is the domain name you are using to serve the system.
The Authorization callback URL is your domainname with the path /oauth. (http://example.com/oauth)


##Dependencies##

When compiling these following libraies need to be included;
- [goauth2][]
- [go-github][]
- [go-dockerclient][]
- [diskv][]

The runtime of the test enviorment are virtualized in [docker].

[goauth2]: https://code.google.com/p/goauth2/
[go-github]: https://github.com/google/go-github
[docker]: https://www.docker.com/
[go-dockerclient]: https://github.com/fsouza/go-dockerclient
[diskv]: https://github.com/hfurubotten/diskv
