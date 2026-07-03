package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yourusername/argus/internal/detector"
	prom "github.com/yourusername/argus/internal/prometheus"
)

// color helpers — these make CLI output readable at a glance
var (
	okStyle       = color.New(color.FgGreen, color.Bold)
	warnStyle     = color.New(color.FgYellow, color.Bold)
	criticalStyle = color.New(color.FgRed, color.Bold)
	dimStyle      = color.New(color.Faint)
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a one-shot check of all metrics and print anomalies",
	Long: `Scan connects to Prometheus, evaluates all configured rules,
and prints any anomalies it finds. Exits with code 1 if anomalies are found.`,
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) {
	host := viper.GetString("prometheus.host")
	client := prom.NewClient(host)

	fmt.Printf("\n🔍 Argus scanning %s\n\n", dimStyle.Sprint(host))

	anomaliesFound := false

	for _, rule := range detector.DefaultRules {
		results, err := client.Query(rule.PromQL)
		if err != nil {
			// Don't crash the whole scan if one query fails
			warnStyle.Printf("  ⚠  Could not query '%s': %v\n", rule.Name, err)
			continue
		}

		if len(results) == 0 {
			dimStyle.Printf("  –  %s: no data\n", rule.Name)
			continue
		}

		// Check each time series returned by this query
		for _, result := range results {
			anomaly := detector.Check(rule, result.Value)

			if anomaly == nil {
				okStyle.Printf("  ✓  %s: %.4f (threshold: %.4f)\n",
					rule.Name, result.Value, rule.Threshold)
				continue
			}

			anomaliesFound = true

			switch anomaly.Severity {
			case detector.SeverityCritical:
				criticalStyle.Printf("  ✗  %s\n", anomaly)
			case detector.SeverityWarn:
				warnStyle.Printf("  ⚠  %s\n", anomaly)
			}
		}
	}

	fmt.Println()

	if anomaliesFound {
		criticalStyle.Println("Anomalies detected.")
		// Exit code 1 means "problems found" — useful for CI pipelines
		os.Exit(1)
	} else {
		okStyle.Println("All checks passed.")
	}
}
