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
	fmt.Fprintln(w, "名称\t类型\t加载器\t保护\t创建时间")
	for _, ks := range keyStores {
		protectedStatus := "否"
		if ks.Protected {
			protectedStatus = "是"
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
	return nil
}

func keyStoreCreate(args []string) error {
	fs := flagSet("create")
	name := fs.String("name", "", "keystore 名称")
	loaderType := fs.String("loader-type", "file", "加载器类型 (file/named/skf/sdf)")
	signCert := fs.String("sign-cert", "", "签名证书文件路径")
	signKey := fs.String("sign-key", "", "签名密钥文件路径")
	encCert := fs.String("enc-cert", "", "加密证书文件路径 (TLCP)")
	encKey := fs.String("enc-key", "", "加密密钥文件路径 (TLCP)")
	protected := fs.Bool("protected", false, "是否受保护")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("请指定 --name")
	}

	files := make(map[string][]byte)

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

func keyStoreGenerate(args []string) error {
	fs := flagSet("generate")
	name := fs.String("name", "", "keystore 名称")
	ksType := fs.String("type", "tlcp", "类型 (tlcp/tls)")
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
	if *commonName == "" {
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
		CertConfig: client.GenerateKeyStoreCertConfig{
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
		},
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

	if err := fs.Parse(args); err != nil {
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
