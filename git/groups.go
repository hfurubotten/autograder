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

type Group struct {
	ID            int
	Active        bool
	Course        string
	Members       map[string]interface{}
	CurrentLabNum int

	store *diskv.Diskv
}

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
		CurrentLabNum: 1,
		store:         store,
	}

	return
}

func (g *Group) Activate() {
	g.Active = true

	for username, _ := range g.Members {
		user, err := NewMemberFromUsername(username)
		if err != nil {
			continue
		}

		opt := user.Courses[g.Course]
		if !opt.IsGroupMember {
			opt.IsGroupMember = true
			opt.GroupNum = g.ID
			user.Courses[g.Course] = opt
			user.Save()
		}
	}
}

func (g *Group) AddMember(user string) {
	g.Members[user] = nil
}

func (g *Group) Lock() {
	// TODO: implement locking
}

func (g *Group) Unlock() {
	// TODO: implement locking
}

func (g Group) Save() error {
	return g.store.WriteGob(strconv.Itoa(g.ID), g)
}

func (g *Group) Delete() error {
	for username, _ := range g.Members {
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

func GetGroupStore(org string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:     global.Basepath + "diskv/groups/" + org + "/",
		CacheSizeMax: 1024 * 1024 * 256,
	})
}

func HasGroup(org string, groupid int) bool {
	storage := GetGroupStore(org)
	return storage.Has(strconv.Itoa(groupid))
}
