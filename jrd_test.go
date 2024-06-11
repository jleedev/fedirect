package main

import (
	"fmt"
	"testing"
)

func TestParseJrd(t *testing.T) {
	for i, tc := range []struct {
		Input   string
		Subject string
		Href    string
	}{
		{
			`{"subject":"acct:joshly@oulipo.social","aliases":["https://oulipo.social/@joshly","https://oulipo.social/users/joshly"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://oulipo.social/@joshly"},{"rel":"self","type":"application/activity+json","href":"https://oulipo.social/users/joshly"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://oulipo.social/authorize_interaction?uri={uri}"}]}`,
			"acct:joshly@oulipo.social",
			"https://oulipo.social/@joshly",
		},
		{
			`{"subject":"acct:thunderbird@tilvids.com","aliases":["https://tilvids.com/accounts/thunderbird"],"links":[{"rel":"self","type":"application/activity+json","href":"https://tilvids.com/accounts/thunderbird"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://tilvids.com/remote-interaction?uri={uri}"}]}`,
			"acct:thunderbird@tilvids.com",
			"https://tilvids.com/accounts/thunderbird",
		},
		{
			// per lemmy, mastodon seems to prioritize the last webfinger item in case of duplicates, and so they put the community last with this outcome in mind
			`{"subject":"acct:android@lemmy.world","links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://lemmy.world/u/Android","template":null},{"rel":"self","type":"application/activity+json","href":"https://lemmy.world/u/Android","template":null,"properties":{"https://www.w3.org/ns/activitystreams#type":"Person"}},{"rel":"http://ostatus.org/schema/1.0/subscribe","type":null,"href":null,"template":"https://lemmy.world/activitypub/externalInteraction?uri={uri}"},{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://lemmy.world/c/android","template":null},{"rel":"self","type":"application/activity+json","href":"https://lemmy.world/c/android","template":null,"properties":{"https://www.w3.org/ns/activitystreams#type":"Group"}}]}`,
			"acct:android@lemmy.world",
			"https://lemmy.world/c/android",
		},
	} {
		w := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result, err := parseJrd([]byte(w.Input))
			if err != nil {
				t.Fatal(err)
			}
			if *result != (JRDLookupResult{w.Subject, w.Href}) {
				t.Errorf("Expected %#v, got %#v", w.Href, result.Href)
			}
		})
	}
}
