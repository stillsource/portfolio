package di

import (
	"net"
	"net/http"
	"time"
)

// GetHTTPClient returns the HTTP client shared by every drive infrastructure
// adapter. Connection pooling is tuned for concurrent downloads.
func (c *Container) GetHTTPClient() *http.Client {
	if c.httpClient == nil {
		timeout := c.cfg.HTTPTimeout
		if timeout <= 0 {
			timeout = 60
		}
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          64,
			MaxIdleConnsPerHost:   16,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		c.httpClient = &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeout) * time.Second,
		}
	}
	return c.httpClient
}
