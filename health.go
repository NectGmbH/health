package health

import (
    "fmt"
    "math/rand"
    "net"
    "strings"
    "time"

    "github.com/golang/glog"
)

const defaultRetention = time.Second

func GetHealthCheckProvider(name string) (HealthCheckProvider, error) {
    name = strings.ToLower(strings.TrimSpace(name))

    if name == "none" {
        return DefaultNoneHealthCheckProvider, nil
    } else if name == "tcp" {
        return DefaultTCPHealthCheckProvider, nil
    } else if name == "http" {
        return DefaultHTTPHealthCheckProvider, nil
    } else if name == "https" {
        return DefaultHTTPSHealthCheckProvider, nil
    }

    return nil, fmt.Errorf("unknown health check protocol `%s` use either \"none\", \"tcp\" or \"http\"", name)
}

// HealthCheck represents a health monitoring resource
type HealthCheck struct {
    IP               net.IP
    Port             int
    Provider         HealthCheckProvider
    Healthy          bool
    LastTimeHealthy  time.Time
    LastCheck        time.Time
    LastMessage      string
    PlannedRetention time.Duration
    Retention        time.Duration
    MaxRetention     time.Duration
    MaxResponseTime  time.Duration
}

// HealthCheckStatus represents the current status of a HealthCheck endpoint.
type HealthCheckStatus struct {
    IP        net.IP
    Port      int
    Healthy   bool
    Message   string
    DidChange bool
}

// GetAddress returns the endpoint (i.e. 127.0.0.1:80) of the current HealthCheckStatus.
func (h *HealthCheckStatus) GetAddress() string {
    return fmt.Sprintf("%s:%d", h.IP, h.Port)
}

// String returns a string representation of the current status.
func (s *HealthCheckStatus) String() string {
    sign := "UP"

    if !s.Healthy {
        sign = "DOWN"
    }

    return fmt.Sprintf("%s %s:%d - %s", sign, s.IP, s.Port, s.Message)
}

// NewHealthCheck creates a new HealthCheck instance with the specified parameters.
func NewHealthCheck(
    ip net.IP,
    port int,
    provider HealthCheckProvider,
    plannedRetention time.Duration,
    maxRetention time.Duration,
    maxResponseTime time.Duration,
) *HealthCheck {
    h := &HealthCheck{
        IP:               ip,
        Port:             port,
        Provider:         provider,
        Healthy:          false,
        PlannedRetention: plannedRetention,
        Retention:        plannedRetention,
        MaxRetention:     maxRetention,
        MaxResponseTime:  maxResponseTime,
    }

    return h
}

// GetAddress returns the endpoint (i.e. 127.0.0.1:80) of the current HealthCheck.
func (h *HealthCheck) GetAddress() string {
    return fmt.Sprintf("%s:%d", h.IP, h.Port)
}

// Monitor starts monitoring the endpoint configured in the HealthCheck.
func (h *HealthCheck) Monitor(stopChan chan struct{}) chan HealthCheckStatus {
    notificationChan := make(chan HealthCheckStatus)

    go (func() {
        // Add some random delay so not all healthchecks happen at the very same second
        time.Sleep(time.Duration(rand.Float64() * float64(time.Second)))

        glog.V(5).Infof("Starting monitoring %s:%d", h.IP, h.Port)

        for {
            select {
            case <-stopChan:
                glog.V(5).Infof("Stopped monitoring %s:%d", h.IP, h.Port)
                close(notificationChan)
                return
            default:
            }

            isFirst := h.LastCheck.IsZero()
            before := h.Healthy
            h.CheckHealth()
            after := h.Healthy

            notificationChan <- HealthCheckStatus{
                IP:        h.IP,
                Port:      h.Port,
                Healthy:   h.Healthy,
                Message:   h.LastMessage,
                DidChange: isFirst || after != before,
            }

            time.Sleep(h.Retention)
        }
    })()

    return notificationChan
}

// CheckHealth updates the current HealthCheck (e.g. the healthy-property)
func (h *HealthCheck) CheckHealth() {
    h.LastMessage, h.Healthy = h.Provider.CheckHealth(h)

    // Add some randomness so not all checks get executed at the same time
    retention := h.PlannedRetention + time.Duration((rand.Float64()/2)*float64(time.Second))

    h.LastCheck = time.Now()
    if h.Healthy {
        h.LastTimeHealthy = h.LastCheck
        h.Retention = retention
    } else if h.Retention < h.MaxRetention {
        h.Retention += defaultRetention
    }
}

// HealthCheckProvider defines any implementation of the CheckHealth func
type HealthCheckProvider interface {
    CheckHealth(healthCheck *HealthCheck) (string, bool)
}
