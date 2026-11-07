package jsoncodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eCT prüft eine Liste von Kopfzeilen: gültig liefert den Body, ungültig `content_type` ohne Ort und ohne Read.
func eCT(t *testing.T, was string, art model.EingabeArt, ct []string, gueltig bool) {
	t.Helper()
	s := eStromText(eObj("Id", "1"))
	a := eA("Mitten", art, "")
	a.src, a.ct = s, ct
	got, err := a.los()
	if gueltig {
		eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.Mitte", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
		return
	}
	eFehler(t, was, got, err, model.EingabeContentType, "")
	if s.readN != 0 {
		t.Errorf("%s: %d Reads vor dem Content-Type", was, s.readN)
	}
}

// TestEingabeMedientyp hält Zeile 5: nur `application/json` ohne Beachtung der ASCII-Schreibweise, Rand getrimmt. Menge 17 × 3.
func TestEingabeMedientyp(t *testing.T) {
	gueltig := []string{"application/json", "Application/JSON", "APPLICATION/JSON", " \tapplication/json\t ", "application/json;"}
	ungueltig := []string{"", "   ", "text/plain", "application/xml", "application/atom+xml", "application/jsonx", "xapplication/json",
		"application/json/", "application/", "json", "application/json x", "application/j\xc5\xbfon"}
	n := 0
	for _, art := range eArten() {
		for _, ct := range gueltig {
			eCT(t, "gültig "+ct, art, []string{ct}, true)
			n++
		}
		for _, ct := range ungueltig {
			eCT(t, "ungültig "+ct, art, []string{ct}, false)
			n++
		}
	}
	if n != 17*3 {
		t.Errorf("Menge %d, erwartet 51", n)
	}
}

// TestEingabeCharset hält Zeile 6: jedes Vorkommen von `charset` muss `utf-8` sein, der Name zählt ohne Beachtung der Schreibweise. Menge 14 × 3
// (6 gültige und 8 ungültige Formen; `CHARSET=utf-16` fängt den Namen allein).
func TestEingabeCharset(t *testing.T) {
	gueltig := []string{"application/json; charset=utf-8", "application/json;charset=UTF-8", "application/json;CHARSET=Utf-8",
		"application/json; charset = utf-8", "application/json;odata.metadata=minimal", "application/json;;charset=utf-8"}
	ungueltig := []string{"application/json;charset=iso-8859-1", "application/json;charset=utf-16", `application/json;charset="utf-8"`,
		"application/json;charset=", "application/json;charset", "application/json;charset=utf-8;charset=utf-16",
		"application/json;charset=utf-16;charset=utf-8", "application/json;CHARSET=utf-16"}
	n := 0
	for _, art := range eArten() {
		for _, ct := range gueltig {
			eCT(t, "gültig "+ct, art, []string{ct}, true)
			n++
		}
		for _, ct := range ungueltig {
			eCT(t, "ungültig "+ct, art, []string{ct}, false)
			n++
		}
	}
	if n != 14*3 {
		t.Errorf("Menge %d, erwartet 42", n)
	}
}

// TestEingabeCharsetPositionen hält Zeile 6 für die Klasse „Parameter `charset` an jeder Position mit jedem Rand“: 4 Positionen (hinter 0, 1, 2, 3 anderen
// Parametern) × 81 Ränder (Leerzeichen, Tabulator oder keiner in den vier Lücken vor dem Namen, hinter dem Namen, hinter dem `=`, hinter dem Wert) ×
// Wert (`utf-8` gültig; `iso-8859-1`, `utf-16`, `iso-8859-1=x` ungültig) = 4 × 81 × 4 = 1296 Elemente × 3 Arten. Dazu andere Randzeichen (CR, LF, FF, VT,
// NBSP): am Medientyp und am Wert sind sie kein Rand und machen den Medientyp bzw. den Wert ungültig, am Namen ist der Teil kein `charset` und ohne Wirkung.
// Das Orakel ist die Konstruktion der Elemente, nie der Leser.
func TestEingabeCharsetPositionen(t *testing.T) {
	raender := []string{"", " ", "\t"}
	n := 0
	for pos := range 4 {
		for _, wert := range []string{"utf-8", "iso-8859-1", "utf-16", "iso-8859-1=x"} {
			for _, r1 := range raender {
				for _, r2 := range raender {
					for _, r3 := range raender {
						for _, r4 := range raender {
							arg := r1 + "charset" + r2 + "=" + r3 + wert + r4
							ct := "application/json" + strings.Repeat(";a=b", pos) + ";" + arg + strings.Repeat(";c=d", 3-pos)
							for _, art := range eArten() {
								eCT(t, ct, art, []string{ct}, wert == "utf-8")
								n++
							}
						}
					}
				}
			}
		}
	}
	if n != 4*4*81*3 {
		t.Errorf("Menge %d, erwartet %d", n, 4*4*81*3)
	}
	// andere Randzeichen
	for _, c := range []string{"\r", "\n", "\f", "\v", "\xc2\xa0"} {
		for _, art := range eArten() {
			eCT(t, "Medientyp davor", art, []string{c + "application/json"}, false)
			eCT(t, "Medientyp dahinter", art, []string{"application/json" + c}, false)
			eCT(t, "Wert davor", art, []string{"application/json;charset=" + c + "utf-8"}, false)
			eCT(t, "Wert dahinter", art, []string{"application/json;charset=utf-8" + c}, false)
			eCT(t, "Name davor", art, []string{"application/json;" + c + "charset=utf-16"}, true)
			eCT(t, "Name dahinter", art, []string{"application/json;charset" + c + "=utf-16"}, true)
		}
	}
}

