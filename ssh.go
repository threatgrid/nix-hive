package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sshCmd)
	sshCmd.Flags().SetInterspersed(false)
}

var sshCmd = &cobra.Command{
	Use:   `ssh`,
	Short: `SSHs to a managed instance`,
	Long:  `ssh will generate an ssh config and use it with your arguments`,
	RunE:  runSSH,
}

func runSSH(cmd *cobra.Command, args []string) error {
	err := generateSshConfig(cmd.Context())
	if err != nil {
		return err
	}
	proc := exec.CommandContext(cmd.Context(), `ssh`, append([]string{
		`-F`, filepath.Join(tmp, `ssh_config`),
	}, args...)...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	return proc.Run()
}
