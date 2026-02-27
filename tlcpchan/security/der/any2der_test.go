package der

import (
	"os"
	"testing"
)

func TestAny2DER(t *testing.T) {
	tlsEcdsaCertDER, err := os.ReadFile("testdata/tls-ecdsa/cert.der")
	if err != nil {
		t.Fatalf("加载 TLS ECDSA 证书 DER 失败: %v", err)
	}

	tlsEcdsaPrivDER, err := os.ReadFile("testdata/tls-ecdsa/priv.der")
	if err != nil {
		t.Fatalf("加载 TLS ECDSA 私钥 DER 失败: %v", err)
	}

	tlsRsaCertDER, err := os.ReadFile("testdata/tls-rsa/cert.der")
	if err != nil {
		t.Fatalf("加载 TLS RSA 证书 DER 失败: %v", err)
	}

	tlsRsaPrivDER, err := os.ReadFile("testdata/tls-rsa/priv.der")
	if err != nil {
		t.Fatalf("加载 TLS RSA 私钥 DER 失败: %v", err)
	}

	tlcpSignCertDER, err := os.ReadFile("testdata/tlcp-sign/cert.der")
	if err != nil {
		t.Fatalf("加载 TLCP 签名证书 DER 失败: %v", err)
	}

	tlcpSignPrivDER, err := os.ReadFile("testdata/tlcp-sign/priv.der")
	if err != nil {
		t.Fatalf("加载 TLCP 签名私钥 DER 失败: %v", err)
	}

	tlcpEncCertDER, err := os.ReadFile("testdata/tlcp-enc/cert.der")
	if err != nil {
		t.Fatalf("加载 TLCP 加密证书 DER 失败: %v", err)
	}

	tlcpEncPrivDER, err := os.ReadFile("testdata/tlcp-enc/priv.der")
	if err != nil {
		t.Fatalf("加载 TLCP 加密私钥 DER 失败: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		want    []byte
		wantErr bool
	}{
		{
			name:    "TLS ECDSA PEM 证书",
			file:    "testdata/tls-ecdsa/cert.pem",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA PEM 私钥",
			file:    "testdata/tls-ecdsa/priv.pem",
			want:    tlsEcdsaPrivDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA DER 证书",
			file:    "testdata/tls-ecdsa/cert.der",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA HEX 证书",
			file:    "testdata/tls-ecdsa/cert.hex",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA HEX 证书带 0x 前缀",
			file:    "testdata/tls-ecdsa/cert.hex0x",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA HEX 证书带空格",
			file:    "testdata/tls-ecdsa/cert.hex-sp",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA Base64 证书",
			file:    "testdata/tls-ecdsa/cert.base64",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS ECDSA Base64 证书带换换行",
			file:    "testdata/tls-ecdsa/cert.base64-nl",
			want:    tlsEcdsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS RSA PEM 证书",
			file:    "testdata/tls-rsa/cert.pem",
			want:    tlsRsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS RSA PEM 私钥",
			file:    "testdata/tls-rsa/priv.pem",
			want:    tlsRsaPrivDER,
			wantErr: false,
		},
		{
			name:    "TLS RSA DER 证书",
			file:    "testdata/tls-rsa/cert.der",
			want:    tlsRsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS RSA HEX 证书",
			file:    "testdata/tls-rsa/cert.hex",
			want:    tlsRsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLS RSA Base64 证书",
			file:    "testdata/tls-rsa/cert.base64",
			want:    tlsRsaCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 签名 PEM 证书",
			file:    "testdata/tlcp-sign/cert.pem",
			want:    tlcpSignCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 签名 PEM 私钥",
			file:    "testdata/tlcp-sign/priv.pem",
			want:    tlcpSignPrivDER,
			wantErr: false,
		},
		{
			name:    "TLCP 签名 DER 证书",
			file:    "testdata/tlcp-sign/cert.der",
			want:    tlcpSignCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 签名 HEX 证书",
			file:    "testdata/tlcp-sign/cert.hex",
			want:    tlcpSignCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 签名 Base64 证书",
			file:    "testdata/tlcp-sign/cert.base64",
			want:    tlcpSignCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 加密 PEM 证书",
			file:    "testdata/tlcp-enc/cert.pem",
			want:    tlcpEncCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 加密 PEM 私钥",
			file:    "testdata/tlcp-enc/priv.pem",
			want:    tlcpEncPrivDER,
			wantErr: false,
		},
		{
			name:    "TLCP 加密 DER 证书",
			file:    "testdata/tlcp-enc/cert.der",
			want:    tlcpEncCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 加密 HEX 证书",
			file:    "testdata/tlcp-enc/cert.hex",
			want:    tlcpEncCertDER,
			wantErr: false,
		},
		{
			name:    "TLCP 加密 Base64 证书",
			file:    "testdata/tlcp-enc/cert.base64",
			want:    tlcpEncCertDER,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("读取文件失败: %v", err)
			}

			got, err := Any2DER(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Any2DER() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalBytes(got, tt.want) {
				t.Errorf("Any2DER() 结果不匹配")
			}
		})
	}
}

func TestAny2DER_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"空数据", []byte{}},
		{"nil 数据", nil},
		{"无效格式", []byte("invalid data")},
		{"无效 HEX", []byte("not hex")},
		{"无效 Base64", []byte("!!!invalid base64!!!")},
		{"不完整的 DER", []byte{0x30, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Any2DER(tt.data)
			if err == nil {
				t.Error("Any2DER() 期望返回错误，但没有返回")
			}
		})
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
