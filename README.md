<div align="center">

# ⛵ Vessel

**From `git push` to production on your own VPS — in one command.**

Vessel is a lightweight, self-hosted PaaS written in Go. It builds your apps natively, runs them as systemd services, and serves them behind Caddy with automatic HTTPS. No Docker. No Kubernetes. No vendor lock-in.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](./LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)](https://www.linux.org)

<br>

![Vessel Demo](demo.gif)

*Deploy a full-stack app to your own VPS in 5 seconds.*

</div>

---

## ✨ Features

- ⚡ **One-command deploys** — `vessel deploy` takes a Git URL and ships it. That's the entire ceremony.
- 🐳 **No Docker, no Kubernetes** — Your code builds natively on the host. No image layers, no registries, no cluster babysitting.
- 🏗️ **Native builds** — Vessel compiles real binaries on your VPS. Faster builds, instant cold starts, zero daemon overhead.
- 🔧 **systemd-native** — Every app runs as a hardened, first-class systemd service. Logs stream straight from `journalctl`.
- 🔒 **Automatic HTTPS** — Built-in Caddy integration routes your domain and provisions and renews TLS certificates for you. Zero config.
- 🪶 **Ridiculously lightweight** — A single static Go binary. Installs in seconds, idles in megabytes of RAM.
- 🏠 **Truly self-hosted** — Your code stays on your machine, your data stays on your server. You are the platform.

---

## 🚀 Quick Start

> All you need: a Linux VPS with SSH access and [Go 1.22+](https://go.dev/dl/).

### 1. Install

```bash
go install github.com/vessel-sh/vessel@latest
```

### 2. Deploy

```bash
# Vessel syncs, builds, and serves your app at https://yourdomain.com
# — TLS certificate issued automatically.
sudo vessel deploy https://github.com/you/your-app
```

---

## ⚙️ Configuration

Drop a `vessel.yaml` in your repository root (optional — sensible defaults apply):

```yaml
# vessel.yaml
name: my-app
domain: myapp.example.com

git:
  branch: main

build:
  command: go build -o bin/my-app ./cmd/server
  artifacts:
    - bin/my-app

run:
  command: ./bin/my-app
  port: 8080
  env:
    - GIN_MODE=release

health:
  path: /healthz
  interval: 10s
```

---

## 🔄 How It Works

**Sync → Build → Serve → Proxy**

1. **Sync** — Vessel connects to your VPS over SSH and pulls the latest revision from your Git repository. No agents, no daemons to babysit.
2. **Build** — Your app is compiled natively on the host. Real binaries in, real binaries out — no container images involved.
3. **Serve** — The build is registered as a hardened systemd unit and started, with restart policies and health checks wired up.
4. **Proxy** — Caddy fronts your app on ports 80/443, routes your domain, and handles TLS issuance and renewal automatically.

---

## 📄 License

Vessel is [MIT licensed](./LICENSE) — free to use, modify, and ship.
