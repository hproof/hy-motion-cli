package commands

import (
	"fmt"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a motion generation task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := api.NewClient(cfg)

		resp, err := client.SubmitTask(text)
		if err != nil {
			return fmt.Errorf("failed to submit task: %w", err)
		}

		fmt.Printf("Task submitted successfully!\n")
		fmt.Printf("  Task ID: %s\n", resp.TaskID)
		fmt.Printf("  Status:  %s\n", resp.Status)

		return nil
	},
}

func init() {
	submitCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	rootCmd.AddCommand(submitCmd)
}
