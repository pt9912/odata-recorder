package jsoncodec_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.ODataSchreiber = (*jsoncodec.Schreiber)(nil)

// bs ist der Backslash als Byte: ein Escape aus Backslash, `u` und Hex-Ziffern entsteht aus Bytewerten (AGENTS.md §3.18).
const bs = "\x5c"

func esc(hex string) string { return bs + "u" + hex }

const wurzel = "http://h:8080/odata"

var sch = jsoncodec.NewSchreiber()

// ziel ist der Fake von dst: er hält jeden Aufruf fest und zählt jede Operation, auch die, die der Schreiber nie rufen darf.
type ziel struct {
	log       *[]string
	writes    [][]byte
	stringN   int
	readFromN int
	closeN    int
	err       error
	kurz      int // -1: vollständig; -2: Länge-1; sonst die Zahl der Bytes, die er meldet
}

func neuesZiel() *ziel { return &ziel{log: &[]string{}, kurz: -1} }

func (z *ziel) Write(p []byte) (int, error) {
	*z.log = append(*z.log, "Write")
	z.writes = append(z.writes, append([]byte{}, p...))
	switch {
	case z.err != nil:
		return 0, z.err
	case z.kurz == -2:
		return len(p) - 1, nil
	case z.kurz >= 0 && z.kurz < len(p):
		return z.kurz, nil
	}
	return len(p), nil
}

func (z *ziel) WriteString(s string) (int, error) { z.stringN++; return len(s), nil }

func (z *ziel) ReadFrom(io.Reader) (int64, error) { z.readFromN++; return 0, nil }

func (z *ziel) Close() error { z.closeN++; return nil }

func (z *ziel) body() string {
	var b strings.Builder
	for _, w := range z.writes {
		b.Write(w)
	}
	return b.String()
}

// leer prüft, dass nichts in dst steht und keine Methode gerufen wurde.
func (z *ziel) leer(t *testing.T, was string) {
	t.Helper()
	if len(z.writes) != 0 || z.stringN != 0 || z.readFromN != 0 || z.closeN != 0 {
		t.Errorf("%s: dst berührt: Write %d, WriteString %d, ReadFrom %d, Close %d", was, len(z.writes), z.stringN, z.readFromN, z.closeN)
	}
}

// nurEinWrite prüft, dass genau ein Write und sonst keine Methode gerufen wurde.
func (z *ziel) nurEinWrite(t *testing.T, was string) {
	t.Helper()
	if len(z.writes) != 1 || z.stringN != 0 || z.readFromN != 0 || z.closeN != 0 {
		t.Errorf("%s: Write %d, WriteString %d, ReadFrom %d, Close %d", was, len(z.writes), z.stringN, z.readFromN, z.closeN)
	}
}

// kontext ist der Fake von ctx: er zählt Err, Done und Deadline und teilt sein Ereignisprotokoll mit dem Ziel.
type kontext struct {
	log                 *[]string
	err                 error
	errN, doneN, deadlN int
}

func neuerKontext(z *ziel, err error) *kontext { return &kontext{log: z.log, err: err} }

// Value liefert nichts; der Schreiber liest keinen Wert aus dem Kontext.
func (*kontext) Value(any) any { return nil }

func (k *kontext) Err() error {
	k.errN++
	*k.log = append(*k.log, "Err")
	return k.err
}

func (k *kontext) Done() <-chan struct{} {
	k.doneN++
	*k.log = append(*k.log, "Done")
	return nil
}

func (k *kontext) Deadline() (time.Time, bool) {
	k.deadlN++
	*k.log = append(*k.log, "Deadline")
	return time.Time{}, false
}

// ---- Modell --------------------------------------------------------------------------------------------------------

func prop(name, typ string, nullable bool) model.Property {
	return model.Property{Name: name, EdmType: typ, Nullable: nullable}
}

