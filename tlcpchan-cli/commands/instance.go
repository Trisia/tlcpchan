package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

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
	fmt.Printf("TLCP认证: %s\n", inst.Config.TLCP.ClientAuthType)
	fmt.Printf("TLS认证: %s\n", inst.Config.TLS.ClientAuthType)
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
	tlcpClientAuthType := fs.String("tlcp-client-auth-type", "no-client-cert", "TLCP客户端认证类型")
	tlsClientAuthType := fs.String("tls-client-auth-type", "no-client-cert", "TLS客户端认证类型")
	keystoreName := fs.String("keystore-name", "", "keystore 名称")
	enabled := fs.Bool("enabled", true, "是否启用")
	sni := fs.String("sni", "", "SNI 名称")
	bufferSize := fs.Int("buffer-size", 0, "缓冲区大小")

	tlcpSignCert := fs.String("tlcp-sign-cert", "", "TLCP 签名证书路径")
	tlcpSignKey := fs.String("tlcp-sign-key", "", "TLCP 签名密钥路径")
	tlcpEncCert := fs.String("tlcp-enc-cert", "", "TLCP 加密证书路径")
	tlcpEncKey := fs.String("tlcp-enc-key", "", "TLCP 加密密钥路径")

	tlsSignCert := fs.String("tls-sign-cert", "", "TLS 签名证书路径")
	tlsSignKey := fs.String("tls-sign-key", "", "TLS 签名密钥路径")

	clientCA := fs.String("client-ca", "", "客户端CA证书路径，多个用逗号分隔")
	serverCA := fs.String("server-ca", "", "服务端CA证书路径，多个用逗号分隔")

	timeoutDial := fs.Int("timeout-dial", 0, "连接建立超时（秒）")
	timeoutRead := fs.Int("timeout-read", 0, "读取超时（秒）")
	timeoutWrite := fs.Int("timeout-write", 0, "写入超时（秒）")
	timeoutHandshake := fs.Int("timeout-handshake", 0, "握手超时（秒）")

	tlcpMinVersion := fs.String("tlcp-min-version", "", "TLCP最小协议版本")
	tlcpMaxVersion := fs.String("tlcp-max-version", "", "TLCP最大协议版本")
	tlcpCipherSuites := fs.String("tlcp-cipher-suites", "", "TLCP密码套件，多个用逗号分隔")
	tlcpCurvePreferences := fs.String("tlcp-curve-preferences", "", "TLCP椭圆曲线偏好，多个用逗号分隔")
	tlcpSessionTickets := fs.Bool("tlcp-session-tickets", false, "启用TLCP会话票据")
	tlcpSessionCache := fs.Bool("tlcp-session-cache", false, "启用TLCP会话缓存")
	tlcpInsecureSkipVerify := fs.Bool("tlcp-insecure-skip-verify", false, "跳过TLCP证书验证（不安全）")

	tlsMinVersion := fs.String("tls-min-version", "", "TLS最小协议版本")
	tlsMaxVersion := fs.String("tls-max-version", "", "TLS最大协议版本")
	tlsCipherSuites := fs.String("tls-cipher-suites", "", "TLS密码套件，多个用逗号分隔")
	tlsCurvePreferences := fs.String("tls-curve-preferences", "", "TLS椭圆曲线偏好，多个用逗号分隔")
	tlsSessionTickets := fs.Bool("tls-session-tickets", false, "启用TLS会话票据")
	tlsSessionCache := fs.Bool("tls-session-cache", false, "启用TLS会话缓存")
	tlsInsecureSkipVerify := fs.Bool("tls-insecure-skip-verify", false, "跳过TLS证书验证（不安全）")

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
		Enabled:    *enabled,
		SNI:        *sni,
		BufferSize: *bufferSize,
	}

	if *clientCA != "" {
		cfg.ClientCA = splitString(*clientCA, ",")
	}
	if *serverCA != "" {
		cfg.ServerCA = splitString(*serverCA, ",")
	}

	if *timeoutDial > 0 || *timeoutRead > 0 || *timeoutWrite > 0 || *timeoutHandshake > 0 {
		cfg.Timeout = &client.TimeoutConfig{}
		if *timeoutDial > 0 {
			cfg.Timeout.Dial = time.Duration(*timeoutDial) * time.Second
		}
		if *timeoutRead > 0 {
			cfg.Timeout.Read = time.Duration(*timeoutRead) * time.Second
		}
		if *timeoutWrite > 0 {
			cfg.Timeout.Write = time.Duration(*timeoutWrite) * time.Second
		}
		if *timeoutHandshake > 0 {
			cfg.Timeout.Handshake = time.Duration(*timeoutHandshake) * time.Second
		}
	}

	if *keystoreName != "" {
		if *protocol == "tlcp" || *protocol == "auto" {
			cfg.TLCP = &client.TLCPConfig{
				ClientAuthType: *tlcpClientAuthType,
				Keystore: &client.KeyStoreConfig{
					Name: *keystoreName,
				},
			}
			populateTLCPConfig(cfg.TLCP, tlcpMinVersion, tlcpMaxVersion, tlcpCipherSuites,
				tlcpCurvePreferences, tlcpSessionTickets, tlcpSessionCache, tlcpInsecureSkipVerify)
		}
		if *protocol == "tls" || *protocol == "auto" {
			cfg.TLS = &client.TLSConfig{
				ClientAuthType: *tlsClientAuthType,
				Keystore: &client.KeyStoreConfig{
					Name: *keystoreName,
				},
			}
			populateTLSConfig(cfg.TLS, tlsMinVersion, tlsMaxVersion, tlsCipherSuites,
				tlsCurvePreferences, tlsSessionTickets, tlsSessionCache, tlsInsecureSkipVerify)
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
					ClientAuthType: *tlcpClientAuthType,
					Keystore: &client.KeyStoreConfig{
						Type:   "file",
						Params: params,
					},
				}
				populateTLCPConfig(cfg.TLCP, tlcpMinVersion, tlcpMaxVersion, tlcpCipherSuites,
					tlcpCurvePreferences, tlcpSessionTickets, tlcpSessionCache, tlcpInsecureSkipVerify)
			}
		}

		if *tlsSignCert != "" && *tlsSignKey != "" {
			if *protocol == "tls" || *protocol == "auto" {
				params := map[string]string{
					"sign-cert": *tlsSignCert,
					"sign-key":  *tlsSignKey,
				}
				cfg.TLS = &client.TLSConfig{
					ClientAuthType: *tlsClientAuthType,
					Keystore: &client.KeyStoreConfig{
						Type:   "file",
						Params: params,
					},
				}
				populateTLSConfig(cfg.TLS, tlsMinVersion, tlsMaxVersion, tlsCipherSuites,
					tlsCurvePreferences, tlsSessionTickets, tlsSessionCache, tlsInsecureSkipVerify)
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
	tlcpClientAuthType := fs.String("tlcp-client-auth-type", "", "TLCP客户端认证类型")
	tlsClientAuthType := fs.String("tls-client-auth-type", "", "TLS客户端认证类型")
	keystoreName := fs.String("keystore-name", "", "keystore 名称")
	enabled := fs.Bool("enabled", false, "是否启用")
	sni := fs.String("sni", "", "SNI 名称")
	bufferSize := fs.Int("buffer-size", 0, "缓冲区大小")

	tlcpSignCert := fs.String("tlcp-sign-cert", "", "TLCP 签名证书路径")
	tlcpSignKey := fs.String("tlcp-sign-key", "", "TLCP 签名密钥路径")
	tlcpEncCert := fs.String("tlcp-enc-cert", "", "TLCP 加密证书路径")
	tlcpEncKey := fs.String("tlcp-enc-key", "", "TLCP 加密密钥路径")

	tlsSignCert := fs.String("tls-sign-cert", "", "TLS 签名证书路径")
	tlsSignKey := fs.String("tls-sign-key", "", "TLS 签名密钥路径")

	clientCA := fs.String("client-ca", "", "客户端CA证书路径，多个用逗号分隔")
	serverCA := fs.String("server-ca", "", "服务端CA证书路径，多个用逗号分隔")

	timeoutDial := fs.Int("timeout-dial", 0, "连接建立超时（秒）")
	timeoutRead := fs.Int("timeout-read", 0, "读取超时（秒）")
	timeoutWrite := fs.Int("timeout-write", 0, "写入超时（秒）")
	timeoutHandshake := fs.Int("timeout-handshake", 0, "握手超时（秒）")

	tlcpMinVersion := fs.String("tlcp-min-version", "", "TLCP最小协议版本")
	tlcpMaxVersion := fs.String("tlcp-max-version", "", "TLCP最大协议版本")
	tlcpCipherSuites := fs.String("tlcp-cipher-suites", "", "TLCP密码套件，多个用逗号分隔")
	tlcpCurvePreferences := fs.String("tlcp-curve-preferences", "", "TLCP椭圆曲线偏好，多个用逗号分隔")
	tlcpSessionTickets := fs.Bool("tlcp-session-tickets", false, "启用TLCP会话票据")
	tlcpSessionCache := fs.Bool("tlcp-session-cache", false, "启用TLCP会话缓存")
	tlcpInsecureSkipVerify := fs.Bool("tlcp-insecure-skip-verify", false, "跳过TLCP证书验证（不安全）")

	tlsMinVersion := fs.String("tls-min-version", "", "TLS最小协议版本")
	tlsMaxVersion := fs.String("tls-max-version", "", "TLS最大协议版本")
	tlsCipherSuites := fs.String("tls-cipher-suites", "", "TLS密码套件，多个用逗号分隔")
	tlsCurvePreferences := fs.String("tls-curve-preferences", "", "TLS椭圆曲线偏好，多个用逗号分隔")
	tlsSessionTickets := fs.Bool("tls-session-tickets", false, "启用TLS会话票据")
	tlsSessionCache := fs.Bool("tls-session-cache", false, "启用TLS会话缓存")
	tlsInsecureSkipVerify := fs.Bool("tls-insecure-skip-verify", false, "跳过TLS证书验证（不安全）")

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
	if *tlcpClientAuthType != "" {
		cfg.TLCP.ClientAuthType = *tlcpClientAuthType
	}
	if *tlsClientAuthType != "" {
		cfg.TLS.ClientAuthType = *tlsClientAuthType
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

	if *clientCA != "" {
		cfg.ClientCA = splitString(*clientCA, ",")
	}
	if *serverCA != "" {
		cfg.ServerCA = splitString(*serverCA, ",")
	}

	if *timeoutDial > 0 || *timeoutRead > 0 || *timeoutWrite > 0 || *timeoutHandshake > 0 {
		if cfg.Timeout == nil {
			cfg.Timeout = &client.TimeoutConfig{}
		}
		if *timeoutDial > 0 {
			cfg.Timeout.Dial = time.Duration(*timeoutDial) * time.Second
		}
		if *timeoutRead > 0 {
			cfg.Timeout.Read = time.Duration(*timeoutRead) * time.Second
		}
		if *timeoutWrite > 0 {
			cfg.Timeout.Write = time.Duration(*timeoutWrite) * time.Second
		}
		if *timeoutHandshake > 0 {
			cfg.Timeout.Handshake = time.Duration(*timeoutHandshake) * time.Second
		}
	}

	if *keystoreName != "" {
		if cfg.Protocol == "tlcp" || cfg.Protocol == "auto" {
			if cfg.TLCP == nil {
				cfg.TLCP = &client.TLCPConfig{}
			}
			cfg.TLCP.Keystore = &client.KeyStoreConfig{
				Name: *keystoreName,
			}
			populateTLCPConfig(cfg.TLCP, tlcpMinVersion, tlcpMaxVersion, tlcpCipherSuites,
				tlcpCurvePreferences, tlcpSessionTickets, tlcpSessionCache, tlcpInsecureSkipVerify)
		}
		if cfg.Protocol == "tls" || cfg.Protocol == "auto" {
			if cfg.TLS == nil {
				cfg.TLS = &client.TLSConfig{}
			}
			cfg.TLS.Keystore = &client.KeyStoreConfig{
				Name: *keystoreName,
			}
			populateTLSConfig(cfg.TLS, tlsMinVersion, tlsMaxVersion, tlsCipherSuites,
				tlsCurvePreferences, tlsSessionTickets, tlsSessionCache, tlsInsecureSkipVerify)
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
				if *tlcpClientAuthType != "" {
					cfg.TLCP.ClientAuthType = *tlcpClientAuthType
				}
				populateTLCPConfig(cfg.TLCP, tlcpMinVersion, tlcpMaxVersion, tlcpCipherSuites,
					tlcpCurvePreferences, tlcpSessionTickets, tlcpSessionCache, tlcpInsecureSkipVerify)
			}
		}

		if *tlsSignCert != "" && *tlsSignKey != "" {
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
				params["sign-cert"] = *tlsSignCert
				params["sign-key"] = *tlsSignKey
				cfg.TLS.Keystore = &client.KeyStoreConfig{
					Type:   "file",
					Params: params,
				}
				if *tlsClientAuthType != "" {
					cfg.TLS.ClientAuthType = *tlsClientAuthType
				}
				populateTLSConfig(cfg.TLS, tlsMinVersion, tlsMaxVersion, tlsCipherSuites,
					tlsCurvePreferences, tlsSessionTickets, tlsSessionCache, tlsInsecureSkipVerify)
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

func splitString(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func populateTLCPConfig(cfg *client.TLCPConfig, minVersion, maxVersion, cipherSuites, curvePreferences *string,
	sessionTickets, sessionCache, insecureSkipVerify *bool) {
	if *minVersion != "" {
		cfg.MinVersion = *minVersion
	}
	if *maxVersion != "" {
		cfg.MaxVersion = *maxVersion
	}
	if *cipherSuites != "" {
		cfg.CipherSuites = splitString(*cipherSuites, ",")
	}
	if *curvePreferences != "" {
		cfg.CurvePreferences = splitString(*curvePreferences, ",")
	}
	if *sessionTickets {
		cfg.SessionTickets = *sessionTickets
	}
	if *sessionCache {
		cfg.SessionCache = *sessionCache
	}
	if *insecureSkipVerify {
		cfg.InsecureSkipVerify = *insecureSkipVerify
	}
}

func populateTLSConfig(cfg *client.TLSConfig, minVersion, maxVersion, cipherSuites, curvePreferences *string,
	sessionTickets, sessionCache, insecureSkipVerify *bool) {
	if *minVersion != "" {
		cfg.MinVersion = *minVersion
	}
	if *maxVersion != "" {
		cfg.MaxVersion = *maxVersion
	}
	if *cipherSuites != "" {
		cfg.CipherSuites = splitString(*cipherSuites, ",")
	}
	if *curvePreferences != "" {
		cfg.CurvePreferences = splitString(*curvePreferences, ",")
	}
	if *sessionTickets {
		cfg.SessionTickets = *sessionTickets
	}
	if *sessionCache {
		cfg.SessionCache = *sessionCache
	}
	if *insecureSkipVerify {
		cfg.InsecureSkipVerify = *insecureSkipVerify
	}
}
