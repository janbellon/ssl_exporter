package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ssl-exporter/internal/config"
)

func TestCheckAuthorization(t *testing.T) {
	cases := []struct {
		header string
		want   bool
	}{
		{"Bearer secret", true},
		{"bearer secret", true},
		{"Bearer wrong", false},
		{"Bearer secretx", false},
		{"Bearer ", false},
		{"Basic secret", false},
		{"secret", false},
		{"", false},
	}
	for _, c := range cases {
		if got := CheckAuthorization(c.header, "secret"); got != c.want {
			t.Errorf("CheckAuthorization(%q) = %v, want %v", c.header, got, c.want)
		}
	}
}

func TestParseTarget(t *testing.T) {
	valid := map[string]string{
		"example.com":      "example.com:443",
		"example.com:8443": "example.com:8443",
		"10.0.0.1:443":     "10.0.0.1:443",
		"[::1]:8443":       "[::1]:8443",
		"[::1]":            "[::1]:443",
		"::1":              "[::1]:443",
	}
	for in, want := range valid {
		got, err := parseTarget(in)
		if err != nil || got != want {
			t.Errorf("parseTarget(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	invalid := []string{
		"",
		"example.com/path",
		"example.com:443/path?x=1",
		"user:pass@example.com",
		"user@example.com:443",
		"example.com:0",
		"example.com:99999",
		"example.com:https",
		"exa mple.com",
		"example.com\"} forged 1\n",
		"example.com#frag",
		strings.Repeat("a", 300),
	}
	for _, in := range invalid {
		if got, err := parseTarget(in); err == nil {
			t.Errorf("parseTarget(%q) = %q, want error", in, got)
		}
	}
}

func TestFormatMetricsEscapesLabels(t *testing.T) {
	out := formatMetrics([]MetricsResponse{{
		MetricType: "gauge",
		MetricName: "ssl_expiration_delay",
		MetricHelp: "help",
		MetricsList: []Metric{{
			Labels: map[string]string{"target": "t", "domain": "evil\"} 1\nforged{x=\"y\\"},
			Value:  "42",
		}},
	}})
	want := "# HELP ssl_expiration_delay help\n" +
		"# TYPE ssl_expiration_delay gauge\n" +
		`ssl_expiration_delay{domain="evil\"} 1\nforged{x=\"y\\",target="t"} 42` + "\n"
	if out != want {
		t.Errorf("unexpected output:\n%s\nwant:\n%s", out, want)
	}
}

func newTestListener(bearer string) *HTTPListener {
	return NewHTTPListener(&config.Config{
		Listen: config.ListenConfig{Host: "127.0.0.1", Port: 9123, Bearer: bearer},
		Client: config.ClientConfig{Timeout: 2},
	})
}

func TestMetricsRequiresBearer(t *testing.T) {
	l := newTestListener("secret")

	req := httptest.NewRequest(http.MethodGet, "/metrics?target=example.com", nil)
	rec := httptest.NewRecorder()
	l.server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics?target=example.com", nil)
	req.Header.Set("Authorization", "Bearer nope")
	rec = httptest.NewRecorder()
	l.server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("bad token: status = %d, want 401", rec.Code)
	}
}

func TestMetricsRejectsBadTargets(t *testing.T) {
	l := newTestListener("")

	for _, q := range []string{"", "target=" + url.QueryEscape("example.com/admin?x=1")} {
		req := httptest.NewRequest(http.MethodGet, "/metrics?"+q, nil)
		rec := httptest.NewRecorder()
		l.server.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("query %q: status = %d, want 400", q, rec.Code)
		}
	}
}

func TestMetricsEndToEnd(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the exporter must not send an HTTP request to the target")
	}))
	defer srv.Close()
	target := strings.TrimPrefix(srv.URL, "https://")

	l := newTestListener("secret")
	req := httptest.NewRequest(http.MethodGet, "/metrics?target="+url.QueryEscape(target)+"&target="+url.QueryEscape(target), nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	l.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if n := strings.Count(body, `target="`+target+`"`); n != 1 {
		t.Errorf("expected exactly one sample for the target, got %d:\n%s", n, body)
	}
}
