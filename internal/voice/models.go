package voice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ModelBaseURL hosts the official whisper.cpp ggml models.
const ModelBaseURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

// Model is a downloadable whisper.cpp model.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	File        string `json:"file"`
	SizeMB      int    `json:"sizeMb"`
}

// Models are quantised multilingual models that run well on laptops.
var Models = []Model{
	{ID: "base-q5_1", Name: "Base", Description: "Schnell, gut für kurze Notizen", File: "ggml-base-q5_1.bin", SizeMB: 57},
	{ID: "small-q5_1", Name: "Small", Description: "Empfohlen – deutlich genauer", File: "ggml-small-q5_1.bin", SizeMB: 181},
	{ID: "medium-q5_0", Name: "Medium", Description: "Sehr genau, braucht mehr Zeit und RAM", File: "ggml-medium-q5_0.bin", SizeMB: 514},
}

// FindModel returns the model with the given ID.
func FindModel(id string) (Model, bool) {
	for _, m := range Models {
		if m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}

// DataDir returns ~/.local/share/kimai-tracker (honouring XDG_DATA_HOME).
func DataDir() (string, error) {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "kimai-tracker"), nil
	}
	switch runtime.GOOS {
	case "darwin", "windows":
		d, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(d, "kimai-tracker"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "kimai-tracker"), nil
}

// ModelPath returns where a downloaded model is stored.
func ModelPath(m Model) (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "models", m.File), nil
}

// ggmlMagic starts every ggml model file ("ggml" as little-endian uint32).
var ggmlMagic = []byte{0x6c, 0x6d, 0x67, 0x67}

// Download fetches url into dest, reporting progress. The file is written to
// a temporary name and only renamed once complete and recognised as a ggml
// model, so an interrupted download never leaves a broken model behind.
func Download(ctx context.Context, client *http.Client, url, dest string, progress func(done, total int64)) error {
	if client == nil {
		client = &http.Client{}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "kimai-tracker")
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("Download fehlgeschlagen: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download fehlgeschlagen: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".download-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	total := resp.ContentLength
	var done int64
	var head []byte
	buf := make([]byte, 256<<10)
	lastReport := int64(-1)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if len(head) < len(ggmlMagic) {
				head = append(head, buf[:min(n, len(ggmlMagic)-len(head))]...)
			}
			if _, err := tmp.Write(buf[:n]); err != nil {
				tmp.Close()
				return err
			}
			done += int64(n)
			if progress != nil && (done-lastReport >= 1<<20 || done == total) {
				progress(done, total)
				lastReport = done
			}
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			tmp.Close()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("Download abgebrochen: %v", rerr)
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if total > 0 && done != total {
		return fmt.Errorf("Download unvollständig (%d von %d Bytes)", done, total)
	}
	if !bytes.Equal(head, ggmlMagic) {
		return errors.New("Die heruntergeladene Datei ist kein whisper.cpp-Modell")
	}
	return os.Rename(tmp.Name(), dest)
}

// binaryNames are the names whisper.cpp's CLI is installed under
// (upstream, Homebrew, distribution packages).
var binaryNames = []string{"whisper-cli", "whisper-cpp", "whisper.cpp", "whisper-cpp-cli"}

// FindBinary looks for whisper.cpp's command-line tool in PATH, next to the
// app and in the app's data directory. It returns "" if none is found.
func FindBinary() string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if d, err := DataDir(); err == nil {
		dirs = append(dirs, filepath.Join(d, "bin"))
	}
	dirs = append(dirs, "/usr/local/bin", "/opt/homebrew/bin", "/home/linuxbrew/.linuxbrew/bin")
	for _, name := range binaryNames {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
		for _, d := range dirs {
			p := filepath.Join(d, name)
			if info, err := os.Stat(p); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
				return p
			}
		}
	}
	return ""
}
