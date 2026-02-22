package provider

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestFromURI_VMess_PortAndAidNumeric(t *testing.T) {
	raw := vmessURI(t, map[string]any{
		"v":    "2",
		"add":  "example.com",
		"port": 443,
		"id":   "2d67b1be-5e23-40b0-a826-4fd8dd4e650f",
		"aid":  0,
		"net":  "tcp",
		"tls":  "tls",
		"type": "http",
		"path": "/search",
		"scy":  "auto",
	})

	p, err := FromURI(raw)
	if err != nil {
		t.Fatalf("FromURI() error = %v", err)
	}
	vm, ok := p.(*VMess)
	if !ok {
		t.Fatalf("provider type = %T, want *VMess", p)
	}
	if vm.Port != 443 {
		t.Fatalf("vm.Port = %d, want 443", vm.Port)
	}
	if vm.AlterID != 0 {
		t.Fatalf("vm.AlterID = %d, want 0", vm.AlterID)
	}
}

func TestFromURI_VMess_PortAndAidString(t *testing.T) {
	raw := vmessURI(t, map[string]any{
		"v":    "2",
		"add":  "example.com",
		"port": "443",
		"id":   "2d67b1be-5e23-40b0-a826-4fd8dd4e650f",
		"aid":  "1",
		"net":  "tcp",
		"tls":  "tls",
		"type": "http",
		"path": "/search",
		"scy":  "auto",
	})

	p, err := FromURI(raw)
	if err != nil {
		t.Fatalf("FromURI() error = %v", err)
	}
	vm, ok := p.(*VMess)
	if !ok {
		t.Fatalf("provider type = %T, want *VMess", p)
	}
	if vm.Port != 443 {
		t.Fatalf("vm.Port = %d, want 443", vm.Port)
	}
	if vm.AlterID != 1 {
		t.Fatalf("vm.AlterID = %d, want 1", vm.AlterID)
	}
}

func TestVMessOutbound_TCPHTTP_NoHostHeaderWhenEmpty(t *testing.T) {
	vm := &VMess{
		Address:  "example.com",
		Port:     443,
		ID:       "2d67b1be-5e23-40b0-a826-4fd8dd4e650f",
		Network:  "tcp",
		Type:     "http",
		Path:     "/search",
		TLS:      "tls",
		Security: "auto",
	}

	out, err := vm.Outbound()
	if err != nil {
		t.Fatalf("Outbound() error = %v", err)
	}

	stream := mustMap(t, out["streamSettings"])
	tcp := mustMap(t, stream["tcpSettings"])
	header := mustMap(t, tcp["header"])
	req := mustMap(t, header["request"])
	if _, ok := req["headers"]; ok {
		t.Fatalf("request.headers exists; want omitted when host is empty")
	}
}

func TestVMessOutbound_TCPHTTP_HostHeaderIncluded(t *testing.T) {
	vm := &VMess{
		Address:  "example.com",
		Port:     443,
		ID:       "2d67b1be-5e23-40b0-a826-4fd8dd4e650f",
		Network:  "tcp",
		Type:     "http",
		Path:     "/search",
		Host:     "a.example.com,b.example.com",
		TLS:      "tls",
		Security: "auto",
	}

	out, err := vm.Outbound()
	if err != nil {
		t.Fatalf("Outbound() error = %v", err)
	}

	stream := mustMap(t, out["streamSettings"])
	tcp := mustMap(t, stream["tcpSettings"])
	header := mustMap(t, tcp["header"])
	req := mustMap(t, header["request"])
	headers := mustMap(t, req["headers"])
	hosts := mustStringSlice(t, headers["Host"])
	if len(hosts) != 2 || hosts[0] != "a.example.com" || hosts[1] != "b.example.com" {
		t.Fatalf("Host headers = %#v, want [a.example.com b.example.com]", hosts)
	}
}

