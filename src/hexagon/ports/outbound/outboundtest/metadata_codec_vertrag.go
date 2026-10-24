package outboundtest

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

//go:embed testdata/referenz.xml testdata/vererbung.xml testdata/composite.xml
var dokumente embed.FS

// CodecNeu liefert je Aufruf einen Leser; der Vertrag ruft nur die Schnittstelle des Ports.
type CodecNeu func() outbound.MetadataCodec

const (
	edmNs  = "http://docs.oasis-open.org/odata/ns/edm"
	edmxNs = "http://docs.oasis-open.org/odata/ns/edmx"
	edmxV3 = "http://schemas.microsoft.com/ado/2007/06/edmx"
)

// speicherLimit ist die Obergrenze des Heaps der Suite. Eine Schleife des Lesers, die ohne Ende wächst (etwa nach dem
// Vertauschen von Basistypen und Ketten beim Auflösen, dessen Zyklus sonst nie endet), endet damit nach Sekunden mit einem
// Fehler statt mit dem Speicher des Hosts; ein gesunder Lauf bleibt weit darunter. Die Grenze liegt bei einem Viertel von
// 2 GiB, denn eine wachsende Liste hält beim Umkopieren das Alte und das Neue.
const speicherLimit = 512 << 20

// speicherwaechter bricht den Prozess ab, sobald der Heap speicherLimit überschreitet, und nennt in der Meldung die Zeile,
// die zuletzt zeile gesetzt hat; stop beendet ihn. Die Zeilen laufen nacheinander, der Fall in der Zeile bleibt offen.
func speicherwaechter() (zeile func(name string), stop func()) {
	var laufend atomic.Value
	laufend.Store("")
	ende := make(chan struct{})
	go func() {
		tick := time.NewTicker(20 * time.Millisecond)
		defer tick.Stop()
		var ms runtime.MemStats
		for {
			select {
			case <-ende:
				return
			case <-tick.C:
			}
			if runtime.ReadMemStats(&ms); ms.HeapAlloc > speicherLimit {
				panic(fmt.Sprintf("Speicherwächter: Heap %d MiB über %d MiB in der Zeile %q, eine Schleife des Lesers wächst ohne Ende",
					ms.HeapAlloc>>20, speicherLimit>>20, laufend.Load()))
			}
		}
	}()
	return func(name string) { laufend.Store(name) }, func() { close(ende) }
}

// MetadataCodecVertrag prüft die Zusagen von SPEC-125 (Zeilen A1 bis I3 der Tabelle Zusage · Fänger)
// gegen jeden Adapter von outbound.MetadataCodec.
func MetadataCodecVertrag(t *testing.T, neu CodecNeu) {
	zeile, stop := speicherwaechter()
	defer stop()
	for _, z := range []struct {
		name string
		lauf func(*testing.T, CodecNeu)
	}{
		{"A1 Argumente", zeileA1}, {"A2 Kontext", zeileA2}, {"A3 Fehler von src", zeileA3},
		{"A4 Zerlegung in Reads", zeileA4}, {"A5 Strom gehört dem Aufrufer", zeileA5},
		{"B1 leer und unvollständig", zeileB1}, {"B2 Kodierung", zeileB2}, {"B3 DOCTYPE", zeileB3},
		{"B4 Referenzen", zeileB4}, {"B5 Kommentare, PI, Nachlauf", zeileB5}, {"B6 Namensräume", zeileB6},
		{"C1 Wurzel", zeileC1}, {"C2 Version", zeileC2},
		{"D1 ohne Schlüssel", zeileD1}, {"D2 nicht auflösbar", zeileD2}, {"D3 Zyklus und Kette", zeileD3},
		{"D4 Schlüsseltypen", zeileD4}, {"D5 Formen des Property-Typs", zeileD5}, {"D6 Namen und Eindeutigkeit", zeileD6},
		{"D7 Ketten", zeileD7}, {"D8 Rangfolge", zeileD8}, {"D9 jeder Verstoß ist ODR-E6004", zeileD9},
		{"D10 alles andere ist gültig", zeileD10},
		{"E1 Größe", zeileE1}, {"E2 Tiefe", zeileE2}, {"E3 Summe der Members", zeileE3},
		{"F1 Modell Referenzdienst", zeileF1}, {"F2 Vererbung im Modell", zeileF2}, {"F3 Composite-Schlüssel", zeileF3},
		{"F4 Aliase und Schemas", zeileF4}, {"F6 Randformen des Modells", zeileF6},
		{"G1 Funde", zeileG1}, {"G2 ohne Wirkung", zeileG2}, {"G3 leerer Bericht", zeileG3},
		{"G4 Reihenfolge der Funde", zeileG4}, {"G5 nie als unterstützt", zeileG5},
		{"H1 Fehlerform", zeileH1},
		{"I1 Wiederholung", zeileI1}, {"I2 gleichzeitig, geordnet", zeileI2}, {"I3 gleichzeitig, unabhängig", zeileI3},
	} {
		zeile(z.name)
		t.Run(z.name, func(t *testing.T) { z.lauf(t, neu) })
	}
}

