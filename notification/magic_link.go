package notification

import (
	"net/url"
	"time"
)

type MagicLinkData struct {
	Origin    *url.URL
	TTL       time.Duration
	MagicLink *url.URL
}
