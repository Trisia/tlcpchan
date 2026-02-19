package commands

import (
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
	name := fs.String("name", "", "实例名称（必需）")
	instType := fs.String("type", "server", "类型（server/client/http-server/http-client）")
	listen := fs.String("listen", "", "监听地址（必需）")
	target := fs.String("target", "", "目标地址（必需）")
	protocol := fs.String("protocol", "auto", "协议（auto/tlcp/tls）")
	auth := fs.String("auth", "one-way", "认证模式（none/one-way/mutual）")
	keystoreName := fs.String("keystore-name", "", "keystore 名称")
	enabled := fs.Bool("enabled", true, "是否启用")
	sni := fs.String("sni", "", "SNI 名称")
	bufferSize := fs.Int("buffer-size", 0, "缓冲区大小")

	tlcpSignCert := fs.String("tlcp-sign-cert", "", "TLCP 签名证书路径")
	tlcpSignKey := fs.String("tlcp-sign-key", "", "TLCP 签名密钥路径")
	tlcpEncCert := fs.String("tlcp-enc-cert", "", "TLCP 加密证书路径")
	tlcpEncKey := fs.String("tlcp-enc-key", "", "TLCP 加密密钥路径")

	tlsCert := fs.String("tls-cert", "", "TLS 证书路径")
	tlsKey := fs.String("tls-key", "", "TLS 密钥路径")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("请指定 --name")
	}
	if *listen == "" {
		return fmt.Errorf("请指定 --listen")
	}
	if *target == "" {
		return fmt.Errorf("请指定 --target")
	}

	cfg := client.InstanceConfig{
		Name:       *name,
		Type:       *instType,
		Listen:     *listen,
		Target:     *target,
		Protocol:   *protocol,
		Auth:       *auth,
		Enabled:    *enabled,
		SNI:        *sni,
		BufferSize: *bufferSize,
	}

	if *keystoreName != "" {
		if *protocol == "tlcp" || *protocol == "auto" {
			cfg.TLCP = &client.TLCPConfig{
				Auth: *auth,
				Keystore: &client.KeyStoreConfig{
					Name: *keystoreName,
				},
			}
		}
		if *protocol == "tls" || *protocol == "auto" {
			cfg.TLS = &client.TLSConfig{
				Auth: *auth,
				Keystore: &client.KeyStoreConfig{
					Name: *keystoreName,
				},
			}
		}
	} else {
		if (*tlcpSignCert != "" && *tlcpSignKey != "") || (*tlcpEncCert != "" && *tlcpEncKey != "") {
			if *protocol == "tlcp" || *protocol == "auto" {
				params := make(map[string]string)
				if *tlcpSignCert != "" {
					params["sign-cert"] = *tlcpSignCert
				}
				if *tlcpSignKey != "" {
					params["sign-key"] = *tlcpSignKey
				}
				if *tlcpEncCert != "" {
					params["enc-cert"] = *tlcpEncCert
				}
				if *tlcpEncKey != "" {
					params["enc-key"] = *tlcpEncKey
				}
				cfg.TLCP = &client.TLCPConfig{
					Auth: *auth,
					Keystore: &client.KeyStoreConfig{
						Type:   "file",
						Params: params,
					},
				}
			}
		}

		if *tlsCert != "" && *tlsKey != "" {
			if *protocol == "tls" || *protocol == "auto" {
				params := map[string]string{
					"cert": *tlsCert,
					"key":  *tlsKey,
				}
				cfg.TLS = &client.TLSConfig{
					Auth: *auth,
					Keystore: &client.KeyStoreConfig{
						Type:   "file",
						Params: params,
					},
				}
			}
		}
	}

	if err := cli.CreateInstance(&cfg); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例创建成功",
			"name":    cfg.Name,
		})
	}

	fmt.Printf("实例 %s 创建成功\n", cfg.Name)
	return nil
}

