// Package config handles loading and validation of etoro-cli configuration
// from YAML files and environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the full application configuration
type Config struct {
	Etoro   EtoroConfig   `mapstructure:"etoro"`
	Trading TradingConfig `mapstructure:"trading"`
	Output  OutputConfig  `mapstructure:"output"`
}

// EtoroConfig holds eToro API credentials and settings
type EtoroConfig struct {
	PublicKey string `mapstructure:"public_key"`
	UserKey   string `mapstructure:"user_key"`
	BaseURL   string `mapstructure:"base_url"`
}

// TradingConfig holds trading safety and execution settings
type TradingConfig struct {
	ExecutionEnabled bool     `mapstructure:"execution_enabled"`
	KillSwitch       bool     `mapstructure:"kill_switch"`
	MaxTradeUSD      float64  `mapstructure:"max_trade_usd"`
	MaxPositions     int      `mapstructure:"max_positions"`
	SymbolAllowlist  []string `mapstructure:"symbol_allowlist"`
	DefaultLeverage  int      `mapstructure:"default_leverage"`
}

// OutputConfig holds output formatting preferences
type OutputConfig struct {
	Format   string `mapstructure:"format"`
	Timezone string `mapstructure:"timezone"`
}

// AppConfig is the global configuration instance
var AppConfig Config

// Load reads configuration from file and environment variables
func Load(configPath string) error {
	v := viper.New()

	// Set defaults
	v.SetDefault("etoro.base_url", "https://api.etoro.com")
	v.SetDefault("trading.execution_enabled", false)
	v.SetDefault("trading.kill_switch", true)
	v.SetDefault("trading.max_trade_usd", 500)
	v.SetDefault("trading.max_positions", 10)
	v.SetDefault("trading.default_leverage", 1)
	v.SetDefault("output.format", "table")
	v.SetDefault("output.timezone", "local")

	// Config file search paths
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Check current directory first
		v.AddConfigPath(".")
		v.SetConfigName("etoro")

		// Then check home directory
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".etoro"))
			v.SetConfigName("config")
		}
	}

	v.SetConfigType("yaml")

	// Environment variable bindings
	v.SetEnvPrefix("ETORO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit env bindings
	v.BindEnv("etoro.public_key", "ETORO_PUBLIC_KEY")
	v.BindEnv("etoro.user_key", "ETORO_USER_KEY")
	v.BindEnv("trading.execution_enabled", "ETORO_EXECUTION_ENABLED")
	v.BindEnv("trading.kill_switch", "ETORO_KILL_SWITCH")
	v.BindEnv("trading.max_trade_usd", "ETORO_MAX_TRADE_USD")

	// Read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into struct
	if err := v.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	return nil
}

// Validate checks that required configuration values are present
func Validate() error {
	if AppConfig.Etoro.PublicKey == "" {
		return fmt.Errorf("etoro.public_key is required (set via config or ETORO_PUBLIC_KEY env)")
	}
	if AppConfig.Etoro.UserKey == "" {
		return fmt.Errorf("etoro.user_key is required (set via config or ETORO_USER_KEY env)")
	}
	if AppConfig.Etoro.BaseURL == "" {
		AppConfig.Etoro.BaseURL = "https://api.etoro.com"
	}
	return nil
}

// IsKillSwitchActive returns true if the kill switch is enabled
func IsKillSwitchActive() bool {
	return AppConfig.Trading.KillSwitch
}

// IsExecutionEnabled returns true if trade execution is allowed
func IsExecutionEnabled() bool {
	return AppConfig.Trading.ExecutionEnabled
}

// CheckMaxTradeUSD validates if an order amount is within limits
func CheckMaxTradeUSD(amount float64) error {
	if AppConfig.Trading.MaxTradeUSD > 0 && amount > AppConfig.Trading.MaxTradeUSD {
		return fmt.Errorf("order amount $%.2f exceeds max_trade_usd limit of $%.2f", amount, AppConfig.Trading.MaxTradeUSD)
	}
	return nil
}

// IsSymbolAllowed checks if a symbol is in the allowlist (empty = all allowed)
func IsSymbolAllowed(symbol string) bool {
	if len(AppConfig.Trading.SymbolAllowlist) == 0 {
		return true
	}
	symbol = strings.ToUpper(symbol)
	for _, s := range AppConfig.Trading.SymbolAllowlist {
		if strings.ToUpper(s) == symbol {
			return true
		}
	}
	return false
}

// GetConfirmFromEnv checks ETORO_CONFIRM environment variable
func GetConfirmFromEnv() bool {
	val := os.Getenv("ETORO_CONFIRM")
	return val == "1" || strings.ToLower(val) == "true"
}
