package health

import (
	"gotest.tools/assert"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestCheckHealthHTTPCorrect(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	assert.NilError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	h := NewHealthCheck(
		net.IPv4(127, 0, 0, 1),
		port,
		DefaultHTTPHealthCheckProvider,
		time.Second,
		60*time.Second,
		1*time.Second)

	go (func() {
		hand := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		http.Serve(listener, hand)
	})()

	h.CheckHealth()

	assert.Assert(t, h.Healthy)
}

func TestCheckHealthHTTPIncorrect(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	assert.NilError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	h := NewHealthCheck(
		net.IPv4(127, 0, 0, 1),
		port,
		DefaultHTTPHealthCheckProvider,
		time.Second,
		60*time.Second,
		1*time.Second)

	go (func() {
		hand := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(418)
		})

		http.Serve(listener, hand)
	})()

	h.CheckHealth()

	assert.Assert(t, !h.Healthy)
}

func TestCheckHealthHTTPTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	assert.NilError(t, err)
	defer listener.Close()

	h := NewHealthCheck(
		net.IPv4(127, 0, 0, 1),
		listener.Addr().(*net.TCPAddr).Port,
		DefaultHTTPHealthCheckProvider,
		time.Second,
		60*time.Second,
		1*time.Second)

	go (func() {
		_, err := listener.Accept()
		if err != nil {
			t.Errorf("listener couldnt accept connection, see: %v", err)
		}
	})()

	timeBefore := time.Now()
	h.CheckHealth()
	timeAfter := time.Now()
	timeDiff := timeAfter.Sub(timeBefore).Seconds()

	assertLowerThan(t, timeDiff, 1.5, "timeout")
	assertBiggerThan(t, timeDiff, 0.5, "timeout")

	assert.Assert(t, !h.Healthy)
}
