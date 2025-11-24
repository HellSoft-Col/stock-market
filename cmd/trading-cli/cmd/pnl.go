package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	detailed   bool
	exportFile string
)

var pnlCmd = &cobra.Command{
	Use:   "pnl",
	Short: "💰 Show P&L report",
	Long:  `Display detailed Profit & Loss report for all bots`,
	RunE:  runPnL,
}

func init() {
	rootCmd.AddCommand(pnlCmd)
	pnlCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "show detailed breakdown")
	pnlCmd.Flags().StringVarP(&exportFile, "export", "e", "", "export to CSV file")
}

func runPnL(cmd *cobra.Command, args []string) error {
	printHeader("💰 Profit & Loss Report")

	// Sample data - in real implementation, this would fetch from running/stopped bots
	botData := []BotPnL{
		{
			Name:           "auto-producer-1",
			Strategy:       "auto_producer",
			InitialBalance: 10000,
			CurrentBalance: 10125.50,
			PnL:            125.50,
			OrdersSent:     45,
			OrdersFilled:   42,
			Productions:    30,
			Uptime:         "2h 15m",
		},
		{
			Name:           "market-maker-1",
			Strategy:       "market_maker",
			InitialBalance: 10000,
			CurrentBalance: 10087.25,
			PnL:            87.25,
			OrdersSent:     234,
			OrdersFilled:   210,
			Productions:    0,
			Uptime:         "2h 15m",
		},
		{
			Name:           "momentum-trader-1",
			Strategy:       "momentum_trader",
			InitialBalance: 10000,
			CurrentBalance: 10045.80,
			PnL:            45.80,
			OrdersSent:     12,
			OrdersFilled:   11,
			Productions:    0,
			Uptime:         "2h 15m",
		},
	}

	if exportFile != "" {
		return exportPnLToCSV(botData, exportFile)
	}

	displayPnLTable(botData, detailed)
	displayPnLSummary(botData)

	return nil
}

type BotPnL struct {
	Name           string
	Strategy       string
	InitialBalance float64
	CurrentBalance float64
	PnL            float64
	OrdersSent     int
	OrdersFilled   int
	Productions    int
	Uptime         string
}

func displayPnLTable(data []BotPnL, detailed bool) {
	fmt.Println()

	if detailed {
		// Detailed table header
		fmt.Printf("%-20s %-16s %-12s %-12s %-15s %-10s %-8s %-8s %-6s %-10s\n",
			bold(yellow("Bot Name")),
			bold(yellow("Strategy")),
			bold(yellow("Initial")),
			bold(yellow("Current")),
			bold(yellow("P&L")),
			bold(yellow("P&L %")),
			bold(yellow("Orders")),
			bold(yellow("Fills")),
			bold(yellow("Prod")),
			bold(yellow("Uptime")))
		fmt.Println(strings.Repeat("─", 120))

		for _, bot := range data {
			pnlPercent := (bot.PnL / bot.InitialBalance) * 100

			pnlStr := fmt.Sprintf("$%.2f", bot.PnL)
			pnlPercentStr := fmt.Sprintf("%.2f%%", pnlPercent)

			if bot.PnL > 0 {
				pnlStr = green(pnlStr + " ↑")
			} else if bot.PnL < 0 {
				pnlStr = red(pnlStr + " ↓")
			}

			fmt.Printf("%-20s %-16s %-12s %-12s %-20s %-10s %-8d %-8d %-6d %-10s\n",
				bot.Name,
				bot.Strategy,
				fmt.Sprintf("$%.2f", bot.InitialBalance),
				fmt.Sprintf("$%.2f", bot.CurrentBalance),
				pnlStr,
				pnlPercentStr,
				bot.OrdersSent,
				bot.OrdersFilled,
				bot.Productions,
				bot.Uptime)
		}
	} else {
		// Simple table header
		fmt.Printf("%-20s %-18s %-20s %-12s %-12s\n",
			bold(yellow("Bot Name")),
			bold(yellow("Strategy")),
			bold(yellow("P&L")),
			bold(yellow("P&L %")),
			bold(yellow("Fill Rate")))
		fmt.Println(strings.Repeat("─", 80))

		for _, bot := range data {
			pnlPercent := (bot.PnL / bot.InitialBalance) * 100
			fillRate := float64(bot.OrdersFilled) / float64(bot.OrdersSent) * 100

			pnlStr := fmt.Sprintf("$%.2f", bot.PnL)
			pnlPercentStr := fmt.Sprintf("%.2f%%", pnlPercent)

			if bot.PnL > 0 {
				pnlStr = green(pnlStr + " ↑")
			} else if bot.PnL < 0 {
				pnlStr = red(pnlStr + " ↓")
			}

			fmt.Printf("%-20s %-18s %-25s %-12s %-12s\n",
				bot.Name,
				bot.Strategy,
				pnlStr,
				pnlPercentStr,
				fmt.Sprintf("%.1f%%", fillRate))
		}
	}

	fmt.Println()
}

