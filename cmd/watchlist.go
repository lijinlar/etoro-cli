package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var (
	watchlistAdd    string
	watchlistRemove string
)

var watchlistCmd = &cobra.Command{
	Use:   "watchlist",
	Short: "View or manage watchlist",
	Long: `View and manage your eToro watchlist.

Actions:
  - Without flags: Display current watchlist
  - --add SYMBOL: Add instrument to watchlist
  - --remove SYMBOL: Remove instrument from watchlist

Examples:
  etoro watchlist                    # View watchlist
  etoro watchlist --add AAPL         # Add AAPL
  etoro watchlist --remove TSLA      # Remove TSLA
  etoro watchlist --json             # JSON output`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		printer := output.NewPrinter(jsonOutput)

		// Handle add
		if watchlistAdd != "" {
			instrument, err := c.GetInstrumentBySymbol(watchlistAdd)
			if err != nil {
				exitWithError(err, 1)
			}

			if err := c.AddToWatchlist(instrument.InstrumentID); err != nil {
				exitWithError(err, 1)
			}

			if jsonOutput {
				printer.PrintJSON(map[string]interface{}{
					"action":  "add",
					"symbol":  watchlistAdd,
					"success": true,
				})
			} else {
				fmt.Printf("Added %s to watchlist\n", watchlistAdd)
			}
			return
		}

		// Handle remove
		if watchlistRemove != "" {
			instrument, err := c.GetInstrumentBySymbol(watchlistRemove)
			if err != nil {
				exitWithError(err, 1)
			}

			if err := c.RemoveFromWatchlist(instrument.InstrumentID); err != nil {
				exitWithError(err, 1)
			}

			if jsonOutput {
				printer.PrintJSON(map[string]interface{}{
					"action":  "remove",
					"symbol":  watchlistRemove,
					"success": true,
				})
			} else {
				fmt.Printf("Removed %s from watchlist\n", watchlistRemove)
			}
			return
		}

		// View watchlist
		watchlist, err := c.GetWatchlist()
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(watchlist)
		} else {
			if len(watchlist.Items) == 0 {
				fmt.Println("Watchlist is empty")
				return
			}

			headers := []string{"Symbol", "Name", "Added"}
			rows := make([][]string, len(watchlist.Items))
			for i, item := range watchlist.Items {
				rows[i] = []string{
					item.Symbol,
					item.Name,
					item.AddedAt,
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	watchlistCmd.Flags().StringVar(&watchlistAdd, "add", "", "add symbol to watchlist")
	watchlistCmd.Flags().StringVar(&watchlistRemove, "remove", "", "remove symbol from watchlist")

	rootCmd.AddCommand(watchlistCmd)
}
