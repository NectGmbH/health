package health

// DefaultNoneHealthCheckProvider defines a nop-healthcheck provider which always returns "up".
var DefaultNoneHealthCheckProvider = &NoneHealthCheckProvider{}

// NoneHealthCheckProvider is a dummy provider which does no checking at all and always returns "up"
type NoneHealthCheckProvider struct {
}

// CheckHealth with the none-provider always returns "up" for the current service.
func (c *NoneHealthCheckProvider) CheckHealth(h *HealthCheck) (string, bool) {
	return "unknown", true
}
