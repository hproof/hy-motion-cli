package commands

import (
	"fmt"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Show queue statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := api.NewClient(cfg)

		queue, err := client.GetQueue()
		if err != nil {
			return fmt.Errorf("failed to get queue: %w", err)
		}

		fmt.Printf("Queue Status:\n")
		fmt.Printf("  Pending: %d\n", queue.Pending)
		fmt.Printf("  Running: %d\n", queue.Running)

		return nil
	},
}

func init() {
	queueCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	rootCmd.AddCommand(queueCmd)
}