// fall ist ein Dokument mit dem erwarteten Ausgang; ursache leer heißt gültig.
type fall struct {
	name    string
	doc     []byte
	max     int64
	ursache model.MetadatenUrsache
	ort     string
	gross   bool
}

func gut(name string, doc []byte) fall { return fall{name: name, doc: doc} }

func schlecht(name string, doc []byte, u model.MetadatenUrsache, ort string) fall {
	return fall{name: name, doc: doc, ursache: u, ort: ort}
}

func festes(t *testing.T, name string) []byte {
	t.Helper()
	b, err := dokumente.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.TrimRight(b, "\n")
}

func edmx(inhalt string) string {
	return `<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"><edmx:DataServices>` + inhalt + `</edmx:DataServices></edmx:Edmx>`
}

func schema(ns, inhalt string) string {
	return `<Schema xmlns="` + edmNs + `" Namespace="` + ns + `">` + inhalt + `</Schema>`
}

// eins ist ein Dokument mit einem Schema `Ns`.
func eins(inhalt string) []byte { return []byte(edmx(schema("Ns", inhalt))) }

func prop(name, typ string) string { return `<Property Name="` + name + `" Type="` + typ + `"/>` }

const schluesselID = `<Key><PropertyRef Name="Id"/></Key>` + `<Property Name="Id" Type="Edm.Int32"/>`

// typ ist ein Entity-Typ mit dem Schlüssel `Id`; typOhne hat keinen.
func typ(name, inhalt string) string {
	return `<EntityType Name="` + name + `">` + schluesselID + inhalt + `</EntityType>`
}

func typOhne(name, inhalt string) string {
	return `<EntityType Name="` + name + `">` + inhalt + `</EntityType>`
}

func ableitung(name, basis, inhalt string) string {
	return `<EntityType Name="` + name + `" BaseType="` + basis + `">` + inhalt + `</EntityType>`
}

// aufruf ruft einen frischen Leser; eine Panik meldet es als Fehler, damit der Test mit dem Namen der Zeile rot wird.
func (neu CodecNeu) aufruf(ctx context.Context, src io.Reader, grenze int64) (m model.ServiceModel, b model.ValidierungsBericht, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panik im Leser: %v", r)
		}
	}()
	return neu().ParseMetadata(ctx, src, grenze)
}

// grenze ist die Obergrenze des Aufrufs; 0 heißt: ein Gibibyte, weit über jedem Dokument der Fälle.
func (f fall) grenze() int64 {
	if f.max == 0 {
		return 1 << 30
	}
	return f.max
}

func lies(neu CodecNeu, f fall) (model.ServiceModel, model.ValidierungsBericht, error) {
	return neu.aufruf(context.Background(), bytes.NewReader(f.doc), f.grenze())
}

func nullwerte(t *testing.T, m model.ServiceModel, b model.ValidierungsBericht) {
	t.Helper()
	if m.Version() != "" || len(m.EntityTypes()) != 0 || len(m.EntitySets()) != 0 || b.Funde != nil {
		t.Errorf("Modell und Bericht sind keine Nullwerte: %q %d %v", m.Version(), len(m.EntityTypes()), b.Funde)
	}
}

// erwarteFehler prüft ODR-E6004 mit Ursache und Ort samt Form des Fehlers und Nullwerten.
func erwarteFehler(t *testing.T, err error, u model.MetadatenUrsache, ort string) {
	t.Helper()
	var mf *model.MetadatenFehler
	switch {
	case !errors.As(err, &mf):
		t.Errorf("Fehler %v, erwartet %s %q", err, u, ort)
	case mf.Ursache != u || mf.Ort != ort:
		t.Errorf("Ursache %q Ort %q, erwartet %q %q", mf.Ursache, mf.Ort, u, ort)
	case mf.Code() != model.CodeE6004 || !errors.Is(err, model.ErrMetadatenUngueltig):
		t.Errorf("Code %v, ErrMetadatenUngueltig %v", mf.Code(), errors.Is(err, model.ErrMetadatenUngueltig))
	}
}

