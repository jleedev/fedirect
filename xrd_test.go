package main

import "testing"

func TestParseXrd(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<XRD xmlns="http://docs.oasis-open.org/ns/xri/xrd-1.0">
  <Link rel="lrdd" template="https://oulipo.social/.well-known/webfinger?resource={uri}"/>
</XRD>`
	result, err := parseXrd([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if result != "https://oulipo.social/.well-known/webfinger?resource={uri}" {
		t.Error(result)
	}
}
