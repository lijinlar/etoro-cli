// Package cmd implements all CLI commands for etoro-cli.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/config"
)

var (
	cfgFile   string
	jsonOutput bool
	dryRun    bool
	verbose   bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "etoro",
	Short: "eToro CLI - Trading automation for eToro",
	Long: `etoro-cli is a production-quality command-line interface for eToro trading automation.

Designed for both human traders and AI agents, it provides full trading capabilities
with built-in safety guardrails, JSON output support, and comprehensive help text.

Features:
  - View account info, positions, orders, and portfolio
  - Place buy/sell orders with safety guards
  - Close positions and cancel orders
  - Search instruments and track prices
  - Manage watchlists and view trade history
  - Risk analysis and monitoring

Safety:
  - Kill switch for emergency trading halt
  - Execution toggle for dry-run mode
  - Maximum trade amount limits
  - Symbol allowlist restrictions

Get started:
  1. Create config file: cp config.example.yaml etoro.yaml
  2. Add your API keys to etoro.yaml
  3. Run: etoro account

Documentation: https://github.com/lijinlar/etoro-cli`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for help commands
		if cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}

		// Load configuration
		if err := config.Load(cfgFile); err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(2)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: ./etoro.yaml or ~/.etoro/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON (recommended for agent/script use)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate actions without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show HTTP requests and responses")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("etoro-cli v1.0.0")
		},
	})
}

// exitWithError prints an error and exits with the specified code
func exitWithError(err error, code int) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(code)
}

// checkConfig validates configuration and exits if invalid
func checkConfig() {
	if err := config.Validate(); err != nil {
		exitWithError(err, 2)
	}
}

// checkKillSwitch checks if trading is halted
func checkKillSwitch() {
	if config.IsKillSwitchActive() {
		exitWithError(fmt.Errorf("Kill switch is active. Set kill_switch: false in config to enable trading"), 3)
	}
}

// checkSymbolAllowed validates symbol against allowlist
func checkSymbolAllowed(symbol string) {
	if !config.IsSymbolAllowed(symbol) {
		exitWithError(fmt.Errorf("Symbol %s is not in the allowed list. Update symbol_allowlist in config to allow trading this symbol", symbol), 3)
	}
}

// checkMaxTrade validates trade amount against limits
func checkMaxTrade(amount float64) {
	if err := config.CheckMaxTradeUSD(amount); err != nil {
		exitWithError(err, 3)
	}
}

// isConfirmed checks if action is confirmed via flag or env
func isConfirmed(confirmFlag bool) bool {
	return confirmFlag || config.GetConfirmFromEnv()
}

// isExecutionEnabled checks if trading execution is allowed
func isExecutionEnabled() bool {
	return config.IsExecutionEnabled() && !dryRun
}
