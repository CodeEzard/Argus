package sysinfo

import (
	"fmt"
	"strings"
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
	for i, metric := range snap.CorrelatedMetrics {
		fmt.Fprintf(&b, "Metric %d: %s %.4f %.2f %s\n",
			i,
			metric.Name,
			metric.CurrentValue,
			metric.ZScore,
			metric.Severity,
		)
	}

	// Section 3 — running processes
	for i, process := range snap.SystemInfo.Processes {
		fmt.Fprintf(&b, "Process %d: %s\n", i, process)
	}

	// Section 4 — instruction
	fmt.Fprintf(&b, "Respond ONLY in this JSON format:\n")
	fmt.Fprintf(&b, `{"severity":"...","diagnosis":"...","commands":["..."],"long_term_fix":"...","confidence":0.0}`)

	return b.String()
}