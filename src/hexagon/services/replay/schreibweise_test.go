package replay_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Klasse dieser Datei ist die Schreibweise (SPEC-135, Kopf): Header-Namen und der Wert identity vergleicht der
// Dienst allein nach ASCII; die Methode zählt byte-gleich. Jede Menge ist ein Produkt aus Stelle, Name, Form der
// Schreibweise und Rand; die Verdikte entstehen aus dem Bau der Eingabe (ASCII-Form: gleicher Name, Faltung: anderer
// Name), nicht aus dem Code.

// faltungPartner sind die Zeichen jenseits von ASCII, die eine Unicode-Faltung auf den ASCII-Buchstaben c abbildet:
// das Kelvin-Zeichen auf k, das lange s auf s, das punktierte große I (ToLower) und das punktlose kleine i
// (ToUpper) auf i. Sie entstehen aus Codepunkten, nicht aus Literalen (AGENTS.md §3.18).
func faltungPartner(c byte) []string {
	switch c {
	case 'k', 'K':
		return []string{string(rune(0x212A))}
	case 's', 'S':
		return []string{string(rune(0x17F))}
	case 'i', 'I':
		return []string{string(rune(0x130)), string(rune(0x131))}
	}
	return nil
}

// faltungen liefert zu name jede Fassung, in der genau ein Buchstabe durch einen Partner ersetzt ist.
func faltungen(name string) []string {
	var r []string
	for i := 0; i < len(name); i++ {
		for _, p := range faltungPartner(name[i]) {
			r = append(r, name[:i]+p+name[i+1:])
		}
	}
	return r
}

// asciiFormen sind die vier ASCII-Schreibweisen eines ASCII-Namens: wie gegeben, klein, groß, abwechselnd.
func asciiFormen(n string) []string {
	var g strings.Builder
	for i := 0; i < len(n); i++ {
		c := n[i]
		if i%2 == 0 {
			c = strings.ToUpper(string(c))[0]
		} else {
			c = strings.ToLower(string(c))[0]
		}
		g.WriteByte(c)
	}
	return []string{n, strings.ToLower(n), strings.ToUpper(n), g.String()}
}

// TestSteuerheaderSchreibweise hält Z10 für die Schreibweise: Menge 26 (3 Namen; je 4 ASCII-Schreibweisen als
// Gutfall, 12; je Name jede Faltung als anderer Name und damit fehlender Header, 2 + 7 + 5 = 14).
func TestSteuerheaderSchreibweise(t *testing.T) {
	namen := []string{"X-Recorder-Workspace", "X-Recorder-Recording-Session", "X-Recorder-Replay-Session"}
	wert := map[string]string{namen[0]: wsID, namen[1]: sesID, namen[2]: repID}
	gut, schlecht := 0, 0
	for _, n := range namen {
		for _, f := range asciiFormen(n) {
			gut++
			u := neu(t)
			erfolg(u)
			a := anf()
			delete(a.Header, n)
			a.Header[f] = []string{wert[n]}
			resp, err := u.replay(context.Background(), a)
			if err != nil {
				t.Errorf("%q: %v", f, err)
				continue
			}
			_ = resp.Body.Close()
		}
		for _, f := range faltungen(n) {
			schlecht++
			u := neu(t)
			erfolg(u)
			a := anf()
			delete(a.Header, n)
			a.Header[f] = []string{wert[n]}
			_, err := u.replay(context.Background(), a)
			erwarteCode(t, "Faltung "+f, err, model.CodeE5003)
			if u.p.gesamt() != 0 {
				t.Errorf("%q: Port gerufen: %v", f, u.p.ops())
			}
		}
	}
	if gut != 12 || schlecht != 14 {
		t.Errorf("Menge: %d Gutfälle und %d Faltungen, erwartet 12 und 14", gut, schlecht)
	}
}

// randZeichen sind die Ränder, die ein Wert von Content-Encoding tragen kann: jedes einzelne Byte 0x00 bis 0xFF
// (256), die beiden Zeichen U+0085 und U+00A0 in UTF-8 und jedes übrige Zeichen, das unicode.IsSpace nennt (17).
func randZeichen(t *testing.T) []string {
	t.Helper()
	var r []string
	for b := 0; b < 256; b++ {
		r = append(r, string([]byte{byte(b)}))
	}
	r = append(r, string(rune(0x85)), string(rune(0xA0)))
	for c := rune(0x100); c <= unicode.MaxRune; c++ {
		if unicode.IsSpace(c) {
			r = append(r, string(c))
		}
	}
	if len(r) != 275 {
		t.Fatalf("%d Ränder, erwartet 275", len(r))
	}
	return r
}

