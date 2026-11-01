package outboundtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen W8 bis W10 prüfen den Zeitpunkt (SPEC-172 Nr. 4, E5): `Edm.DateTime` in der Schreibweise `/Date(<ms>)/` ergibt
// `Edm.DateTimeOffset` des Modells mit der Art `datetimeoffset` und dem Zeitpunkt in UTC.

// v2ZeitGrenzen sind die Millisekunden, zwischen denen der Zeitpunkt liegen darf (SPEC-172 Nr. 4).
const (
	v2ZeitMin = -62135596800000
	v2ZeitMax = 253402300799999
)

// v2Zeitfall ist ein Wert in Millisekunden mit dem erwarteten Text des Zeitpunkts, oder `wert`.
type v2Zeitfall struct {
	ms      int64
	text    string
	gueltig bool
}

// v2ZeitBruchteil schreibt die Millisekunden f (0 bis 999) als Bruchteil ohne Nullen am Ende, leer bei null.
func v2ZeitBruchteil(f int) string {
	if f == 0 {
		return ""
	}
	return "." + strings.TrimRight(fmt.Sprintf("%03d", f), "0")
}

// v2ZeitGrenzfaelle sind die Einzelfälle: die Grenzen des Bereichs und eine Millisekunde daneben, die Null, die Nachbarn der Null
// und die Tage um die Schalttage 1900 (kein Schaltjahr), 2000 (Schaltjahr) und 2100 (kein Schaltjahr).
func v2ZeitGrenzfaelle() []v2Zeitfall {
	out := []v2Zeitfall{
		{v2ZeitMin, "0001-01-01T00:00:00Z", true}, {v2ZeitMin - 1, "", false}, {v2ZeitMax, "9999-12-31T23:59:59.999Z", true}, {v2ZeitMax + 1, "", false},
		{0, "1970-01-01T00:00:00Z", true}, {-1, "1969-12-31T23:59:59.999Z", true}, {1, "1970-01-01T00:00:00.001Z", true},
		{-62135596799999, "0001-01-01T00:00:00.001Z", true}, {253402300799000, "9999-12-31T23:59:59Z", true}, {253402300800000, "", false},
	}
	for _, j := range []int{1900, 2000, 2100} {
		for _, d := range []struct{ m, t int }{{2, 28}, {2, 29}, {3, 1}} {
			z := time.Date(j, time.Month(d.m), d.t, 0, 0, 0, 0, time.UTC)
			// 29. Februar 1900 und 2100 gibt es nicht: time.Date rechnet auf den 1. März, die Eingabe nennt dann den 1. März.
			out = append(out, v2Zeitfall{z.UnixMilli(), fmt.Sprintf("%04d-%02d-%02dT00:00:00Z", z.Year(), int(z.Month()), z.Day()), true})
		}
	}
	return out
}

