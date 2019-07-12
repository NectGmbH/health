package health

import (
	"net"
)

// DefaultTCPHealthCheckProvider holds a tcp health monitoring provider.
var DefaultTCPHealthCheckProvider = &TCPHealthCheckProvider{}

// TCPHealthCheckProvider represents a HealthCheckProvider which monitors a tcp endpoint.
type TCPHealthCheckProvider struct {
}

// CheckHealth validates whether the current endpoint is up
func (c *TCPHealthCheckProvider) CheckHealth(h *HealthCheck) (string, bool) {
	con, err := net.DialTimeout("tcp", h.GetAddress(), h.MaxResponseTime)
	if err != nil {
		return err.Error(), false
	}

	defer con.Close()

	return "success", true
}
