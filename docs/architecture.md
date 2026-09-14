# Architecture notes

## Wails v3 instead of v2

Kimai Tracker is a tray-first application: it lives in the background, shows the
running timer in the tray menu and is toggled via hotkey. Wails v3 ships a native,
cross-platform system tray (`app.SystemTray`), multi-window support and window
hooks (e.g. cancel *close* → hide). Wails v2 has no tray API; it would need a
third-party library such as `fyne.io/systray` or `getlantern/systray`, which run
their own main loop and clash with GTK's main loop on Linux.

Wails v3 is still tagged *beta* (pinned: `v3.0.0-beta.28`), but its API has been
stable for many releases and is used in production by a number of apps. The version
is pinned in `go.mod`, `frontend/package.json` (`@wailsio/runtime`) and the
`Makefile` (`wails3` CLI) – upgrade all three together.

### Linux GTK stack

Wails v3 uses **GTK 4 + WebKitGTK 6.0** on Linux by default (Ubuntu 24.04+,
Debian 13+, Fedora 39+). The legacy **GTK 3 + WebKit2GTK 4.1** stack is available
via the `gtk3` build tag (`make build GTK3=1`) for Ubuntu 22.04; Wails plans to drop
it in v3.1.

## Window

- Frameless, transparent window of 380 × 500 px. The visible card is 360 × 480; the
  10 px transparent gutter holds the CSS drop shadow (`--shadow-gutter` in
  `frontend/src/app.css`). Rounded corners are drawn by CSS, so they look identical
  on X11 and Wayland.
- The drag region is the title bar (`--wails-draggable: drag`); interactive children
  opt out with `no-drag`.
- Linux has no portable "blur behind window" API (KDE offers one, GNOME doesn't).
  The frosted look is therefore a tinted, semi-transparent surface plus a subtle
  noise texture. On macOS/Windows the native translucent backdrop is enabled.
- Single instance: a second launch focuses the running window. This is also the
  basis for the Wayland hotkey fallback (bind a desktop shortcut to the binary).

## Dark mode

The frontend follows `prefers-color-scheme`. On Linux, WebKitGTK does not always pick
up the desktop preference, so the Go side additionally reads the XDG desktop portal
(`org.freedesktop.appearance color-scheme`) and emits `theme:changed` events.

## Frontend ↔ backend

Go services in `internal/app` are bound via Wails; TypeScript bindings are generated
into `frontend/bindings` by `wails3 generate bindings` (run automatically during
`make build` / `make dev`).
