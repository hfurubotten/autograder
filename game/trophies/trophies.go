package trophies

const (
	LABELACTION int = 1 + iota
	ASSIGNACTION
	ISSUEACTION
	PUSHACTION
	CODEACTION
	TALKACTION
)

const (
	BRONCE_RANK int = 1 + iota
	SILVER_RANK
	GOLD_RANK
	PLATINUM_RANK
	ONYX_RANK
)

var RANKNAMES []string = []string{
	"Bronce",
	"Silver",
	"Gold",
	"Platinum",
	"Onyx",
}

var StandardThrophyChest *TrophyChest = &TrophyChest{
	Store: map[int]*Trophy{
		LABELACTION:  LabelerTrophy,
		ASSIGNACTION: AssignerTrophy,
		ISSUEACTION:  IssueTrophy,
		PUSHACTION:   PusherTrophy,
		TALKACTION:   TalkerTrophy,
		CODEACTION:   HackerTrophy,
	},
}

var (
	IssueTrophy *Trophy = &Trophy{
		Action: ISSUEACTION,
		Name:   "Issuer",
		Desc:   "React fast and lets people know when there is a problem.",
		Steps: []int{
			1, 10, 20,
		},
	}

	TalkerTrophy *Trophy = &Trophy{
		Action: TALKACTION,
		Name:   "Talker",
		Desc:   "Lets put this in a comment, shall we...",
		Steps: []int{
			1, 10, 20,
		},
	}

	HackerTrophy *Trophy = &Trophy{
		Action: CODEACTION,
		Name:   "Hacker",
		Desc:   "This is the person who implements line after line after line.",
		Steps: []int{
			1, 10, 20,
		},
	}

	PusherTrophy *Trophy = &Trophy{
		Action: PUSHACTION,
		Name:   "Pusher",
		Desc:   "Loves fast implementation and let people know about this right away.",
		Steps: []int{
			1, 10, 20,
		},
	}

	LabelerTrophy *Trophy = &Trophy{
		Action: LABELACTION,
		Name:   "Labeler",
		Desc:   "Keep it organized!",
		Steps: []int{
			1, 10, 20,
		},
	}

	AssignerTrophy *Trophy = &Trophy{
		Action: ASSIGNACTION,
		Name:   "Assigner",
		Desc:   "Knows who can deal best with each thing.",
		Steps: []int{
			1, 10, 20,
		},
	}
)

type TrophyChest struct {
	Store map[int]*Trophy
}

type TrophyHunter interface {
	GetTrophyChest() *TrophyChest
}

func NewTrophyChest() *TrophyChest {
	t := new(TrophyChest)
	t.Store = make(map[int]*Trophy)
	return t
}

type Trophy struct {
	Action      int
	Occurrences int
	Name        string
	Desc        string
	Steps       []int
	Rank        int
	RankName    string
}

//
func (t *Trophy) BumpRank() {
	if t.Rank > len(t.Steps) {
		return
	}

	if t.Occurrences >= t.Steps[t.Rank] {
		t.Rank += 1
		t.RankName = RANKNAMES[t.Rank-1]
	}
}
