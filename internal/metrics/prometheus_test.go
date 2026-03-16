package metrics

import "testing"

func TestPrometheusCollectors_Usable(t *testing.T) {
	if RequestCounter == nil {
		t.Fatal("RequestCounter must be initialized")
	}
	if RequestDuration == nil {
		t.Fatal("RequestDuration must be initialized")
	}

	RequestCounter.WithLabelValues("GET", "/health", "OK").Inc()
	RequestDuration.WithLabelValues("GET", "/health").Observe(0.01)
}
