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

## Configuration and secrets

- `internal/config` stores non-secret settings in `~/.config/kimai-tracker/config.json`
  (honours `XDG_CONFIG_HOME`), written atomically with mode `0600`. A corrupt file is
  moved to `config.json.broken` and the app starts with defaults.
- The API token lives in the system keyring via `zalando/go-keyring`
  (service `kimai-tracker`, key `kimai-api-token`). The token is never sent back to
  the frontend; the UI only learns whether one is stored.

## Kimai client

`internal/kimai` is a small, dependency-free client for the Kimai 2.x REST API
(Bearer token auth). Notes from testing against Kimai 2.69:

- Relations come as plain IDs from collection endpoints but as nested objects from
  `/api/timesheets/active`; `kimai.Ref` accepts both.
- Timestamps are returned as `2026-10-13T09:15:00+0200` and must be sent as
  `2026-10-13T09:15:00` in the **user's Kimai time zone**. The time zone is a top-level
  field of `/api/users/me` (older versions: a `timezone` preference). The client formats
  `begin` in that zone, so desktop and Kimai time zones may differ.
- `/api/activities?project=ID` returns project-specific and global activities.
- Errors are mapped to sentinels (`ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`,
  `ErrOffline`, `ErrCertificate`, `ErrNotHTTPS`, `ErrUnexpectedResponse`); the app layer
  turns them into German UI messages (`app.DescribeError`).

Unit tests use `httptest`; `integration_test.go` (build tag `integration`) runs the
full start → update → stop flow against a real instance (`make kimai-up`).

## Services

- `AppService` – app info, dark mode, opening links.
- `SettingsService` – settings, token, connection test; emits `connection:changed`.
- `Connection` (not bound) – shared config/secret access and the Kimai client factory
  for the services added in later steps.
