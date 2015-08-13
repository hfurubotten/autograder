package git

import (
	"testing"
)

func TestNewCodeReview(t *testing.T) {
	startpoint := GetNextCodeReviewID()
	for i := 1 + startpoint; i <= testGetNextCodeReviewIDIterations; i++ {
		cr, err := NewCodeReview()
		if err != nil {
			t.Error("Error while creating a new code review:", err)
		}

		if cr.ID != i {
			t.Errorf("Creating a new code review didnt generate correct ID. Got %d, want %d.", cr.ID, i)
		}
	}
}

var testGetAndSaveCodeReviewInput = []*CodeReview{
	&CodeReview{
		ID:    1,
		Title: "Hello World",
		Ext:   "go",
		Desc:  "Hello World a",
		User:  "User1",
		URL:   "github.com/org/repo/1",
	},
	&CodeReview{
		ID:    2,
		Title: "Hello World 2",
		Ext:   "java",
		Desc:  "Hello World v",
		User:  "User1",
		URL:   "github.com/org/repo/2",
	},
	&CodeReview{
		ID:    3,
		Title: "Hello World 3",
		Ext:   "py",
		Desc:  "Hello World e",
		User:  "User3",
		URL:   "github.com/org/repo/3",
	},
	&CodeReview{
		ID:    4,
		Title: "World Hello",
		Ext:   "go",
		Desc:  "Hello World g",
		User:  "User4",
		URL:   "github.com/org/repo/4",
	},
	&CodeReview{
		ID:    5,
		Title: "Hello World",
		Ext:   "py",
		Desc:  "Hello World iiik",
		User:  "User5",
		URL:   "github.com/org/repo/5",
	},
}

func TestGetAndSaveCodeReview(t *testing.T) {
	for _, cr := range testGetAndSaveCodeReviewInput {
		if err := cr.Save(); err != nil {
			t.Error("Error saving code review:", err)
			continue
		}

		cr2, err := GetCodeReview(cr.ID)
		if err != nil {
			t.Error("Error gettin code review:", err)
			continue
		}

		compareCodeReviewObjects(cr, cr2, t)
	}
}

var testGetNextCodeReviewIDIterations = 100

func TestGetNextCodeReviewID(t *testing.T) {
	startpoint := GetNextCodeReviewID()
	for i := 1 + startpoint; i <= testGetNextCodeReviewIDIterations; i++ {
		nextID := GetNextCodeReviewID()
		if nextID != i {
			t.Errorf("Error with counting in getting next group ID. Got %d, want %d.", nextID, i)
		}
	}
}

func compareCodeReviewObjects(cr1, cr2 *CodeReview, t *testing.T) {
	if cr1.ID != cr2.ID {
		t.Errorf("Field value ID does not match. %v != %v", cr1.ID, cr2.ID)
	}

	if cr1.Title != cr2.Title {
		t.Errorf("Field value Title does not match. %v != %v", cr1.Title, cr2.Title)
	}

	if cr1.Ext != cr2.Ext {
		t.Errorf("Field value Ext does not match. %v != %v", cr1.Ext, cr2.Ext)
	}

	if cr1.Desc != cr2.Desc {
		t.Errorf("Field value Desc does not match. %v != %v", cr1.Desc, cr2.Desc)
	}

	if cr1.Code != cr2.Code {
		t.Errorf("Field value Code does not match. %v != %v", cr1.Code, cr2.Code)
	}

	if cr1.User != cr2.User {
		t.Errorf("Field value User does not match. %v != %v", cr1.User, cr2.User)
	}

	if cr1.URL != cr2.URL {
		t.Errorf("Field value URL does not match. %v != %v", cr1.URL, cr2.URL)
	}
}
