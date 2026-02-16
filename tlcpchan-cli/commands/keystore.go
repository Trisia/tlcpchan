package commands

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

var (
	ksName       string
	ksType       string
	ksSignCert   string
	ksSignKey    string
	ksEncCert    string
	ksEncKey     string
	ksUpdateSign bool
	ksUpdateEnc  bool
)

func init() {
}

func keyStoreList(args []string) error {
	resp, err := cli.Get("/api/v1/keystores")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		KeyStores []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			HasSignCert bool   `json:"hasSignCert"`
			HasSignKey  bool   `json:"hasSignKey"`
			HasEncCert  bool   `json:"hasEncCert"`
			HasEncKey   bool   `json:"hasEncKey"`
			KeyParams   struct {
				Algorithm string `json:"algorithm"`
				Length    int    `json:"length"`
			} `json:"keyParams"`
			CreatedAt string `json:"createdAt"`
		} `json:"keystores"`
	}

	if err := parseJSONResponse(resp, &result); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(result)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t类型\t算法\t签名\t加密\t创建时间")
	fmt.Fprintln(w, "----\t----\t----\t----\t----\t--------")

	for _, ks := range result.KeyStores {
		signStatus := "证/钥"
		if !ks.HasSignCert {
			signStatus = strings.Replace(signStatus, "证", "-", 1)
		}
		if !ks.HasSignKey {
			signStatus = strings.Replace(signStatus, "钥", "-", 1)
		}

		encStatus := "-"
		if ks.Type == "tlcp" {
			encStatus = "证/钥"
			if !ks.HasEncCert {
				encStatus = strings.Replace(encStatus, "证", "-", 1)
			}
			if !ks.HasEncKey {
				encStatus = strings.Replace(encStatus, "钥", "-", 1)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s/%d\t%s\t%s\t%s\n",
			ks.Name,
			strings.ToUpper(ks.Type),
			ks.KeyParams.Algorithm,
			ks.KeyParams.Length,
			signStatus,
			encStatus,
			ks.CreatedAt[:19],
		)
	}

	return w.Flush()
}

