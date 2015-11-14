package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var basePath = "testfiles"

func TestMain(m *testing.M) {
	err := os.Mkdir(basePath, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	m.Run()
	err = os.RemoveAll(basePath)
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

		if conf.URL != in.url {
			t.Errorf("URL field mismatch, got: %v, expected: %v", conf.URL, in.url)
		}
		if conf.OAuthClientID != in.id {
			t.Errorf("OAuthClientID field mismatch, got: %v, expected: %v", conf.OAuthClientID, in.id)
		}
		if conf.OAuthClientSecret != in.secret {
			t.Errorf("OAuthClientSecret field mismatch, got: %v, expected: %v", conf.OAuthClientSecret, in.secret)
		}
		if conf.BasePath != in.path {
			t.Errorf("BasePath field mismatch, got: %v, expected: %v", conf.BasePath, in.path)
		}
	}
}

func TestGetPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Failed to recover from panic")
		}
	}()
	Get()
	t.Fatal("Expected panic on call to Get() before initialization")
}

func TestSetCurrent(t *testing.T) {
	// this will run setCurrent only once
	for _, in := range testNewConfigInput[0:1] {
		conf, _ := NewConfig(in.url, in.id, in.secret, in.path)
		conf.SetCurrent()
	}
}

func TestSetCurrentPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Failed to recover from panic")
		}
	}()
	for _, in := range testNewConfigInput {
		conf, _ := NewConfig(in.url, in.id, in.secret, in.path)
		// this will panic since it was already initialized in TestSetCurrent()
		conf.SetCurrent()
	}
	t.Fatal("Expected panic on repeated SetCurrent() invocations")
}

func TestGet(t *testing.T) {
	// we can call Get() here because it was initialized above in TestSetCurrent()
	want := Get()
	for i := 0; i < 5; i++ {
		got := Get()
		if want != got {
			t.Errorf("want %v, got %v", want, got)
		}
	}
	c := Get()
	_ = c
}

var testNewConfigRemoveSuffixInput = []struct {
	url, id, secret, path, out string
}{
	{"http://example1.com", "1234acd544", "abcd455813aa", "/tmp", "http://example1.com"},
	{"http://example2.com/", "1234acd544", "abcd455813aa", "/tmp", "http://example2.com"},
}

