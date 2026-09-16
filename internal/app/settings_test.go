package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/klein-it-solutions/kimai-desktop-app/internal/config"
)

func newTestService(t *testing.T) (*SettingsService, config.MemorySecrets) {
	t.Helper()
	store, err := config.Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	secrets := config.MemorySecrets{}
	return NewSettingsService(NewConnection(store, secrets)), secrets
}

func fakeKimai(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"code":401,"message":"Invalid credentials"}`)
			return
		}
		switch r.URL.Path {
		case "/api/users/me":
			_, _ = io.WriteString(w, `{"id":1,"username":"anna","alias":"Anna","timezone":"Europe/Berlin"}`)
		case "/api/version":
			_, _ = io.WriteString(w, `{"version":"2.40.0"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSaveSettingsStoresTokenInSecretsOnly(t *testing.T) {
	svc, secrets := newTestService(t)
	got, err := svc.SaveSettings(SettingsInput{KimaiURL: "kimai.example.com/api/", Token: " tok "})
	if err != nil {
		t.Fatal(err)
	}
	if got.KimaiURL != "https://kimai.example.com" || !got.HasToken {
		t.Errorf("settings = %+v", got)
	}
	if secrets[config.SecretKimaiToken] != "tok" {
		t.Errorf("token = %q", secrets[config.SecretKimaiToken])
	}

	// An empty token keeps the stored one.
	if _, err := svc.SaveSettings(SettingsInput{KimaiURL: "https://kimai.example.com"}); err != nil {
		t.Fatal(err)
	}
	if secrets[config.SecretKimaiToken] != "tok" {
		t.Error("token was cleared by an empty input")
	}
}

func TestSaveSettingsRejectsInvalidURL(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.SaveSettings(SettingsInput{KimaiURL: "ftp://x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestTestConnection(t *testing.T) {
	srv := fakeKimai(t)
	svc, _ := newTestService(t)

	res := svc.TestConnection(srv.URL, "good")
	if !res.OK || res.UserName != "Anna" || res.ServerVersion != "2.40.0" {
		t.Errorf("ok result = %+v", res)
	}

	res = svc.TestConnection(srv.URL, "bad")
	if res.OK || res.Status != StatusUnauthorized {
		t.Errorf("bad token result = %+v", res)
	}

	res = svc.TestConnection("not a url ::", "good")
	if res.Status != StatusInvalidURL {
		t.Errorf("invalid url result = %+v", res)
	}

	res = svc.TestConnection(srv.URL, "")
	if res.Status != StatusUnauthorized {
		t.Errorf("missing token result = %+v", res)
	}
}

func TestCheckConnectionUsesStoredSettings(t *testing.T) {
	srv := fakeKimai(t)
	svc, _ := newTestService(t)

	if res := svc.CheckConnection(); res.Status != StatusNotConfigured {
		t.Fatalf("unconfigured = %+v", res)
	}
	if _, err := svc.SaveSettings(SettingsInput{KimaiURL: srv.URL, Token: "good"}); err != nil {
		t.Fatal(err)
	}
	if res := svc.CheckConnection(); !res.OK {
		t.Fatalf("configured = %+v", res)
	}
	if res := svc.LastConnection(); !res.OK {
		t.Errorf("last = %+v", res)
	}

	c, err := svc.conn.Client()
	if err != nil {
		t.Fatal(err)
	}
	if c.Location == nil || c.Location.String() != "Europe/Berlin" {
		t.Errorf("client location = %v, want Kimai user time zone", c.Location)
	}
}

func TestOfflineMessage(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	svc, _ := newTestService(t)
	if res := svc.TestConnection(url, "good"); res.Status != StatusOffline {
		t.Errorf("result = %+v", res)
	}
}