// encodingLauf sendet einen POST mit JSON-Body und dem Header name: wert und liefert den Fehler und die gelesenen
// Bytes des Bodys; der Aufzeichnung steht unter dem Key des Requests.
func encodingLauf(t *testing.T, name string, wert ...string) (int, error) {
	t.Helper()
	u := neu(t)
	u.k.Pfade = pfadPassword(t)
	body := []byte(`{"a":1}`)
	a := mitBodyAnfrage(body, "application/json")
	rd := a.Body.(*zaehlLeser)
	a.Header[name] = wert
	rb, rr := redigiert(t, body, u.k.Pfade)
	u.lege(u.austausch(1, u.keyMitBody(a, rb, rr), true))
	resp, err := u.replay(context.Background(), a)
	if err == nil {
		_ = resp.Body.Close()
	}
	return rd.gelesen, err
}

// TestContentEncodingSchreibweise hält Z18 für Name, Wert und Rand: Menge 1118 = 8 (4 ASCII-Schreibweisen des
// Namens × gzip, identity) + 2 (Faltungen des Namens: ein anderer Header, ohne Wirkung) + 8 (Wert identity in 4
// ASCII-Schreibweisen, gut, und in 4 Faltungen, schlecht) + 1100 (275 Ränder × 4 Stellen: Anfang des Werts, Ende,
// Anfang des zweiten Teils, Ende des ersten Teils); nur Leerzeichen und Tabulator am Rand sind gut.
func TestContentEncodingSchreibweise(t *testing.T) {
	n := 0
	pruefe := func(was string, gut bool, name string, wert ...string) {
		t.Helper()
		n++
		gelesen, err := encodingLauf(t, name, wert...)
		if gut {
			if err != nil {
				t.Errorf("%s: %v", was, err)
			}
			return
		}
		erwarteCode(t, was, err, model.CodeE5001)
		if gelesen != 0 {
			t.Errorf("%s: %d Bytes gelesen", was, gelesen)
		}
	}
	for _, f := range asciiFormen("Content-Encoding") {
		pruefe("Name "+f+" gzip", false, f, "gzip")
		pruefe("Name "+f+" identity", true, f, "identity")
	}
	if fs := faltungen("Content-Encoding"); len(fs) != 2 {
		t.Fatalf("%d Faltungen des Namens, erwartet 2", len(fs))
	} else {
		for _, f := range fs {
			pruefe("Name "+f+" gzip", true, f, "gzip")
		}
	}
	for _, f := range asciiFormen("identity") {
		pruefe("Wert "+f, true, "Content-Encoding", f)
	}
	if fs := faltungen("identity"); len(fs) != 4 {
		t.Fatalf("%d Faltungen des Wertes, erwartet 4", len(fs))
	} else {
		for _, f := range fs {
			pruefe("Wert "+f, false, "Content-Encoding", f)
		}
	}
	stellen := []func(r string) string{
		func(r string) string { return r + "identity" },
		func(r string) string { return "identity" + r },
		func(r string) string { return "identity," + r + "identity" },
		func(r string) string { return "identity" + r + ",identity" },
	}
	for _, r := range randZeichen(t) {
		for i, bau := range stellen {
			pruefe(fmt.Sprintf("Rand Stelle %d Zeichen %s", i, quoted(r)), r == " " || r == "\t", "Content-Encoding", bau(r))
		}
	}
	if n != 1118 {
		t.Errorf("Menge: %d Elemente, erwartet 1118", n)
	}
}

func quoted(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		b.WriteString("\\x")
		b.WriteString(string("0123456789abcdef"[s[i]>>4]))
		b.WriteString(string("0123456789abcdef"[s[i]&15]))
	}
	return b.String()
}

// TestContentTypeSchreibweise hält Z17 für den Namen von Content-Type: Menge 5 (4 ASCII-Schreibweisen, gut; zwei
// Schreibweisen desselben Namens sind zwei Werte und damit ein Verstoß). Der Name enthält keinen Buchstaben mit
// Faltungspartner.
func TestContentTypeSchreibweise(t *testing.T) {
	if fs := faltungen("Content-Type"); len(fs) != 0 {
		t.Fatalf("%d Faltungen von Content-Type, erwartet 0", len(fs))
	}
	n := 0
	lauf := func(h map[string][]string) error {
		u := neu(t)
		u.k.Pfade = pfadPassword(t)
		body := []byte(`{"a":1}`)
		a := mitBodyAnfrage(body)
		for k, v := range h {
			a.Header[k] = v
		}
		rb, rr := redigiert(t, body, u.k.Pfade)
		u.lege(u.austausch(1, u.keyMitBody(a, rb, rr), true))
		resp, err := u.replay(context.Background(), a)
		if err == nil {
			_ = resp.Body.Close()
		}
		return err
	}
	for _, f := range asciiFormen("Content-Type") {
		n++
		if err := lauf(map[string][]string{f: {"application/json"}}); err != nil {
			t.Errorf("Name %q: %v", f, err)
		}
	}
	n++
	err := lauf(map[string][]string{"Content-Type": {"application/json"}, "content-type": {"application/json"}})
	erwarteCode(t, "zwei Schreibweisen", err, model.CodeE5001)
	if n != 5 {
		t.Errorf("Menge: %d, erwartet 5", n)
	}
}

