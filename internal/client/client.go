// Package client provides the HTTP client for interacting with the eToro API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lijinlar/etoro-cli/internal/config"
)

// Client is the eToro API client
type Client struct {
	baseURL    string
	publicKey  string
	userKey    string
	httpClient *http.Client
	verbose    bool
}

// New creates a new eToro API client
func New(verbose bool) *Client {
	return &Client{
		baseURL:   config.AppConfig.Etoro.BaseURL,
		publicKey: config.AppConfig.Etoro.PublicKey,
		userKey:   config.AppConfig.Etoro.UserKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}
}

// doRequest performs an HTTP request with authentication headers
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers
	req.Header.Set("X-Api-Key", c.publicKey)
	req.Header.Set("X-User-Key", c.userKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-Id", uuid.New().String())

	if c.verbose {
		fmt.Printf("[HTTP] %s %s\n", method, fullURL)
		if body != nil {
			jsonBody, _ := json.MarshalIndent(body, "", "  ")
			fmt.Printf("[HTTP] Request Body:\n%s\n", string(jsonBody))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("[HTTP] Response Status: %d\n", resp.StatusCode)
		fmt.Printf("[HTTP] Response Body:\n%s\n", string(respBody))
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		apiErr.Code = resp.StatusCode
		return nil, apiErr
	}

	return respBody, nil
}

// GetPnL retrieves the full portfolio P&L from the real account
func (c *Client) GetPnL() (*PnLResponse, error) {
	body, err := c.doRequest("GET", "/trading/info/real/pnl", nil)
	if err != nil {
		return nil, err
	}

	var pnl PnLResponse
	if err := json.Unmarshal(body, &pnl); err != nil {
		return nil, fmt.Errorf("failed to parse P&L response: %w", err)
	}

	return &pnl, nil
}

// GetAccount retrieves account information (derived from PnL endpoint)
func (c *Client) GetAccount() (*AccountInfo, error) {
	pnl, err := c.GetPnL()
	if err != nil {
		return nil, err
	}

	return &AccountInfo{
		Balance:      pnl.ClientPortfolio.Credit,
		UnrealizedPL: pnl.ClientPortfolio.UnrealizedPnL,
		Positions:    pnl.ClientPortfolio.Positions,
		OpenOrders:   pnl.ClientPortfolio.OrdersForOpen,
	}, nil
}

// searchResult is the paginated response wrapper from /market-data/search
type searchResult struct {
	Page      int          `json:"page"`
	PageSize  int          `json:"pageSize"`
	TotalItems int         `json:"totalItems"`
	Items     []Instrument `json:"items"`
}

// SearchInstruments searches for instruments by query string
func (c *Client) SearchInstruments(query string) ([]Instrument, error) {
	path := "/market-data/search?phrase=" + url.QueryEscape(query)
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result searchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse instruments response: %w", err)
	}

	return result.Items, nil
}

// GetInstrumentBySymbol finds an instrument by its exact symbol
func (c *Client) GetInstrumentBySymbol(symbol string) (*Instrument, error) {
	path := "/market-data/search?internalSymbolFull=" + url.QueryEscape(strings.ToUpper(symbol))
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result searchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse instrument response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("instrument not found: %s", symbol)
	}

	return &result.Items[0], nil
}

// GetInstrumentRate gets live rates for an instrument
func (c *Client) GetInstrumentRate(instrumentID int) (*InstrumentRate, error) {
	path := fmt.Sprintf("/market-data/instruments/rates?instrumentIds=%d", instrumentID)
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rates []InstrumentRate `json:"rates"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse rate response: %w", err)
	}

	if len(result.Rates) == 0 {
		return nil, fmt.Errorf("no rate data for instrument %d", instrumentID)
	}

	rate := result.Rates[0]
	rate.Spread = rate.Ask - rate.Bid
	return &rate, nil
}

// GetPositions retrieves open positions (from PnL endpoint)
func (c *Client) GetPositions() ([]Position, error) {
	pnl, err := c.GetPnL()
	if err != nil {
		return nil, err
	}

	return pnl.ClientPortfolio.Positions, nil
}

// GetOrders retrieves pending orders (from PnL endpoint)
func (c *Client) GetOrders() ([]Order, error) {
	pnl, err := c.GetPnL()
	if err != nil {
		return nil, err
	}

	// Combine all pending order types
	orders := pnl.ClientPortfolio.Orders
	orders = append(orders, pnl.ClientPortfolio.OrdersForOpen...)
	return orders, nil
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(req *OrderRequest) (*OrderResponse, error) {
	body, err := c.doRequest("POST", "/trading/orders", req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}

	return &resp, nil
}

// CancelOrder cancels a pending order
func (c *Client) CancelOrder(orderID int) (*CancelOrderResponse, error) {
	path := fmt.Sprintf("/trading/orders/%d", orderID)
	body, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	var resp CancelOrderResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse cancel response: %w", err)
	}

	return &resp, nil
}

// ClosePosition closes an open position
func (c *Client) ClosePosition(positionID int, partialQty float64) (*ClosePositionResponse, error) {
	path := fmt.Sprintf("/trading/positions/%d/close", positionID)
	var reqBody interface{}
	if partialQty > 0 {
		reqBody = &ClosePositionRequest{PartialQuantity: partialQty}
	}

	body, err := c.doRequest("POST", path, reqBody)
	if err != nil {
		return nil, err
	}

	var resp ClosePositionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse close response: %w", err)
	}

	return &resp, nil
}

// GetHistory retrieves closed trade history
func (c *Client) GetHistory(from, to, symbol string, limit int) ([]TradeHistory, error) {
	params := url.Values{}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := "/trading/history"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var history []TradeHistory
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	return history, nil
}

// GetWatchlist retrieves the user's watchlist
func (c *Client) GetWatchlist() (*Watchlist, error) {
	body, err := c.doRequest("GET", "/watchlist", nil)
	if err != nil {
		return nil, err
	}

	var watchlist Watchlist
	if err := json.Unmarshal(body, &watchlist); err != nil {
		return nil, fmt.Errorf("failed to parse watchlist response: %w", err)
	}

	return &watchlist, nil
}

// AddToWatchlist adds a symbol to the watchlist
func (c *Client) AddToWatchlist(instrumentID int) error {
	req := map[string]int{"instrumentId": instrumentID}
	_, err := c.doRequest("POST", "/watchlist", req)
	return err
}

// RemoveFromWatchlist removes a symbol from the watchlist
func (c *Client) RemoveFromWatchlist(instrumentID int) error {
	path := fmt.Sprintf("/watchlist/%d", instrumentID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}
