package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Address struct {
	User string
	Host string
}

var kAddressRe = regexp.MustCompile(`^(.*?)@(.*)$`)

func ParseAddress(id string) (*Address, error) {
	m := kAddressRe.FindStringSubmatch(id)
	if m == nil {
		return nil, errors.New("parse error")
	}
	return &Address{m[1], m[2]}, nil
}

// Account is cached? Return account
// Fetch host meta if not cached
// If host meta exists,

func (a Address) String() string {
	return a.User + "@" + a.Host
}

type FedirectHandler struct {
	// keys: user@host we would like to look up
	// values: profile page url
	accountCache   map[string]JRDLookupResult
	accountCacheMu sync.RWMutex
	// keys: host part of usernames
	// values: the web finger template url
	hostCache   map[string]string
	hostCacheMu sync.RWMutex
}

func NewFedirectHandler() *FedirectHandler {
	return &FedirectHandler{
		accountCache: make(map[string]JRDLookupResult),
		hostCache:    make(map[string]string),
	}
}

func (f *FedirectHandler) setHost(host string, url string) {
	f.hostCacheMu.Lock()
	defer f.hostCacheMu.Unlock()
	f.hostCache[host] = url
}

func (f *FedirectHandler) getHost(host string) (string, bool) {
	f.hostCacheMu.RLock()
	defer f.hostCacheMu.RUnlock()
	url, ok := f.hostCache[host]
	return url, ok
}

func (f *FedirectHandler) setAccount(name string, result JRDLookupResult) {
	f.accountCacheMu.Lock()
	defer f.accountCacheMu.Unlock()
	f.accountCache[name] = result
}

func (f *FedirectHandler) getAccount(name string) *JRDLookupResult {
	f.accountCacheMu.RLock()
	defer f.accountCacheMu.RUnlock()
	result, ok := f.accountCache[name]
	if ok {
		return &result
	} else {
		return nil
	}
}

func (f *FedirectHandler) LookupAccount(account Address) (*JRDLookupResult, error) {
	if cached := f.getAccount(account.String()); cached != nil {
		return cached, nil
	}

	template, err := f.LookupHost(account.Host)
	if err != nil {
		return nil, err
	}
	uri := url.QueryEscape("acct:" + account.String())
	webfingerUrl := strings.Replace(template, "{uri}", uri, 1)
	req, err := http.NewRequest("GET", webfingerUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "fedirect")
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
	jrd, err := parseJrd(body)
	if err != nil {
		return nil, err
	}
	f.setAccount(account.String(), *jrd)
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
	cached, ok := f.getHost(host)
	if ok {
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
	req.Header.Set("User-Agent", "fedirect")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		webfingerUrl := (&url.URL{
			Scheme:   "https",
			Host:     host,
			Path:     "/.well-known/webfinger",
			RawQuery: "resource={uri}",
		}).String()
		f.setHost(host, webfingerUrl)
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
	if err != nil {
		return "", err
	}
	f.setHost(host, template)
	return template, nil
}

func (f *FedirectHandler) DoLookup(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		http.Error(w, "?id=", http.StatusBadRequest)
		return
	}
	addr, err := ParseAddress(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profile, err := f.LookupAccount(*addr)
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
	case req.Method == http.MethodGet && req.URL.Path == "/":
		f.DoLookup(w, req)
	default:
		http.NotFound(w, req)
	}
}

func main() {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	listen := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Now listening on http://", ln.Addr())
	handler := NewFedirectHandler()
	http.Handle("/", handler)
	log.Fatal(http.Serve(ln, nil))
}
