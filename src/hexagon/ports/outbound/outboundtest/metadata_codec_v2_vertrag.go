package outboundtest

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"io"
	"math"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

//go:embed testdata/v2_referenz.xml testdata/v2_vererbung.xml testdata/v2_composite.xml
var dokumenteV2 embed.FS

// Namensräume der Dokumente von V2 (SPEC-171 Nr. 2, 3).
const (
	v2EdmxNs = "http://schemas.microsoft.com/ado/2007/06/edmx"
	v2EdmNs  = "http://schemas.microsoft.com/ado/2008/09/edm"
	v2MetaNs = "http://schemas.microsoft.com/ado/2007/08/dataservices/metadata"
)

// MetadataCodecV2Vertrag prüft die Zusagen von SPEC-171 (Zeilen A1 bis I3 der Tabelle Zusage · Fänger des Plans von
// slice-v2-metadaten) gegen jeden Adapter von outbound.MetadataCodec für die Version 2.0.
func MetadataCodecV2Vertrag(t *testing.T, neu CodecNeu) {
	zeile, stop := speicherwaechter()
	defer stop()
	for _, z := range []struct {
		name string
		lauf func(*testing.T, CodecNeu)
	}{
		{"A1 Argumente", zeileV2A1}, {"A2 Kontext", zeileV2A2}, {"A3 Fehler von src", zeileV2A3},
		{"A4 Zerlegung in Reads", zeileV2A4}, {"A5 Strom gehört dem Aufrufer", zeileV2A5},
		{"A6 leere Reads", zeileV2A6},
		{"B1 leer und unvollständig", zeileV2B1}, {"B2 Kodierung", zeileV2B2}, {"B3 DOCTYPE", zeileV2B3},
		{"B4 Referenzen", zeileV2B4}, {"B5 Kommentare, PI, Nachlauf", zeileV2B5}, {"B6 Namensräume", zeileV2B6},
		{"C1 Wurzel", zeileV2C1}, {"C2 Version", zeileV2C2}, {"C3 DataServiceVersion", zeileV2C3},
		{"C4 Namensräume des Schemas", zeileV2C4}, {"C5 Wurzel steht fest", zeileV2C5},
		{"D1 ohne Schlüssel", zeileV2D1}, {"D2 nicht auflösbar", zeileV2D2}, {"D3 Zyklus und Kette", zeileV2D3},
		{"D4 Schlüsseltypen", zeileV2D4}, {"D5 Formen des Property-Typs", zeileV2D5}, {"D6 Pflichtangaben und Namen", zeileV2D6},
		{"D7 Eindeutigkeit", zeileV2D7}, {"D8 Ketten", zeileV2D8}, {"D9 Rangfolge", zeileV2D9},
		{"D10 jeder Verstoß ist ODR-E6004", zeileV2D10}, {"D11 alles andere ist gültig", zeileV2D11},
		{"N1 Navigation löst", zeileV2N1}, {"N2 Navigation im Modell", zeileV2N2}, {"N3 Multiplicity", zeileV2N3},
		{"N4 Bindings", zeileV2N4}, {"N5 gültig ohne Binding", zeileV2N5}, {"N6 AssociationSet löst", zeileV2N6},
		{"N7 Zahl der End", zeileV2N7}, {"N8 Grenze der Bindings", zeileV2N8},
		{"E1 Größe", zeileV2E1}, {"E2 Tiefe", zeileV2E2}, {"E3 Summe der Members", zeileV2E3},
		{"F1 Modell Referenzdienst", zeileV2F1}, {"F2 Vererbung im Modell", zeileV2F2}, {"F3 Composite-Schlüssel", zeileV2F3},
		{"F4 Aliase und Schemas", zeileV2F4}, {"F5 Typabbildung", zeileV2F5}, {"F6 Randformen des Modells", zeileV2F6},
		{"G1 Funde", zeileV2G1}, {"G2 ohne Wirkung", zeileV2G2}, {"G3 leerer Bericht", zeileV2G3},
		{"G4 Reihenfolge der Funde", zeileV2G4}, {"G5 nie als unterstützt", zeileV2G5}, {"G6 Funde V2-eigener Stellen", zeileV2G6},
		{"H1 Fehlerform", zeileV2H1},
		{"I1 Wiederholung", zeileV2I1}, {"I2 gleichzeitig, geordnet", zeileV2I2}, {"I3 gleichzeitig, unabhängig", zeileV2I3},
		{"Z Zählprüfung der Mengen", zeileV2Z},
	} {
		zeile(z.name)
		t.Run(z.name, func(t *testing.T) { z.lauf(t, neu) })
	}
}

