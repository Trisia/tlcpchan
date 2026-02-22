package proxy

import (
	"testing"

	"github.com/Trisia/tlcpchan/config"
)

func TestValidateClientConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.InstanceConfig
		wantErr bool
	}{
		{
			name: "默认配置",
			cfg: &config.InstanceConfig{
				Protocol: string(config.ProtocolAuto),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want ProtocolType
	}{
		{"TLCP握手", []byte{0x16, 0x01, 0x01, 0x01, 0x01, 0x00}, ProtocolTLCP},
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