func produktProps() []model.Property {
	return []model.Property{
		prop("Id", "Edm.Int32", false), prop("Name", "Edm.String", true), prop("Preis", "Edm.Decimal", true),
		prop("Gross", "Edm.Int64", true), prop("Aktiv", "Edm.Boolean", true), prop("Ref", "Edm.Guid", true),
		prop("Tag", "Edm.Date", true), prop("Zeit", "Edm.DateTimeOffset", true), prop("Blob", "Edm.Binary", true),
	}
}

// alleProps hat die acht Typen je nicht nullbar (S, B, I32, I64, Dec, G, D, T) und nullbar (NS, NB, ...).
func alleProps() []model.Property {
	typen := model.UnterstuetzteEDMTypen()
	namen := []string{"S", "B", "I32", "I64", "Dec", "G", "D", "T"}
	var out []model.Property
	for i, n := range namen {
		out = append(out, prop(n, typen[i], false))
	}
	for i, n := range namen {
		out = append(out, prop("N"+n, typen[i], true))
	}
	return out
}

func ringTyp(name, basis string) model.EntityType {
	return model.EntityType{Name: name, BaseType: basis, Keys: []string{"Id"}, Properties: []model.Property{prop("Id", "Edm.Int32", false)}}
}

func testModell() model.ServiceModel { return modellVersion("4.0") }

func modellVersion(v string) model.ServiceModel {
	extra := append(produktProps(), prop("Extra", "Edm.String", true))
	typen := []model.EntityType{
		{Name: "Ns.Product", Keys: []string{"Id"}, Properties: produktProps(),
			NavigationProperties: []model.NavigationProperty{{Name: "Kategorie", Target: "Ns.Product"}}},
		{Name: "Ns.Sonder", BaseType: "Ns.Product", Keys: []string{"Id"}, Properties: extra},
		{Name: "Ns.Sonder2", BaseType: "Ns.Sonder", Keys: []string{"Id"}, Properties: extra},
		{Name: "Ns.Abstrakt", BaseType: "Ns.Product", Abstract: true, Keys: []string{"Id"}, Properties: produktProps()},
		ringTyp("Ns.Fremd", ""),
		ringTyp("Ns.R1", "Ns.R1"),
		ringTyp("Ns.R2a", "Ns.R2b"), ringTyp("Ns.R2b", "Ns.R2a"),
		ringTyp("Ns.R3a", "Ns.R3b"), ringTyp("Ns.R3b", "Ns.R3c"), ringTyp("Ns.R3c", "Ns.R3a"),
		ringTyp("Ns.R4a", "Ns.R4b"), ringTyp("Ns.R4b", "Ns.R4c"), ringTyp("Ns.R4c", "Ns.R4d"), ringTyp("Ns.R4d", "Ns.R4a"),
		{Name: "Ns.Alle", Keys: []string{"S"}, Properties: alleProps()},
	}
	sets := []model.EntitySet{
		{Name: "Produkte", EntityType: "Ns.Product"}, {Name: "Fremde", EntityType: "Ns.Fremd"},
		{Name: "Alles", EntityType: "Ns.Alle"}, {Name: "OhneTyp", EntityType: "Ns.Fehlt"}, {Name: "Produkte2", EntityType: "Ns.Product"},
	}
	return model.NeuesServiceModel(v, []string{"Ns"}, typen, sets, nil)
}

var modell = testModell()

func pw(name string, art model.WertArt, text string) model.PropertyWert {
	return model.PropertyWert{Name: name, Wert: model.Wert{Art: art, Text: text}}
}

func produkt(id string, name string) model.Entity {
	return model.Entity{Typ: "Ns.Product", Schluessel: id, Properties: []model.PropertyWert{
		pw("Id", model.ArtInt32, id), pw("Name", model.ArtString, name),
	}}
}

func gutEntity() model.EntityAntwort {
	return model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: produkt("1", "A")}
}

func gutCollection() model.CollectionAntwort {
	return model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: []model.Entity{produkt("1", "A"), produkt("2", "B"), produkt("3", "C")}}
}

func fehlerCode() model.Fehlerobjekt {
	f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[0])
	return f
}

