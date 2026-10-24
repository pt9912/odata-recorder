package outboundtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// ODataCodecNeu liefert je Aufruf einen Leser; der Vertrag ruft nur die Schnittstelle des Ports.
type ODataCodecNeu func() outbound.ODataCodec

// jsuite hält die Modelle und den Leser einer Ausführung des Vertrags.
type jsuite struct {
	neu     ODataCodecNeu
	modelle map[string]model.ServiceModel
	faelle  []jfall
	rufe    atomic.Uint64 // Zahl der Aufrufe des Lesers, für drossel
}

// Der Leser legt je Aufruf einen Puffer von 32 KiB an. Die Zeilen mit vielen Aufrufen auf vielen Goroutinen (A4, J1, die
// Generatoren von B2 und B3) erzeugen damit Müll schneller, als eine Müllsammlung auf einem ausgelasteten Host nachkommt, und
// der Speicherwächter (512 MiB) schlägt vor ihr an. drossel wartet deshalb ab gHeapGrenze Heap eine Müllsammlung ab; geprüft
// wird jeder gHeapTakt-te Aufruf des Lesers.
const (
	gHeapGrenze = 128 << 20
	gHeapTakt   = 128
)

// gHeap ist der Stand des Heaps (lebende Objekte und noch nicht gesammelter Müll) in Bytes.
func gHeap() uint64 {
	s := []metrics.Sample{{Name: "/memory/classes/heap/objects:bytes"}}
	metrics.Read(s)
	return s[0].Value.Uint64()
}

// drossel zählt einen Aufruf des Lesers und wartet bei hohem Stand des Heaps eine Müllsammlung ab.
func (s *jsuite) drossel() {
	if s.rufe.Add(1)%gHeapTakt == 0 && gHeap() > gHeapGrenze {
		runtime.GC()
	}
}

// ODataCodecVertrag prüft die Zusagen von SPEC-126 (Zeilen A1 bis L1 der Tabelle Zusage · Fänger) gegen jeden
// Adapter von outbound.ODataCodec. meta liest die CSDL-Dokumente unter testdata/ in die Modelle der Fixtures.
func ODataCodecVertrag(t *testing.T, neu ODataCodecNeu, meta outbound.MetadataCodec) {
	zeile, stop := speicherwaechter()
	defer stop()
	defer debug.SetMaxStack(debug.SetMaxStack(256 << 20))
	s := neueJSuite(t, neu, meta)
	for _, z := range []struct {
		name string
		lauf func(*testing.T, *jsuite)
	}{
		{"A1 Argumente", jA1}, {"A2 Kontext", jA2}, {"A3 Fehler von src", jA3},
		{"A4 Zerlegung in Reads", jA4}, {"A5 Strom gehört dem Aufrufer", jA5},
		{"B1 leer und unvollständig", jB1}, {"B2 Grammatik", jB2}, {"B3 Kodierung", jB3},
		{"B4 Tiefe", jB4}, {"B5 Größe", jB5}, {"B6 doppelte Namen", jB6},
		{"C1 Wurzel", jC1}, {"C2 Aufbau der Collection", jC2}, {"C3 die Methode entscheidet", jC3},
		{"D1 Kontext gültig", jD1}, {"D2 Kontext ungültig", jD2},
		{"E1 nextLink", jE1}, {"E2 count", jE2}, {"E3 odata.type", jE3}, {"E4 Annotationen", jE4}, {"E5 Mitglieder", jE5},
		{"F1 String", jF1}, {"F2 Boolean", jF2}, {"F3 Int32", jF3}, {"F4 Int64", jF4}, {"F5 Decimal", jF5},
		{"F6 Guid", jF6}, {"F7 Date", jF7}, {"F8 DateTimeOffset", jF8}, {"Z1 Composite-Schlüssel mit Guid", jZ1},
		{"F9 null gegen nicht selektiert", jF9}, {"F10 JSON-Typ gegen EDM-Typ", jF10}, {"F11 Marker", jF11},
		{"G1 Schlüssel eines Typs", jG1}, {"G2 zusammengesetzter Schlüssel", jG2}, {"G3 Verletzungen des Schlüssels", jG3},
		{"H1 Rangfolge im Lesen", jH1}, {"H2 Rangfolge im Deuten", jH2}, {"H3 jeder Verstoß ist ODR-E6101", jH3},
		{"H4 alles andere ist gültig", jH4},
		{"I1 Wiederholung und Reihenfolge", jI1}, {"I2 Ergebnis gehört dem Aufrufer", jI2}, {"I3 Dubletten", jI3},
		{"J1 gleichzeitig, geordnet", jJ1}, {"J2 gleichzeitig, unabhängig", jJ2},
		{"K1 Fehlerform und Ort", jK1}, {"K2 Abbildung", jK2}, {"K3 kein Typ der Bibliothek", jK3}, {"K4 Port", jK4},
		{"L1 Speicher des Ergebnisses", jL1},
	} {
		zeile(z.name)
		t.Run(z.name, func(t *testing.T) { z.lauf(t, s) })
	}
}

