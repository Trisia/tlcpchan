package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/Trisia/tlcpchan/security/rootcert"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListRootCertsInput 列出根证书输入（无参数）
type ListRootCertsInput struct{}

// ListRootCertsOutput 列出根证书输出
type ListRootCertsOutput struct {
	Certs []RootCertInfo `json:"certs"`
}

// RootCertInfo 根证书信息
type RootCertInfo struct {
	Filename     string   `json:"filename"`
	Subject      string   `json:"subject"`
	Issuer       string   `json:"issuer"`
	NotBefore    string   `json:"notBefore"`
	NotAfter     string   `json:"notAfter"`
	KeyType      string   `json:"keyType"`
	SerialNumber string   `json:"serialNumber"`
	Version      int      `json:"version"`
	IsCA         bool     `json:"isCA"`
	KeyUsage     []string `json:"keyUsage"`
}

// AddRootCertInput 添加根证书输入
type AddRootCertInput struct {
	Filename string `json:"filename"` // 证书文件名
	Content  string `json:"content"`  // 证书内容（Base64 或 PEM）
}

// AddRootCertOutput 添加根证书输出
type AddRootCertOutput struct {
	Filename     string   `json:"filename"`
	Subject      string   `json:"subject"`
	Issuer       string   `json:"issuer"`
	NotBefore    string   `json:"notBefore"`
	NotAfter     string   `json:"notAfter"`
	KeyType      string   `json:"keyType"`
	SerialNumber string   `json:"serialNumber"`
	Version      int      `json:"version"`
	IsCA         bool     `json:"isCA"`
	KeyUsage     []string `json:"keyUsage"`
}

// RemoveRootCertInput 删除根证书输入
type RemoveRootCertInput struct {
	Filename string `json:"filename"` // 证书文件名
}

// RemoveRootCertOutput 删除根证书输出
type RemoveRootCertOutput struct {
	Filename string `json:"filename"`
}

// registerRootCertTools 注册根证书管理工具
//
// 参数:
//   - c: MCP 控制器实例
//
// 注意:
//   - 注册 3 个根证书管理工具
func (c *MCPController) registerRootCertTools() {
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "list_root_certs",
		Description: "获取系统中的所有根证书列表",
	}, c.handleListRootCerts)

	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "add_root_cert",
		Description: "添加新的根证书到系统",
	}, c.handleAddRootCert)

	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "remove_root_cert",
		Description: "从系统中删除指定的根证书",
	}, c.handleRemoveRootCert)
}

// handleListRootCerts 处理获取获取根证书列表
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 输入参数（无）
//
// 返回:
//   - result: 工具结果（成功为 nil）
//   - output: 根证书列表输出
//   - error: 错误信息
//
// 注意:
//   - 调用 RootCertManager.List() 获取所有证书
//   - 将证书信息转换为输出格式
func (c *MCPController) handleListRootCerts(ctx context.Context, req *mcp.CallToolRequest, input ListRootCertsInput) (
	*mcp.CallToolResult,
	ListRootCertsOutput,
	error,
) {
	if c.rootCertMgr == nil {
		return nil, ListRootCertsOutput{}, fmt.Errorf("根证书管理器未初始化")
	}

	certs := c.rootCertMgr.List()
	output := ListRootCertsOutput{
		Certs: convertRootCertsToOutput(certs),
	}

	return nil, output, nil
}

