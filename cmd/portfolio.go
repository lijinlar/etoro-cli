package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Full portfolio snapshot with P&L summary",
	Long: `Display a comprehensive portfolio overview including:
  - Total number of open positions
  - Total portfolio value
  - Total unrealized P&L
  - Top 3 gaining positions
  - Top 3 losing positions

This command aggregates data from all open positions to provide
a quick snapshot of your portfolio performance.

Examples:
  etoro portfolio            # Display portfolio summary
  etoro portfolio --json     # JSON output for scripts/agents`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		// Get all positions
		positions, err := c.GetPositions()
		if err != nil {
			exitWithError(err, 1)
		}

		// Calculate totals
		var totalValue, totalPL float64
		for _, p := range positions {
			totalValue += p.CurrentPrice * p.Quantity
			totalPL += p.PL
		}

		// Sort for top gainers/losers
		sorted := make([]client.Position, len(positions))
		copy(sorted, positions)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].PL > sorted[j].PL
		})

		// Get top 3 gainers and losers
		var topGainers, topLosers []client.Position
		for i, p := range sorted {
			if i < 3 && p.PL > 0 {
				topGainers = append(topGainers, p)
			}
		}
		for i := len(sorted) - 1; i >= 0 && len(topLosers) < 3; i-- {
			if sorted[i].PL < 0 {
				topLosers = append(topLosers, sorted[i])
			}
		}

		summary := &client.PortfolioSummary{
			TotalPositions: len(positions),
			TotalValue:     totalValue,
			UnrealizedPL:   totalPL,
			TopGainers:     topGainers,
			TopLosers:      topLosers,
		}

		if jsonOutput {
			printer.PrintJSON(summary)
		} else {
			fmt.Println("=== PORTFOLIO SUMMARY ===")
			fmt.Printf("Total Positions:  %d\n", summary.TotalPositions)
			fmt.Printf("Total Value:      %s\n", output.FormatMoney(summary.TotalValue))
			fmt.Printf("Unrealized P&L:   %s\n", output.FormatPL(summary.UnrealizedPL))
			fmt.Println()

			if len(topGainers) > 0 {
				fmt.Println("--- Top Gainers ---")
				for _, p := range topGainers {
					fmt.Printf("  %s: %s (%s)\n", p.Symbol, output.FormatPL(p.PL), output.FormatPercent(p.PLPercent))
				}
				fmt.Println()
			}

			if len(topLosers) > 0 {
				fmt.Println("--- Top Losers ---")
				for _, p := range topLosers {
					fmt.Printf("  %s: %s (%s)\n", p.Symbol, output.FormatPL(p.PL), output.FormatPercent(p.PLPercent))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(portfolioCmd)
}
