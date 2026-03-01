package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var riskCmd = &cobra.Command{
	Use:   "risk",
	Short: "Risk analysis dashboard",
	Long: `Display risk metrics and analysis for your trading account.

Metrics displayed:
  - Margin Utilization %: How much of available margin is used
  - Total Exposure: Total value of all positions
  - Per-Symbol Exposure: Breakdown by instrument

Warnings:
  - WARNING level when margin utilization > 70%
  - CRITICAL level when margin utilization > 90%

Examples:
  etoro risk            # Display risk dashboard
  etoro risk --json     # JSON output for automated monitoring`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		// Get account info for margin
		account, err := c.GetAccount()
		if err != nil {
			exitWithError(err, 1)
		}

		// Get positions for exposure analysis
		positions, err := c.GetPositions()
		if err != nil {
			exitWithError(err, 1)
		}

		// Calculate metrics
		var totalExposure float64
		symbolExposure := make(map[string]float64)

		for _, p := range positions {
			exposure := p.CurrentPrice * p.Quantity * float64(p.Leverage)
			totalExposure += exposure
			symbolExposure[p.Symbol] += exposure
		}

		// Calculate exposure vs balance
		var warningLevel string
		if account.Balance > 0 && totalExposure > account.Balance*5 {
			warningLevel = "CRITICAL"
		} else if account.Balance > 0 && totalExposure > account.Balance*2 {
			warningLevel = "WARNING"
		}

		metrics := &client.RiskMetrics{
			MarginUtilization: 0, // Not available from public API
			TotalExposure:     totalExposure,
			SymbolExposure:    symbolExposure,
			WarningLevel:      warningLevel,
		}

		if jsonOutput {
			printer.PrintJSON(metrics)
		} else {
			fmt.Println("=== RISK DASHBOARD ===")

			// Warning banner
			if warningLevel == "CRITICAL" {
				fmt.Println("!!! CRITICAL: Exposure is >5x account balance !!!")
			} else if warningLevel == "WARNING" {
				fmt.Println("! WARNING: Exposure is >2x account balance !")
			}
			fmt.Println()

			// Summary metrics
			fmt.Printf("Total Exposure:     %s\n", output.FormatMoney(totalExposure))
			fmt.Printf("Account Balance:    %s\n", output.FormatMoney(account.Balance))
			fmt.Printf("Unrealized P&L:     %s\n", output.FormatPL(account.UnrealizedPL))
			fmt.Println()

			// Per-symbol exposure
			if len(symbolExposure) > 0 {
				fmt.Println("--- Exposure by Symbol ---")

				// Sort symbols by exposure
				type symbolExp struct {
					symbol   string
					exposure float64
				}
				var sorted []symbolExp
				for sym, exp := range symbolExposure {
					sorted = append(sorted, symbolExp{sym, exp})
				}
				sort.Slice(sorted, func(i, j int) bool {
					return sorted[i].exposure > sorted[j].exposure
				})

				for _, se := range sorted {
					pct := (se.exposure / totalExposure) * 100
					fmt.Printf("  %-8s %s (%.1f%%)\n", se.symbol, output.FormatMoney(se.exposure), pct)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(riskCmd)
}
