package entities

import (
	"testing"
)

var testNewTokenInput = []string{
	"54143585152ac4",
	"5414afdc52ac4",
	"54143585152ac4",
	"654131583aad4542dd",
	"855d5c32b31a35d35",
	"adc531351d53ddd",
	"cdv3513513",
	"789564321dcbad",
}

var testHasTokenInStoreInput = []struct {
	Token, Username string
}{
	{"abfc13a1c351a3cf51531cfa", "user1"},
	{"abfc12435abcd1531cfa", "user2"},
	{"15313543513513cf51531cfa", "user3"},
	{"abfc13a1c351a3cf515135315", "user4"},
	{"abfc13a1c351a3cf51531", "user5"},
	{"13a1c351a3cf51531cfa", "user6"},
	{"abfc15135ad3551c3a1c351a3cf51531cfa", "user7"},
	{"abfc13a1c3fdbdbdb51531cfa", "user8"},
	{"abfc13a1cdd55d5d55d1531cfa", "user9"},
	{"1381535a5a31", "user10"},
}

func TestHasSetGetAndRemoveTokenInStore(t *testing.T) {
	for _, in := range testHasTokenInStoreInput {
		if has(in.Token) {
			t.Errorf("Found the token \"%v\" in store before save.", in.Token)
		}

		err := put(in.Token, in.Username)
		if err != nil {
			t.Error("Error saving token in database:", err)
			continue
		}

		if !has(in.Token) {
			t.Errorf("Could not find the token \"%v\" in store before saving it to username \"%v\".", in.Token, in.Username)
		}

		username, err := get(in.Token)
		if err != nil {
			t.Error("Error while getting username from token store:", err)
		}
		if username != in.Username {
			t.Errorf("Username gotten from token store does not match saved username. %v != %v", username, in.Username)
		}

		remove(in.Token)
		if has(in.Token) {
			t.Errorf("Could find the token \"%v\" in store after removing it.", in.Token)
		}
	}
}
