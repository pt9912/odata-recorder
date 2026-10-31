package baselineimport

import "github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"

// Hilfen für die externen Tests: die reinen Funktionen des Hashes der Anfrage, der Kanonform, der Ordnung von
// unsupported_features.entities, des Content-Type und der ASCII-Kleinschreibung.

// AnfrageHash ist anfrageHash (SPEC-127 Nr. 3 f).
func AnfrageHash(a inbound.ImportAnfrage) string { return anfrageHash(a) }

// Kanon ist kanon (SPEC-127 Nr. 14); v ist ein Wert aus nil, bool, int, int64, string, []any und map[string]any.
func Kanon(v any) []byte { return kanon(v) }

// VergleicheZeilen ist die Ordnung der Einträge von unsupported_features.entities (SPEC-127 Nr. 14 b); ein
// Eintrag ist (entity_set, name, reason).
func VergleicheZeilen(a, b [3]string) int {
	return vergleicheZeilen(zeile{a[0], a[1], a[2]}, zeile{b[0], b[1], b[2]})
}

// Inhaltstyp ist inhaltstypOK (SPEC-127 Nr. 8 g): h sind die Header der Antwort, medien die zulässigen Medientypen.
func Inhaltstyp(h map[string][]string, medien ...string) (ok, ohneKontext bool) {
	return inhaltstypOK(h, medien...)
}

// ASCIIKlein ist asciiKlein (SPEC-127 Nr. 8 g, ASCII-klein).
func ASCIIKlein(s string) string { return asciiKlein(s) }

// OptionenPaare ist optionen (SPEC-127 Nr. 8 i, SPEC-162 Nr. 1): je Option der nachsichtig dekodierte Name und der Wert.
func OptionenPaare(raw string) [][2]string {
	var out [][2]string
	for _, o := range optionen(raw) {
		out = append(out, [2]string{o.name, o.wert})
	}
	return out
}
