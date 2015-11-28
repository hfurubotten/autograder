package entities

import (
	"encoding/gob"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hfurubotten/autograder/database"
)

// OrganizationBucketName is the bucket/table name for organizations in the DB.
var GroupsBucketName = "groups"

func init() {
	gob.Register(Group{})
	database.RegisterBucket(GroupsBucketName)
}

// Group represents a group of students in a course.
type Group struct {
	// synchronization variables (must be package private to avoid storing to DB)
	mu *sync.RWMutex

	ID      int //TODO to be removed later??
	TeamID  int
	Active  bool
	Name    string
	Course  string
	Members map[string]interface{}

	CurrentLabNum int
	Assignments   map[int]*Assignment

	lock sync.Mutex //TODO remove me later
}

// NewGroup creates a new group with the provided name for the given course.
func NewGroup(course, name string) (g *Group) {
	return &Group{
		Course:        course,
		Name:          name,
		Members:       make(map[string]interface{}),
		CurrentLabNum: 1,
		Assignments:   make(map[int]*Assignment),
	}
}

// NewGroupWithID creates a new group for the given course
// with a unique group ID.
func NewGroupWithID(course string) (*Group, error) {
	gid, err := nextGroupID()
	if err != nil {
		return nil, err
	}
	groupName := GroupRepoPrefix + strconv.Itoa(gid)
	return NewGroup(course, groupName), nil
}

// GetGroup returns the group associated with the given groupName.
func GetGroup(groupName string) (g *Group, err error) {
	err = database.Get(GroupsBucketName, groupName, &g)
	if err != nil {
		return nil, err
	}
	g.mu = &sync.RWMutex{}
	// groupName found in database
	return g, nil
}

// Save will store the group information in the database.
func (g *Group) Save() error {
	return database.Put(GroupsBucketName, g.Name, g)
}

// Activate will activate/approve a group.
func (g *Group) Activate() {
	g.Active = true

	for username := range g.Members {
		user, err := GetMember(username)
		if err != nil {
			log.Println(err)
			continue
		}

		opt := user.Courses[g.Course]
		if !opt.IsGroupMember {
			opt.IsGroupMember = true
			opt.GroupNum = g.ID
			opt.GroupName = g.Name
			user.Courses[g.Course] = opt
			err := user.Save()
			if err != nil {
				//return error
			}
		}
	}
}

// AddMember will add a new member to the group.
func (g *Group) AddMember(user string) {
	g.Members[user] = nil
}

// RemoveMember will remove a member from the group.
func (g *Group) RemoveMember(user string) {
	if len(g.Members) <= 1 {
		g.Delete()
	}
	delete(g.Members, user)
}

// AddBuildResult will add a build result to the group.
func (g *Group) AddBuildResult(lab, buildid int) {
	if g.Assignments == nil {
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
	}

	g.Assignments[lab].AddBuildResult(buildid)
}

// GetLastBuildID will get the last build ID added to a lab assignment.
func (g *Group) GetLastBuildID(lab int) int {
	if assignment, ok := g.Assignments[lab]; ok {
		if assignment.Builds == nil {
			return -1
		}
		if len(assignment.Builds) == 0 {
			return -1
		}

		return assignment.Builds[len(assignment.Builds)-1]
	}

	return -1
}

// SetApprovedBuild will put the approved build results in
func (g *Group) SetApprovedBuild(labnum, buildid int, date time.Time) {
	if _, ok := g.Assignments[labnum]; !ok {
		g.Assignments[labnum] = NewAssignment()
	}

	g.Assignments[labnum].ApproveDate = date
	g.Assignments[labnum].ApprovedBuild = buildid

	if g.CurrentLabNum <= labnum {
		g.CurrentLabNum = labnum + 1
	}
}

// AddNotes will add notes to a lab assignment.
func (g *Group) AddNotes(lab int, notes string) {
	if g.Assignments == nil {
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
	}

	g.Assignments[lab].Notes = notes
}

// GetNotes will get notes from a lab assignment.
func (g *Group) GetNotes(lab int) string {
	if g.Assignments == nil {
		g.Assignments = make(map[int]*Assignment)
	}

	if _, ok := g.Assignments[lab]; !ok {
		g.Assignments[lab] = NewAssignment()
	}

	return g.Assignments[lab].Notes
}

//TODO: We should never export lock functions. That's asking for trouble!!

// Lock will put a writers lock on the group.
func (g *Group) Lock() {
	g.lock.Lock()
}

// Unlock will remove a writers lock on the group. If there is no lock this
// method will panic.
func (g *Group) Unlock() {
	g.lock.Unlock()
}

// Delete will remove the group object.
func (g *Group) Delete() error {
	for username := range g.Members {
		user, err := GetMember(username)
		if err != nil {
			continue
		}

		courseopt := user.Courses[g.Course]
		if courseopt.GroupNum == g.ID {
			user.Lock()
			courseopt.IsGroupMember = false
			courseopt.GroupNum = 0
			user.Courses[g.Course] = courseopt
			if err = user.Save(); err != nil {
				user.Unlock()
				log.Println(err)
			}
		}
	}

	return database.GetPureDB().Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(GroupsBucketName)).Delete([]byte(strconv.Itoa(g.ID)))
	})
}

// HasGroup will check if the group is in storage.
func HasGroup(groupid int) bool {
	return database.Has(GroupsBucketName, strconv.Itoa(groupid))
}

// nextGroupID will get the next group id available.
func nextGroupID() (int, error) {
	id, err := database.NextID(GroupsBucketName)
	return int(id), err
}
