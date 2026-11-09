package httpserver

import (
	"log/slog"
	"net/http"
)

func (l *HTTPListener) GetRoot(w http.ResponseWriter, r *http.Request) {
	writeText(w, http.StatusOK, "Metric available at /metrics")
}

func (l *HTTPListener) GetHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, NewResponse(StatusOK, "service operational"))
}

func (l *HTTPListener) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if l.config.Listen.Bearer != "" {
		if !CheckAuthorization(r.Header.Get("Authorization"), l.config.Listen.Bearer) {
			slog.Warn("GET /metrics", "source_ip", r.RemoteAddr, "error", "Failed authentication")
			w.Header().Set("WWW-Authenticate", `Bearer realm="ssl_exporter"`)
			writeJSON(w, http.StatusUnauthorized, NewResponse(StatusError, "unauthorized"))
			return
		}
	}

	targets, addrs, err := parseTargets(r.URL.Query()["target"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, NewResponse(StatusError, err.Error()))
		return
	}

	sslMetric := l.CollectMetrics(r.Context(), targets, addrs)
	if r.Context().Err() != nil {
		// Client went away (or scrape timed out): nothing to answer.
		return
	}
	writeMetrics(w, http.StatusOK, []MetricsResponse{*sslMetric})
}
