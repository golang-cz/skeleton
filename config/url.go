package config

import (
	"fmt"
	"net/url"

	"github.com/goware/urlx"
)

type URL struct {
	url.URL
}

func (u *URL) UnmarshalText(text []byte) error {
	var err error
	parsedUrl, err := urlx.Parse(string(text))
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	*u = URL{*parsedUrl}

	return nil
}

func (u *URL) String() string {
	return u.URL.String()
}
