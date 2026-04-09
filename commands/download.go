package commands

import (
	"fmt"
	"path/filepath"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var format string
var outputPath string
var version int

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "下载任务结果文件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		client := api.NewClient(cfg)

		// 如果未指定输出路径，生成默认文件名（带序号后缀）
		if outputPath == "" {
			ext := ".fbx"
			if format == "dict" {
				ext = ".html"
			}
			// 生成带序号的文件名，如 task_id_000.fbx
			suffix := fmt.Sprintf("_%03d", version)
			outputPath = filepath.Join(".", taskID+suffix+ext)
		}

		if err := client.DownloadTask(taskID, format, outputPath, version); err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}

		fmt.Printf("文件已保存: %s\n", outputPath)
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&format, "format", "f", "fbx", "下载格式: fbx 或 dict")
	downloadCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件路径 (默认: {task_id}_{version}.fbx)")
	downloadCmd.Flags().IntVarP(&version, "version", "v", 0, "FBX 版本号 (0 ~ 3)，对应不同随机种子生成的结果")
	rootCmd.AddCommand(downloadCmd)
}
