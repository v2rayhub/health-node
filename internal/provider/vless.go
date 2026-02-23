package provider

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type vlessParser struct{}

func (p *vlessParser) Scheme() string { return "vless" }

func (p *vlessParser) Parse(u *url.URL, _ string) (Provider, error) {
	if u.User == nil {
		return nil, errors.New("vless URI missing user id")
	}
	id := u.User.Username()
	if id == "" {
		return nil, errors.New("vless URI has empty user id")
	}

	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("vless host/port parse failed: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid vless port: %w", err)
	}

	q := u.Query()
	return &VLESS{
		Address:     host,
		Port:        port,
		ID:          id,
		Flow:        q.Get("flow"),
		Encryption:  valueOrDefault(q.Get("encryption"), "none"),
		Network:     valueOrDefault(q.Get("type"), "tcp"),
		Security:    valueOrDefault(q.Get("security"), "none"),
		HeaderType:  q.Get("headerType"),
		Host:        q.Get("host"),
		Path:        q.Get("path"),
		SNI:         q.Get("sni"),
		ALPN:        q.Get("alpn"),
		Service:     q.Get("serviceName"),
		Fingerprint: q.Get("fp"),
		PublicKey:   q.Get("pbk"),
		ShortID:     q.Get("sid"),
		SpiderX:     q.Get("spx"),
		PQV:         q.Get("pqv"),
	}, nil
}

func (v *VLESS) Name() string { return "vless" }

func (v *VLESS) Outbound() (map[string]any, error) {
	user := map[string]any{
		"id":         v.ID,
		"encryption": valueOrDefault(v.Encryption, "none"),
	}
	if v.Flow != "" {
		user["flow"] = v.Flow
	}

	out := map[string]any{
		"tag":      "proxy",
		"protocol": "vless",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": v.Address,
				"port":    v.Port,
				"users":   []any{user},
			}},
		},
	}

	stream := map[string]any{
		"network":  v.Network,
		"security": v.Security,
	}
	if strings.EqualFold(v.Network, "tcp") && strings.EqualFold(v.HeaderType, "http") {
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
	if strings.EqualFold(v.Network, "grpc") {
		stream["grpcSettings"] = map[string]any{"serviceName": v.Service}
	}
	if strings.EqualFold(v.Security, "tls") {
		stream["tlsSettings"] = map[string]any{
			"serverName": firstNonEmpty(v.SNI, v.Host, v.Address),
			"alpn":       splitCSV(v.ALPN),
		}
	}
	if strings.EqualFold(v.Security, "reality") {
		reality := map[string]any{
			"fingerprint": valueOrDefault(v.Fingerprint, "chrome"),
			"serverName":  firstNonEmpty(v.SNI, v.Host, v.Address),
			"publicKey":   v.PublicKey,
			"shortId":     v.ShortID,
			"spiderX":     v.SpiderX,
		}
		if strings.TrimSpace(v.PQV) != "" {
			reality["mldsa65Verify"] = v.PQV
		}
		stream["realitySettings"] = reality
	}

	out["streamSettings"] = stream
	return out, nil
}
