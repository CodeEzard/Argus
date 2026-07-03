package detector

import (
	"fmt"

	"github.com/codeezard/argus/internal/prometheus"
)

// Anomaly represents a detected problem
type Anomaly struct {
	MetricName string
	Value      float64
	Threshold  float64
	Message    string
	Severity   string // "warn" or "critical"
}

// Rule defines what to check and when to fire
type Rule struct {
	Name      string
	Query     string  // PromQL query to run
	Threshold float64 // fire if value exceeds this
	Severity  string
}

// DefaultRules are the built-in rules Argus ships with
var DefaultRules = []Rule{
	{
		Name:      "High CPU usage",
		Query:     `rate(process_cpu_seconds_total[1m])`,
		Threshold: 0.8,
		Severity:  "warn",
	},
	{
		Name:      "Prometheus scrape failures",
		Query:     `rate(prometheus_target_scrape_pool_exceeded_target_limit_total[5m])`,
		Threshold: 0,
		Severity:  "warn",
	},
	{
    	Name:      "High memory usage",
    	Query:     `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`,
    	Threshold: 80.0,
    	Severity:  "warn",
	},
}

// Detector runs rules against metric samples and returns anomalies
type Detector struct {
	rules []Rule
}

func New(rules []Rule) *Detector {
	return &Detector{rules: rules}
}

// Check takes samples for a single rule and returns anomalies if threshold exceeded
func (d *Detector) Check(rule Rule, samples []prometheus.MetricSample) []Anomaly {
	var anomalies []Anomaly

	for _, s := range samples {
		if s.Value > rule.Threshold {
			anomalies = append(anomalies, Anomaly{
				MetricName: rule.Name,
				Value:      s.Value,
				Threshold:  rule.Threshold,
				Severity:   rule.Severity,
				Message: fmt.Sprintf(
					"%s — current value %.4f exceeds threshold %.4f",
					rule.Name, s.Value, rule.Threshold,
				),
			})
		}
	}

	return anomalies
}