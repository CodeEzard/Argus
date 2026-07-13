package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/codeezard/argus/internal/store"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show a history of detected anomalies from the database",
	RunE:  runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	dbPath := viper.GetString("store.path")
	if dbPath == "" {
		dbPath = "argus.db"
	}
	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database store: %w", err)
	}

	events, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list history: %w", err)
	}

	if len(events) == 0 {
		fmt.Println("No anomalies recorded in history yet.")
		return nil
	}

	fmt.Printf("\n📋 Showing %d recorded anomalies:\n\n", len(events))

	// Set up tabwriter to print tables nicely aligned
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	headerColor := color.New(color.Bold, color.FgCyan)
	fmt.Fprintln(w, headerColor.Sprint("ID\tTimestamp\tMetric\tValue\tZ-Score\tSeverity\tDiagnosis"))
	fmt.Fprintln(w, "──\t─────────\t──────\t─────\t───────\t────────\t─────────")

	for _, e := range events {
		// Format Timestamp
		tsStr := e.Timestamp
		if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
			tsStr = t.Format("2006-01-02 15:04:05")
		}

		// Color severity
		var sevCol string
		switch e.Severity {
		case "high":
			sevCol = color.RedString("HIGH")
		case "medium":
			sevCol = color.YellowString("MEDIUM")
		case "low":
			sevCol = color.BlueString("LOW")
		default:
			sevCol = e.Severity
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%.4f\t%.2f\t%s\t%s\n",
			e.ID,
			tsStr,
			e.Metric,
			e.Value,
			e.ZScore,
			sevCol,
			e.Diagnosis,
		)
	}
	w.Flush()
	fmt.Println()

	return nil
}
