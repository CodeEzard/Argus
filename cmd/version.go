package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the Argus version",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("argus v0.1.0")
        return nil
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}