// TestMethodeSchreibweise hält die Zelle Methode × fremd: Menge 25 (7 Methoden × kanonisch, klein, erster
// Buchstabe groß = 21, dazu 4 Faltungen: POST ein s, OPTIONS ein i mit zwei Partnern und ein s). Die Methode geht
// byte-gleich in den Key; nur die kanonische Schreibweise trifft die Aufzeichnung, jede andere ist ODR-E5001.
func TestMethodeSchreibweise(t *testing.T) {
	n := 0
	for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		varianten := []string{m, strings.ToLower(m), m[:1] + strings.ToLower(m[1:])}
		varianten = append(varianten, faltungen(m)...)
		for _, v := range varianten {
			n++
			u := neu(t)
			a := anf()
			a.Methode = m
			key := u.keyFuer(a, svcID, "/Products")
			u.lege(u.austausch(1, key, true))
			a.Methode = v
			resp, err := u.replay(context.Background(), a)
			if v == m {
				if err != nil {
					t.Errorf("Methode %q: %v", v, err)
				} else {
					_ = resp.Body.Close()
				}
				continue
			}
			erwarteCode(t, "Methode "+quoted(v), err, model.CodeE5001)
		}
	}
	if n != 25 {
		t.Errorf("Menge: %d, erwartet 25", n)
	}
}

// antwortLauf liefert die Header der Auslieferung für einen Exchange mit dem Header name: "v" und den flüchtigen
// Namen der Konfig.
func antwortLauf(t *testing.T, name string, fl []string) map[string][]string {
	t.Helper()
	u := neu(t)
	u.k.Regeln.VolatileHeaders = fl
	e := u.austausch(1, u.keyA(), true)
	e.ResponseHeaders[name] = []string{"v"}
	u.lege(e)
	resp, err := u.replay(context.Background(), anf())
	if err != nil {
		t.Fatalf("Header %q: %v", name, err)
	}
	_ = resp.Body.Close()
	return resp.Header
}

// TestAntwortHeaderFaltung hält Z36 für die Schreibweise der Namen: Menge 32. Zehn Namen (die sieben flüchtigen,
// Content-Length, Content-Encoding, X-Recorder-Match-Redacted) in abwechselnder Schreibweise entfallen (10); jede
// Faltung eines dieser Namen ist ein anderer Name und bleibt (12); ein flüchtiger Name der Konfig in Faltung
// entfernt den ASCII-Namen nicht (10).
func TestAntwortHeaderFaltung(t *testing.T) {
	fest := []string{"Content-Length", "Content-Encoding", "X-Recorder-Match-Redacted"}
	alle := append(append([]string{}, fluechtig()...), fest...)
	n := 0
	for _, name := range alle {
		n++
		f := asciiFormen(name)[3]
		if _, da := antwortLauf(t, f, fluechtig())[f]; da {
			t.Errorf("Name %q bleibt", f)
		}
	}
	faltung := 0
	for _, name := range alle {
		for _, f := range faltungen(name) {
			n++
			faltung++
			if got := antwortLauf(t, f, fluechtig())[f]; len(got) != 1 || got[0] != "v" {
				t.Errorf("Faltung %s wurde entfernt", quoted(f))
			}
		}
	}
	konfig := 0
	for _, name := range fluechtig() {
		for _, f := range faltungen(name) {
			n++
			konfig++
			if got := antwortLauf(t, name, []string{f})[name]; len(got) != 1 || got[0] != "v" {
				t.Errorf("Konfig-Name %s entfernt %s", quoted(f), name)
			}
		}
	}
	if faltung != 12 || konfig != 10 || n != 32 {
		t.Errorf("Menge: %d Faltungen, %d der Konfig, %d insgesamt; erwartet 12, 10 und 32", faltung, konfig, n)
	}
}
