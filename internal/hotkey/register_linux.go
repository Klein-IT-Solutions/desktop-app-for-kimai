package hotkey

import (
	"context"
	"os"
	"strings"
)

// fallbackHint tells users how to bind the shortcut themselves.
const fallbackHint = "Lege in den Tastatur-Einstellungen deines Desktops ein eigenes Kürzel für den Befehl „kimai-tracker --toggle“ an."

func register(ctx context.Context, acc Accelerator, onPress func(), onStatus func(Status)) Status {
	wayland := strings.EqualFold(os.Getenv("XDG_SESSION_TYPE"), "wayland") || os.Getenv("WAYLAND_DISPLAY") != ""
	if !wayland {
		err := listenX11(ctx, acc, onPress)
		if err == nil {
			return Status{Active: true, Backend: BackendX11, Shortcut: acc.String()}
		}
		// Fall through to the portal (e.g. KDE on X11), then to the hint.
		if s, ok := listenPortal(ctx, acc, onPress, onStatus); ok {
			return s
		}
		return Status{Shortcut: acc.String(), Message: x11ErrorMessage(err) + " " + fallbackHint}
	}
	if s, ok := listenPortal(ctx, acc, onPress, onStatus); ok {
		return s
	}
	return Status{
		Shortcut: acc.String(),
		Message:  "Unter Wayland dürfen Apps keine globalen Tastenkürzel abfangen, und dein Desktop bietet das GlobalShortcuts-Portal nicht an. " + fallbackHint,
	}
}
