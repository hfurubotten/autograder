package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hfurubotten/diskv"

	newgit "github.com/hfurubotten/autograder/entities"
	oldgit "github.com/hfurubotten/autograder/git"

	// "github.com/boltdb/bolt"
	"github.com/hfurubotten/autograder/ci"
	"github.com/hfurubotten/autograder/config"
	"github.com/hfurubotten/autograder/database"
)

func main() {
	// Signal backup warning and wait for confirmation.
	warningMsg()

	// load configuration values and store to JSON
	if err := loadConfigurations(); err != nil {
		log.Println("Error loading configurations:", err)
		return
	}

	database.Start(config.StandardBasePath + "autograder.db")
	defer database.Close()

	// load tokens. Actually no, those are encrypted and cant be found.

	// load users to database
	if err := convertUsers(); err != nil {
		log.Println("Error loading courses:", err)
		return
	}

	// Load orgs to database
	// Extract code reviews and store in database.
	// load groups to database
	if err := convertOrgs(); err != nil {
		log.Println("Error loading courses:", err)
		return
	}
}

func warningMsg() {
	fmt.Println("This will load the stored course and user information from the " +
		"old storage into the database. \nWarning: If you have stored data in the " +
		"database from before, this information might get overwritten if information" +
		"is stored under the same name.")

	fmt.Print("Do you want to continue? (Y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	action := strings.ToUpper(strings.TrimSpace(scanner.Text()))

	if action != "Y" && action != "YES" {
		fmt.Println("Shutting down.")
		os.Exit(1)
	}
}

func convertOrgs() error {
	log.Println("Porting Courses")
	oldorgs := oldgit.ListRegisteredOrganizations()
	for _, oldorg := range oldorgs {
		log.Println("Adding course:", oldorg.Name)

		neworg, err := newgit.NewOrganization(oldorg.Name, false)
		if err != nil {
			return err
		}

		neworg.TotalScore = oldorg.TotalScore
		neworg.WeeklyScore = oldorg.WeeklyScore
		neworg.MonthlyScore = oldorg.MonthlyScore
		neworg.TotalLeaderboard = oldorg.TotalLeaderboard
		neworg.WeeklyLeadboard = oldorg.WeeklyLeadboard
		neworg.MonthlyLeadboard = oldorg.MonthlyLeadboard
		neworg.TrackingWeek = oldorg.TrackingWeek
		neworg.TrackingMonth = oldorg.TrackingMonth

		neworg.ScreenName = oldorg.ScreenName
		neworg.Description = oldorg.Description
		neworg.Location = oldorg.Location
		neworg.Company = oldorg.Company
		neworg.HTMLURL = oldorg.HTMLURL
		neworg.AvatarURL = oldorg.AvatarURL

		neworg.GroupAssignments = oldorg.GroupAssignments
		neworg.IndividualAssignments = oldorg.IndividualAssignments
		neworg.IndividualLabFolders = oldorg.IndividualLabFolders
		neworg.GroupLabFolders = oldorg.GroupLabFolders
		neworg.IndividualDeadlines = oldorg.IndividualDeadlines
		neworg.GroupDeadlines = oldorg.GroupDeadlines
		neworg.StudentTeamID = oldorg.StudentTeamID
		neworg.OwnerTeamID = oldorg.OwnerTeamID
		neworg.Private = oldorg.Private

		neworg.PendingUser = oldorg.PendingUser
		neworg.Members = oldorg.Members
		neworg.Teachers = oldorg.Teachers

		neworg.CodeReview = oldorg.CodeReview
		neworg.AdminToken = oldorg.AdminToken
		neworg.CI = newgit.CIOptions{
			Basepath: oldorg.CI.Basepath,
			Secret:   oldorg.CI.Secret,
		}

		// Converting groups
		for gname := range oldorg.Groups {
			log.Println("Converting group:", gname)

			gnum, err := strconv.Atoi(gname[len("group"):])
			if err != nil {
				return err
			}

			oldgroup, err := oldgit.NewGroup(oldorg.Name, gnum)
			if err != nil {
				return err
			}

			nextgroupid := newgit.GetNextGroupID()
			if nextgroupid < 0 {
				return errors.New("Error finding next group id")
			}

			newgroup, err := newgit.NewGroup(oldorg.Name, nextgroupid, false)
			if err != nil {
				return err
			}

			newgroup.TeamID = oldgroup.TeamID
			newgroup.CurrentLabNum = oldgroup.CurrentLabNum
			newgroup.Course = oldgroup.Course
			newgroup.Members = oldgroup.Members

			// loading results for Group
			gcistore := GetCIStorage(oldorg.Name, "group"+strconv.Itoa(oldgroup.ID))
			keys := gcistore.Keys()
			for key := range keys {
				build, err := ci.NewBuildResult()
				if err != nil {
					log.Println(err)
					continue
				}
				err = gcistore.ReadGob(key, build, false)
				if err != nil {
					log.Println(err)
					continue
				}

				labnum := -1
				for i, name := range oldorg.GroupLabFolders {
					if name == key {
						labnum = i
						break
					}
				}
				if labnum < 0 {
					log.Println("No lab with that name found.")
					continue
				}

				newgroup.AddBuildResult(labnum, build.ID)

				if err = build.Save(); err != nil {
					log.Println(err)
					continue
				}
			}

			newgroup.Activate()
			if err = newgroup.Save(); err != nil {
				return err
			}

			neworg.Groups["group"+strconv.Itoa(newgroup.ID)] = nil
		}

		for user := range oldorg.Members {
			// loading results for user
			gcistore := GetCIStorage(oldorg.Name, user)
			keys := gcistore.Keys()
			for key := range keys {
				build, err := ci.NewBuildResult()
				if err != nil {
					log.Println(err)
					continue
				}
				err = gcistore.ReadGob(key, build, false)
				if err != nil {
					log.Println(err)
					continue
				}

				labnum := -1
				for i, name := range oldorg.IndividualLabFolders {
					if name == key {
						labnum = i
						break
					}
				}
				if labnum < 0 {
					log.Println("No lab with that name found.")
					continue
				}

				member, err := newgit.NewMemberFromUsername(user, false)
				if err != nil {
					log.Println(err)
					continue
				}

				member.AddBuildResult(oldorg.Name, labnum, build.ID)

				if err = build.Save(); err != nil {
					log.Println(err)
					continue
				}

				if err = member.Save(); err != nil {
					member.Unlock()
					log.Println(err)
					continue
				}
			}
		}

		// Extracting code reviews
		for _, oldcr := range oldorg.CodeReviewlist {
			newcr, err := newgit.NewCodeReview()
			if err != nil {
				return err
			}

			newcr.Title = oldcr.Title
			newcr.Ext = oldcr.Ext
			newcr.Desc = oldcr.Desc
			newcr.Code = oldcr.Code
			newcr.User = oldcr.User
			newcr.URL = oldcr.URL

			if err := newcr.Save(); err != nil {
				return err
			}

			neworg.CodeReviewlist = append(neworg.CodeReviewlist, newcr.ID)
		}

		if err = neworg.Save(); err != nil {
			return err
		}
	}

	return nil
}

func convertUsers() error {
	log.Println("Porting Users")
	oldusers := oldgit.ListAllMembers()
	for _, olduser := range oldusers {
		log.Println("Adding User:", olduser.Username)

		newuser, err := newgit.NewMemberFromUsername(olduser.Username, false)

		newuser.Name = olduser.Name
		newuser.Email = olduser.Email
		newuser.Location = olduser.Location
		newuser.Active = olduser.Active
		newuser.PublicProfile = olduser.PublicProfile
		newuser.TotalScore = olduser.TotalScore
		newuser.WeeklyScore = olduser.WeeklyScore
		newuser.MonthlyScore = olduser.MonthlyScore
		newuser.Level = olduser.Level
		newuser.Trophies = olduser.Trophies
		newuser.AvatarURL = olduser.AvatarURL
		newuser.ProfileURL = olduser.ProfileURL
		newuser.Scope = olduser.Scope

		newuser.StudentID = olduser.StudentID
		newuser.IsTeacher = olduser.IsTeacher
		newuser.IsAssistant = olduser.IsAssistant
		newuser.IsAdmin = olduser.IsAdmin
		newuser.Teaching = olduser.Teaching
		newuser.AssistantCourses = olduser.AssistantCourses

		for n, c := range olduser.Courses {
			newc := newgit.NewCourseOptions(c.Course)
			newc.CurrentLabNum = c.CurrentLabNum
			newuser.Courses[n] = newc
		}

		if err = newuser.Save(); err != nil {
			return err
		}
	}

	return nil
}

var optionstore = diskv.New(diskv.Options{
	BasePath:     "diskv/options/",
	CacheSizeMax: 1024 * 1024 * 256,
})

func loadConfigurations() error {
	var hname string
	err := optionstore.ReadGob("hostname", &hname, false)
	if err != nil {
		return err
	}

	var id string
	var secret string
	err = optionstore.ReadGob("OAuthID", &id, false)
	if err != nil {
		return err
	}

	err = optionstore.ReadGob("OAuthSecret", &secret, false)
	if err != nil {
		return err
	}

	conf := config.Configuration{
		Hostname:    hname,
		OAuthID:     id,
		OAuthSecret: secret,
	}

	if conf.Validate() != nil {
		if err := conf.QuickFix(); err != nil {
			return err
		}
	}

	return conf.Save()
}

// GetCIStorage will create a Diskv object used to store the test results.
func GetCIStorage(course, user string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:     "diskv/CI/" + course + "/" + user,
		CacheSizeMax: 1024 * 1024 * 256,
	})
}