// operation ist eine der vier Operationen mit einem gültigen und einem ungültigen Argument.
type operation struct {
	name          string
	gut, schlecht func(ctx context.Context, dst io.Writer) (model.Darstellung, error)
}

func operationen() []operation {
	schlechteEntity := gutEntity()
	schlechteEntity.Set = "Unbekannt"
	schlechteCollection := gutCollection()
	schlechteCollection.Set = "Unbekannt"
	return []operation{
		{"SchreibeEntity",
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeEntity(c, d, modell, gutEntity())
			},
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeEntity(c, d, modell, schlechteEntity)
			}},
		{"SchreibeCollection",
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeCollection(c, d, modell, gutCollection())
			},
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeCollection(c, d, modell, schlechteCollection)
			}},
		{"SchreibeAnzahl",
			func(c context.Context, d io.Writer) (model.Darstellung, error) { return sch.SchreibeAnzahl(c, d, 5) },
			func(c context.Context, d io.Writer) (model.Darstellung, error) { return sch.SchreibeAnzahl(c, d, -1) }},
		{"SchreibeFehler",
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeFehler(c, d, fehlerCode())
			},
			func(c context.Context, d io.Writer) (model.Darstellung, error) {
				return sch.SchreibeFehler(c, d, model.Fehlerobjekt{})
			}},
	}
}

// istUngueltig prüft die Klasse des Fehlers: ErrUngueltig, ohne Code, ohne DatenFehler (SPEC-168 Nr. 1).
func istUngueltig(t *testing.T, err error, was string) {
	t.Helper()
	var coded interface{ Code() model.Code }
	var df *model.DatenFehler
	switch {
	case !errors.Is(err, model.ErrUngueltig):
		t.Errorf("%s: kein ErrUngueltig: %v", was, err)
	case errors.As(err, &coded):
		t.Errorf("%s: trägt einen Code: %v", was, err)
	case errors.As(err, &df):
		t.Errorf("%s: ist ein DatenFehler: %v", was, err)
	}
}

// scheitert ruft f und prüft die Klasse ErrUngueltig, die Nullwerte von Darstellung und dst: der Ergebnis-Helfer aller Fehlerfälle.
func scheitert(t *testing.T, was string, f func(ctx context.Context, dst io.Writer) (model.Darstellung, error)) {
	t.Helper()
	z := neuesZiel()
	d, err := f(context.Background(), z)
	istUngueltig(t, err, was)
	if d != (model.Darstellung{}) {
		t.Errorf("%s: Darstellung %+v, will den Nullwert", was, d)
	}
	z.leer(t, was)
}

func entityScheitert(t *testing.T, was string, a model.EntityAntwort) {
	t.Helper()
	scheitert(t, was, func(c context.Context, d io.Writer) (model.Darstellung, error) {
		return sch.SchreibeEntity(c, d, modell, a)
	})
}

func collectionScheitert(t *testing.T, was string, a model.CollectionAntwort) {
	t.Helper()
	scheitert(t, was, func(c context.Context, d io.Writer) (model.Darstellung, error) {
		return sch.SchreibeCollection(c, d, modell, a)
	})
}

// entityBody schreibt die Entity und liefert den Body; ein Fehler ist ein Fehlschlag des Tests.
func entityBody(t *testing.T, was string, a model.EntityAntwort) string {
	t.Helper()
	z := neuesZiel()
	_, err := sch.SchreibeEntity(context.Background(), z, modell, a)
	if err != nil {
		t.Fatalf("%s: %v", was, err)
	}
	z.nurEinWrite(t, was)
	return z.body()
}

func collectionBody(t *testing.T, was string, a model.CollectionAntwort) string {
	t.Helper()
	z := neuesZiel()
	_, err := sch.SchreibeCollection(context.Background(), z, modell, a)
	if err != nil {
		t.Fatalf("%s: %v", was, err)
	}
	z.nurEinWrite(t, was)
	return z.body()
}
