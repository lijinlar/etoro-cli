package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var ordersSymbol string

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List pending orders",
	Long: `Display all pending (unfilled) orders in your eToro account.

Columns displayed:
  - ID: Order identifier (use with 'cancel' command)
  - Symbol: Trading instrument symbol
  - Direction: BUY or SELL
  - Type: Order type (market, limit, stop)
  - Qty: Number of units
  - Limit Price: Target price for limit orders
  - SL: Stop loss price
  - TP: Take profit price
  - Created: Order creation timestamp

Filtering:
  Use --symbol to filter orders by instrument symbol.

Examples:
  etoro orders                    # List all pending orders
  etoro orders --symbol TSLA      # Filter by symbol
  etoro orders --json             # Output as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		orders, err := c.GetOrders()
		if err != nil {
			exitWithError(err, 1)
		}

		// Filter by symbol if specified
		if ordersSymbol != "" {
			symbol := strings.ToUpper(ordersSymbol)
			filtered := make([]client.Order, 0)
			for _, o := range orders {
				if strings.ToUpper(o.Symbol) == symbol {
					filtered = append(filtered, o)
				}
			}
			orders = filtered
		}

		printer := output.NewPrinter(jsonOutput)

		if jsonOutput {
			printer.PrintJSON(orders)
		} else {
			if len(orders) == 0 {
				fmt.Println("No pending orders")
				return
			}

			headers := []string{"ID", "Symbol", "Direction", "Type", "Qty", "Limit Price", "SL", "TP", "Created"}
			rows := make([][]string, len(orders))
			for i, o := range orders {
				limitPrice := "-"
				if o.LimitPrice > 0 {
					limitPrice = fmt.Sprintf("$%.2f", o.LimitPrice)
				}
				sl := "-"
				if o.StopLoss > 0 {
					sl = fmt.Sprintf("$%.2f", o.StopLoss)
				}
				tp := "-"
				if o.TakeProfit > 0 {
					tp = fmt.Sprintf("$%.2f", o.TakeProfit)
				}

				rows[i] = []string{
					fmt.Sprintf("%d", o.OrderID),
					o.Symbol,
					o.Direction,
					o.OrderType,
					fmt.Sprintf("%.4f", o.Quantity),
					limitPrice,
					sl,
					tp,
					o.CreatedAt.Format("2006-01-02 15:04"),
				}
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	ordersCmd.Flags().StringVar(&ordersSymbol, "symbol", "", "filter by symbol (e.g., TSLA)")
	rootCmd.AddCommand(ordersCmd)
}
