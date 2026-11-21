package tlstest_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// aufzeichnung hält die Meldungen von Fatalf fest und läuft weiter; Helper kommt vom echten Test.
type aufzeichnung struct {
	testing.TB
	meldungen []string
}

func (a *aufzeichnung) Fatalf(format string, args ...any) {
	a.meldungen = append(a.meldungen, fmt.Sprintf(format, args...))
}

// TestDateiMeldetFehler: ein Fehler beim Anlegen oder Schreiben endet mit einer Meldung, die mit `tlstest: ` beginnt.
func TestDateiMeldetFehler(t *testing.T) {
	datei := tlstest.Datei(t, t.TempDir(), "d.pem", []byte("x"), 0o600)
	a := &aufzeichnung{TB: t}
	tlstest.Datei(a, datei, "unter-einer-datei.pem", []byte("x"), 0o600)
	if len(a.meldungen) == 0 || !strings.HasPrefix(a.meldungen[0], "tlstest: ") {
		t.Fatalf("Meldungen %q, erwartet mindestens eine mit Präfix `tlstest: `", a.meldungen)
	}
}

// TestDateiRechte: die Datei trägt genau die Rechte, die der Aufruf nennt, auch wenn umask Bits nähme.
func TestDateiRechte(t *testing.T) {
	for _, modus := range []os.FileMode{0o600, 0o666} {
		pfad := tlstest.Datei(t, t.TempDir(), "k.pem", []byte("x"), modus)
		info, err := os.Stat(pfad)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if got := info.Mode().Perm(); got != modus {
			t.Errorf("Rechte %o, erwartet %o", got, modus)
		}
	}
}
