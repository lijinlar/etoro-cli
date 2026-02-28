// Package output provides helpers for formatting CLI output as tables or JSON.
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

// Format represents the output format type
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Printer handles output formatting
type Printer struct {
	format Format
}

// NewPrinter creates a new output printer
func NewPrinter(jsonOutput bool) *Printer {
	format := FormatTable
	if jsonOutput {
		format = FormatJSON
	}
	return &Printer{format: format}
}

// PrintJSON outputs data as formatted JSON
func (p *Printer) PrintJSON(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// PrintTable outputs data as a formatted table
func (p *Printer) PrintTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(false)
	table.AppendBulk(rows)
	table.Render()
}

// Print outputs data in the configured format
func (p *Printer) Print(data interface{}, headers []string, rowFunc func() [][]string) error {
	if p.format == FormatJSON {
		return p.PrintJSON(data)
	}
	p.PrintTable(headers, rowFunc())
	return nil
}

// PrintMessage prints a simple message (respects JSON format)
func (p *Printer) PrintMessage(message string) {
	if p.format == FormatJSON {
		p.PrintJSON(map[string]string{"message": message})
	} else {
		fmt.Println(message)
	}
}

// PrintError prints an error message (respects JSON format)
func (p *Printer) PrintError(err error) {
	if p.format == FormatJSON {
		p.PrintJSON(map[string]string{"error": err.Error()})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// PrintSuccess prints a success message
func (p *Printer) PrintSuccess(message string) {
	if p.format == FormatJSON {
		p.PrintJSON(map[string]interface{}{
			"success": true,
			"message": message,
		})
	} else {
		fmt.Printf("Success: %s\n", message)
	}
}

// PrintWarning prints a warning message
func (p *Printer) PrintWarning(message string) {
	if p.format == FormatJSON {
		p.PrintJSON(map[string]interface{}{
			"warning": true,
			"message": message,
		})
	} else {
		fmt.Printf("Warning: %s\n", message)
	}
}

// FormatMoney formats a float as currency
func FormatMoney(amount float64) string {
	if amount >= 0 {
		return fmt.Sprintf("$%.2f", amount)
	}
	return fmt.Sprintf("-$%.2f", -amount)
}

// FormatPercent formats a float as percentage
func FormatPercent(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.2f%%", pct)
	}
	return fmt.Sprintf("%.2f%%", pct)
}

// FormatPL formats P&L with color indicators (for table output)
func FormatPL(pl float64) string {
	if pl >= 0 {
		return fmt.Sprintf("+$%.2f", pl)
	}
	return fmt.Sprintf("-$%.2f", -pl)
}
