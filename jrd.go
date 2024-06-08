package main

import (
	"encoding/json"
	"errors"
)

type JRDLookupResult struct {
	Subject string
	Href    string
}

func parseJrd(data []byte) (*JRDLookupResult, error) {
	var jrd struct {
		Subject string `json:"subject"`
		Link    []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := json.Unmarshal([]byte(data), &jrd); err != nil {
		return nil, err
	}
	for _, link := range jrd.Link {
		if link.Rel == "http://webfinger.net/rel/profile-page" {
			return &JRDLookupResult{jrd.Subject, link.Href}, nil
		}
	}
	for _, link := range jrd.Link {
		if link.Rel == "self" {
			return &JRDLookupResult{jrd.Subject, link.Href}, nil
		}
	}
	return nil, errors.New("not found")
}
