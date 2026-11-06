package jsoncodec_test

import (
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
// (der Plan nennt 13; dazu `CHARSET=utf-16`, der den Namen allein fängt).
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
