package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"

	"golang.org/x/net/idna"
)

type Address struct {
	User string
	Host string
}

var kAddressRe = regexp.MustCompile(`^(.*?)@(.*)$`)

const kUserAgent string = "fedirect/0"

//go:embed index.html
var kIndexHtml string

func footer() string {
	revision := ""
	dirty := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value[:12]
			case "vcs.modified":
				if setting.Value == "true" {
					dirty = "-dirty"
				}
			}
		}
		if revision == "" {
			return info.Main.Version
		}
	}
	return revision + dirty
}

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
	label, err := idna.ToASCII(a.Host)
	if err == nil {
		return a.User + "@" + label
	} else {
		return a.User + "@" + a.Host
	}
}

func DefaultWebFinger(host string) *url.URL {
	return &url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     "/.well-known/webfinger",
		RawQuery: "resource={uri}",
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
	http.Handle("GET /{$}", handler)
	log.Fatal(http.Serve(ln, nil))
}
