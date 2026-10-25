package baselineimport

import "github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"

// Hilfen für die externen Tests: die reinen Funktionen des Hashes der Anfrage und der Kanonform.

// AnfrageHash ist anfrageHash (SPEC-127 Nr. 3 f).
func AnfrageHash(a inbound.ImportAnfrage) string { return anfrageHash(a) }

// Kanon ist kanon (SPEC-127 Nr. 14); v ist ein Wert aus nil, bool, int, int64, string, []any und map[string]any.
func Kanon(v any) []byte { return kanon(v) }
