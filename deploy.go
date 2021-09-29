package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   `deploy`,
	Short: `Deploys systems to NixOS instances`,
	Long:  `Deploy will build, push and activate systems on NixOS instances.`,
	RunE:  runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	instances, err := inv.matchInstances(args...)
	if err != nil {
		return err
	}
	systems := inv.instanceSystems(instances...)
	err = inv.build(ctx, systems...)
	if err != nil {
		return err
	}
	err = inv.push(ctx, instances, ``)
	if err != nil {
		return err
	}
	err = generateSshConfig(ctx)
	if err != nil {
		return err
	}
	return inv.deploy(ctx, instances...)
}

func (inv *Inventory) deploy(ctx context.Context, instances ...string) error {
	for _, instance := range instances {
		err := inv.deployInstance(ctx, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inv *Inventory) deployInstance(ctx context.Context, instance string) error {
	cfg := inv.Instances[instance]
	path := inv.Systems[cfg.System].Result
	args := []string{`-F`, filepath.Join(tmp, `ssh_config`), instance, `sudo`, path + `/bin/switch-to-configuration`, `switch`}
	opts := os.Getenv("NIX_COPYOPTS") //TODO: DOC this and NIX_SSHOPTS
	if opts != `` {
		args = append(args, strings.Split(opts, " ")...)
	}
	inform(ctx, `running ssh %v`, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, `ssh`, args...)
	cmd.Env = append(os.Environ(), `NIX_SSHOPTS=-F `+filepath.Join(tmp, `ssh_config`)+` `+os.Getenv(`NIX_SSHOPTS`))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
