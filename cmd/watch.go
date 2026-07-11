package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/codeezard/argus/internal/detector"
	"github.com/codeezard/argus/internal/llm"
	prom "github.com/codeezard/argus/internal/prometheus"
	"github.com/codeezard/argus/internal/sysinfo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "See realtime Metrics for each cycle",
	RunE:  runWatch,
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

	d := detector.NewDetector(60) // lives outside loop — history accumulates
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for range ticker.C {
		fmt.Printf("── %s ──\n", time.Now().Format("15:04:05"))
		fmt.Printf("── %s ──\n", time.Now().Format("15:04:05"))

		// TEMPORARY: simulate an anomaly to test LLM pipeline
		// delete this block after testing
		trigger := sysinfo.AnomalousMetric{
			Name:         "rate(process_cpu_seconds_total[1m])",
			CurrentValue: 0.95,
			ZScore:       3.8,
			Severity:     "high",
			DetectedAt:   time.Now(),
		}
		snap := sysinfo.Collect(trigger, []sysinfo.AnomalousMetric{})
		provider := &llm.OllamaProvider{Model: "llama3.2:3b"}
		llmClient := llm.NewClient(provider)
		suggestion, err := llmClient.Diagnose(snap)
		if err != nil {
			fmt.Printf("  LLM error: %v\n", err)
		} else {
			fmt.Printf("  🚨 Diagnosis  : %s\n", suggestion.Diagnosis)
			fmt.Printf("  💊 Commands   :\n")
			for _, cmd := range suggestion.Commands {
				fmt.Printf("       $ %s\n", cmd)
			}
			fmt.Printf("  🔧 Long term  : %s\n", suggestion.LongTermFix)
			fmt.Printf("  📊 Confidence : %.0f%%\n", suggestion.Confidence*100)
		}

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
				// Step 1 — build context snapshot
				trigger := sysinfo.AnomalousMetric{
					Name:         anomaly.MetricName,
					CurrentValue: anomaly.Value,
					ZScore:       anomaly.ZScore,
					Severity:     anomaly.Severity,
					DetectedAt:   time.Now(),
				}
				snap := sysinfo.Collect(trigger, []sysinfo.AnomalousMetric{})

				// Step 2 — call LLM
				provider := &llm.OllamaProvider{Model: "llama3.2:3b"}
				llmClient := llm.NewClient(provider)
				suggestion, err := llmClient.Diagnose(snap)
				if err != nil {
					fmt.Printf("  ⚠️  ANOMALY detected (LLM unavailable: %v)\n", err)
					fmt.Printf("     Metric  : %s\n", anomaly.MetricName)
					fmt.Printf("     Value   : %.4f\n", anomaly.Value)
					fmt.Printf("     Z-Score : %.4f\n", anomaly.ZScore)
				} else {
					// Step 3 — print suggestion
					fmt.Printf("  🚨 ANOMALY: %s\n", anomaly.MetricName)
					fmt.Printf("     Severity   : %s\n", suggestion.Severity)
					fmt.Printf("     Diagnosis  : %s\n", suggestion.Diagnosis)
					fmt.Printf("     Commands   :\n")
					for _, cmd := range suggestion.Commands {
						fmt.Printf("       $ %s\n", cmd)
					}
					fmt.Printf("     Long term  : %s\n", suggestion.LongTermFix)
					fmt.Printf("     Confidence : %.0f%%\n", suggestion.Confidence*100)
				}
			}
		}

	}

	return nil
}
