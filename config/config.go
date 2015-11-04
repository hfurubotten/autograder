package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
func NewConfig(hostname, oauthid, oauthsecret, path string) (*Configuration, error) {
	return &Configuration{
		Hostname:    hostname,
		OAuthID:     oauthid,
		OAuthSecret: oauthsecret,
		BasePath:    path,
	}, nil
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
	return conf, nil
}

// ExportToGlobalVars will export the configuration ot the global variables.
func (c *Configuration) ExportToGlobalVars() {
	global.Hostname = c.Hostname
	global.OAuthClientID = c.OAuthID
	global.OAuthClientSecret = c.OAuthSecret
}

// Validate will try to validate the information in the configuration. Returns
// error if the information cant be validated.
func (c *Configuration) Validate() error {
	if !strings.HasPrefix(c.Hostname, "http://") && !strings.HasPrefix(c.Hostname, "https://") {
		return errors.New("The domain url is not a valid url")
	}

	if strings.Count(c.Hostname, "/") > 2 {
		return errors.New("The hostename contains too many elements")
	}

	if c.OAuthID == "" {
		return errors.New("Missing OAuth ID")
	}

	if c.OAuthSecret == "" {
		return errors.New("Missing OAuth Secret hash")
	}

	if c.BasePath == "" {
		return errors.New("Missing basepath for storing support files")
	}

	return nil
}

// QuickFix will try to fix minor errors in the configuration. Returns error if
// it cant be fixed or cant be validated afterwards.
func (c *Configuration) QuickFix() error {
	if strings.HasSuffix(c.Hostname, "/") {
		c.Hostname = c.Hostname[:len(c.Hostname)-1]
	}
	if c.BasePath == "" {
		c.BasePath = StandardBasePath
	}
	if !strings.HasSuffix(c.BasePath, "/") {
		c.BasePath = c.BasePath + "/"
	}
	return c.Validate()
}

// Save saves the configuration file in basepath.
func (c *Configuration) Save() error {
	info, err := os.Stat(c.BasePath)
	if err != nil {
		err := os.Mkdir(c.BasePath, 0777)
		if err != nil {
			return err
		}
	} else if !info.IsDir() {
		return errors.New("basepath is not a directory")
	}

	jsondata, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(c.BasePath, ConfigFileName), jsondata, 0666)
}
