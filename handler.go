package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type LookupRequest struct {
	// the user@host of the request
	Id Address
	// empty string, or the type which was requested, converted to lowercase
	Type string
}

type LookupResponse struct {
	// the user@host of the webfinger response
	Subject string
	// the profile url to redirect to
	Href string
}

type FedirectHandler struct {
	// keys: user@host we would like to look up
	// values: profile page url
	accountCache RWMap[LookupRequest, LookupResponse]
	// keys: host part of usernames
	// values: the web finger template url
	hostCache RWMap[string, string]
}

func NewFedirectHandler() *FedirectHandler {
	return &FedirectHandler{
		accountCache: NewRWMap[LookupRequest, LookupResponse](),
		hostCache:    NewRWMap[string, string](),
	}
}

func (f *FedirectHandler) LookupAccount(lr LookupRequest) (*LookupResponse, error) {
	if cached, ok := f.accountCache.Get(lr); ok {
		return &cached, nil
	}
	template, err := f.LookupHost(lr.Id.Host)
	if err != nil {
		return nil, err
	}
	uri := url.QueryEscape("acct:" + lr.Id.String())
	webfingerUrl := strings.Replace(template, "{uri}", uri, 1)
	req, err := http.NewRequest("GET", webfingerUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", kUserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %#v from %#v", resp.Status, webfingerUrl)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jrd, err := parseJrd(body, lr.Type)
	if err != nil {
		return nil, err
	}
	f.accountCache.Set(lr, *jrd)
	return jrd, nil
}

// Determine the WebFinger URL for the hostname, or error if it was not possible to determine.
//
// 1. Fetch the https://{host}/.well-known/host-meta.
//
// 2. If this is a 404, return "https://{host}/.well-known/webfinger?resource={uri}", where {host} is substituted with the argument but {uri} is literal.
//
// 3. Parse the XML, extract `//XRD/Link[@rel="lrdd"]/@template`, and return that.
//
// This function is safe to call from multiple threads.
func (f *FedirectHandler) LookupHost(host string) (string, error) {
	if cached, ok := f.hostCache.Get(host); ok {
		return cached, nil
	}
	hostMetaUrl := (&url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/.well-known/host-meta",
	}).String()
	req, err := http.NewRequest("GET", hostMetaUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", kUserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	mediatype, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if resp.StatusCode == http.StatusNotFound || resp.Header.Get("Content-Length") == "0" || mediatype != "application/xrd+xml" {
		webfingerUrl := DefaultWebFinger(host).String()
		f.hostCache.Set(host, webfingerUrl)
		return webfingerUrl, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %#v from %#v", resp.Status, hostMetaUrl)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	template, err := parseXrd(body)
	if err == nil {
		f.hostCache.Set(host, template)
		return template, nil
	} else if err == ErrorNotFound {
		webfingerUrl := DefaultWebFinger(host).String()
		f.hostCache.Set(host, webfingerUrl)
		return webfingerUrl, nil
	} else {
		return "", err
	}
}

func (f *FedirectHandler) DoLookup(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		io.WriteString(w, kIndexHtml)
		if footer := footer(); footer != "" {
			fmt.Fprintf(w, "<p>\n<address>%v</address>\n", footer)
		}
		return
	}
	addr, err := ParseAddress(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wanted_type := req.FormValue("type")

	profile, err := f.LookupAccount(LookupRequest{*addr, wanted_type})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=259200, public")
	http.Redirect(w, req, profile.Href, http.StatusFound)

	fmt.Fprintf(w, "Request succeeded for %v\n", addr)
	resolvedAcct := strings.TrimPrefix(profile.Subject, "acct:")
	fmt.Fprintf(w, "Found %v at %v\n", resolvedAcct, profile.Href)
	fmt.Fprintf(w, "Redirecting …\n")
}

func (f *FedirectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/", req.Method == http.MethodHead && req.URL.Path == "/":
		f.DoLookup(w, req)
	default:
		http.NotFound(w, req)
	}
}
