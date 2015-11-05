package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hfurubotten/autograder/global"
)

// ConfigFileName is the default file name for the JSON configuration file.
var ConfigFileName = "config"

// Configuration contains the necessary configuration data for the system.
type Configuration struct {
	Hostname    string `json:",omitempty"`
	OAuthID     string `json:",omitempty"`
	OAuthSecret string `json:",omitempty"`
	BasePath    string `json:",omitempty"`
}

// NewConfig creates a new configuration object.
func NewConfig(url, clientID, clientSecret, path string) (*Configuration, error) {
	conf := &Configuration{
		Hostname:    url,
		OAuthID:     clientID,
		OAuthSecret: clientSecret,
		BasePath:    path,
	}
	conf.quickFix()
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return conf, nil
}

// Load loads the configuration file in the provided directory.
func Load(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(filepath.Join(path, ConfigFileName))
	if err != nil {
		return nil, err
	}
	conf := new(Configuration)
	err = json.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}
	conf.quickFix()
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return conf, nil
}

// ExportToGlobalVars will export the configuration ot the global variables.
func (c *Configuration) ExportToGlobalVars() {
	global.Hostname = c.Hostname
	global.OAuthClientID = c.OAuthID
	global.OAuthClientSecret = c.OAuthSecret
}

// validate does rudimentary checks on the configuration object's data.
func (c *Configuration) validate() error {
	if c.Hostname == "" {
		return errors.New("homepage url is required")
	}
	if !strings.HasPrefix(c.Hostname, "http://") && !strings.HasPrefix(c.Hostname, "https://") {
		return errors.New("homepage url is not a valid url")
	}
	if strings.Count(c.Hostname, "/") > 2 {
		return errors.New("homepage url cannot contain path elements")
	}
	if c.OAuthID == "" {
		return errors.New("clientID is required")
	}
	if c.OAuthSecret == "" {
		return errors.New("clientSecret is required")
	}
	if c.BasePath == "" {
		return errors.New("basepath is required")
	}
	return nil
}

// quickFix removes extra slash at end of URL and basepath.
func (c *Configuration) quickFix() {
	c.Hostname = strings.TrimSuffix(c.Hostname, "/")
	c.BasePath = strings.TrimSuffix(c.BasePath, "/")
}

// Save saves the configuration file in basepath.
func (c *Configuration) Save() error {
	info, err := os.Stat(c.BasePath)
	if err != nil {
		err := os.MkdirAll(c.BasePath, 0700)
		if err != nil {
			return err
		}
		log.Printf("Created %s directory: %s", SystemName, c.BasePath)
	} else if !info.IsDir() {
		return errors.New("basepath is not a directory")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(c.BasePath, ConfigFileName), data, 0600)
	if err == nil {
		log.Printf("Saved %s configuration file in: %s", SystemName, c.BasePath)
	}
	return err
}
