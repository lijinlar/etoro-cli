package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var priceCmd = &cobra.Command{
	Use:   "price SYMBOL [SYMBOL2...]",
	Short: "Get live prices for symbols",
	Long: `Retrieve real-time price data for one or more trading instruments.

Data displayed:
  - Symbol: Instrument ticker
  - Bid: Current bid price
  - Ask: Current ask price
  - Spread: Difference between ask and bid
  - Daily Change%: Price change since market open
  - Daily High: Highest price today
  - Daily Low: Lowest price today

Multiple symbols can be specified to get prices in a single call.

Examples:
  etoro price AAPL                    # Single symbol
  etoro price AAPL TSLA GOOGL         # Multiple symbols
  etoro price AAPL --json             # JSON output for scripts`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		var rates []client.InstrumentRate

		for _, symbol := range args {
			// Find instrument by symbol
			instrument, err := c.GetInstrumentBySymbol(symbol)
			if err != nil {
				exitWithError(fmt.Errorf("symbol %s: %w", symbol, err), 1)
			}

			// Get rate for instrument
			rate, err := c.GetInstrumentRate(instrument.InstrumentID)
			if err != nil {
				exitWithError(fmt.Errorf("symbol %s: %w", symbol, err), 1)
			}

			rate.Symbol = instrument.Symbol
			rates = append(rates, *rate)
		}

		if jsonOutput {
			if len(rates) == 1 {
				printer.PrintJSON(rates[0])
			} else {
				printer.PrintJSON(rates)
			}
		} else {
			headers := []string{"Symbol", "Bid", "Ask", "Spread", "Daily Change%", "High", "Low"}
			rows := make([][]string, len(rates))
			for i, r := range rates {
				rows[i] = []string{
					r.Symbol,
					fmt.Sprintf("$%.2f", r.Bid),
					fmt.Sprintf("$%.2f", r.Ask),
					fmt.Sprintf("$%.4f", r.Spread),
					output.FormatPercent(r.DailyChange),
					fmt.Sprintf("$%.2f", r.DailyHigh),
					fmt.Sprintf("$%.2f", r.DailyLow),
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)
}
