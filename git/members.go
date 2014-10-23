package git

import (
	"encoding/gob"
	"log"
	"fmt"
	"crypto/sha256"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(member{})
}

var store = diskv.New(diskv.Options{
	CacheSizeMax: 1024 * 1024 * 256,
})

type member struct {
	githubclient *github.Client
	Username     string
	Name         string
	StudentID    int

	AccessToken string
}

func NewMember(oauthtoken string) member {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: oauthtoken},
	}
	g := github.NewClient(t.Client())
	return member{githubclient: g, AccessToken: oauthtoken}
}

func (m *member) LoadData() (err error) {
	if hash := sha256.Sum256([]byte(m.AccessToken)); store.Has(fmt.Sprintf("%x", hash)) {
		var tmp member

    	err = store.ReadGob(m.Username, &tmp, false)
    	m.Copy(tmp)

    	return
	}

	user, _, err := m.githubclient.Users.Get("")
	if err != nil {
		log.Println(err)
		return
	}

	if user.Login != nil {
		m.Username = *user.Login
	}

	if user.Name != nil {
		m.Name = *user.Name
	}

	return
}

func (m member) StickToSystem() (err error) {
	hash := sha256.Sum256([]byte(m.AccessToken))
    err = store.WriteGob(fmt.Sprintf("%x", hash), m)

    return 
}

func (m *member) FetchFromSystem(accesstoken string) (err error) {

	var tmp member

	hash := sha256.Sum256([]byte(accesstoken))
    err = store.ReadGob(fmt.Sprintf("%x", hash), &tmp, false)
    m.Copy(tmp)

    return
}

func (m *member) Copy(tmp member){
	m.Username     = tmp.Username
	m.Name         = tmp.Name
	m.StudentID    = tmp.StudentID

	m.AccessToken  = tmp.AccessToken
}