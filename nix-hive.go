package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var err error
	tmp, err = os.MkdirTemp(``, `hive-`)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			panic(err)
		}
	}()

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:                `nix-hive`,
	Short:              `A system for managing identical NixOS systems`,
	Long:               `Nix-Hive is a tool that builds NixOS systems and deploys them efficiently.`,
	PersistentPreRunE:  preRun,
	PersistentPostRunE: postRun,
	SilenceUsage:       true,
}

func preRun(cmd *cobra.Command, args []string) error {
	err := loadInventory(cmd, args)
	if err != nil {
		return err
	}
	err = applyState()
	if err != nil {
		return err
	}
	return nil
}

func postRun(cmd *cobra.Command, args []string) error {
	err := saveState()
	if err != nil {
		return err
	}
	return nil
}

var tmp = ``

func inform(ctx context.Context, msg string, info ...interface{}) {
	fmt.Fprintf(os.Stderr, ".. "+msg+"\n", info...)
}

func warn(ctx context.Context, msg string, info ...interface{}) {
	fmt.Fprintf(os.Stderr, ":: "+msg+"\n", info...)
}
