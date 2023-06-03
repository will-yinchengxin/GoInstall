package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/alecthomas/kingpin"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	goEnvPath     = "/etc/profile"
	goDownloadURL = "https://studygolang.com/dl/golang"
)

var (
	version                                       string
	goPath, goRoot, goBinPath, goInstallationPath string
)

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o GoDownload main.go
func main() {
	var (
		goVersion  = kingpin.Flag("gv", "go版本号：1.20 ｜ 1.19 ｜ 1.18 ｜ 1.17 ｜ 1.16").Default("1.20").String()
		goRootPath = kingpin.Flag("gr", "go root 目录，如：/usr/local").Default("/usr/local").String()
	)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.Version("1.0.0")
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()
	version = *goVersion
	goPath = *goRootPath + "/go/path"
	goRoot = *goRootPath + "/go"
	goBinPath = *goRootPath + "/go/bin"
	goInstallationPath = *goRootPath + "/go"

	DownLoad()
}

func DownLoad() {
	// 获取操作系统信息
	goOS := runtime.GOOS
	// 获取体系结构信息
	goArch := runtime.GOARCH

	downloadURL := fmt.Sprintf("%s/go%s.%s-%s.tar.gz", goDownloadURL, version, goOS, goArch)
	err := downloadFile(downloadURL, fmt.Sprintf("go%s.%s-%s.tar.gz", version, goOS, goArch))
	if err != nil {
		fmt.Printf("下载失败: %s\n", err.Error())
		return
	}

	err = extractTarGz(fmt.Sprintf("go%s.%s-%s.tar.gz", version, goOS, goArch), "./")
	if err != nil {
		// open go/src/crypto/sha1/fallback_test.go: too many open files
		// ulimit -u 查看限制
		// ulimit -n 数字， 修改限制
		fmt.Printf("解压失败: %s\n", err.Error())
		fmt.Println("如果错误为：open go/src/crypto/sha1/fallback_test.go: too many open files 请执行 「 ulimit -n 65535 」")
		return
	}

	err = moveGoInstallation("./go", goInstallationPath)
	if err != nil {
		fmt.Printf("移动Go安装文件失败: %s\n", err.Error())
		return
	}

	err = createDirectory(goPath)
	if err != nil {
		fmt.Printf("创建目录失败: %s\n", err.Error())
		return
	}

	err = updateProfile()
	if err != nil {
		fmt.Printf("更新环境变量失败: %s\n", err.Error())
		fmt.Println("")
		return
	}

	err = sourceProfile()
	if err != nil {
		fmt.Printf("执行source /etc/profile失败: %s\n", err.Error())
		return
	}

	fmt.Println("配置环境变量成功")

	err = configureGoEnv()
	if err != nil {
		fmt.Printf("配置Go环境变量失败: %s\n", err.Error())
		return
	}

	fmt.Printf("Go版本 %s 下载、安装、配置 成功 ", version)

	err = updateAndSourceProfile()
	if err != nil {
		fmt.Printf("\n 执行脚本失败: %s, 请手动执行 「 source /etc/profile 」\n", err.Error())
		return
	}

	fmt.Println("大功告成 ！！！")
}

func updateAndSourceProfile() error {
	script := `#!/bin/bash
source /etc/profile
`
	scriptPath := "/tmp/updateenv.sh"

	err := os.WriteFile(scriptPath, []byte(script), 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", scriptPath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func downloadFile(url, filepath string) error {
	outputFile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return err
	}
	defer outputFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return err
	}
	fmt.Println("ZIP 文件下载完成")
	return nil
}

func extractTarGz(tarGzFile, destPath string) error {
	file, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		filename := header.Name
		if strings.HasPrefix(filename, "go/") {
			filename = filepath.Join(destPath, filename)
			mode := header.FileInfo().Mode()

			switch header.Typeflag {
			case tar.TypeDir:
				err := os.MkdirAll(filename, mode)
				if err != nil {
					return err
				}

			case tar.TypeReg:
				file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, mode)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(file, tarReader)
				if err != nil {
					return err
				}
			}
		}
	}
	fmt.Println("解压 ZIP 文件成功")
	return nil
}

func moveGoInstallation(sourcePath, destinationPath string) error {
	err := os.RemoveAll(destinationPath)
	if err != nil {
		return err
	}

	return os.Rename(sourcePath, destinationPath)
}

func createDirectory(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func updateProfile() error {
	profileContent, err := os.ReadFile(goEnvPath)
	if err != nil {
		return err
	}

	content := string(profileContent)
	if !strings.Contains(content, goPath) {
		content += fmt.Sprintf("\nexport PATH=$PATH:%s\n", goBinPath)
		content += fmt.Sprintf("export GOPATH=%s\n", goPath)
	}

	err = os.WriteFile(goEnvPath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func sourceProfile() error {
	// source命令是bash shell内置命令，exec.Command函数在默认情况下只能执行外部命令，无法执行shell内置命令。
	// 需要执行的时候添加 bash -c

	// 在子进程中执行source命令后，环境变量只在子进程中生效，而不会传递回父进程中, 导致这里内容没有生效
	// 为了使环境变量在当前会话中生效，可以使用os/exec包提供的Command结构的Run方法，该方法会在一个子进程中执行命令
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && env", goEnvPath))
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 使用os.Setenv将环境变量设置到当前进程中
			os.Setenv(key, value)
		}
	}

	// 这种方法仍然无法影响到父进程的环境变量。如果需要在当前会话的父进程中设置环境变量，还要在终端中手动执行命令或重启终端会话。
	return nil
}

func configureGoEnv() error {
	cmds := []string{
		// 执行命令之前一定要先报漏 GOROOT（非 bin 目录）
		fmt.Sprintf("export GOROOT=%s", goRoot),
		"go env -w GO111MODULE=on",
		fmt.Sprintf("go env -w GOPATH=%s", goPath),
		"go env -w GOPROXY=https://goproxy.cn,https://goproxy.io,direct",
	}

	for _, cmdStr := range cmds {
		cmd := exec.Command("sh", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
