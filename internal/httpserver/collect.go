package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type MetricsResponse struct {
	MetricType  string
	MetricName  string
	MetricsList []Metric
	MetricHelp  string
}

type Metric struct {
	Labels map[string]string
	Value  string
}

func (l *HTTPListener) CollectMetrics(targets []string) *MetricsResponse {
	metricsResponse := MetricsResponse{
		MetricType:  "gauge",
		MetricName:  "ssl_expiration_delay",
		MetricHelp:  "Delay before ssl certificate expiration",
		MetricsList: make([]Metric, 0),
	}

	for _, target := range targets {
		url := "https://" + target

		slog.Info(fmt.Sprintf("Sending GET request to %s", url))

		req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}

		resp, err := l.client.Do(req)
		if err != nil {
			// Errors handling
			slog.Warn(fmt.Sprintf("Skipping target %s: %v", target, err))
			continue
		}

		func() {
			defer resp.Body.Close()

			if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
				slog.Warn(fmt.Sprintf("Skipping target %s: no TLS certificate found", target))
				return
			}

			cert := resp.TLS.PeerCertificates[0]

			certDomain := cert.Subject.CommonName
			if certDomain == "" && len(cert.DNSNames) > 0 {
				certDomain = cert.DNSNames[0]
			}

			expirationDelay := time.Until(cert.NotAfter).Seconds()

			metricsResponse.MetricsList = append(metricsResponse.MetricsList, Metric{
				Labels: map[string]string{
					"target": target,
					"domain": certDomain,
				},
				Value: fmt.Sprintf("%.0f", expirationDelay),
			})
		}()
	}

	return &metricsResponse
}
