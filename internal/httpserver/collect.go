package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
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

// CollectMetrics checks every target (label value -> dial address) and
// returns the expiration delay of their certificates. Targets are expected
// to have been validated with parseTargets.
func (l *HTTPListener) CollectMetrics(ctx context.Context, targets []string, addrs map[string]string) *MetricsResponse {
	metricsResponse := MetricsResponse{
		MetricType:  "gauge",
		MetricName:  "ssl_expiration_delay",
		MetricHelp:  "Delay before ssl certificate expiration",
		MetricsList: make([]Metric, 0, len(targets)),
	}

	const workers = 10

	results := make([]*Metric, len(targets))
	indexCh := make(chan int)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for idx := range indexCh {
				target := targets[idx]
				targetMetric, err := l.checkTarget(ctx, target, addrs[target])
				if err != nil {
					slog.Warn("Skipping target", "target", target, "error", err.Error())
					continue
				}
				results[idx] = targetMetric
			}
		}()
	}

feed:
	for idx := range targets {
		select {
		case indexCh <- idx:
		case <-ctx.Done():
			break feed
		}
	}
	close(indexCh)
	wg.Wait()

	for _, metric := range results {
		if metric != nil {
			metricsResponse.MetricsList = append(metricsResponse.MetricsList, *metric)
		}
	}

	return &metricsResponse
}

// checkTarget performs a TLS handshake with the target and reads the leaf
// certificate. No application request is sent: nothing can be redirected and
// no response body is ever read from the target.
func (l *HTTPListener) checkTarget(ctx context.Context, target string, addr string) (*Metric, error) {
	slog.Debug("Checking target", "target", target, "address", addr)

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{},
		Config: &tls.Config{
			// Verification is deliberately disabled: the exporter must be
			// able to report on expired, self-signed or otherwise invalid
			// certificates. The connection is closed right after the
			// handshake and no data is exchanged over it.
			InsecureSkipVerify: true, // #nosec G402
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.(*tls.Conn).ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no TLS certificate found")
	}

	cert := certs[0]

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
