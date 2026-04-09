package commands

import (
	"fmt"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "显示队列统计",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		client := api.NewClient(cfg)

		queue, err := client.GetQueue()
		if err != nil {
			return fmt.Errorf("获取队列失败: %w", err)
		}

		fmt.Printf("队列状态:\n")
		fmt.Printf("  等待中: %d\n", queue.Pending)
		fmt.Printf("  运行中: %d\n", queue.Running)
		fmt.Printf("  已完成: %d\n", queue.Completed)
		fmt.Printf("  失败:   %d\n", queue.Failed)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(queueCmd)
}
