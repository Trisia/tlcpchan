package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Trisia/tlcpchan-cli/client"
)

func keyStoreList(args []string) error {
	keyStores, err := cli.ListKeyStores()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(keyStores)
	}

	if len(keyStores) == 0 {
		fmt.Println("无 keystore")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 仅当列表中确实存在 IBC keystore 时才附加 IBC 标识列，避免证书用户看到无意义的空列
	hasIBC := false
	for _, ks := range keyStores {
		if ks.IBC != nil {
			hasIBC = true
			break
		}
	}

	if hasIBC {
		fmt.Fprintln(w, "名称\t类型\t加载器\t保护\tIBC标识\t创建时间")
	} else {
		fmt.Fprintln(w, "名称\t类型\t加载器\t保护\t创建时间")
	}
	for _, ks := range keyStores {
		protectedStatus := "否"
		if ks.Protected {
			protectedStatus = "是"
		}
		if hasIBC {
			identity := "-"
			if ks.IBC != nil && ks.IBC.Identity != "" {
				identity = ks.IBC.Identity
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				ks.Name, ks.Type, ks.LoaderType, protectedStatus, identity, ks.CreatedAt)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ks.Name, ks.Type, ks.LoaderType, protectedStatus, ks.CreatedAt)
	}
	w.Flush()
	return nil
}

func keyStoreShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}

	ks, err := cli.GetKeyStore(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(ks)
	}

	fmt.Printf("名称: %s\n", ks.Name)
	fmt.Printf("类型: %s\n", ks.Type)
	fmt.Printf("加载器: %s\n", ks.LoaderType)
	fmt.Printf("受保护: %v\n", ks.Protected)
	fmt.Printf("创建时间: %s\n", ks.CreatedAt)
	fmt.Printf("更新时间: %s\n", ks.UpdatedAt)
	fmt.Println("参数:")
	for k, v := range ks.Params {
		fmt.Printf("  %s: %s\n", k, v)
	}
	if ks.IBC != nil {
		fmt.Println()
		printIBCInfo(ks.IBC)
	}
	return nil
}

func keyStoreShowDetail(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}

	ks, err := cli.GetKeyStore(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(ks)
	}

	fmt.Println("========== Keystore 详情 ==========")
	fmt.Printf("名称: %s\n", ks.Name)
	fmt.Printf("类型: %s\n", ks.Type)
	fmt.Printf("加载器类型: %s\n", ks.LoaderType)
	fmt.Printf("受保护: %v\n", ks.Protected)
	fmt.Printf("创建时间: %s\n", ks.CreatedAt)
	fmt.Printf("更新时间: %s\n", ks.UpdatedAt)

	// IBC 身份不含证书，单独命名小节以免误解
	paramsSection := "---------- 证书密钥参数 ----------"
	if ks.Type == "ibc" {
		paramsSection = "---------- IBC 身份参数 ----------"
	}
	fmt.Printf("\n%s\n", paramsSection)
	if ks.Protected {
		fmt.Println("⚠️  受保护的 keystore 不允许修改")
	}
	if ks.LoaderType != "file" && ks.LoaderType != "ibc-file" {
		fmt.Println("ℹ️  只有文件类型(file/ibc-file)的 keystore 支持编辑参数")
	}

	fmt.Println("当前参数:")
	for k, v := range ks.Params {
		fmt.Printf("  %s: %s\n", k, v)
	}

	if ks.IBC != nil {
		fmt.Println()
		printIBCInfo(ks.IBC)
	}

	fmt.Println("\n---------- 关联实例 ----------")
	instances, err := cli.GetKeyStoreInstances(args[0])
	if err != nil {
		fmt.Printf("⚠️  获取关联实例失败: %v\n", err)
		return nil
	}

	if len(instances) == 0 {
		fmt.Println("暂无关联实例")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "实例名称\t协议类型\t状态")
	runningCount := 0
	for _, inst := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\n", inst.Name, inst.Protocol, inst.Status)
		if inst.Status == "running" {
			runningCount++
		}
	}
	w.Flush()

	if runningCount > 0 {
		fmt.Printf("\n⚠️  有 %d 个运行中的实例，修改参数后需要重新加载\n", runningCount)
		fmt.Printf("使用命令: tlcpchan-cli instance reload <instance-name>\n")
	}

	return nil
}

