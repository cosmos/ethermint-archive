// This file contains all the utilities to initialise and start
// Tendermint.

package cli

import (
	"context"
	"os/exec"
	"time"
)

// InvokeTendermint invokes Tendermint in-process
func InvokeTendermint(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return invokeTendermint(ctx, args...)
}

func invokeTendermint(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "tendermint", args...)
	return cmd.CombinedOutput()
}
