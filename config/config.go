package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// current holds the current configuration of the running system.
// It is global state and can only be initialized once.
var current struct {
	*Configuration
	sync.Once
}

// Configuration contains the necessary configuration data for the system.
type Configuration struct {
	URL               string `json:",omitempty"`
	OAuthClientID     string `json:",omitempty"`
	OAuthClientSecret string `json:",omitempty"`
	BasePath          string `json:",omitempty"`
}

// Get returns a copy of the current configuration.
func Get() Configuration {
	if current.Configuration == nil {
		panic("current configuration has not been set")
	}
	return *current.Configuration
}

// NewConfig creates a new configuration object.
func NewConfig(url, clientID, clientSecret, path string) (*Configuration, error) {
	conf := &Configuration{
		URL:               url,
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		BasePath:          path,
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return conf, nil
}

// Load loads the configuration file in the provided directory.
func Load(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(filepath.Join(path, FileName))
	if err != nil {
		return nil, err
	}
	conf := new(Configuration)
	err = json.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return conf, nil
}

// SetCurrent sets the current configuration. After this it cannot be set again.
func (c *Configuration) SetCurrent() {
	current.Do(func() {
		current.Configuration = c
	})
	if current.Configuration != c {
		panic("current configuration cannot be set again")
	}
}

// validate does rudimentary checks on the configuration object's data.
func (c *Configuration) validate() error {
	// remove extra slash at end of URL and basepath.
	c.URL = strings.TrimSuffix(c.URL, "/")
	c.BasePath = strings.TrimSuffix(c.BasePath, "/")
	if c.URL == "" {
		return errors.New("homepage url is required")
	}
	if !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		return errors.New("homepage url is not a valid url")
	}
	if strings.Count(c.URL, "/") > 2 {
		return errors.New("homepage url cannot contain path elements")
	}
	if c.OAuthClientID == "" {
		return errors.New("clientID is required")
	}
	if c.OAuthClientSecret == "" {
		return errors.New("clientSecret is required")
	}
	if c.BasePath == "" {
		return errors.New("basepath is required")
	}
	return nil
}

// Save saves the configuration file in basepath.
func (c *Configuration) Save() error {
	info, err := os.Stat(c.BasePath)
	if err != nil {
		err := os.MkdirAll(c.BasePath, 0700)
		if err != nil {
			return err
		}
		log.Printf("Created %s directory: %s", SysName, c.BasePath)
	} else if !info.IsDir() {
		return errors.New("basepath is not a directory")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	confFile := filepath.Join(c.BasePath, FileName)
	err = ioutil.WriteFile(confFile, data, 0600)
	if err == nil {
		log.Printf("Saved %s configuration file in: %s", SysName, confFile)
	}
	return err
}
