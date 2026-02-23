# proxy-node

[![GitHub stars](https://img.shields.io/github/stars/v2rayhub/proxy-node)](https://github.com/v2rayhub/proxy-node/stargazers)
[![GitHub release](https://img.shields.io/github/v/release/v2rayhub/proxy-node)](https://github.com/v2rayhub/proxy-node/releases)
[![Release Workflow](https://github.com/v2rayhub/proxy-node/actions/workflows/release.yml/badge.svg)](https://github.com/v2rayhub/proxy-node/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/v2rayhub/proxy-node)](https://github.com/v2rayhub/proxy-node/blob/main/go.mod)

`proxy-node` is a lightweight V2Ray and Xray proxy client utility written in Go for Linux and macOS, providing local SOCKS/HTTP proxy, connectivity probe, and speed test from `vless://`, `vmess://`, and `ss://` links.

## Features

- Open local SOCKS5 or HTTP proxy from share links.
- Probe connectivity through the started proxy.
- Measure download speed through SOCKS5.
- Install Xray/V2Ray core binaries from GitHub releases.
- Minimal single-binary CLI workflow.

## Install

Prebuilt binaries:

https://github.com/v2rayhub/proxy-node/releases

Build locally:

```bash
go build -o proxy-node ./cmd/proxy-node
```

## Usage

### Install Core

```bash
./proxy-node install-core
```

Examples:

```bash
./proxy-node install-core --repo XTLS/Xray-core --version v26.2.6
./proxy-node install-core --repo v2fly/v2ray-core --version v5.20.0 --dest ./core
./proxy-node install-core --force
```

Core auto-detection order:
- `./xray` or `./v2ray`
- `./core/xray` or `./core/v2ray`
- `PATH`

### Run Local Proxy

SOCKS5 on port `1080`:

```bash
./proxy-node proxy --uri 'vmess://BASE64_JSON' --inbound socks --local-port 1080
```

HTTP proxy on port `8080`:

```bash
./proxy-node proxy --uri 'vmess://BASE64_JSON' --inbound http --local-port 8080
```

Notes:
- `socks` is alias for `proxy --inbound socks`.
- Use `--print-requests` to stream core logs.
- Use `--no-traffic` to disable traffic meter output.

### Probe

Default probe URL:
- `https://www.cloudflare.com/cdn-cgi/trace`

```bash
./proxy-node probe --uri 'vless://UUID@server.example.com:443?type=ws&security=tls&host=server.example.com&path=%2Fws&sni=server.example.com'
```

### Speed

Default speed URL:
- `https://speed.cloudflare.com/__down?bytes=10000000`

```bash
./proxy-node speed --uri 'vmess://BASE64_JSON'
```

Useful flags:
- `--retries` retry count on failure (default: `1`).
- `--max-bytes` stop after N bytes (`0` means full response).
- `--timeout` overall timeout.

### Help

```bash
./proxy-node --help
./proxy-node proxy --help
./proxy-node probe --help
./proxy-node speed --help
./proxy-node install-core --help
```

## Troubleshooting

- Always quote full URIs in shell:
  - `./proxy-node probe --uri 'vless://...&security=reality&pbk=...#tag'`
- If logs show `accepted tcp:... [proxy]` then reset/EOF, local SOCKS is up and remote path is dropping streams.
- VLESS/REALITY profiles can behave differently across clients. If VMess works but VLESS fails, verify `pbk`, `sid`, `sni`, `fp`, and server-side config for that node.

## Development

```bash
gofmt -w ./cmd ./internal
go test ./...
go build -o proxy-node ./cmd/proxy-node
```

Targeted tests:

```bash
go test ./internal/provider -v
```

Project layout:
- `cmd/proxy-node/main.go`: CLI commands and runtime orchestration.
- `internal/core`: core process/config runner.
- `internal/provider`: protocol parsers and outbound builders.
- `internal/installer`: core downloader/installer.
- `internal/proxy`: SOCKS5 client path for probe/speed.
