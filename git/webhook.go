package git

import (
	"encoding/json"
	"io"

	//"github.com/google/go-github/github"
)

type HookPayload struct {
	User            string
	Repo            string
	Fullname        string
	Organization    string
	LatestCommitTag string
}

func DecodeHookPayload(r io.Reader) (load HookPayload, err error) {
	// work around go-github bug decoding the payload
	var v map[string]interface{}

	load = HookPayload{}

	dec := json.NewDecoder(r)
	err = dec.Decode(&v)
	if err != nil {
		return
	}

	if userdata, ok := v["pusher"]; ok {
		if userdata, ok := userdata.(map[string]interface{}); ok {
			load.User = userdata["name"].(string)
		}
	}

	if repodata, ok := v["repository"]; ok {
		if repodata, ok := repodata.(map[string]interface{}); ok {
			load.Repo = repodata["name"].(string)
			load.Fullname = repodata["full_name"].(string)
		}
	}

	if orgdata, ok := v["organization"]; ok {
		if orgdata, ok := orgdata.(map[string]interface{}); ok {
			load.Organization = orgdata["login"].(string)
		}
	}

	if commitdata, ok := v["after"]; ok {
		load.LatestCommitTag = commitdata.(string)
	}

	return
}
