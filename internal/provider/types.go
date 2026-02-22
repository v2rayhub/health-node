package provider

import "encoding/json"

// Provider builds outbound configs for V2Ray-compatible cores.
type Provider interface {
	Outbound() (map[string]any, error)
	Name() string
}

type VLESS struct {
	Address     string
	Port        int
	ID          string
	Flow        string
	Encryption  string
	Network     string
	Security    string
	HeaderType  string
	Host        string
	Path        string
	SNI         string
	ALPN        string
	Service     string
	Fingerprint string
	PublicKey   string
	ShortID     string
	SpiderX     string
	PQV         string
}

type VMess struct {
	Address    string          `json:"add"`
	PortRaw    json.RawMessage `json:"port"`
	ID         string          `json:"id"`
	AlterIDRaw json.RawMessage `json:"aid"`
	Network    string          `json:"net"`
	Host       string          `json:"host"`
	Path       string          `json:"path"`
	TLS        string          `json:"tls"`
	SNI        string          `json:"sni"`
	ALPN       string          `json:"alpn"`
	Type       string          `json:"type"`
	Security   string          `json:"scy"`
	Port       int             `json:"-"`
	AlterID    int             `json:"-"`
}

type Shadowsocks struct {
	Address    string
	Port       int
	Method     string
	Password   string
	Network    string
	Security   string
	HeaderType string
	Host       string
	Path       string
	SNI        string
	ALPN       string
	Service    string
}
