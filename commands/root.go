package commands

var (
	configPath string
	rootCmd    = &cobra.Command{
		Use:   "hy-motion",
		Short: "HY-Motion CLI - Control motion generation tasks",
		Long: `HY-Motion CLI allows you to submit and manage motion generation tasks.

Examples:
  hy-motion submit "hello world"
  hy-motion status <task_id>
  hy-motion queue`,
	},
)

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	rootCmd.Execute()
}
