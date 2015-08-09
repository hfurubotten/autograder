# Installation #

## Installing from source ##
When installing from source you need to have Go compiler installed and a GOPATH
set up. When cloning the Autograder repository, the dependent libraries need to
be cloned as well.

Follow these steps to complile and run Autograder:
- Run the `go get` command on the main repository to clone the Autograder repository.

	   go get github.com/hfurubotten/autograder

     This will also download all dependent libraries.
- This process will also compile the source code into a runnable file. This file
  can be found at `$GOPATH/bin/autograder`.
- Set the go bin folder as current working directory. Run `cd $GOPATH/bin/`
- Run Autograder with "sudo ./autograder". Add the needed first time time
  configurations described below as flags after the command to set its behavior.
- Optional: To let the autograder application to run while the terminal window
  is closed, it can be opened in a screen session. Run the command
  `screen -S autograder sudo ./autograder`. Disconnect from the screen session
  with `ctrl+a, d`.

## Updating from source ##
- Close down the current running Autograder instance.
- Run the command

      go get -u github.com/hfurubotten/autograder

    in the first step in the installation and go through the steps again.

## Configuration ##

The binary file can take a number of flags to configure its behavior. These
configurations only need to be set at first start up, as the system remember
last configuration given.

	-admin="": Sets up an admin user in the system. The value has to be a valid Github username.
	-clientid="": The application ID used in the OAuth process against Github. This can be generated at your settings page at Github.
	-domain="": Give the domain name for the autogradersystem.
	-help=false: List the startup options for the autograder.
	-secret="": The secret application code used in the OAuth process against Github. This can be generated at your settings page at Github.

### Github Application codes ###

To register a new application at GitHub, go to this address to generate OAuth
tokens: [https://github.com/settings/applications/new]

If you already have OAuth codes, you can find then on this address:
[https://github.com/settings/applications]

Homepage URL is the domain name you are using to serve Autograder.
The Authorization callback URL is your domain name with the path /oauth.
(http://example.com/oauth)
