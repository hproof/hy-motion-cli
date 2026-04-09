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
	Short: "查看任务状态",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		client := api.NewClient(cfg)

		task, err := client.GetTaskStatus(taskID)
		if err != nil {
			return fmt.Errorf("获取任务状态失败: %w", err)
		}

		fmt.Printf("任务 ID:    %s\n", task.TaskID)
		fmt.Printf("状态:       %s\n", task.Status)
		fmt.Printf("文本:       %s\n", task.Text)
		fmt.Printf("创建时间:   %s\n", task.CreatedAt.Format(time.RFC3339))
		fmt.Printf("更新时间:   %s\n", task.UpdatedAt.Format(time.RFC3339))

		if !task.CompletedAt.IsZero() {
			fmt.Printf("完成时间:   %s\n", task.CompletedAt.Format(time.RFC3339))
		}

		if task.Result != nil && task.Result.FbxFiles != nil && len(task.Result.FbxFiles) > 0 {
			fmt.Printf("文件版本数: %d\n", len(task.Result.FbxFiles))
		}

		if task.Error != "" {
			fmt.Printf("错误:       %s\n", task.Error)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