// pruefeFall fährt einen Fall und prüft den erwarteten Ausgang; bei einem gültigen Dokument liefert
// es Modell und Bericht.
func pruefeFall(t *testing.T, neu CodecNeu, f fall) (model.ServiceModel, model.ValidierungsBericht) {
	t.Helper()
	m, b, err := lies(neu, f)
	if f.ursache == "" {
		if err != nil {
			t.Errorf("%s: Fehler %v bei gültigem Dokument", f.name, err)
		}
		return m, b
	}
	if err == nil {
		t.Errorf("%s: gültig, erwartet %s %q", f.name, f.ursache, f.ort)
		return m, b
	}
	t.Run(f.name, func(t *testing.T) {
		erwarteFehler(t, err, f.ursache, f.ort)
		nullwerte(t, m, b)
	})
	return m, b
}

func pruefeFaelle(t *testing.T, neu CodecNeu, faelle []fall) {
	t.Helper()
	for _, f := range faelle {
		pruefeFall(t, neu, f)
	}
}

// teilReader liefert Teile als einzelne Reads und zählt, was der Leser mit ihm tut.
type teilReader struct {
	teile     [][]byte
	reads     int
	nachEnde  int
	closes    int
	geliefert int
	nach      func(read int)
	ende      error
	mitEnde   bool
	zuEnde    bool
}

func (r *teilReader) endeFehler() error {
	if r.ende != nil {
		return r.ende
	}
	return io.EOF
}

func (r *teilReader) Close() error { r.closes++; return nil }

func (r *teilReader) Read(p []byte) (int, error) {
	r.reads++
	if len(r.teile) == 0 {
		if r.zuEnde {
			r.nachEnde++
		}
		r.zuEnde = true
		return 0, r.endeFehler()
	}
	n := copy(p, r.teile[0])
	r.teile[0] = r.teile[0][n:]
	if len(r.teile[0]) == 0 {
		r.teile = r.teile[1:]
	}
	r.geliefert += n
	if r.nach != nil {
		r.nach(r.reads)
	}
	if len(r.teile) == 0 && r.mitEnde {
		r.zuEnde = true
		return n, r.endeFehler()
	}
	return n, nil
}

func teile(doc []byte, stuecke ...int) *teilReader {
	r := &teilReader{}
	vor := 0
	for _, s := range append(stuecke, len(doc)) {
		r.teile = append(r.teile, doc[vor:s])
		vor = s
	}
	return r
}

// zeileA1: die Argumente und ihre Reihenfolge (SPEC-125 Nr. 1 a bis c).
func zeileA1(t *testing.T, neu CodecNeu) {
	var keinSrc io.Reader
	kein := func() context.Context { return nil }
	hin := context.Background
	for _, f := range []struct {
		name string
		ctx  func() context.Context
		src  bool
		max  int64
	}{
		{"ctx nil", kein, true, 1000}, {"grenze 0", hin, true, 0}, {"grenze -1", hin, true, -1},
		{"grenze 0 und beendeter ctx", abgebrochen, true, 0}, {"src nil", hin, false, 1000},
		{"ctx nil und src nil", kein, false, 1000}, {"src nil und grenze 0", hin, false, 0},
		{"grenze MinInt64", hin, true, math.MinInt64}, {"ctx nil und grenze 0", kein, true, 0}, {"ctx nil, src nil und grenze 0", kein, false, 0},
		{"src nil und beendeter ctx", abgebrochen, false, 1000}, {"src nil, grenze 0 und beendeter ctx", abgebrochen, false, 0},
	} {
		src := io.Reader(&teilReader{teile: [][]byte{[]byte("<")}})
		if !f.src {
			src = keinSrc
		}
		m, b, err := neu.aufruf(f.ctx(), src, f.max)
		var mf *model.MetadatenFehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &mf) {
			t.Errorf("%s: Fehler %v, erwartet model.ErrUngueltig ohne MetadatenFehler", f.name, err)
		}
		if tr, ist := src.(*teilReader); ist && tr.reads != 0 {
			t.Errorf("%s: %d Reads, erwartet 0", f.name, tr.reads)
		}
		nullwerte(t, m, b)
	}
	pruefeFall(t, neu, fall{name: "grenze 1 ist gültig", doc: []byte("<"), max: 1, ursache: model.UrsacheSyntax})
}