func keyStoreUpdateParams(args []string) error {
	fs := flagSet("update")
	loaderType := fs.String("loader-type", "", "加载器类型 (file/ibc-file)，用于校验可更新的参数；留空则按参数自动判断")
	signCert := fs.String("sign-cert", "", "签名证书文件路径")
	signKey := fs.String("sign-key", "", "签名密钥文件路径（ibc-file 时为签名私钥 hid=0x01）")
	encCert := fs.String("enc-cert", "", "加密证书文件路径 (TLCP)")
	encKey := fs.String("enc-key", "", "加密密钥文件路径（TLCP；ibc-file 时为加密私钥 hid=0x03）")
	cert := fs.String("cert", "", "证书文件路径 (TLS)")
	key := fs.String("key", "", "密钥文件路径 (TLS)")
	identity := fs.String("identity", "", "IBC 本端标识文件路径 (ibc-file)")
	paramsFile := fs.String("params", "", "IBC KGC 公共参数文件路径 (ibc-file)")
	kexKey := fs.String("kex-key", "", "IBC 密钥交换私钥文件路径 (ibc-file, hid=0x02)")
	if err := fs.Parse(reorderFlagsFirst(fs, args)); err != nil {
		return err
	}

	if len(fs.Args()) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}
	name := fs.Args()[0]

	if *loaderType != "" && *loaderType != "file" && *loaderType != "ibc-file" {
		return fmt.Errorf("不支持的 --loader-type: %s (可选 file/ibc-file)", *loaderType)
	}

	// 未显式指定加载器时，按是否出现 IBC 专有参数自动判定
	isIBC := *loaderType == "ibc-file" || *identity != "" || *paramsFile != "" || *kexKey != ""
	if isIBC && (*signCert != "" || *encCert != "" || *cert != "" || *key != "") {
		return fmt.Errorf("ibc-file keystore 不支持 --sign-cert / --enc-cert / --cert / --key，请使用 --identity / --params / --sign-key / --enc-key / --kex-key")
	}

	params := make(map[string]string)
	if *signCert != "" {
		params["sign-cert"] = *signCert
	}
	if *signKey != "" {
		params["sign-key"] = *signKey
	}
	if *encCert != "" {
		params["enc-cert"] = *encCert
	}
	if *encKey != "" {
		params["enc-key"] = *encKey
	}
	if *cert != "" {
		params["cert"] = *cert
	}
	if *key != "" {
		params["key"] = *key
	}
	if *identity != "" {
		params["identity"] = *identity
	}
	if *paramsFile != "" {
		params["params"] = *paramsFile
	}
	if *kexKey != "" {
		params["kex-key"] = *kexKey
	}

	if len(params) == 0 {
		return fmt.Errorf("请指定至少一个参数 (--identity, --params, --sign-key, --enc-key, --kex-key, --sign-cert, --enc-cert, --cert, --key)")
	}

	ks, err := cli.UpdateKeyStoreParams(name, params)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "keystore 参数更新成功",
			"name":    ks.Name,
		})
	}

	fmt.Printf("keystore %s 参数更新成功\n", ks.Name)

	instances, err := cli.GetKeyStoreInstances(name)
	if err != nil {
		return nil
	}

	runningCount := 0
	for _, inst := range instances {
		if inst.Status == "running" {
			runningCount++
		}
	}

	if runningCount > 0 {
		fmt.Printf("\n⚠️  警告: 有 %d 个运行中的实例引用此 keystore\n", runningCount)
		fmt.Println("修改后需要重新加载这些实例才能生效:")
		for _, inst := range instances {
			if inst.Status == "running" {
				fmt.Printf("  - tlcpchan-cli instance reload %s\n", inst.Name)
			}
		}
	}

	return nil
}

