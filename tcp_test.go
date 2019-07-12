package health

import (
	"gotest.tools/assert"
	"net"
	"testing"
	"time"
)

func TestCheckHealthTCPCorrect(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	assert.NilError(t, err)
	defer listener.Close()

	h := NewHealthCheck(
		net.IPv4(127, 0, 0, 1),
		listener.Addr().(*net.TCPAddr).Port,
		DefaultTCPHealthCheckProvider,
		time.Second,
		60*time.Second,
		1*time.Second)

	h.CheckHealth()

	assert.Assert(t, h.Healthy)
}

func TestCheckHealthTCPIncorrect(t *testing.T) {
	h := NewHealthCheck(
		net.IPv4(127, 0, 0, 1),
		31337,
		DefaultTCPHealthCheckProvider,
		time.Second,
		60*time.Second,
		1*time.Second)

	h.CheckHealth()

	assert.Assert(t, !h.Healthy)
}
