package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scpCmd)
	scpCmd.Flags().SetInterspersed(false)
}

var scpCmd = &cobra.Command{
	Use:   `scp`,
	Short: `Exchanges files with a managed instance via SCP`,
	Long:  `scp will generate an scp config and use it with your arguments`,
	RunE:  runSCP,
}

func runSCP(cmd *cobra.Command, args []string) error {
	err := generateSshConfig(cmd.Context())
	if err != nil {
		return err
	}
	proc := exec.CommandContext(cmd.Context(), `scp`, append([]string{
		`-F`, filepath.Join(tmp, `ssh_config`),
	}, args...)...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	return proc.Run()
}
