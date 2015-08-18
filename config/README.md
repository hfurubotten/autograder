# Configuration #

When starting up Autograder the first time, a set of configuration variables
need to be set. These variables get stored in `autograder.config`.

This file contains JSON structured data with the following variables:
- Hostname
- OAuthID
- OAuthSecret
- BasePath

Example file:
```javascript
{
  "Hostname": "http://example.com",
  "OAuthID": "123456789",
  "OAuthSecret": "123456789abcdef",
  "BasePath": "/usr/share/autograder/"
}
```

The standard path to where the configuration file is stored is
`/usr/share/autograder/`, but a custom configuration file can be loaded using
the flag `-config="/path/to/config.json"` when starting the application. This
custom file need to contain the mentioned variables to be accepted. 
