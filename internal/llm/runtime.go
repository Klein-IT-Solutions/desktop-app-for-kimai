package llm

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/klein-it-solutions/kimai-desktop-app/internal/fetch"
)

// ReleaseBaseURL hosts llama.cpp's official release builds.
const ReleaseBaseURL = "https://github.com/ggml-org/llama.cpp/releases"

// FallbackTag is used when the latest release cannot be determined.
const FallbackTag = "b11586"

// serverName is the llama.cpp server binary.
var serverName = func() string {
	if runtime.GOOS == "windows" {
		return "llama-server.exe"
	}
	return "llama-server"
}()

// AssetName returns the release archive for this platform, following the
// naming in llama.cpp's .github/workflows/release.yml (CPU builds).
func AssetName(tag, goos, goarch string) (string, error) {
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[goarch]
	if arch == "" {
		return "", fmt.Errorf("keine llama.cpp-Builds für %s/%s", goos, goarch)
	}
	switch goos {
	case "linux":
		return fmt.Sprintf("llama-%s-bin-ubuntu-%s.tar.gz", tag, arch), nil
	case "darwin":
		return fmt.Sprintf("llama-%s-bin-macos-%s.tar.gz", tag, arch), nil
	case "windows":
		return fmt.Sprintf("llama-%s-bin-win-cpu-%s.zip", tag, arch), nil
	}
	return "", fmt.Errorf("keine llama.cpp-Builds für %s", goos)
}

var tagPattern = regexp.MustCompile(`/releases/tag/(b\d+)`)

// LatestTag resolves the newest release via the redirect of
// /releases/latest (no API token or rate limit involved).
func LatestTag(ctx context.Context, client *http.Client, base string) (string, error) {
	if client == nil {
		client = &http.Client{}
	}
	noRedirect := *client
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, base+"/latest", nil)
	if err != nil {
		return "", err
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if m := tagPattern.FindStringSubmatch(resp.Header.Get("Location")); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("unerwartete Antwort (HTTP %d)", resp.StatusCode)
}

// RuntimeDir returns where llama.cpp builds are installed.
func RuntimeDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "llama.cpp"), nil
}

// FindServer looks for llama-server: an installed build in the data
// directory first, then PATH. It returns "" if none is found.
func FindServer() string {
	if dir, err := RuntimeDir(); err == nil {
		var found string
		_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && d.Name() == serverName && found == "" {
				found = p
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	if p, err := exec.LookPath(serverName); err == nil {
		return p
	}
	return ""
}

// Install downloads and unpacks the llama.cpp build for this platform and
// returns the path to llama-server. base is ReleaseBaseURL or a mirror.
func Install(ctx context.Context, client *http.Client, base string, progress func(done, total int64)) (string, error) {
	tag, err := LatestTag(ctx, client, base)
	if err != nil {
		tag = FallbackTag
	}
	asset, err := AssetName(tag, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	dir, err := RuntimeDir()
	if err != nil {
		return "", err
	}
	archive := filepath.Join(dir, asset)
	magic := fetch.MagicGzip
	if strings.HasSuffix(asset, ".zip") {
		magic = fetch.MagicZip
	}
	if err := fetch.File(ctx, client, base+"/download/"+tag+"/"+asset, archive, magic, progress); err != nil {
		return "", err
	}
	defer os.Remove(archive)

	dest := filepath.Join(dir, tag)
	_ = os.RemoveAll(dest)
	if magic[0] == 'P' {
		err = extractZip(archive, dest)
	} else {
		err = extractTarGz(archive, dest)
	}
	if err != nil {
		_ = os.RemoveAll(dest)
		return "", fmt.Errorf("llama.cpp konnte nicht entpackt werden: %v", err)
	}
	// Remove older installed builds.
	if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if e.IsDir() && e.Name() != tag {
				_ = os.RemoveAll(filepath.Join(dir, e.Name()))
			}
		}
	}
	var server string
	_ = filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == serverName {
			server = p
		}
		return nil
	})
	if server == "" {
		return "", errors.New("llama-server fehlt im heruntergeladenen Paket")
	}
	return server, nil
}

// safeJoin rejects archive entries that would escape dest ("zip slip").
func safeJoin(dest, name string) (string, error) {
	clean := path.Clean("/" + strings.ReplaceAll(name, "\\", "/"))
	p := filepath.Join(dest, filepath.FromSlash(clean))
	if !strings.HasPrefix(p, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("ungültiger Pfad im Archiv: %s", name)
	}
	return p, nil
}

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		p, err := safeJoin(dest, h.Name)
		if err != nil {
			return err
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeFile(p, tr, os.FileMode(h.Mode)&0o755|0o600); err != nil {
				return err
			}
		case tar.TypeSymlink:
			// Shared-library symlinks (libfoo.so -> libfoo.so.0) must stay inside dest.
			if strings.Contains(h.Linkname, "/") || strings.Contains(h.Linkname, "..") {
				return fmt.Errorf("unerlaubter Link im Archiv: %s", h.Name)
			}
			_ = os.MkdirAll(filepath.Dir(p), 0o755)
			if err := os.Symlink(h.Linkname, p); err != nil {
				return err
			}
		}
	}
}

func extractZip(archive, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		p, err := safeJoin(dest, f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeFile(p, rc, 0o755)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(p string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
