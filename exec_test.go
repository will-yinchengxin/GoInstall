package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"testing"
)

func TestLookPath(t *testing.T) {
	// 在环境变量PATH指定的目录中搜索可执行文件，如file中有斜杠，则只在当前目录搜索。返回完整路径或者相对于当前目录的一个相对路径。
	path, err := exec.LookPath("./GoDownload")
	t.Log(path, err)
}

func TestStart(t *testing.T) {
	// Start开始执行c包含的命令，但并不会等待该命令完成即返回
	cmd := exec.Command("sleep", "5")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
}

func TestRun(t *testing.T) {
	// Run执行c包含的命令，并阻塞直到完成。
	cmd := exec.Command("sleep", "5")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Command finished ")
}

func TestOutPut(t *testing.T) {
	output, err := exec.Command("sh", "-c", "ls -la | grep go").Output()
	if err != nil {
		log.Fatal(err.Error())
	}
	t.Log(string(output))
}

func TestLSCommand(t *testing.T) {
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
			log.Fatal(err)
		}
	}
	log.Println("Success !!! ")
}

func TestQuote(t *testing.T) {
	t.Log(strconv.Quote("will\""))
}