func TestFromURI_VLESS_RealityFieldsMapped(t *testing.T) {
	raw := "vless://80cbb58b-74c0-4fb5-a66e-818ffc81a3cd@example.com:443?type=tcp&headerType=http&security=reality&pbk=PUBKEY123&sid=abcd1234&fp=chrome&sni=aparat.com&spx=%2F&pqv=VERIFY123&encryption=none&path=%2Ftest"

	p, err := FromURI(raw)
	if err != nil {
		t.Fatalf("FromURI() error = %v", err)
	}
	v, ok := p.(*VLESS)
	if !ok {
		t.Fatalf("provider type = %T, want *VLESS", p)
	}
	if v.Security != "reality" {
		t.Fatalf("v.Security = %q, want reality", v.Security)
	}
	if v.PublicKey != "PUBKEY123" {
		t.Fatalf("v.PublicKey = %q, want PUBKEY123", v.PublicKey)
	}
	if v.ShortID != "abcd1234" {
		t.Fatalf("v.ShortID = %q, want abcd1234", v.ShortID)
	}
	if v.Fingerprint != "chrome" {
		t.Fatalf("v.Fingerprint = %q, want chrome", v.Fingerprint)
	}
}

func TestVLESSOutbound_RealitySettingsPresent(t *testing.T) {
	v := &VLESS{
		Address:     "example.com",
		Port:        443,
		ID:          "80cbb58b-74c0-4fb5-a66e-818ffc81a3cd",
		Encryption:  "none",
		Network:     "tcp",
		HeaderType:  "http",
		Path:        "/test",
		Security:    "reality",
		SNI:         "aparat.com",
		Fingerprint: "chrome",
		PublicKey:   "PUBKEY123",
		ShortID:     "abcd1234",
		SpiderX:     "/",
		PQV:         "VERIFY123",
	}

	out, err := v.Outbound()
	if err != nil {
		t.Fatalf("Outbound() error = %v", err)
	}
	stream := mustMap(t, out["streamSettings"])
	reality := mustMap(t, stream["realitySettings"])

	if got := reality["serverName"]; got != "aparat.com" {
		t.Fatalf("reality.serverName = %#v, want aparat.com", got)
	}
	if got := reality["password"]; got != "PUBKEY123" {
		t.Fatalf("reality.password = %#v, want PUBKEY123", got)
	}
	if got := reality["shortId"]; got != "abcd1234" {
		t.Fatalf("reality.shortId = %#v, want abcd1234", got)
	}
	if got := reality["fingerprint"]; got != "chrome" {
		t.Fatalf("reality.fingerprint = %#v, want chrome", got)
	}
	if got := reality["mldsa65Verify"]; got != "VERIFY123" {
		t.Fatalf("reality.mldsa65Verify = %#v, want VERIFY123", got)
	}
}

func TestFromURI_VLESS_Minimal(t *testing.T) {
	raw := "vless://80cbb58b-74c0-4fb5-a66e-818ffc81a3cd@example.com:443"
	p, err := FromURI(raw)
	if err != nil {
		t.Fatalf("FromURI() error = %v", err)
	}
	v, ok := p.(*VLESS)
	if !ok {
		t.Fatalf("provider type = %T, want *VLESS", p)
	}
	if v.Address != "example.com" {
		t.Fatalf("v.Address = %q, want example.com", v.Address)
	}
	if v.Port != 443 {
		t.Fatalf("v.Port = %d, want 443", v.Port)
	}
	if v.Network != "tcp" {
		t.Fatalf("v.Network = %q, want tcp", v.Network)
	}
	if v.Security != "none" {
		t.Fatalf("v.Security = %q, want none", v.Security)
	}
}

func TestFromURI_VLESS_InvalidHostPort(t *testing.T) {
	raw := "vless://80cbb58b-74c0-4fb5-a66e-818ffc81a3cd@example.com"
	_, err := FromURI(raw)
	if err == nil {
		t.Fatal("FromURI() expected error for missing port")
	}
	if !strings.Contains(err.Error(), "host/port") {
		t.Fatalf("error = %v, want host/port parse error", err)
	}
}