func festesV2(t *testing.T, name string) []byte {
	t.Helper()
	b, err := dokumenteV2.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.TrimRight(b, "\n")
}

func edmx2(inhalt string) string {
	return `<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"><edmx:DataServices>` + inhalt + `</edmx:DataServices></edmx:Edmx>`
}

func schema2(ns, inhalt string) string {
	return `<Schema xmlns="` + v2EdmNs + `" Namespace="` + ns + `">` + inhalt + `</Schema>`
}

func schemaAlias2(ns, alias, inhalt string) string {
	return `<Schema xmlns="` + v2EdmNs + `" Namespace="` + ns + `" Alias="` + alias + `">` + inhalt + `</Schema>`
}

// eins2 ist ein Dokument mit einem Schema `Ns`.
func eins2(inhalt string) []byte { return []byte(edmx2(schema2("Ns", inhalt))) }

// nav2 ist eine Navigation-Property auf die Association `Ns.R` (Rollen F und G).
func nav2(name string) string {
	return `<NavigationProperty Name="` + name + `" Relationship="Ns.R" FromRole="F" ToRole="G"/>`
}

// assoz2 ist die Association `R` von `Ns.T` nach `Ns.T` (Rolle F mit Vielfachheit *, Rolle G mit 0..1).
const assoz2 = `<Association Name="R"><End Type="Ns.T" Role="F" Multiplicity="*"/><End Type="Ns.T" Role="G" Multiplicity="0..1"/></Association>`

// zeileV2A1: die Argumente und ihre Reihenfolge (SPEC-125 Nr. 1 a bis d): Schleife über jede nichtleere Teilmenge der
// drei Verstöße mal Kontext beendet oder nicht, die Grenzen 0, -1 und MinInt64.
func zeileV2A1(t *testing.T, neu CodecNeu) {
	faelle := a1FaelleV2()
	for _, f := range faelle {
		var src io.Reader
		rd := &teilReader{teile: [][]byte{[]byte("<")}}
		if !f.srcNil {
			src = rd
		}
		m, b, err := neu.aufruf(f.ctx(), src, f.grenze)
		var mf *model.MetadatenFehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &mf) {
			t.Errorf("%+v: Fehler %v, erwartet model.ErrUngueltig ohne MetadatenFehler", f, err)
		}
		if rd.reads != 0 {
			t.Errorf("%+v: %d Reads, erwartet 0", f, rd.reads)
		}
		nullwerte(t, m, b)
	}
	if len(faelle) < 14 {
		t.Errorf("%d Läufe, erwartet mindestens 14", len(faelle))
	}
	pruefeFall(t, neu, fall{name: "grenze 1 ist gültig", doc: []byte("<"), max: 1, ursache: model.UrsacheSyntax})
}

// a1FallV2 ist eine nichtleere Teilmenge der Verstöße von SPEC-125 Nr. 1 a bis c (Bits 1, 2, 4: ctx nil, src nil, grenze
// kleiner 1) mit einem beendeten Kontext oder nicht.
type a1FallV2 struct {
	maske   int
	srcNil  bool
	grenze  int64
	ctxNil  bool
	beendet bool
}

func (f a1FallV2) ctx() context.Context {
	switch {
	case f.ctxNil:
		return nil
	case f.beendet:
		return abgebrochen()
	}
	return context.Background()
}

func a1FaelleV2() []a1FallV2 {
	var out []a1FallV2
	for maske := 1; maske < 8; maske++ {
		grenzen := []int64{1000}
		if maske&4 != 0 {
			grenzen = []int64{0, -1, math.MinInt64}
		}
		for _, ende := range []bool{false, true} {
			for _, g := range grenzen {
				out = append(out, a1FallV2{maske: maske, srcNil: maske&2 != 0, grenze: g, ctxNil: maske&1 != 0, beendet: ende})
			}
		}
	}
	return out
}

