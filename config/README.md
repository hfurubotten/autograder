# Configuration #

When starting up Autograder the first time, a set of configuration variables
need to be set. These variables get stored in a JSON formatted file named
`autograder.config`.

Here is an example configuration file:
```json
{
  "HomepageURL": "http://example.com/",
  "ClientID": "123456789",
  "ClientSecret": "123456789abcdef",
  "BasePath": "/usr/share/autograder/"
}
```

The configuration file is stored in the base path.