// beendbar ist ein Kontext, dessen Err() erst nach beende einen Fehler liefert; so endet er mit jedem Kontextfehler an
// einer vorgegebenen Stelle des Aufrufs. Der Leser ruft ihn aus der Goroutine des Aufrufs, ein Schutz ist nicht nötig.
type beendbar struct {
	done chan struct{}
	err  error
}

func neuBeendbar(err error) *beendbar {
	c := &beendbar{done: make(chan struct{})}
	if err != nil {
		c.beende(err)
	}
	return c
}

func (c *beendbar) beende(err error) { c.err = err; close(c.done) }

func (*beendbar) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *beendbar) Done() <-chan struct{}     { return c.done }
func (c *beendbar) Err() error                { return c.err }
func (*beendbar) Value(any) any               { return nil }

// kontextFehler sind die beiden Fehler eines beendeten Kontexts (Nr. 1 d).
func kontextFehler() []error { return []error{context.Canceled, context.DeadlineExceeded} }

// zeileA2: der Kontext wird vor jedem Read und einmal vor dem Auflösen geprüft (Nr. 1 d), und der Fehler des Kontexts
// kommt unverändert zurück: Schleife über die Kontextfehler und die Stellen, an denen der Kontext endet.
func zeileA2(t *testing.T, neu CodecNeu) {
	r := festes(t, "referenz.xml")
	nichtLoesbar := eins(typ("T", prop("P", "Ns.Nope")))
	syntaxFehler := []byte("<a b>")
	abbruchBei := func(ce error, doc []byte, stuecke []int, bei int, mitEnde bool) (*teilReader, error) {
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
	for _, ce := range kontextFehler() {
		for _, f := range []struct {
			name    string
			doc     []byte
			stuecke []int
			bei     int
			mitEnde bool
			reads   int
		}{
			{"Abbruch im ersten Read", r, []int{100, 900}, 1, false, 1},
			{"Abbruch im mittleren Read", r, []int{100, 900}, 2, false, 2},
			{"Abbruch im letzten Read, der io.EOF liefert", r, []int{100, 900}, 3, true, 3},
			{"Abbruch im letzten Read, Auflösungsfehler dahinter", nichtLoesbar, []int{50}, 2, true, 2},
		} {
			rd, err := abbruchBei(ce, f.doc, f.stuecke, f.bei, f.mitEnde)
			if err != ce {
				t.Errorf("%s (%v): Fehler %v, erwartet den Fehler des Kontexts", f.name, ce, err)
			}
			if rd.reads != f.reads {
				t.Errorf("%s (%v): %d Reads, erwartet %d", f.name, ce, rd.reads, f.reads)
			}
			if _, ok := outbound.Abbilden(err, outbound.GegenstandImport); ok {
				t.Errorf("%s (%v): der Fehler des Kontexts bekommt einen Code", f.name, ce)
			}
		}
		_, err := abbruchBei(ce, syntaxFehler, nil, 1, false)
		erwarteFehler(t, err, model.UrsacheSyntax, "")
		rd := teile(r)
		m, b, err := neu.aufruf(neuBeendbar(ce), rd, 1<<20)
		if err != ce || rd.reads != 0 {
			t.Errorf("beendet vor dem Aufruf (%v): %v, %d Reads", ce, err, rd.reads)
		}
		nullwerte(t, m, b)
	}
	ctx, abbruch := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer abbruch()
	if _, _, err := neu.aufruf(ctx, teile(r), 1<<20); err != context.DeadlineExceeded {
		t.Errorf("abgelaufene Frist: %v", err)
	}
	if _, _, err := neu.aufruf(abgebrochen(), teile(r), 1<<20); err != context.Canceled {
		t.Errorf("beendet über CancelFunc: %v", err)
	}
}

var errQuelle = errors.New("Quelle gestört")

// zeileA3: ein Fehler von src ist kein MetadatenFehler (Nr. 2).
func zeileA3(t *testing.T, neu CodecNeu) {
	r := festes(t, "referenz.xml")
	for _, f := range []struct {
		name    string
		rd      *teilReader
		mitEnde bool
	}{
		{"vor dem ersten Byte", &teilReader{ende: errQuelle}, false},
		{"nach dem ersten Byte", teile(r[:1]), false},
		{"mitten im Element", teile(r[:60]), false},
		{"nach dem vollständigen Dokument", teile(r), false},
		{"Daten und Fehler im selben Read", teile(r[:60]), true},
	} {
		f.rd.ende, f.rd.mitEnde = errQuelle, f.mitEnde
		m, b, err := neu.aufruf(context.Background(), f.rd, 1<<20)
		var mf *model.MetadatenFehler
		if !errors.Is(err, errQuelle) || errors.As(err, &mf) {
			t.Errorf("%s: Fehler %v, erwartet errQuelle ohne MetadatenFehler", f.name, err)
		}
		nullwerte(t, m, b)
	}
	rd := &teilReader{teile: [][]byte{r}, mitEnde: true}
	if _, _, err := neu.aufruf(context.Background(), rd, 1<<20); err != nil {
		t.Errorf("Daten mit io.EOF im letzten Read: %v", err)
	}
}

// nurFaelle liefert die Dokumente der Fälle mit erwarteten Funden.
func nurFaelle(gs []g1Fall) []fall {
	out := make([]fall, len(gs))
	for i, g := range gs {
		out[i] = g.fall
	}
	return out
}

func alleFaelle(t *testing.T) []fall {
	t.Helper()
	var out []fall
	for _, p := range [][]fall{
		faelleB1(t), faelleB2(t), faelleB3(t), faelleB4(), faelleB5(t), faelleB6(), faelleC1(), faelleC2(),
		faelleD1(), faelleD2(), faelleD3(), faelleD4(), faelleD5(), faelleD6(), faelleD7(), faelleD8(t),
		faelleE1(t), faelleE2(), faelleE3(), faelleF(t), faelleG(t), nurFaelle(faelleB5Attribute()),
	} {
		out = append(out, p...)
	}
	return out
}

func ergebnis(m model.ServiceModel, b model.ValidierungsBericht, err error) string {
	return fmt.Sprintf("%q|%q|%+v|%+v|%+v|%+v|%+v|%v", m.Version(), m.Namespaces(), m.EntityTypes(), m.EntitySets(), m.NavigationBindings(), b.Funde, b.Funde == nil, err)
}

// Grenzen der Menge von A4: bis a4Voll Bytes ist jede Zweiteilung ein Lauf, bis a4Probe Bytes sind es die ersten und die
// letzten a4Rand Positionen und a4Streu verteilte, darüber nur ganz und Byte für Byte.
const (
	a4Voll  = 5000
	a4Probe = 100000
	a4Rand  = 64
	a4Streu = 16
)

// zerlegtReader liefert doc in Reads: schnitt 0 ganz, schnitt -1 Byte für Byte, sonst in zwei Teilen, geteilt vor dem
// Byte an der Stelle schnitt; mit leerVor liefert er vor jedem Read einmal `0, nil`. Das Ende kommt als eigener Read.
type zerlegtReader struct {
	doc       []byte
	schnitt   int
	leerVor   bool
	pos       int
	leerGetan bool
}

func (r *zerlegtReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.doc) {
		return 0, io.EOF
	}
	if r.leerVor && !r.leerGetan {
		r.leerGetan = true
		return 0, nil
	}
	r.leerGetan = false
	ende := len(r.doc)
	switch {
	case r.schnitt == -1:
		ende = r.pos + 1
	case r.schnitt > r.pos:
		ende = r.schnitt
	}
	n := copy(p, r.doc[r.pos:ende])
	r.pos += n
	return n, nil
}

