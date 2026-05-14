package term

import (
	"io"
	"os"
	"testing"
)

func TestNewSession(t *testing.T) {
	sess := NewSession()
	if sess == nil {
		t.Fatal("expected session")
	}
}

func TestStartPrintsTerminalSetup(t *testing.T) {
	printed := captureStdout(t, func() {
		NewSession().Start()
	})

	if printed != "\033[1;1H\033[0J\033[?25l" {
		t.Fatalf("unexpected start output %q", printed)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}
