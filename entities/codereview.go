package git

import (
	"errors"
	"fmt"
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

type codeReviewID int

func (id *codeReviewID) String() string {
	return strconv.Itoa(int(*id))
}

// CodeReview represent a code review stored in autograder.
type CodeReview struct {
	ID    codeReviewID
	Title string
	Ext   string
	Desc  string
	Code  string
	User  string

	// Data from Github
	URL string
}

func (cr *CodeReview) String() string {
	return fmt.Sprintf(
		"ID: %v, Title: %s, Ext: %s, Desc: %s, Code: %s, User: %s, URL: %s",
		cr.ID, cr.Title, cr.Ext, cr.Desc, cr.Code, cr.User, cr.URL)
}

// Equal returns true if cr equals other.
func (cr *CodeReview) Equal(other *CodeReview) bool {
	return cr.ID == other.ID &&
		cr.Title == other.Title &&
		cr.Ext == other.Ext &&
		cr.Desc == other.Desc &&
		cr.Code == other.Code &&
		cr.User == other.User &&
		cr.URL == other.URL
}

// NewCodeReview creates a new code review object.
func NewCodeReview() (*CodeReview, error) {
	nextid, err := nextCodeReviewID()
	if err != nil {
		return nil, err
	}
	return &CodeReview{ID: nextid}, nil
}

// GetCodeReview returns the code review for the given reviewID.
func GetCodeReview(id codeReviewID) (*CodeReview, error) {
	var cr *CodeReview
	err := database.Get(CodeReviewBucketName, id.String(), &cr)
	if err != nil {
		return nil, err
	}
	return cr, nil
}

// Save stores the code review to the database.
func (cr *CodeReview) Save() error {
	return database.Put(CodeReviewBucketName, cr.ID.String(), cr)
}

// nextCodeReviewID will find the next available CodeReview ID.
func nextCodeReviewID() (nextid codeReviewID, err error) {
	err = database.GetPureDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CodeReviewBucketName))
		if b == nil {
			return errors.New("unknown bucket: " + CodeReviewBucketName)
		}

		var er error
		data := b.Get([]byte(CodeReviewLengtKey))
		if data != nil {
			er = database.GobDecode(data, &nextid)
			if er != nil {
				return er
			}
		}

		nextid++
		data, er = database.GobEncode(nextid)
		if er != nil {
			return er
		}
		return b.Put([]byte(CodeReviewLengtKey), data)
	})

	return nextid, err
}
