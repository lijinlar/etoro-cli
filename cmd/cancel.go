package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/config"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var cancelConfirm bool

var cancelCmd = &cobra.Command{
	Use:   "cancel ORDER_ID",
	Short: "Cancel a pending order",
	Long: `Cancel an existing pending (unfilled) order.

To find the ORDER_ID, use 'etoro orders' command.

Safety:
  Requires --confirm flag or ETORO_CONFIRM=1 environment variable.

Examples:
  etoro cancel 67890 --confirm    # Cancel order
  etoro cancel 67890 --json       # Dry-run with JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		orderID, err := strconv.Atoi(args[0])
		if err != nil {
			exitWithError(fmt.Errorf("invalid order ID: %s", args[0]), 3)
		}

		printer := output.NewPrinter(jsonOutput)

		// Safety checks
		checkKillSwitch()

		// Check if we should execute
		shouldExecute := isConfirmed(cancelConfirm) && isExecutionEnabled()

		if !shouldExecute {
			// Dry-run mode
			dryRunResult := &client.DryRunResult{
				Action:       "CANCEL",
				WouldExecute: false,
				Message:      "DRY-RUN: Order not cancelled. Use --confirm to execute.",
			}

			if !isConfirmed(cancelConfirm) {
				dryRunResult.Message = "DRY-RUN: Add --confirm flag to cancel this order."
			} else if !config.IsExecutionEnabled() {
				dryRunResult.Message = "DRY-RUN: Set execution_enabled: true in config to allow trading."
			} else if dryRun {
				dryRunResult.Message = "DRY-RUN: Remove --dry-run flag to cancel this order."
			}

			if jsonOutput {
				printer.PrintJSON(map[string]interface{}{
					"action":       "CANCEL",
					"orderId":      orderID,
					"wouldExecute": false,
					"message":      dryRunResult.Message,
				})
			} else {
				fmt.Println("=== DRY-RUN CANCEL PREVIEW ===")
				fmt.Printf("Action:    CANCEL\n")
				fmt.Printf("Order ID:  %d\n", orderID)
				fmt.Println("==============================")
				fmt.Println(dryRunResult.Message)
			}
			return
		}

		c := client.New(verbose)

		// Execute cancel
		resp, err := c.CancelOrder(orderID)
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(resp)
		} else {
			fmt.Println("=== ORDER CANCELLED ===")
			fmt.Printf("Order ID: %d\n", resp.OrderID)
			fmt.Printf("Status:   %s\n", resp.Status)
			if resp.Message != "" {
				fmt.Printf("Message:  %s\n", resp.Message)
			}
		}
	},
}

func init() {
	cancelCmd.Flags().BoolVar(&cancelConfirm, "confirm", false, "confirm and execute the cancellation")

	rootCmd.AddCommand(cancelCmd)
}
