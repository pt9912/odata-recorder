package model_test

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// etagOrakel baut den ETag zu n unabhängig von EtagFuer: Ziffer für Ziffer aus dem Rest der Division (SPEC-280 Nr. 1).
func etagOrakel(n int64) string {
	if n == 0 {
		return string([]byte{'W', '/', '"', '0', '"'})
	}
	var rueckwaerts []byte
	for ; n > 0; n /= 10 {
		rueckwaerts = append(rueckwaerts, byte('0'+n%10))
	}
	b := []byte{'W', '/', '"'}
	for i := len(rueckwaerts) - 1; i >= 0; i-- {
		b = append(b, rueckwaerts[i])
	}
	return string(append(b, '"'))
}

// TestEtagFuer (E1): die Literale der Randwerte und die Schleife 0 bis 300 gegen den unabhängigen Aufbau; 9 + 301 Elemente.
func TestEtagFuer(t *testing.T) {
	literale := []struct {
		n    int64
		want string
	}{
		{0, `W/"0"`}, {1, `W/"1"`}, {9, `W/"9"`}, {10, `W/"10"`}, {99, `W/"99"`}, {100, `W/"100"`},
		{2147483647, `W/"2147483647"`}, {4294967296, `W/"4294967296"`}, {9223372036854775807, `W/"9223372036854775807"`},
	}
	for _, l := range literale {
		if got := model.EtagFuer(l.n); got != l.want {
			t.Errorf("EtagFuer(%d) = %q, erwartet %q", l.n, got, l.want)
		}
		if got := etagOrakel(l.n); got != l.want {
			t.Errorf("Orakel(%d) = %q, erwartet %q", l.n, got, l.want)
		}
	}
	if len(literale) != 9 {
		t.Errorf("%d Literale, erwartet 9", len(literale))
	}
	zaehler := 0
	for n := int64(0); n <= 300; n++ {
		zaehler++
		if got, want := model.EtagFuer(n), etagOrakel(n); got != want {
			t.Errorf("EtagFuer(%d) = %q, erwartet %q", n, got, want)
		}
		if got := model.EtagFuer(n); got[:3] != "W/\"" || got[len(got)-1] != '"' || got[3:len(got)-1] != strconv.FormatInt(n, 10) {
			t.Errorf("EtagFuer(%d) = %q hat nicht die Form W/\"<n>\"", n, got)
		}
	}
	if zaehler != 301 {
		t.Errorf("%d Elemente der Schleife, erwartet 301", zaehler)
	}
}

// etagKlasse ist die Menge der Versionen aus dem Wortlaut der Zusage (SPEC-280 Nr. 1: jede nichtnegative 63-Bit-Zahl): 0 bis 300, jede
// Zehnerpotenz 10^k mit ihren Nachbarn (k = 1 bis 18), 2^k mit ihren Nachbarn (k = 1 bis 62), 2^63-1 und 2000 Zahlen einer Folge mit festem Seed.
func etagKlasse() []int64 {
	menge := map[int64]bool{}
	for n := int64(0); n <= 300; n++ {
		menge[n] = true
	}
	p := int64(1)
	for k := 1; k <= 18; k++ {
		p *= 10
		menge[p-1], menge[p], menge[p+1] = true, true, true
	}
	for k := 1; k <= 62; k++ {
		q := int64(1) << k
		menge[q-1], menge[q], menge[q+1] = true, true, true
	}
	menge[9223372036854775807] = true
	r := rand.New(rand.NewPCG(20261108, 15))
	for i := 0; i < 2000; i++ {
		menge[r.Int64N(1<<63-1)] = true
		menge[r.Int64N(1<<(1+r.IntN(62)))] = true
	}
	alle := make([]int64, 0, len(menge))
	for n := range menge {
		alle = append(alle, n)
	}
	return alle
}

// TestEtagFuerKlasse (E1, AGENTS.md §3.17): jedes Element der Klasse gegen das Orakel der Ziffernrechnung; die Zählprüfung nennt die Menge.
// Die Menge zwischen den Stichproben ist unbeschränkt (2^63 Werte); Prüfgrenze und Folge stehen im Plan.
func TestEtagFuerKlasse(t *testing.T) {
	alle := etagKlasse()
	if len(alle) < 4000 {
		t.Fatalf("Menge hat %d Elemente, erwartet mindestens 4000", len(alle))
	}
	for _, n := range alle {
		if got, want := model.EtagFuer(n), etagOrakel(n); got != want {
			t.Errorf("EtagFuer(%d) = %q, erwartet %q", n, got, want)
		}
	}
	t.Logf("Menge: %d Versionen", len(alle))
}

// etagDaten sind die acht Entities von abDaten, jede mit einem eigenen Etag (aus ihrer Id, nicht gleich der Id).
func etagDaten() []model.Entity {
	es := abDaten()
	for i := range es {
		es[i].Etag = model.EtagFuer(int64(i)*7 + 3)
	}
	return es
}

// TestFuehreAusBehaeltEtag (E10): jede Auswahl × Ordnung × Filter × $top × $skip, darin jede Projektion: jedes Element des Ergebnisses
// trägt den Etag der Entity mit demselben Schlüssel aus der Eingabe.
func TestFuehreAusBehaeltEtag(t *testing.T) {
	typ := abTyp()
	ein := etagDaten()
	soll := map[string]string{}
	for _, e := range ein {
		soll[e.Schluessel] = e.Etag
	}
	selects := []*model.SelectOption{nil, {Alle: true}}
	for _, p := range typ.Properties {
		selects = append(selects, abSel(p.Name))
	}
	selects = append(selects, abSel("Name", "Price"), abSel("Id", "Name"), abSel("Id"))
	ordnungen := [][]model.Sortierung{nil, {abOrd("Name", false), abOrd("Price", true)}, {abOrd("Id", true)}, {abOrd("Price", false)}}
	filter := []*model.FilterOption{nil, abF(abIDGt("3"))}
	tops := []model.ZahlOption{{}, abZ(0), abZ(3), abZ(100)}
	skips := []model.ZahlOption{{}, abZ(0), abZ(2), abZ(6), abZ(100)}
	faelle, elemente := 0, 0
	for si, sel := range selects {
		for oi, ord := range ordnungen {
			for fi, f := range filter {
				for ti, top := range tops {
					for ki, skip := range skips {
						faelle++
						opt := model.Optionen{Select: sel, OrderBy: ord, Filter: f, Top: top, Skip: skip}
						name := fmt.Sprintf("select %d ordnung %d filter %d top %d skip %d", si, oi, fi, ti, ki)
						r := abOk(t, name, typ, etagDaten(), opt, nil)
						for _, e := range r.Entities {
							elemente++
							if e.Etag != soll[e.Schluessel] || e.Etag == "" {
								t.Errorf("%s: Entity %s trägt Etag %q, erwartet %q", name, e.Schluessel, e.Etag, soll[e.Schluessel])
							}
						}
					}
				}
			}
		}
	}
	if want := len(selects) * len(ordnungen) * len(filter) * len(tops) * len(skips); faelle != want || faelle != 14*4*2*4*5 {
		t.Errorf("%d Fälle, erwartet %d", faelle, want)
	}
	if elemente == 0 {
		t.Error("kein Element des Ergebnisses geprüft")
	}
}
