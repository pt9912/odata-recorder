package urlparser_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
)

// TestParserOhneZustand (Zeile 14, SPEC-160 Nr. 7): der Parser hat genau ein Feld, den Leser `filter` vom Kind Interface. Jedes
// weitere Feld wäre Zustand der Instanz, den ein Aufruf nach der Rückkehr des FilterLeser schreiben und am Ende lesen könnte;
// dieses Fenster hat keinen Haltepunkt, den ein Test erzwingen kann, die Gestalt des Typs ist deshalb der Fänger. Zustand in einer
// Variable des Pakets fängt `make lint` (`gochecknoglobals`), der Rest ist das akzeptierte Negativ (3) des Plans.
func TestParserOhneZustand(t *testing.T) {
	typ := reflect.TypeFor[urlparser.Parser]()
	if typ.NumField() != 1 {
		names := make([]string, typ.NumField())
		for i := range names {
			names[i] = typ.Field(i).Name
		}
		t.Fatalf("Parser hat %d Felder %v, erwartet genau eines (`filter`); SPEC-160 Nr. 7: kein veränderlicher Zustand", typ.NumField(), names)
	}
	if f := typ.Field(0); f.Name != "filter" || f.Type.Kind() != reflect.Interface {
		t.Errorf("Feld %s vom Kind %v, erwartet `filter` vom Kind Interface; SPEC-160 Nr. 7: kein veränderlicher Zustand", f.Name, f.Type.Kind())
	}
}
