package jsoncodec_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eBodyForm ist ein Element der Menge der Body-Formen (zweite Tabelle, Zeilen 9 bis 15): der Body und die Erwartung (Ursache; leer: die Wurzel
// ist ein Objekt und der Body gültig).
type eBodyForm struct {
	name, body string
	ursache    model.EingabeUrsache
}

// eBodyFormen sind die 43 Elemente: 7 Wurzeln × 5 Zusätze und 8 BOM-Formen. Das Orakel ist die Tabelle: eine Wurzel, die kein Objekt ist, ist
// `struktur`, jeder Zusatz hinter dem Wert, der kein Leerraum ist, ist `syntax`.
func eBodyFormen() []eBodyForm {
	wurzeln := []string{"{}", "[]", `""`, "0", "true", "false", "null"}
	zusaetze := []struct{ name, davor, dahinter string }{
		{"keiner", "", ""}, {"Leerraum", " \t\r\n", " \t\r\n"}, {"x", "", "x"}, {"{}", "", " {}"}, {"Komma", "", ","},
	}
	var out []eBodyForm
	for _, w := range wurzeln {
		for _, z := range zusaetze {
			u := model.EingabeUrsache("")
			switch {
			case z.dahinter != "" && strings.TrimLeft(z.dahinter, " \t\r\n") != "":
				u = model.EingabeSyntax
			case w != "{}":
				u = model.EingabeStruktur
			}
			out = append(out, eBodyForm{"Wurzel " + w + " Zusatz " + z.name, z.davor + w + z.dahinter, u})
		}
	}
	for _, b := range []struct {
		name, bytes string
		u           model.EingabeUrsache
	}{
		{"EF BB BF", "\xef\xbb\xbf", model.EingabeKodierung}, {"FF FE", "\xff\xfe", model.EingabeKodierung}, {"FE FF", "\xfe\xff", model.EingabeKodierung},
		{"EF", "\xef", model.EingabeSyntax}, {"EF BB", "\xef\xbb", model.EingabeSyntax}, {"EF BB BE", "\xef\xbb\xbe", model.EingabeSyntax},
		{"FF", "\xff", model.EingabeSyntax}, {"FE", "\xfe", model.EingabeSyntax},
	} {
		out = append(out, eBodyForm{"BOM " + b.name, b.bytes + "{}", b.u})
	}
	return out
}

// TestEingabeBodyFormen hält die zweite Tabelle der Zeilen 9 bis 15 als Generator: Wurzeln × Zusätze und die BOM-Formen. Menge 43 × 3 Arten.
func TestEingabeBodyFormen(t *testing.T) {
	n := 0
	for _, f := range eBodyFormen() {
		for _, art := range eArten() {
			got, err := eA("Mitten", art, f.body).los()
			if f.ursache == "" {
				eMitteLeer(t, f.name, art, got, err)
			} else {
				eFehler(t, f.name+" "+string(art), got, err, f.ursache, "")
			}
			n++
		}
	}
	if n != 43*3 {
		t.Errorf("Menge %d, erwartet 129", n)
	}
}

// eZeichen ist ein Element der Menge der Zeichen (zweite Tabelle, 101): der Text im JSON-String, der dekodierte Text oder die Ursache.
type eZeichen struct {
	json, dec string
	ursache   model.EingabeUrsache
}

// eZeichenMenge sind die 101 Zeichen und Folgen: gültige Escapes 48, ungültige Escapes 6, einzelne Surrogate 5, rohe Steuerzeichen 32, rohes U+007F,
// Mehrbyte 3, ungültige UTF-8-Folgen 6; Escape-Folgen aus Bytewerten (AGENTS.md §3.18), die Hex-Ziffern in drei Schreibweisen.
func eZeichenMenge() []eZeichen {
	var z []eZeichen
	for _, k := range []struct{ kurz, dec string }{{`"`, `"`}, {bs, bs}, {"/", "/"}, {"b", "\b"}, {"f", "\f"}, {"n", "\n"}, {"r", "\r"}, {"t", "\t"}} {
		z = append(z, eZeichen{bs + k.kurz, k.dec, ""})
	}
	for c := range 32 {
		z = append(z, eZeichen{eJSONZeichen(string(rune(c)), c), string(rune(c)), ""})
	}
	for i, r := range []rune{0x7F, 0xE9, 0xD7FF, 0xE000, 0xFFFD, 0xFEFF} {
		z = append(z, eZeichen{eJSONZeichen(string(r), i), string(r), ""})
	}
	z = append(z, eZeichen{esc("D800") + esc("DC00"), string(rune(0x10000)), ""}, eZeichen{esc("dbff") + esc("DFFF"), string(rune(0x10FFFF)), ""})
	// Backslash am Ende des Strings steht in TestEingabeBackslashAmEnde: an Anfang und Mitte ergäbe er mit dem Folgezeichen ein gültiges Escape.
	for _, s := range []string{bs + "a", bs + "x", bs + "U0041", bs + "u123G", bs + "u00G0", bs + "1"} {
		z = append(z, eZeichen{s, "", model.EingabeSyntax})
	}
	for _, s := range []string{esc("D800"), esc("DBFF"), esc("DC00"), esc("DFFF"), esc("D800") + esc("0041")} {
		z = append(z, eZeichen{s, "", model.EingabeKodierung})
	}
	for c := range 32 {
		z = append(z, eZeichen{string([]byte{byte(c)}), "", model.EingabeSyntax})
	}
	z = append(z, eZeichen{"\x7f", "\x7f", ""}, eZeichen{"\xc3\xa9", "\xc3\xa9", ""}, eZeichen{"\xe2\x82\xac", "\xe2\x82\xac", ""},
		eZeichen{"\xf0\x9f\x98\x80", "\xf0\x9f\x98\x80", ""})
	for _, s := range []string{"\xc0\x80", "\xc3", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xff", "\x80"} {
		z = append(z, eZeichen{s, "", model.EingabeKodierung})
	}
	return z
}

func eOrtVon(name string) string {
	return "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(name)
}

// TestEingabeEscapeRohbytes hält die Menge der Escape-Folgen und Rohbytes in Namen und Werten: 101 Zeichen × 4 Stellen × 3 Seiten mit Art PATCH,
// dazu 101 × 2 Arten am Wert. Menge 1414; das Orakel ist die Tabelle.
func TestEingabeEscapeRohbytes(t *testing.T) {
	zeichen := eZeichenMenge()
	if len(zeichen) != 101 {
		t.Fatalf("%d Zeichen, erwartet 101", len(zeichen))
	}
	seiten := []func(c string) string{
		func(c string) string { return c + "ab" }, func(c string) string { return "a" + c + "b" }, func(c string) string { return "ab" + c },
	}
	n := 0
	stellen := []struct {
		name string
		body func(inhalt string) string
		ok   func(dec string) (model.EntityEingabe, model.EingabeUrsache, string)
	}{
		{"Name einer Property", func(i string) string { return `{"Id":1,"` + i + `":1}` },
			func(dec string) (model.EntityEingabe, model.EingabeUrsache, string) {
				return model.EntityEingabe{}, model.EingabeUnbekannteProperty, eOrtVon(dec)
			}},
		{"Wert eines Edm.String", func(i string) string { return `{"Id":1,"S":"` + i + `"}` },
			func(dec string) (model.EntityEingabe, model.EingabeUrsache, string) {
				return model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("S", model.ArtString, dec)}}, "", ""
			}},
		{"Wert einer Annotation", func(i string) string { return `{"Id":1,"@a":"` + i + `"}` }, eNurId},
		{"Name in einem Objekt einer Annotation", func(i string) string { return `{"Id":1,"@a":{"` + i + `":1}}` }, eNurId},
	}
	for _, st := range stellen {
		for _, z := range zeichen {
			for s, seite := range seiten {
				arten := []model.EingabeArt{model.EingabePatch}
				if st.name == "Wert eines Edm.String" && s == 1 {
					arten = eArten()
				}
				for _, art := range arten {
					was := fmt.Sprintf("%s %q Seite %d %s", st.name, z.json, s, art)
					got, err := eA("Ws", art, st.body(seite(z.json))).los()
					soll, u, ort := st.ok(seite(z.dec))
					switch {
					case z.ursache != "":
						eFehler(t, was, got, err, z.ursache, "")
					case u != "":
						eFehler(t, was, got, err, u, ort)
					default:
						eOK(t, was, got, err, soll)
					}
					n++
				}
			}
		}
	}
	if n != 1212+202 { // POST und PUT nur an der Seite „Mitte“ des Wertes
		t.Errorf("Menge %d, erwartet 1414", n)
	}
}

// TestEingabeBackslashAmEnde: ein Backslash vor dem schließenden Anführungszeichen schließt den String nicht (`syntax`).
func TestEingabeBackslashAmEnde(t *testing.T) {
	for _, art := range eArten() {
		got, err := eA("Ws", art, `{"Id":1,"S":"ab`+bs+`"}`).los()
		eFehler(t, "Backslash am Ende", got, err, model.EingabeSyntax, "")
	}
}

// eNurId ist das Ergebnis für einen Inhalt, den der Leser nicht deutet.
func eNurId(string) (model.EntityEingabe, model.EingabeUrsache, string) {
	return model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}}, "", ""
}