// TestEingabeKopfzeilen hält Zeile 7: genau eine Kopfzeile; keine und mehr als eine sind `content_type`, ohne Read. Menge 7 × 3.
func TestEingabeKopfzeilen(t *testing.T) {
	g := "application/json"
	listen := []struct {
		name    string
		ct      []string
		gueltig bool
	}{
		{"nil", nil, false}, {"leer", []string{}, false}, {"eine gültige", []string{g}, true}, {"zwei gültige", []string{g, g}, false},
		{"gültig und leer", []string{g, ""}, false}, {"leer und gültig", []string{"", g}, false}, {"drei gültige", []string{g, g, g}, false},
	}
	n := 0
	for _, art := range eArten() {
		for _, l := range listen {
			eCT(t, l.name, art, l.ct, l.gueltig)
			n++
		}
	}
	if n != 21 {
		t.Errorf("Menge %d, erwartet 21", n)
	}
}

// TestEingabeContentTypeLiestNicht hält Zeile 7 für jeden Verstoß der Zeilen 5 bis 7 samt einem Verstoß im Body: `content_type`
// geht dem Body voraus, `src` wird nicht gelesen.
func TestEingabeContentTypeLiestNicht(t *testing.T) {
	for _, ct := range [][]string{nil, {"text/plain"}, {"application/json;charset=utf-16"}, {"application/json", "application/json"}} {
		for _, art := range eArten() {
			s := eStromText(`{"Id":`)
			a := eA("Mitten", art, "")
			a.src, a.ct = s, ct
			got, err := a.los()
			eFehler(t, "Verstoß im Body und im Content-Type", got, err, model.EingabeContentType, "")
			if s.readN != 0 {
				t.Errorf("%v: %d Reads", ct, s.readN)
			}
		}
	}
}

// TestEingabeParameterOhneWirkung hält Zeile 8: außer `charset` hat kein Parameter eine Wirkung, auch `IEEE754Compatible`.
// Menge 60 = 5 Parameter × 2 Zahlformen × 2 Typen × 3 Arten.
func TestEingabeParameterOhneWirkung(t *testing.T) {
	params := []string{"IEEE754Compatible=true", "IEEE754Compatible=false", "odata.metadata=none", "odata.streaming=true", "foo=bar"}
	typen := []struct {
		name, zahl, text string
		art              model.WertArt
	}{{"I64", "5", "5", model.ArtInt64}, {"Dec", "1.5", "1.5", model.ArtDecimal}}
	n := 0
	for _, p := range params {
		for _, form := range []string{"Zahl", "Zeichenkette"} {
			for _, ty := range typen {
				for _, art := range eArten() {
					wert := ty.zahl
					if form == "Zeichenkette" {
						wert = `"` + ty.zahl + `"`
					}
					a := eA("Ws", art, eObj("Id", "1", ty.name, wert))
					a.ct = []string{"application/json;" + p}
					got, err := a.los()
					eOK(t, p+" "+form+" "+ty.name, got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1",
						Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW(ty.name, ty.art, ty.text)}})
					n++
				}
			}
		}
	}
	if n != 60 {
		t.Errorf("Menge %d, erwartet 60", n)
	}
}
