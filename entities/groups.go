package entities

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hfurubotten/autograder/database"
)

// OrganizationBucketName is the bucket/table name for organizations in the DB.
var GroupsBucketName = "groups"

// InMemoryOrgs is a mapper where pointers to all the Organization are kept in memory.
var InMemoryGroups = make(map[int]*Group)

// InMemoryOrgsLock is the locking for the org mapper.
var InMemoryGroupsLock sync.Mutex

// GroupLengthKey is the key for finding current count of new group IDs.
var GroupLengthKey = "length"

func init() {
	gob.Register(Group{})

	database.RegisterBucket(GroupsBucketName)
}

// Group represents a group of students in a course.
type Group struct {
	ID      int
	TeamID  int
	Active  bool
	Course  string
	Members map[string]interface{}

	CurrentLabNum int
	Assignments   map[int]*Assignment

	lock sync.Mutex
}

// NewGroup will try to fetch a group for storage, if non is found it creates a new one.
func NewGroup(org string, groupid int, readonly bool) (g *Group, err error) {
	InMemoryGroupsLock.Lock()
	defer InMemoryGroupsLock.Unlock()

	g = &Group{
		ID:            groupid,
		Active:        false,
		Course:        org,
		Members:       make(map[string]interface{}),
		Assignments:   make(map[int]*Assignment),
		CurrentLabNum: 1,
	}

	if _, ok := InMemoryGroups[groupid]; ok {
		g = InMemoryGroups[groupid]
		if !readonly {
			g.Lock()
		}

		return g, nil
	}

	err = g.loadStoredData(!readonly)
	if err != nil {
		if err.Error() == "No data in database" {
			return nil, err
		}
	}
	// Add the org to in memory mapper.
	InMemoryGroups[g.ID] = g

	return g, nil
}

func (g *Group) loadStoredData(lock bool) error {
	err := database.GetPureDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GroupsBucketName))
		if b == nil {
			return errors.New("Bucket not found. Are you sure the bucket was registered correctly?")
		}

		data := b.Get([]byte(strconv.Itoa(g.ID)))
		if data == nil {
			return errors.New("No data in database.")
		}

		buf := &bytes.Buffer{}
		decoder := gob.NewDecoder(buf)

		n, _ := buf.Write(data)

		if n != len(data) {
			return errors.New("Couldn't write all data to buffer while getting data from database. " + strconv.Itoa(n) + " != " + strconv.Itoa(len(data)))
		}

		err := decoder.Decode(g)
		if err != nil {
			return err
		}

		return nil
	})

	//TODO: What is this?? Why have an option to lock or not?? Bad practice.

	// locks the object directly in order to ensure consistent info from DB.
	if lock {
		g.Lock()
	}

	return err
}

// Activate will activate/approve a group.
func (g *Group) Activate() {
	g.Active = true

	for username := range g.Members {
		user, err := GetMember(username)
		if err != nil {
			continue
		}

		opt := user.Courses[g.Course]
		if !opt.IsGroupMember {
			user.Lock()
			opt.IsGroupMember = true
			opt.GroupNum = g.ID
			user.Courses[g.Course] = opt
			err := user.Save()
			if err != nil {
				user.Unlock()
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

// Save will store the group to memory and disk.
func (g *Group) Save() error {
	return database.GetPureDB().Update(func(tx *bolt.Tx) (err error) {
		// open the bucket
		b := tx.Bucket([]byte(GroupsBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		buf := &bytes.Buffer{}
		encoder := gob.NewEncoder(buf)

		if err = encoder.Encode(g); err != nil {
			return
		}

		data, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		err = b.Put([]byte(strconv.Itoa(g.ID)), data)
		if err != nil {
			return err
		}

		g.Unlock()
		return nil
	})
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

	delete(InMemoryGroups, g.ID)

	return database.GetPureDB().Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(GroupsBucketName)).Delete([]byte(strconv.Itoa(g.ID)))
	})
}

// HasGroup will check if the group is in storage.
func HasGroup(groupid int) bool {
	return database.Has(GroupsBucketName, strconv.Itoa(groupid))
}

// GetNextGroupID will get the next group id available.
// Returns -1 on error.
func GetNextGroupID() int {
	nextid := -1
	if err := database.GetPureDB().Update(func(tx *bolt.Tx) error {
		// open the bucket
		b := tx.Bucket([]byte(GroupsBucketName))

		// Checks if the bucket was opened, and creates a new one if not existing. Returns error on any other situation.
		if b == nil {
			return errors.New("Missing bucket")
		}

		var err error
		data := b.Get([]byte(GroupLengthKey))
		if data == nil {
			nextid = 0
		} else {
			nextid, err = strconv.Atoi(string(data))
			if err != nil {
				return err
			}
		}

		nextid++

		data = []byte(strconv.Itoa(nextid))

		err = b.Put([]byte(GroupLengthKey), data)
		if err != nil {
			return err
		}

		return nil

	}); err != nil {
		return -1
	}

	return nextid
}
