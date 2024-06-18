package main

import (
	"encoding/json"
	"slices"
	"strings"
)

func parseJrd(data []byte, lr LookupRequest) (*LookupResponse, error) {
	var jrd struct {
		Subject string
		Links   []struct {
			Rel        string
			Href       string
			Type       string
			Properties struct {
				Type string `json:"https://www.w3.org/ns/activitystreams#type"`
			}
		}
	}
	if err := json.Unmarshal([]byte(data), &jrd); err != nil {
		return nil, err
	}
	if lr.Type != "" {
		for _, link := range jrd.Links {
			if strings.EqualFold(lr.Type, link.Properties.Type) {
				return &LookupResponse{jrd.Subject, link.Href}, nil
			}
		}
		if !lr.Fallback {
			return nil, ErrorNotFound
		}
	}
	slices.Reverse(jrd.Links)
	for _, link := range jrd.Links {
		if link.Rel == "http://webfinger.net/rel/profile-page" || link.Rel == "https://webfinger.net/rel/profile-page" {
			return &LookupResponse{jrd.Subject, link.Href}, nil
		}
	}
	// Warning, no profile page found. It's possible this redirect may not be useful in a web browser, but only for an ActivityPub agent.
	for _, link := range jrd.Links {
		if link.Rel == "self" {
			return &LookupResponse{jrd.Subject, link.Href}, nil
		}
	}
	return nil, ErrorNotFound
}