// keyStoreCreate 创建 keystore
// 参数：
//   - args: 命令行参数，格式为 [--name <名称>] [--loader-type <类型>] [材料文件选项]
//
// 返回：
//   - error: 错误信息
//
// 说明：
//   - loader-type=ibc-file 时按 IBC（SM9）语义读取材料：--identity / --params / --sign-key /
//     --enc-key / --kex-key，multipart 字段名固定为 identity / params / signKey / encKey / kexKey；
//   - 其他加载器沿用证书材料选项：--sign-cert / --sign-key / --enc-cert / --enc-key。
func keyStoreCreate(args []string) error {
	fs := flagSet("create")
	name := fs.String("name", "", "keystore 名称")
	loaderType := fs.String("loader-type", "file", "加载器类型 (file/named/skf/sdf/ibc-file)")
	signCert := fs.String("sign-cert", "", "签名证书文件路径")
	signKey := fs.String("sign-key", "", "签名密钥文件路径（ibc-file 时为签名私钥 hid=0x01）")
	encCert := fs.String("enc-cert", "", "加密证书文件路径 (TLCP)")
	encKey := fs.String("enc-key", "", "加密密钥文件路径（TLCP；ibc-file 时为加密私钥 hid=0x03）")
	identity := fs.String("identity", "", "IBC 本端标识文件路径 (ibc-file，必需)")
	paramsFile := fs.String("params", "", "IBC KGC 公共参数文件路径 (ibc-file)")
	kexKey := fs.String("kex-key", "", "IBC 密钥交换私钥文件路径 (ibc-file, hid=0x02)")
	protected := fs.Bool("protected", false, "是否受保护")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("请指定 --name")
	}

	files := make(map[string][]byte)

	if *loaderType == "ibc-file" {
		if *signCert != "" || *encCert != "" {
			return fmt.Errorf("ibc-file 类型不支持 --sign-cert / --enc-cert，请使用 --identity / --params / --sign-key / --enc-key / --kex-key")
		}
		if *identity == "" {
			return fmt.Errorf("ibc-file 类型必须指定 --identity（本端标识）")
		}

		// IBC 材料的 multipart 字段名固定为驼峰形式，与证书材料字段区分
		ibcFiles := []struct {
			field string
			path  string
			label string
		}{
			{"identity", *identity, "IBC 标识"},
			{"params", *paramsFile, "IBC KGC 公共参数"},
			{"signKey", *signKey, "IBC 签名私钥"},
			{"encKey", *encKey, "IBC 加密私钥"},
			{"kexKey", *kexKey, "IBC 密钥交换私钥"},
		}
		for _, f := range ibcFiles {
			if f.path == "" {
				continue
			}
			data, err := os.ReadFile(f.path)
			if err != nil {
				return fmt.Errorf("读取%s失败: %w", f.label, err)
			}
			files[f.field] = data
		}
	} else {
		if *loaderType != "file" && *loaderType != "named" && *loaderType != "skf" && *loaderType != "sdf" {
			return fmt.Errorf("不支持的 --loader-type: %s (可选 file/named/skf/sdf/ibc-file)", *loaderType)
		}

		if *signCert != "" {
			data, err := os.ReadFile(*signCert)
			if err != nil {
				return fmt.Errorf("读取签名证书失败: %w", err)
			}
			files["sign-cert"] = data
		}

		if *signKey != "" {
			data, err := os.ReadFile(*signKey)
			if err != nil {
				return fmt.Errorf("读取签名密钥失败: %w", err)
			}
			files["sign-key"] = data
		}

		if *encCert != "" {
			data, err := os.ReadFile(*encCert)
			if err != nil {
				return fmt.Errorf("读取加密证书失败: %w", err)
			}
			files["enc-cert"] = data
		}

		if *encKey != "" {
			data, err := os.ReadFile(*encKey)
			if err != nil {
				return fmt.Errorf("读取加密密钥失败: %w", err)
			}
			files["enc-key"] = data
		}
	}

	ks, err := cli.CreateKeyStoreWithFiles(*name, *loaderType, files, *protected)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "keystore 创建成功",
			"name":    ks.Name,
		})
	}

	fmt.Printf("keystore %s 创建成功\n", ks.Name)
	return nil
}

