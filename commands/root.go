package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cliVersion = "dev" // cliVersion 为编译时注入版本号，默认为 dev

var rootCmd = &cobra.Command{
	Use:   "hy-motion-cli",
	Short: "HY-Motion CLI - 控制运动生成任务",
	Long:  "HY-Motion CLI 用于提交和管理运动生成任务。",
	Run: func(cmd *cobra.Command, args []string) {
		if usage, _ := cmd.Flags().GetBool("usage"); usage {
			printUsage()
			os.Exit(0)
		}
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Println(cliVersion)
			os.Exit(0)
		}
		cmd.Help()
	},
}

func printUsage() {
	usage := map[string]interface{}{
		"program": "hy-motion-cli",
		"description": "HY-Motion CLI - 控制运动生成任务",
		"version": cliVersion,
		"commands": []map[string]string{
			{
				"name":        "health",
				"description": "检查服务健康状态",
				"usage":       "hy-motion-cli health",
				"args":        "无参数",
			},
			{
				"name":        "submit",
				"description": "提交运动生成任务",
				"usage":       "hy-motion-cli submit <text> [-d duration] [-s seeds] [-c cfg-scale] [-o output-format]（不提供 seeds 时自动生成）",
				"args":        "接受 1 个参数: text (要生成运动的文本)",
			},
			{
				"name":        "status",
				"description": "查看任务状态",
				"usage":       "hy-motion-cli status <task_id>",
				"args":        "接受 1 个参数: task_id (任务 ID)",
			},
			{
				"name":        "download",
				"description": "下载任务结果文件",
				"usage":       "hy-motion-cli download <task_id> [-f format] [-o output_path] [-v version]",
				"args":        "接受 1 个参数: task_id (任务 ID)",
			},
			{
				"name":        "queue",
				"description": "显示队列统计",
				"usage":       "hy-motion-cli queue",
				"args":        "无参数",
			},
			{
				"name":        "config",
				"description": "交互式配置 CLI 设置",
				"usage":       "hy-motion-cli config",
				"args":        "无参数（交互式）",
			},
			{
				"name":        "upgrade",
				"description": "升级 CLI 到最新版本",
				"usage":       "hy-motion-cli upgrade",
				"args":        "无参数",
			},
		},
		"flags": []map[string]string{
			{
				"name":        "--version",
				"description": "显示版本号",
			},
			{
				"name":        "--usage",
				"description": "输出机器可读的 usage 信息 (用于 AI 自动发现)",
			},
			{
				"name":        "--help, -h",
				"description": "显示帮助信息",
			},
		},
		"config_file_example": map[string]string{
			"api.url":      "http://localhost:8000",
			"api.timeout":  "30",
			"auth.user_id": "你的用户ID",
			"auth.token":   "你的令牌",
		},
	}

	data, _ := json.MarshalIndent(usage, "", "  ")
	fmt.Println(string(data))
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("usage", "", false, "输出机器可读的 usage 信息 (用于 AI 自动发现)")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().Bool("version", false, "显示版本号")
}
