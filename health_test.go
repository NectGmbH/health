package health

import (
    "fmt"
    "gotest.tools/assert"
    "net"
    "testing"
    "time"
)

type MockHealthCheckProvider struct {
    MockFunc func(h *HealthCheck) (string, bool)
}

func NewMockHealthCheckProvider(
    mock func(h *HealthCheck) (string, bool)) *MockHealthCheckProvider {
    return &MockHealthCheckProvider{MockFunc: mock}
}

func (c *MockHealthCheckProvider) CheckHealth(h *HealthCheck) (string, bool) {
    return c.MockFunc(h)
}

func TestCheckHealthCorrect(t *testing.T) {
    mockFunc := func(h *HealthCheck) (string, bool) {
        return "message", true
    }

    h := NewHealthCheck(
        net.IPv4(0, 0, 0, 0),
        0,
        NewMockHealthCheckProvider(mockFunc),
        time.Second,
        60*time.Second,
        1*time.Second)

    timeBefore := time.Now()

    h.CheckHealth()

    assert.Assert(t, h.Healthy)

    assertTimeBetweenTimes(
        t, h.LastCheck, timeBefore, time.Now(), "LastCheck date incorrect")

    assert.Equal(t, h.LastCheck, h.LastTimeHealthy)

    if h.Retention < defaultRetention || h.Retention > time.Duration(1.5*float64(defaultRetention)) {
        t.Fatalf("expected retention to be around 1s, but its %s", h.Retention.String())
    }

    assert.Equal(t, h.LastMessage, "message")

    timeBefore = time.Now()
    h.CheckHealth()

    assert.Assert(t, h.Healthy)

    assertTimeBetweenTimes(
        t, h.LastCheck, timeBefore, time.Now(), "LastCheck date incorrect")

    assert.Equal(t, h.LastCheck, h.LastTimeHealthy)

    if h.Retention < defaultRetention || h.Retention > time.Duration(1.5*float64(defaultRetention)) {
        t.Fatalf("expected retention to be around 1s, but its %s", h.Retention.String())
    }
}

func TestCheckHealthIncorrectRetention(t *testing.T) {
    mockFunc := func(h *HealthCheck) (string, bool) {
        return "message", false
    }

    h := NewHealthCheck(
        net.IPv4(0, 0, 0, 0),
        0,
        NewMockHealthCheckProvider(mockFunc),
        time.Second,
        60*time.Second,
        1*time.Second)

    timeBefore := time.Now()

    h.CheckHealth()

    assert.Assert(t, !h.Healthy)

    assertTimeBetweenTimes(
        t, h.LastCheck, timeBefore, time.Now(), "LastCheck date incorrect")

    assertTimeBefore(t, h.LastTimeHealthy, timeBefore, "LastTimeHealthy")

    if h.Retention < 2*defaultRetention || h.Retention > time.Duration(2.5*float64(defaultRetention)) {
        t.Fatalf("expected retention to be around 2s, but its %s", h.Retention.String())
    }

    assert.Equal(t, h.LastMessage, "message")

    timeBefore = time.Now()
    h.CheckHealth()

    assert.Assert(t, !h.Healthy)

    assertTimeBetweenTimes(
        t, h.LastCheck, timeBefore, time.Now(), "LastCheck date incorrect")

    if h.Retention < 3*defaultRetention || h.Retention > 4*defaultRetention {
        t.Fatalf("expected retention to be around 3.5s, but its %s", h.Retention.String())
    }
}

func TestMonitor(t *testing.T) {
    i := 0

    mockFunc := func(h *HealthCheck) (string, bool) {
        i++
        return fmt.Sprintf("msg %d", i), i < 5
    }

    h := NewHealthCheck(
        net.IPv4(42, 42, 42, 42),
        1337,
        NewMockHealthCheckProvider(mockFunc),
        time.Second,
        60*time.Second,
        1*time.Second)

    stopChan := make(chan struct{})
    defer close(stopChan)

    notificationChan := h.Monitor(stopChan)

    for i2 := 1; i < 8; i2++ {
        timeBefore := time.Now()
        status := <-notificationChan
        timeAfter := time.Now()

        assert.DeepEqual(t, status.IP, net.IPv4(42, 42, 42, 42))
        assert.Equal(t, status.Port, 1337)
        assert.Equal(t, status.Healthy, i2 < 5)
        assert.Equal(t, status.Message, fmt.Sprintf("msg %d", i2))

        timeDiff := timeAfter.Sub(timeBefore).Seconds()

        if i2 >= 5 {
            assertLowerThan(t, timeDiff, float64(i2-4)+1, "retention low")
            assertBiggerThan(t, timeDiff, 1+float64(i2-5), "retention big")
        } else {
            if h.Retention < defaultRetention || h.Retention > time.Duration(1.5*float64(defaultRetention)) {
                t.Fatalf("expected retention to be around 1s, but its %s", h.Retention.String())
            }

            assertLowerThan(t, timeDiff, 1.5, "retention low")
        }
    }
}

func assertLowerThan(t *testing.T, a float64, b float64, msg string) {
    if a >= b {
        t.Errorf("Expected `%v` to be lower than `%v`: %v", a, b, msg)
    }
}

func assertBiggerThan(t *testing.T, a float64, b float64, msg string) {
    if a <= b {
        t.Errorf("Expected `%v` to be bigger than `%v`: %v", a, b, msg)
    }
}

func assertTimeBetweenTimes(
    t *testing.T,
    testDate time.Time,
    lowerTime time.Time,
    upperTime time.Time,
    message string) {

    if testDate.After(lowerTime) && testDate.Before(upperTime) {
        return
    }

    t.Errorf(
        "Expected timestamp `%s` to be between `%s` and `%s`: %s",
        testDate,
        lowerTime,
        upperTime,
        message)
}

func assertTimeBefore(t *testing.T, test time.Time, lo time.Time, msg string) {
    if !test.Before(lo) {
        t.Errorf(
            "Expected timestamp `%s` to be before `%s`: %s",
            test,
            lo,
            msg)
    }
}