// a4RandDetektor ist die Zahl der Positionen am Anfang und am Ende, die A4 unter dem Detektor prüft: die Zweiteilung
// einer Zerlegung hat keinen Zustand, den der Detektor sehen könnte, und die Allaussage über alle Zweiteilungen fährt
// `make test`.
const a4RandDetektor = 8

// a4Zweiteilungen sind die Stellen, an denen A4 ein Dokument der Länge l in zwei Reads teilt: alle bis a4Voll Bytes,
// sonst eine Probe; unter dem Detektor immer die Probe.
func a4Zweiteilungen(l int) []int {
	var out []int
	rand := a4Rand
	if unterDetektor {
		rand = a4RandDetektor
	}
	switch {
	case l <= a4Voll && !unterDetektor:
		for s := 1; s < l; s++ {
			out = append(out, s)
		}
	case l <= a4Probe:
		stellen := map[int]bool{}
		for i := 1; i <= rand; i++ {
			stellen[i], stellen[l-i] = true, true
		}
		for i := 1; i <= a4Streu; i++ {
			stellen[l*i/(a4Streu+1)] = true
		}
		for s := 1; s < l; s++ {
			if stellen[s] {
				out = append(out, s)
			}
		}
	}
	return out
}

// zeileA4: das Ergebnis hängt nicht von der Zerlegung in Reads ab (Nr. 2): Schleife über alle Fälle der Suite
// (gültige und fehlerhafte, auch die großen) und die Zerlegungen ganz, Byte für Byte, jede Zweiteilung bis a4Voll
// Bytes und darüber bis a4Probe Bytes eine Probe (akzeptiertes Negativ im Plan); die Zweiteilungen mit und ohne leere
// Reads. Der Test zählt die Läufe der Dokumente bis a4Voll Bytes gegen Dokumente × (L−1) × 2; unter dem Detektor
// (unterDetektor) sind auch diese Dokumente nur eine Probe (a4RandDetektor), und die Zählung entfällt.
func zeileA4(t *testing.T, neu CodecNeu) {
	var dokumente, laeufe, erwartet atomic.Int64
	arbeit := make(chan fall)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), 8) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range arbeit {
				if len(f.doc) <= a4Voll {
					dokumente.Add(1)
					erwartet.Add(int64(max(len(f.doc)-1, 0) * 2))
				}
				laeufe.Add(a4Fall(t, neu, f))
			}
		}()
	}
	var gross []fall
	for _, f := range alleFaelle(t) {
		if len(f.doc) > a4Probe {
			gross = append(gross, f)
			continue
		}
		arbeit <- f
	}
	close(arbeit)
	wg.Wait()
	for _, f := range gross {
		a4Fall(t, neu, f)
	}
	if (laeufe.Load() != erwartet.Load() && !unterDetektor) || dokumente.Load() < 400 {
		t.Errorf("%d Dokumente bis %d Bytes, %d Läufe der Zweiteilungen, erwartet %d und mindestens 400 Dokumente",
			dokumente.Load(), a4Voll, laeufe.Load(), erwartet.Load())
	}
	t.Logf("%d Dokumente bis %d Bytes, %d Zweiteilungen", dokumente.Load(), a4Voll, laeufe.Load())
}

