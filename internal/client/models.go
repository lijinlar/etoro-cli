// Package client provides the HTTP client and data models for the eToro API.
package client

import "time"

// PnLResponse is the top-level response from /trading/info/real/pnl
type PnLResponse struct {
	ClientPortfolio ClientPortfolio `json:"clientPortfolio"`
}

// ClientPortfolio contains the account's live portfolio data
type ClientPortfolio struct {
	Credit          float64    `json:"credit"`
	BonusCredit     float64    `json:"bonusCredit"`
	UnrealizedPnL   float64    `json:"unrealizedPnL"`
	AccountCurrencyID int      `json:"accountCurrencyId"`
	Positions       []Position `json:"positions"`
	Orders          []Order    `json:"orders"`
	OrdersForOpen   []Order    `json:"ordersForOpen"`
	OrdersForClose  []Order    `json:"ordersForClose"`
}

// AccountInfo represents the account summary response
type AccountInfo struct {
	Balance        float64    `json:"balance"`
	UnrealizedPL   float64    `json:"unrealizedPL"`
	Positions      []Position `json:"positions"`
	OpenOrders     []Order    `json:"openOrders"`
}

// Instrument represents a tradeable instrument from the eToro API
type Instrument struct {
	InstrumentID   int     `json:"internalInstrumentId"`
	Symbol         string  `json:"internalSymbolFull"`
	Name           string  `json:"internalInstrumentDisplayName"`
	AssetClass     string  `json:"internalAssetClassName"`
	Exchange       string  `json:"internalExchangeName"`
	CurrentRate    float64 `json:"currentRate"`
	DailyChange    float64 `json:"dailyPriceChange"`
	IsTradable     bool    `json:"isCurrentlyTradable"`
	IsBuyEnabled   bool    `json:"isBuyEnabled"`
}

// InstrumentRate represents live price data for an instrument from the eToro API
type InstrumentRate struct {
	InstrumentID int     `json:"instrumentID"`
	Symbol       string  `json:"-"` // populated client-side after lookup
	Bid          float64 `json:"bid"`
	Ask          float64 `json:"ask"`
	Spread       float64 `json:"-"` // calculated
	DailyChange  float64 `json:"-"` // not in rates endpoint
	DailyHigh    float64 `json:"-"` // not in rates endpoint
	DailyLow     float64 `json:"-"` // not in rates endpoint
	LastUpdated  string  `json:"date"`
}

// Position represents an open trading position
type Position struct {
	PositionID   int       `json:"positionId"`
	InstrumentID int       `json:"instrumentId"`
	Symbol       string    `json:"symbol"`
	Direction    string    `json:"direction"`
	Quantity     float64   `json:"quantity"`
	OpenPrice    float64   `json:"openPrice"`
	CurrentPrice float64   `json:"currentPrice"`
	PL           float64   `json:"pl"`
	PLPercent    float64   `json:"plPercent"`
	Leverage     int       `json:"leverage"`
	StopLoss     float64   `json:"stopLoss,omitempty"`
	TakeProfit   float64   `json:"takeProfit,omitempty"`
	OpenDate     time.Time `json:"openDate"`
}

// Order represents a pending order
type Order struct {
	OrderID      int       `json:"orderId"`
	InstrumentID int       `json:"instrumentId"`
	Symbol       string    `json:"symbol"`
	Direction    string    `json:"direction"`
	OrderType    string    `json:"orderType"`
	Quantity     float64   `json:"quantity"`
	Amount       float64   `json:"amount"`
	LimitPrice   float64   `json:"limitPrice,omitempty"`
	StopLoss     float64   `json:"stopLoss,omitempty"`
	TakeProfit   float64   `json:"takeProfit,omitempty"`
	Leverage     int       `json:"leverage"`
	CreatedAt    time.Time `json:"createdAt"`
	Status       string    `json:"status"`
}

// OrderRequest represents a new order to be placed
type OrderRequest struct {
	InstrumentID int     `json:"instrumentId"`
	Direction    string  `json:"direction"`
	OrderType    string  `json:"orderType"`
	Amount       float64 `json:"amount,omitempty"`
	Quantity     float64 `json:"quantity,omitempty"`
	LimitPrice   float64 `json:"limitPrice,omitempty"`
	StopLoss     float64 `json:"stopLoss,omitempty"`
	TakeProfit   float64 `json:"takeProfit,omitempty"`
	Leverage     int     `json:"leverage"`
}

// OrderResponse represents the response after placing an order
type OrderResponse struct {
	OrderID    int    `json:"orderId"`
	PositionID int    `json:"positionId,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

// ClosePositionRequest represents a request to close a position
type ClosePositionRequest struct {
	PartialQuantity float64 `json:"partialQuantity,omitempty"`
}

// ClosePositionResponse represents the response after closing a position
type ClosePositionResponse struct {
	PositionID int     `json:"positionId"`
	ClosedPL   float64 `json:"closedPL"`
	Status     string  `json:"status"`
	Message    string  `json:"message,omitempty"`
}

// CancelOrderResponse represents the response after canceling an order
type CancelOrderResponse struct {
	OrderID int    `json:"orderId"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// TradeHistory represents a closed trade
type TradeHistory struct {
	PositionID  int       `json:"positionId"`
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"`
	Quantity    float64   `json:"quantity"`
	OpenPrice   float64   `json:"openPrice"`
	ClosePrice  float64   `json:"closePrice"`
	PL          float64   `json:"pl"`
	PLPercent   float64   `json:"plPercent"`
	OpenDate    time.Time `json:"openDate"`
	CloseDate   time.Time `json:"closeDate"`
}

// WatchlistItem represents an item in the watchlist
type WatchlistItem struct {
	InstrumentID int    `json:"instrumentId"`
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	AddedAt      string `json:"addedAt"`
}

// Watchlist represents the user's watchlist
type Watchlist struct {
	Items []WatchlistItem `json:"items"`
}

// PortfolioSummary represents a portfolio overview
type PortfolioSummary struct {
	TotalPositions int        `json:"totalPositions"`
	TotalValue     float64    `json:"totalValue"`
	UnrealizedPL   float64    `json:"unrealizedPL"`
	TopGainers     []Position `json:"topGainers"`
	TopLosers      []Position `json:"topLosers"`
}

// RiskMetrics represents risk analysis data
type RiskMetrics struct {
	MarginUtilization float64            `json:"marginUtilization"`
	TotalExposure     float64            `json:"totalExposure"`
	SymbolExposure    map[string]float64 `json:"symbolExposure"`
	WarningLevel      string             `json:"warningLevel,omitempty"`
}

// APIError represents an error response from the eToro API
type APIError struct {
	Code         int    `json:"code"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func (e APIError) Error() string {
	if e.ErrorMessage != "" {
		if e.ErrorCode != "" {
			return e.ErrorCode + ": " + e.ErrorMessage
		}
		return e.ErrorMessage
	}
	return "unknown API error"
}

// DryRunResult represents the result of a dry-run operation
type DryRunResult struct {
	Action      string      `json:"action"`
	Symbol      string      `json:"symbol"`
	Direction   string      `json:"direction"`
	Amount      float64     `json:"amount,omitempty"`
	Quantity    float64     `json:"quantity,omitempty"`
	OrderType   string      `json:"orderType,omitempty"`
	LimitPrice  float64     `json:"limitPrice,omitempty"`
	StopLoss    float64     `json:"stopLoss,omitempty"`
	TakeProfit  float64     `json:"takeProfit,omitempty"`
	Leverage    int         `json:"leverage,omitempty"`
	WouldExecute bool       `json:"wouldExecute"`
	Message     string      `json:"message"`
}
