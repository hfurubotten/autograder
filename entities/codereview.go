package git

import (
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/hfurubotten/autograder/database"
)

// CodeReviewBucketName is the bucket/table name in the database
var CodeReviewBucketName = "codereviews"

// CodeReviewLengtKey is the key used to count the ID for code reviews.
var CodeReviewLengtKey = "length"

func init() {
	//TODO Is this necessary?
	// gob.Register(CodeReview{})

	database.RegisterBucket(CodeReviewBucketName)
}

// CodeReview represent a code review stored in autograder.
type CodeReview struct {
	ID    int
	Title string
	Ext   string
	Desc  string
	Code  string
	User  string

	// Data from Github
	URL string
}

// NewCodeReview will create a new code review object.
func NewCodeReview() (*CodeReview, error) {
	nextid := GetNextCodeReviewID()
	if nextid < 0 {
		return nil, errors.New("Error occured while generating Build ID")
	}

	return &CodeReview{
		ID: nextid,
	}, nil
}

// GetCodeReview will get an already store code review from the database.
func GetCodeReview(reviewid int) (*CodeReview, error) {
	cr := &CodeReview{
		ID: reviewid,
	}

	if err := cr.loadStoredData(); err != nil {
		return nil, err
	}

	return cr, nil
}

func (cr *CodeReview) loadStoredData() error {
	return database.Get(CodeReviewBucketName, strconv.Itoa(cr.ID), cr)
}

// Save will store the code review to the database.
func (cr *CodeReview) Save() error {
	return database.Put(CodeReviewBucketName, strconv.Itoa(cr.ID), cr)
}

// GetNextCodeReviewID will find the next available CodeReview ID.
func GetNextCodeReviewID() int {
	nextid := -1
	err := database.GetPureDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CodeReviewBucketName))
		if b == nil {
			return errors.New("unknown bucket: " + CodeReviewBucketName)
		}

		var err error
		data := b.Get([]byte(CodeReviewLengtKey))
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
		err = b.Put([]byte(CodeReviewLengtKey), data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return -1
	}
	return nextid
}