// keyStoreGenerate 生成 keystore
// 参数：
//   - args: 命令行参数，格式为 [--name <名称>] [--type <tlcp|tls|ibc>] [生成选项]
//
// 返回：
//   - error: 错误信息
//
// 说明：
//   - type=tlcp/tls 时使用证书生成选项（--cn 等），证书字段必填 --cn；
//   - type=ibc 时使用 --identity（必需）与可选的 --district-name / --district-serial，
//     服务端从 IBC 信任池取目标 KGC 公共参数并派生签名/加密/密钥交换三把私钥。
func keyStoreGenerate(args []string) error {
	fs := flagSet("generate")
	name := fs.String("name", "", "keystore 名称")
	ksType := fs.String("type", "tlcp", "类型 (tlcp/tls/ibc)")
	identity := fs.String("identity", "", "IBC 本端标识，如 server@tlcpchan.local (仅 ibc)")
	districtName := fs.String("district-name", "", "IBC 目标 KGC 属地区域名 (仅 ibc，留空由服务端从信任池选择)")
	districtSerial := fs.Int("district-serial", -1, "IBC 目标 KGC 序号 (仅 ibc，-1 表示由服务端选择)")
	commonName := fs.String("cn", "", "证书通用名称 (CN)")
	country := fs.String("c", "", "国家 (C, 2字母代码)")
	stateOrProvince := fs.String("st", "", "省/州 (ST)")
	locality := fs.String("l", "", "地区/城市 (L)")
	org := fs.String("org", "tlcpchan", "组织名称 (O)")
	orgUnit := fs.String("org-unit", "", "组织单位 (OU)")
	email := fs.String("email", "", "邮箱地址")
	years := fs.Int("years", 0, "证书有效期 (年)")
	days := fs.Int("days", 0, "证书有效期 (天, 优先级高于years)")
	keyAlg := fs.String("key-alg", "ecdsa", "密钥算法 (ecdsa/rsa, 仅TLS有效)")
	keyBits := fs.Int("key-bits", 2048, "密钥位数 (仅RSA有效)")
	dnsNames := fs.String("dns", "", "DNS名称, 多个用逗号分隔")
	ipAddrs := fs.String("ip", "", "IP地址, 多个用逗号分隔")
	protected := fs.Bool("protected", false, "是否受保护")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("请指定 --name")
	}

	isIBC := *ksType == "ibc"
	if isIBC {
		if *identity == "" {
			return fmt.Errorf("ibc 类型必须指定 --identity（本端标识）")
		}
	} else if *commonName == "" {
		return fmt.Errorf("请指定 --cn")
	}

	var dnsList []string
	if *dnsNames != "" {
		dnsList = splitAndTrim(*dnsNames, ",")
	}

	var ipList []string
	if *ipAddrs != "" {
		ipList = splitAndTrim(*ipAddrs, ",")
	}

	req := client.GenerateKeyStoreRequest{
		Name:      *name,
		Type:      *ksType,
		Protected: *protected,
	}
	if isIBC {
		req.Identity = *identity
		req.DistrictName = *districtName
		if *districtSerial >= 0 {
			serial := *districtSerial
			req.DistrictSerial = &serial
		}
	} else {
		req.CertConfig = &client.GenerateKeyStoreCertConfig{
			CommonName:      *commonName,
			Country:         *country,
			StateOrProvince: *stateOrProvince,
			Locality:        *locality,
			Org:             *org,
			OrgUnit:         *orgUnit,
			EmailAddress:    *email,
			Years:           *years,
			Days:            *days,
			KeyAlgorithm:    *keyAlg,
			KeyBits:         *keyBits,
			DNSNames:        dnsList,
			IPAddresses:     ipList,
		}
	}

	ks, err := cli.GenerateKeyStore(req)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "keystore 生成成功",
			"name":    ks.Name,
		})
	}

	fmt.Printf("keystore %s 生成成功\n", ks.Name)
	return nil
}

func splitAndTrim(s, sep string) []string {
	var result []string
	parts := strings.Split(s, sep)
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func keyStoreDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}

	if err := cli.DeleteKeyStore(args[0]); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "keystore 已删除",
			"name":    args[0],
		})
	}

	fmt.Printf("keystore %s 已删除\n", args[0])
	return nil
}

