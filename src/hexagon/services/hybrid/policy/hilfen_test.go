package policy_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// einzel ruft New mit einem einzigen Eintrag in der Liste des Schlüssels (Upstream aus) und liefert
// die Ursache des Fehlers, leer ohne Fehler. Ein Fehler trägt genau einen *KonfigFehler mit dem
// Schlüssel, Eintrag 1 und keine Policy.
func einzel(t *testing.T, schluessel, eintrag string) string {
	t.Helper()
	var w policy.Werte
	switch schluessel {
	case policy.SchluesselHosts:
		w.Hosts = []string{eintrag}
	case policy.SchluesselSchreibfreigaben:
		w.Schreibfreigaben = []string{eintrag}
	case policy.SchluesselReplayKategorien:
		w.ReplayKategorien = []string{eintrag}
	default:
		t.Fatalf("Schlüssel %q", schluessel)
	}
	p, err := policy.New(w)
	if err == nil {
		if p == nil {
			t.Fatalf("%s %q: weder Fehler noch Policy", schluessel, eintrag)
		}
		return ""
	}
	if p != nil {
		t.Errorf("%s %q: Policy trotz Fehler", schluessel, eintrag)
	}
	fehler := zerlege(err)
	if len(fehler) != 1 {
		t.Fatalf("%s %q: %d Fehler, erwartet 1", schluessel, eintrag, len(fehler))
	}
	var kf *policy.KonfigFehler
	if !errors.As(fehler[0], &kf) {
		t.Fatalf("%s %q: kein *KonfigFehler: %T", schluessel, eintrag, fehler[0])
	}
	if kf.Schluessel != schluessel || kf.Eintrag != 1 {
		t.Errorf("%s %q: Schlüssel %q Eintrag %d", schluessel, eintrag, kf.Schluessel, kf.Eintrag)
	}
	return kf.Ursache
}

// zerlege liefert die Fehler eines errors.Join in ihrer Reihenfolge; ein anderer Fehler ist einer.
func zerlege(err error) []error {
	var j interface{ Unwrap() []error }
	if errors.As(err, &j) {
		return j.Unwrap()
	}
	return []error{err}
}

// varianten erzeugt aus base jede Form mit einem Byte mehr (jedes Byte 0 bis 255 an jeder Stelle),
// mit einem Byte ersetzt (jedes Byte an jeder Stelle) und mit einem Byte weniger (jede Stelle).
// Die Zahl der Elemente ist (n+1)*256 + n*256 + n.
func varianten(base string) []string {
	var aus []string
	for i := 0; i <= len(base); i++ {
		for b := 0; b < 256; b++ {
			aus = append(aus, base[:i]+string([]byte{byte(b)})+base[i:])
		}
	}
	for i := 0; i < len(base); i++ {
		for b := 0; b < 256; b++ {
			aus = append(aus, base[:i]+string([]byte{byte(b)})+base[i+1:])
		}
	}
	for i := 0; i < len(base); i++ {
		aus = append(aus, base[:i]+base[i+1:])
	}
	return aus
}

func zahlVarianten(base string) int {
	n := len(base)
	return (n+1)*256 + n*256 + n
}
