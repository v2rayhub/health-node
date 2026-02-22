package provider

import (
	"encoding/base64"
	"encoding/json"
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
