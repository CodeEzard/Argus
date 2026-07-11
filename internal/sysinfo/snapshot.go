package sysinfo

import "time"

type ContextSnapshot struct {
    Timestamp         time.Time
    TriggerAnomaly    AnomalousMetric
    CorrelatedMetrics []AnomalousMetric
    RunningServices   []string
    SystemInfo        SystemInfo
}

type AnomalousMetric struct {
    Name         string
    CurrentValue float64
    ZScore       float64
    Severity     string
    DetectedAt   time.Time
}

type SystemInfo struct {
    Hostname  string
    OS        string
	Uptime    time.Duration
    Processes []string 
}

type Suggestion struct {
    Severity    string   `json:"severity"`
    Diagnosis   string   `json:"diagnosis"`
    Commands    []string `json:"commands"`
    LongTermFix string   `json:"long_term_fix"`
    Confidence  float64  `json:"confidence"`
}