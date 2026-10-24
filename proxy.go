package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const caddyConfigDir = "/etc/caddy/conf.d"

// ConfigureProxy writes a Caddy site config and reloads Caddy.
// Falls back to Nginx if Caddy is not installed.
func ConfigureProxy(cfg *AppConfig) (string, error) {
	domain := cfg.Domain
	if domain == "" {
		// Use nip.io for automatic wildcard DNS without user setup.
		domain = fmt.Sprintf("%s.%s.nip.io", cfg.Name, publicIP())
	}

	upstream := fmt.Sprintf("127.0.0.1:%d", cfg.Port)

	if commandExists("caddy") {
		return configureCaddy(cfg.Name, domain, upstream)
	}
	if commandExists("nginx") {
		return configureNginx(cfg.Name, domain, upstream)
	}

	return "", fmt.Errorf("neither caddy nor nginx found — install one first")
}

func configureCaddy(name, domain, upstream string) (string, error) {
	if err := os.MkdirAll(caddyConfigDir, 0755); err != nil {
		return "", err
	}

	conf := fmt.Sprintf(`%s {
	reverse_proxy %s
}
`, domain, upstream)

	confPath := filepath.Join(caddyConfigDir, fmt.Sprintf("vessel-%s.caddy", name))
	if err := os.WriteFile(confPath, []byte(conf), 0644); err != nil {
		return "", err
	}

	// Reload Caddy.
	if err := exec.Command("caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run(); err != nil {
		// Caddy might not be running yet — try start.
		exec.Command("systemctl", "reload", "caddy").Run()
	}

	return "https://" + domain, nil
}

func configureNginx(name, domain, upstream string) (string, error) {
	sitePath := filepath.Join("/etc/nginx/sites-available", fmt.Sprintf("vessel-%s", name))
	enabledPath := filepath.Join("/etc/nginx/sites-enabled", fmt.Sprintf("vessel-%s", name))

	conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
`, domain, upstream)

	if err := os.WriteFile(sitePath, []byte(conf), 0644); err != nil {
		return "", err
	}
	os.Symlink(sitePath, enabledPath)

	exec.Command("nginx", "-t").Run()
	exec.Command("systemctl", "reload", "nginx").Run()

	return "http://" + domain, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func publicIP() string {
	out, err := exec.Command("curl", "-s", "ifconfig.me").Output()
	if err != nil {
		return "127.0.0.1"
	}
	return strings.TrimSpace(string(out))
}