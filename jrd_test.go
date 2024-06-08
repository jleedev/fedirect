package main

import "testing"

func TestParseJrd(t *testing.T) {
	data := `{"subject":"acct:joshly@oulipo.social","aliases":["https://oulipo.social/@joshly","https://oulipo.social/users/joshly"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://oulipo.social/@joshly"},{"rel":"self","type":"application/activity+json","href":"https://oulipo.social/users/joshly"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://oulipo.social/authorize_interaction?uri={uri}"}]}`
	result, err := parseJrd([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if *result != (JRDLookupResult{"acct:joshly@oulipo.social", "https://oulipo.social/@joshly"}) {
		t.Error(result)
	}
}
