package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search QUERY",
	Short: "Search for instruments",
	Long: `Search for trading instruments by name or ticker symbol.

Returns matching instruments with:
  - Symbol: Trading ticker
  - Name: Full instrument name
  - Type: Asset type (stock, crypto, etc.)
  - Exchange: Trading exchange
  - Trading: Whether trading is currently enabled

Examples:
  etoro search apple           # Search by name
  etoro search AAPL            # Search by ticker
  etoro search bitcoin         # Search crypto
  etoro search "S&P 500"       # Search indices
  etoro search apple --json    # JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		query := args[0]
		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		instruments, err := c.SearchInstruments(query)
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(instruments)
		} else {
			if len(instruments) == 0 {
				fmt.Printf("No instruments found for: %s\n", query)
				return
			}

			headers := []string{"Symbol", "Name", "Type", "Exchange", "Trading"}
			rows := make([][]string, len(instruments))
			for i, inst := range instruments {
				trading := "No"
				if inst.TradingEnabled {
					trading = "Yes"
				}
				rows[i] = []string{
					inst.Symbol,
					inst.Name,
					inst.Type,
					inst.Exchange,
					trading,
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
