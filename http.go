package health

import (
	"fmt"
	"net/http"
)

// DefaultHTTPHealthCheckProvider holds a http health monitoring provider.
var DefaultHTTPHealthCheckProvider = &HTTPHealthCheckProvider{}

// HTTPHealthCheckProvider represents a HealthCheckProvider which monitors a http endpoint.
type HTTPHealthCheckProvider struct {
}

// CheckHealth validates whether the current endpoint is up
func (c *HTTPHealthCheckProvider) CheckHealth(h *HealthCheck) (string, bool) {
	client := &http.Client{
		Timeout: h.MaxResponseTime,
	}

	resp, err := client.Get("http://" + h.GetAddress() + "/healthz")
	if err != nil {
		return err.Error(), false
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Sprintf("status code is `%d`", resp.StatusCode), false
	}

	return "success", true
}
