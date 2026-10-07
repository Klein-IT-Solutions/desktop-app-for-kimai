package voice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klein-it-solutions/kimai-desktop-app/internal/fetch"
)

// WhisperCLIVersion is the whisper.cpp release built by
// .github/workflows/whisper-cli.yml and published in this repository.
const WhisperCLIVersion = "v1.9.5"

// WhisperCLIBaseURL hosts those builds.
const WhisperCLIBaseURL = "https://github.com/klein-it-solutions/kimai-desktop-app/releases/download"

// BinaryDownloadURL returns the whisper-cli archive for this platform.
// base is WhisperCLIBaseURL or a mirror.
func BinaryDownloadURL(base, goos, goarch string) (string, error) {
	if goos != "linux" {
		return "", errors.New("Für dieses System gibt es noch keinen Download – bitte whisper.cpp installieren (z. B. „brew install whisper-cpp“).")
	}
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[goarch]
	if arch == "" {
		return "", fmt.Errorf("kein whisper-cli-Build für %s", goarch)
	}
	name := fmt.Sprintf("whisper-cli-%s-linux-%s.tar.gz", WhisperCLIVersion, arch)
	return fmt.Sprintf("%s/whisper-cli-%s/%s", strings.TrimRight(base, "/"), WhisperCLIVersion, name), nil
}

// InstallBinary downloads whisper-cli into the app's data directory
// (~/.local/share/kimai-tracker/bin) and returns its path.
func InstallBinary(ctx context.Context, client *http.Client, base string, progress func(done, total int64)) (string, error) {
	url, err := BinaryDownloadURL(base, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(data, 0o755); err != nil {
		return "", err
	}
	tmpDir, err := os.MkdirTemp(data, ".whisper-cli-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	archive := filepath.Join(tmpDir, filepath.Base(url))
	if err := fetch.File(ctx, client, url, archive, fetch.MagicGzip, progress); err != nil {
		return "", fmt.Errorf("%v (%s)", err, filepath.Base(url))
	}
	unpacked := filepath.Join(tmpDir, "x")
	if err := fetch.Extract(archive, unpacked); err != nil {
		return "", fmt.Errorf("whisper-cli konnte nicht entpackt werden: %v", err)
	}
	src := fetch.FindFile(unpacked, "whisper-cli")
	if src == "" {
		return "", errors.New("whisper-cli fehlt im heruntergeladenen Paket")
	}
	dest := filepath.Join(data, "bin", "whisper-cli")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.Chmod(src, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(src, dest); err != nil {
		return "", err
	}
	if lic := fetch.FindFile(unpacked, "LICENSE-whisper.cpp"); lic != "" {
		_ = os.Rename(lic, filepath.Join(data, "bin", "LICENSE-whisper.cpp"))
	}
	return dest, nil
}
