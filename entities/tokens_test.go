package git

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

func TestNewToken(t *testing.T) {
	for _, hash := range testNewTokenInput {
		token := NewToken(hash)
		if token.accessToken != hash {
			t.Errorf("Failed to set token. Got \"%v\" and not \"%v\" from token object ", token.accessToken, hash)
		}
	}
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
		token := NewToken(in.Token)

		if token.HasTokenInStore() {
			t.Errorf("Found the token \"%v\" in store before save.", in.Token)
		}

		err := token.SetUsernameToTokenInStore(in.Username)
		if err != nil {
			t.Error("Error saving token in database:", err)
			continue
		}

		if !token.HasTokenInStore() {
			t.Errorf("Could not find the token \"%v\" in store before saving it to username \"%v\".", in.Token, in.Username)
		}

		username, err := token.GetUsernameFromTokenInStore()
		if err != nil {
			t.Error("Error while getting username from token store:", err)
		}

		if username != in.Username {
			t.Errorf("Username gotten from token store does not match saved username. %v != %v", username, in.Username)
		}

		token.RemoveTokenInStore()

		if token.HasTokenInStore() {
			t.Errorf("Could find the token \"%v\" in store after removing it.", in.Token)
		}
	}
}

func TestHasAndGetToken(t *testing.T) {
	for _, hash := range testNewTokenInput {
		token := Token{}
		if token.HasToken() {
			t.Error("Found a token when it should not be any there")
		}

		token = NewToken(hash)

		if !token.HasToken() {
			t.Error("Found no token when it should be there")
		}

		storedhash := token.GetToken()
		if storedhash != hash {
			t.Errorf("Wrong token gotten from the token object. %v != %v", storedhash, hash)
		}
	}
}
