package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/config"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var (
	sellAmount     float64
	sellQty        float64
	sellType       string
	sellLimitPrice float64
	sellSL         float64
	sellTP         float64
	sellLeverage   int
	sellConfirm    bool
)

var sellCmd = &cobra.Command{
	Use:   "sell SYMBOL",
	Short: "Place a sell (short) order",
	Long: `Place a SELL (short) order for a trading instrument.

Note: This opens a SHORT position, not selling existing holdings.
To close an existing position, use 'etoro close POSITION_ID'.

Order types:
  - market: Execute at current market price (default)
  - limit: Execute when price reaches --limit-price

Amount specification:
  - --amount: USD value to invest
  - --qty: Number of units to sell (alternative to amount)

Risk management:
  - --sl: Stop loss price level
  - --tp: Take profit price level
  - --leverage: Leverage multiplier (default from config)

Safety:
  Without --confirm flag, only shows a dry-run preview.
  With --confirm (or ETORO_CONFIRM=1 env), executes if all guards pass.

Examples:
  etoro sell AAPL --amount 100                        # Dry-run preview
  etoro sell AAPL --amount 100 --confirm              # Execute short
  etoro sell TSLA --qty 10 --type limit --limit-price 250 --confirm
  etoro sell AAPL --amount 100 --json                 # JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		symbol := args[0]
		printer := output.NewPrinter(jsonOutput)

		// Validate inputs
		if sellAmount <= 0 && sellQty <= 0 {
			exitWithError(fmt.Errorf("must specify either --amount or --qty"), 3)
		}

		// Safety checks
		checkKillSwitch()
		checkSymbolAllowed(symbol)
		if sellAmount > 0 {
			checkMaxTrade(sellAmount)
		}

		// Set default leverage from config
		if sellLeverage == 0 {
			sellLeverage = config.AppConfig.Trading.DefaultLeverage
			if sellLeverage == 0 {
				sellLeverage = 1
			}
		}

		// Validate order type
		if sellType != "market" && sellType != "limit" {
			exitWithError(fmt.Errorf("invalid order type: %s (must be 'market' or 'limit')"), 3)
		}

		if sellType == "limit" && sellLimitPrice <= 0 {
			exitWithError(fmt.Errorf("--limit-price is required for limit orders"), 3)
		}

		c := client.New(verbose)

		// Resolve symbol to instrument
		instrument, err := c.GetInstrumentBySymbol(symbol)
		if err != nil {
			exitWithError(err, 1)
		}

		// Build order request
		orderReq := &client.OrderRequest{
			InstrumentID: instrument.InstrumentID,
			Direction:    "SELL",
			OrderType:    sellType,
			Amount:       sellAmount,
			Quantity:     sellQty,
			LimitPrice:   sellLimitPrice,
			StopLoss:     sellSL,
			TakeProfit:   sellTP,
			Leverage:     sellLeverage,
		}

		// Check if we should execute
		shouldExecute := isConfirmed(sellConfirm) && isExecutionEnabled()

		if !shouldExecute {
			// Dry-run mode - show preview
			dryRunResult := &client.DryRunResult{
				Action:       "SELL",
				Symbol:       symbol,
				Direction:    "SELL",
				Amount:       sellAmount,
				Quantity:     sellQty,
				OrderType:    sellType,
				LimitPrice:   sellLimitPrice,
				StopLoss:     sellSL,
				TakeProfit:   sellTP,
				Leverage:     sellLeverage,
				WouldExecute: false,
				Message:      "DRY-RUN: Order not executed. Use --confirm to execute.",
			}

			if !isConfirmed(sellConfirm) {
				dryRunResult.Message = "DRY-RUN: Add --confirm flag to execute this order."
			} else if !config.IsExecutionEnabled() {
				dryRunResult.Message = "DRY-RUN: Set execution_enabled: true in config to allow trading."
			} else if dryRun {
				dryRunResult.Message = "DRY-RUN: Remove --dry-run flag to execute this order."
			}

			if jsonOutput {
				printer.PrintJSON(dryRunResult)
			} else {
				fmt.Println("=== DRY-RUN ORDER PREVIEW ===")
				fmt.Printf("Action:      SELL (Short)\n")
				fmt.Printf("Symbol:      %s\n", symbol)
				fmt.Printf("Order Type:  %s\n", sellType)
				if sellAmount > 0 {
					fmt.Printf("Amount:      %s\n", output.FormatMoney(sellAmount))
				}
				if sellQty > 0 {
					fmt.Printf("Quantity:    %.4f\n", sellQty)
				}
				if sellLimitPrice > 0 {
					fmt.Printf("Limit Price: %s\n", output.FormatMoney(sellLimitPrice))
				}
				if sellSL > 0 {
					fmt.Printf("Stop Loss:   %s\n", output.FormatMoney(sellSL))
				}
				if sellTP > 0 {
					fmt.Printf("Take Profit: %s\n", output.FormatMoney(sellTP))
				}
				fmt.Printf("Leverage:    %dx\n", sellLeverage)
				fmt.Println("==============================")
				fmt.Println(dryRunResult.Message)
			}
			return
		}

		// Execute order
		resp, err := c.PlaceOrder(orderReq)
		if err != nil {
			exitWithError(err, 1)
		}

		if jsonOutput {
			printer.PrintJSON(resp)
		} else {
			fmt.Println("=== ORDER EXECUTED ===")
			fmt.Printf("Order ID:    %d\n", resp.OrderID)
			if resp.PositionID > 0 {
				fmt.Printf("Position ID: %d\n", resp.PositionID)
			}
			fmt.Printf("Status:      %s\n", resp.Status)
			if resp.Message != "" {
				fmt.Printf("Message:     %s\n", resp.Message)
			}
		}
	},
}

func init() {
	sellCmd.Flags().Float64Var(&sellAmount, "amount", 0, "USD amount to invest")
	sellCmd.Flags().Float64Var(&sellQty, "qty", 0, "number of units to sell")
	sellCmd.Flags().StringVar(&sellType, "type", "market", "order type: market or limit")
	sellCmd.Flags().Float64Var(&sellLimitPrice, "limit-price", 0, "limit price for limit orders")
	sellCmd.Flags().Float64Var(&sellSL, "sl", 0, "stop loss price")
	sellCmd.Flags().Float64Var(&sellTP, "tp", 0, "take profit price")
	sellCmd.Flags().IntVar(&sellLeverage, "leverage", 0, "leverage multiplier (default from config)")
	sellCmd.Flags().BoolVar(&sellConfirm, "confirm", false, "confirm and execute the order")

	rootCmd.AddCommand(sellCmd)
}
