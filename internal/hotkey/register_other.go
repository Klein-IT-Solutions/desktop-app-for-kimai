//go:build !linux

package hotkey

import "context"

func register(_ context.Context, acc Accelerator, _ Shortcut, _, _ func(), _ func(Status)) Status {
	// macOS/Windows support follows with the platform ports.
	return Status{Shortcut: acc.String(), Message: "Globale Tastenkürzel werden auf diesem System noch nicht unterstützt."}
}
