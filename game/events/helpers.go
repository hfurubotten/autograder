package events

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/hfurubotten/autograder/entities"
	"github.com/hfurubotten/autograder/game/points"
	"github.com/hfurubotten/autograder/game/trophies"
)

// DistributeScores is a helper function which update the scores on repos and users.
// This function will up also check if the supplied objects also implements the saver
// interface, and if so lock the writing and save the object when done.
// locking is handled internally.
func DistributeScores(score int, user points.SingleScorer, org points.MultiScorer) (err error) {
	if user == nil {
		panic("User parament cannot be nil when distributing scores.")
	}

	usaver, issaver := user.(entities.Saver)

	if issaver {
		usaver.Lock()
		defer func() {
			err = usaver.Save()
			if err != nil {
				usaver.Unlock()
				log.Println(err)
			}
		}()
	}

	user.IncScoreBy(score)

	if org != nil {
		osaver, issaver := org.(entities.Saver)

		if issaver {
			osaver.Lock()
			defer func() {
				err = osaver.Save()
				if err != nil {
					osaver.Unlock()
					log.Println(err)
				}
			}()
		}

		org.IncScoreBy(user.GetUsername(), score)
	}

	return
}

func RegisterAction(action int, user trophies.TrophyHunter) (err error) {
	usaver, issaver := user.(entities.Saver)

	if issaver {
		usaver.Lock()
		defer func() {
			err = usaver.Save()
			if err != nil {
				return
			}
		}()
	}
	chest := user.GetTrophyChest()

	trophy, ok := chest.Store[action]
	if !ok {
		trophy = trophies.StandardThrophyChest.Store[action]
		chest.Store[action] = trophy
	}

	trophy.Occurrences++
	trophy.BumpRank()

	return
}

// PanicHandler is a function that need to be called via a defer.
// It stopes a panicing go rutine and prints the stack trace for
// the panicing go rutine.
func PanicHandler(printstack bool) {
	if r := recover(); r != nil {
		log.Println("Recovered from panicing goroutine: ", r)
		if printstack {
			fmt.Println(string(debug.Stack()))
		}
	}
}