// handleAddRootCert 处理添加根证书
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 输入参数（证书文件名和内容）
//
// 返回:
//   - result: 工具结果（成功为 nil）
//   - output: 添加的证书信息
//   - error: 错误信息
//
// 注意:
//   - 验证文件名和内容不为空
//   - 调用 RootCertManager.Add() 添加证书
//   - 获取添加的证书信息返回
func (c *MCPController) handleAddRootCert(ctx context.Context, req *mcp.CallToolRequest, input AddRootCertInput) (
	*mcp.CallToolResult,
	AddRootCertOutput,
	error,
) {
	if c.rootCertMgr == nil {
		return nil, AddRootCertOutput{}, fmt.Errorf("根证书管理器未初始化")
	}

	// 验证输入
	if input.Filename == "" {
		return nil, AddRootCertOutput{}, fmt.Errorf("证书文件名不能为空")
	}
	if input.Content == "" {
		return nil, AddRootCertOutput{}, fmt.Errorf("证书内容不能为空")
	}

	// 将 Base64 内容转换为字节
	certData := []byte(input.Content)

	// 添加证书
	rootCert, err := c.rootCertMgr.Add(input.Filename, certData)
	if err != nil {
		return nil, AddRootCertOutput{}, fmt.Errorf("添加根证书失败: %w", err)
	}

	if rootCert == nil {
		return nil, AddRootCertOutput{}, fmt.Errorf("添加证书成功但未返回证书信息")
	}

	certInfo := convertRootCertToOutput(rootCert)
	output := AddRootCertOutput{
		Filename:     certInfo.Filename,
		Subject:      certInfo.Subject,
		Issuer:       certInfo.Issuer,
		NotBefore:    certInfo.NotBefore,
		NotAfter:     certInfo.NotAfter,
		KeyType:      certInfo.KeyType,
		SerialNumber: certInfo.SerialNumber,
		Version:      certInfo.Version,
		IsCA:         certInfo.IsCA,
		KeyUsage:     certInfo.KeyUsage,
	}
	return nil, output, nil
}

// handleRemoveRootCert 处理删除根证书
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 输入参数（证书文件名）
//
// 返回:
//   - result: 工具结果（成功为 nil）
//   - output: 删除确认信息
//   - error: 错误信息
//
// 注意:
//   - 验证文件名不为空
//   - 调用 RootCertManager.Delete() 删除证书
func (c *MCPController) handleRemoveRootCert(ctx context.Context, req *mcp.CallToolRequest, input RemoveRootCertInput) (
	*mcp.CallToolResult,
	RemoveRootCertOutput,
	error,
) {
	if c.rootCertMgr == nil {
		return nil, RemoveRootCertOutput{}, fmt.Errorf("根证书管理器未初始化")
	}

	// 验证输入
	if input.Filename == "" {
		return nil, RemoveRootCertOutput{}, fmt.Errorf("证书文件名不能为空")
	}

	// 删除证书
	err := c.rootCertMgr.Delete(input.Filename)
	if err != nil {
		return nil, RemoveRootCertOutput{}, fmt.Errorf("删除根证书失败: %w", err)
	}

	output := RemoveRootCertOutput{
		Filename: input.Filename,
	}

	return nil, output, nil
}

// convertRootCertsToOutput 将根证书信息列表转换为输出格式
//
// 参数:
//   - certs: 根证书信息列表
//
// 返回:
//   - []RootCertInfo: 转换后的证书信息列表
//
// 注意:
//   - 提取证书的各个字段并转换为字符串格式
func convertRootCertsToOutput(certs []*rootcert.RootCert) []RootCertInfo {
	output := make([]RootCertInfo, len(certs))
	for i, cert := range certs {
		output[i] = convertRootCertToOutput(cert)
	}
	return output
}

// convertRootCertToOutput 将单个根证书信息转换为输出格式
//
// 参数:
//   - cert: 根证书信息
//
// 返回:
//   - RootCertInfo: 转换后的证书信息
func convertRootCertToOutput(cert *rootcert.RootCert) RootCertInfo {
	return RootCertInfo{
		Filename:     cert.Filename,
		Subject:      cert.Subject,
		Issuer:       cert.Issuer,
		NotBefore:    timeToString(cert.NotBefore),
		NotAfter:     timeToString(cert.NotAfter),
		KeyType:      cert.KeyType,
		SerialNumber: cert.SerialNumber,
		Version:      cert.Version,
		IsCA:         cert.IsCA,
		KeyUsage:     cert.KeyUsage,
	}
}

// timeToString 将 time.Time 转换为 ISO 8601 格式字符串
//
// 参数:
//   - t: 时间对象，零时间表示空
//
// 返回:
//   - string: ISO 格式时间字符串，零时返回空字符串
//
// 注意:
//   - 使用 "2006-01-02T15:04:05Z" 格式
func timeToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}
