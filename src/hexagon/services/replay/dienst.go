package replay

import (
	"slices"
	"strings"
)

// asciiKlein setzt allein ASCII-Großbuchstaben herab: Header-Namen vergleicht der Dienst ohne
// Beachtung der ASCII-Schreibweise (SPEC-135, Kopf).
func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// werteVon nimmt die Werte aller Schreibweisen eines Namens zusammen, in der Bytefolge der Namen,
// dann in ihrer Reihenfolge (SPEC-135, Kopf); klein ist der Name in Kleinbuchstaben.
func werteVon(h map[string][]string, klein string) []string {
	var namen []string
	for n := range h {
		if asciiKlein(n) == klein {
			namen = append(namen, n)
		}
	}
	slices.Sort(namen)
	var werte []string
	for _, n := range namen {
		werte = append(werte, h[n]...)
	}
	return werte
}

// dienstFuer wählt den Dienst, dessen Root gleich pfad ist oder ihm an einer Segmentgrenze
// vorausgeht, byteweise, den mit dem längsten Root, und liefert den OData-Pfad (SPEC-135 Nr. 4).
func dienstFuer(dienste []Dienst, pfad string) (Dienst, string, bool) {
	var (
		gewaehlt Dienst
		rest     string
		gefunden bool
	)
	for _, d := range dienste {
		r, ok := unterRoot(d.Root, pfad)
		if ok && (!gefunden || len(d.Root) > len(gewaehlt.Root)) {
			gewaehlt, rest, gefunden = d, r, true
		}
	}
	if !gefunden {
		return Dienst{}, "", false
	}
	if rest == "" {
		rest = "/"
	}
	return gewaehlt, rest, true
}

func unterRoot(root, pfad string) (string, bool) {
	switch {
	case root == "/":
		return pfad, strings.HasPrefix(pfad, "/")
	case pfad == root:
		return "", true
	case strings.HasPrefix(pfad, root+"/"):
		return pfad[len(root):], true
	}
	return "", false
}

// einzelwert liefert den einen nichtleeren Wert des Headers klein (SPEC-135 Nr. 5); fehlt er, ist die
// Liste leer, ist der Wert leer oder gibt es mehr als einen, ist ok falsch.
func einzelwert(h map[string][]string, klein string) (string, bool) {
	w := werteVon(h, klein)
	if len(w) != 1 || w[0] == "" {
		return "", false
	}
	return w[0], true
}

// replaySessionGueltig prüft die Form der Replay-Session-ID: 1 bis 128 Bytes aus 0-9, A-Z, a-z, `.`,
// `_` und `-` (SPEC-135 Nr. 5).
func replaySessionGueltig(id string) bool {
	if len(id) < 1 || len(id) > 128 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		ok := c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '.' || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// zuordnung liest die drei Steuerheader (SPEC-135 Nr. 5) und liefert Workspace, Recording-Session und
// Replay-Session byte-gleich; ok ist falsch bei jedem Verstoß.
func zuordnung(h map[string][]string) (workspace, recording, replay string, ok bool) {
	workspace, okW := einzelwert(h, "x-recorder-workspace")
	recording, okR := einzelwert(h, "x-recorder-recording-session")
	replay, okS := einzelwert(h, "x-recorder-replay-session")
	return workspace, recording, replay, okW && okR && okS && replaySessionGueltig(replay)
}