// v2ZeitListe ist die Menge des Generators: jedes Jahr von 1 bis 9999 mal jeder Monat mal {erster Tag 00:00:00.000, letzter Tag
// 23:59:59.999} (239 976 Fälle) und die Bruchteile 0 bis 999 am 1970-01-01T00:00:00 (1000 Fälle). Das Orakel ist das Paket
// time für die Millisekunden und Sprintf der Felder für den Text.
func v2ZeitListe() []v2Zeitfall {
	liste := make([]v2Zeitfall, 0, v2ZeitZahl)
	for j := 1; j <= 9999; j++ {
		for m := 1; m <= 12; m++ {
			erster := time.Date(j, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
			liste = append(liste, v2Zeitfall{erster.UnixMilli(), fmt.Sprintf("%04d-%02d-01T00:00:00Z", j, m), true})
			tag := time.Date(j, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
			letzter := time.Date(j, time.Month(m), tag, 23, 59, 59, 999e6, time.UTC)
			liste = append(liste, v2Zeitfall{letzter.UnixMilli(), fmt.Sprintf("%04d-%02d-%02dT23:59:59.999Z", j, m, tag), true})
		}
	}
	for f := range 1000 {
		liste = append(liste, v2Zeitfall{int64(f), "1970-01-01T00:00:00" + v2ZeitBruchteil(f) + "Z", true})
	}
	return liste
}

// v2ZeitZahl ist die Zahl der Körper des Generators: 9999 Jahre × 12 Monate × 2 Zeitpunkte und 1000 Bruchteile.
const v2ZeitZahl = 9999*12*2 + 1000

// v2ZeitBody ist der Body zu den Millisekunden: die Property `NDateTimeOffset` mit dem Wert, als Entity oder als Collection.
func v2ZeitBody(ms int64, einzel bool) []byte {
	members := `"Id":1,"NDateTimeOffset":` + v2Dt(strconv.FormatInt(ms, 10))
	if einzel {
		return v2Ein(members)
	}
	return v2Kol("{" + members + "}")
}

// v2ZeitPruefer prüft einen Körper des Generators: gültig mit dem Text des Orakels oder `wert` am Ort der Property; die Methode
// wechselt mit der Nummer.
func (s *jsuite) v2ZeitPruefer(liste []v2Zeitfall) gPruefer {
	return func(b []byte, nr int64) (klasse, abweichung string) {
		z := liste[nr]
		einzel := nr%2 == 0
		r := s.aufruf(context.Background(), einzel, bytes.NewReader(b), s.modelle["W"], "Ws", 1<<30)
		ort := "/d/results/0/NDateTimeOffset"
		if einzel {
			ort = "/d/NDateTimeOffset"
		}
		var df *model.DatenFehler
		switch {
		case z.gueltig:
			es := r.entities(jfall{einzel: einzel})
			if r.err != nil || len(es) != 1 || len(es[0].Properties) != 2 || es[0].Properties[1] != pw("NDateTimeOffset", model.ArtDateTimeOffset, z.text) {
				return "gültig", fmt.Sprintf("%d ms: Ergebnis %v %+v, erwartet %s", z.ms, r.err, es, z.text)
			}
			return "gültig", ""
		case !errors.As(r.err, &df) || df.Ursache != model.DatenWert || df.Ort != ort:
			return "wert", fmt.Sprintf("%d ms: Ergebnis %v, erwartet wert", z.ms, r.err)
		}
		return "wert", ""
	}
}

func v2ZeitFaelle(liste []v2Zeitfall) []wertFall {
	var out []wertFall
	for _, z := range liste {
		w := gutW(v2Dt(strconv.FormatInt(z.ms, 10)), model.ArtDateTimeOffset, z.text)
		if !z.gueltig {
			w = schlechtW(v2Dt(strconv.FormatInt(z.ms, 10)), model.DatenWert)
		}
		out = append(out, w)
	}
	return out
}

// v2W8: der Generator der 240 976 Zeitpunkte gegen das Orakel (jedes Jahr, jeder Monat, erster und letzter Tag, jeder Bruchteil)
// und die Einzelfälle der Grenzen; die Zählprüfung nennt die Menge. Unter dem Detektor läuft die Probe: jedes 97. Element.
func v2W8(t *testing.T, s *jsuite) {
	liste := v2ZeitListe()
	if len(liste) != v2ZeitZahl {
		t.Fatalf("W8: %d Zeitpunkte, erwartet %d", len(liste), v2ZeitZahl)
	}
	if unterDetektor {
		var probe []v2Zeitfall
		for i := 0; i < len(liste); i += 97 {
			probe = append(probe, liste[i])
		}
		liste = probe
	}
	folge := func(yield func([]byte) bool) {
		for i, z := range liste {
			if !yield(v2ZeitBody(z.ms, i%2 == 0)) {
				return
			}
		}
	}
	gZaehle(t, "W8 Generator", gFahre(t, "W8", folge, s.v2ZeitPruefer(liste)), int64(len(liste)))
	s.pruefeAlle(t, v2FaelleW8())
	// Der Wert des Modells Edm.DateTimeOffset (aus Edm.DateTime) bleibt datetimeoffset, nicht date und nicht string (SPEC-170 Nr. 4).
	r := s.lauf(jfall{m: "W", set: "Ws", einzel: true, body: v2ZeitBody(0, true)})
	if r.err != nil || r.entity.Properties[1].Wert.Art != model.ArtDateTimeOffset || r.entity.Properties[1].Wert.Art == model.ArtDate || r.entity.Properties[1].Wert.Art == model.ArtString {
		t.Errorf("W8 Art: %+v %v", r.entity, r.err)
	}
}

// v2FaelleW8 sind die Fälle der Zeilen W8 bis W10 als Fälle der Suite: die Grenzfälle des Zeitpunkts, die Formen des Textes, die
// Schreibweisen des Schrägstrichs.
func v2FaelleW8() []jfall {
	return verbinde(v2WertFaelle("W8", "Edm.DateTimeOffset", v2ZeitFaelle(v2ZeitGrenzfaelle()), false), v2FaelleW9(), v2FaelleW10())
}

// v2Datum schreibt den Inhalt als JSON-String: der Schrägstrich als Backslash und Schrägstrich, wie V2 ihn schreibt.
func v2Datum(inner string) string { return `"` + strings.ReplaceAll(inner, "/", "\x5c/") + `"` }

// v2DatumFormen sind mindestens 30 Formen des Textes `/Date(...)/`, die `wert` sind (SPEC-172 Nr. 4, E5): ein Zusatz hinter der
// Zahl, `-0`, ein `+`, eine führende Null, die leere Zahl, eine Dezimalzahl, ein Exponent, eine Zahl über int64, eine andere
// Schreibweise von Date, fehlende oder zusätzliche Zeichen, ein Zeitpunkt außerhalb des Bereichs, ein ISO-String.
func v2DatumFormen() []string {
	return []string{
		"/Date(0+0060)/", "/Date(0-0100)/", "/Date(0+0000)/", "/Date(0 )/", "/Date( 0)/", "/Date(-0)/", "/Date(+5)/", "/Date(05)/", "/Date(-05)/",
		"/Date()/", "/Date(-)/", "/Date(1.5)/", "/Date(1e3)/", "/Date(1E3)/", "/Date(9223372036854775808)/", "/Date(12345678901234567890)/",
		"/Date(-9223372036854775809)/", "/Date(18446744073709551617)/", "/Date(-18446744073709551617)/", "/date(0)/", "/DATE(0)/", "/Date(0)", "/Date(0)/ ", " /Date(0)/", "Date(0)", "/Date(0)//", "/Date(0/", "/Date0)/",
		"/Date(0))/", "/Date((0))/", "/Date(0)/x", "x/Date(0)/", "/Date(--5)/", "/Date(5-)/", "/Date(0x10)/", "//Date(0)/",
		"/Date(" + strconv.FormatInt(v2ZeitMin-1, 10) + ")/", "/Date(" + strconv.FormatInt(v2ZeitMax+1, 10) + ")/", "/Date(-99999999999999)/",
		"/Date(0,5)/", "/Date(0.0)/", "2020-01-01T00:00:00Z", "2020-01-01", "", "0", "/Date(/", "/Date(a)/", "/Date(\xd9\xa3)/",
	}
}

func v2FaelleW9() []jfall {
	var out []jfall
	for _, f := range v2DatumFormen() {
		out = append(out, v2WertFaelle("W9", "Edm.DateTimeOffset", []wertFall{{json: v2Datum(f), ursache: model.DatenWert, name: fmt.Sprintf("%q", f)}}, false)...)
	}
	for _, j := range []string{`0`, `915148800000`, `true`, `false`, `{}`, `[]`, `{"ms":0}`} {
		out = append(out, v2WertFaelle("W9", "Edm.DateTimeOffset", []wertFall{schlechtW(j, model.DatenTyp)}, false)...)
	}
	for _, g := range []struct{ ms, text string }{
		{"0", "1970-01-01T00:00:00Z"}, {"-1", "1969-12-31T23:59:59.999Z"}, {"915148800000", "1999-01-01T00:00:00Z"}, {"1057017600000", "2003-07-01T00:00:00Z"},
		{"1", "1970-01-01T00:00:00.001Z"}, {"10", "1970-01-01T00:00:00.01Z"}, {"100", "1970-01-01T00:00:00.1Z"}, {"1000", "1970-01-01T00:00:01Z"},
	} {
		out = append(out, v2WertFaelle("W9 gültig", "Edm.DateTimeOffset", []wertFall{gutW(v2Dt(g.ms), model.ArtDateTimeOffset, g.text)}, false)...)
	}
	return out
}

func v2W9(t *testing.T, s *jsuite) {
	if n := len(v2DatumFormen()); n < 30 {
		t.Errorf("W9: %d Formen, erwartet mindestens 30", n)
	}
	s.pruefeAlle(t, v2FaelleW9())
}

// v2SchreibweisenDatum liefert zu den Millisekunden die Schreibweisen desselben Textes `/Date(<ms>)/`: mit Backslash vor dem
// Schrägstrich, ohne ihn, und mit einem einzigen Zeichen als Escape aus Backslash, `u` und vier Hex-Ziffern (klein und groß), an
// jeder Stelle des Textes; der Escape des Schrägstrichs steht auch als `\/`.
func v2SchreibweisenDatum(ms string) []string {
	inner := "/Date(" + ms + ")/"
	out := []string{`"` + inner + `"`, v2Datum(inner), `"` + esc("002F") + `Date(` + ms + `)` + esc("002f") + `"`}
	for i := range len(inner) {
		hex := fmt.Sprintf("%04x", inner[i])
		for _, h := range []string{hex, strings.ToUpper(hex)} {
			out = append(out, `"`+inner[:i]+esc(h)+inner[i+1:]+`"`)
		}
	}
	return out
}

func v2FaelleW10() []jfall {
	var out []jfall
	for _, g := range []struct{ ms, text string }{
		{"0", "1970-01-01T00:00:00Z"}, {"915148800000", "1999-01-01T00:00:00Z"}, {"-1", "1969-12-31T23:59:59.999Z"}, {"253402300799999", "9999-12-31T23:59:59.999Z"},
	} {
		for _, j := range v2SchreibweisenDatum(g.ms) {
			out = append(out, v2WertFaelle("W10", "Edm.DateTimeOffset", []wertFall{gutW(j, model.ArtDateTimeOffset, g.text)}, false)...)
		}
	}
	return out
}

func v2W10(t *testing.T, s *jsuite) {
	if n := len(v2SchreibweisenDatum("0")); n != 3+2*len("/Date(0)/") {
		t.Errorf("W10: %d Schreibweisen von /Date(0)/, erwartet %d", n, 3+2*len("/Date(0)/"))
	}
	s.pruefeAlle(t, v2FaelleW10())
}
