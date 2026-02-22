# health-node

[![Release](https://img.shields.io/github/v/release/v2rayhub/health-node)](https://github.com/v2rayhub/health-node/releases)
[![Release Workflow](https://github.com/v2rayhub/health-node/actions/workflows/release.yml/badge.svg)](https://github.com/v2rayhub/health-node/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/v2rayhub/health-node)](https://github.com/v2rayhub/health-node/blob/main/go.mod)

Small V2Ray/Xray client utility for Linux and macOS.
It can open local SOCKS/HTTP proxy from `vless://`, `vmess://`, `ss://`, run health checks, and run speed tests.

## How To Use

### Download / Build

Prebuilt binaries:

https://github.com/v2rayhub/health-node/releases

Or build locally:

```bash
go build -o health-node ./cmd/health-node
```

### Install Core (Automatic)

```bash
./health-node install-core
```

Examples:

```bash
# specific Xray version
./health-node install-core --repo XTLS/Xray-core --version v26.2.6

# v2ray-core into ./core
./health-node install-core --repo v2fly/v2ray-core --version v5.20.0 --dest ./core

# overwrite existing core binary
./health-node install-core --force
```

Core auto-detection order:
- same folder: `./xray` or `./v2ray`
- subfolder: `./core/xray` or `./core/v2ray`
- `PATH` fallback

### Open Local Proxy (Main Mode)

SOCKS5 (default):

```bash
./health-node proxy --uri 'vmess://BASE64_JSON' --local-port 1080
```

HTTP proxy:

```bash
./health-node proxy --uri 'vmess://BASE64_JSON' --inbound http --local-port 8080
```

Notes:
- `socks` is alias for `proxy --inbound socks`
- Traffic monitor is on by default
- Disable traffic monitor with `--no-traffic`
- Tune refresh cost with `--traffic-interval 2s` (or slower, e.g. `5s`)

### Health Check / Speed Test

Connectivity probe:

```bash
./health-node probe --uri 'vless://UUID@server.example.com:443?type=ws&security=tls&host=server.example.com&path=%2Fws&sni=server.example.com'
```

Speed test:

```bash
./health-node speed \
  --uri 'vmess://BASE64_JSON' \
  --url 'https://ash-speed.hetzner.com/1GB.bin' \
  --max-bytes 0 \
  --timeout 5m
```

Help:

```bash
./health-node --help
./health-node proxy --help
./health-node probe --help
./health-node speed --help
./health-node install-core --help
```

## How To Develop And Extend

### Local Dev Cycle

```bash
# format
gofmt -w ./cmd ./internal

# test
go test ./...

# build
go build -o health-node ./cmd/health-node
```

Targeted tests:

```bash
go test ./internal/provider -v
```

### Project Structure

- `cmd/health-node/main.go`
CLI commands and runtime orchestration.

- `internal/core`
Core process/config runner for xray/v2ray.

- `internal/provider`
Protocol parsers and outbound builders.

- `internal/provider/registry.go`
Parser interface/registry for protocol extension.

- `internal/installer`
GitHub release downloader/installer for core binaries.

- `internal/proxy`
SOCKS5 client used by probe/speed code paths.

### Add A New Protocol

1. Add a protocol file in `internal/provider/` (for example `trojan.go`).
2. Implement parser + provider output using existing pattern:
- parser implements `URIParser`
- provider struct implements `Provider`
3. Register parser in `internal/provider/registry.go` default registry.
4. Add tests in `internal/provider/*_test.go` for parse + outbound generation.

### Release Automation

Tag push triggers multi-platform release build and upload:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Workflow file:
- `.github/workflows/release.yml`
