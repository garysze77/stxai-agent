package display

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	headerColor = color.New(color.FgCyan, color.Bold)
	bullColor   = color.New(color.FgGreen)
	bearColor   = color.New(color.FgRed)
	leadColor   = color.New(color.FgYellow)
	signalColor = color.New(color.FgMagenta, color.Bold)
	errorColor  = color.New(color.FgRed, color.Bold)
	priceColor  = color.New(color.FgWhite, color.Bold)
	dimColor    = color.New(color.Faint)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func ShowPrice(ticker string, price float64, currency string) {
	headerColor.Printf("\n  %s", ticker)
	if price > 0 {
		priceColor.Printf("  $%.2f %s", price, currency)
	}
	fmt.Println()
}

func ShowNodeHeader(node string) {
	switch {
	case strings.Contains(node, "bullish"):
		bullColor.Println("\n  ── Bull Case ──\n")
	case strings.Contains(node, "bearish"):
		bearColor.Println("\n  ── Bear Case ──\n")
	case strings.Contains(node, "lead"):
		leadColor.Println("\n  ── Synthesis ──\n")
	case strings.Contains(node, "signal"):
		signalColor.Println("\n  ── Signal ──\n")
	default:
		dimColor.Printf("\n  ── %s ──\n\n", node)
	}
}

func ShowNodeContent(node string, data json.RawMessage) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	switch {
	case strings.Contains(node, "bullish"):
		if thesis, ok := payload["thesis"]; ok {
			printJSONString(thesis)
		}
	case strings.Contains(node, "bearish"):
		if thesis, ok := payload["thesis"]; ok {
			printJSONString(thesis)
		}
	case strings.Contains(node, "lead"):
		if report, ok := payload["report"]; ok {
			printJSONString(report)
		}
	case strings.Contains(node, "signal"):
		if card, ok := payload["signal_card"]; ok {
			printJSONString(card)
		}
		if bias, ok := payload["directional_bias"]; ok {
			var s string
			json.Unmarshal(bias, &s)
			if s != "" {
				signalColor.Printf("\n  Direction: %s", s)
			}
		}
		if score, ok := payload["confidence_score"]; ok {
			var n float64
			json.Unmarshal(score, &n)
			if n > 0 {
				signalColor.Printf(" | Confidence: %.0f/100", n)
			}
		}
		if strength, ok := payload["signal_strength"]; ok {
			var s string
			json.Unmarshal(strength, &s)
			if s != "" {
				signalColor.Printf(" | Strength: %s", s)
			}
		}
		fmt.Println()
	}
}

func printJSONString(raw json.RawMessage) {
	var s string
	json.Unmarshal(raw, &s)
	fmt.Println(s)
}

// Spinner shows an animated spinner with a label. Returns a stop function.
func Spinner(label string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				fmt.Fprint(os.Stderr, "\r\033[K")
				close(done)
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r  %s %s", spinnerFrames[i%len(spinnerFrames)], label)
				i++
			}
		}
	}()

	return func() {
		close(stop)
		<-done
	}
}

func ShowError(data json.RawMessage) {
	var payload map[string]interface{}
	json.Unmarshal(data, &payload)
	node, _ := payload["node"].(string)
	errMsg, _ := payload["error"].(string)
	errorColor.Printf("\n  ✗ Error in %s: %s\n", node, errMsg)
}

func ShowHeartbeat(data json.RawMessage) {
	// Heartbeats are handled by the spinner — nothing to display
}
