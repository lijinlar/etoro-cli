package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var positionsSymbol string

var positionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "List open positions",
	Long: `Display all open trading positions in your eToro account.

Columns displayed:
  - ID: Position identifier (use with 'close' command)
  - Symbol: Trading instrument symbol
  - Direction: BUY or SELL
  - Qty: Number of units held
  - Open Price: Entry price
  - Current Price: Current market price
  - P&L: Unrealized profit/loss in USD
  - P&L%: Unrealized profit/loss percentage

Filtering:
  Use --symbol to filter positions by instrument symbol.

Examples:
  etoro positions                    # List all open positions
  etoro positions --symbol AAPL      # Filter by symbol
  etoro positions --json             # Output as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		positions, err := c.GetPositions()
		if err != nil {
			exitWithError(err, 1)
		}

		// Filter by symbol if specified
		if positionsSymbol != "" {
			symbol := strings.ToUpper(positionsSymbol)
			filtered := make([]client.Position, 0)
			for _, p := range positions {
				if strings.ToUpper(p.Symbol) == symbol {
					filtered = append(filtered, p)
				}
			}
			positions = filtered
		}

		printer := output.NewPrinter(jsonOutput)

		if jsonOutput {
			printer.PrintJSON(positions)
		} else {
			if len(positions) == 0 {
				fmt.Println("No open positions")
				return
			}

			headers := []string{"ID", "Symbol", "Direction", "Qty", "Open Price", "Current Price", "P&L", "P&L%"}
			rows := make([][]string, len(positions))
			for i, p := range positions {
				rows[i] = []string{
					fmt.Sprintf("%d", p.PositionID),
					p.Symbol,
					p.Direction,
					fmt.Sprintf("%.4f", p.Quantity),
					fmt.Sprintf("$%.2f", p.OpenPrice),
					fmt.Sprintf("$%.2f", p.CurrentPrice),
					output.FormatPL(p.PL),
					output.FormatPercent(p.PLPercent),
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	positionsCmd.Flags().StringVar(&positionsSymbol, "symbol", "", "filter by symbol (e.g., AAPL)")
	rootCmd.AddCommand(positionsCmd)
}
