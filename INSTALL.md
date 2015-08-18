# Pre-requisits #

To install autograder, you'll need to install Go and set the `GOPATH`
environment variable appropriately. See [here](http://golang.org/doc/install).

# Installation #

To install autograder, complete these steps:

    go get github.com/hfurubotten/autograder

This will produce an executable file in `$GOPATH/bin/autograder`.
Note that this will also pull in several libraries on which autograder depends.

## Running (first time: configure) ##

1. `cd $GOPATH/bin/`
2. Run `sudo autograder`.
3. Optional (replaces step 2): `screen -S autograder sudo autograder`. This lets you run the autograder in the background allowing you to close the terminal window. To disconnect from the screen session, use `ctrl+a, d`.

## Running (configuration completed) ##

1. `cd $GOPATH/bin/`
2. Run `sudo autograder`.
3. Optional (replaces step 2): `screen -S autograder sudo autograder`. This lets you run the autograder in the background allowing you to close the terminal window. To disconnect from the screen session, use `ctrl+a, d`.

## Upgrading ##

1. Shut down the currently running autograder instance.
2. Run the command

    go get -u github.com/hfurubotten/autograder



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
