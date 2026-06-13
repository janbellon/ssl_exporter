package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (h *HTTPListener) GetRoot(w http.ResponseWriter, r *http.Request) {
	writeText(w, http.StatusOK, "Metric available at /metrics")
}

func (h *HTTPListener) GetHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, NewResponse(StatusOK, "service operational"))
}

func (l *HTTPListener) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if l.config.Listen.Bearer != "" {
		err := CheckAuthorization(r.Header["Authorization"], l.config.Listen.Bearer)
		if err != nil {
			writeJSON(w, http.StatusOK, NewResponse(StatusError, fmt.Sprint(err)))
			slog.Warn("GET /checktoken", "source_ip", r.RemoteAddr, "error", "Failed authentication")
			return
		}
	}
	targets := r.URL.Query()["target"]
	var metrics []MetricsResponse
	sslMetric := l.CollectMetrics(targets)
	metrics = append(metrics, *sslMetric)
	writeMetrics(w, 200, &metrics)
}
