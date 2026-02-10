package jikan

import (
	"net/http"
	"time"
)

type Option func(*Client)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.client = hc }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.client = &http.Client{
			Timeout:   d,
			Transport: http.DefaultTransport,
		}
	}
}

func WithCache(cache Cache, ttl time.Duration) Option {
	return func(c *Client) {
		c.cache = cache
		c.cacheTTL = ttl
	}
}

func WithRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}
