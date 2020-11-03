package cmd

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	digestTag           = "digest-without-tag"
	FormatV1Tarball     = "v1-tarball"
	FormatLegacyTarball = "legacy-tarball"
	FormatV1Layout      = "v1-layout"
)

var showHash, formatManifest bool

func apiOptions() (bool, string, []remote.Option) {
	var (
		options = []remote.Option{}
		msg     []string
	)

	if password == "" && username == "" && proxyUrl == "" && !httpClient && !anonymous && pullWriteFormat == FormatV1Tarball {
		return true, "simple API", nil
	}

	msg = append(msg, "advanced API")

	switch {
	case anonymous:
		msg = append(msg, "Anonymous auth")
		options = append(options, remote.WithAuth(authn.Anonymous))
	case username != "" || password != "":
		msg = append(msg, "username password auth")
		options = append(options, remote.WithAuth(authn.FromConfig(authn.AuthConfig{Username: username, Password: password})))
	case httpClient:
		msg = append(msg, "custom http.Client")
		tr := &http.Transport{}
		options = append(options, remote.WithTransport(tr))
	case proxyUrl != "":
		msg = append(msg, "custom http.Client with proxy")
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			log.Fatalf("invalid proxy URL %s: %v", proxyUrl, err)
		}
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		options = append(options, remote.WithTransport(tr))
	}
	return false, strings.Join(msg, " "), options
}
