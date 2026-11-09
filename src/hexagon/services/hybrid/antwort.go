package hybrid

import (
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Kennzeichnung der Antwort (SPEC-284 Nr. 9): der Header und sein Wert für den Replay-Fallback.
const (
	headerFallback = "X-Recorder-Fallback"
	wertReplay     = "replay"
)

// asciiKlein setzt allein ASCII-Großbuchstaben herab: Header-Namen vergleicht der Dienst ohne Beachtung der
// ASCII-Schreibweise (SPEC-284 Nr. 9).
func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// gekennzeichnet liefert die Antwort mit neuer Header-Map: ohne jeden Header X-Recorder-Fallback der Stufe, mit dem
// einen des Dienstes; die Map der Stufe bleibt unverändert, Status und Body gehen unverändert durch.
func gekennzeichnet(a inbound.Antwort, wert string) inbound.Antwort {
	h := make(map[string][]string, len(a.Header)+1)
	for name, werte := range a.Header {
		if asciiKlein(name) != asciiKlein(headerFallback) {
			h[name] = slices.Clone(werte)
		}
	}
	h[headerFallback] = []string{wert}
	a.Header = h
	return a
}

// verwirf schließt den Body einer Antwort, die zusammen mit einem Fehler zurückkam, genau einmal (SPEC-284 Nr. 10).
func verwirf(a inbound.Antwort) {
	if a.Body != nil {
		_ = a.Body.Close()
	}
}
