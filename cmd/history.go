package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var (
	historyFrom   string
	historyTo     string
	historySymbol string
	historyLimit  int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View closed trade history",
	Long: `Display history of closed trades with optional filters.

Filters:
  --from DATE     Start date (YYYY-MM-DD format)
  --to DATE       End date (YYYY-MM-DD format)
  --symbol SYMBOL Filter by instrument symbol
  --limit N       Limit number of results (default: 50)

Columns displayed:
  - ID: Position identifier
  - Symbol: Trading instrument
  - Direction: BUY or SELL
  - Qty: Number of units
  - Open Price: Entry price
  - Close Price: Exit price
  - P&L: Realized profit/loss
  - P&L%: Percentage return
  - Closed: Close timestamp

Examples:
  etoro history                              # Recent history
  etoro history --from 2024-01-01            # From date
  etoro history --symbol AAPL --limit 10     # Filter by symbol
  etoro history --json                       # JSON output`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		history, err := c.GetHistory(historyFrom, historyTo, historySymbol, historyLimit)
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(history)
		} else {
			if len(history) == 0 {
				fmt.Println("No trade history found")
				return
			}

			headers := []string{"ID", "Symbol", "Direction", "Qty", "Open Price", "Close Price", "P&L", "P&L%", "Closed"}
			rows := make([][]string, len(history))
			for i, h := range history {
				rows[i] = []string{
					fmt.Sprintf("%d", h.PositionID),
					h.Symbol,
					h.Direction,
					fmt.Sprintf("%.4f", h.Quantity),
					fmt.Sprintf("$%.2f", h.OpenPrice),
					fmt.Sprintf("$%.2f", h.ClosePrice),
					output.FormatPL(h.PL),
					output.FormatPercent(h.PLPercent),
					h.CloseDate.Format("2006-01-02 15:04"),
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	historyCmd.Flags().StringVar(&historyFrom, "from", "", "start date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&historyTo, "to", "", "end date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&historySymbol, "symbol", "", "filter by symbol")
	historyCmd.Flags().IntVar(&historyLimit, "limit", 50, "limit number of results")

	rootCmd.AddCommand(historyCmd)
}
