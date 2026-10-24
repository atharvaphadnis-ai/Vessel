package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "vessel",
		Short: "Deploy apps from Git to your own server. No Docker required.",
		Long: `Vessel is a lightweight PaaS for your VPS.
It clones your repo, builds it natively, and runs it as a systemd service behind a reverse proxy.`,
	}

	root.AddCommand(deployCmd())
	root.AddCommand(listCmd())
	root.AddCommand(removeCmd())
	root.AddCommand(versionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("vessel %s\n", version)
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List deployed apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			apps, err := ListApps()
			if err != nil {
				return err
			}
			if len(apps) == 0 {
				fmt.Println("No apps deployed yet.")
				return nil
			}
			fmt.Printf("%-20s %-10s %-45s %s\n", "NAME", "PORT", "URL", "STATUS")
			for _, a := range apps {
				status := "running"
				if !IsServiceActive(a.Name) {
					status = "stopped"
				}
				fmt.Printf("%-20s %-10d %-45s %s\n", a.Name, a.Port, a.URL, status)
			}
			return nil
		},
	}
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <app-name>",
		Short: "Stop and remove an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RemoveApp(args[0])
		},
	}
}