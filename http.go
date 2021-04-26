package health

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// DefaultHTTPHealthCheckProvider holds a http health monitoring provider.
var DefaultHTTPHealthCheckProvider = &HTTPHealthCheckProvider{}
var DefaultHTTPSHealthCheckProvider = &HTTPHealthCheckProvider{
	HTTPS: true,
}

// HTTPHealthCheckProvider represents a HealthCheckProvider which monitors a http endpoint.
type HTTPHealthCheckProvider struct {
	HTTPS              bool
	InsecureSkipVerify bool
	Path               string
}

func NewHTTPHealthCheckProvider(https bool, insecureSkipVerify bool, path string) *HTTPHealthCheckProvider {
	return &HTTPHealthCheckProvider{
		HTTPS:              https,
		InsecureSkipVerify: insecureSkipVerify,
		Path:               path,
	}
}

// CheckHealth validates whether the current endpoint is up
func (c *HTTPHealthCheckProvider) CheckHealth(h *HealthCheck) (string, bool) {
	client := &http.Client{
		Timeout: h.MaxResponseTime,
	}

	if c.InsecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	schema := "http"
	if c.HTTPS {
		schema = "https"
	}

	path := c.Path
	if path == "" {
		path = "/healthz"
	} else if path[0] != '/' {
		path = "/" + path
	}

	resp, err := client.Get(schema + "://" + h.GetAddress() + path)
	if err != nil {
		return err.Error(), false
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Sprintf("status code is `%d`", resp.StatusCode), false
	}

	return "success", true
}
