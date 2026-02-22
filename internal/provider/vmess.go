package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type vmessParser struct{}

func (p *vmessParser) Scheme() string { return "vmess" }

func (p *vmessParser) Parse(_ *url.URL, raw string) (Provider, error) {
	const prefix = "vmess://"
	payload := strings.TrimPrefix(raw, prefix)
	if payload == raw {
		return nil, errors.New("invalid vmess URI")
	}

	decoded, err := decodeBase64Any(payload)
	if err != nil {
		return nil, fmt.Errorf("vmess base64 decode failed: %w", err)
	}

	var vm VMess
	if err := json.Unmarshal(decoded, &vm); err != nil {
		return nil, fmt.Errorf("vmess JSON decode failed: %w", err)
	}

	port, err := parseIntField(vm.PortRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid vmess port: %w", err)
	}
	if vm.Address == "" || vm.ID == "" {
		return nil, errors.New("vmess JSON missing add/port/id")
	}
	vm.Port = port
	if vm.Port <= 0 || vm.Port > 65535 {
		return nil, errors.New("vmess port out of range")
	}
	if len(vm.AlterIDRaw) > 0 {
		if aid, err := parseIntField(vm.AlterIDRaw); err == nil {
			vm.AlterID = aid
		}
	}
	if vm.Network == "" {
		vm.Network = "tcp"
	}
	if vm.Security == "" {
		vm.Security = "auto"
	}
	return &vm, nil
}

func (v *VMess) Name() string { return "vmess" }

func (v *VMess) Outbound() (map[string]any, error) {
	out := map[string]any{
		"tag":      "proxy",
		"protocol": "vmess",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": v.Address,
				"port":    v.Port,
				"users": []any{map[string]any{
					"id":       v.ID,
					"alterId":  v.AlterID,
					"security": valueOrDefault(v.Security, "auto"),
				}},
			}},
		},
	}

	security := "none"
	if strings.EqualFold(v.TLS, "tls") {
		security = "tls"
	}
	stream := map[string]any{
		"network":  valueOrDefault(v.Network, "tcp"),
		"security": security,
	}
	if strings.EqualFold(v.Network, "tcp") && strings.EqualFold(v.Type, "http") {
		request := map[string]any{"path": toPathList(v.Path)}
		if hosts := toHostList(v.Host); len(hosts) > 0 {
			request["headers"] = map[string]any{"Host": hosts}
		}
		stream["tcpSettings"] = map[string]any{
			"header": map[string]any{
				"type":    "http",
				"request": request,
			},
		}
	}
	if strings.EqualFold(v.Network, "ws") {
		ws := map[string]any{"path": valueOrDefault(v.Path, "/")}
		if strings.TrimSpace(v.Host) != "" {
			ws["headers"] = map[string]any{"Host": v.Host}
		}
		stream["wsSettings"] = ws
	}
	if security == "tls" {
		stream["tlsSettings"] = map[string]any{
			"serverName": firstNonEmpty(v.SNI, v.Host, v.Address),
			"alpn":       splitCSV(v.ALPN),
		}
	}
	out["streamSettings"] = stream
	return out, nil
}
