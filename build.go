package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   `build`,
	Short: `Builds systems for deployment`,
	Long:  `Build will build NixOS systems locally for deployment.`,
	RunE:  runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	systems, err := inv.matchSystems(args...)
	if err != nil {
		return err
	}
	err = inv.build(ctx, systems...)
	if err != nil {
		return err
	}
	results := make(map[string]string, len(systems))
	for _, system := range systems {
		results[system] = inv.Systems[system].Result
	}
	return json.NewEncoder(os.Stdout).Encode(inv)
}

// build builds systems for a specific target, such as "system" for a NixOS system or "vhd" for a disk image.
func (inv *Inventory) build(ctx context.Context, systems ...string) error {
	for _, system := range systems {
		err := inv.Systems[system].build(ctx, system)
		if err != nil {
			return fmt.Errorf(`%w while building %q`, err, system)
		}
	}
	return nil
}

func (cfg *System) build(ctx context.Context, system string) error {
	if cfg.Result != `` {
		return nil // already built.
	}
	inform(ctx, `building %q`, system)
	link := filepath.Join(tmp, `system-`+system)
	args := []string{`build`, `--out-link`, link, `--include`, `deployment=` + deploymentPath}
	paths := inv.systemPaths(system)
	for _, path := range paths {
		args = append(args, `--include`, path)
	}
	args = append(args, `--argstr`, `name`, system)
	args = append(args, `(import <hive/build.nix>)`)
	_, err := execNix(ctx, args...)
	if err != nil {
		return err
	}
	result, err := os.Readlink(link)
	if err != nil {
		return err
	}
	cfg.Result = result
	return err
}
