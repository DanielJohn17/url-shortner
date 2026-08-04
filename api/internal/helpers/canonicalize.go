package helpers

import (
	"fmt"
	"net/url"
	"strings"
)

func CanonicalizeBasic(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("Invalid URL")
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = u.Hostname()
	}

	if escaped := u.EscapedPath(); escaped != u.Path {
		u.RawPath = escaped
	}

	return u.String(), nil
}
