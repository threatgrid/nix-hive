package main

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sshConfigCmd)
	sshConfigCmd.Flags().SetInterspersed(false)
}

var sshConfigCmd = &cobra.Command{
	Use:   `ssh-config`,
	Short: `Writes the SSH configuration to stdout`,
	RunE:  runSSHConfig,
}

func runSSHConfig(cmd *cobra.Command, args []string) error {
	_, err := os.Stdout.WriteString(inv.SSH)
	return err
}
