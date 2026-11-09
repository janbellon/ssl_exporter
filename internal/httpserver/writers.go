package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Return JSON data with status, service and content
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Return simple text
func writeText(w http.ResponseWriter, status int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	w.Write([]byte(text))
}

var (
	labelValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	helpEscaper       = strings.NewReplacer(`\`, `\\`, "\n", `\n`)
)

// formatMetrics renders metrics in the Prometheus text exposition format.
// Label values come from the request and from remote certificates, so they
// are escaped to prevent injection of forged samples.
func formatMetrics(metrics []MetricsResponse) string {
	var b strings.Builder
	for _, metric := range metrics {
		fmt.Fprintf(&b, "# HELP %s %s\n", metric.MetricName, helpEscaper.Replace(metric.MetricHelp))
		fmt.Fprintf(&b, "# TYPE %s %s\n", metric.MetricName, metric.MetricType)
		for _, metricValue := range metric.MetricsList {
			keys := make([]string, 0, len(metricValue.Labels))
			for k := range metricValue.Labels {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			labelParts := make([]string, 0, len(keys))
			for _, k := range keys {
				v := strings.ToValidUTF8(metricValue.Labels[k], "\uFFFD")
				labelParts = append(labelParts, fmt.Sprintf(`%s="%s"`, k, labelValueEscaper.Replace(v)))
			}
			fmt.Fprintf(&b, "%s{%s} %s\n", metric.MetricName, strings.Join(labelParts, ","), metricValue.Value)
		}
	}
	return b.String()
}

func writeMetrics(w http.ResponseWriter, status int, metrics []MetricsResponse) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	w.Write([]byte(formatMetrics(metrics)))
}
