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

	fmt.Printf("根证书 %s 添加成功\n", cert.Filename)
	return nil
}

func rootCertGenerate(args []string) error {
	fs := flagSet("generate")
	commonName := fs.String("cn", "tlcpchan-root-ca", "证书通用名称 (CN)")
	org := fs.String("org", "tlcpchan", "组织名称 (O)")
	orgUnit := fs.String("org-unit", "", "组织单位 (OU)")
	years := fs.Int("years", 10, "证书有效期 (年)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req := client.GenerateRootCARequest{
		CommonName: *commonName,
		Org:        *org,
		OrgUnit:    *orgUnit,
		Years:      *years,
	}

	cert, err := cli.GenerateRootCA(req)
	if err != nil {
		return err
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
	fmt.Printf("根证书 %s 已删除\n", args[0])
	return nil
}

func rootCertReload(args []string) error {
	if err := cli.ReloadRootCerts(); err != nil {
		return err
	}
	fmt.Println("根证书已重新加载")
	return nil
}
