package admin_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// wsFelder sind die vier Felder, die einen Wert tragen (Index der Route in wsBodyRouten): die Summe der Schlüssel der drei
// Routen mit Body (2 + 1 + 1).
func wsFelder() []string { return []string{"baseline_snapshot_id", "id", "name", "snapshot_id"} }

// wsMitFeld bildet den Body, in dem das Feld den Wert roh (wie im Body geschrieben) trägt, und liefert den Index der Route und
// das Orakel der Argumente der Operation mit dem dekodierten Wert dek.
func wsMitFeld(feld, roh, dek string) (br int, body []byte, args [2]string) {
	switch feld {
	case "baseline_snapshot_id":
		return 0, []byte(`{"baseline_snapshot_id":"` + roh + `"}`), [2]string{dek, ""}
	case "id":
		return 0, []byte(`{"baseline_snapshot_id":"base-1","id":"` + roh + `"}`), [2]string{"base-1", dek}
	case "name":
		return 1, []byte(`{"name":"` + roh + `"}`), [2]string{wsPfadID, dek}
	}
	return 2, []byte(`{"snapshot_id":"` + roh + `"}`), [2]string{wsPfadID, dek}
}

// wsBGueltig sind die 54 gültigen Formen von M-Body gültig (SPEC-298 Nr. 5 a bis f): 17 Formen, 9 mit anderem Content-Type,
// 12 der Achse Zeichen im Wert und 16 der Achse Randleerraum im Wert; das Orakel sind die Argumente der Operation als Literale.
func wsBGueltig() []wsBForm {
	json := map[string][]string{"Content-Type": {"application/json"}}
	ae, euro, smiley, _ := nichtASCII()
	leer := "\r\n\t "
	hinten := " \t\r\n"
	rohEsc, dekEsc := "a"+esc("0076")+rueck()+"n"+ae+euro+smiley, "av\n"+ae+euro+smiley
	formen := []wsBForm{
		{name: "beide Schlüssel", br: 0, body: []byte(`{"baseline_snapshot_id":"base-1","id":"ws-neu"}`), args: [2]string{"base-1", "ws-neu"}},
		{name: "nur baseline_snapshot_id", br: 0, body: []byte(`{"baseline_snapshot_id":"base-1"}`), args: [2]string{"base-1", ""}},
		{name: "id zuerst", br: 0, body: []byte(`{"id":"ws-neu","baseline_snapshot_id":"base-1"}`), args: [2]string{"base-1", "ws-neu"}},
		{name: "id leer", br: 0, body: []byte(`{"baseline_snapshot_id":"base-1","id":""}`), args: [2]string{"base-1", ""}},
		{
			name: "Leerraum", br: 0, args: [2]string{" base ", " a "},
			body: []byte(leer + `{ "baseline_snapshot_id" : " base " ,` + "\n" + ` "id" : " a " }` + hinten),
		},
		{
			name: "Escape und Nicht-ASCII", br: 0, args: [2]string{"av\nb", ae + euro + smiley},
			body: []byte(`{"baseline_snapshot_id":"a` + esc("0076") + rueck() + `nb","id":"` + ae + euro + smiley + `"}`),
		},
		{name: "1048576 Bytes", br: 0, body: auffuellen(`{"baseline_snapshot_id":"base-1"}`, bodyGrenze), args: [2]string{"base-1", ""}},
	}
	for i, k := range []string{"name", "snapshot_id"} {
		br := i + 1
		formen = append(formen,
			wsBForm{name: k, br: br, body: []byte(`{"` + k + `":"v"}`), args: [2]string{wsPfadID, "v"}},
			wsBForm{name: k + " leer", br: br, body: []byte(`{"` + k + `":""}`), args: [2]string{wsPfadID, ""}},
			wsBForm{name: k + " Leerraum", br: br, body: []byte(leer + `{ "` + k + `" : " v " }` + hinten), args: [2]string{wsPfadID, " v "}},
			wsBForm{name: k + " Escape und Nicht-ASCII", br: br, body: []byte(`{"` + k + `":"` + rohEsc + `"}`), args: [2]string{wsPfadID, dekEsc}},
			wsBForm{name: k + " 1048576 Bytes", br: br, body: auffuellen(`{"`+k+`":"v"}`, bodyGrenze), args: [2]string{wsPfadID, "v"}})
	}
	for i := range formen {
		formen[i].achse, formen[i].kopf, formen[i].gueltig = "gültig", json, true
	}
	for br := range 3 {
		erste := slices.IndexFunc(formen, func(f wsBForm) bool { return f.br == br })
		for _, ct := range [][]string{nil, {"text/plain"}, {"application/x-www-form-urlencoded"}} {
			f := formen[erste]
			f.achse, f.name, f.kopf = "Content-Type", "Content-Type "+strings.Join(ct, "")+" "+formen[erste].name, nil
			if ct != nil {
				f.kopf = map[string][]string{"Content-Type": ct}
			}
			formen = append(formen, f)
		}
	}
	bomZeichen, ersatz := string([]byte{0xEF, 0xBB, 0xBF}), string([]byte{0xEF, 0xBF, 0xBD})
	nbsp := string([]byte{0xC2, 0xA0})
	for _, feld := range wsFelder() {
		for _, z := range []struct{ name, roh, dek string }{
			{"U+FEFF", "s" + bomZeichen + "x", "s" + bomZeichen + "x"},
			{"hohes Surrogat", "s" + esc("d800") + "x", "s" + ersatz + "x"},
			{"tiefes Surrogat", "s" + esc("dc00") + "x", "s" + ersatz + "x"},
		} {
			br, body, args := wsMitFeld(feld, z.roh, z.dek)
			formen = append(formen, wsBForm{achse: "Zeichen im Wert", name: z.name + " in " + feld, br: br, body: body, args: args, gueltig: true})
		}
		for _, z := range []struct{ name, roh, dek string }{
			{"Leerzeichen", " x ", " x "}, {"Tabulator", esc("0009") + "x" + esc("0009"), "\tx\t"},
			{"NBSP", nbsp + "x" + nbsp, nbsp + "x" + nbsp}, {"nur Leerraum", "   ", "   "},
		} {
			br, body, args := wsMitFeld(feld, z.roh, z.dek)
			formen = append(formen, wsBForm{achse: "Randleerraum", name: z.name + " in " + feld, br: br, body: body, args: args, gueltig: true})
		}
	}
	return formen
}

