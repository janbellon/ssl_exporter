package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Return JSON data with status, service and content
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Return simple text
func writeText(w http.ResponseWriter, status int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(text))
}

func writeMetrics(w http.ResponseWriter, status int, metrics *[]MetricsResponse) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	var text string
	for _, metric := range *metrics {
		text += fmt.Sprintf("# HELP %s %s# TYPE %s %s\n", metric.MetricName, metric.MetricHelp, metric.MetricName, metric.MetricType)
		for _, metricValue := range metric.MetricsList {
			var labelParts []string
			for k, v := range metricValue.Labels {
				labelParts = append(labelParts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			labelString := strings.Join(labelParts, ",")
			text += fmt.Sprintf("%s{%s} %s\n",
				metric.MetricName,
				labelString,
				metricValue.Value,
			)
		}
	}
	w.Write([]byte(text))
}
