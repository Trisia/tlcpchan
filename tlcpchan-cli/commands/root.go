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
	cliVersion = version
	flag.StringVar(&apiURL, "api", "http://localhost:20080", "API服务地址")
	flag.StringVar(&apiURL, "a", "http://localhost:20080", "API服务地址(缩写)")
	flag.StringVar(&output, "output", "table", "输出格式 (table|json)")
	flag.StringVar(&output, "o", "table", "输出格式(缩写) (table|json)")
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
				"list":    {Name: "list", Description: "列出所有实例", Usage: "list", Run: instanceList},
				"show":    {Name: "show", Description: "显示实例详情", Usage: "show <name>", Run: instanceShow},
				"create":  {Name: "create", Description: "创建实例", Usage: "create [-f file]", Run: instanceCreate},
				"update":  {Name: "update", Description: "更新实例配置", Usage: "update <name> [-f file]", Run: instanceUpdate},
				"delete":  {Name: "delete", Description: "删除实例", Usage: "delete <name>", Run: instanceDelete},
				"start":   {Name: "start", Description: "启动实例", Usage: "start <name>", Run: instanceStart},
				"stop":    {Name: "stop", Description: "停止实例", Usage: "stop <name>", Run: instanceStop},
				"reload":  {Name: "reload", Description: "重载实例", Usage: "reload <name>", Run: instanceReload},
				"restart": {Name: "restart", Description: "重启实例", Usage: "restart <name>", Run: instanceRestart},
				"stats":   {Name: "stats", Description: "查看统计信息", Usage: "stats <name>", Run: instanceStats},
				"logs":    {Name: "logs", Description: "查看日志", Usage: "logs <name>", Run: instanceLogs},
				"health":  {Name: "health", Description: "健康检查", Usage: "health <name> [-t timeout]", Run: instanceHealth},
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
		"keystore": {
			Name:        "keystore",
			Description: "密钥管理",
			Usage:       "keystore <子命令>",
			SubCommands: map[string]Command{
				"list":       {Name: "list", Description: "列出所有 keystore", Usage: "list", Run: keyStoreList},
				"show":       {Name: "show", Description: "显示 keystore 信息", Usage: "show <name>", Run: keyStoreShow},
				"detail":     {Name: "detail", Description: "显示 keystore 详情（含关联实例）", Usage: "detail <name>", Run: keyStoreShowDetail},
				"update":     {Name: "update", Description: "更新 keystore 参数", Usage: "update <name> [选项]", Run: keyStoreUpdateParams},
				"create":     {Name: "create", Description: "创建 keystore", Usage: "create [选项]", Run: keyStoreCreate},
				"generate":   {Name: "generate", Description: "生成 keystore（含证书）", Usage: "generate [选项]", Run: keyStoreGenerate},
				"export-csr": {Name: "export-csr", Description: "导出证书请求(CSR)", Usage: "export-csr <name> [选项]", Run: keyStoreExportCSR},
				"delete":     {Name: "delete", Description: "删除 keystore", Usage: "delete <name>", Run: keyStoreDelete},
			},
		},
		"rootcert": {
			Name:        "rootcert",
			Description: "根证书管理",
			Usage:       "rootcert <子命令>",
			SubCommands: map[string]Command{
				"list":     {Name: "list", Description: "列出所有根证书", Usage: "list", Run: rootCertList},
				"download": {Name: "download", Description: "下载根证书文件", Usage: "download <filename> [-o output]", Run: rootCertDownload},
				"add":      {Name: "add", Description: "添加根证书", Usage: "add [选项]", Run: rootCertAdd},
				"generate": {Name: "generate", Description: "生成根 CA 证书", Usage: "generate [选项]", Run: rootCertGenerate},
				"delete":   {Name: "delete", Description: "删除根证书", Usage: "delete <filename>", Run: rootCertDelete},
				"reload":   {Name: "reload", Description: "重载所有根证书", Usage: "reload", Run: rootCertReload},
			},
		},
		"system": {
			Name:        "system",
			Description: "系统信息",
			Usage:       "system <子命令>",
			SubCommands: map[string]Command{
				"info":   {Name: "info", Description: "显示系统信息", Usage: "info", Run: systemInfo},
				"health": {Name: "health", Description: "健康检查", Usage: "health", Run: systemHealth},
				"logs": {Name: "logs", Description: "日志管理", Usage: "logs <子命令>", SubCommands: map[string]Command{
					"list":         {Name: "list", Description: "列出日志文件", Usage: "list", Run: systemLogsList},
					"content":      {Name: "content", Description: "读取日志内容", Usage: "content [选项]", Run: systemLogsContent},
					"download":     {Name: "download", Description: "下载单个日志文件", Usage: "download <filename> [选项]", Run: systemLogsDownload},
					"download-all": {Name: "download-all", Description: "打包下载所有日志", Usage: "download-all [选项]", Run: systemLogsDownloadAll},
				}},
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
