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
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", baseURL, owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("请求失败: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	// 去掉 v 前缀
	tagName = strings.TrimPrefix(release.TagName, "v")

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

	// 查找解压后的 exe 文件
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return err
	}
	var newPath string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".exe") {
			newPath = filepath.Join(tmpDir, entry.Name())
			break
		}
	}

	if newPath == "" {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("解压后未找到 exe 文件")
	}

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
		os.RemoveAll(tmpDir)
		return fmt.Errorf("创建安装脚本失败: %w", err)
	}

	// 启动脚本并退出
	cmd := exec.Command("cmd", "/c", "start", "/B", scriptPath)
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		os.Remove(scriptPath)
		return fmt.Errorf("启动安装脚本失败: %w", err)
	}

	fmt.Println("正在升级，请稍候...")
	os.Exit(0)
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
