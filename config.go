package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rf := rootCmd.PersistentFlags()
	rf.StringVarP(
		&deploymentPath, `config`, `c`, `hive.nix`, `Nix deployment configuration path`)
	rf.StringVarP(
		&statePath, `state`, `s`, `.hive.state`, `Nix-Hive state path`)
	rf.StringVar(
		&no, `no`, ``, `List of steps to skip if previously complete, separated by ","`)
}

// loadInventory evaluates <hive/inventory.nix> to establish inventory.
func loadInventory(cmd *cobra.Command, args []string) error {
	data, err := execNix(
		cmd.Context(), `eval`, `--json`,
		`--include`, `deployment=`+deploymentPath,
		`(import <hive/config.nix>)`)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &inv)
	if err != nil {
		return err
	}
	return applyState()
}

func applyState() error {
	data, err := ioutil.ReadFile(statePath)
	switch {
	case err == nil: // that's good!
	case os.IsNotExist(err): // that's.. fine.
		return nil
	default:
		return err
	}
	var dont struct {
		build bool
	}
	for _, step := range strings.Split(no, ",") {
		switch step {
		case `build`:
			dont.build = true
		}
	}
	_, err = processState(data, func(predicate string, terms ...string) error {
		switch predicate {
		case `r0`: // result, variant 0.
			if len(terms) != 2 {
				return fmt.Errorf(`expected a system and result path, for r0, got %v terms`, len(terms))
			}
			cfg, ok := inv.Systems[terms[0]]
			if !ok {
				return nil // system no longer exists.
			}
			if dont.build && pathExists(terms[1]) {
				cfg.Result = terms[1]
			}
		}
		return nil
	})
	return err
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func saveState() error {
	state := make([]byte, 0, len(inv.Systems)*64)
	state = appendSystemResultFacts(state)
	return ioutil.WriteFile(statePath, state, 0600)
}

func appendSystemResultFacts(state []byte) []byte {
	names := make([]string, 0, len(inv.Systems))
	for name := range inv.Systems {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		result := inv.Systems[name].Result
		if result != `` {
			state = appendFact(state, `r0`, name, result)
		}
	}
	return state
}

var deploymentPath = `./hive.nix`
var statePath = `.hive.state`
var no = ``

var inv Inventory

type Inventory struct {
	// Paths is a list of Nix paths that should be passed to nix-build when building systems in the inventory.  These
	// paths may be extended and overridden by the per-system Paths list.
	Paths []string `json:"paths,omitempty"`

	// Systems maps system information by system name.
	Systems map[string]*System `json:"systems"`

	// SSH contains a literal ssh_config (see "man ssh_config") that hive will use when accessing remote
	// hosts, including when it uses "nix copy" to transfer paths to a store.
	SSH string `json:"ssh"`

	// Instances maps instance information by instance name.
	Instances map[string]*Instance `json:"instances"`
}

// instanceSystems identifies unique systems associated with instances in the provided patterns, in pattern order.
func (inv *Inventory) instanceSystems(instances ...string) []string {
	added := make(map[string]struct{}, len(inv.Systems))
	systems := make([]string, 0, len(inv.Systems))
	for _, instance := range instances {
		system := inv.Instances[instance].System
		if _, dup := added[system]; dup {
			continue
		}
		added[system] = struct{}{}
		systems = append(systems, system)
	}
	return systems
}

// matchSystems identifies unique systems that match the provided patterns.  If no patterns are provided, then each
// system is returned.
func (inv *Inventory) matchSystems(patterns ...string) ([]string, error) {
	rows := make([][]string, 0, len(inv.Systems))
	for name := range inv.Systems {
		rows = append(rows, []string{name})
	}
	return matchPatterns(patterns, rows...)
}

// matchInstances identifies unique instances that match the provided patterns.  If no patterns are provided, then
// each instance is returned.
func (inv *Inventory) matchInstances(patterns ...string) ([]string, error) {
	rows := make([][]string, 0, len(inv.Instances))
	for name, cfg := range inv.Instances {
		row := make([]string, 2+len(cfg.Tags))
		row[0] = name
		row[1] = cfg.System
		copy(row[2:], cfg.Tags)
		rows = append(rows, row)
	}
	return matchPatterns(patterns, rows...)
}

func (inv *Inventory) systemPaths(system string) []string {
	cfg := inv.Systems[system]
	seq := make([]string, 0, len(inv.Paths)+len(cfg.Paths))
	seq = append(seq, inv.Paths...)
	seq = append(seq, cfg.Paths...)
	return seq
}

// An Instance describes a host where a system configuration should be deployed.
type Instance struct {
	// System names the system configuration that should be deployed to this instance.
	System string `json:"system"`

	// Store is the URL to a Nix store where built systems should be uploaded prior to transferring to the
	// instance.  This is crucial to avoiding redundant transfers of a shared system to multiple instances.
	Store string `json:"store,omitempty"`

	// Tags provide a way to group instances so they can be targeted for a deployment without using their name.
	Tags []string `json:"tags,omitempty"`
}

type System struct {
	// Paths is a list of Nix paths that should be passed to nix-build when building the system, in --include format.
	Paths []string `json:"paths,omitempty"`

	// Result identifies the path to the built system.  This is populated by the build method, and not by
	// <hive/inventory.nix>.
	Result string `json:"result,omitempty"`
}

// matchPatterns searches rows for items that match a set of patterns, returning the first item in each row for hit,
// in the order that they occur.  matchPatterns is inherently slow, since it assumes the patterns are globs and uses
// path.Match -- faster behavior would be achieved by using a regex.
//
// IOW, this is O(n*m) and might need remedation for more than a thousand rows or a few patterns.
func matchPatterns(patterns []string, rows ...[]string) (hits []string, err error) {
	if len(patterns) == 0 {
		patterns = []string{`*`}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	added := make(map[string]struct{}, len(rows))
	hits = make([]string, 0, len(rows))
	for _, pattern := range patterns {
		used := false
		for _, row := range rows {
			name := row[0]
			for _, item := range row {
				if match, err := path.Match(pattern, item); err != nil {
					return nil, fmt.Errorf(`%v while parsing pattern %q`, err, pattern)
				} else if !match {
					continue
				}
				used = true
				if _, dup := added[name]; dup {
					continue
				}
				added[name] = struct{}{}
				hits = append(hits, name)
				break
			}
		}
		if !used {
			return nil, fmt.Errorf(`%q did not match anything`, pattern)
		}
	}
	return
}