func TestNewConfigRemoveSuffix(t *testing.T) {
	for _, in := range testNewConfigRemoveSuffixInput {
		conf, err := NewConfig(in.url, in.id, in.secret, in.path)
		if err != nil {
			t.Error(err)
			continue
		}
		if conf.URL != in.out {
			t.Errorf("Expected url: '%v', got: '%v'", in.out, conf.URL)
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
			t.Errorf("Expected <nil>, got: %v", conf)
			continue
		}
		if err == nil {
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
		filedata: `{
  "URL":"http://example.com",
  "OAuthClientID":"1234",
  "OAuthClientSecret":"abcd",
  "BasePath":"/example/"
}`,
		conf: &Configuration{
			URL:               "http://example.com",
			OAuthClientID:     "1234",
			OAuthClientSecret: "abcd",
			BasePath:          "/example",
		},
	},
	{
		filedata: `{
  "URL":"http://example2.com",
  "OAuthClientID":"123456789",
  "OAuthClientSecret":"abcdef123456789",
  "BasePath":"/usr/"
}`,
		conf: &Configuration{
			URL:               "http://example2.com",
			OAuthClientID:     "123456789",
			OAuthClientSecret: "abcdef123456789",
			BasePath:          "/usr",
		},
	},
	{
		filedata: `{
  "URL":"http://example3.com",
  "OAuthClientID":"123454685139",
  "OAuthClientSecret":"abcdef123454accd58be5f5ee6789",
  "BasePath":"/usr/share/"
}`,
		conf: &Configuration{
			URL:               "http://example3.com",
			OAuthClientID:     "123454685139",
			OAuthClientSecret: "abcdef123454accd58be5f5ee6789",
			BasePath:          "/usr/share",
		},
	},
	{
		filedata: `{
		"URL": "http://example3.com/",
		"OAuthClientID": "1234",
		"OAuthClientSecret": "abcd",
		"BasePath": "/example/"
		}`,
		conf: &Configuration{
			URL:               "http://example3.com",
			OAuthClientID:     "1234",
			OAuthClientSecret: "abcd",
			BasePath:          "/example",
		},
	},
	{
		filedata: `{
		"URL": "http://example7.com",
		"OAuthClientID": "12039857",
	  "OAuthClientSecret": "abcdef123456789",
	  "BasePath": "/usr/share/"
		}`,
		conf: &Configuration{
			URL:               "http://example7.com",
			OAuthClientID:     "12039857",
			OAuthClientSecret: "abcdef123456789",
			BasePath:          "/usr/share",
		},
	},
}

func TestLoadStandardConfigFile(t *testing.T) {
	for _, in := range testLoadStandardConfigFileInput {
		err := ioutil.WriteFile(filepath.Join(basePath, FileName), []byte(in.filedata), 0666)
		if err != nil {
			t.Error(err)
			continue
		}

		conf, err := Load(basePath)
		if err != nil {
			t.Error(err)
			continue
		}

		compareConfigObjects(conf, in.conf, t)
	}
}

var testLoadNonValidInput = []struct {
	filedata string
}{
	{
		filedata: `{
		"URL": "",
	  "OAuthClientID": "12039857",
	  "OAuthClientSecret": "abcdef123456789",
	  "BasePath": "/usr/"
		}`,
	},
	{
		filedata: `{
		"URL": "http://example2.com",
	  "OAuthClientID": "",
	  "OAuthClientSecret": "abcdef123456789",
	  "BasePath": "/usr/"
		}`,
	},
	{
		filedata: `{
		"URL": "http://example3.com",
		"OAuthClientID": "12039857",
	  "OAuthClientSecret": "",
	  "BasePath": "/usr/"
		}`,
	},
	{
		filedata: `{
		"URL": "http://example4.com",
		"OAuthClientID": "12039857",
	  "OAuthClientSecret": "abcdef123456789",
	  "BasePath": ""
		}`,
	},
	{
		filedata: `{
		"URL": "smb://example5.com",
	  "OAuthClientID": "23123",
	  "OAuthClientSecret": "abcdef123456789",
	  "BasePath": "/usr/"
		}`,
	},
	{
		filedata: `{
		"URL": "http://example6.com/",
		"OAuthClientID": "123454685139",
		"OAuthClientSecret": "abcdef123454accd58be5f5ee6789",
		"BasePath": ""
		}`,
	},
}

func TestLoadNonValidInput(t *testing.T) {
	for _, in := range testLoadNonValidInput {
		err := ioutil.WriteFile(filepath.Join(basePath, FileName), []byte(in.filedata), 0666)
		if err != nil {
			t.Error(err)
			continue
		}

		conf, err := Load(basePath)
		if conf != nil {
			t.Errorf("Expected <nil> conf, got %v", conf)
			continue
		}
		if err == nil {
			t.Error("Expected non-nil error, didn't get any errors")
			continue
		}
	}
}

var testValidateInput = []struct {
	valid bool
	conf  *Configuration
}{
	{
		valid: true,
		conf: &Configuration{
			URL:               "http://example.com",
			OAuthClientID:     "1234",
			OAuthClientSecret: "abcd",
			BasePath:          "/example/",
		},
	},
	{
		valid: true,
		conf: &Configuration{
			URL:               "http://example2.com/",
			OAuthClientID:     "123456789",
			OAuthClientSecret: "abcdef123456789",
			BasePath:          "/usr/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			URL:               "http://example3.com",
			OAuthClientID:     "123454685139",
			OAuthClientSecret: "abcdef123454accd58be5f5ee6789",
			BasePath:          "",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			URL:               "example.com",
			OAuthClientID:     "1234",
			OAuthClientSecret: "abcd",
			BasePath:          "/example/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			URL:               "http://example.com",
			OAuthClientID:     "1234",
			OAuthClientSecret: "",
			BasePath:          "/example/",
		},
	},
	{
		valid: false,
		conf: &Configuration{
			URL:               "http://example.com",
			OAuthClientID:     "",
			OAuthClientSecret: "abcd",
			BasePath:          "/example/",
		},
	},
}

func TestValidate(t *testing.T) {
	for _, in := range testValidateInput {
		err := in.conf.validate()
		if err != nil && in.valid {
			t.Error("A valid configuration object was not validated:", err)
		} else if err == nil && !in.valid {
			t.Error("A not valid configuration object was validated. Object:", in.conf)
		}
	}
}

var testSaveInput = []*Configuration{
	&Configuration{
		URL:               "http://example.com",
		OAuthClientID:     "1234",
		OAuthClientSecret: "abcd",
		BasePath:          "example/",
	},
	&Configuration{
		URL:               "http://example2.com",
		OAuthClientID:     "123456789",
		OAuthClientSecret: "abcdef123456789",
		BasePath:          "share/autograder/",
	},
	&Configuration{
		URL:               "http://example3.com",
		OAuthClientID:     "123454685139",
		OAuthClientSecret: "abcdef123454accd58be5f5ee6789",
		BasePath:          "share",
	},
}

func TestSave(t *testing.T) {
	// TODO: also test for non existing dir path and wrong dir type
	for _, conf := range testSaveInput {
		err := conf.validate()
		if err != nil {
			t.Error(err)
			continue
		}
		if err = conf.Save(); err != nil {
			t.Error(err)
			continue
		}
		conf2, err := Load(conf.BasePath)
		if err != nil {
			t.Error(err)
			continue
		}
		compareConfigObjects(conf, conf2, t)
		err = os.RemoveAll(conf.BasePath)
		if err != nil {
			t.Error(err)
			continue
		}
	}
}

func compareConfigObjects(c1, c2 *Configuration, t *testing.T) {
	if c1.URL != c2.URL {
		t.Errorf("Field value URL does not match. %v != %v", c1.URL, c2.URL)
	}
	if c1.OAuthClientID != c2.OAuthClientID {
		t.Errorf("Field value OAuthClientID does not match. %v != %v", c1.OAuthClientID, c2.OAuthClientID)
	}
	if c1.OAuthClientSecret != c2.OAuthClientSecret {
		t.Errorf("Field value OAuthClientSecret does not match. %v != %v", c1.OAuthClientSecret, c2.OAuthClientSecret)
	}
	if c1.BasePath != c2.BasePath {
		t.Errorf("Field value BasePath does not match. %v != %v", c1.BasePath, c2.BasePath)
	}
}
