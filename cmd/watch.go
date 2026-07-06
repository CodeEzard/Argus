package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/codeezard/argus/internal/detector"
	prom "github.com/codeezard/argus/internal/prometheus"
)

var watchCmd = &cobra.Command{
    Use:   "watch",
    Short: "See realtime Metrics for each cycle",
    RunE: runWatch,
}

func init() {
    rootCmd.AddCommand(watchCmd)
    watchCmd.Flags().Int("interval", 30, "Polling interval in seconds")
}

func runWatch(cmd *cobra.Command, args []string) error {
    host := viper.GetString("prometheus.host")
    client := prom.NewClient(host)
    interval, _ := cmd.Flags().GetInt("interval")

    fmt.Printf("\n👁  Argus watching %s (every %ds)...\n\n", host, interval)

    d := detector.NewDetector(60)  // lives outside loop — history accumulates
    ticker := time.NewTicker(time.Duration(interval) * time.Second)

    for range ticker.C {
        fmt.Printf("── %s ──\n", time.Now().Format("15:04:05"))

        // your query loop here — same as scan.go
        // but use d.Observe and d.Check
		for _, query := range queries {
    		fmt.Printf("  Checking: %s\n", query)

    		samples, err := client.Query(query)
    		if err != nil {
    		    fmt.Fprintf(os.Stderr, "  ⚠️  Failed to query: %v\n", err)
    		    continue
    		}

    		if len(samples) == 0 {
    		    fmt.Printf("  –  No data returned\n")
    		    continue
    		}

    		value := samples[0].Value

    		d.Observe(query, value)
    		anomaly := d.Check(query, value)

    		if anomaly == nil {
    		    fmt.Printf("  ✓  OK (value: %.4f)\n", value)
    		} else {
    		    fmt.Printf("  ⚠️  ANOMALY detected\n")
    		    fmt.Printf("     Metric  : %s\n", anomaly.MetricName)
    		    fmt.Printf("     Value   : %.4f\n", anomaly.Value)
    		    fmt.Printf("     Z-Score : %.4f\n", anomaly.ZScore)
    		    fmt.Printf("     Severity: %s\n", anomaly.Severity)
    		}
	}

    }

    return nil
}