func neueJSuite(t *testing.T, neu ODataCodecNeu, meta outbound.MetadataCodec) *jsuite {
	t.Helper()
	s := &jsuite{neu: neu, modelle: literalModelle()}
	lade := func(doc []byte) model.ServiceModel {
		m, _, err := meta.ParseMetadata(context.Background(), bytes.NewReader(doc), 1<<20)
		if err != nil {
			t.Fatalf("Modell der Fixture: %v", err)
		}
		return m
	}
	s.modelle["R"] = lade(festes(t, "referenz.xml"))
	s.modelle["Z"] = lade(festes(t, "vererbung.xml"))
	s.modelle["K"] = lade([]byte(compositeOhneZ))
	s.modelle["KZ"] = lade(festes(t, "composite.xml"))
	return s
}

// compositeOhneZ ist das Dokument von testdata/composite.xml mit Edm.Int64 statt Edm.Guid, Edm.Date und
// Edm.DateTimeOffset.
const compositeOhneZ = `<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"><edmx:DataServices>` +
	`<Schema xmlns="` + edmNs + `" Namespace="Shop" Alias="S">` +
	`<EntityType Name="OrderLine"><Key><PropertyRef Name="OrderID"/><PropertyRef Name="LineNo"/></Key>` +
	`<Property Name="LineNo" Type="Edm.Int32" Nullable="false"/><Property Name="OrderID" Type="Edm.Int64" Nullable="false"/>` +
	`<Property Name="Note" Type="Edm.String"/><Property Name="Flag" Type="Edm.Boolean"/>` +
	`<Property Name="Total" Type="Edm.Int64"/><Property Name="Amount" Type="Edm.Decimal"/></EntityType>` +
	`<EntityType Name="Gift" BaseType="S.OrderLine"/>` +
	`<EntityContainer Name="Container"><EntitySet Name="Lines" EntityType="S.OrderLine"/></EntityContainer>` +
	`</Schema></edmx:DataServices></edmx:Edmx>`

// esc ist der JSON-Escape der UTF-16-Einheit mit den vier Hex-Ziffern: ein Backslash, `u` und die Ziffern.
func esc(hex string) string { return "\x5cu" + hex }

// jfall ist ein Aufruf mit erwartetem Ausgang; ursache leer heißt gültig.
type jfall struct {
	name       string
	m          string
	set        string
	einzel     bool
	body       []byte
	grenze     int64
	ursache    model.DatenUrsache
	ort        string
	gross      bool
	props      []model.PropertyWert
	schluessel string
	typ        string
	ueber      []model.Uebersprungen
	nach       func(*testing.T, jergebnis)
}

// grenzeOderStandard ist die Obergrenze des Aufrufs; 0 heißt: ein Gibibyte, weit über jedem Body der Fälle.
func (f jfall) grenzeOderStandard() int64 {
	if f.grenze == 0 {
		return 1 << 30
	}
	return f.grenze
}

// jergebnis fasst die Rückgabe beider Methoden: bei ParseEntity ist Entities die eine Entity.
type jergebnis struct {
	seite  model.EntitySeite
	entity model.Entity
	err    error
}

// ruf ruft die Methode des Falls; eine Panik des Lesers wird zum Fehler, damit der Test mit dem Namen der Zeile rot wird.
func (s *jsuite) ruf(ctx context.Context, src io.Reader, f jfall) jergebnis {
	return s.aufruf(ctx, f.einzel, src, s.modelle[f.m], f.set, f.grenzeOderStandard())
}

// aufruf ruft einen frischen Leser über die Schnittstelle.
func (s *jsuite) aufruf(ctx context.Context, einzel bool, src io.Reader, m model.ServiceModel, set string, grenze int64) (r jergebnis) {
	defer func() {
		if p := recover(); p != nil {
			r.err = fmt.Errorf("Panik im Leser: %v", p)
		}
	}()
	s.drossel()
	c := s.neu()
	if einzel {
		r.entity, r.err = c.ParseEntity(ctx, src, m, set, grenze)
	} else {
		r.seite, r.err = c.ParseCollection(ctx, src, m, set, grenze)
	}
	return r
}

func (s *jsuite) lauf(f jfall) jergebnis {
	return s.ruf(context.Background(), bytes.NewReader(f.body), f)
}

// entities liefert die gelesenen Entities beider Methoden.
func (r jergebnis) entities(f jfall) []model.Entity {
	if f.einzel {
		return []model.Entity{r.entity}
	}
	return r.seite.Entities
}

// nullwerte prüft, dass beide Rückgaben die Nullwerte sind (Kopf von SPEC-126).
func (r jergebnis) nullwerte(t *testing.T, name string) {
	t.Helper()
	if !reflect.DeepEqual(r.seite, model.EntitySeite{}) || !reflect.DeepEqual(r.entity, model.Entity{}) {
		t.Errorf("%s: Rückgabe ist kein Nullwert: %+v %+v", name, r.seite, r.entity)
	}
}

