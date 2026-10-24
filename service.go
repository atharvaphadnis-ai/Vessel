package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const systemdDir = "/etc/systemd/system"

var unitTemplate = `[Unit]
Description=Vessel app: {{.Name}}
After=network.target

[Service]
Type=simple
User=vessel
WorkingDirectory={{.Dir}}
ExecStart={{.Start}}
Restart=always
RestartSec=3
Environment=PORT={{.Port}}
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
`

func InstallService(cfg *AppConfig, appDir string) error {
	unitName := fmt.Sprintf("vessel-%s.service", cfg.Name)
	unitPath := filepath.Join(systemdDir, unitName)

	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(unitPath)
	if err != nil {
		return fmt.Errorf("cannot write unit file (run with sudo?): %w", err)
	}
	defer f.Close()

	startCmd := cfg.Start
	if startCmd == "" {
		return fmt.Errorf("no start command detected — add vessel.yaml")
	}

	if err := tmpl.Execute(f, map[string]interface{}{
		"Name":  cfg.Name,
		"Dir":   appDir,
		"Start": startCmd,
		"Port":  cfg.Port,
	}); err != nil {
		return err
	}

	// Ensure vessel user exists.
	exec.Command("useradd", "-r", "-s", "/bin/false", "vessel").Run()
	exec.Command("chown", "-R", "vessel:vessel", appDir).Run()

	// Reload and restart.
	runCmd("systemctl", "daemon-reload")
	if err := runCmd("systemctl", "enable", "--now", unitName); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("→ Service %s installed and running\n", unitName)
	return nil
}

func IsServiceActive(name string) bool {
	unit := fmt.Sprintf("vessel-%s.service", name)
	err := exec.Command("systemctl", "is-active", "--quiet", unit).Run()
	return err == nil
}

func StopService(name string) error {
	unit := fmt.Sprintf("vessel-%s.service", name)
	exec.Command("systemctl", "stop", unit).Run()
	return exec.Command("systemctl", "disable", unit).Run()
}