package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/spf13/cobra"
)

var (
	watch   bool
	refresh int
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "📊 Show bot status",
	Long:  `Display current status of all running bots with real-time updates`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch mode with live updates")
	statusCmd.Flags().IntVarP(&refresh, "refresh", "r", 3, "refresh interval in seconds (for watch mode)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	if watch {
		return watchStatus(cfg)
	}

	return showStatus(cfg)
}

func showStatus(cfg *config.Config) error {
	printHeader("📊 Bot Status Dashboard")

	enabledClients := cfg.GetEnabledClients()

	// Print table header
	fmt.Println()
	fmt.Printf("%s\n", bold(cyan("┌────────────────────────────────────────────────────────────────────────────────┐")))
	fmt.Printf("│ %-20s │ %-12s │ %-15s │ %-12s │ %-12s │\n",
		bold("Bot Name"), bold("Status"), bold("Strategy"), bold("Species"), bold("Config"))
	fmt.Printf("%s\n", bold(cyan("├────────────────────────────────────────────────────────────────────────────────┤")))

	for _, client := range enabledClients {
		status := green("✅ Configured")
		if !client.Enabled {
			status = red("❌ Disabled")
		}

		tokenStr := client.Token
		if len(tokenStr) > 8 {
			tokenStr = tokenStr[:8] + "..."
		}

		fmt.Printf("│ %-20s │ %-20s │ %-15s │ %-12s │ %-12s │\n",
			client.Name,
			status,
			client.Strategy,
			client.Species,
			tokenStr)
	}

	fmt.Printf("%s\n", bold(cyan("└────────────────────────────────────────────────────────────────────────────────┘")))

	fmt.Printf("\n%s Total bots: %s enabled\n",
		"📈",
		green(fmt.Sprintf("%d", len(enabledClients))))

	printInfo("Use --watch flag for live updates")
	printInfo("Use 'trading-cli start' to launch bots")

	return nil
}

func watchStatus(cfg *config.Config) error {
	printInfo(fmt.Sprintf("Starting watch mode (refresh every %ds)", refresh))
	printWarning("Press Ctrl+C to exit")
	fmt.Println()

	ticker := time.NewTicker(time.Duration(refresh) * time.Second)
	defer ticker.Stop()

	// Initial display
	clearScreen()
	displayLiveStatus(cfg)

	for range ticker.C {
		clearScreen()
		displayLiveStatus(cfg)
	}

	return nil
}

func displayLiveStatus(cfg *config.Config) {
	fmt.Println(bold("╔═══════════════════════════════════════════════════════╗"))
	fmt.Println(bold("║  🥑 Andorian Avocado Exchange - Live Status         ║"))
	fmt.Println(bold("╚═══════════════════════════════════════════════════════╝"))
	fmt.Println()

	fmt.Printf("%s %s\n", cyan("Last Update:"), time.Now().Format("15:04:05"))
	fmt.Printf("%s %s\n\n", cyan("Server:"), fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))

	enabledClients := cfg.GetEnabledClients()

	// Print table header
	fmt.Printf("%-20s %-15s %-18s %-15s %-12s\n",
		bold(cyan("Bot")),
		bold(cyan("Status")),
		bold(cyan("Strategy")),
		bold(cyan("P&L")),
		bold(cyan("Orders")))
	fmt.Println(strings.Repeat("─", 80))

	for _, client := range enabledClients {
		status := green("● READY")
		pnl := "-"
		orders := "-"

		// In real implementation, this would fetch live data from running bots
		// For now, show as configured

		fmt.Printf("%-20s %-15s %-18s %-15s %-12s\n",
			client.Name,
			status,
			client.Strategy,
			pnl,
			orders)
	}

	fmt.Println()
	fmt.Printf("%s Use 'q' + Enter to quit, Ctrl+C to exit\n", cyan("💡"))
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
