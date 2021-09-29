package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   `run`,
	Short: `Runs a command on managed hosts.`,
	Long:  `Run runs a command on managed hosts in an optional Nix shell`,
	RunE:  runCommand,
}

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if len(args) < 2 {
		return fmt.Errorf(`run expects an instance pattern followed by a command to run`)
	}
	//TODO: support user@pattern
	pattern, command := args[0], args[1:]
	instances, err := inv.matchInstances(pattern)
	if err != nil {
		return err
	}
	err = generateSshConfig(cmd.Context())
	if err != nil {
		return err
	}
	failed := 0
	for _, instance := range instances {
		println(`##`, instance)
		err := runCommandOn(ctx, instance, command)
		if err != nil {
			println(`!!`, err.Error())
			failed++
		}
	}
	switch failed {
	case 0:
		return nil
	case 1:
		return fmt.Errorf(`one instance failed`)
	}
	return fmt.Errorf(`%v instances failed`, failed)
}

func runCommandOn(ctx context.Context, instance string, command []string) error {
	proc := exec.CommandContext(ctx, `ssh`, append([]string{
		`-F`, filepath.Join(tmp, `ssh_config`), instance,
	}, command...)...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	return proc.Run()
}

// buildShell builds a Nix expression at path similar to how nix-shell does it and returns a path to it in the Nix
// store.
func buildShell(ctx context.Context, path string) (string, error) {
	// https://github.com/NixOS/nix/blob/27444d40cf726129334899a42d68d69f73baa988/src/nix-build/nix-build.cc
	// https://github.com/NixOS/nixpkgs/blob/master/pkgs/build-support/mkshell/default.nix

	panic(`TODO`)

	// mkshell produces an intentionally broken derivation which is treated specially by nix / nix-build / nix-shell
	// to produce a file like:
	/*
		rm -rf '/tmp/nix-shell-50765-0'
		[ -n "$PS1" ] && [ -e ~/.bashrc ] && source ~/.bashrc
		p=$PATH
		dontAddDisableDepTrack=1
		[ -e $stdenv/setup ] && source $stdenv/setup
		PATH=$PATH:$p
		unset p;
		PATH="/nix/store/mp1q3wb5jdx85jklmhywqswqkk9wrkg1-bash-interactive-4.4-p23/bin:$PATH"
		SHELL=/nix/store/mp1q3wb5jdx85jklmhywqswqkk9wrkg1-bash-interactive-4.4-p23/bin/bash
		set +e;
		[ -n "$PS1" ] && PS1='\n\[\033[1;32m\][nix-shell:\w]\$\[\033[0m\] '
		if [ "$(type -t runHook)" = function ]; then runHook shellHook; fi
		unset NIX_ENFORCE_PURITY
		shopt -u nullglob
		unset TZ
		shopt -s execfail
		/bin/true
		exit
	*/
	// This file is run with "bash --rcfile /tmp/nix-shell-38231-0/rc" and it, along with the tempdir, is removed
	// by the parent process after bash exits.
	//
	// In order to replicate this in a more maintainable way, we have <nix/shell.nix> that overlays pkgs.mkShell with
	// our own variant that has its own phases and buildPhase that captures the build environment.  (We assume that all
	// the complexity in nix-build.cc is a vestige meant to allow hacking on unbuilt derivations.)
}
