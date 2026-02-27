package der

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strings"
)

// Any2DER 尝试将输入数据解析为 DER 格式
// 支持的格式：PEM、HEX、Base64、DER
//
// 参数:
//   - data: 输入数据，可以是 PEM、HEX、Base64 或 DER 格式
//
// 返回:
//   - []byte: 解析后的 DER 数据
//   - error: 如果所有解析方式都失败则返回错误
//
// 解析顺序：
//  1. PEM 格式 - 检查 PEM 前綴，提取 PEM block 的 Bytes
//  2. HEX 格式 - 检查是否为有效的十六进制字符串（可選 0x 前綴）
//  3. Base64 格式 - 檢查是否為有效的 Base64 編碼
//  4. DER 格式 - 檢查 ASN.1 SEQUENCE 前綴（0x30）
//  5. 失敗 - 返回錯誤
func Any2DER(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	// 1. 尝试 PEM 格式
	if block, _ := pem.Decode(data); block != nil && len(block.Bytes) > 0 {
		return block.Bytes, nil
	}

	// 2. 尝试 HEX 格式
	if der, err := tryParseHex(data); err == nil && der != nil {
		return der, nil
	}

	// 3. 尝试 Base64 格式
	if der, err := tryParseBase64(data); err == nil && der != nil {
		return der, nil
	}

	// 4. 尝试 DER 格式（检查 ASN.1 SEQUENCE 前綴 0x30）
	if isDER(data) {
		return data, nil
	}

	return nil, errors.New("无法识别数据格式")
}

// tryParseHex 尝试解析十六进制格式
func tryParseHex(data []byte) ([]byte, error) {
	str := strings.TrimSpace(string(data))

	// 去除 0x 前缀
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}

	// 去除所有空白字符
	str = strings.Join(strings.Fields(str), "")

	if len(str) == 0 {
		return nil, errors.New("十六进制字符串为空")
	}

	// 检查是否为有效的十六进制字符串
	for _, c := range str {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return nil, errors.New("无效的十六进制字符")
		}
	}

	// 检查长度是否为偶数
	if len(str)%2 != 0 {
		return nil, errors.New("十六进制字符串长度必须为偶数")
	}

	return hex.DecodeString(str)
}

// tryParseBase64 尝试解析 Base64 格式
func tryParseBase64(data []byte) ([]byte, error) {
	str := strings.TrimSpace(string(data))

	// 去除所有空白字符
	str = strings.Join(strings.Fields(str), "")

	if len(str) == 0 {
		return nil, errors.New("Base64 字符串为空")
	}

	// 尝试标准 Base64
	if der, err := base64.StdEncoding.DecodeString(str); err == nil {
		return der, nil
	}

	// 尝试 URL 安全 Base64
	if der, err := base64.URLEncoding.DecodeString(str); err == nil {
		return der, nil
	}

	return nil, errors.New("无效的 Base64 编码")
}

// isDER 检查数据是否为 DER 格式
func isDER(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// DER 格式的第一个字节应该是 0x30（ASN.1 SEQUENCE）
	if data[0] != 0x30 {
		return false
	}

	// 检查长度字节
	lengthByte := int(data[1])
	if lengthByte < 0x80 {
		// 短格式：长度直接存储在第二个字节
		if len(data) < 2+lengthByte {
			return false
		}
	} else if lengthByte == 0x81 {
		// 长格式：长度存储在下一个字节
		if len(data) < 3 {
			return false
		}
		length := int(data[2])
		if len(data) < 3+length {
			return false
		}
	} else if lengthByte == 0x82 {
		// 长格式：长度存储在接下来的两个字节
		if len(data) < 4 {
			return false
		}
		length := int(data[2])<<8 | int(data[3])
		if len(data) < 4+length {
			return false
		}
	} else {
		// 不支持的长度格式
		return false
	}

	return true
}
