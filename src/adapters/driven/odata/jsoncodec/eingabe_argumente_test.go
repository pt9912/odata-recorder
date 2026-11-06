package jsoncodec_test

import (
	"context"
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eVerstoss ist ein Programmfehler des Aufrufers (Zeile 1): eine Änderung am gültigen Aufruf.
type eVerstoss struct {
	name string
	mach func(a *eAufruf)
}

// eVerstoesse sind die elf Verstöße der Zeile 1 ohne `art`: ctx, src, grenze (3), entitySet (5), Nullwert-Modell.
func eVerstoesse() []eVerstoss {
	return []eVerstoss{
		{"ctx nil", func(a *eAufruf) { a.ctx = nil }},
		{"src nil", func(a *eAufruf) { a.src = nil }},
		{"grenze 0", func(a *eAufruf) { a.grenze = 0 }},
		{"grenze -1", func(a *eAufruf) { a.grenze = -1 }},
		{"grenze MinInt64", func(a *eAufruf) { a.grenze = math.MinInt64 }},
		{"set leer", func(a *eAufruf) { a.set = "" }},
		{"set unbekannt", func(a *eAufruf) { a.set = "Unbekannt" }},
		{"set andere Schreibweise", func(a *eAufruf) { a.set = "ws" }},
		{"set mit Namensraum", func(a *eAufruf) { a.set = "Ns.Ws" }},
		{"set ohne Typ", func(a *eAufruf) { a.set = "OhneTyp" }},
		{"Nullwert-Modell", func(a *eAufruf) { a.m = model.ServiceModel{} }},
	}
}

// eFalscheArten sind die 24 `art` der Zeile 1, die keine der drei Konstanten sind: je Konstante sieben Varianten (klein, Anfangsbuchstabe groß, Leerzeichen
// dahinter und davor, Tabulator dahinter, ein Zeichen zu viel, ein Zeichen zu wenig) und `DELETE`, `GET`, leer.
func eFalscheArten() []model.EingabeArt {
	var out []model.EingabeArt
	for _, k := range []string{"POST", "PATCH", "PUT"} {
		kl := strings.ToLower(k)
		for _, v := range []string{kl, strings.ToUpper(kl[:1]) + kl[1:], k + " ", " " + k, k + "\t", k + "X", k[:len(k)-1]} {
			out = append(out, model.EingabeArt(v))
		}
	}
	return append(out, "DELETE", "GET", "")
}

// TestEingabeProgrammfehler hält Zeile 1: Verstoß × Art, `src` wird nie gelesen; die Gegenfälle `grenze` 1 und MaxInt64 sind
// keine Programmfehler. Menge 63 = 11 × 3 + 24 + 3 × 2.
func TestEingabeProgrammfehler(t *testing.T) {
	n := 0
	for _, v := range eVerstoesse() {
		for _, art := range eArten() {
			s := eStromText(eObj("Id", "1"))
			a := eA("Ws", art, "")
			a.src = s
			v.mach(&a)
			got, err := a.los()
			eProg(t, v.name+" "+string(art), got, err)
			if s.readN != 0 {
				t.Errorf("%s %s: %d Reads", v.name, art, s.readN)
			}
			n++
		}
	}
	for _, art := range eFalscheArten() {
		s := eStromText(eObj("Id", "1"))
		a := eA("Ws", art, "")
		a.src = s
		got, err := a.los()
		eProg(t, "art "+string(art), got, err)
		if s.readN != 0 {
			t.Errorf("art %q: %d Reads", art, s.readN)
		}
		n++
	}
	for _, g := range []int64{1, math.MaxInt64} {
		for _, art := range eArten() {
			s := eStromText("5")
			a := eA("Ws", art, "")
			a.src, a.grenze = s, g
			got, err := a.los()
			eFehler(t, "Gegenfall grenze "+string(art), got, err, model.EingabeStruktur, "")
			if s.readN == 0 {
				t.Errorf("grenze %d %s: kein Read", g, art)
			}
			n++
		}
	}
	if n != 63 {
		t.Errorf("Menge %d, erwartet 63", n)
	}
}

// eBeendet liefert die beiden beendeten Kontexte der Zeilen 2 und 3 (abgebrochen, Frist abgelaufen) mit ihrem Fehler.
func eBeendet() (map[string]context.Context, func()) {
	abgebrochen, c1 := context.WithCancel(context.Background())
	c1()
	frist, c2 := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	return map[string]context.Context{"abgebrochen": abgebrochen, "Frist": frist}, c2
}

// TestEingabeArgumentVorKontext hält Zeile 2: ein Programmfehler geht dem beendeten Kontext vor. Menge 68 = (10 Verstöße ohne
// `ctx` + 24 `art`) × 2 Zustände.
func TestEingabeArgumentVorKontext(t *testing.T) {
	kontexte, ende := eBeendet()
	defer ende()
	n := 0
	for zustand, ctx := range kontexte {
		for _, v := range eVerstoesse()[1:] {
			a := eA("Ws", model.EingabePatch, eObj("Id", "1"))
			a.ctx = ctx
			v.mach(&a)
			got, err := a.los()
			eProg(t, v.name+" "+zustand, got, err)
			n++
		}
		for _, art := range eFalscheArten() {
			a := eA("Ws", art, eObj("Id", "1"))
			a.ctx = ctx
			got, err := a.los()
			eProg(t, "art "+string(art)+" "+zustand, got, err)
			n++
		}
	}
	if n != 68 {
		t.Errorf("Menge %d, erwartet 68", n)
	}
}

// TestEingabeKontextBeendet hält Zeile 3: bei gültigen Argumenten ist der Fehler `ctx.Err()` unverändert (`==`), ohne Code, ohne
// Read. Menge 18 = 3 Zustände × 3 Arten × {gültiger, falscher `Content-Type`} (der Kontext gewinnt), dazu der Gegenfall.
func TestEingabeKontextBeendet(t *testing.T) {
	kontexte, ende := eBeendet()
	defer ende()
	ursache := errors.New("ursache")
	mitUrsache, cancel := context.WithCancelCause(context.Background())
	cancel(ursache)
	kontexte["Ursache"] = mitUrsache
	n := 0
	for zustand, ctx := range kontexte {
		for _, art := range eArten() {
			for _, ct := range [][]string{eJSONCT(), {"text/plain"}} {
				s := eStromText(eObj("Id", "1"))
				a := eA("Ws", art, "")
				a.ctx, a.src, a.ct = ctx, s, ct
				got, err := a.los()
				if err != ctx.Err() {
					t.Errorf("%s %s: Fehler %v, erwartet ctx.Err() %v unverändert", zustand, art, err, ctx.Err())
				}
				var coded interface{ Code() model.Code }
				if errors.As(err, &coded) || errors.Is(err, ursache) || got.Properties != nil {
					t.Errorf("%s %s: Code, Ursache oder Ergebnis: %v %+v", zustand, art, err, got)
				}
				if s.readN != 0 {
					t.Errorf("%s %s: %d Reads", zustand, art, s.readN)
				}
				n++
			}
		}
	}
	if n != 3*3*2 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
	got, err := eA("Ws", model.EingabePost, eObj("Id", "1")).los()
	eOK(t, "aktiver Kontext", got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
}

// eLueckenAufruf ist der Aufruf der Zeile 4: ein Body in drei Reads, der letzte mit io.EOF; `ab` ist die Nummer des Aufrufs von
// `Err`, ab dem der Kontext endet. Die Folge der Aufrufe ist Err (Argumente), Err, Read, Err, Read, Err, Read, Err (nach dem
// letzten Read).
func eLueckenAufruf(art model.EingabeArt, teile [3]string, ab int, endet error) (eAufruf, *eStrom, *eKontext) {
	s := eNeuerStrom(eSchritt{daten: []byte(teile[0])}, eSchritt{daten: []byte(teile[1])}, eSchritt{daten: []byte(teile[2]), err: errEOF()})
	k := &eKontext{log: s.log, err: endet, ab: ab}
	a := eA("Mitten", art, "")
	a.ctx, a.src = k, s
	return a, s, k
}

// TestEingabeKontextLuecken hält Zeile 4: der Kontext endet vor dem 1., 2., 3. Read oder nach dem letzten; Fehler des Kontexts
// unverändert, kein weiterer Read. Menge 15 = 4 Stellen × 3 Arten + Paar `syntax` im 1. Read × 3.
func TestEingabeKontextLuecken(t *testing.T) {
	gut := [3]string{`{"Id":1,"Na`, `me":"x"`, `}`}
	endet := errors.New("kontext endet")
	n := 0
	for _, art := range eArten() {
		a, s, _ := eLueckenAufruf(art, gut, 0, nil)
		got, err := a.los()
		eOK(t, "ohne Abbruch", got, err, model.EntityEingabe{Typ: "Ns.Mitte", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("Name", model.ArtString, "x")}})
		soll := []string{"Err", "Err", "Read", "Err", "Read", "Err", "Read", "Err"}
		if !slices.Equal(*s.log, soll) {
			t.Errorf("Ereignisse %v, erwartet %v", *s.log, soll)
		}
		for ab := 2; ab <= 5; ab++ {
			a, s, _ := eLueckenAufruf(art, gut, ab, endet)
			got, err := a.los()
			if err != endet || got.Properties != nil {
				t.Errorf("%s ab %d: Fehler %v Ergebnis %+v, erwartet den Fehler des Kontexts unverändert", art, ab, err, got)
			}
			if reads := min(ab-2, 3); s.readN != reads {
				t.Errorf("%s ab %d: %d Reads, erwartet %d", art, ab, s.readN, reads)
			}
			s.verboten(t, "Lücke")
			n++
		}
		a, s, _ = eLueckenAufruf(art, [3]string{`{x`, `"}`, ``}, 3, endet)
		got, err = a.los()
		eFehler(t, "syntax vor Abbruch "+string(art), got, err, model.EingabeSyntax, "")
		if s.readN != 1 {
			t.Errorf("syntax: %d Reads", s.readN)
		}
		n++
	}
	if n != 15 {
		t.Errorf("Menge %d, erwartet 15", n)
	}
}
