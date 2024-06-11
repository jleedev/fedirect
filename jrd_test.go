package main

import (
	"fmt"
	"testing"
)

func TestParseJrd(t *testing.T) {
	const LemmyInput string = `{"subject":"acct:android@lemmy.world","links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://lemmy.world/u/Android","template":null},{"rel":"self","type":"application/activity+json","href":"https://lemmy.world/u/Android","template":null,"properties":{"https://www.w3.org/ns/activitystreams#type":"Person"}},{"rel":"http://ostatus.org/schema/1.0/subscribe","type":null,"href":null,"template":"https://lemmy.world/activitypub/externalInteraction?uri={uri}"},{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://lemmy.world/c/android","template":null},{"rel":"self","type":"application/activity+json","href":"https://lemmy.world/c/android","template":null,"properties":{"https://www.w3.org/ns/activitystreams#type":"Group"}}]}`
	for i, tc := range []struct {
		Input   string
		Type    string
		Subject string
		Href    string
	}{
		{
			Input:   `{"subject":"acct:joshly@oulipo.social","aliases":["https://oulipo.social/@joshly","https://oulipo.social/users/joshly"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://oulipo.social/@joshly"},{"rel":"self","type":"application/activity+json","href":"https://oulipo.social/users/joshly"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://oulipo.social/authorize_interaction?uri={uri}"}]}`,
			Subject: "acct:joshly@oulipo.social",
			Href:    "https://oulipo.social/@joshly",
		},
		{
			Input:   `{"subject":"acct:thunderbird@tilvids.com","aliases":["https://tilvids.com/accounts/thunderbird"],"links":[{"rel":"self","type":"application/activity+json","href":"https://tilvids.com/accounts/thunderbird"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://tilvids.com/remote-interaction?uri={uri}"}]}`,
			Subject: "acct:thunderbird@tilvids.com",
			Href:    "https://tilvids.com/accounts/thunderbird",
		},
		{
			// per lemmy, mastodon seems to prioritize the last webfinger item in case of duplicates, and so they put the community last with this outcome in mind
			Input:   LemmyInput,
			Subject: "acct:android@lemmy.world",
			Href:    "https://lemmy.world/c/android",
		},
		{
			Input:   LemmyInput,
			Type:    "Person",
			Subject: "acct:android@lemmy.world",
			Href:    "https://lemmy.world/u/Android",
		},
		{
			Input:   LemmyInput,
			Type:    "Group",
			Subject: "acct:android@lemmy.world",
			Href:    "https://lemmy.world/c/android",
		},
	} {
		w := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result, err := parseJrd([]byte(w.Input), w.Type)
			if err != nil {
				t.Fatal(err)
			}
			if *result != (LookupResponse{w.Subject, w.Href}) {
				t.Errorf("Expected %#v, got %#v", w.Href, result.Href)
			}
		})
	}
}
