package model_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestDatenFehlerText (Zeile K1): der Text nennt Code und Ursache, der Ort steht in Anführungszeichen dahinter.
func TestDatenFehlerText(t *testing.T) {
	const kopf = "ODR-E6101: Baseline-Import ungültig oder inkonsistent: "
	text := func(u model.DatenUrsache, ort string) string {
		return (&model.DatenFehler{Ursache: u, Ort: ort}).Error()
	}
	if got := text(model.DatenSyntax, ""); got != kopf+"syntax" {
		t.Errorf("ohne Ort: %q", got)
	}
	if got := text(model.DatenTyp, "/value/0/ID"); got != kopf+`typ "/value/0/ID"` {
		t.Errorf("mit Ort: %q", got)
	}
	if got := text(model.DatenUnbekannteProperty, "/a\"b\n"); got != kopf+`unbekannte_property "/a\"b\n"` {
		t.Errorf("mit Anführungszeichen und Zeilenumbruch im Ort: %q", got)
	}
}

// TestDatenFehlerCode (Zeile K1): Code und Unwrap, auch durch Verpacken.
func TestDatenFehlerCode(t *testing.T) {
	e := &model.DatenFehler{Ursache: model.DatenSchluessel, Ort: "/Id"}
	if e.Code() != model.CodeE6101 {
		t.Errorf("Code %v, erwartet ODR-E6101", e.Code())
	}
	var err error = e
	if errors.Unwrap(err) != model.ErrDatenUngueltig || !errors.Is(fmt.Errorf("lesen: %w", err), model.ErrDatenUngueltig) {
		t.Error("ErrDatenUngueltig über Unwrap oder errors.Is nicht erreichbar")
	}
	var df *model.DatenFehler
	if !errors.As(fmt.Errorf("lesen: %w", err), &df) || df != e {
		t.Error("errors.As über Verpackung")
	}
}

// TestDatenUrsachen: die fünfzehn Ursachen sind verschieden und tragen die Schreibweise von SPEC-126 Nr. 14.
func TestDatenUrsachen(t *testing.T) {
	soll := []string{
		"syntax", "kodierung", "tiefe", "zu_gross", "doppelter_name", "struktur", "kontext", "odata_typ", "unbekannte_property",
		"typ", "wert", "null", "zahl_zu_lang", "schluessel", "entity_set",
	}
	ist := []model.DatenUrsache{
		model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName, model.DatenStruktur,
		model.DatenKontext, model.DatenODataTyp, model.DatenUnbekannteProperty, model.DatenTyp, model.DatenWert, model.DatenNull,
		model.DatenZahlZuLang, model.DatenSchluessel, model.DatenEntitySet,
	}
	if len(ist) != len(soll) {
		t.Fatalf("%d Ursachen, erwartet %d", len(ist), len(soll))
	}
	for i := range soll {
		if string(ist[i]) != soll[i] {
			t.Errorf("Ursache %d: %q, erwartet %q", i, ist[i], soll[i])
		}
	}
}

// TestWertArten: die zehn Arten tragen die Schreibweise von SPEC-126 Nr. 10, die drei Gründe die von Nr. 9.
func TestWertArten(t *testing.T) {
	soll := map[model.WertArt]string{
		model.ArtNull: "null", model.ArtString: "string", model.ArtBoolean: "boolean", model.ArtInt32: "int32", model.ArtInt64: "int64",
		model.ArtDecimal: "decimal", model.ArtGUID: "guid", model.ArtDate: "date", model.ArtDateTimeOffset: "datetimeoffset",
		model.ArtRedigiert: "redigiert",
	}
	for art, text := range soll {
		if string(art) != text {
			t.Errorf("Art %q, erwartet %q", art, text)
		}
	}
	gruende := map[model.UebersprungGrund]string{model.UebersprungTyp: "typ", model.UebersprungNavigation: "navigation", model.UebersprungOffen: "offen"}
	for g, text := range gruende {
		if string(g) != text {
			t.Errorf("Grund %q, erwartet %q", g, text)
		}
	}
}
