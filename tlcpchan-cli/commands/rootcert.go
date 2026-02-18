package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Trisia/tlcpchan-cli/client"
)

func rootCertList(args []string) error {
	certs, err := cli.ListRootCerts()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(certs)
	}

	if len(certs) == 0 {
		fmt.Println("无根证书")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "文件名\t主题\t颁发者\t过期时间")
	for _, cert := range certs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", cert.Filename, cert.Subject, cert.Issuer, cert.NotAfter)
	}
	w.Flush()
	return nil
}

func rootCertShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定根证书文件名")
	}

	cert, err := cli.GetRootCert(args[0])
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(cert)
	}

	fmt.Printf("文件名: %s\n", cert.Filename)
	fmt.Printf("主题: %s\n", cert.Subject)
	fmt.Printf("颁发者: %s\n", cert.Issuer)
	fmt.Printf("过期时间: %s\n", cert.NotAfter)
	return nil
}

func rootCertAdd(args []string) error {
	fs := flagSet("add")
	filename := fs.String("filename", "", "保存的证书文件名")
	certFile := fs.String("cert", "", "证书文件路径")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *filename == "" {
		return fmt.Errorf("请指定 --filename")
	}
	if *certFile == "" {
		return fmt.Errorf("请指定 --cert")
	}

	certData, err := os.ReadFile(*certFile)
	if err != nil {
		return fmt.Errorf("读取证书文件失败: %w", err)
	}

	cert, err := cli.AddRootCert(*filename, certData)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "根证书添加成功",
			"filename": cert.Filename,
		})
	}

	fmt.Printf("根证书 %s 添加成功\n", cert.Filename)
	return nil
}

func rootCertGenerate(args []string) error {
	fs := flagSet("generate")
	commonName := fs.String("cn", "tlcpchan-root-ca", "证书通用名称 (CN)")
	country := fs.String("c", "", "国家 (C, 2字母代码)")
	stateOrProvince := fs.String("st", "", "省/州 (ST)")
	locality := fs.String("l", "", "地区/城市 (L)")
	org := fs.String("org", "tlcpchan", "组织名称 (O)")
	orgUnit := fs.String("org-unit", "", "组织单位 (OU)")
	email := fs.String("email", "", "邮箱地址")
	years := fs.Int("years", 0, "证书有效期 (年)")
	days := fs.Int("days", 0, "证书有效期 (天, 优先级高于years)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req := client.GenerateRootCARequest{
		CommonName:      *commonName,
		Country:         *country,
		StateOrProvince: *stateOrProvince,
		Locality:        *locality,
		Org:             *org,
		OrgUnit:         *orgUnit,
		EmailAddress:    *email,
		Years:           *years,
		Days:            *days,
	}

	cert, err := cli.GenerateRootCA(req)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "根 CA 证书生成成功",
			"filename": cert.Filename,
		})
	}

	fmt.Printf("根 CA 证书 %s 生成成功\n", cert.Filename)
	return nil
}

func rootCertDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定根证书文件名")
	}

	if err := cli.DeleteRootCert(args[0]); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "根证书已删除",
			"filename": args[0],
		})
	}

	fmt.Printf("根证书 %s 已删除\n", args[0])
	return nil
}

func rootCertReload(args []string) error {
	if err := cli.ReloadRootCerts(); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "根证书已重新加载",
		})
	}

	fmt.Println("根证书已重新加载")
	return nil
}