func instanceUpdate(args []string) error {
	fs := flagSet("update")
	instType := fs.String("type", "", "类型（server/client/http-server/http-client）")
	listen := fs.String("listen", "", "监听地址")
	target := fs.String("target", "", "目标地址")
	protocol := fs.String("protocol", "", "协议（auto/tlcp/tls）")
	auth := fs.String("auth", "", "认证模式（none/one-way/mutual）")
	keystoreName := fs.String("keystore-name", "", "keystore 名称")
	enabled := fs.Bool("enabled", false, "是否启用")
	sni := fs.String("sni", "", "SNI 名称")
	bufferSize := fs.Int("buffer-size", 0, "缓冲区大小")

	tlcpSignCert := fs.String("tlcp-sign-cert", "", "TLCP 签名证书路径")
	tlcpSignKey := fs.String("tlcp-sign-key", "", "TLCP 签名密钥路径")
	tlcpEncCert := fs.String("tlcp-enc-cert", "", "TLCP 加密证书路径")
	tlcpEncKey := fs.String("tlcp-enc-key", "", "TLCP 加密密钥路径")

	tlsCert := fs.String("tls-cert", "", "TLS 证书路径")
	tlsKey := fs.String("tls-key", "", "TLS 密钥路径")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("请指定实例名称")
	}
	name := remaining[0]

	inst, err := cli.GetInstance(name)
	if err != nil {
		return err
	}

	cfg := inst.Config

	if *instType != "" {
		cfg.Type = *instType
	}
	if *listen != "" {
		cfg.Listen = *listen
	}
	if *target != "" {
		cfg.Target = *target
	}
	if *protocol != "" {
		cfg.Protocol = *protocol
	}
	if *auth != "" {
		cfg.Auth = *auth
	}
	if *sni != "" {
		cfg.SNI = *sni
	}
	if *bufferSize > 0 {
		cfg.BufferSize = *bufferSize
	}
	if *enabled {
		cfg.Enabled = true
	}

	if *keystoreName != "" {
		if cfg.Protocol == "tlcp" || cfg.Protocol == "auto" {
			if cfg.TLCP == nil {
				cfg.TLCP = &client.TLCPConfig{}
			}
			cfg.TLCP.Keystore = &client.KeyStoreConfig{
				Name: *keystoreName,
			}
			if *auth != "" {
				cfg.TLCP.Auth = *auth
			}
		}
		if cfg.Protocol == "tls" || cfg.Protocol == "auto" {
			if cfg.TLS == nil {
				cfg.TLS = &client.TLSConfig{}
			}
			cfg.TLS.Keystore = &client.KeyStoreConfig{
				Name: *keystoreName,
			}
			if *auth != "" {
				cfg.TLS.Auth = *auth
			}
		}
	} else {
		if (*tlcpSignCert != "" && *tlcpSignKey != "") || (*tlcpEncCert != "" && *tlcpEncKey != "") {
			if cfg.Protocol == "tlcp" || cfg.Protocol == "auto" {
				if cfg.TLCP == nil {
					cfg.TLCP = &client.TLCPConfig{}
				}

				params := make(map[string]string)
				if cfg.TLCP.Keystore != nil && cfg.TLCP.Keystore.Params != nil {
					for k, v := range cfg.TLCP.Keystore.Params {
						params[k] = v
					}
				}
				if *tlcpSignCert != "" {
					params["sign-cert"] = *tlcpSignCert
				}
				if *tlcpSignKey != "" {
					params["sign-key"] = *tlcpSignKey
				}
				if *tlcpEncCert != "" {
					params["enc-cert"] = *tlcpEncCert
				}
				if *tlcpEncKey != "" {
					params["enc-key"] = *tlcpEncKey
				}
				cfg.TLCP.Keystore = &client.KeyStoreConfig{
					Type:   "file",
					Params: params,
				}
				if *auth != "" {
					cfg.TLCP.Auth = *auth
				}
			}
		}

		if *tlsCert != "" && *tlsKey != "" {
			if cfg.Protocol == "tls" || cfg.Protocol == "auto" {
				if cfg.TLS == nil {
					cfg.TLS = &client.TLSConfig{}
				}

				params := make(map[string]string)
				if cfg.TLS.Keystore != nil && cfg.TLS.Keystore.Params != nil {
					for k, v := range cfg.TLS.Keystore.Params {
						params[k] = v
					}
				}
				params["cert"] = *tlsCert
				params["key"] = *tlsKey
				cfg.TLS.Keystore = &client.KeyStoreConfig{
					Type:   "file",
					Params: params,
				}
				if *auth != "" {
					cfg.TLS.Auth = *auth
				}
			}
		}
	}

	if err := cli.UpdateInstance(name, &cfg); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例更新成功",
			"name":    name,
		})
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

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例已删除",
			"name":    args[0],
		})
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

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例已启动",
			"name":    args[0],
		})
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

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例已停止",
			"name":    args[0],
		})
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

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例已重载",
			"name":    args[0],
		})
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

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "实例已重启",
			"name":    args[0],
		})
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
