# Pre-requisits #

To install autograder, you'll need to install Go and set the `GOPATH`
environment variable appropriately. See [here](http://golang.org/doc/install).

# Installation #

To install autograder, complete these steps:

    go get github.com/hfurubotten/autograder

This will produce an executable file in `$GOPATH/bin/autograder`.
Note that this will also pull in several libraries on which autograder depends.

# First time use #

The first time you start Autograder you will need to supply a few details
about your host environment, the administrator and the git repository hosting
environment. Currently, we only support GitHub for hosting git repositories.

Autograder can either read a configuration file with the necessary
information (see the example below), or you can provide these details as
command line arguments (also shown below).

Here is an example autograder.config file:

{
  "HomepageURL": "http://example.com/",
  "ClientID": "123456789",
  "ClientSecret": "123456789abcdef",
  "BasePath": "/usr/share/autograder/"
}

Before you can start you will need to register the Autograder application
at GitHub; you will need to do this from the administrator account.

1. Go to [https://github.com/settings/applications/new]
2. Enter the information requested.
   - Application name: e.g. "Autograder at University of Stavanger"
   - Homepage URL: e.g. "http://autograder.ux.uis.no"
   - Authorization callback URL: e.g. "http://autograder.ux.uis.no/oauth"

Note that, the Homepage URL must be a fully qualified URL, including http://.
This must be the hostname (or an alias) of server running the 'autograder'
program. This server must have a public IP address, since GitHub will make calls
to this server to support Autograder's functionality. Further, Autograder
requires that the Authorization callback URL is the same as the Homepage URL
with the added "/oauth" path.

Once you have completed the above steps, the Client ID and Client Secret will be
available from the GitHub web interface. Simply copy each of these OAuth tokens
and paste them into the configuration file, or on the command line when starting
Autograder for the first time. You will not need to repeat this process
when starting Autograder in the future.

If you need to obtain the OAuth tokens at a later time, e.g. if you have deleted
the configuration file, go to: [https://github.com/settings/developers] and
select your Application to be able to view the OAuth tokens again.

## Running (first time: configure) ##

1. `cd $GOPATH/bin/`

2. Run `sudo ./autograder -admin=<githubusername> -id=<Client ID> -secret=<Client Secret> -url=<http://your.domain.com>`. Insert your system details after the equal sign.

3. Optional (replaces step 2): You can configure the details in a config file and
  run `sudo ./autograder -config=/path/to/config.json`. This is explained in
  more details in the [config package](https://github.com/hfurubotten/autograder/tree/master/config).

4. Optional (replaces step 2): `screen -S autograder sudo autograder`. This lets
  you run the autograder in the background allowing you to close the terminal
  window. To disconnect from the screen session, use `ctrl+a, d`.

## Running (configuration completed) ##

1. `cd $GOPATH/bin/`

2. Run `sudo autograder`.

3. Optional (replaces step 2): `screen -S autograder sudo ./autograder`.
   This lets you run the autograder in the background allowing you to close the
   terminal window. To disconnect from the screen session, use `ctrl+a, d`.

## Upgrading ##

Shut down the currently running autograder instance, and run the command:

    go get -u github.com/hfurubotten/autograder

Restart according to instructions above.

## Configuration ##

The autograder executable takes a number of command line arguments, which is
necessary to configure its behavior. These configuration parameters are only
need during installation, and will be remembered across updates and restarts.

Here is the Usage of autograder:

```
  -admin string
    	Admin must be a valid GitHub username
  -basepath string
    	Path for data storage for Autograder
  -config string
    	Path to a custom config file
  -help
    	Helpful instructions
  -id string
    	Client ID for OAuth with Github
  -secret string
    	Client Secret for OAuth with Github
  -url string
    	Homepage URL for Autograder
```
