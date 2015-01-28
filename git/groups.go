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

func NewGroup(org string, groupnum int) (g Group, err error) {
	store := GetGroupStore(org)
	num := strconv.Itoa(groupnum)

	if store.Has(num) {
		err = store.ReadGob(num, &g, false)
		if err != nil {
			return
		}
		g.store = store
		return
	}

	g = Group{
		ID:            groupnum,
		Active:        false,
		Course:        org,
		Members:       make(map[string]interface{}),
		CurrentLabNum: 1,
		store:         store,
	}

	return
}

func (g *Group) AddMember(user string) {
	g.Members[user] = nil
}

func (g Group) StickToSystem() error {
	return g.store.WriteGob(strconv.Itoa(g.ID), g)
}

func (g *Group) Delete() error {
	for username, _ := range g.Members {
		user := NewMemberFromUsername(username)
		courseopt := user.Courses[g.Course]
		courseopt.IsGroupMember = false
		courseopt.GroupNum = 0
		user.Courses[g.Course] = courseopt
		user.StickToSystem()
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
