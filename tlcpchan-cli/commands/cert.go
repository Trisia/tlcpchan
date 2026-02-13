package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func certList(args []string) error {
	certs, err := cli.ListCertificates()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(certs)
	}

	if len(certs) == 0 {
		fmt.Println("无证书")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t类型\t过期时间")
	for _, cert := range certs {
		fmt.Fprintf(w, "%s\t%s\t%s\n", cert.Name, cert.Type, cert.ExpiresAt)
	}
	w.Flush()
	return nil
}

func certShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定证书名称")
	}

	certs, err := cli.ListCertificates()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		if cert.Name == args[0] {
			if isJSONOutput() {
				return printJSON(cert)
			}
			fmt.Printf("名称: %s\n", cert.Name)
			fmt.Printf("类型: %s\n", cert.Type)
			if cert.ExpiresAt != "" {
				fmt.Printf("过期时间: %s\n", cert.ExpiresAt)
			}
			return nil
		}
	}

	return fmt.Errorf("证书 %s 不存在", args[0])
}

func certGenerate(args []string) error {
	fs := flagSet("generate")
	name := fs.String("name", "", "证书名称")
	certType := fs.String("type", "tlcp", "证书类型 (tlcp|tls)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("请使用 -name 指定证书名称")
	}

	fmt.Printf("证书生成功能需要通过API上传证书文件\n")
	fmt.Printf("请使用: curl -F 'file=@%s.crt' %s/api/v1/certificates\n", *name, apiURL)
	fmt.Printf("证书类型: %s\n", *certType)
	return nil
}

func certReload(args []string) error {
	if err := cli.ReloadCertificates(); err != nil {
		return err
	}
	fmt.Println("证书已重新加载")
	return nil
}

func certDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定证书名称")
	}

	if err := cli.DeleteCertificate(args[0]); err != nil {
		return err
	}
	fmt.Printf("证书 %s 已删除\n", args[0])
	return nil
}