// zeileV2A2: der Kontext wird vor jedem Read und einmal vor dem Auflösen geprüft, der Fehler kommt unverändert zurück:
// Schleife über beide Kontextfehler und jeden Read eines in sechs Teile zerlegten Dokuments.
func zeileV2A2(t *testing.T, neu CodecNeu) {
	r := festesV2(t, "v2_referenz.xml")
	stuecke := []int{300, 600, 900, 1200, 1500}
	for _, ce := range kontextFehler() {
		for bei := 1; bei <= len(stuecke)+1; bei++ {
			for _, mitEnde := range []bool{false, true} {
				rd, err := a2AbbruchBei(neu, ce, r, stuecke, bei, mitEnde)
				if err != ce || rd.reads != bei {
					t.Errorf("Abbruch im Read %d (%v, mit Ende %v): Fehler %v, %d Reads", bei, ce, mitEnde, err, rd.reads)
				}
				if _, ok := outbound.Abbilden(err, outbound.GegenstandImport); ok {
					t.Errorf("Abbruch im Read %d (%v): der Fehler des Kontexts bekommt einen Code", bei, ce)
				}
			}
		}
		zeileV2A2Raender(t, neu, ce, r)
	}
}

// a2AbbruchBei beendet den Kontext mit dem Fehler ce im Read bei des in Stücke zerlegten Dokuments.
func a2AbbruchBei(neu CodecNeu, ce error, doc []byte, stuecke []int, bei int, mitEnde bool) (*teilReader, error) {
	ctx := neuBeendbar(nil)
	rd := teile(doc, stuecke...)
	rd.mitEnde = mitEnde
	rd.nach = func(n int) {
		if n == bei {
			ctx.beende(ce)
		}
	}
	_, _, err := neu.aufruf(ctx, rd, math.MaxInt64)
	return rd, err
}

// zeileV2A2Raender: der Abbruch im letzten Read vor dem Auflösen, ein Verstoß in den gelesenen Bytes vor dem Abbruch und
// ein Kontext, der vor dem Aufruf beendet ist.
func zeileV2A2Raender(t *testing.T, neu CodecNeu, ce error, r []byte) {
	rd, err := a2AbbruchBei(neu, ce, eins2(typ("T", prop("P", "Ns.Nope"))), []int{50}, 2, true)
	if err != ce || rd.reads != 2 {
		t.Errorf("Abbruch im letzten Read, Auflösungsfehler dahinter (%v): %v, %d Reads", ce, err, rd.reads)
	}
	_, err = a2AbbruchBei(neu, ce, []byte("<a b>"), nil, 1, false)
	erwarteFehler(t, err, model.UrsacheSyntax, "")
	rd = teile(r)
	m, b, err := neu.aufruf(neuBeendbar(ce), rd, 1<<20)
	if err != ce || rd.reads != 0 {
		t.Errorf("beendet vor dem Aufruf (%v): %v, %d Reads", ce, err, rd.reads)
	}
	nullwerte(t, m, b)
}

// zeileV2A3: ein Fehler von src ist kein MetadatenFehler und schlägt jede Auflösung, die nach den Daten käme (Nr. 2):
// Schleife über Fehlerort mal gültiges und ungültiges Dokument.
func zeileV2A3(t *testing.T, neu CodecNeu) {
	r := festesV2(t, "v2_referenz.xml")
	nichtLoesbar := eins2(typ("T", prop("P", "Ns.Nope")))
	for _, f := range []struct {
		name    string
		rd      *teilReader
		mitEnde bool
		syntax  bool
	}{
		{"vor dem ersten Byte", &teilReader{}, false, false},
		{"nach dem ersten Byte", teile(r[:1]), false, false},
		{"mitten im Element", teile(r[:60]), false, false},
		{"nach dem vollständigen Dokument", teile(r), false, false},
		{"Daten und Fehler im selben Read", teile(r[:60]), true, false},
		{"nach dem vollständigen Dokument mit Auflösungsfehler", teile(nichtLoesbar), false, false},
		{"Daten und Fehler im selben Read, Auflösungsfehler", teile(nichtLoesbar), true, false},
		{"Verstoß in den Daten des Fehler-Reads", teile([]byte("<a b>")), true, true},
		{"Verstoß vor dem Fehler", teile([]byte("<a b>"), 2), false, true},
	} {
		f.rd.ende, f.rd.mitEnde = errQuelle, f.mitEnde
		m, b, err := neu.aufruf(context.Background(), f.rd, 1<<20)
		var mf *model.MetadatenFehler
		switch {
		case f.syntax:
			erwarteFehler(t, err, model.UrsacheSyntax, "")
		case !errors.Is(err, errQuelle) || errors.As(err, &mf):
			t.Errorf("%s: Fehler %v, erwartet errQuelle ohne MetadatenFehler", f.name, err)
		}
		nullwerte(t, m, b)
	}
	rd := &teilReader{teile: [][]byte{r}, mitEnde: true}
	if _, _, err := neu.aufruf(context.Background(), rd, 1<<20); err != nil {
		t.Errorf("Daten mit io.EOF im letzten Read: %v", err)
	}
}

