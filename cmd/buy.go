package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lijinlar/etoro-cli/internal/client"
	"github.com/lijinlar/etoro-cli/internal/config"
	"github.com/lijinlar/etoro-cli/internal/output"
)

var (
	buyAmount     float64
	buyQty        float64
	buyType       string
	buyLimitPrice float64
	buySL         float64
	buyTP         float64
	buyLeverage   int
	buyConfirm    bool
)

var buyCmd = &cobra.Command{
	Use:   "buy SYMBOL",
	Short: "Place a buy order",
	Long: `Place a BUY order for a trading instrument.

Order types:
  - market: Execute at current market price (default)
  - limit: Execute when price reaches --limit-price

Amount specification:
  - --amount: USD value to invest
  - --qty: Number of units to buy (alternative to amount)

Risk management:
  - --sl: Stop loss price level
  - --tp: Take profit price level
  - --leverage: Leverage multiplier (default from config)

Safety:
  Without --confirm flag, only shows a dry-run preview.
  With --confirm (or ETORO_CONFIRM=1 env), executes if all guards pass.

Safety guards checked:
  1. Kill switch must be disabled
  2. Execution must be enabled in config
  3. Amount must not exceed max_trade_usd
  4. Symbol must be in allowlist (if configured)

Examples:
  etoro buy AAPL --amount 100                        # Dry-run preview
  etoro buy AAPL --amount 100 --confirm              # Execute buy
  etoro buy AAPL --qty 5 --type limit --limit-price 150 --confirm
  etoro buy TSLA --amount 500 --sl 200 --tp 300 --confirm
  etoro buy AAPL --amount 100 --json                 # JSON output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()

		symbol := args[0]
		printer := output.NewPrinter(jsonOutput)

		// Validate inputs
		if buyAmount <= 0 && buyQty <= 0 {
			exitWithError(fmt.Errorf("must specify either --amount or --qty"), 3)
		}

		// Safety checks
		checkKillSwitch()
		checkSymbolAllowed(symbol)
		if buyAmount > 0 {
			checkMaxTrade(buyAmount)
		}

		// Set default leverage from config
		if buyLeverage == 0 {
			buyLeverage = config.AppConfig.Trading.DefaultLeverage
			if buyLeverage == 0 {
				buyLeverage = 1
			}
		}

		// Validate order type
		if buyType != "market" && buyType != "limit" {
			exitWithError(fmt.Errorf("invalid order type: %s (must be 'market' or 'limit')"), 3)
		}

		if buyType == "limit" && buyLimitPrice <= 0 {
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
			Direction:    "BUY",
			OrderType:    buyType,
			Amount:       buyAmount,
			Quantity:     buyQty,
			LimitPrice:   buyLimitPrice,
			StopLoss:     buySL,
			TakeProfit:   buyTP,
			Leverage:     buyLeverage,
		}

		// Check if we should execute
		shouldExecute := isConfirmed(buyConfirm) && isExecutionEnabled()

		if !shouldExecute {
			// Dry-run mode - show preview
			dryRunResult := &client.DryRunResult{
				Action:       "BUY",
				Symbol:       symbol,
				Direction:    "BUY",
				Amount:       buyAmount,
				Quantity:     buyQty,
				OrderType:    buyType,
				LimitPrice:   buyLimitPrice,
				StopLoss:     buySL,
				TakeProfit:   buyTP,
				Leverage:     buyLeverage,
				WouldExecute: false,
				Message:      "DRY-RUN: Order not executed. Use --confirm to execute.",
			}

			if !isConfirmed(buyConfirm) {
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
				fmt.Printf("Action:      BUY\n")
				fmt.Printf("Symbol:      %s\n", symbol)
				fmt.Printf("Order Type:  %s\n", buyType)
				if buyAmount > 0 {
					fmt.Printf("Amount:      %s\n", output.FormatMoney(buyAmount))
				}
				if buyQty > 0 {
					fmt.Printf("Quantity:    %.4f\n", buyQty)
				}
				if buyLimitPrice > 0 {
					fmt.Printf("Limit Price: %s\n", output.FormatMoney(buyLimitPrice))
				}
				if buySL > 0 {
					fmt.Printf("Stop Loss:   %s\n", output.FormatMoney(buySL))
				}
				if buyTP > 0 {
					fmt.Printf("Take Profit: %s\n", output.FormatMoney(buyTP))
				}
				fmt.Printf("Leverage:    %dx\n", buyLeverage)
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
	buyCmd.Flags().Float64Var(&buyAmount, "amount", 0, "USD amount to invest")
	buyCmd.Flags().Float64Var(&buyQty, "qty", 0, "number of units to buy")
	buyCmd.Flags().StringVar(&buyType, "type", "market", "order type: market or limit")
	buyCmd.Flags().Float64Var(&buyLimitPrice, "limit-price", 0, "limit price for limit orders")
	buyCmd.Flags().Float64Var(&buySL, "sl", 0, "stop loss price")
	buyCmd.Flags().Float64Var(&buyTP, "tp", 0, "take profit price")
	buyCmd.Flags().IntVar(&buyLeverage, "leverage", 0, "leverage multiplier (default from config)")
	buyCmd.Flags().BoolVar(&buyConfirm, "confirm", false, "confirm and execute the order")

	rootCmd.AddCommand(buyCmd)
}
