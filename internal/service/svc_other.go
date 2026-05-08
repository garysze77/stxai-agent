//go:build !windows

package service

import "context"

// RunWindowsService is a no-op on non-Windows platforms.
func RunWindowsService(ctx context.Context, run func(context.Context) error) error {
	return run(ctx)
}
