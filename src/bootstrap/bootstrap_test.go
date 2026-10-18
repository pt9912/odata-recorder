package bootstrap_test

import (
	"bytes"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"ohne Argument gibt die Version aus", nil, 0, "odata-recorder 1.2.3\n", ""},
		{"version gibt die Version aus", []string{"version"}, 0, "odata-recorder 1.2.3\n", ""},
		{"unbekanntes Kommando endet mit 2", []string{"foo"}, 2, "", "Unbekanntes Kommando: foo\n"},
		{"version mit Zusatzargument endet mit 2", []string{"version", "x"}, 2, "", "Unbekanntes Kommando: version\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := bootstrap.Run(tc.args, "1.2.3", &stdout, &stderr)
			if code != tc.wantCode {
				t.Errorf("Exit-Code = %d, erwartet %d", code, tc.wantCode)
			}
			if stdout.String() != tc.wantStdout {
				t.Errorf("stdout = %q, erwartet %q", stdout.String(), tc.wantStdout)
			}
			if stderr.String() != tc.wantStderr {
				t.Errorf("stderr = %q, erwartet %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}
