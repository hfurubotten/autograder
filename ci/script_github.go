package ci

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"text/template"
)

// Script needs to be documented later
func Script(opt DaemonOptions) []string {

	// TODO why does BaseFolder have trailing slash (/)?
	ciScript := `mkdir -p {{.BaseFolder}}
git clone https://{{.AdminToken}}:x-oauth-basic@github.com/{{.Org}}/{{.UserRepo}}.git {{.BaseFolder}}{{.DestFolder}}
git clone https://{{.AdminToken}}:x-oauth-basic@github.com/{{.Org}}/{{.TestRepo}}.git {{.BaseFolder}}{{.TestRepo}}
{{/* Copy AG tests from test repo to dest folder in docker VM instance */}}
cp -rf {{.BaseFolder}}{{.TestRepo}}/* {{.BaseFolder}}{{.DestFolder}}
cd {{.BaseFolder}}{{.DestFolder}}
{{/* We may want to check for file existence before */}}
chmod 700 dependencies.sh {{.LabFolder}}/test.sh
./dependencies.sh
{{.LabFolder}}/test.sh

{{/* This should be identical to the original version, if the version above doesn't work! */}}
/bin/bash -c "cp -rf "{{.BaseFolder}}{{.TestRepo}}/*" "{{.BaseFolder}}{{.DestFolder}}" "
chmod 700 {{.BaseFolder}}{{.DestFolder}}/dependencies.sh
chmod 700 {{.BaseFolder}}{{.DestFolder}}/{{.LabFolder}}/test.sh
/bin/sh -c "(cd "{{.BaseFolder}}{{.DestFolder}}" && ./dependencies.sh)"
/bin/sh -c "(cd "{{.BaseFolder}}{{.DestFolder}}/{{.LabFolder}}" && ./test.sh)"
  `
	tmpl, err := template.New("script").Parse(ciScript)
	if err != nil {
		// We just log it; should never panic
		log.Println(err)
		return []string{}
	}
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)
	err = tmpl.Execute(w, opt)
	if err != nil {
		// We just log it; should never panic
		log.Println(err)
		return []string{}
	}

	// fmt.Println()
	// cmds := []struct {
	// 	Cmd       string
	// 	Breakable bool
	// }{
	// 	{"mkdir -p " + opt.BaseFolder, true},
	// 	{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + opt.UserRepo + ".git" + " " + opt.BaseFolder + opt.DestFolder + "/", true},
	// 	{"git clone https://" + opt.AdminToken + ":x-oauth-basic@github.com/" + opt.Org + "/" + git.TestRepoName + ".git" + " " + opt.BaseFolder + git.TestRepoName + "/", true},
	// 	{"/bin/bash -c \"cp -rf \"" + opt.BaseFolder + git.TestRepoName + "/*\" \"" + opt.BaseFolder + opt.DestFolder + "/\" \"", true},
	//
	// 	{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/dependencies.sh", true},
	// 	{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/\" && ./dependencies.sh)\"", true},
	// 	{"chmod 777 " + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/test.sh", true},
	// 	{"/bin/sh -c \"(cd \"" + opt.BaseFolder + opt.DestFolder + "/" + opt.LabFolder + "/\" && ./test.sh)\"", false},
	// }
	// for _, k := range cmds {
	// 	fmt.Println(k.Cmd)
	// }

	return strings.Split(string(buf.Bytes()), "\n")
}
