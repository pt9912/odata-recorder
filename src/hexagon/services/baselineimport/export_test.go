package baselineimport

import "github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"

// Hilfen für die externen Tests: die reinen Funktionen des Hashes der Anfrage, der Kanonform und der
// Ordnung von unsupported_features.entities.

// AnfrageHash ist anfrageHash (SPEC-127 Nr. 3 f).
func AnfrageHash(a inbound.ImportAnfrage) string { return anfrageHash(a) }

// Kanon ist kanon (SPEC-127 Nr. 14); v ist ein Wert aus nil, bool, int, int64, string, []any und map[string]any.
func Kanon(v any) []byte { return kanon(v) }

// VergleicheZeilen ist die Ordnung der Einträge von unsupported_features.entities (SPEC-127 Nr. 14 b); ein
// Eintrag ist (entity_set, name, reason).
func VergleicheZeilen(a, b [3]string) int {
	return vergleicheZeilen(zeile{a[0], a[1], a[2]}, zeile{b[0], b[1], b[2]})
}
