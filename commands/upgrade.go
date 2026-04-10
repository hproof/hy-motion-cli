package commands

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	owner   = "hproof"
	repo    = "hy-motion-cli"
	baseURL = "https://api.github.com"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 CLI 到最新版本",
	Long:  "从 GitHub 下载并安装最新版本的 hy-motion-cli",
	Run:   runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) {
	currentVersion := getCurrentVersion()
	fmt.Printf("当前版本: %s\n", currentVersion)

	latestVersion, downloadURL, err := getLatestRelease()
	if err != nil {
		fmt.Printf("检查更新失败: %v\n", err)
		os.Exit(1)
	}

	if currentVersion == latestVersion {
		fmt.Println("已是最新版本，无需升级")
		return
	}

	fmt.Printf("发现新版本: %s\n", latestVersion)
	fmt.Println("开始下载...")

	tmpFile, err := downloadFile(downloadURL)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	if err := installBinary(tmpFile); err != nil {
		fmt.Printf("安装失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功升级到 %s\n", latestVersion)
}

func getCurrentVersion() string {
	return cliVersion
}

func getLatestRelease() (tagName, downloadURL string, err error) {
	// 通过重定向获取最新 release 的版本号
	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)
	resp, err := http.Head(url)
	if err != nil {
		return "", "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 获取重定向后的 URL
	location := resp.Request.URL.String()

	// 从 URL 中提取版本号，格式如: https://github.com/hproof/hy-motion-cli/releases/tag/v0.4.0
	if idx := strings.LastIndex(location, "/tag/"); idx >= 0 {
		tagName = strings.TrimPrefix(location[idx+len("/tag/"):], "v")
	} else {
		return "", "", fmt.Errorf("无法从 URL 解析版本: %s", location)
	}

	// 构造 release API URL 获取 assets
	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/tags/v%s", baseURL, owner, repo, tagName)
	resp, err = http.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果 API 限速，尝试直接从 GitHub 下载页面解析
		return getLatestReleaseFromPage(tagName)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	// 查找匹配当前平台的二进制文件
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 匹配格式: hy-motion-cli_{version}_{os}_{arch}.tar.gz 或 hy-motion-cli_{os}_{arch}.exe
	expectedPatterns := []string{
		fmt.Sprintf("hy-motion-cli_%s_%s_%s.tar.gz", tagName, osName, arch),
		fmt.Sprintf("hy-motion-cli_%s_%s.tar.gz", osName, arch),
		fmt.Sprintf("hy-motion-cli_%s_%s.exe", osName, arch),
	}

	for _, asset := range release.Assets {
		for _, pattern := range expectedPatterns {
			if asset.Name == pattern {
				return tagName, asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", "", fmt.Errorf("未找到匹配 %s/%s 的二进制文件", osName, arch)
}

// getLatestReleaseFromPage 当 API 限速时，从 GitHub 页面直接解析下载链接
func getLatestReleaseFromPage(tagName string) (string, string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 尝试所有支持的命名格式
	patterns := []string{
		fmt.Sprintf("hy-motion-cli_%s_%s_%s.tar.gz", tagName, osName, arch),
		fmt.Sprintf("hy-motion-cli_%s_%s.tar.gz", osName, arch),
		fmt.Sprintf("hy-motion-cli_%s_%s.exe", osName, arch),
	}

	var downloadURL string
	for _, filename := range patterns {
		url := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s", owner, repo, tagName, filename)
		// 用 HEAD 请求检查文件是否存在
		resp, err := http.Head(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			downloadURL = url
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("未找到匹配 %s/%s 的二进制文件", osName, arch)
	}
	return tagName, downloadURL, nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "hy-motion-cli-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", err
	}
	tmpFile.Close()

	return tmpPath, nil
}

func installBinary(tmpPath string) error {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// 解压 tar.gz
	tmpDir, err := os.MkdirTemp("", "hy-motion-cli-*")
	if err != nil {
		return err
	}

	if err := untar(tmpPath, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("解压失败: %w", err)
	}

	// 查找解压后的二进制文件
	newPath, err := findBinary(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return err
	}

	if runtime.GOOS == "windows" {
		return installWindows(newPath, execPath, tmpDir)
	} else {
		return installUnix(newPath, execPath, tmpDir)
	}
}

// findBinary 查找解压后的正确二进制文件
func findBinary(tmpDir string) (string, error) {
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}

	// 明确的二进制名称
	var expectedName string
	if runtime.GOOS == "windows" {
		expectedName = "hy-motion-cli.exe"
	} else {
		expectedName = "hy-motion-cli"
	}

	// 先精确匹配二进制名称
	for _, entry := range entries {
		if entry.Name() == expectedName {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			// 验证是常规文件
			if info.Mode().IsRegular() {
				return filepath.Join(tmpDir, entry.Name()), nil
			}
		}
	}

	// 回退：Windows 匹配 .exe，Unix 匹配无扩展名的可执行文件
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		name := entry.Name()
		if runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe") {
			return filepath.Join(tmpDir, name), nil
		} else if runtime.GOOS != "windows" && !strings.Contains(name, ".") {
			return filepath.Join(tmpDir, name), nil
		}
	}

	return "", fmt.Errorf("解压后未找到可执行文件")
}

func installWindows(newPath, execPath, tmpDir string) error {
	// 创建安装脚本
	scriptContent := fmt.Sprintf(`@echo off
ping 127.0.0.1 -n 2 >nul
copy /Y "%s" "%s"
if errorlevel 1 (
    ping 127.0.0.1 -n 2 >nul
    copy /Y "%s" "%s"
)
del "%s"
rmdir /S /Q "%s"
del "%%~f0"
`, newPath, execPath, newPath, execPath, newPath, tmpDir)

	scriptPath := filepath.Join(os.TempDir(), "hy-motion-cli-upgrade.bat")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("创建安装脚本失败: %w", err)
	}

	// 启动脚本并退出
	cmd := exec.Command("cmd", "/c", "start", "/B", scriptPath)
	if err := cmd.Start(); err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("启动安装脚本失败: %w", err)
	}

	fmt.Println("正在升级，请稍候...")
	os.Exit(0)
	return nil
}

func installUnix(newPath, execPath, tmpDir string) error {
	// Unix 系统直接替换或原子替换
	if err := os.Chmod(newPath, 0755); err != nil {
		return err
	}

	// 尝试直接 rename
	if err := os.Rename(newPath, execPath); err != nil {
		// 检测是否是跨设备错误（EXDEV）
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err.Error() == "exe cross-device link" {
			// 跨设备：复制到目标目录的临时文件，再原子替换
			if err := atomicReplace(newPath, execPath); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	os.RemoveAll(tmpDir)
	return nil
}

// atomicReplace 原子替换：先复制到目标目录的临时文件，再 rename
func atomicReplace(src, dst string) error {
	// 在目标文件所在目录创建临时文件
	tmpDir := filepath.Dir(dst)
	tmpFile, err := os.CreateTemp(tmpDir, filepath.Base(dst)+".*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()

	// 复制内容
	srcFile, err := os.Open(src)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 确保数据写入磁盘
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("同步文件失败: %w", err)
	}

	tmpFile.Close()

	// 设置正确的权限
	info, err := os.Stat(src)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// 原子替换
	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("替换文件失败: %w", err)
	}

	// 删除源文件
	os.Remove(src)
	return nil
}

func untar(tarPath, dest string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(target, header.FileInfo().Mode())
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 获取源文件权限并设置到目标文件
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}
