package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hfurubotten/autograder/global"
)

func TestMain(m *testing.M) {
	StandardBasePath = "testfiles"
	err := os.Mkdir(StandardBasePath, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	m.Run()
	err = os.RemoveAll(StandardBasePath)
	if err != nil {
		log.Println("Unable to remove configuration files after test")
	}
}

var testNewConfigInput = []struct {
	url, id, secret, path string
}{
	{"http://example.com", "1234", "abcd", "/tmp"},
	{"http://example2.com", "123456789", "abcdef", "/tmp"},
	{"http://example3.com", "987654321", "abcd123456789", "/tmp"},
	{"http://example4.com", "1234acd544", "abcd455813aa", "/tmp"},
	{"http://example5.com", "123accd224", "422abcd54adcb4442272882adedff35f3fe3", "/usr/share/ag"},
}

func TestNewConfig(t *testing.T) {
	for _, in := range testNewConfigInput {
		conf, err := NewConfig(in.url, in.id, in.secret, in.path)
		if err != nil {
			t.Error(err)
			continue
		}

		if conf.Hostname != in.url {
			t.Errorf("Field value Hostname does not match. %v != %v", conf.Hostname, in.url)
		}
		if conf.OAuthID != in.id {
			t.Errorf("Field value OAuthID does not match. %v != %v", conf.OAuthID, in.id)
		}
		if conf.OAuthSecret != in.secret {
			t.Errorf("Field value OAuthSecret does not match. %v != %v", conf.OAuthSecret, in.secret)
		}
		if conf.BasePath != in.path {
			t.Errorf("BasePath does not match. %v != %v", conf.BasePath, in.path)
		}
	}
}

var testNewConfigFailInput = []struct {
	url, id, secret, path string
}{
	{"", "1234", "abcd", "/tmp"},
	{"http://example2.com", "", "abcdef", "/tmp"},
	{"http://example3.com", "987654321", "", "/tmp"},
	{"http://example4.com", "1234acd544", "abcd455813aa", ""},
	{"http://example5.com", "123accd224", "", "/usr/share/ag"},
	{"", "", "", "/usr/share/ag"},
	{"", "", "", ""},
}

func TestNewConfigFail(t *testing.T) {
	for _, in := range testNewConfigFailInput {
		conf, err := NewConfig(in.url, in.id, in.secret, in.path)
		if conf != nil {
			// should not happen
			t.Errorf("Expected <nil>, got: %v", conf)
			continue
		}
		if err == nil {
			// should not happen
			t.Error("Expected non-nil error, didn't get any errors")
			continue
		}
	}
}

var testLoadStandardConfigFileInput = []struct {
	filedata string
	conf     *Configuration
}{
	{
		filedata: "{\n" +
			"\"Hostname\":\"http://example.com\",\n" +
			"\"OAuthID\":\"1234\",\n" +
			"\"OAuthSecret\":\"abcd\",\n" +
			"\"BasePath\":\"/example/\"\n" +
			"}",
		conf: &Configuration{
			Hostname:    "http://example.com",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
	{
		filedata: "{\n" +
			"\"Hostname\":\"http://example2.com\",\n" +
			"\"OAuthID\":\"123456789\",\n" +
			"\"OAuthSecret\":\"abcdef123456789\",\n" +
			"\"BasePath\":\"/usr/\"\n" +
			"}",
		conf: &Configuration{
			Hostname:    "http://example2.com",
			OAuthID:     "123456789",
			OAuthSecret: "abcdef123456789",
			BasePath:    "/usr/",
		},
	},
	{
		filedata: "{\n" +
			"\"Hostname\":\"http://example3.com\",\n" +
			"\"OAuthID\":\"123454685139\",\n" +
			"\"OAuthSecret\":\"abcdef123454accd58be5f5ee6789\",\n" +
			"\"BasePath\":\"/usr/share/\"\n" +
			"}",
		conf: &Configuration{
			Hostname:    "http://example3.com",
			OAuthID:     "123454685139",
			OAuthSecret: "abcdef123454accd58be5f5ee6789",
			BasePath:    "/usr/share/",
		},
	},
}

func TestLoadStandardConfigFile(t *testing.T) {
	for _, in := range testLoadStandardConfigFileInput {
		err := ioutil.WriteFile(filepath.Join(StandardBasePath, ConfigFileName), []byte(in.filedata), 0666)
		if err != nil {
			t.Error(err)
			continue
		}

		conf, err := Load(StandardBasePath)
		if err != nil {
			t.Error(err)
			continue
		}

		compareConfigObjects(conf, in.conf, t)
	}
}

var testExportToGlobalVarsInput = []*Configuration{
	&Configuration{
		Hostname:    "http://example.com",
		OAuthID:     "1234",
		OAuthSecret: "abcd",
		BasePath:    "/example/",
	},
	&Configuration{
		Hostname:    "http://example2.com",
		OAuthID:     "123456789",
		OAuthSecret: "abcdef123456789",
		BasePath:    "/usr/",
	},
	&Configuration{
		Hostname:    "http://example3.com",
		OAuthID:     "123454685139",
		OAuthSecret: "abcdef123454accd58be5f5ee6789",
		BasePath:    "/usr/share/",
	},
}

func TestExportToGlobalVars(t *testing.T) {
	for _, conf := range testExportToGlobalVarsInput {
		conf.ExportToGlobalVars()
		compareConfigObjectsToGlobal(conf, t)
	}
}

var testValidateInput = []struct {
	valid bool
	conf  *Configuration
}{
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example.com",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "http://example2.com/",
			OAuthID:     "123456789",
			OAuthSecret: "abcdef123456789",
			BasePath:    "/usr/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "http://example3.com",
			OAuthID:     "123454685139",
			OAuthSecret: "abcdef123454accd58be5f5ee6789",
			BasePath:    "",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "example.com",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "http://example.com",
			OAuthID:     "1234",
			OAuthSecret: "",
			BasePath:    "/example/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "http://example.com",
			OAuthID:     "",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
}

func TestValidate(t *testing.T) {
	for _, in := range testValidateInput {
		err := in.conf.Validate()
		if err != nil && in.valid {
			t.Error("A valid configuration object was not validated:", err)
		} else if err == nil && !in.valid {
			t.Error("A not valid configuration object was validated. Object:", in.conf)
		}
	}
}

var testQuickFixInput = []struct {
	valid bool
	conf  *Configuration
}{
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example.com/",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example2.com",
			OAuthID:     "123456789",
			OAuthSecret: "abcdef123456789",
			BasePath:    "",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example3.com",
			OAuthID:     "123454685139",
			OAuthSecret: "abcdef123454accd58be5f5ee6789",
			BasePath:    "/usr",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example.com/",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example.com",
			OAuthID:     "1234",
			OAuthSecret: "abcd",
			BasePath:    "/example/",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			Hostname:    "http://example.com/",
			OAuthID:     "abcd",
			OAuthSecret: "abcd",
			BasePath:    "",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			Hostname:    "http://example.com/",
			OAuthID:     "abcd",
			OAuthSecret: "",
			BasePath:    "",
		},
	},
}

func TestQuickFix(t *testing.T) {
	for _, in := range testQuickFixInput {
		err := in.conf.QuickFix()
		if err != nil && in.valid {
			t.Error("A valid configuration object was not validated, after quick fix:", err)
		} else if err == nil && !in.valid {
			t.Error("A not valid configuration object was validated, after quick fix. Object:", in.conf)
		}
	}
}

var testSaveInput = []*Configuration{
	&Configuration{
		Hostname:    "http://example.com",
		OAuthID:     "1234",
		OAuthSecret: "abcd",
	},
	&Configuration{
		Hostname:    "http://example2.com",
		OAuthID:     "123456789",
		OAuthSecret: "abcdef123456789",
	},
	&Configuration{
		Hostname:    "http://example3.com",
		OAuthID:     "123454685139",
		OAuthSecret: "abcdef123454accd58be5f5ee6789",
	},
}

func TestSave(t *testing.T) {
	// TODO: also test for non existing dir path and wrong dir type
	for _, conf := range testSaveInput {
		err := conf.QuickFix()
		if err != nil {
			t.Error(err)
			continue
		}

		if err = conf.Save(); err != nil {
			t.Error(err)
			continue
		}

		conf2, err := Load(StandardBasePath)
		if err != nil {
			t.Error(err)
			continue
		}

		compareConfigObjects(conf, conf2, t)
	}
}

func compareConfigObjects(c1, c2 *Configuration, t *testing.T) {
	if c1.Hostname != c2.Hostname {
		t.Errorf("Field value Hostname does not match. %v != %v", c1.Hostname, c2.Hostname)
	}
	if c1.OAuthID != c2.OAuthID {
		t.Errorf("Field value OAuthID does not match. %v != %v", c1.OAuthID, c2.OAuthID)
	}
	if c1.OAuthSecret != c2.OAuthSecret {
		t.Errorf("Field value OAuthSecret does not match. %v != %v", c1.OAuthSecret, c2.OAuthSecret)
	}
	if c1.BasePath != c2.BasePath {
		t.Errorf("Field value BasePath does not match. %v != %v", c1.BasePath, c2.BasePath)
	}
}

func compareConfigObjectsToGlobal(c1 *Configuration, t *testing.T) {
	if c1.Hostname != global.Hostname {
		t.Errorf("Field value Hostname does not match. %v != %v", c1.Hostname, global.Hostname)
	}
	if c1.OAuthID != global.OAuthClientID {
		t.Errorf("Field value OAuthID does not match. %v != %v", c1.OAuthID, global.OAuthClientID)
	}
	if c1.OAuthSecret != global.OAuthClientSecret {
		t.Errorf("Field value OAuthSecret does not match. %v != %v", c1.OAuthSecret, global.OAuthClientSecret)
	}
}
