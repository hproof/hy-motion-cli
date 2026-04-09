package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置 CLI 设置",
	Long:  "交互式配置 CLI，包括 API 地址、认证信息等。",
	Run:   runConfig,
}

const configFileName = ".hy-motion-cli.toml"

func runConfig(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// 尝试加载现有配置
	existingCfg := loadExistingConfig()

	// 1. 询问保存位置
	saveLocation := askSaveLocation(reader, existingCfg)
	configPath := getConfigPath(saveLocation)

	// 2. 依次询问各选项
	apiURL := askAPIURL(reader, existingCfg)
	apiTimeout := askAPITimeout(reader, existingCfg)
	userID := askUserID(reader, existingCfg)
	token := askToken(reader, existingCfg)

	// 3. 生成配置内容
	configContent := generateConfigContent(apiURL, apiTimeout, userID, token)

	// 4. 写入文件
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		fmt.Printf("错误: 写入配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n配置已保存到: %s\n", configPath)
}

func loadExistingConfig() map[string]string {
	cfg := make(map[string]string)

	// 尝试加载当前目录配置
	cwd, _ := os.Getwd()
	candidate := filepath.Join(cwd, configFileName)
	if data, err := os.ReadFile(candidate); err == nil {
		parseConfig(string(data), cfg)
		return cfg
	}

	// 尝试加载 home 目录配置
	home, _ := os.UserHomeDir()
	candidate = filepath.Join(home, configFileName)
	if data, err := os.ReadFile(candidate); err == nil {
		parseConfig(string(data), cfg)
	}

	return cfg
}

func parseConfig(content string, cfg map[string]string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "url") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				url := strings.Trim(strings.Trim(parts[1], " "), `"`)
				// 修复缺少协议前缀的 URL
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				cfg["api.url"] = url
			}
		} else if strings.HasPrefix(line, "timeout") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				cfg["api.timeout"] = strings.Trim(parts[1], " ")
			}
		} else if strings.HasPrefix(line, "user_id") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				cfg["auth.user_id"] = strings.Trim(strings.Trim(parts[1], " "), `"`)
			}
		} else if strings.HasPrefix(line, "token") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				cfg["auth.token"] = strings.Trim(strings.Trim(parts[1], " "), `"`)
			}
		}
	}
}

func askSaveLocation(reader *bufio.Reader, existingCfg map[string]string) string {
	fmt.Println("配置保存位置：")
	fmt.Println("1) 当前目录")
	fmt.Println("2) home 目录")

	defaultChoice := "1"
	if existingCfg["_source"] == "home" {
		defaultChoice = "2"
	}

	fmt.Printf("请选择 [默认: %s]: ", defaultChoice)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		input = defaultChoice
	}

	for input != "1" && input != "2" {
		fmt.Print("请输入 1 或 2: ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			input = defaultChoice
		}
	}

	return input
}

func getConfigPath(location string) string {
	if location == "2" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, configFileName)
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, configFileName)
}

func askAPIURL(reader *bufio.Reader, existingCfg map[string]string) string {
	defaultURL := getString(existingCfg, "api.url", "http://localhost:8000")
	fmt.Printf("\nAPI 地址 (api.url):\n")
	fmt.Printf("请输入 [默认: %s]: ", defaultURL)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultURL
	}
	// 确保 URL 包含协议前缀
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "http://" + input
		fmt.Printf("已自动添加 http:// 前缀\n")
	}
	return input
}

func askAPITimeout(reader *bufio.Reader, existingCfg map[string]string) int {
	defaultTimeout := getInt(existingCfg, "api.timeout", 30)
	fmt.Printf("\nAPI 超时时间（秒）(api.timeout):\n")
	fmt.Printf("请输入 [默认: %d]: ", defaultTimeout)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultTimeout
	}
	var timeout int
	fmt.Sscanf(input, "%d", &timeout)
	if timeout <= 0 {
		return defaultTimeout
	}
	return timeout
}

func askUserID(reader *bufio.Reader, existingCfg map[string]string) string {
	defaultUserID := getString(existingCfg, "auth.user_id", "")
	fmt.Printf("\n用户 ID (auth.user_id):\n")
	fmt.Printf("请输入 [默认: %s]: ", defaultUserID)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultUserID
	}
	return input
}

func askToken(reader *bufio.Reader, existingCfg map[string]string) string {
	defaultToken := getString(existingCfg, "auth.token", "")
	fmt.Printf("\n认证令牌 (auth.token):\n")
	fmt.Printf("请输入 [默认: %s]: ", defaultToken)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultToken
	}
	return input
}

func generateConfigContent(apiURL string, apiTimeout int, userID, token string) string {
	return fmt.Sprintf(`[api]
url = "%s"
timeout = %d

[auth]
user_id = "%s"
token = "%s"
`, apiURL, apiTimeout, userID, token)
}

func getString(cfg map[string]string, key, defaultVal string) string {
	if val, ok := cfg[key]; ok && val != "" {
		return val
	}
	return defaultVal
}

func getInt(cfg map[string]string, key string, defaultVal int) int {
	if val, ok := cfg[key]; ok && val != "" {
		var result int
		fmt.Sscanf(val, "%d", &result)
		if result > 0 {
			return result
		}
	}
	return defaultVal
}

func init() {
	GetRootCmd().AddCommand(configCmd)
}
