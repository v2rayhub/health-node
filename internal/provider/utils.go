package provider

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func parseIntField(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, errors.New("empty field")
	}

	var asInt int
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}

	var asStr string
	if err := json.Unmarshal(raw, &asStr); err == nil {
		asStr = strings.TrimSpace(asStr)
		if asStr == "" {
			return 0, errors.New("empty string")
		}
		n, err := strconv.Atoi(asStr)
		if err != nil {
			return 0, err
		}
		return n, nil
	}

	return 0, errors.New("expected number or numeric string")
}

func decodeBase64Any(s string) ([]byte, error) {
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.RawURLEncoding.DecodeString(s)
}

func valueOrDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func toHostList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return splitCSV(v)
}

func toPathList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return []string{"/"}
	}
	return []string{v}
}
