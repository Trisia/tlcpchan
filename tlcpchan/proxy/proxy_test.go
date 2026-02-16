package proxy

import (
	"testing"

	"github.com/Trisia/tlcpchan/config"
)

func TestParseProtocolType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ProtocolType
	}{
		{"auto默认", "", ProtocolAuto},
		{"auto", "auto", ProtocolAuto},
		{"tlcp", "tlcp", ProtocolTLCP},
		{"tls", "tls", ProtocolTLS},
		{"未知类型", "unknown", ProtocolAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseProtocolType(tt.input)
			if got != tt.want {
				t.Errorf("ParseProtocolType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestProtocolType_String(t *testing.T) {
	tests := []struct {
		name string
		p    ProtocolType
		want string
	}{
		{"auto", ProtocolAuto, "auto"},
		{"tlcp", ProtocolTLCP, "tlcp"},
		{"tls", ProtocolTLS, "tls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("ProtocolType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateClientConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.InstanceConfig
		wantErr bool
	}{
		{
			name: "空配置",
			cfg: &config.InstanceConfig{
				Protocol: string(config.ProtocolAuto),
			},
			wantErr: false,
		},
		{
			name: "ECDHE密码套件需要双向认证-tlcp",
			cfg: &config.InstanceConfig{
				Protocol: string(config.ProtocolTLCP),
				TLCP: config.TLCPConfig{
					Auth:         string(config.AuthOneWay),
					CipherSuites: []string{"ECDHE_SM4_CBC_SM3"},
				},
			},
			wantErr: true,
		},
		{
			name: "ECDHE密码套件在双向认证下-tlcp",
			cfg: &config.InstanceConfig{
				Protocol: string(config.ProtocolTLCP),
				TLCP: config.TLCPConfig{
					Auth:         string(config.AuthMutual),
					CipherSuites: []string{"ECDHE_SM4_CBC_SM3"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateClientConfig(tt.cfg); (err != nil) != tt.wantErr {
				t.Errorf("ValidateClientConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want ProtocolType
	}{
		{"TLCP握手", []byte{0x16, 0x03, 0x01, 0x01, 0x01, 0x00}, ProtocolTLCP},
		{"TLS 1.0握手", []byte{0x16, 0x03, 0x01, 0x03, 0x01, 0x00}, ProtocolTLS},
		{"TLS 1.2握手", []byte{0x16, 0x03, 0x01, 0x03, 0x03, 0x00}, ProtocolTLS},
		{"数据太短", []byte{0x16, 0x03}, ProtocolTLS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectProtocol(tt.data); got != tt.want {
				t.Errorf("detectProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsECDHECipherSuite(t *testing.T) {
	tests := []struct {
		name   string
		suites []uint16
		want   bool
	}{
		{"空列表", []uint16{}, false},
		{"非ECDHE", []uint16{0xC011}, false},
		{"ECDHE_SM4_CBC_SM3", []uint16{0xC013}, true},
		{"ECDHE_SM4_GCM_SM3", []uint16{0xC014}, true},
		{"ECDHE_SM4_CCM_SM3", []uint16{0xC01A}, true},
		{"混合包含ECDHE", []uint16{0xC011, 0xC013}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isECDHECipherSuite(tt.suites); got != tt.want {
				t.Errorf("isECDHECipherSuite() = %v, want %v", got, tt.want)
			}
		})
	}
}