func TestFromURI_VMess_InvalidBase64(t *testing.T) {
	_, err := FromURI("vmess://not-base64")
	if err == nil {
		t.Fatal("FromURI() expected base64 decode error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("error = %v, want decode error", err)
	}
}

func TestFromURI_VMess_MissingFields(t *testing.T) {
	raw := vmessURI(t, map[string]any{
		"v":   "2",
		"add": "example.com",
	})
	_, err := FromURI(raw)
	if err == nil {
		t.Fatal("FromURI() expected missing fields error")
	}
	if !strings.Contains(err.Error(), "invalid vmess port") {
		t.Fatalf("error = %v, want invalid vmess port error", err)
	}
}

func TestFromURI_VMess_PortOutOfRange(t *testing.T) {
	raw := vmessURI(t, map[string]any{
		"v":    "2",
		"add":  "example.com",
		"port": 70000,
		"id":   "2d67b1be-5e23-40b0-a826-4fd8dd4e650f",
	})
	_, err := FromURI(raw)
	if err == nil {
		t.Fatal("FromURI() expected out-of-range port error")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error = %v, want out-of-range error", err)
	}
}

func TestFromURI_UnsupportedScheme(t *testing.T) {
	_, err := FromURI("trojan://example")
	if err == nil {
		t.Fatal("FromURI() expected unsupported scheme error")
	}
	if !strings.Contains(err.Error(), "unsupported scheme") {
		t.Fatalf("error = %v, want unsupported scheme error", err)
	}
}

func TestFromURI_Shadowsocks_Base64UserInfo(t *testing.T) {
	raw := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388?type=tcp&security=tls&sni=aparat.com"
	p, err := FromURI(raw)
	if err != nil {
		t.Fatalf("FromURI() error = %v", err)
	}
	ss, ok := p.(*Shadowsocks)
	if !ok {
		t.Fatalf("provider type = %T, want *Shadowsocks", p)
	}
	if ss.Method != "aes-256-gcm" {
		t.Fatalf("ss.Method = %q, want aes-256-gcm", ss.Method)
	}
	if ss.Password != "password" {
		t.Fatalf("ss.Password = %q, want password", ss.Password)
	}
	if ss.Port != 8388 {
		t.Fatalf("ss.Port = %d, want 8388", ss.Port)
	}
}

func TestFromURI_Shadowsocks_Invalid(t *testing.T) {
	_, err := FromURI("ss://invalid@example.com:8388")
	if err == nil {
		t.Fatal("FromURI() expected error for invalid ss credentials")
	}
}

func TestShadowsocksOutbound_TLS(t *testing.T) {
	ss := &Shadowsocks{
		Address:  "example.com",
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "password",
		Network:  "tcp",
		Security: "tls",
		SNI:      "aparat.com",
		ALPN:     "h2,http/1.1",
	}
	out, err := ss.Outbound()
	if err != nil {
		t.Fatalf("Outbound() error = %v", err)
	}
	stream := mustMap(t, out["streamSettings"])
	if got := stream["security"]; got != "tls" {
		t.Fatalf("stream.security = %#v, want tls", got)
	}
	tls := mustMap(t, stream["tlsSettings"])
	if got := tls["serverName"]; got != "aparat.com" {
		t.Fatalf("tls.serverName = %#v, want aparat.com", got)
	}
}

func vmessURI(t *testing.T, payload map[string]any) string {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(b)
}

func mustMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("value type = %T, want map[string]any", v)
	}
	return m
}

func mustStringSlice(t *testing.T, v any) []string {
	t.Helper()
	raw, ok := v.([]string)
	if ok {
		return raw
	}

	ai, ok := v.([]any)
	if !ok {
		t.Fatalf("value type = %T, want []string or []any", v)
	}
	out := make([]string, 0, len(ai))
	for _, e := range ai {
		s, ok := e.(string)
		if !ok {
			t.Fatalf("slice elem type = %T, want string", e)
		}
		out = append(out, s)
	}
	return out
}
