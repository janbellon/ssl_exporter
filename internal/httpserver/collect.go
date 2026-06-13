package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
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

	const workers = 10

	targetCh := make(chan string)
	resultCh := make(chan Metric)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for target := range targetCh {
				targetMetric, err := l.checkTarget(target)
				if err != nil {
					slog.Warn(fmt.Sprint(err))
					continue
				}

				resultCh <- *targetMetric
			}
		}()
	}

	go func() {
		for _, target := range targets {
			targetCh <- target
		}

		close(targetCh)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for metric := range resultCh {
		metricsResponse.MetricsList = append(metricsResponse.MetricsList, metric)
	}

	return &metricsResponse
}

func (l *HTTPListener) checkTarget(target string) (*Metric, error) {
	url := "https://" + target

	slog.Info(fmt.Sprintf("Sending GET request to %s", url))

	req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("skipping target %s: %w", target, err)
	}

	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("skipping target %s: no TLS certificate found", target)
	}

	cert := resp.TLS.PeerCertificates[0]

	certDomain := cert.Subject.CommonName
	if certDomain == "" && len(cert.DNSNames) > 0 {
		certDomain = cert.DNSNames[0]
	}

	expirationDelay := time.Until(cert.NotAfter).Seconds()

	return &Metric{
		Labels: map[string]string{
			"target": target,
			"domain": certDomain,
		},
		Value: fmt.Sprintf("%.0f", expirationDelay),
	}, nil
}
