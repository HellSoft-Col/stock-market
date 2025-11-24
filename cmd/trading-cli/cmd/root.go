package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	logLevel string

	// Color helpers
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "trading-cli",
	Short: "🥑 Andorian Avocado Exchange Trading Bot Manager",
	Long: `
╔═══════════════════════════════════════════════════════╗
║  🥑 Andorian Avocado Exchange - Trading Bot Manager  ║
║                                                       ║
║  Manage automated trading bots with ease             ║
║  • Start/Stop bots                                   ║
║  • Monitor P&L in real-time                          ║
║  • View detailed statistics                          ║
╚═══════════════════════════════════════════════════════╝

Built with ❤️  for maximum profitability!
`,
}

// Execute adds all child commands and sets flags
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "automated-clients.yaml", "config file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "log level (debug, info, warn, error)")
}

// Helper functions for colored output
func printSuccess(msg string) {
	fmt.Printf("%s %s\n", green("✅"), msg)
}

func printError(msg string) {
	fmt.Printf("%s %s\n", red("❌"), msg)
}

func printWarning(msg string) {
	fmt.Printf("%s %s\n", yellow("⚠️"), msg)
}

func printInfo(msg string) {
	fmt.Printf("%s %s\n", cyan("ℹ️"), msg)
}

func printHeader(title string) {
	fmt.Printf("\n%s %s %s\n", blue("━━━"), bold(title), blue("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
}
