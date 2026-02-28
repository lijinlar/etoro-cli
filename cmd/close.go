package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/config"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var (
	closePartial float64
	closeConfirm bool
)

var closeCmd = &cobra.Command{
	Use:   "close POSITION_ID",
	Short: "Close an open position",
	Long: `Close an existing open trading position.

To find the POSITION_ID, use 'etoro positions' command.

Partial closing:
  Use --partial to close only a portion of the position.
  The remaining units will stay open.

Safety:
  Requires --confirm flag or ETORO_CONFIRM=1 environment variable.

Examples:
  etoro close 12345 --confirm                # Close entire position
  etoro close 12345 --partial 5 --confirm    # Close 5 units only
  etoro close 12345 --json                   # Dry-run with JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		positionID, err := strconv.Atoi(args[0])
		if err != nil {
			exitWithError(fmt.Errorf("invalid position ID: %s", args[0]), 3)
		}

		printer := output.NewPrinter(jsonOutput)

		// Safety checks
		checkKillSwitch()

		// Check if we should execute
		shouldExecute := isConfirmed(closeConfirm) && isExecutionEnabled()

		if !shouldExecute {
			// Dry-run mode
			action := "CLOSE"
			if closePartial > 0 {
				action = fmt.Sprintf("PARTIAL_CLOSE (%.4f units)", closePartial)
			}

			dryRunResult := &client.DryRunResult{
				Action:       action,
				WouldExecute: false,
				Message:      "DRY-RUN: Position not closed. Use --confirm to execute.",
			}

			if !isConfirmed(closeConfirm) {
				dryRunResult.Message = "DRY-RUN: Add --confirm flag to close this position."
			} else if !config.IsExecutionEnabled() {
				dryRunResult.Message = "DRY-RUN: Set execution_enabled: true in config to allow trading."
			} else if dryRun {
				dryRunResult.Message = "DRY-RUN: Remove --dry-run flag to close this position."
			}

			if jsonOutput {
				printer.PrintJSON(map[string]interface{}{
					"action":       action,
					"positionId":   positionID,
					"partialQty":   closePartial,
					"wouldExecute": false,
					"message":      dryRunResult.Message,
				})
			} else {
				fmt.Println("=== DRY-RUN CLOSE PREVIEW ===")
				fmt.Printf("Action:      %s\n", action)
				fmt.Printf("Position ID: %d\n", positionID)
				if closePartial > 0 {
					fmt.Printf("Partial Qty: %.4f\n", closePartial)
				}
				fmt.Println("=============================")
				fmt.Println(dryRunResult.Message)
			}
			return
		}

		c := client.New(verbose)

		// Execute close
		resp, err := c.ClosePosition(positionID, closePartial)
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(resp)
		} else {
			fmt.Println("=== POSITION CLOSED ===")
			fmt.Printf("Position ID: %d\n", resp.PositionID)
			fmt.Printf("Realized P&L: %s\n", output.FormatPL(resp.ClosedPL))
			fmt.Printf("Status:      %s\n", resp.Status)
			if resp.Message != "" {
				fmt.Printf("Message:     %s\n", resp.Message)
			}
		}
	},
}

func init() {
	closeCmd.Flags().Float64Var(&closePartial, "partial", 0, "partial quantity to close (optional)")
	closeCmd.Flags().BoolVar(&closeConfirm, "confirm", false, "confirm and execute the close")

	rootCmd.AddCommand(closeCmd)
}
