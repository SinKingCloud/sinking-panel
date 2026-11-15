//go:build darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import "context"

func sources(ctx context.Context) ([]string, error) {
	return fileSources(ctx)
}
