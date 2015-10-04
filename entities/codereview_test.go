package git

import (
	"testing"
)

var iter = 100

func TestNewCodeReview(t *testing.T) {
	first, err := NewCodeReview()
	if err != nil {
		t.Error("Error creating new code review: ", err)
	}
	if int(first.ID) != 1 {
		t.Errorf("Got %v, wanted 1.", first.ID)
	}

	for i := int(first.ID) + 1; i <= iter; i++ {
		nxt, err := NewCodeReview()
		if err != nil {
			t.Error("Error creating new code review: ", err)
		}
		if int(nxt.ID) != i {
			t.Errorf("Got %v, wanted %d.", nxt.ID, i)
		}
	}
}

func TestConcurrentNewCodeReview(t *testing.T) {
	first, err := NewCodeReview()
	if err != nil {
		t.Error("Error creating new code review: ", err)
	}
	if int(first.ID) != 101 {
		t.Errorf("Got %v, wanted 1.", first.ID)
	}

	for i := int(first.ID) + 1; i <= iter; i++ {
		go func() {
			nxt, err := NewCodeReview()
			if err != nil {
				t.Error("Error creating new code review: ", err)
			}
			if int(nxt.ID) != i {
				t.Errorf("Got %v, wanted %d.", nxt.ID, i)
			}
		}()
	}
}

var crInput = []*CodeReview{
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

func TestSaveAndGetCodeReview(t *testing.T) {
	for _, cr := range crInput {
		if err := cr.Save(); err != nil {
			t.Error("Error saving code review:", err)
			continue
		}

		cr2, err := GetCodeReview(cr.ID)
		if err != nil {
			t.Error("Error getting code review:", err)
			continue
		}
		if !cr.Equal(cr2) {
			t.Errorf("\ncr : %v\ncr2: %v\n", cr, cr2)
		}
	}
}