func keyStoreExportCSR(args []string) error {
	fs := flagSet("export-csr")
	keyType := fs.String("key-type", "sign", "密钥类型 (sign/enc)")
	commonName := fs.String("cn", "", "证书通用名称 (CN)")
	country := fs.String("c", "", "国家 (C, 2字母代码)")
	stateOrProvince := fs.String("st", "", "省/州 (ST)")
	locality := fs.String("l", "", "地区/城市 (L)")
	org := fs.String("org", "", "组织名称 (O)")
	orgUnit := fs.String("org-unit", "", "组织单位 (OU)")
	email := fs.String("email", "", "邮箱地址")
	dnsNames := fs.String("dns", "", "DNS名称, 多个用逗号分隔")
	ipAddrs := fs.String("ip", "", "IP地址, 多个用逗号分隔")
	outputPath := fs.String("output", "", "输出文件路径")

	if err := fs.Parse(reorderFlagsFirst(fs, args)); err != nil {
		return err
	}

	if *commonName == "" {
		return fmt.Errorf("请指定 --cn")
	}

	if len(fs.Args()) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}
	name := fs.Args()[0]

	var dnsList []string
	if *dnsNames != "" {
		dnsList = splitAndTrim(*dnsNames, ",")
	}

	var ipList []string
	if *ipAddrs != "" {
		ipList = splitAndTrim(*ipAddrs, ",")
	}

	req := client.ExportCSRRequest{
		KeyType: *keyType,
	}
	req.CSRParams.CommonName = *commonName
	req.CSRParams.Country = *country
	req.CSRParams.StateOrProvince = *stateOrProvince
	req.CSRParams.Locality = *locality
	req.CSRParams.Org = *org
	req.CSRParams.OrgUnit = *orgUnit
	req.CSRParams.EmailAddress = *email
	req.CSRParams.DNSNames = dnsList
	req.CSRParams.IPAddresses = ipList

	data, filename, err := cli.ExportKeyStoreCSR(name, req)
	if err != nil {
		return err
	}

	outputFile := *outputPath
	if outputFile == "" {
		outputFile = filename
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "CSR导出成功",
			"name":    name,
			"output":  outputFile,
		})
	}

	fmt.Printf("CSR已导出到: %s\n", outputFile)
	return nil
}

