package sysinfo

import (
	"fmt"
	"strings"
	"sort"
)

func BuildPrompt(snap ContextSnapshot) string {
	var b strings.Builder

	// Section 1 — what triggered
	fmt.Fprintf(&b, "You are an expert SRE. An anomaly was detected.\n\n")
	fmt.Fprintf(&b, "TRIGGER:\n")
	fmt.Fprintf(&b, "  Metric   : %s\n", snap.TriggerAnomaly.Name)
	fmt.Fprintf(&b, "  Value    : %.4f\n", snap.TriggerAnomaly.CurrentValue)
	fmt.Fprintf(&b, "  Z-Score  : %.2f\n", snap.TriggerAnomaly.ZScore)
	fmt.Fprintf(&b, "  Severity : %s\n\n", snap.TriggerAnomaly.Severity)

	// Section 2 — correlated metrics
	fmt.Fprintf(&b, "CORRELATED ANOMALIES (sorted by detection time):\n")
	sort.Slice(snap.CorrelatedMetrics, func(i, j int) bool {
    	return snap.CorrelatedMetrics[i].DetectedAt.Before(snap.CorrelatedMetrics[j].DetectedAt)
	})
	for _, metric := range snap.CorrelatedMetrics {
	    fmt.Fprintf(&b, "  - %s | value: %.4f | z-score: %.2f | severity: %s | detected: %s\n",
	        metric.Name,
	        metric.CurrentValue,
    	    metric.ZScore,
        	metric.Severity,
        	metric.DetectedAt.Format("15:04:05"),
    	)
	}
	fmt.Fprintf(&b, "\n")

	// Section 3 — running processes
	fmt.Fprintf(&b, "RUNNING SERVICES:\n")
	for _, process := range snap.SystemInfo.Processes {
		fmt.Fprintf(&b, "Process %d: %s\n", process)
	}

	// Section 4 — instruction
	fmt.Fprintf(&b, "IMPORTANT: Respond with ONLY a JSON object. No markdown, no backticks, no extra text.\n")
	fmt.Fprintf(&b, "The JSON must have exactly these fields with exactly these types:\n")
	fmt.Fprintf(&b, `{"severity":"string","diagnosis":"string","commands":["string","string"],"long_term_fix":"string","confidence":0.0}`)
	fmt.Fprintf(&b, "\ncommands must be a flat array of strings. long_term_fix must be a plain string.\n")

	return b.String()
}