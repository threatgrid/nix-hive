package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   `push`,
	Short: `Pushes built systems to stores`,
	Long:  `Push will transfer NixOS systems to remote stores for deployment.`,
	RunE:  runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
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
	//TODO: let users specify a path to push.
	err = inv.push(ctx, instances, ``)
	if err != nil {
		return err
	}
	return nil
}

// push pushes a path to each instance, caching the paths in the instance stores as necessary.  If a path is an empty
// string, the system path for the instance will be used.
func (inv *Inventory) push(ctx context.Context, instances []string, paths ...string) error {
	err := generateSshConfig(ctx)
	if err != nil {
		return err
	}

	type push struct {
		Path  string
		Store string
	}
	pushes := make([]push, 0, len(instances)*2)
	added := make(map[push]struct{}, cap(pushes))
	stores := make([]string, 0, cap(pushes))
	jobs := make(map[string][]string, cap(stores))

	addPush := func(path, store string) {
		item := push{path, store}
		if _, dup := added[item]; dup {
			return
		}
		added[item] = struct{}{}
		pushes = append(pushes, item)
		job, ok := jobs[store]
		if !ok {
			job = make([]string, 0, len(instances))
			stores = append(stores, store)
		}
		job = append(job, path)
		jobs[store] = job
	}

	for _, instance := range instances {
		cfg := inv.Instances[instance]
		for _, path := range paths {
			if path == `` {
				path = inv.Systems[cfg.System].Result
			}
			if path == `` {
				continue // no system, no path, nothing to do.
			}
			if cfg.Store != `` {
				addPush(path, cfg.Store)
			}
			addPush(path, `ssh://`+instance)
		}
	}

	for _, store := range stores {
		err := inv.pushNixPaths(ctx, store, jobs[store]...)
		if err != nil {
			return fmt.Errorf(`%w while pushing to %q`, err, store)
		}
	}
	return nil
}

func (inv *Inventory) pushNixPaths(ctx context.Context, store string, paths ...string) error {
	paths = uniqueStrings(paths)
	args := make([]string, 0, len(paths)+8)
	args = append(args, `copy`, `--to`, store, `--substitute-on-destination`)
	opts := os.Getenv("NIX_COPYOPTS") //TODO: DOC this and NIX_SSHOPTS
	if opts != `` {
		args = append(args, strings.Split(opts, " ")...)
	}
	args = append(args, paths...)

	inform(ctx, `running nix %v`, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, `nix`, args...)
	cmd.Env = append(os.Environ(), `NIX_SSHOPTS=-F `+filepath.Join(tmp, `ssh_config`)+` `+os.Getenv(`NIX_SSHOPTS`))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func uniqueStrings(seq []string) []string {
	seen := make(map[string]struct{}, len(seq))
	ret := make([]string, 0, len(seq))
	for _, item := range seq {
		if _, dup := seen[item]; dup {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}

func generateSshConfig(ctx context.Context) error {
	generateSshConfigOnce.Do(func() {
		generateSshConfigErr = ioutil.WriteFile(filepath.Join(tmp, `ssh_config`), []byte(inv.SSH), 0600)
	})
	return generateSshConfigErr
}

var generateSshConfigOnce sync.Once
var generateSshConfigErr error
