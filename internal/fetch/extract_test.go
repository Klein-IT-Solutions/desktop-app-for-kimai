package fetch

import "testing"

func TestSafeJoinClamps(t *testing.T) {
	for in, want := range map[string]string{
		"../../etc/passwd":    "/data/etc/passwd",
		"llama-b1/bin/server": "/data/llama-b1/bin/server",
		`..\..\win.ini`:       "/data/win.ini",
	} {
		if p, err := safeJoin("/data", in); err != nil || p != want {
			t.Errorf("safeJoin(%q) = %q, %v; want %q", in, p, err, want)
		}
	}
}
