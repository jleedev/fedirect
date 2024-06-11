package main

import (
	"encoding/json"
	"slices"
	"strings"
)

func parseJrd(data []byte, wanted_type string) (*LookupResponse, error) {
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
	if wanted_type != "" {
		for _, link := range jrd.Links {
			if strings.EqualFold(wanted_type, link.Properties.Type) {
				return &LookupResponse{jrd.Subject, link.Href}, nil
			}
		}
		return nil, ErrorNotFound
	}
	slices.Reverse(jrd.Links)
	for _, link := range jrd.Links {
		if link.Rel == "http://webfinger.net/rel/profile-page" {
			return &LookupResponse{jrd.Subject, link.Href}, nil
		}
	}
	for _, link := range jrd.Links {
		if link.Rel == "self" {
			return &LookupResponse{jrd.Subject, link.Href}, nil
		}
	}
	return nil, ErrorNotFound
}
