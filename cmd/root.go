package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "argus",
	Short: "Argus — AI-powered infrastructure monitor",
	Long: `Argus watches your infrastructure, detects anomalies,
and suggests fixes before they become incidents.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags available on every subcommand
	rootCmd.PersistentFlags().String("host", "http://localhost:9090", "Prometheus host URL")
	viper.BindPFlag("prometheus.host", rootCmd.PersistentFlags().Lookup("host"))
}

func initConfig() {
	viper.SetConfigName("argus")   // looks for argus.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")       // current directory first
	viper.AddConfigPath("$HOME/.argus") // then home dir

	// Read config file if it exists — don't error if it doesn't
	// You'll add `argus init` in Phase 3 to generate this
	viper.ReadInConfig()

	// Env vars override config file — prefix with ARGUS_
	// e.g. ARGUS_PROMETHEUS_HOST=http://myserver:9090
	viper.SetEnvPrefix("ARGUS")
	viper.AutomaticEnv()
}
