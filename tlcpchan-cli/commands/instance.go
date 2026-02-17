package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Trisia/tlcpchan-cli/client"
)

func instanceList(args []string) error {
	instances, err := cli.ListInstances()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(instances)
	}

	if len(instances) == 0 {
		fmt.Println("无实例")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t状态\t类型\t监听\t目标\t启用")
	for _, inst := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%v\n",
			inst.Name, inst.Status, inst.Config.Type, inst.Config.Listen, inst.Config.Target, inst.Enabled)
	}
	w.Flush()
	return nil
}

func instanceShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	inst, err := cli.GetInstance(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(inst)
	}

	fmt.Printf("名称: %s\n", inst.Name)
	fmt.Printf("状态: %s\n", inst.Status)
	fmt.Printf("类型: %s\n", inst.Config.Type)
	fmt.Printf("监听: %s\n", inst.Config.Listen)
	fmt.Printf("目标: %s\n", inst.Config.Target)
	fmt.Printf("协议: %s\n", inst.Config.Protocol)
	fmt.Printf("认证: %s\n", inst.Config.Auth)
	fmt.Printf("启用: %v\n", inst.Enabled)
	return nil
}

func instanceCreate(args []string) error {
	fs := flagSet("create")
	file := fs.String("file", "", "配置文件路径(JSON)")
	fs.StringVar(file, "f", "", "配置文件路径(JSON)(缩写)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var cfg client.InstanceConfig
	if *file != "" {
		data, err := os.ReadFile(*file)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("解析配置失败: %w", err)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			decoder := json.NewDecoder(os.Stdin)
			if err := decoder.Decode(&cfg); err != nil {
				return fmt.Errorf("解析标准输入失败: %w", err)
			}
		} else {
			return fmt.Errorf("请使用 -f 指定配置文件或通过标准输入提供JSON配置")
		}
	}

	if err := cli.CreateInstance(&cfg); err != nil {
		return err
	}
	fmt.Printf("实例 %s 创建成功\n", cfg.Name)
	return nil
}

func instanceUpdate(args []string) error {
	fs := flagSet("update")
	file := fs.String("file", "", "配置文件路径(JSON)")
	fs.StringVar(file, "f", "", "配置文件路径(JSON)(缩写)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("请指定实例名称")
	}
	name := remaining[0]

	var cfg client.InstanceConfig
	if *file != "" {
		data, err := os.ReadFile(*file)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("解析配置失败: %w", err)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			decoder := json.NewDecoder(os.Stdin)
			if err := decoder.Decode(&cfg); err != nil {
				return fmt.Errorf("解析标准输入失败: %w", err)
			}
		} else {
			return fmt.Errorf("请使用 -f 指定配置文件或通过标准输入提供JSON配置")
		}
	}

	cfg.Name = name
	if err := cli.UpdateInstance(name, &cfg); err != nil {
		return err
	}
	fmt.Printf("实例 %s 更新成功\n", name)
	return nil
}

func instanceDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	if err := cli.DeleteInstance(args[0]); err != nil {
		return err
	}
	fmt.Printf("实例 %s 已删除\n", args[0])
	return nil
}

func instanceStart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	if err := cli.StartInstance(args[0]); err != nil {
		return err
	}
	fmt.Printf("实例 %s 已启动\n", args[0])
	return nil
}

func instanceStop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	if err := cli.StopInstance(args[0]); err != nil {
		return err
	}
	fmt.Printf("实例 %s 已停止\n", args[0])
	return nil
}

func instanceReload(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	if err := cli.ReloadInstance(args[0]); err != nil {
		return err
	}
	fmt.Printf("实例 %s 已重载\n", args[0])
	return nil
}

func instanceRestart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	if err := cli.RestartInstance(args[0]); err != nil {
		return err
	}
	fmt.Printf("实例 %s 已重启\n", args[0])
	return nil
}

func instanceStats(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	stats, err := cli.InstanceStats(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(stats)
	}

	for k, v := range stats {
		fmt.Printf("%s: %v\n", k, v)
	}
	return nil
}

func instanceLogs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	logs, err := cli.InstanceLogs(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(logs)
	}

	for _, log := range logs {
		fmt.Printf("[%s] %s: %s\n", log["level"], log["timestamp"], log["message"])
	}
	return nil
}

func instanceHealth(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	fs := flagSet("health")
	timeout := fs.Int("timeout", 0, "超时时间（秒）")
	fs.IntVar(timeout, "t", 0, "超时时间（秒）(缩写)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("请指定实例名称")
	}

	name := remaining[0]

	var timeoutPtr *int
	if *timeout > 0 {
		timeoutPtr = timeout
	}

	health, err := cli.InstanceHealth(name, timeoutPtr)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(health)
	}

	fmt.Printf("实例: %s\n\n", health.Instance)
	for _, result := range health.Results {
		fmt.Printf("协议: %s\n", result.Protocol)
		if result.Success {
			fmt.Printf("  状态: 成功\n")
			fmt.Printf("  延迟: %dms\n", result.Latency)
		} else {
			fmt.Printf("  状态: 失败\n")
			fmt.Printf("  错误: %s\n", result.Error)
		}
		fmt.Println()
	}
	return nil
}