// datenFehler prüft einen DatenFehler mit Ursache und Ort samt Code und errors.Is.
func datenFehler(t *testing.T, name string, err error, u model.DatenUrsache, ort string) {
	t.Helper()
	var df *model.DatenFehler
	switch {
	case !errors.As(err, &df):
		t.Errorf("%s: Fehler %v, erwartet %s %q", name, err, u, ort)
	case df.Ursache != u || df.Ort != ort:
		t.Errorf("%s: Ursache %q Ort %q, erwartet %q %q", name, df.Ursache, df.Ort, u, ort)
	case df.Code() != model.CodeE6101 || !errors.Is(err, model.ErrDatenUngueltig):
		t.Errorf("%s: Code %v, ErrDatenUngueltig %v", name, df.Code(), errors.Is(err, model.ErrDatenUngueltig))
	}
}

// pruefe fährt einen Fall und prüft den erwarteten Ausgang.
func (s *jsuite) pruefe(t *testing.T, f jfall) {
	t.Helper()
	r := s.lauf(f)
	if f.ursache != "" {
		datenFehler(t, f.name, r.err, f.ursache, f.ort)
		r.nullwerte(t, f.name)
		return
	}
	if r.err != nil {
		t.Errorf("%s: Fehler %v bei gültigem Body", f.name, r.err)
		return
	}
	f.pruefeErste(t, r.entities(f))
	if f.nach != nil {
		f.nach(t, r)
	}
}

// pruefeErste prüft, was der Fall von der ersten gelesenen Entity erwartet.
func (f jfall) pruefeErste(t *testing.T, es []model.Entity) {
	t.Helper()
	if len(es) == 0 {
		if f.props != nil || f.schluessel != "" || f.typ != "" || f.ueber != nil {
			t.Errorf("%s: keine Entity gelesen", f.name)
		}
		return
	}
	e := es[0]
	if f.props != nil && !reflect.DeepEqual(e.Properties, f.props) {
		t.Errorf("%s: Properties %+v, erwartet %+v", f.name, e.Properties, f.props)
	}
	if f.schluessel != "" && e.Schluessel != f.schluessel {
		t.Errorf("%s: Schlüssel %q, erwartet %q", f.name, e.Schluessel, f.schluessel)
	}
	if f.typ != "" && e.Typ != f.typ {
		t.Errorf("%s: Typ %q, erwartet %q", f.name, e.Typ, f.typ)
	}
	if f.ueber != nil && !reflect.DeepEqual(e.Uebersprungen, f.ueber) {
		t.Errorf("%s: Uebersprungen %+v, erwartet %+v", f.name, e.Uebersprungen, f.ueber)
	}
}

func (s *jsuite) pruefeAlle(t *testing.T, faelle []jfall) {
	t.Helper()
	for _, f := range faelle {
		s.pruefe(t, f)
	}
}

// kontextAn ist der Wert von `@odata.context` für das Set: die Form der Collection oder der Einzel-Entity.
func kontextAn(set string, einzel bool) string {
	if einzel {
		return `"@odata.context":"$metadata#` + set + `/$entity"`
	}
	return `"@odata.context":"$metadata#` + set + `"`
}

// jKol ist der Body einer Collection mit den Elementen.
func jKol(set string, elems ...string) []byte {
	return []byte(`{` + kontextAn(set, false) + `,"value":[` + strings.Join(elems, ",") + `]}`)
}

// jEin ist der Body einer Einzel-Entity mit den Mitgliedern.
func jEin(set, members string) []byte {
	if members == "" {
		return []byte(`{` + kontextAn(set, true) + `}`)
	}
	return []byte(`{` + kontextAn(set, true) + `,` + members + `}`)
}

// beide bildet aus den Mitgliedern einer Entity einen Fall je Methode: als einziges Element von `value` und als
// Wurzel der Einzel-Entity. ort ist der Pointer in der Entity; bei der Collection steht /value/0 davor.
func beide(name, m, set, members string, u model.DatenUrsache, ort string) []jfall {
	return []jfall{
		{name: name + " (Collection)", m: m, set: set, body: jKol(set, "{"+members+"}"), ursache: u, ort: ortVor(u, "/value/0"+ort)},
		{name: name + " (Entity)", m: m, set: set, einzel: true, body: jEin(set, members), ursache: u, ort: ortVor(u, ort)},
	}
}

// ortVor liefert den Ort für eine Ursache; bei einem gültigen Fall und bei den Ursachen des Lesens steht keiner.
func ortVor(u model.DatenUrsache, ort string) string {
	switch u {
	case "", model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName:
		return ""
	}
	return ort
}

// beideGut ist beide für einen gültigen Fall; props sind die erwarteten Properties der Entity.
func beideGut(name, m, set, members string, props ...model.PropertyWert) []jfall {
	fs := beide(name, m, set, members, "", "")
	for i := range fs {
		fs[i].props = props
	}
	return fs
}

// pw ist eine erwartete Property.
func pw(name string, art model.WertArt, text string) model.PropertyWert {
	return model.PropertyWert{Name: name, Wert: model.Wert{Art: art, Text: text}}
}

func verbinde(gruppen ...[]jfall) []jfall {
	var out []jfall
	for _, g := range gruppen {
		out = append(out, g...)
	}
	return out
}