// a4Fall fährt ein Dokument in allen Zerlegungen von A4 gegen den Lauf ganz und liefert die Zahl der Läufe der
// Zweiteilungen, die es für ein Dokument bis a4Voll Bytes ausgeführt hat (sonst 0). Der erste Unterschied endet es.
func a4Fall(t *testing.T, neu CodecNeu, f fall) (zweiteilungen int64) {
	type ergebnisWert struct {
		m   model.ServiceModel
		b   model.ValidierungsBericht
		err error
	}
	lauf := func(schnitt int, leer bool) ergebnisWert {
		m, b, err := neu.aufruf(context.Background(), &zerlegtReader{doc: f.doc, schnitt: schnitt, leerVor: leer}, f.grenze())
		return ergebnisWert{m, b, err}
	}
	soll := lauf(0, false)
	stimmt := func(schnitt int, leer bool) bool {
		got := lauf(schnitt, leer)
		if reflect.DeepEqual(got, soll) {
			return true
		}
		t.Errorf("%s (%d Bytes) zerlegt bei %d (leere Reads %v): %s, ganz: %s", f.name, len(f.doc), schnitt, leer,
			ergebnis(got.m, got.b, got.err), ergebnis(soll.m, soll.b, soll.err))
		return false
	}
	if !stimmt(-1, false) || (len(f.doc) <= a4Probe && !stimmt(-1, true)) {
		return 0
	}
	for _, s := range a4Zweiteilungen(len(f.doc)) {
		for _, leer := range []bool{false, true} {
			if !stimmt(s, leer) {
				return 0
			}
			if len(f.doc) <= a4Voll {
				zweiteilungen++
			}
		}
	}
	return zweiteilungen
}

// endlos liefert Leerzeichen ohne Ende und zählt sie.
type endlos struct {
	geliefert int
	closes    int
}

func (e *endlos) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	e.geliefert += len(p)
	return len(p), nil
}

func (e *endlos) Close() error { e.closes++; return nil }

// zeileA5: der Leser schließt src nie, liest nach dem Ende nicht weiter und höchstens grenze+1 Bytes (Nr. 2, 5).
func zeileA5(t *testing.T, neu CodecNeu) {
	alle := append(alleFaelle(t), gut("Referenzdienst", festes(t, "referenz.xml")))
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
}
