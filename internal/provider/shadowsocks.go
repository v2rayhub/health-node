package provider

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type shadowsocksParser struct{}

func (p *shadowsocksParser) Scheme() string { return "ss" }

func (p *shadowsocksParser) Parse(u *url.URL, raw string) (Provider, error) {
	server := u.Host
	credPart := ""
	if strings.Contains(raw, "@") {
		trimmed := strings.TrimPrefix(raw, "ss://")
		if i := strings.Index(trimmed, "#"); i >= 0 {
			trimmed = trimmed[:i]
		}
		if i := strings.Index(trimmed, "?"); i >= 0 {
			trimmed = trimmed[:i]
		}
		at := strings.LastIndex(trimmed, "@")
		if at > 0 {
			credPart = trimmed[:at]
			server = trimmed[at+1:]
		}
	}

	method := ""
	password := ""
	if credPart != "" {
		if b, err := decodeBase64Any(credPart); err == nil {
			if m, p, ok := strings.Cut(string(b), ":"); ok {
				method, password = m, p
			}
		}
		if method == "" {
			if m, p, ok := strings.Cut(credPart, ":"); ok {
				method, password = m, p
			}
		}
	} else if u.User != nil {
		user := u.User.Username()
		if b, err := decodeBase64Any(user); err == nil {
			if m, p, ok := strings.Cut(string(b), ":"); ok {
				method, password = m, p
			}
		}
	}
	if method == "" || password == "" {
		return nil, errors.New("ss URI missing method/password")
	}

	host, portStr, err := net.SplitHostPort(server)
	if err != nil {
		return nil, fmt.Errorf("ss host/port parse failed: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("invalid ss port")
	}

	q := u.Query()
	return &Shadowsocks{
		Address:    host,
		Port:       port,
		Method:     method,
		Password:   password,
		Network:    valueOrDefault(q.Get("type"), "tcp"),
		Security:   valueOrDefault(q.Get("security"), "none"),
		HeaderType: q.Get("headerType"),
		Host:       q.Get("host"),
		Path:       q.Get("path"),
		SNI:        q.Get("sni"),
		ALPN:       q.Get("alpn"),
		Service:    q.Get("serviceName"),
	}, nil
}

func (s *Shadowsocks) Name() string { return "shadowsocks" }

func (s *Shadowsocks) Outbound() (map[string]any, error) {
	out := map[string]any{
		"tag":      "proxy",
		"protocol": "shadowsocks",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address":  s.Address,
				"port":     s.Port,
				"method":   s.Method,
				"password": s.Password,
			}},
		},
	}

	stream := map[string]any{
		"network":  valueOrDefault(s.Network, "tcp"),
		"security": valueOrDefault(s.Security, "none"),
	}
	if strings.EqualFold(s.Network, "tcp") && strings.EqualFold(s.HeaderType, "http") {
		request := map[string]any{"path": toPathList(s.Path)}
		if hosts := toHostList(s.Host); len(hosts) > 0 {
			request["headers"] = map[string]any{"Host": hosts}
		}
		stream["tcpSettings"] = map[string]any{
			"header": map[string]any{
				"type":    "http",
				"request": request,
			},
		}
	}
	if strings.EqualFold(s.Network, "ws") {
		ws := map[string]any{"path": valueOrDefault(s.Path, "/")}
		if strings.TrimSpace(s.Host) != "" {
			ws["headers"] = map[string]any{"Host": s.Host}
		}
		stream["wsSettings"] = ws
	}
	if strings.EqualFold(s.Network, "grpc") {
		stream["grpcSettings"] = map[string]any{"serviceName": s.Service}
	}
	if strings.EqualFold(s.Security, "tls") {
		stream["tlsSettings"] = map[string]any{
			"serverName": firstNonEmpty(s.SNI, s.Host, s.Address),
			"alpn":       splitCSV(s.ALPN),
		}
	}

	out["streamSettings"] = stream
	return out, nil
}