// keyStoreUploadCertificates 上传材料文件以更新 keystore
// 参数：
//   - args: 命令行参数，格式为 <keystore-name> [--loader-type <file|ibc-file>] [材料文件选项]
//
// 返回：
//   - error: 错误信息
//
// 说明：
//   - ibc-file（IBC/SM9）材料读取 --identity / --params / --sign-key / --enc-key / --kex-key，
//     multipart 字段名固定为 identity / params / signKey / encKey / kexKey；
//   - file（证书）材料读取 --sign-cert / --sign-key / --enc-cert / --enc-key，
//     multipart 字段名固定为 signCert / signKey / encCert / encKey；
//   - 未指定 --loader-type 时按是否出现 IBC 专有参数自动判断。
func keyStoreUploadCertificates(args []string) error {
	fs := flagSet("upload")
	loaderType := fs.String("loader-type", "", "加载器类型 (file/ibc-file)，用于选择材料字段；留空则按参数自动判断")
	signCert := fs.String("sign-cert", "", "签名证书文件路径（TLS类型时为证书，TLCP类型时为签名证书）")
	signKey := fs.String("sign-key", "", "签名密钥文件路径（TLS类型时为密钥，TLCP类型时为签名密钥；ibc-file 时为签名私钥 hid=0x01）")
	encCert := fs.String("enc-cert", "", "加密证书文件路径（仅TLCP类型有效）")
	encKey := fs.String("enc-key", "", "加密密钥文件路径（仅TLCP类型有效；ibc-file 时为加密私钥 hid=0x03）")
	identity := fs.String("identity", "", "IBC 本端标识文件路径 (ibc-file)")
	paramsFile := fs.String("params", "", "IBC KGC 公共参数文件路径 (ibc-file)")
	kexKey := fs.String("kex-key", "", "IBC 密钥交换私钥文件路径 (ibc-file, hid=0x02)")
	if err := fs.Parse(reorderFlagsFirst(fs, args)); err != nil {
		return err
	}

	if len(fs.Args()) == 0 {
		return fmt.Errorf("请指定 keystore 名称")
	}
	name := fs.Args()[0]

	if *loaderType != "" && *loaderType != "file" && *loaderType != "ibc-file" {
		return fmt.Errorf("不支持的 --loader-type: %s (可选 file/ibc-file)", *loaderType)
	}

	// 未显式指定加载器时，按是否出现 IBC 专有参数自动判定
	isIBC := *loaderType == "ibc-file" || *identity != "" || *paramsFile != "" || *kexKey != ""
	if isIBC && (*signCert != "" || *encCert != "") {
		return fmt.Errorf("ibc-file keystore 不支持 --sign-cert / --enc-cert，请使用 --identity / --params / --sign-key / --enc-key / --kex-key")
	}
	if !isIBC && (*identity != "" || *paramsFile != "" || *kexKey != "") {
		return fmt.Errorf("--loader-type file 不支持 IBC 材料参数 (--identity / --params / --kex-key)")
	}

	files := make(map[string][]byte)

	if isIBC {
		// IBC 材料的 multipart 字段名固定为驼峰形式
		ibcFiles := []struct {
			field string
			path  string
			label string
		}{
			{"identity", *identity, "IBC 标识"},
			{"params", *paramsFile, "IBC KGC 公共参数"},
			{"signKey", *signKey, "IBC 签名私钥"},
			{"encKey", *encKey, "IBC 加密私钥"},
			{"kexKey", *kexKey, "IBC 密钥交换私钥"},
		}
		for _, f := range ibcFiles {
			if f.path == "" {
				continue
			}
			data, err := os.ReadFile(f.path)
			if err != nil {
				return fmt.Errorf("读取%s失败: %w", f.label, err)
			}
			files[f.field] = data
		}
	} else {
		if *signCert != "" {
			data, err := os.ReadFile(*signCert)
			if err != nil {
				return fmt.Errorf("读取签名证书失败: %w", err)
			}
			files["signCert"] = data
		}

		if *signKey != "" {
			data, err := os.ReadFile(*signKey)
			if err != nil {
				return fmt.Errorf("读取签名密钥失败: %w", err)
			}
			files["signKey"] = data
		}

		if *encCert != "" {
			data, err := os.ReadFile(*encCert)
			if err != nil {
				return fmt.Errorf("读取加密证书失败: %w", err)
			}
			files["encCert"] = data
		}

		if *encKey != "" {
			data, err := os.ReadFile(*encKey)
			if err != nil {
				return fmt.Errorf("读取加密密钥失败: %w", err)
			}
			files["encKey"] = data
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("请指定至少一个文件 (--identity, --params, --sign-key, --enc-key, --kex-key, --sign-cert, --enc-cert)")
	}

	ks, err := cli.UpdateKeyStoreCertificates(name, files)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "keystore 证书和密钥更新成功",
			"name":    ks.Name,
		})
	}

	fmt.Printf("keystore %s 证书和密钥更新成功\n", ks.Name)

	instances, err := cli.GetKeyStoreInstances(name)
	if err != nil {
		return nil
	}

	runningCount := 0
	for _, inst := range instances {
		if inst.Status == "running" {
			runningCount++
		}
	}

	if runningCount > 0 {
		fmt.Printf("\n⚠️  警告: 有 %d 个运行实例引用此 keystore\n", runningCount)
		fmt.Println("修改后需要重新加载这些实例才能生效:")
		for _, inst := range instances {
			if inst.Status == "running" {
				fmt.Printf("  - tlcpchan-cli instance reload %s\n", inst.Name)
			}
		}
	}

	return nil
}

// printIBCInfo 以键值对形式打印 IBC（SM9）身份元信息
// 参数：
//   - ibc: keystore 的 IBC 元信息，为 nil 时不输出任何内容
//
// 返回：无
func printIBCInfo(ibc *client.IBCInfo) {
	if ibc == nil {
		return
	}
	fmt.Println("---------- IBC 身份信息 ----------")
	fmt.Printf("标识: %s\n", ibc.Identity)
	fmt.Printf("含公共参数: %v\n", ibc.HasParams)
	fmt.Printf("KGC 区域: %s#%d\n", ibc.DistrictName, ibc.DistrictSerial)
	fmt.Printf("公共参数生效时间: %s\n", formatIBCValidity(ibc.NotBefore))
	fmt.Printf("公共参数失效时间: %s\n", formatIBCValidity(ibc.NotAfter))
	fmt.Printf("签名私钥(hid=0x01): %v\n", ibc.HasSignKey)
	fmt.Printf("加密私钥(hid=0x03): %v\n", ibc.HasEncryptKey)
	fmt.Printf("密钥交换私钥(hid=0x02): %v\n", ibc.HasKeyExchangeKey)
}
