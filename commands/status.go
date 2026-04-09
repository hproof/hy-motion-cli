package commands

import (
	"fmt"
	"os"
	"time"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [task_id]",
	Short: "Check task status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := api.NewClient(cfg)

		task, err := client.GetTaskStatus(taskID)
		if err != nil {
			return fmt.Errorf("failed to get task status: %w", err)
		}

		fmt.Printf("Task ID:      %s\n", task.TaskID)
		fmt.Printf("Status:       %s\n", task.Status)
		fmt.Printf("Text:         %s\n", task.Text)
		fmt.Printf("Created At:   %s\n", task.CreatedAt.Format(time.RFC3339))

		if !task.CompletedAt.IsZero() {
			fmt.Printf("Completed At: %s\n", task.CompletedAt.Format(time.RFC3339))
		}

		if task.Error != "" {
			fmt.Printf("Error:        %s\n", task.Error)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	rootCmd.AddCommand(statusCmd)
}
