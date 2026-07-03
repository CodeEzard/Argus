package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/codeezard/argus/internal/detector"
	prom "github.com/codeezard/argus/internal/prometheus"
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

	d := detector.New(detector.DefaultRules)
	anomaliesFound := false

	for _, rule := range detector.DefaultRules {
		fmt.Printf("  Checking: %s\n", rule.Name)

		samples, err := client.Query(rule.Query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️  Failed to query '%s': %v\n", rule.Name, err)
			continue
		}

		if len(samples) == 0 {
			fmt.Printf("  –  No data returned\n")
			continue
		}

		anomalies := d.Check(rule, samples)

		if len(anomalies) == 0 {
			fmt.Printf("  ✓  OK (value: %.4f)\n", samples[0].Value)
		} else {
			for _, a := range anomalies {
				fmt.Printf("  ⚠️  ANOMALY: %s\n", a.Message)
				fmt.Printf("     Severity : %s\n", a.Severity)
				fmt.Printf("     Value    : %.4f\n", a.Value)
				fmt.Printf("     Threshold: %.4f\n", a.Threshold)
				anomaliesFound = true
			}
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