// wsPruefeGueltig hält eine gültige Form: der Status der Route, genau ein Aufruf der Operation mit den Argumenten als
// Literalen.
func wsPruefeGueltig(t *testing.T, f wsBForm) {
	t.Helper()
	e := wsNeu(t)
	r := f.route()
	rec := e.rufe(f.request())
	name := f.achse + " " + f.name
	if rec.Code != r.status {
		t.Errorf("%s: Status %d, erwartet %d (%q)", name, rec.Code, r.status, rec.Body.String())
	}
	e.p.nurDiagonale(t, name, r.op)
	if got := e.p.liste(); len(got) != 1 || got[0].a != f.args[0] || got[0].b != f.args[1] {
		t.Errorf("%s: Aufrufe %+v, erwartet die Argumente %q", name, got, f.args)
	}
}

// wsPruefeUngueltig hält eine ungültige Form gegen SPEC-298 Nr. 5 g: 400 mit dem Fehlerobjekt von `ODR-E2102`, genau ein
// Zählen, kein Aufruf eines Dienstes (der Dienst würde `ODR-E7001` liefern), kein Audit, kein Wert aus dem Body in der
// Antwort; der Handler allein liefert model.ErrUngueltig ohne Wert im Text und keine Antwort.
func wsPruefeUngueltig(t *testing.T, f wsBForm) {
	t.Helper()
	name := f.achse + " " + f.name
	e := wsNeu(t)
	e.wf.err, e.sf.err = wsE7001(), wsE7001()
	rec := e.rufe(f.request())
	pruefeFehlerAntwort(t, name, rec, 400, objekt400, "")
	if codes, andere := zaehlungenVon(e.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 {
		t.Errorf("%s: CountError %q, andere Aufrufe %d", name, codes, andere)
	}
	if e.p.gesamt() != 0 || e.audit.Aufrufe() != 0 {
		t.Errorf("%s: %d Aufrufe der Dienste, %d von Audit, erwartet 0 und 0", name, e.p.gesamt(), e.audit.Aufrufe())
	}
	begrenzt := http.MaxBytesReader(nil, io.NopCloser(f.leser()), bodyGrenze)
	ant, err := e.routen[wsBodyRouten()[f.br].route].Handler(context.Background(), admin.Anfrage{
		Methode: "POST", Body: begrenzt, Pfadwerte: map[string]string{"id": wsPfadID},
	})
	if !errors.Is(err, model.ErrUngueltig) || strings.Contains(err.Error(), canary()) || ant.Status != 0 || len(ant.Body) != 0 {
		t.Errorf("%s: Handler liefert %q, %v, erwartet model.ErrUngueltig ohne Wert und keine Antwort", name, ant.Body, err)
	}
	if e.p.gesamt() != 0 {
		t.Errorf("%s: der Handler rief einen Dienst", name)
	}
}

// wsAchsen liefert die Zahl der Formen je Achse.
func wsAchsen(formen []wsBForm) map[string]int {
	out := map[string]int{}
	for _, f := range formen {
		out[f.achse]++
	}
	return out
}

// TestWS6Zahlen hält M-Body gegen die Quelle des Tests: 54 gültige Formen (17 + 9 + 12 + 16), 145 ungültige (70 + 75) und
// 1536 des Rands; die Felder sind die Schlüssel der drei Routen (2 + 1 + 1), die Paare sind 6 über 2 gleich 15 je Route, und
// die Namen und Bodies der Formen einer Achse sind verschieden.
func TestWS6Zahlen(t *testing.T) {
	wsFrist(t)
	g := wsBGueltig()
	gleichAuf(t, "gültige Formen", len(g), 54)
	gleichAuf(t, "Achsen der gültigen Formen", wsAchsen(g), map[string]int{"gültig": 17, "Content-Type": 9, "Zeichen im Wert": 12, "Randleerraum": 16})
	u := wsUngueltigeBodys("x")
	gleichAuf(t, "ungültige Formen", len(u), 145)
	gleichAuf(t, "Achsen der ungültigen Formen", wsAchsen(u), map[string]int{
		"Wertform": 21, "Struktur": 21, "Schlüssel": 19, "Zeichen": 9, "Pflicht": 4, "Werte": 20, "Lesen": 6, "Paar": 45,
	})
	gleichAuf(t, "Rand", len(wsRand()), 1536)
	gleichAuf(t, "Rand gültig", len(slices.DeleteFunc(wsRand(), func(f wsBForm) bool { return !f.gueltig })), 3*8)
	gleichAuf(t, "Felder gleich Schlüssel der Routen", len(wsFelder()), 2+1+1)
	gleichAuf(t, "Paare je Route", 6*5/2, 15)
	namen := map[string]bool{}
	for _, f := range u {
		namen[f.achse+"/"+strconv.Itoa(f.br)+"/"+f.name] = true
	}
	gleichAuf(t, "verschiedene Namen", len(namen), 145)
	for _, feld := range wsFelder() {
		_, body, _ := wsMitFeld(feld, strings.Repeat("a", wsHoechstwert(feld)), "")
		gleichAuf(t, "Body mit dem größten Wert in "+feld, len(body), bodyGrenze)
	}
	for _, bis := range []int{1048000, wsHoechstwert("name"), 100, 5000} {
		laengen := wsLaengen(1, bis)
		enthalten := func(n int) bool { return n < 1 || n > bis || slices.Contains(laengen, n) }
		gleichAuf(t, "aufsteigend und endet an bis", laengen[len(laengen)-1] == bis && slices.IsSorted(laengen) && laengen[0] >= 1, true)
		for _, basis := range []int{2, 10} {
			for p := 1; p <= bis+1; p *= basis {
				if !enthalten(p-1) || !enthalten(p) || !enthalten(p+1) {
					t.Errorf("wsLaengen(1, %d) lässt eine Länge um %d aus", bis, p)
				}
			}
		}
		for _, n := range laengenDerWerte() {
			if !enthalten(n) {
				t.Errorf("wsLaengen(1, %d) lässt %d aus", bis, n)
			}
		}
	}
}

// TestWS6GueltigerBody hält WS6 (SPEC-298 Nr. 5 a bis f): jede der 54 gültigen Formen kommt dekodiert, unverändert und
// ungekürzt an den Dienst.
func TestWS6GueltigerBody(t *testing.T) {
	wsFrist(t)
	for _, f := range wsBGueltig() {
		wsPruefeGueltig(t, f)
	}
}

// TestWS7UngueltigerBody hält WS7 (SPEC-298 Nr. 5 b, c, g): die 70 Formen der Achsen Wertform, Struktur, Schlüssel und Zeichen
// sind 400 `ODR-E2102` ohne Aufruf eines Dienstes.
func TestWS7UngueltigerBody(t *testing.T) {
	wsFrist(t)
	n := 0
	for _, f := range wsUngueltigeBodys(canary()) {
		if slices.Contains([]string{"Wertform", "Struktur", "Schlüssel", "Zeichen"}, f.achse) {
			wsPruefeUngueltig(t, f)
			n++
		}
	}
	gleichAuf(t, "Formen", n, 70)
}

// TestWS7Rand hält WS7 für die Achse Rand (SPEC-298 Nr. 5 b): jeder Bytewert 0 bis 255 vor und hinter dem Objekt; gültig sind
// genau die vier Zeichen des Leerraums (8 je Route), jedes andere Byte ist 400 ohne Aufruf eines Dienstes.
func TestWS7Rand(t *testing.T) {
	wsFrist(t)
	gueltig, ungueltig := 0, 0
	for _, f := range wsRand() {
		if f.gueltig {
			wsPruefeGueltig(t, f)
			gueltig++
			continue
		}
		e := wsNeu(t)
		if rec := e.rufe(f.request()); rec.Code != 400 || e.p.gesamt() != 0 {
			t.Errorf("%s: Status %d, %d Aufrufe, erwartet 400 und 0", f.name, rec.Code, e.p.gesamt())
		}
		ungueltig++
	}
	gleichAuf(t, "gültige und ungültige Fälle", []int{gueltig, ungueltig}, []int{3 * 4 * 2, 3 * 252 * 2})
}

// TestWS8UngueltigerBody hält WS8 (SPEC-298 Nr. 5 a, d, g): die 75 Formen der Achsen Pflicht, Werte, Lesen und Paare sind 400
// `ODR-E2102` ohne Aufruf eines Dienstes; jedes der 45 Paare zweier Verstoßarten ist dieselbe Antwort.
func TestWS8UngueltigerBody(t *testing.T) {
	wsFrist(t)
	n := 0
	for _, f := range wsUngueltigeBodys(canary()) {
		if slices.Contains([]string{"Pflicht", "Werte", "Lesen", "Paar"}, f.achse) {
			wsPruefeUngueltig(t, f)
			n++
		}
	}
	gleichAuf(t, "Formen", n, 75)
}

// wsWert bildet den Request mit dem Wert im Feld (roh geschrieben und dekodiert gleich) und liefert die Route und das
// Orakel der Argumente.
func wsWert(feld, wert string) (wsBForm, [2]string) {
	br, body, args := wsMitFeld(feld, wert, wert)
	return wsBForm{br: br, body: body}, args
}

// TestWS9Laenge hält WS9 (SPEC-298 Nr. 5 d) für die Menge Länge × Feld × Zeichenklasse: die Klassen klein und Mehrbyte mit
// jeder Länge von wsLaengen(1, wsHoechstwert(Feld)) (jede Zweier- und Zehnerpotenz um 1 bis zum größten Wert, den der Body
// von 1 MiB im Feld fasst), die fünf anderen Klassen mit den 12 Längen von laengenDerWerte, je × 4 Felder, je der Status der
// Route und der Wert byte-gleich am Dienst. Prüfgrenze (akzeptiertes Negativ (1)): die Längen dazwischen und die Mischungen
// der Klassen in einem Wert.
func TestWS9Laenge(t *testing.T) {
	wsFrist(t)
	faelle, soll := 0, 0
	for _, feld := range wsFelder() {
		for _, klasse := range klassenDerWerte() {
			laengen := laengenDerWerte()
			if klasse == "klein" || klasse == "mehrbyte" {
				laengen = wsLaengen(1, wsHoechstwert(feld))
			}
			soll += len(laengen)
			for _, n := range laengen {
				wert := wertDerLaenge(klasse, n)
				if len(wert) != n {
					t.Fatalf("Orakel: Klasse %s Länge %d liefert %d Bytes", klasse, n, len(wert))
				}
				f, args := wsWert(feld, wert)
				f.name = feld + " " + klasse + " " + strconv.Itoa(n)
				f.args = args
				wsPruefeGueltig(t, f)
				faelle++
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, soll)
}

// TestWS9Zeichen hält WS9 für die Menge Zeichen × Feld: 139 Zeichen und Folgen roh in einer Zeichenkette × 4 Felder = 556:
// 32 Steuerzeichen sind ungültig, 94 ASCII-Zeichen und 7 gültige Folgen gehen unverändert an den Dienst, 6 ungültige Folgen
// sind 400.
func TestWS9Zeichen(t *testing.T) {
	wsFrist(t)
	gueltigeFolgen := [][]byte{{0xC2, 0x80}, {0xDF, 0xBF}, {0xE0, 0xA0, 0x80}, {0xEF, 0xBF, 0xBF}, {0xF0, 0x90, 0x80, 0x80}, {0xF4, 0x8F, 0xBF, 0xBF}, {0xEF, 0xBB, 0xBF}}
	ungueltigeFolgen := [][]byte{{0x80}, {0xC0, 0x80}, {0xC3}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0xFF}}
	faelle := 0
	for _, feld := range wsFelder() {
		versuche := func(wert string, gueltig bool) {
			f, args := wsWert(feld, "s"+wert+"x")
			f.args, f.achse, f.name = args, "Zeichen", fmt.Sprintf("%s %x", feld, wert)
			faelle++
			if gueltig {
				wsPruefeGueltig(t, f)
				return
			}
			wsPruefeUngueltig(t, f)
		}
		for c := range byte(0x80) {
			if c != '"' && c != 0x5c {
				versuche(string([]byte{c}), c >= 0x20)
			}
		}
		for _, s := range gueltigeFolgen {
			versuche(string(s), true)
		}
		for _, s := range ungueltigeFolgen {
			versuche(string(s), false)
		}
	}
	gleichAuf(t, "Fälle", faelle, (32+94+7+6)*4)
}

// TestWS9Stelle hält WS9 für die Stelle des Bytes 0xFF im Wert: jede Stelle von wsLaengen(0, wsHoechstwert(Feld) - 3) (der
// Wert mit den zwei Bytes dahinter füllt den Body von 1 MiB) × 4 Felder, jeder Fall 400 ohne Aufruf.
func TestWS9Stelle(t *testing.T) {
	wsFrist(t)
	faelle, soll := 0, 0
	for _, feld := range wsFelder() {
		stellen := wsLaengen(0, wsHoechstwert(feld)-3)
		soll += len(stellen)
		for _, p := range stellen {
			f, _ := wsWert(feld, strings.Repeat("a", p)+string([]byte{0xFF})+"bb")
			f.achse, f.name = "Stelle", feld+" "+strconv.Itoa(p)
			wsPruefeUngueltig(t, f)
			faelle++
		}
	}
	gleichAuf(t, "Fälle", faelle, soll)
}

// wsEscape ist ein Element von M-Escape: der Text des Escapes in einer Zeichenkette des Bodys und der dekodierte Wert als Bytes.
type wsEscape struct{ name, roh, dek string }

// wsEscapes sind die Elemente von M-Escape (SPEC-298 Nr. 5 d): jedes der 128 Zeichen U+0000 bis U+007F als Escape mit vier
// Hexziffern, die 48 mit einem Buchstaben in der Hexfolge zusätzlich in Großschreibung, die 8 Kurz-Escapes, 19 Zeichen jenseits
// von ASCII (bei Zeichen jenseits der Grundebene als Paar von Surrogaten; an jeder Grenze der Länge in UTF-8 und je eines aus
// den Klassen C1, Leerzeichen, Zeilentrenner, Privatgebrauch, Nichtzeichen und Ersatzzeichen), die 14 mit einem Buchstaben
// zusätzlich in Großschreibung (14 + 18 = 19 + 13 Formen). Der dekodierte Wert entsteht aus Bytewerten, nie aus einer Dekodierung; kein
// Escape steht im Quelltext. Prüfgrenze (akzeptiertes Negativ (1)): die übrigen Zeichen jenseits von ASCII.
func wsEscapes() []wsEscape {
	var out []wsEscape
	gesehen := map[string]bool{}
	add := func(name, roh, dek string) {
		if !gesehen[roh] {
			gesehen[roh] = true
			out = append(out, wsEscape{name, roh, dek})
		}
	}
	for c := range 128 {
		h := fmt.Sprintf("%04x", c)
		add(fmt.Sprintf("U+%04X klein", c), esc(h), string([]byte{byte(c)}))
		add(fmt.Sprintf("U+%04X groß", c), esc(strings.ToUpper(h)), string([]byte{byte(c)}))
	}
	for _, k := range []struct{ name, roh, dek string }{
		{"Anführungszeichen", rueck() + `"`, `"`}, {"Backslash", rueck() + rueck(), rueck()}, {"Schrägstrich", rueck() + "/", "/"},
		{"b", rueck() + "b", "\b"}, {"f", rueck() + "f", "\f"}, {"n", rueck() + "n", "\n"}, {"r", rueck() + "r", "\r"}, {"t", rueck() + "t", "\t"},
	} {
		add("Kurz-Escape "+k.name, k.roh, k.dek)
	}
	for _, z := range []struct {
		name string
		hex  []string
		dek  []byte
	}{
		{"U+0080", []string{"0080"}, []byte{0xC2, 0x80}}, {"U+07FF", []string{"07ff"}, []byte{0xDF, 0xBF}},
		{"U+0800", []string{"0800"}, []byte{0xE0, 0xA0, 0x80}}, {"U+FEFF", []string{"feff"}, []byte{0xEF, 0xBB, 0xBF}},
		{"U+FFFF", []string{"ffff"}, []byte{0xEF, 0xBF, 0xBF}}, {"U+10000", []string{"d800", "dc00"}, []byte{0xF0, 0x90, 0x80, 0x80}},
		{"U+10FFFF", []string{"dbff", "dfff"}, []byte{0xF4, 0x8F, 0xBF, 0xBF}}, {"U+1F600", []string{"d83d", "de00"}, []byte{0xF0, 0x9F, 0x98, 0x80}},
		{"U+0085", []string{"0085"}, []byte{0xC2, 0x85}}, {"U+00A0", []string{"00a0"}, []byte{0xC2, 0xA0}},
		{"U+2028", []string{"2028"}, []byte{0xE2, 0x80, 0xA8}}, {"U+2029", []string{"2029"}, []byte{0xE2, 0x80, 0xA9}},
		{"U+3000", []string{"3000"}, []byte{0xE3, 0x80, 0x80}}, {"U+D7FF", []string{"d7ff"}, []byte{0xED, 0x9F, 0xBF}},
		{"U+E000", []string{"e000"}, []byte{0xEE, 0x80, 0x80}}, {"U+FDD0", []string{"fdd0"}, []byte{0xEF, 0xB7, 0x90}},
		{"U+FFFD", []string{"fffd"}, []byte{0xEF, 0xBF, 0xBD}}, {"U+E0001", []string{"db40", "dc01"}, []byte{0xF3, 0xA0, 0x80, 0x81}},
		{"U+F0000", []string{"db80", "dc00"}, []byte{0xF3, 0xB0, 0x80, 0x80}},
	} {
		var klein, gross string
		for _, h := range z.hex {
			klein, gross = klein+esc(h), gross+esc(strings.ToUpper(h))
		}
		add(z.name+" klein", klein, string(z.dek))
		add(z.name+" groß", gross, string(z.dek))
	}
	return out
}

// TestWS9Escape hält WS9 (SPEC-298 Nr. 5 d, „die Route prüft kein Zeichen eines Wertes“) für die Menge Feld × Zeichen als
// Escape (M-Escape, 216 je Feld) × Stelle im Wert (der Wert ist das Zeichen allein, am Anfang, in der Mitte oder am Ende
// zwischen Buchstaben) × 4 Felder = 3456 Fälle: jedes Zeichen, das ein Escape in eine Zeichenkette bringt, kommt dekodiert und
// unverändert an den Dienst (Orakel: die Bytes als Literale); roh steht dieselbe Menge in TestWS9Zeichen.
func TestWS9Escape(t *testing.T) {
	wsFrist(t)
	gleichAuf(t, "Elemente von M-Escape", len(wsEscapes()), 128+6*8+8+14+18)
	faelle := 0
	for _, feld := range wsFelder() {
		for _, z := range wsEscapes() {
			for i, stelle := range []string{"allein", "Anfang", "Mitte", "Ende"} {
				vor, nach := [4]string{"", "", "s", "s"}[i], [4]string{"", "s", "x", ""}[i]
				br, body, args := wsMitFeld(feld, vor+z.roh+nach, vor+z.dek+nach)
				wsPruefeGueltig(t, wsBForm{achse: "Escape", name: feld + " " + z.name + " " + stelle, br: br, body: body, args: args, gueltig: true})
				faelle++
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, len(wsFelder())*4*(128+6*8+8+14+18))
}

// TestWS7Fremd hält WS7 (SPEC-298 Nr. 5 c) für die Menge der fremden Schlüssel: jedes Paar aus einer Route mit Body und einem
// Schlüssel einer anderen Route mit Body (M-Fremd) × jede Stelle des fremden Schlüssels (allein und an jeder der Stellen der
// erlaubten Schlüssel) × jeder der 7 Werte (Zeichenkette, leer, `null`, Zahl, Wahrheitswert, Array, Objekt) ist 400 `ODR-E2102`
// ohne Aufruf eines Dienstes. Die Zählprüfung bildet Paare, Stellen und Werte aus den Routen, ihren Schlüsseln und der Liste
// der Werte, ohne feste Zahl; die Routen mit Body sind die der Liste der Routen mit einem Body.
func TestWS7Fremd(t *testing.T) {
	wsFrist(t)
	mitBody := 0
	for _, r := range wsRouten() {
		if r.body != "" {
			mitBody++
		}
	}
	alle := map[string]bool{}
	for _, br := range wsBodyRouten() {
		for _, k := range br.schluessel {
			alle[k] = true
		}
	}
	gleichAuf(t, "Routen mit Body", len(wsBodyRouten()), mitBody)
	gleichAuf(t, "Felder gleich Schlüssel der Routen", wsFelder(), slices.Sorted(maps.Keys(alle)))
	paare, stellen := 0, 0
	for _, br := range wsBodyRouten() {
		fremde := len(alle) - len(br.schluessel)
		paare += fremde
		stellen += fremde * (1 + len(br.schluessel) + 1) * len(wsFremdeWerte())
	}
	formen := wsFremde()
	gleichAuf(t, "Formen", len(formen), stellen)
	for _, f := range formen {
		wsPruefeUngueltig(t, f)
	}
	gleichAuf(t, "Paare aus Route und fremdem Schlüssel", paare, 8)
}

// TestWS8Lesen hält WS8 für die Achse *Lesen* (SPEC-298 Nr. 5 a, M-Lesen): jede der 15 Arten von wsLesefehlerArten (das Ende,
// 8 Lesefehler, 6 Kontextfehler) × 4 Stellen, an denen der Fehler hinter den Bytes kommt (vor dem ersten Byte, vor dem letzten
// Byte des Objekts, hinter dem vollständigen Objekt, hinter einem Leerzeichen dahinter) × 4 Längen des Objekts (klein, 1 MiB - 1,
// 1 MiB, 1 MiB + 1) × 3 Routen mit Body = 720 Zellen. Orakel: wer mehr als 1 MiB Bytes liefert, ist 400 (die Grenze des Rahmens,
// welche Art auch folgt); sonst bricht ein Kontextfehler die Verbindung ab (kein Fehlerobjekt, kein Zählen, kein Aufruf eines
// Dienstes, der Handler liefert den Fehler unverändert), ein Lesefehler ist 400 (auch hinter einem vollständigen Objekt), und
// das Ende gilt: der Erfolg der Route mit einem Aufruf, wenn das Objekt vollständig ist, sonst 400.
func TestWS8Lesen(t *testing.T) {
	wsFrist(t)
	arten := map[string]int{}
	for _, lf := range wsLesefehlerArten() {
		arten[lf.art]++
	}
	gleichAuf(t, "Arten des Lesens", arten, map[string]int{"ende": 1, "lese": 8, "kontext": 6})
	zellen := 0
	for _, br := range wsBodyRouten() {
		r := wsRouten()[br.route]
		for stelle := range 4 {
			for _, n := range wsLesenLaengen(r) {
				daten := wsLesenDaten(r, stelle, n)
				for _, lf := range wsLesefehlerArten() {
					zellen++
					name := fmt.Sprintf("%s: %s, Stelle %d, Länge %d", r.name, lf.name, stelle, n)
					e := wsNeu(t)
					rec, p := wsRahmenAntwort(e, r.request(wsPfadID, "", nil, wsNeuerLeser(daten, lf.err)))
					codes, andere := zaehlungenVon(e.tele)
					handler := func() error {
						_, err := e.routen[br.route].Handler(context.Background(), admin.Anfrage{
							Methode: r.methode, Body: wsNeuerLeser(daten, lf.err), Pfadwerte: map[string]string{"id": wsPfadID},
						})
						return err
					}
					switch {
					case len(daten) <= bodyGrenze && lf.art == "kontext":
						if p != http.ErrAbortHandler || rec.Body.Len() != 0 || len(codes) != 0 || andere != 0 || e.p.gesamt() != 0 {
							t.Errorf("%s: Panik %v, Body %q, Zählungen %q, %d Aufrufe: erwartet den Abbruch ohne Antwort", name, p, rec.Body.String(), codes, e.p.gesamt())
						}
						if err := handler(); err != lf.err {
							t.Errorf("%s: Handler liefert %v, erwartet den Fehler unverändert", name, err)
						}
					case len(daten) <= bodyGrenze && lf.art == "ende" && stelle >= 2:
						if got := e.p.liste(); p != nil || rec.Code != r.status || len(got) != 1 || got[0].op != r.op || got[0].a != r.args[0] || got[0].b != r.args[1] {
							t.Errorf("%s: Panik %v, Status %d, Aufrufe %+v: erwartet %d und einen Aufruf mit %q", name, p, rec.Code, got, r.status, r.args)
						}
					default:
						pruefeFehlerAntwort(t, name, rec, 400, objekt400, "")
						if p != nil || !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 || e.p.gesamt() != 0 {
							t.Errorf("%s: Panik %v, Zählungen %q, %d Aufrufe: erwartet 400 ohne Aufruf", name, p, codes, e.p.gesamt())
						}
						if len(daten) <= bodyGrenze {
							if err := handler(); !errors.Is(err, model.ErrUngueltig) {
								t.Errorf("%s: Handler liefert %v, erwartet model.ErrUngueltig", name, err)
							}
						}
					}
					if n := e.audit.Aufrufe(); n != 0 {
						t.Errorf("%s: %d Aufrufe von Audit, erwartet 0", name, n)
					}
				}
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, len(wsBodyRouten())*len(wsLesefehlerArten())*4*4)
}

// TestWS9Dienst hält WS9 für die Werte, die der Dienst abweist: model.ErrUngueltig des Dienstes ist 400 `ODR-E2102` mit genau
// einem Aufruf (8 Werte), eine leere `snapshot_id` erreicht `Restore` als `""`, und `ODR-E7002` ist 404 (1 Fall).
func TestWS9Dienst(t *testing.T) {
	wsFrist(t)
	faelle := []struct{ feld, roh, dek string }{
		{"baseline_snapshot_id", "", ""}, {"id", strings.Repeat("a", 65), strings.Repeat("a", 65)}, {"id", "A", "A"}, {"id", "_", "_"},
		{"name", "", ""}, {"name", strings.Repeat("a", 129), strings.Repeat("a", 129)}, {"name", esc("0001"), "\x01"}, {"name", esc("007F"), "\x7f"},
	}
	for _, c := range faelle {
		br, body, args := wsMitFeld(c.feld, c.roh, c.dek)
		e := wsNeu(t)
		e.wf.err, e.sf.err = model.ErrUngueltig, model.ErrUngueltig
		f := wsBForm{br: br, body: body}
		rec := e.rufe(f.request())
		pruefeFehlerAntwort(t, c.feld, rec, 400, objekt400, "")
		got := e.p.liste()
		if len(got) != 1 || got[0].a != args[0] || got[0].b != args[1] {
			t.Errorf("%s %q: Aufrufe %+v, erwartet genau einen mit %q", c.feld, c.dek, got, args)
		}
	}
	e := wsNeu(t)
	e.wf.err = &inbound.Fehler{Code: model.CodeE7002, Status: 404}
	br, body, args := wsMitFeld("snapshot_id", "", "")
	rec := e.rufe(wsBForm{br: br, body: body}.request())
	pruefeFehlerAntwort(t, "leere snapshot_id", rec, 404, wsObjE7002, "")
	if got := e.p.liste(); len(got) != 1 || got[0].op != wsRestore || got[0].a != args[0] || got[0].b != "" {
		t.Errorf("leere snapshot_id: Aufrufe %+v, erwartet Restore mit %q", got, args)
	}
	gleichAuf(t, "Fälle", len(faelle)+1, 9)
}
