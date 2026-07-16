package sysinfo

import (
    "fmt"
    "strings"
)

// Observe returns a human-readable summary of the snapshot
// This runs before the LLM and gives the operator instant context
func Observe(snap ContextSnapshot) string {
    var lines []string

    // 1. describe the trigger
    lines = append(lines, fmt.Sprintf(
        "%s is anomalous (value: %.2f, z-score: %.2f, severity: %s)",
        snap.TriggerAnomaly.Name,
        snap.TriggerAnomaly.CurrentValue,
        snap.TriggerAnomaly.ZScore,
        snap.TriggerAnomaly.Severity,
    ))

    // 2. describe which detectors fired
    if len(snap.TriggerAnomaly.Signals) > 0 {
        lines = append(lines, fmt.Sprintf(
            "Triggered by: %s",
            strings.Join(snap.TriggerAnomaly.Signals, ", "),
        ))
    }

    // 3. describe correlated metrics
    if len(snap.CorrelatedMetrics) > 0 {
        lines = append(lines, fmt.Sprintf(
            "%d correlated metric(s) also showing anomalous behaviour:",
            len(snap.CorrelatedMetrics),
        ))
        for _, m := range snap.CorrelatedMetrics {
            lines = append(lines, fmt.Sprintf(
                "  → %s (%s, detected %s ago)",
                m.Name,
                m.Severity,
                snap.Timestamp.Sub(m.DetectedAt).Round(1e9),
            ))
        }
    }

    // 4. describe running services
    if len(snap.SystemInfo.Processes) > 0 {
        lines = append(lines, fmt.Sprintf(
            "Running services: %s",
            strings.Join(snap.SystemInfo.Processes, ", "),
        ))
    }

    return strings.Join(lines, "\n")
}