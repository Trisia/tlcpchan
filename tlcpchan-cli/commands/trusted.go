package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func trustedList(args []string) error {
	certs, err := cli.ListTrustedCertificates()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(certs)
	}

	if len(certs) == 0 {
		fmt.Println("无信任证书")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t类型\t主题\t颁发者\t是否CA\t过期时间")
	for _, cert := range certs {
		isCA := "否"
		if cert.IsCA {
			isCA = "是"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", cert.Name, cert.Type, cert.Subject, cert.Issuer, isCA, cert.ExpiresAt)
	}
	w.Flush()
	return nil
}

func trustedShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定信任证书名称")
	}

	certs, err := cli.ListTrustedCertificates()
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
			if cert.SerialNumber != "" {
				fmt.Printf("序列号: %s\n", cert.SerialNumber)
			}
			if cert.Subject != "" {
				fmt.Printf("主题: %s\n", cert.Subject)
			}
			if cert.Issuer != "" {
				fmt.Printf("颁发者: %s\n", cert.Issuer)
			}
			isCA := "否"
			if cert.IsCA {
				isCA = "是"
			}
			fmt.Printf("是否CA: %s\n", isCA)
			if cert.ExpiresAt != "" {
				fmt.Printf("过期时间: %s\n", cert.ExpiresAt)
			}
			return nil
		}
	}

	return fmt.Errorf("信任证书 %s 不存在", args[0])
}

func trustedReload(args []string) error {
	if err := cli.ReloadTrustedCertificates(); err != nil {
		return err
	}
	fmt.Println("信任证书已重新加载")
	return nil
}

func trustedDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定信任证书名称")
	}

	if err := cli.DeleteTrustedCertificate(args[0]); err != nil {
		return err
	}
	fmt.Printf("信任证书 %s 已删除\n", args[0])
	return nil
}