func keyStoreShow(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请指定密钥名称")
	}
	name := args[0]

	resp, err := cli.Get(fmt.Sprintf("/api/v1/keystores/%s", name))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ks struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		HasSignCert bool   `json:"hasSignCert"`
		HasSignKey  bool   `json:"hasSignKey"`
		HasEncCert  bool   `json:"hasEncCert"`
		HasEncKey   bool   `json:"hasEncKey"`
		KeyParams   struct {
			Algorithm string `json:"algorithm"`
			Length    int    `json:"length"`
			Type      string `json:"type"`
		} `json:"keyParams"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	if err := parseJSONResponse(resp, &ks); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(ks)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "名称:\t%s\n", ks.Name)
	fmt.Fprintf(w, "类型:\t%s\n", strings.ToUpper(ks.Type))
	fmt.Fprintf(w, "算法:\t%s/%d\n", ks.KeyParams.Algorithm, ks.KeyParams.Length)
	fmt.Fprintf(w, "签名证书:\t%s\n", boolToStatus(ks.HasSignCert))
	fmt.Fprintf(w, "签名密钥:\t%s\n", boolToStatus(ks.HasSignKey))
	if ks.Type == "tlcp" {
		fmt.Fprintf(w, "加密证书:\t%s\n", boolToStatus(ks.HasEncCert))
		fmt.Fprintf(w, "加密密钥:\t%s\n", boolToStatus(ks.HasEncKey))
	}
	fmt.Fprintf(w, "创建时间:\t%s\n", ks.CreatedAt)
	fmt.Fprintf(w, "更新时间:\t%s\n", ks.UpdatedAt)

	return w.Flush()
}

func keyStoreCreate(args []string) error {
	fs := flagSet("keystore create")
	fs.StringVar(&ksName, "name", "", "密钥名称")
	fs.StringVar(&ksType, "type", "tlcp", "类型 (tlcp/tls)")
	fs.StringVar(&ksSignCert, "sign-cert", "", "签名证书文件路径")
	fs.StringVar(&ksSignKey, "sign-key", "", "签名密钥文件路径")
	fs.StringVar(&ksEncCert, "enc-cert", "", "加密证书文件路径 (仅tlcp)")
	fs.StringVar(&ksEncKey, "enc-key", "", "加密密钥文件路径 (仅tlcp)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if ksName == "" {
		return fmt.Errorf("请指定 --name")
	}
	if ksSignCert == "" {
		return fmt.Errorf("请指定 --sign-cert")
	}
	if ksSignKey == "" {
		return fmt.Errorf("请指定 --sign-key")
	}
	if ksType == "tlcp" && ksEncCert == "" {
		return fmt.Errorf("国密类型请指定 --enc-cert")
	}
	if ksType == "tlcp" && ksEncKey == "" {
		return fmt.Errorf("国密类型请指定 --enc-key")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", ksName)
	_ = writer.WriteField("type", ksType)

	if err := addFileToForm(writer, "signCert", ksSignCert); err != nil {
		return err
	}
	if err := addFileToForm(writer, "signKey", ksSignKey); err != nil {
		return err
	}
	if ksType == "tlcp" && ksEncCert != "" {
		if err := addFileToForm(writer, "encCert", ksEncCert); err != nil {
			return err
		}
	}
	if ksType == "tlcp" && ksEncKey != "" {
		if err := addFileToForm(writer, "encKey", ksEncKey); err != nil {
			return err
		}
	}

	writer.Close()

	resp, err := cli.PostMultipart("/api/v1/keystores", writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Fprintf(os.Stderr, "密钥创建成功: %s\n", ksName)
		return nil
	}

	return fmt.Errorf("创建失败: %s", resp.Status)
}

func keyStoreUpdateCert(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请指定密钥名称")
	}
	name := args[0]

	fs := flagSet("keystore update-cert")
	fs.StringVar(&ksSignCert, "sign-cert", "", "签名证书文件路径")
	fs.StringVar(&ksEncCert, "enc-cert", "", "加密证书文件路径 (仅tlcp)")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if ksSignCert == "" && ksEncCert == "" {
		return fmt.Errorf("请至少指定一个证书文件")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if ksSignCert != "" {
		if err := addFileToForm(writer, "signCert", ksSignCert); err != nil {
			return err
		}
	}
	if ksEncCert != "" {
		if err := addFileToForm(writer, "encCert", ksEncCert); err != nil {
			return err
		}
	}

	writer.Close()

	resp, err := cli.PostMultipart(fmt.Sprintf("/api/v1/keystores/%s/certificates", name), writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Fprintf(os.Stderr, "证书更新成功: %s\n", name)
		return nil
	}

	return fmt.Errorf("更新失败: %s", resp.Status)
}

func keyStoreDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请指定密钥名称")
	}
	name := args[0]

	fmt.Fprintf(os.Stderr, "确定要删除密钥 %s 吗？(y/N): ", name)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		return fmt.Errorf("操作已取消")
	}

	resp, err := cli.Delete(fmt.Sprintf("/api/v1/keystores/%s", name))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Fprintf(os.Stderr, "密钥删除成功: %s\n", name)
		return nil
	}

	return fmt.Errorf("删除失败: %s", resp.Status)
}

func keyStoreReload(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请指定密钥名称")
	}
	name := args[0]

	resp, err := cli.Post(fmt.Sprintf("/api/v1/keystores/%s/reload", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Fprintf(os.Stderr, "密钥重载成功: %s\n", name)
		return nil
	}

	return fmt.Errorf("重载失败: %s", resp.Status)
}

func addFileToForm(writer *multipart.Writer, fieldName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败 %s: %w", filePath, err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	return err
}

func boolToStatus(b bool) string {
	if b {
		return "✓ 已上传"
	}
	return "✗ 未上传"
}
