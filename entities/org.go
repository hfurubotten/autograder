package entities

import (
	"encoding/gob"
	//	"errors"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/hfurubotten/autograder/game/points"
)

func init() {
	gob.Register(OrganizationX{})
}

//
type OrganizationX struct {
	points.Leaderboard
	lock sync.Mutex

	Name        string
	ScreenName  string
	Description string
	Location    string
	Company     string

	// URLs
	HTMLURL   string
	AvatarURL string
}

func NewOrganizationX(name string) (org *OrganizationX, err error) {
	o := new(OrganizationX)
	o.Name = name
	o.TotalScore = make(map[string]int64)
	o.WeeklyScore = make(map[int]map[string]int64)
	o.MonthlyScore = make(map[time.Month]map[string]int64)

	return o, nil
}

func NewOrganizationWithGithubDataX(gorg *github.Organization) (org *OrganizationX, err error) {
	org, err = NewOrganizationX(*gorg.Login)
	if err != nil {
		return nil, err
	}

	org.ImportGithubDataX(gorg)
	return
}

// ImportGithubData imports data from the given github
// data object and stores it in the given Organization object.
func (o *OrganizationX) ImportGithubDataX(gorg *github.Organization) {
	if gorg == nil {
		return
	}

	if gorg.Name != nil {
		o.ScreenName = *gorg.Name
	}

	// Missing from go-github
	//if gorg.Description != nil {
	//	o.Description = gorg.Description
	//}

	if gorg.Location != nil {
		o.Location = *gorg.Location
	}

	if gorg.Company != nil {
		o.Company = *gorg.Company
	}
}

// LoadStoredData fetches the organization data stored on disk or in cached memory.
// ATM a NO-OP
func (o *OrganizationX) LoadStoredDataX() (err error) {
	return nil
}

// Lock will lock the organization name from being written to by
// other instances of the same organization. This has to be used
// when new info is written, to prevent race conditions. Unlock
// occures when data is finished written to storage.
func (o *OrganizationX) Lock() {
	o.lock.Lock()
}

// Unlock will unlock the writers block on the orgnization.
func (o *OrganizationX) Unlock() {
	o.lock.Unlock()
}

// Save will store the Organization object to disk and be cached in
// memory. The save function will also unlock the organization for
// writing. If the org is not locked before saving, a runtime error
// will be called.
// ATM a NO-OP
func (o *OrganizationX) SaveX() error {
	o.Unlock()
	return nil
}
