package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const deployRoot = "/opt/vessel"

type AppConfig struct {
	Name    string `yaml:"name"`
	Port    int    `yaml:"port"`
	Build   string `yaml:"build,omitempty"`
	Start   string `yaml:"start,omitempty"`
	Domain  string `yaml:"domain,omitempty"`
	Runtime string `yaml:"runtime,omitempty"`
	URL     string `yaml:"url,omitempty"`
}

func deployCmd() *cobra.Command {
	var name, domain string
	var port int

	cmd := &cobra.Command{
		Use:   "deploy <git-url>",
		Short: "Deploy an app from a Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return DeployApp(args[0], name, port, domain)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "App name (defaults to repo name)")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on (defaults to 3000)")
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Custom domain (defaults to <name>.<server-ip>.nip.io)")
	return cmd
}

func DeployApp(gitURL, name string, port int, domain string) error {
	// 1. Derive app name from URL if not provided.
	if name == "" {
		base := filepath.Base(gitURL)
		name = strings.TrimSuffix(base, ".git")
	}

	// 2. Clone or pull repo.
	appDir := filepath.Join(deployRoot, "apps", name)
	if err := syncRepo(gitURL, appDir); err != nil {
		return fmt.Errorf("sync repo: %w", err)
	}

	// 3. Detect build/start commands.
	cfg, err := DetectApp(appDir)
	if err != nil {
		return fmt.Errorf("detect app: %w", err)
	}
	cfg.Name = name
	if port > 0 {
		cfg.Port = port
	}
	if cfg.Port == 0 {
		cfg.Port = 3000
	}
	if domain != "" {
		cfg.Domain = domain
	}

	// 4. Build.
	if cfg.Build != "" {
		fmt.Printf("→ Building: %s\n", cfg.Build)
		if err := runInDir(appDir, cfg.Build); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	// 5. Install systemd service.
	if err := InstallService(cfg, appDir); err != nil {
		return fmt.Errorf("install service: %w", err)
	}

	// 6. Configure reverse proxy.
	url, err := ConfigureProxy(cfg)
	if err != nil {
		return fmt.Errorf("configure proxy: %w", err)
	}
	cfg.URL = url

	// 7. Save state.
	if err := SaveApp(cfg); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	fmt.Printf("\n✅ Deployed %s\n", name)
	fmt.Printf("   URL:  %s\n", url)
	fmt.Printf("   Port: %d\n", cfg.Port)
	return nil
}

func syncRepo(gitURL, dir string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		fmt.Println("→ Pulling latest changes...")
		return runInDir(dir, "git pull origin HEAD")
	}
	fmt.Println("→ Cloning repository...")
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	return runCmd("git", "clone", "--depth", "1", gitURL, dir)
}

// DetectApp inspects the repo for common language markers.
func DetectApp(dir string) (*AppConfig, error) {
	cfg := &AppConfig{}

	// Check vessel.yaml first.
	if data, err := os.ReadFile(filepath.Join(dir, "vessel.yaml")); err == nil {
		if err := parseYAML(data, cfg); err == nil && cfg.Build != "" {
			return cfg, nil
		}
	}

	// Auto-detect by marker files.
	switch {
	case fileExists(dir, "go.mod"):
		cfg.Build = "go build -o app ."
		cfg.Start = "./app"
		cfg.Runtime = "go"
	case fileExists(dir, "package.json"):
		cfg.Build = "npm install && npm run build"
		cfg.Start = "npm start"
		cfg.Runtime = "node"
	case fileExists(dir, "requirements.txt"):
		cfg.Build = "pip install -r requirements.txt"
		cfg.Start = "python app.py"
		cfg.Runtime = "python"
	case fileExists(dir, "Cargo.toml"):
		cfg.Build = "cargo build --release"
		cfg.Start = "./target/release/app"
		cfg.Runtime = "rust"
	default:
		return nil, fmt.Errorf("could not detect app type — add a vessel.yaml")
	}
	return cfg, nil
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func runInDir(dir, command string) error {
	parts := strings.Fields(command)
	c := exec.Command(parts[0], parts[1:]...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runCmd(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}