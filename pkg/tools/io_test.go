package tools

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPadString(t *testing.T) {
	if got := PadString("hi", 8, "="); got != "== hi ==" {
		t.Fatalf("expected centered padded string, got %q", got)
	}

	if got := PadString("", 5, "-"); got != "-----" {
		t.Fatalf("expected empty content padding, got %q", got)
	}
}

func TestHeaderFooterAndSeparator(t *testing.T) {
	header := GetHeader(24, " Test")
	if !strings.Contains(header, "[hlstail] Test") || !strings.HasSuffix(header, "\r\n") {
		t.Fatalf("unexpected header output %q", header)
	}

	footer := GetFooter(10, "done")
	if !strings.Contains(footer, "done") || !strings.HasSuffix(footer, "\r\n") {
		t.Fatalf("unexpected footer output %q", footer)
	}

	separator := GetSeparator(6, "-")
	if separator != "------\r\n" {
		t.Fatalf("unexpected separator output %q", separator)
	}
}

func TestPrintBufferAndPrintLoading(t *testing.T) {
	printed := captureStdout(t, func() {
		PrintBuffer("hello")
	})

	if printed != "\033[1;1H\033[0J\r\n\r\nhello" {
		t.Fatalf("unexpected print buffer output %q", printed)
	}

	loading := captureStdout(t, func() {
		PrintLoading(30)
	})

	if !strings.Contains(loading, "Loading...") || !strings.Contains(loading, "[hlstail]") {
		t.Fatalf("unexpected loading output %q", loading)
	}
}

func TestLogToFile(t *testing.T) {
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	LogToFile("debug")

	content, err := ioutil.ReadFile(filepath.Join(tmp, "output.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "debug\n" {
		t.Fatalf("unexpected log content %q", string(content))
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
