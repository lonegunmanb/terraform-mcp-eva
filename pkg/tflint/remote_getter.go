package tflint

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	getter "github.com/hashicorp/go-getter/v2"
)

// RemoteGetter defines interface for fetching remote config sources using go-getter
// Get should download src to dst (exact file path) with built-in timeout handling.
type RemoteGetter interface {
	Get(dst, src string) error
}

// remoteConfigGetter is a package-level variable to allow test stubbing. Initialized directly
// with the production implementation so we don't need an init() function.
var remoteConfigGetter RemoteGetter = goGetterImpl{}

// goGetterImpl implements RemoteGetter using go-getter for all remote downloads
type goGetterImpl struct{}

func (g goGetterImpl) Get(dst, src string) error {
	// Apply timeout with env var override (default 60s, override via TFLINT_REMOTE_CONFIG_TIMEOUT_SECONDS)
	timeout := 60 * time.Second
	if v := os.Getenv("TFLINT_REMOTE_CONFIG_TIMEOUT_SECONDS"); v != "" {
		if secs, parseErr := strconv.Atoi(v); parseErr == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if _, err := getter.GetFile(ctx, dst, src); err != nil {
		return fmt.Errorf("go-getter GetFile failed: %w", err)
	}
	return nil
}
