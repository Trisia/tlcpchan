package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Trisia/tlcpchan-cli/client"
)

var (
	apiURL   string
	output   string
	cli      *client.Client
	commands map[string]Command
)

type Command struct {
	Name        string
	Description string
	Usage       string
	Run         func(args []string) error
	SubCommands map[string]Command
}

func Execute(version string) error {
	flag.StringVar(&apiURL, "api", "http://localhost:8080", "API服务地址")
	flag.StringVar(&output, "output", "table", "输出格式 (table|json)")
	flag.Usage = printUsage
	flag.Parse()

	cli = client.NewClient(apiURL)
	commands = getCommands()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("请指定命令")
	}

	cmd, remaining := findCommand(args)
	if cmd.Run == nil {
		if len(remaining) == 0 {
			printCommandHelp(cmd)
			return fmt.Errorf("请指定子命令")
		}
		return fmt.Errorf("未知命令: %s", remaining[0])
	}

	if err := cmd.Run(remaining); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return err
	}
	return nil
}

func getCommands() map[string]Command {
	return map[string]Command{
		"instance": {
			Name:        "instance",
			Description: "实例管理",
			Usage:       "instance <子命令>",
			SubCommands: map[string]Command{
				"list":   {Name: "list", Description: "列出所有实例", Usage: "list", Run: instanceList},
				"show":   {Name: "show", Description: "显示实例详情", Usage: "show <name>", Run: instanceShow},
				"create": {Name: "create", Description: "创建实例", Usage: "create [-f file]", Run: instanceCreate},
				"update": {Name: "update", Description: "更新实例配置", Usage: "update <name> [-f file]", Run: instanceUpdate},
				"delete": {Name: "delete", Description: "删除实例", Usage: "delete <name>", Run: instanceDelete},
				"start":  {Name: "start", Description: "启动实例", Usage: "start <name>", Run: instanceStart},
				"stop":   {Name: "stop", Description: "停止实例", Usage: "stop <name>", Run: instanceStop},
				"reload": {Name: "reload", Description: "重载实例", Usage: "reload <name>", Run: instanceReload},
				"stats":  {Name: "stats", Description: "查看统计信息", Usage: "stats <name>", Run: instanceStats},
				"logs":   {Name: "logs", Description: "查看日志", Usage: "logs <name>", Run: instanceLogs},
			},
		},
		"config": {
			Name:        "config",
			Description: "配置管理",
			Usage:       "config <子命令>",
			SubCommands: map[string]Command{
				"show":     {Name: "show", Description: "显示当前配置", Usage: "show", Run: configShow},
				"reload":   {Name: "reload", Description: "重载配置", Usage: "reload", Run: configReload},
				"validate": {Name: "validate", Description: "验证配置", Usage: "validate [-f file]", Run: configValidate},
			},
		},
		"cert": {
			Name:        "cert",
			Description: "证书管理",
			Usage:       "cert <子命令>",
			SubCommands: map[string]Command{
				"list":     {Name: "list", Description: "列出所有证书", Usage: "list", Run: certList},
				"show":     {Name: "show", Description: "显示证书详情", Usage: "show <name>", Run: certShow},
				"generate": {Name: "generate", Description: "生成证书", Usage: "generate [选项]", Run: certGenerate},
				"reload":   {Name: "reload", Description: "热更新证书", Usage: "reload", Run: certReload},
				"delete":   {Name: "delete", Description: "删除证书", Usage: "delete <name>", Run: certDelete},
			},
		},
		"system": {
			Name:        "system",
			Description: "系统信息",
			Usage:       "system <子命令>",
			SubCommands: map[string]Command{
				"info":     {Name: "info", Description: "显示系统信息", Usage: "info", Run: systemInfo},
				"health":   {Name: "health", Description: "健康检查", Usage: "health", Run: systemHealth},
				"shutdown": {Name: "shutdown", Description: "关闭服务", Usage: "shutdown", Run: systemShutdown},
			},
		},
		"version": {
			Name:        "version",
			Description: "显示版本",
			Usage:       "version",
			Run:         versionCmd,
		},
	}
}

func findCommand(args []string) (Command, []string) {
	cmds := commands
	var cmd Command
	var lastCmd Command

	for i, arg := range args {
		if c, ok := cmds[arg]; ok {
			cmd = c
			lastCmd = c
			if c.SubCommands != nil {
				cmds = c.SubCommands
			} else {
				return cmd, args[i+1:]
			}
		} else {
			if lastCmd.SubCommands != nil {
				return lastCmd, args[i:]
			}
			return cmd, args[i:]
		}
	}
	return lastCmd, []string{}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "用法: tlcpchan-cli [选项] <命令> [参数]\n\n")
	fmt.Fprintf(os.Stderr, "选项:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n命令:\n")
	for _, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Fprintf(os.Stderr, "\n使用 \"tlcpchan-cli <命令> --help\" 获取更多信息\n")
}

func printCommandHelp(cmd Command) {
	fmt.Fprintf(os.Stderr, "用法: tlcpchan-cli %s\n\n", cmd.Usage)
	if cmd.SubCommands != nil {
		fmt.Fprintf(os.Stderr, "子命令:\n")
		for _, sub := range cmd.SubCommands {
			fmt.Fprintf(os.Stderr, "  %-12s %s\n", sub.Name, sub.Description)
		}
	}
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func isJSONOutput() bool {
	return strings.ToLower(output) == "json"
}

func jsonEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: tlcpchan-cli %s [选项]\n\n选项:\n", name)
		fs.PrintDefaults()
	}
	return fs
}
