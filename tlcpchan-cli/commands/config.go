package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

func configShow(args []string) error {
	cfg, err := cli.GetConfig()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(cfg)
	}

	encoder := jsonEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func configReload(args []string) error {
	if err := cli.ReloadConfig(); err != nil {
		return err
	}
	fmt.Println("配置已重新加载")
	return nil
}

func configValidate(args []string) error {
	fs := flagSet("validate")
	file := fs.String("f", "", "配置文件路径(YAML)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("请使用 -f 指定配置文件")
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	fmt.Printf("配置文件 %s 格式有效\n", *file)
	return nil
}
