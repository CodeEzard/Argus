package cmd

import (
	"fmt"
	"os"

	"github.com/codeezard/argus/internal/detector"
	prom "github.com/codeezard/argus/internal/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a one-shot check of all metrics and print anomalies",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	host := viper.GetString("prometheus.host")
	client := prom.NewClient(host)

	fmt.Printf("\n🔍 Argus scanning %s...\n\n", host)

	d := detector.NewDetector(60)
	anomaliesFound := false

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
			anomaliesFound = true
		}
	}

	fmt.Println()
	if anomaliesFound {
		fmt.Println("🚨 Anomalies detected.")
		os.Exit(1)
	}
	fmt.Println("✅ All checks passed.")
	return nil
}
