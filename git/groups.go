package git

import (
	"encoding/gob"
	"strconv"

	"github.com/hfurubotten/autograder/global"
	"github.com/hfurubotten/diskv"
)

func init() {
	gob.Register(Group{})
}

// Group represents a group of students in a course.
type Group struct {
	ID      int
	TeamID  int
	Active  bool
	Course  string
	Members map[string]interface{}

	CurrentLabNum int
	Assignments   map[int]LabAssignmentOptions

	store *diskv.Diskv
}

// NewGroup will try to fetch a group for storage, if non is found it creates a new one.
func NewGroup(org string, groupnum int) (g *Group, err error) {
	store := GetGroupStore(org)
	num := strconv.Itoa(groupnum)
	g = new(Group)

	if store.Has(num) {
		err = store.ReadGob(num, g, false)
		if err != nil {
			return
		}
		g.store = store
		return
	}

	g = &Group{
		ID:            groupnum,
		Active:        false,
		Course:        org,
		Members:       make(map[string]interface{}),
		Assignments:   make(map[int]LabAssignmentOptions),
		CurrentLabNum: 1,
		store:         store,
	}

	return
}

// Activate will activate/approve a group.
func (g *Group) Activate() {
	g.Active = true

	for username := range g.Members {
		user, err := NewMemberFromUsername(username)
		if err != nil {
			continue
		}

		user.Lock()
		defer user.Unlock()

		opt := user.Courses[g.Course]
		if !opt.IsGroupMember {
			opt.IsGroupMember = true
			opt.GroupNum = g.ID
			user.Courses[g.Course] = opt
			user.Save()
		}
	}
}

// AddMember will add a new member to the group.
func (g *Group) AddMember(user string) {
	g.Members[user] = nil
}

// Lock will put a writers lock on the group.
//
// Not yet implemented
func (g *Group) Lock() {
	// TODO: implement locking
}

// Unlock will remove a writers lock on the group. If there is no lock this method will panic.
//
// Not yet implemented
func (g *Group) Unlock() {
	// TODO: implement locking
}

// Save will store the group to memory and disk.
func (g *Group) Save() error {
	return g.store.WriteGob(strconv.Itoa(g.ID), g)
}

// Delete will remove the group object.
func (g *Group) Delete() error {
	for username := range g.Members {
		user, err := NewMemberFromUsername(username)
		if err != nil {
			continue
		}

		courseopt := user.Courses[g.Course]
		if courseopt.GroupNum == g.ID {
			courseopt.IsGroupMember = false
			courseopt.GroupNum = 0
			user.Courses[g.Course] = courseopt
			user.Save()
		}
	}

	return g.store.Erase(strconv.Itoa(g.ID))
}

var groupstore map[string]*diskv.Diskv

// GetGroupStore will get the Diskv object used to fetch the group object from storage.
func GetGroupStore(org string) *diskv.Diskv {
	if groupstore == nil {
		groupstore = make(map[string]*diskv.Diskv)
	}

	if _, ok := groupstore[org]; !ok {
		groupstore[org] = diskv.New(diskv.Options{
			BasePath:     global.Basepath + "diskv/groups/" + org + "/",
			CacheSizeMax: 1024 * 1024 * 256,
		})
	}

	return groupstore[org]
}

// HasGroup will check if the group is in storage.
func HasGroup(org string, groupid int) bool {
	storage := GetGroupStore(org)
	return storage.Has(strconv.Itoa(groupid))
}
