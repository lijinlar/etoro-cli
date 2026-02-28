package cmd

import (
	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Show account summary",
	Long: `Display your eToro account information including:
  - Login ID
  - Balance (available cash)
  - Equity (total account value)
  - Margin (used margin)
  - Available margin
  - Unrealized P&L (from open positions)
  - Realized P&L today

Examples:
  etoro account              # Display as table
  etoro account --json       # Output as JSON (for scripts/agents)`,
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		c := client.New(verbose)
		account, err := c.GetAccount()
		if err != nil {
			exitWithError(err, 1)
		}

		printer := output.NewPrinter(jsonOutput)

		if jsonOutput {
			printer.PrintJSON(account)
		} else {
			headers := []string{"Field", "Value"}
			rows := [][]string{
				{"Login ID", account.LoginID},
				{"Balance", output.FormatMoney(account.Balance)},
				{"Equity", output.FormatMoney(account.Equity)},
				{"Margin", output.FormatMoney(account.Margin)},
				{"Available Margin", output.FormatMoney(account.AvailableMargin)},
				{"Unrealized P&L", output.FormatPL(account.UnrealizedPL)},
				{"Realized P&L Today", output.FormatPL(account.RealizedPLToday)},
			}
			printer.PrintTable(headers, rows)
		}
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
