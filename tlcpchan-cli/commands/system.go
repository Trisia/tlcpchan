package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func systemInfo(args []string) error {
	info, err := cli.GetSystemInfo()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(info)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "操作系统:\t%s\n", info.OS)
	fmt.Fprintf(w, "架构:\t%s\n", info.Arch)
	fmt.Fprintf(w, "CPU核心数:\t%d\n", info.NumCPU)
	fmt.Fprintf(w, "Goroutine数:\t%d\n", info.NumGoroutine)
	fmt.Fprintf(w, "内存分配:\t%d MB\n", info.MemAllocMB)
	fmt.Fprintf(w, "内存总量:\t%d MB\n", info.MemTotalMB)
	fmt.Fprintf(w, "系统内存:\t%d MB\n", info.MemSysMB)
	fmt.Fprintf(w, "启动时间:\t%s\n", info.StartTime)
	fmt.Fprintf(w, "运行时长:\t%s\n", info.Uptime)
	w.Flush()
	return nil
}

func systemHealth(args []string) error {
	health, err := cli.HealthCheck()
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
		return err
	}

	if isJSONOutput() {
		return printJSON(health)
	}

	fmt.Printf("状态: %s\n", health.Status)
	fmt.Printf("版本: %s\n", health.Version)
	return nil
}

func systemShutdown(args []string) error {
	fmt.Println("服务关闭功能暂未实现")
	return nil
}
