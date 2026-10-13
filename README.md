# Kimai Tracker

A minimalist, open-source desktop time tracker for [Kimai](https://www.kimai.org) –
built with Go, [Wails v3](https://v3.wails.io), Svelte, TypeScript and Tailwind CSS.

<a href="https://klein-it.solutions"><img src="docs/brand/klein-it-solutions.png" alt="Klein IT Solutions" width="40" align="left"></a>

Developed and maintained by **[Klein IT Solutions](https://klein-it.solutions)**.
<br clear="left">

> **Status:** early release for Linux (Ubuntu 24.04+, X11 and Wayland). macOS and Windows
> are planned, see [Roadmap](#roadmap).

| Light | Dark |
| ----- | ---- |
| ![Light mode](docs/screenshots/light.png) | ![Dark mode](docs/screenshots/dark.png) |

## Features

- Compact, frameless window (360 × 480) with rounded corners, translucent surface and soft shadow
- Native-looking window buttons per platform (macOS, Windows, GNOME, KDE), honouring
  the GNOME button layout (left/right, minimize shown or hidden)
- Automatic light/dark mode following the system
- Customer → project → activity selection: searchable lists, keyboard navigation,
  recently used entries on top; the next list opens automatically and single
  options are picked for you
- Big start/stop button with live elapsed time (Ctrl+Enter); picks up a timer already
  running in Kimai and notices when it is stopped elsewhere (synced every 60 s)
- Description can be edited while the timer runs and is saved to Kimai automatically
- Works with flaky connections: lists are cached locally, the connection status
  recovers on its own
- Runs in the background: tray icon with status, running time and start/stop; closing
  the window hides it to the tray
- Global hotkey to show/hide the window (default `Super+Shift+T`, configurable)
- Desktop notifications on start/stop, optional start at login
- Voice input for the description: hold the mic button (or a global push-to-talk key) and
  speak; transcribed locally with whisper.cpp or via an OpenAI-compatible API
- Optional local AI turns notes and dictations into short bullet points – fully offline,
  model and runtime are downloaded by the app
- API token stored in the system keyring, never in plain text
- Connection test with clear messages for wrong URL, invalid token (401), missing
  permissions (403), offline server and untrusted TLS certificates

## Installation

Prebuilt packages are on the
[Releases](https://github.com/klein-it-solutions/kimai-desktop-app/releases) page:

- **Ubuntu 24.04+ / Debian 13+:** `sudo apt install ./kimai-tracker_<version>_amd64.deb`
- **Other distributions:** `Kimai-Tracker-<version>-x86_64.AppImage` – make it executable
  (`chmod +x`) and start it.

### Build from source (Ubuntu)

Requirements:

- Go ≥ 1.25
- Node.js ≥ 20 and npm
- GTK 4 and WebKitGTK 6.0 development files (Ubuntu 24.04 or newer)
- CMake and Git (whisper.cpp is fetched and compiled into the app on the first build)

```bash
sudo apt install build-essential pkg-config cmake git libgtk-4-dev libwebkitgtk-6.0-dev

git clone https://github.com/klein-it-solutions/kimai-desktop-app.git
cd kimai-desktop-app
make deps      # installs the wails3 CLI and frontend packages
make build     # → ./bin/kimai-tracker
./bin/kimai-tracker
```

The Makefile finds the `wails3` CLI in `$(go env GOPATH)/bin` automatically. To call
`wails3` directly, add that directory to your `PATH`.

**Ubuntu 22.04:** GTK 4/WebKitGTK 6.0 are not available there. Use the GTK 3 stack instead:

```bash
sudo apt install build-essential pkg-config cmake git libgtk-3-dev libwebkit2gtk-4.1-dev
make build GTK3=1
```

### Development

| Command      | Description                                       |
| ------------ | ------------------------------------------------- |
| `make dev`   | Hot-reload development mode (Go + Vite)           |
| `make build` | Production binary in `./bin`                      |
| `make test`  | Go unit tests and frontend type check             |
| `make lint`  | `go vet` and `gofmt` check                        |
| `make kimai-up` / `make kimai-down` | Local Kimai test instance in Docker |
| `make test-integration` | API tests against a real Kimai          |
| `make version` | Version that would be built (from the latest `v*` tag) |
| `make help`  | All targets                                       |

### Releasing

Every release gets its own version and Git tag. The version is compiled into the app
(shown in the settings) and used for the package names.

- **From GitHub:** *Actions → Release → Run workflow*, choose `patch`, `minor` or `major`.
  The workflow takes the latest `v*` tag, increases it (`v0.1.0` → `v0.1.1`; the first
  release is `v0.1.0`), creates the tag, runs the tests, builds the `.deb` and AppImage
  and publishes a GitHub release with checksums and generated release notes.
- **From the command line:** `git tag v1.2.0 && git push origin v1.2.0` builds exactly
  that version. Tags with a suffix (`v1.2.0-rc.1`) become pre-releases.

Local builds are versioned via `git describe` (e.g. `0.1.0-3-gabc123`); override with
`make package VERSION=1.2.3`. Every push and pull request runs the CI workflow
(lint, tests, build).

### Testing against a local Kimai

`make kimai-up` starts Kimai 2 + MariaDB via Docker Compose, creates an API token and
seeds customers, projects and activities. It prints URL and token, e.g.:

```
Web UI:     http://localhost:8001  (admin / changeme123)
API token:  kt_dev_token_0123456789abcdef0123456789
```

Enter both in the app's settings, or run the integration tests, which start, update
and stop a real timesheet:

```bash
KIMAI_TEST_URL=http://localhost:8001 \
KIMAI_TEST_TOKEN=kt_dev_token_0123456789abcdef0123456789 make test-integration
```

`make kimai-down` removes the instance including its data. You can also point the
integration tests at your own Kimai – use a test account, they create and delete a timesheet.

## Kimai API token

Kimai Tracker talks to the Kimai 2.x REST API and authenticates with a personal API token.

1. Log in to your Kimai instance.
2. Open your profile (avatar, top right) → **API Access** / **API-Zugang**.
3. Create a new token, give it a name such as "Kimai Tracker" and copy it –
   it is shown only once.
4. In Kimai Tracker open **Settings**, enter your Kimai URL
   (e.g. `https://kimai.example.com`) and paste the token.
5. Click **Test connection**.

The token is stored in the system keyring (GNOME Keyring / KWallet via the Secret Service API,
macOS Keychain, Windows Credential Manager) – not in the configuration file.
Non-secret settings live in `~/.config/kimai-tracker/config.json`.

Your Kimai user needs permission to use the API and to create/edit their own timesheets.

If the keyring cannot be used (no Secret Service running, locked collection, dismissed
unlock prompt, unsupported provider), saving shows the reason and offers to store the
token **unencrypted** in `~/.config/kimai-tracker/secrets.json` (mode 0600) instead.
You can move it back into the keyring from the settings later.

To diagnose keyring problems, run:

```bash
./bin/kimai-tracker --check-keyring
```

It shows which program provides the keyring (GNOME Keyring, KWallet, KeePassXC …) and
tests storing, reading and deleting a throw-away value.

The URL may be entered with or without `https://` and with or without a trailing `/api`.
Without a scheme, `https://` is assumed (`http://` for `localhost`).

## Linux notes

- **Transparency & blur:** The window background is semi-transparent when a compositor
  is running (default on GNOME, KDE and most Wayland sessions). Linux has no common API
  to blur what is *behind* a window, so the frosted look is approximated with a tinted
  surface and subtle grain. Without a compositor (rare on X11) the transparent margin
  appears dark.
- **Wayland vs. X11:** Both are supported. Wayland restricts global hotkeys, see
  [Background operation](#background-operation).

## Background operation

**Tray icon.** Shows whether a timer runs (grey / green), the running time and project,
and lets you start/stop the timer, show the window, open the settings or quit.
Closing the window hides it to the tray; quitting does not stop a running timer – it
keeps running in Kimai.

- GNOME needs the *AppIndicator and KStatusNotifierItem Support* extension (pre-installed
  and enabled on Ubuntu). Without a tray, closing the window minimises it instead.
- KDE Plasma, Xfce, Cinnamon, MATE and others show the icon out of the box.

**Global hotkey** (default `Super+Shift+T`, change it in *Settings → Background*):

| Session | How it works |
| ------- | ------------ |
| X11 | Registered directly with the X server. If another app already uses the combination, the settings say so. |
| Wayland | Apps may not grab keys on Wayland. Kimai Tracker uses the XDG *GlobalShortcuts* portal (KDE Plasma 6, GNOME 48+); your desktop asks once to confirm the shortcut. |
| Wayland without portal support | Use the fallback below. |

**Fallback for any desktop:** create a custom keyboard shortcut in your desktop settings
that runs one of these commands – they are forwarded to the running app:

```bash
kimai-tracker --toggle   # show/hide the window
kimai-tracker --timer    # start/stop the timer (last selection)
```

GNOME: *Settings → Keyboard → Keyboard Shortcuts → Custom Shortcuts*.
KDE: *System Settings → Shortcuts → Add New → Command or Script*.

**Start at login** creates `~/.config/autostart/kimai-tracker.desktop`; the app then
starts hidden in the tray (`--hidden`).

## Voice input

Hold the microphone button next to the description and speak – or click once to start and
again to stop (Esc cancels). The recognised text is appended to the description and stays
editable; with a running timer it is saved to Kimai automatically. The global
**push-to-talk** key (default `Super+Shift+V`) works even while the window is hidden.

Choose the transcription in *Settings → Speech input*:

| Option | Notes |
| ------ | ----- |
| **Local (whisper.cpp)** | Fully offline. whisper.cpp is compiled into the app – nothing to install. You only download a model once from within the app (Base 57 MB, Small 181 MB – recommended for German, Medium 514 MB). Optionally an external `whisper-cli` can be configured under *Use external whisper-cli*. |
| **API** | Any OpenAI-compatible `/audio/transcriptions` endpoint: OpenAI (`whisper-1`, `gpt-4o-mini-transcribe`), Groq (`whisper-large-v3`), a self-hosted faster-whisper server, … The key is stored in the system keyring. The recording is sent to that service. |

The language follows *Settings → General → Language* (default German).

Behind a firewall that blocks
Hugging Face, set `KIMAI_TRACKER_MODEL_URL` to a mirror that serves the `ggml-*.bin` files.

## Bullet points with a local AI

*Settings → Bullet points (local AI)* turns descriptions into short bullet points with a
small language model that runs entirely on your computer:

1. Switch it on, click **Load** next to *llama.cpp* (≈ 30 MB, official release build for
   your platform) and next to a model.
2. Use the ✨ button next to the microphone – the result can be undone. Optionally,
   dictations are converted automatically.

| Model | Download | Notes |
| ----- | -------- | ----- |
| Qwen2.5 0.5B Instruct | ≈ 0.5 GB | very fast, simple bullets |
| Qwen2.5 1.5B Instruct | ≈ 1.1 GB | recommended: good German, a few seconds on a laptop CPU |
| Llama 3.2 3B Instruct | ≈ 2.0 GB | best wording, needs ≈ 3 GB RAM |

The model is only loaded when needed (`llama-server` on 127.0.0.1) and unloaded after five
minutes without use. An existing `llama-server` in `PATH` is used as well. Mirrors for
restricted networks: `KIMAI_TRACKER_LLAMA_URL` (llama.cpp releases) and
`KIMAI_TRACKER_LLM_MODEL_URL` (GGUF files).

## Project structure

```
.
├── .github/workflows/      # CI and release builds
├── main.go                 # Wails application & window setup
├── internal/
│   ├── app/                # Services exposed to the frontend
│   ├── config/             # Configuration, keyring, JSON state files
│   ├── desktop/            # Window controls, transparency, tray check, autostart
│   ├── hotkey/             # Global shortcut (X11 grab, Wayland portal)
│   ├── kimai/              # Kimai REST client with tests
│   ├── voice/              # Recording (miniaudio), whisper.cpp, OpenAI API, model download
│   ├── llm/                # Local LLM: llama.cpp runtime install, llama-server, bullet points
│   └── fetch/              # Safe downloads (temp file, progress, type check)
├── frontend/               # Svelte + TypeScript + Tailwind UI
│   ├── bindings/           # Generated TypeScript bindings for Go services
│   └── src/
├── build/                  # Wails build assets (icons, .desktop, nfpm, AppImage)
├── dev/kimai/              # Docker Compose for the local Kimai test instance
├── scripts/                # Test-instance provisioning and seeding
├── docs/                   # Architecture notes, screenshots
├── Makefile
└── Taskfile.yml            # Used internally by the wails3 CLI
```

See [docs/architecture.md](docs/architecture.md) for technical decisions.

## Contributing

Bug reports, ideas and pull requests are welcome – see [CONTRIBUTING.md](CONTRIBUTING.md).
Please report security issues privately as described in [SECURITY.md](SECURITY.md).

## Roadmap

1. ✅ Wails shell with frameless, styled window
2. ✅ Settings + Kimai client including connection test
3. ✅ Customer/project/activity flow and start/stop
4. ✅ Tray, global hotkey, background operation
5. ✅ Voice input
6. ✅ Packaging (.deb, AppImage), CI and versioned releases
7. ⏳ macOS and Windows builds

## License

[MIT](LICENSE) © Klein IT Solutions and the Kimai Tracker contributors.

Kimai Tracker is an independent project and not affiliated with the Kimai project.

---

<sub>Made with care by <a href="https://klein-it.solutions">Klein IT Solutions</a>.</sub>
