package commands

import (
	"fmt"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "检查服务健康状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		client := api.NewClient(cfg)

		health, err := client.GetHealth()
		if err != nil {
			return fmt.Errorf("获取健康状态失败: %w", err)
		}

		fmt.Printf("服务状态:   %s\n", health.Status)
		fmt.Printf("GPU 可用:   %v\n", health.GPUAvailable)
		fmt.Printf("模型已加载: %v\n", health.ModelLoaded)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
