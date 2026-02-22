package provider

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type fakeProvider struct{}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) Outbound() (map[string]any, error) {
	return map[string]any{"tag": "proxy", "protocol": "fake"}, nil
}

type fakeParser struct{}

func (p *fakeParser) Scheme() string { return "fake" }
func (p *fakeParser) Parse(_ *url.URL, _ string) (Provider, error) {
	return &fakeProvider{}, nil
}

func TestSupportedSchemes_Default(t *testing.T) {
	got := SupportedSchemes()
	want := []string{"ss", "vless", "vmess"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedSchemes() = %v, want %v", got, want)
	}
}

func TestRegistry_CustomParser(t *testing.T) {
	r := NewRegistry(&fakeParser{})
	p, err := r.Parse("fake://example")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if p.Name() != "fake" {
		t.Fatalf("provider.Name() = %q, want fake", p.Name())
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	r := NewRegistry(&fakeParser{})
	err := r.Register(&fakeParser{})
	if err == nil {
		t.Fatal("Register() expected duplicate scheme error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("error = %v, want already registered", err)
	}
}
