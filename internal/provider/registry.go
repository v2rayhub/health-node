package provider

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

// URIParser parses one URI scheme into a Provider.
type URIParser interface {
	Scheme() string
	Parse(u *url.URL, raw string) (Provider, error)
}

type Registry struct {
	mu      sync.RWMutex
	parsers map[string]URIParser
}

func NewRegistry(parsers ...URIParser) *Registry {
	r := &Registry{parsers: make(map[string]URIParser)}
	for _, p := range parsers {
		_ = r.Register(p)
	}
	return r
}

func (r *Registry) Register(p URIParser) error {
	if p == nil {
		return fmt.Errorf("parser is nil")
	}
	s := strings.ToLower(strings.TrimSpace(p.Scheme()))
	if s == "" {
		return fmt.Errorf("parser scheme is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.parsers[s]; exists {
		return fmt.Errorf("parser already registered for scheme %q", s)
	}
	r.parsers[s] = p
	return nil
}

func (r *Registry) Parse(raw string) (Provider, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)

	r.mu.RLock()
	p := r.parsers[scheme]
	r.mu.RUnlock()
	if p == nil {
		return nil, fmt.Errorf("unsupported scheme %q (supported: %s)", u.Scheme, strings.Join(r.Schemes(), ", "))
	}
	return p.Parse(u, raw)
}

func (r *Registry) Schemes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.parsers))
	for s := range r.parsers {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

var defaultRegistry = func() *Registry {
	r := NewRegistry()
	mustRegister(r, &vlessParser{})
	mustRegister(r, &vmessParser{})
	mustRegister(r, &shadowsocksParser{})
	return r
}()

func mustRegister(r *Registry, p URIParser) {
	if err := r.Register(p); err != nil {
		panic(err)
	}
}

func FromURI(raw string) (Provider, error) {
	return defaultRegistry.Parse(raw)
}

func SupportedSchemes() []string {
	return defaultRegistry.Schemes()
}

func RegisterParser(p URIParser) error {
	return defaultRegistry.Register(p)
}
