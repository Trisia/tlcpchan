package commands

import (
	"fmt"
	"strings"
)

func configShow(args []string) error {
	cfg, err := cli.GetConfig()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(cfg)
	}

	fmt.Println("当前配置:")
	printMap(cfg, 0)
	return nil
}

func printMap(m map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", prefix, k)
			printMap(val, indent+1)
		case []interface{}:
			fmt.Printf("%s%s:\n", prefix, k)
			for i, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					fmt.Printf("%s  - 项 %d:\n", prefix, i+1)
					printMap(itemMap, indent+2)
				} else {
					fmt.Printf("%s  - %v\n", prefix, item)
				}
			}
		default:
			fmt.Printf("%s%s: %v\n", prefix, k, val)
		}
	}
}

func configReload(args []string) error {
	_, err := cli.ReloadConfig()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "配置已重新加载",
		})
	}

	fmt.Println("配置已重新加载")
	return nil
}

func configValidate(args []string) error {
	fs := flagSet("validate")
	file := fs.String("file", "", "配置文件路径(YAML)，可选，不提供则使用默认配置文件")
	fs.StringVar(file, "f", "", "配置文件路径(YAML)(缩写)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req := make(map[string]string)
	if *file != "" {
		req["path"] = *file
	}

	if err := cli.ValidateConfig(req); err != nil {
		return err
	}

	displayFile := *file
	if displayFile == "" {
		displayFile = "默认配置文件"
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "配置文件格式有效",
			"file":    displayFile,
		})
	}

	fmt.Printf("配置文件 %s 格式有效\n", displayFile)
	return nil
}
