package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/codeezard/argus/internal/detector"
	"github.com/codeezard/argus/internal/llm"
	prom "github.com/codeezard/argus/internal/prometheus"
	"github.com/codeezard/argus/internal/store"
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

	// Initialize store
	dbPath := viper.GetString("store.path")
	if dbPath == "" {
		dbPath = "argus.db"
	}
	dbStore, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

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

			// Store the simulated anomaly event
			evt := &store.Event{
				Timestamp:  trigger.DetectedAt.Format(time.RFC3339),
				Metric:     trigger.Name,
				Value:      trigger.CurrentValue,
				ZScore:     trigger.ZScore,
				Severity:   suggestion.Severity,
				Diagnosis:  suggestion.Diagnosis,
				Commands:   suggestion.Commands,
				Fix:        suggestion.LongTermFix,
				Confidence: suggestion.Confidence,
			}
			if err := dbStore.Save(evt); err != nil {
				fmt.Printf("  💾 DB save error (simulated): %v\n", err)
			} else {
				fmt.Printf("  💾 Stored simulated anomaly (ID: %d)\n", evt.ID)
				// Retrieve with GetById
				savedEvt, err := dbStore.GetById(evt.ID)
				if err != nil {
					fmt.Printf("  ⚠️  Failed to retrieve simulated anomaly: %v\n", err)
				} else {
					fmt.Printf("  ✓  Verified: Retrieved simulated anomaly %d from DB (metric: %s)\n", savedEvt.ID, savedEvt.Metric)
				}
			}
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

					// Step 4 — Store anomaly
					evt := &store.Event{
						Timestamp:  trigger.DetectedAt.Format(time.RFC3339),
						Metric:     anomaly.MetricName,
						Value:      anomaly.Value,
						ZScore:     anomaly.ZScore,
						Severity:   suggestion.Severity,
						Diagnosis:  suggestion.Diagnosis,
						Commands:   suggestion.Commands,
						Fix:        suggestion.LongTermFix,
						Confidence: suggestion.Confidence,
					}
					if err := dbStore.Save(evt); err != nil {
						fmt.Printf("     💾 DB save error: %v\n", err)
					} else {
						fmt.Printf("     💾 Stored anomaly in DB with ID: %d\n", evt.ID)
						// Retrieve with GetById
						savedEvt, err := dbStore.GetById(evt.ID)
						if err != nil {
							fmt.Printf("     ⚠️  Failed to retrieve saved anomaly: %v\n", err)
						} else {
							fmt.Printf("     ✓  Verified: Retrieved saved anomaly %d from DB (metric: %s)\n", savedEvt.ID, savedEvt.Metric)
						}
					}
				}
			}
		}

	}

	return nil
}
