package sysinfo

import (
	"fmt"
	"sort"
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
	fmt.Fprintf(&b, "  Severity : %s\n", snap.TriggerAnomaly.Severity)
	fmt.Fprintf(&b, "  Detected : %s\n\n", snap.TriggerAnomaly.DetectedAt.Format("15:04:05"))

	// Section 1.5 — which detectors fired
	fmt.Fprintf(&b, "DETECTION SIGNALS:\n")
	fmt.Fprintf(&b, "DETECTION SIGNALS:\n")
	for _, sig := range snap.TriggerAnomaly.Signals {
		switch sig {
		case "zscore":
			fmt.Fprintf(&b, "  - Z-Score spike detected\n")
		case "trend":
			fmt.Fprintf(&b, "  - Upward trend detected\n")
		case "rate_of_change":
			fmt.Fprintf(&b, "  - Rapid rate of change detected\n")
		}
	}
	fmt.Fprintf(&b, "\n")

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
	fmt.Fprintf(&b, "RUNNING SERVICES (Docker containers):\n")
	for _, process := range snap.SystemInfo.Processes {
		fmt.Fprintf(&b, "  - %s (Docker container, currently running)\n", process)
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "NOTE: These are Docker containers on a Linux host. Suggest specific Linux and Docker commands relevant to the anomalous metric, not generic container restarts.\n\n")
	// Section 4 — instruction
	fmt.Fprintf(&b, "IMPORTANT: Respond with ONLY a JSON object. No markdown, no backticks, no extra text.\n")
	fmt.Fprintf(&b, "The JSON must have exactly these fields with exactly these types:\n")
	fmt.Fprintf(&b, `{"severity":"string","diagnosis":"string","commands":["string","string"],"long_term_fix":"string","confidence":0.0}`)
	fmt.Fprintf(&b, "\ncommands must be a flat array of strings. long_term_fix must be a plain string.\n")

	return b.String()
}
