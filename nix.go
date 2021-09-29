package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

func execNix(ctx context.Context, args ...string) ([]byte, error) {
	inform(ctx, `running nix %v`, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, `nix`, args...)
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return data, nil
}