// alleFaelleV2 sind die Dokumente der Suite mit erwartetem Ausgang.
func alleFaelleV2(t *testing.T) []fall {
	t.Helper()
	var out []fall
	for _, p := range [][]fall{
		faelleV2B1(t), faelleV2B2(t), faelleV2B3(t), faelleV2B4(), faelleV2B5(t), faelleV2B6(), faelleV2C1(), faelleV2C2(),
		faelleV2C3(), faelleV2C4(), faelleV2C5(), faelleV2D1(), faelleV2D2(), faelleV2D3(), faelleV2D4(), faelleV2D5(), faelleV2D6(),
		faelleV2D7(), faelleV2D8(), faelleV2D9(), faelleV2E1(t), faelleV2E2(), faelleV2E3(), faelleV2F(t), faelleV2G(t),
		faelleV2N(), faelleV2D11(), nurFaelle(faelleV2B5Attribute()),
	} {
		out = append(out, p...)
	}
	return out
}

// zeileV2A4: das Ergebnis hängt nicht von der Zerlegung in Reads ab (Nr. 2): Schleife über alle Fälle der Suite und die
// Zerlegungen von A4 des V4-Vertrags (ganz, Byte für Byte, jede Zweiteilung bis a4Voll Bytes, darüber eine Probe).
func zeileV2A4(t *testing.T, neu CodecNeu) {
	var dokumente, laeufe, erwartet atomic.Int64
	var klein, gross []fall
	for _, f := range alleFaelleV2(t) {
		if len(f.doc) > a4Probe {
			gross = append(gross, f)
		} else {
			klein = append(klein, f)
		}
	}
	arbeiter := min(runtime.GOMAXPROCS(0), 8)
	parallel(arbeiter, func(w int) []string {
		for i := w; i < len(klein); i += arbeiter {
			if len(klein[i].doc) <= a4Voll {
				dokumente.Add(1)
				erwartet.Add(int64(max(len(klein[i].doc)-1, 0) * 2))
			}
			laeufe.Add(a4Fall(t, neu, klein[i]))
		}
		return nil
	})
	for _, f := range gross {
		a4Fall(t, neu, f)
	}
	if (laeufe.Load() != erwartet.Load() && !unterDetektor) || dokumente.Load() < 300 {
		t.Errorf("%d Dokumente bis %d Bytes, %d Läufe der Zweiteilungen, erwartet %d und mindestens 300 Dokumente",
			dokumente.Load(), a4Voll, laeufe.Load(), erwartet.Load())
	}
	t.Logf("%d Dokumente bis %d Bytes, %d Zweiteilungen", dokumente.Load(), a4Voll, laeufe.Load())
}

