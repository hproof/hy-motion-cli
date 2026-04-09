package commands

import (
	"fmt"
	"strconv"
	"strings"

	"hy-motion-cli/api"
	"hy-motion-cli/config"

	"github.com/spf13/cobra"
)

var duration float64
var seeds string
var cfgScale float64
var outputFormat string

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "提交运动生成任务",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]

		// 解析 seeds 字符串为 []int
		var seedsList []int
		if seeds != "" {
			for _, s := range strings.Split(seeds, ",") {
				if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
					seedsList = append(seedsList, v)
				}
			}
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		client := api.NewClient(cfg)

		resp, err := client.SubmitTask(text, duration, seedsList, cfgScale, outputFormat)
		if err != nil {
			return fmt.Errorf("提交任务失败: %w", err)
		}

		fmt.Printf("任务提交成功！\n")
		fmt.Printf("  任务 ID: %s\n", resp.TaskID)
		fmt.Printf("  状态:    %s\n", resp.Status)

		return nil
	},
}

func init() {
	submitCmd.Flags().Float64VarP(&duration, "duration", "d", 5.0, "动作时长（秒），建议 0.5-12 秒")
	submitCmd.Flags().StringVarP(&seeds, "seeds", "s", "", "随机种子列表，逗号分隔（不提供则自动生成）")
	submitCmd.Flags().Float64VarP(&cfgScale, "cfg-scale", "c", 5.0, "CFG 引导强度 (1.0-20.0)")
	submitCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "fbx", "输出格式: fbx 或 dict")
	rootCmd.AddCommand(submitCmd)
}
