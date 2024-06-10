package main

import (
	"encoding/xml"
	"errors"
)

var ErrorNotFound = errors.New("not found")

func parseXrd(data []byte) (string, error) {
	var xrd struct {
		XMLName xml.Name `xml:"http://docs.oasis-open.org/ns/xri/xrd-1.0 XRD"`
		Link    []struct {
			Rel      string `xml:"rel,attr"`
			Template string `xml:"template,attr"`
		}
	}
	if err := xml.Unmarshal(data, &xrd); err != nil {
		return "", err
	}
	for _, link := range xrd.Link {
		if link.Rel == "lrdd" {
			return link.Template, nil
		}
	}
	return "", ErrorNotFound
}