// zeileV2Z ist die Zählprüfung der Mengen der Schleifen und Generatoren: jede Menge hat mindestens die Zahl ihrer
// Elemente, die der Wortlaut des Vertrags ergibt; die Zahlen stehen im Protokoll.
func zeileV2Z(t *testing.T, _ CodecNeu) {
	for _, m := range []struct {
		name   string
		n, min int
	}{
		{"A1 Teilmengen mal Kontext mal Grenze", len(a1FaelleV2()), 30}, {"B1 Elemente", len(faelleV2B1(t)), 12},
		{"B2 Kodierungen und Zeichen", len(faelleV2B2(t)), 100}, {"B3 DOCTYPE", len(faelleV2B3(t)), 11},
		{"B4 Stellen mal Referenzen", len(faelleV2B4()), 305}, {"C1 Name mal Namensraum mal Version", len(faelleV2C1()), 43},
		{"C2 Versionen", len(faelleV2C2()), 13}, {"C3 DataServiceVersion", len(faelleV2C3()), 17}, {"C4 Namensräume des Schemas", len(faelleV2C4()), 7},
		{"C5 Paare", len(faelleV2C5()), 11}, {"D1 Schlüssel mal Abstract mal Kette", len(faelleV2D1()), 26}, {"D2 Referenzart mal Form", len(faelleV2D2()), 50},
		{"D3 Ringe und Ketten", len(faelleV2D3()), 36}, {"D4 Schlüsseltypen", len(faelleV2D4()), 35}, {"D5 Formen des Property-Typs", len(faelleV2D5()), 62},
		{"D5 Namensräume und Aliase mit Edm-Anfang", len(faelleV2D5Edm()), 9}, {"D6 Pflichtangaben und Namen", len(faelleV2D6()), 140}, {"D7 Doppelungen", len(faelleV2D7()), 17}, {"D8 Ketten", len(faelleV2D8()), 14},
		{"D9 Paare der Durchläufe", len(faelleV2D9()), 60}, {"D11 Dokumente der Klasse", len(faelleV2D11()), 156}, {"E1 Grenzen", len(faelleV2E1(t)), 33},
		{"E2 Tiefen", len(faelleV2E2()), 11}, {"E3 Summen", len(faelleV2E3()), 7}, {"G1 Stellen mal Arten", len(g1FaelleV2()), 140},
		{"N1 Formen mal Stellen", len(faelleV2N1()), 36}, {"N3 benannte Werte", len(faelleV2N3Klein()), 54},
		{"N3 Zeichenketten bis Länge 4", len(n3Zeichenketten()), 1555}, {"N3 ungültige Werte", len(n3Werte()), 1566},
		{"N3 Dokumente der Schleife", len(faelleV2N3()), 6264}, {"N6 Verstöße mal End", len(faelleV2N6()), 28},
		{"N7 Zahl der End", len(faelleV2N7()), 11}, {"N8 Fälle der Grenze der Bindings", len(faelleV2N8()), 9},
		{"A6 Dokumente der leeren Reads", len(a6FaelleV2()), 48}, {"alle Dokumente der Suite", len(alleFaelleV2(t)), 1000},
	} {
		t.Logf("%s: %d", m.name, m.n)
		if m.n < m.min {
			t.Errorf("%s: %d Elemente, erwartet mindestens %d", m.name, m.n, m.min)
		}
	}
}

// zeileV2A5: der Leser schließt src nie, liest nach dem Ende nicht weiter und höchstens grenze+1 Bytes (Nr. 2, 5).
func zeileV2A5(t *testing.T, neu CodecNeu) {
	alle := append(alleFaelleV2(t), gut("Referenzdienst", festesV2(t, "v2_referenz.xml")))
	for _, f := range alle {
		if f.gross {
			continue
		}
		rd := teile(f.doc)
		_, _, _ = neu.aufruf(context.Background(), rd, f.grenze())
		if rd.closes != 0 || rd.nachEnde != 0 {
			t.Errorf("%s: Close %d, Reads nach dem Ende %d", f.name, rd.closes, rd.nachEnde)
		}
	}
	e := &endlos{}
	_, _, err := neu.aufruf(context.Background(), e, 1000)
	erwarteFehler(t, err, model.UrsacheZuGross, "")
	if e.geliefert != 1001 || e.closes != 0 {
		t.Errorf("geliefert %d Bytes, Close %d; erwartet 1001 und 0", e.geliefert, e.closes)
	}
	for _, f := range []fall{{doc: []byte("<a"), max: 1}, {doc: festesV2(t, "v2_referenz.xml"), max: 100}} {
		rd := teile(f.doc)
		_, _, _ = neu.aufruf(context.Background(), rd, f.max)
		if rd.geliefert > int(f.max)+1 {
			t.Errorf("grenze %d: %d Bytes geliefert, erwartet höchstens %d", f.max, rd.geliefert, f.max+1)
		}
	}
}