func displayPnLSummary(data []BotPnL) {
	totalPnL := 0.0
	totalInitial := 0.0
	totalOrders := 0
	totalFills := 0
	winners := 0
	losers := 0

	for _, bot := range data {
		totalPnL += bot.PnL
		totalInitial += bot.InitialBalance
		totalOrders += bot.OrdersSent
		totalFills += bot.OrdersFilled

		if bot.PnL > 0 {
			winners++
		} else if bot.PnL < 0 {
			losers++
		}
	}

	totalPnLPercent := (totalPnL / totalInitial) * 100
	totalFillRate := float64(totalFills) / float64(totalOrders) * 100

	printHeader("📈 Summary")

	pnlColor := green
	pnlEmoji := "🚀"
	if totalPnL < 0 {
		pnlColor = red
		pnlEmoji = "📉"
	}

	fmt.Printf("  %s Total P&L:          %s (%s)\n",
		pnlEmoji,
		pnlColor(bold(fmt.Sprintf("$%.2f", totalPnL))),
		pnlColor(fmt.Sprintf("%.2f%%", totalPnLPercent)))

	fmt.Printf("  %s Total Capital:      %s\n",
		"💵",
		cyan(fmt.Sprintf("$%.2f", totalInitial)))

	fmt.Printf("  %s Winning Bots:       %s\n",
		green("✅"),
		green(fmt.Sprintf("%d", winners)))

	if losers > 0 {
		fmt.Printf("  %s Losing Bots:        %s\n",
			red("❌"),
			red(fmt.Sprintf("%d", losers)))
	}

	fmt.Printf("  %s Total Orders:       %s\n",
		"📊",
		fmt.Sprintf("%d", totalOrders))

	fmt.Printf("  %s Filled Orders:      %s\n",
		"✓",
		fmt.Sprintf("%d", totalFills))

	fmt.Printf("  %s Overall Fill Rate:  %s\n",
		"⚡",
		yellow(fmt.Sprintf("%.1f%%", totalFillRate)))

	fmt.Println()

	// Recommendations
	if totalPnL > 0 {
		printSuccess(fmt.Sprintf("Great job! Your bots are profitable. Keep monitoring performance."))
	} else {
		printWarning("Your bots are currently unprofitable. Consider adjusting strategies.")
	}

	if totalFillRate < 50 {
		printWarning("Low fill rate detected. Check order prices and market liquidity.")
	}
}

func exportPnLToCSV(data []BotPnL, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		printError(fmt.Sprintf("Failed to create CSV file: %v", err))
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Timestamp",
		"Bot Name",
		"Strategy",
		"Initial Balance",
		"Current Balance",
		"P&L",
		"P&L %",
		"Orders Sent",
		"Orders Filled",
		"Fill Rate %",
		"Productions",
		"Uptime",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	for _, bot := range data {
		pnlPercent := (bot.PnL / bot.InitialBalance) * 100
		fillRate := float64(bot.OrdersFilled) / float64(bot.OrdersSent) * 100

		row := []string{
			timestamp,
			bot.Name,
			bot.Strategy,
			fmt.Sprintf("%.2f", bot.InitialBalance),
			fmt.Sprintf("%.2f", bot.CurrentBalance),
			fmt.Sprintf("%.2f", bot.PnL),
			fmt.Sprintf("%.2f", pnlPercent),
			fmt.Sprintf("%d", bot.OrdersSent),
			fmt.Sprintf("%d", bot.OrdersFilled),
			fmt.Sprintf("%.2f", fillRate),
			fmt.Sprintf("%d", bot.Productions),
			bot.Uptime,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	printSuccess(fmt.Sprintf("P&L report exported to: %s", filename))
	return nil
}
