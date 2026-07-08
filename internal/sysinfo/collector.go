package sysinfo

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

)
func Collect(trigger AnomalousMetric, correlated []AnomalousMetric) ContextSnapshot {
    // build and return a ContextSnapshot
	hostname, _ := os.Hostname()
	return ContextSnapshot{
    	SystemInfo: SystemInfo{
    		OS:        runtime.GOOS,
    		Hostname:  hostname,
    		Processes: getProcesses(),
		},
		Timestamp:         time.Now(),
		TriggerAnomaly:    trigger,
		CorrelatedMetrics: correlated,
	}
}


func getProcesses() []string {
    out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
    if err != nil {
        return []string{}
    }
    return strings.Split(strings.TrimSpace(string(out)), "\n")
}