package main

import (
	"fmt"
	"os"

	"github.com/garysze77/stxai-agent/cmd"

	"github.com/spf13/cobra"
)

var version = "0.10.0"

var rootCmd = &cobra.Command{
	Use:     "stxai",
	Version: version,
	Short:   "STX AI Agent — Autonomous Financial AI Agent",
	Long: `STX AI is an open-source autonomous AI agent for US & HK stock analysis.

Self-hosted client that connects to the STX AI Cloud API.
Subscribe at https://stxai.app to get your API key.`,
	Run: func(c *cobra.Command, args []string) {
		c.Help()
	},
}

func init() {
	rootCmd.AddCommand(cmd.SetupCmd())
	rootCmd.AddCommand(cmd.StartCmd())
	rootCmd.AddCommand(cmd.ChatCmd())
	rootCmd.AddCommand(cmd.AnalyzeCmd())
	rootCmd.AddCommand(cmd.ServiceCmd())
	rootCmd.AddCommand(cmd.UpdateCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
