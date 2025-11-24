package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/manager"
	"github.com/spf13/cobra"
)

var (
	botName string
	stats   bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "🚀 Start trading bots",
	Long:  `Start all configured trading bots or a specific bot by name`,
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&botName, "name", "n", "", "start specific bot by name")
	startCmd.Flags().BoolVarP(&stats, "stats", "s", true, "show statistics periodically")
}

func runStart(cmd *cobra.Command, args []string) error {
	printHeader("🥑 Starting Trading Bots")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	enabledClients := cfg.GetEnabledClients()

	// Filter by name if specified
	if botName != "" {
		found := false
		for _, client := range enabledClients {
			if client.Name == botName {
				found = true
				enabledClients = []config.ClientConfig{client}
				break
			}
		}
		if !found {
			printError(fmt.Sprintf("Bot '%s' not found in configuration", botName))
			return fmt.Errorf("bot not found: %s", botName)
		}
	}

	printInfo(fmt.Sprintf("Configuration loaded: %s", cfgFile))
	printInfo(fmt.Sprintf("Server: %s:%d", cfg.Server.Host, cfg.Server.Port))
	printSuccess(fmt.Sprintf("Found %d enabled bot(s)", len(enabledClients)))

	// Display bots to start
	fmt.Println()
	for i, client := range enabledClients {
		fmt.Printf("  %d. %s %s (%s strategy, %s species)\n",
			i+1,
			green("●"),
			bold(client.Name),
			cyan(client.Strategy),
			yellow(client.Species))
	}
	fmt.Println()

	// Create client manager
	clientManager := manager.NewClientManager(cfg)

	// Start all clients
	printInfo("Starting bot manager...")
	if err := clientManager.Start(); err != nil {
		printError(fmt.Sprintf("Failed to start clients: %v", err))
		return err
	}

	printSuccess("All bots started successfully!")
	printInfo("Press Ctrl+C to stop all bots")

	// Start statistics reporter if enabled
	if stats {
		go statsReporter(clientManager)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh

	printWarning("\nShutdown signal received...")
	printInfo("Stopping all bots gracefully...")

	// Stop all clients
	if err := clientManager.Stop(); err != nil {
		printError(fmt.Sprintf("Error during shutdown: %v", err))
		return err
	}

	printSuccess("All bots stopped successfully")
	printInfo("Final P&L report:")

	// Show final P&L
	showPnLReport(clientManager)

	return nil
}

func statsReporter(cm *manager.ClientManager) {
	// This will be implemented with real-time updates
	// For now, just a placeholder
}

func showPnLReport(cm *manager.ClientManager) {
	stats := cm.GetStats()

	totalPnL := 0.0
	totalOrders := 0
	totalFills := 0

	fmt.Println()
	printHeader("💰 Final P&L Report")

	for _, sessionStats := range stats {
		name := sessionStats["id"].(string)
		connected := sessionStats["connected"].(bool)

		if !connected {
			continue
		}

		if agentStats, ok := sessionStats["agent"].(map[string]interface{}); ok {
			pnl := agentStats["pnl"].(float64)
			balance := agentStats["balance"].(float64)
			ordersSent := agentStats["ordersSent"].(int)
			fillsReceived := agentStats["fillsReceived"].(int)

			totalPnL += pnl
			totalOrders += ordersSent
			totalFills += fillsReceived

			pnlColor := green
			pnlEmoji := "🚀"
			if pnl < 0 {
				pnlColor = red
				pnlEmoji = "📉"
			}

			fmt.Printf("  %s %-20s  P&L: %s  Balance: $%.2f  Orders: %d/%d\n",
				pnlEmoji,
				bold(name),
				pnlColor(fmt.Sprintf("$%.2f", pnl)),
				balance,
				fillsReceived,
				ordersSent)
		}
	}

	fmt.Println()
	totalColor := green
	if totalPnL < 0 {
		totalColor = red
	}
	fmt.Printf("  %s Total P&L: %s\n", bold("💎"), totalColor(bold(fmt.Sprintf("$%.2f", totalPnL))))
	fmt.Printf("  %s Total Orders: %s filled out of %s sent\n",
		"📊",
		cyan(fmt.Sprintf("%d", totalFills)),
		cyan(fmt.Sprintf("%d", totalOrders)))

	if totalOrders > 0 {
		fillRate := float64(totalFills) / float64(totalOrders) * 100
		fmt.Printf("  %s Fill Rate: %s\n", "⚡", yellow(fmt.Sprintf("%.1f%%", fillRate)))
	}
	fmt.Println()
}
