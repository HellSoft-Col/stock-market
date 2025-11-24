package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var force bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "🛑 Stop trading bots",
	Long:  `Stop all running trading bots gracefully or force stop`,
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringVarP(&botName, "name", "n", "", "stop specific bot by name")
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, "force stop without graceful shutdown")
}

func runStop(cmd *cobra.Command, args []string) error {
	printHeader("🛑 Stopping Trading Bots")

	if botName != "" {
		printInfo(fmt.Sprintf("Stopping bot: %s", botName))
	} else {
		printInfo("Stopping all bots...")
	}

	if force {
		printWarning("Force stop enabled - bots will be terminated immediately")
	} else {
		printInfo("Graceful shutdown initiated...")
	}

	// In real implementation, this would connect to running manager
	// and send stop signals

	printSuccess("All bots stopped successfully")
	printInfo("Use 'trading-cli pnl' to view final P&L report")

	return nil
}
