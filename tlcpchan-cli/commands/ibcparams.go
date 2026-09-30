package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/Trisia/tlcpchan-cli/client"
)

// ibcParamsList 列出 IBC 信任池中的 KGC 公共参数
// 参数：
//   - args: 未使用的命令行参数
//
// 返回：
//   - error: 错误信息
func ibcParamsList(args []string) error {
	params, err := cli.ListIBCParams()
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(params)
	}

	if len(params) == 0 {
		fmt.Println("无 IBC 信任池参数")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "文件名\tKGC区域\t序号\t生效时间\t失效时间\t颁发者标识\t签名主公钥指纹\t加密主公钥指纹")
	for _, param := range params {
		issuer := param.IssuerIdentity
		if issuer == "" {
			issuer = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			param.Filename, param.DistrictName, param.DistrictSerial,
			formatIBCValidity(param.NotBefore), formatIBCValidity(param.NotAfter),
			issuer, shortFingerprint(param.SignKeyFingerprint), shortFingerprint(param.EncKeyFingerprint))
	}
	w.Flush()
	return nil
}

// formatIBCValidity 格式化公共参数有效期文本
// 参数：
//   - value: 服务端返回的 RFC3339 时间文本，空串或零值时间（0001-01-01）表示不限
//
// 返回：
//   - string: 截断到秒的时间文本；不限时返回 "-"
func formatIBCValidity(value string) string {
	if value == "" || strings.HasPrefix(value, "0001-01-01") {
		return "-"
	}
	return truncateTime(value)
}

// shortFingerprint 缩短 SM3 指纹以便表格展示，完整值可通过 -o json 查看
// 参数：
//   - fingerprint: 十六进制指纹文本
//
// 返回：
//   - string: 前 16 个字符加省略号；空串与短指纹原样返回
func shortFingerprint(fingerprint string) string {
	if fingerprint == "" {
		return "-"
	}
	if len(fingerprint) > 16 {
		return fingerprint[:16] + "..."
	}
	return fingerprint
}

// ibcParamsDownload 下载信任池中的 KGC 公共参数文件
// 参数：
//   - args: 命令行参数，格式为 <filename> [--output <路径>]
//
// 返回：
//   - error: 错误信息
func ibcParamsDownload(args []string) error {
	fs := flagSet("download")
	output := fs.String("output", "", "输出文件路径")
	if err := fs.Parse(reorderFlagsFirst(fs, args)); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("请指定 KGC 公共参数文件名")
	}
	filename := remaining[0]

	paramData, err := cli.DownloadIBCParam(filename)
	if err != nil {
		return err
	}

	if *output == "" {
		*output = filename
	}

	if err := os.WriteFile(*output, paramData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "IBC 信任池参数下载成功",
			"filename": *output,
		})
	}

	fmt.Printf("IBC 信任池参数已下载到 %s\n", *output)
	return nil
}

// ibcParamsAdd 添加 KGC 公共参数到 IBC 信任池
// 参数：
//   - args: 命令行参数，格式为 [--filename <文件名>] [--params <文件路径>]
//
// 返回：
//   - error: 错误信息
//
// 注意：公共参数必须通过可信渠道获取，信任池等价于根证书库
func ibcParamsAdd(args []string) error {
	fs := flagSet("add")
	filename := fs.String("filename", "", "信任池中保存的文件名（留空则使用参数文件名）")
	paramsFile := fs.String("params", "", "KGC 公共参数文件路径 (PEM/DER/HEX/Base64)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *paramsFile == "" {
		return fmt.Errorf("请指定 --params")
	}

	paramData, err := os.ReadFile(*paramsFile)
	if err != nil {
		return fmt.Errorf("读取 KGC 公共参数失败: %w", err)
	}

	saveName := *filename
	if saveName == "" {
		saveName = filepath.Base(*paramsFile)
	}

	param, err := cli.AddIBCParam(saveName, paramData)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "IBC 信任池参数添加成功",
			"filename": param.Filename,
		})
	}

	fmt.Printf("IBC 信任池参数 %s 添加成功\n", param.Filename)
	fmt.Println("提示: 修改信任池后需要重载相关实例才会生效 (tlcpchan-cli instance reload <name>)")
	return nil
}

// ibcParamsGenerate 生成测试 KGC 公共参数并加入信任池（仅使用内置测试主密钥）
// 参数：
//   - args: 命令行参数，格式为 [--district-name <区域>] [--district-serial <序号>] [--years <年>]
//
// 返回：
//   - error: 错误信息
//
// 注意：生产环境应由外部 KGC 派生后导入，本命令仅用于本地联调
func ibcParamsGenerate(args []string) error {
	fs := flagSet("generate")
	districtName := fs.String("district-name", "tlcpchan.local", "KGC 属地区域名")
	districtSerial := fs.Int("district-serial", 1, "同一区域下的 KGC 序号")
	years := fs.Int("years", 10, "公共参数有效期 (年)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req := client.GenerateIBCParamRequest{
		DistrictName:   *districtName,
		DistrictSerial: *districtSerial,
		Years:          *years,
	}

	param, err := cli.GenerateIBCParam(req)
	if err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "测试 KGC 公共参数生成成功",
			"filename": param.Filename,
		})
	}

	fmt.Printf("测试 KGC 公共参数 %s 生成成功\n", param.Filename)
	fmt.Println("提示: 修改信任池后需要重载相关实例才会生效 (tlcpchan-cli instance reload <name>)")
	return nil
}

// ibcParamsDelete 删除 IBC 信任池中的 KGC 公共参数
// 参数：
//   - args: 命令行参数，格式为 <filename>
//
// 返回：
//   - error: 错误信息
func ibcParamsDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定 KGC 公共参数文件名")
	}

	if err := cli.DeleteIBCParam(args[0]); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success":  true,
			"message":  "IBC 信任池参数已删除",
			"filename": args[0],
		})
	}

	fmt.Printf("IBC 信任池参数 %s 已删除\n", args[0])
	return nil
}

// ibcParamsReload 重新扫描目录并重建 IBC 信任池
// 参数：
//   - args: 未使用的命令行参数
//
// 返回：
//   - error: 错误信息
func ibcParamsReload(args []string) error {
	if err := cli.ReloadIBCParams(); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(map[string]interface{}{
			"success": true,
			"message": "IBC 信任池已重新加载",
		})
	}

	fmt.Println("IBC 信任池已重新加载")
	fmt.Println("提示: 已运行的实例需执行 instance reload 才会应用新的信任池")
	return nil
}
