package cmd

import (
	"fmt"
	"time"
	"strconv"
	"github.com/codeezard/argus/internal/store"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)
 

var inspectCmd = &cobra.Command{
    Use:   "inspect <id>",
    Short: "Show full details of a specific anomaly",
    Args:  cobra.ExactArgs(1), // enforces exactly one argument
    RunE:  runInspect,
}

func init(){
	rootCmd.AddCommand(inspectCmd)
}

func runInspect(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %s", args[0])
	}

	dbPath := viper.GetString("store.path")
	if dbPath == "" {
		dbPath = "argus.db"
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	e, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("event %d not found: %w", id, err)
	}

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

	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(cyan(fmt.Sprintf("🔍 Anomaly Detail [ID: %d]", e.ID)))
	fmt.Println(color.WhiteString("────────────────────────────────────────────────────────────────────────────────"))
	fmt.Printf("%-15s : %s\n", bold("Timestamp"), tsStr)
	fmt.Printf("%-15s : %s\n", bold("Metric"), e.Metric)
	fmt.Printf("%-15s : %.4f\n", bold("Value"), e.Value)
	fmt.Printf("%-15s : %.2f\n", bold("Z-Score"), e.ZScore)
	fmt.Printf("%-15s : %s\n", bold("Severity"), sevCol)
	fmt.Printf("%-15s : %s\n", bold("Diagnosis"), e.Diagnosis)

	fmt.Printf("%-15s :\n", bold("Commands"))
	if len(e.Commands) == 0 {
		fmt.Println("  None suggested")
	} else {
		for _, cmdStr := range e.Commands {
			fmt.Printf("  - %s\n", color.YellowString(cmdStr))
		}
	}

	fmt.Printf("%-15s : %s\n", bold("Long term fix"), e.Fix)

	// Color confidence based on value
	confPercent := e.Confidence * 100
	var confStr string
	if confPercent >= 80 {
		confStr = color.GreenString("%.0f%%", confPercent)
	} else if confPercent >= 50 {
		confStr = color.YellowString("%.0f%%", confPercent)
	} else {
		confStr = color.RedString("%.0f%%", confPercent)
	}
	fmt.Printf("%-15s : %s\n", bold("Confidence"), confStr)
	fmt.Println(color.WhiteString("────────────────────────────────────────────────────────────────────────────────"))
	fmt.Println()

	return nil
}