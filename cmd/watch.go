package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/codeezard/argus/internal/detector"
	"github.com/codeezard/argus/internal/llm"
	prom "github.com/codeezard/argus/internal/prometheus"
	"github.com/codeezard/argus/internal/store"
	"github.com/codeezard/argus/internal/sysinfo"
	"github.com/fatih/color"
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
		// single timestamp line
		ts := time.Now().Format("15:04:05")
		separator := strings.Repeat("─", 35)
		fmt.Printf("\n%s %s %s\n\n", separator, ts, separator)

		for _, q := range queries {
			samples, err := client.Query(q.PromQL)
			if err != nil {
				fmt.Printf("  ⚠️  %-20s  error: %v\n", q.Name, err)
				continue
			}

			if len(samples) == 0 {
				fmt.Printf("  –  %-20s  no data\n", q.Name)
				continue
			}

			value := samples[0].Value
			d.Observe(q.PromQL, value)
			anomaly := d.CheckMulti(q.PromQL, value)

			if anomaly == nil {
				fmt.Printf("  %s  %-20s  %.2f%%\n",
					color.GreenString("✓"),
					q.Name,
					value,
				)
				continue
			}

			// anomaly line
			fmt.Printf("  %s  %-20s  %.2f%%    %s\n",
    			color.RedString("✗"),
    			color.RedString(q.Name),
   				value,
    			color.YellowString("[%s — z-score: %.2f]", strings.Join(signalKinds(anomaly.Signals), " + "), getZScore(anomaly.Signals)),
			)

			// call LLM
			triggerTime := time.Now()
			correlated := d.FindCorrelated(q.PromQL, triggerTime)

			// convert to sysinfo.AnomalousMetric slice
			var correlatedMetrics []sysinfo.AnomalousMetric
			for _, pair := range correlated {
			    friendlyName := pair.MetricB
			    for _, mq := range queries {
			        if mq.PromQL == pair.MetricB {
			            friendlyName = mq.Name
			            break
			        }
			    }
			    correlatedMetrics = append(correlatedMetrics, sysinfo.AnomalousMetric{
			        Name:       friendlyName,
			        Severity:   pair.Relation,
			        DetectedAt: triggerTime.Add(-pair.TimeDelta),
			    })
			}
			trigger := sysinfo.AnomalousMetric{
    			Name:         q.Name,
			    CurrentValue: anomaly.Value,
			    ZScore:       getZScore(anomaly.Signals),
			    Severity:     anomaly.Severity,
			    DetectedAt:   triggerTime,
				Signals:      signalKinds(anomaly.Signals),
			}

			snap := sysinfo.Collect(trigger, correlatedMetrics)
			provider := &llm.OllamaProvider{Model: "llama3.2:3b"}
			llmClient := llm.NewClient(provider)
			suggestion, err := llmClient.Diagnose(snap)

			if err != nil {
				fmt.Printf("\n  %s Could not get LLM diagnosis: %v\n\n", color.YellowString("⚠️ "), err)
			} else {
				// print observation
				observation := sysinfo.Observe(snap)
				fmt.Printf("\n  📋 %s\n", strings.ReplaceAll(observation, "\n", "\n  "))
				printSuggestion(q.Name, suggestion)
				dbStore.Save(&store.Event{
					Timestamp:  time.Now().Format(time.RFC3339),
					Metric:     q.Name,
					Value:      value,
					ZScore:     getZScore(anomaly.Signals),
					Severity:   suggestion.Severity,
					Diagnosis:  suggestion.Diagnosis,
					Commands:   suggestion.Commands,
					Fix:        suggestion.LongTermFix,
					Confidence: suggestion.Confidence,
				})
			}
		}
	}

	return nil
}

func printSuggestion(metricName string, s sysinfo.Suggestion) {
    width := 74

    sevColor := color.YellowString
    if s.Severity == "high" || s.Severity == "critical" {
        sevColor = color.RedString
    } else if s.Severity == "low" {
        sevColor = color.CyanString
    }

    fmt.Printf("\n  ┌─ %s ─\n", color.RedString("🔴 ANOMALY: %s", metricName))
    fmt.Printf("  │  %-12s : %s\n", "Severity", sevColor(strings.ToUpper(s.Severity)))
    fmt.Printf("  │  %-12s : %s\n", "Diagnosis", s.Diagnosis)

    if len(s.Commands) > 0 {
        fmt.Printf("  │  %-12s :\n", "Commands")
        for _, cmd := range s.Commands {
            fmt.Printf("  │    %s %s\n", color.YellowString("$"), cmd)
        }
    }

    if s.LongTermFix != "" {
        fmt.Printf("  │  %-12s : %s\n", "Long term", s.LongTermFix)
    }

    confStr := color.GreenString("%.0f%%", s.Confidence*100)
    if s.Confidence < 0.5 {
        confStr = color.RedString("%.0f%%", s.Confidence*100)
    } else if s.Confidence < 0.8 {
        confStr = color.YellowString("%.0f%%", s.Confidence*100)
    }
    fmt.Printf("  │  %-12s : %s\n", "Confidence", confStr)
    fmt.Printf("  └%s\n\n", strings.Repeat("─", width))
}

func getZScore(signals []detector.Signal) float64 {
    for _, s := range signals {
        if s.Kind == "zscore" {
            return s.Value
        }
    }
    return 0
}

func signalKinds(signals []detector.Signal) []string {
    kinds := make([]string, len(signals))
    for i, s := range signals {
        kinds[i] = s.Kind
    }
    return kinds
}

