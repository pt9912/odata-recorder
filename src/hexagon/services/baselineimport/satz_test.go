package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// Die Tests dieser Datei prüfen die Wahl des Satzes der Codecs je Dienst (SPEC-127 Nr. 1, SPEC-175 Nr. 3):
// drei Dienste in einer Konfig, D0 ohne Satz (die Codecs der Konfig), D1 und D2 mit je eigenem Satz.

// satzGrenze ist MaxCaptureBody dieser Tests; keine Vorgabe der Welt trägt diesen Wert.
const satzGrenze int64 = 7654321

// satzOps sind die drei Aufrufstellen der Codecs im Import (SPEC-127 Nr. 7 f, 9).
func satzOps() []string { return []string{"ParseMetadata", "ParseCollection", "ParseEntity"} }

type satzCtxSchluessel struct{}

// satzFake ist ein Dienst der Welt mit den Fakes seiner Codecs. Jeder Satz hat ein eigenes Protokoll; der
// Fake von D0 ist der Codec der Konfig und teilt das Protokoll der Welt.
type satzFake struct {
	name, id, ses, set       string
	p                        *protokoll
	meta                     *fakeMetaCodec
	daten                    *fakeDatenCodec
	satz                     *outbound.CodecSatz
	modell                   model.ServiceModel
	metaRef, collRef, entRef string
	ctx                      func() context.Context
}

// modellFuer ist das Modell des Dienstes mit dem einen EntitySet set vom Typ Shop.Product.
func modellFuer(set string) model.ServiceModel {
	return model.NeuesServiceModel("4.0", []string{"Shop"}, modellR().EntityTypes(), []model.EntitySet{{Name: set, EntityType: typProdukt}}, nil)
}

func neuerSatzFake(name, id, ses, set string) *satzFake {
	c := context.WithValue(context.Background(), satzCtxSchluessel{}, name)
	return &satzFake{
		name: name, id: id, ses: ses, set: set, modell: modellFuer(set),
		ctx: func() context.Context { return c },
	}
}

// konfigSatz ist der Dienst ohne Satz: seine Codecs sind die der Konfig der Welt.
func konfigSatz(w *welt, name, id, ses, set string) *satzFake {
	s := neuerSatzFake(name, id, ses, set)
	s.p, s.meta, s.daten = w.p, w.meta, w.daten
	w.meta.vorgabe.modell = s.modell
	return s
}

// neuerSatz ist der Dienst mit einem eigenen Satz; seine Fakes prüfen wie die der Welt Hülle, verbrauchten Strom
// und fehlende Vorgabe.
func neuerSatz(t *testing.T, name, id, ses, set string) *satzFake {
	s := neuerSatzFake(name, id, ses, set)
	s.p = neuesProtokoll()
	s.meta = &fakeMetaCodec{p: s.p, vorgabe: metaVorgabe{modell: s.modell, bericht: model.ValidierungsBericht{Funde: []model.Fund{}}}, nachRef: map[string]metaVorgabe{}}
	s.daten = &fakeDatenCodec{p: s.p, nachRef: map[string]datenVorgabe{}, aufrufe: map[string][]datenAufruf{}}
	s.satz = &outbound.CodecSatz{Version: "4.0", Metadaten: s.meta, Daten: s.daten}
	t.Cleanup(func() {
		if s.meta.fremd+s.daten.fremd > 0 {
			t.Errorf("Satz %s: ein Codec erhielt einen Strom, der nicht der von Open ist (Hülle): %d", name, s.meta.fremd+s.daten.fremd)
		}
		if s.meta.verbraucht+s.daten.verbraucht > 0 {
			t.Errorf("Satz %s: ein Codec erhielt einen Strom, den der Dienst vorher gelesen hat: %d", name, s.meta.verbraucht+s.daten.verbraucht)
		}
		if s.daten.fehlt > 0 {
			t.Errorf("Satz %s: der Fake des Daten-Codecs hatte für %d Aufrufe keine Vorgabe", name, s.daten.fehlt)
		}
	})
	return s
}

func (s *satzFake) anfrage() inbound.ImportAnfrage {
	return inbound.ImportAnfrage{ImportID: "imp-" + s.id, ServiceID: s.id, SourceSessionIDs: []string{s.ses}}
}

// baue trägt den Dienst in die Konfig der Welt ein (mit seinem Satz oder ohne) und legt seine Session an: das
// $metadata, eine Collection und eine Einzel-Entity des Sets, jede mit eigenem Inhalt und eigener Vorgabe.
func (w *welt) baue(s *satzFake) {
	d := baselineimport.Dienst{ID: s.id, UpstreamURL: upstream, Codecs: s.satz}
	if i := indexDienst(w.konfig.Dienste, s.id); i >= 0 {
		w.konfig.Dienste[i] = d
	} else {
		w.konfig.Dienste = append(w.konfig.Dienste, d)
	}
	w.sessions.m[s.ses] = model.Session{ID: s.ses, ServiceID: s.id, CreatedAt: t0(), Status: model.SessionComplete}
	b := &bauer{w: w, ses: s.ses}
	inhalt := func(art string) aenderung {
		return func(e *model.Exchange) {
			bytes := []byte(art + " " + s.name)
			e.ResponseBlob = refVon(bytes)
			w.blobs.inhalt[e.ResponseBlob] = bytes
		}
	}
	s.metaRef = b.metadaten("<edmx " + s.name + "/>").ResponseBlob
	coll := b.roh("/"+s.set, "", inhalt("Collection"))
	ent := b.roh("/"+s.set+"(1)", "", inhalt("Entity"))
	s.collRef, s.entRef = coll.ResponseBlob, ent.ResponseBlob
	s.daten.mu.Lock()
	defer s.daten.mu.Unlock()
	s.daten.nachRef[s.collRef] = datenVorgabe{seite: seite(produkt(1, "A-"+s.name, "1"), produkt(2, "B-"+s.name, "2"))}
	s.daten.nachRef[s.entRef] = datenVorgabe{entity: produkt(1, "A-"+s.name, "1")}
}

func indexDienst(ds []baselineimport.Dienst, id string) int {
	for i, d := range ds {
		if d.ID == id {
			return i
		}
	}
	return -1
}

// satzWelt baut die Welt mit D0, D1 und D2 und fährt den Dienst nicht an.
func satzWelt(t *testing.T) (*welt, []*satzFake) {
	t.Helper()
	w := neu(t)
	w.konfig.MaxCaptureBody = satzGrenze
	alle := []*satzFake{
		konfigSatz(w, "D0", dienstID, sitzungID, "Products"),
		neuerSatz(t, "D1", "d1", "ses-d1", "Waren"),
		neuerSatz(t, "D2", "d2", "ses-d2", "Gueter"),
	}
	for _, s := range alle {
		w.baue(s)
	}
	for i, a := range alle {
		for _, b := range alle[i+1:] {
			if reflect.DeepEqual(a.modell, b.modell) {
				t.Fatalf("die Modelle von %s und %s gleichen sich", a.name, b.name)
			}
		}
	}
	return w, alle
}

// satzStand hält die Zähler der drei Aufrufstellen jedes Satzes fest.
func satzStand(alle []*satzFake) map[*satzFake]map[string]int {
	out := map[*satzFake]map[string]int{}
	for _, s := range alle {
		out[s] = map[string]int{}
		for _, op := range satzOps() {
			out[s][op] = s.p.anzahl(op)
		}
	}
	return out
}

func (w *welt) stromVon(ref string) *strom {
	for _, s := range w.blobs.strome() {
		if s.ref == ref {
			return s
		}
	}
	return nil
}

// pruefeAufruf prüft eine Aufrufstelle eines Imports von ziel: genau ein Aufruf beim Satz von ziel, bei jedem anderen
// so viele, wie sein eigener Import brauchte (fremde: 0, wenn nur ziel importierte, sonst 1); dasselbe ctx, die
// Grenze, das Modell von ParseMetadata desselben Dienstes, das EntitySet und derselbe Strom wie der von Open
// (SPEC-127 Nr. 1, 2, 7 f, 9).
func (w *welt) pruefeAufruf(t *testing.T, alle []*satzFake, ziel *satzFake, op string, vor map[*satzFake]map[string]int, fremde int) {
	t.Helper()
	for _, s := range alle {
		want := fremde
		if s == ziel {
			want = 1
		}
		if got := s.p.anzahl(op) - vor[s][op]; got != want {
			t.Errorf("%s %s: %d Aufrufe bei %s, erwartet %d", ziel.name, op, got, s.name, want)
		}
	}
	n := vor[ziel][op]
	if ctxs := ziel.p.ctxVon(op); len(ctxs) != n+1 || ctxs[n] != ziel.ctx() {
		t.Errorf("%s %s: ctx ist nicht der des Aufrufs", ziel.name, op)
	}
	var src io.Reader
	var grenze int64
	var ref string
	if op == "ParseMetadata" {
		if len(ziel.meta.aufrufe) != n+1 {
			t.Fatalf("%s %s: %d Aufrufe festgehalten", ziel.name, op, len(ziel.meta.aufrufe))
		}
		a := ziel.meta.aufrufe[n]
		src, grenze, ref = a.src, a.grenze, ziel.metaRef
	} else {
		if len(ziel.daten.aufrufe[op]) != n+1 {
			t.Fatalf("%s %s: %d Aufrufe festgehalten", ziel.name, op, len(ziel.daten.aufrufe[op]))
		}
		a := ziel.daten.aufrufe[op][n]
		src, grenze = a.src, a.grenze
		ref = map[string]string{"ParseCollection": ziel.collRef, "ParseEntity": ziel.entRef}[op]
		if !reflect.DeepEqual(a.m, ziel.modell) {
			t.Errorf("%s %s: das Modell ist nicht das von ParseMetadata desselben Dienstes", ziel.name, op)
		}
		if a.entitySet != ziel.set {
			t.Errorf("%s %s: EntitySet %q, erwartet %q", ziel.name, op, a.entitySet, ziel.set)
		}
	}
	if grenze != satzGrenze {
		t.Errorf("%s %s: grenze %d, erwartet %d", ziel.name, op, grenze, satzGrenze)
	}
	if st := w.stromVon(ref); st == nil || src != io.Reader(st) {
		t.Errorf("%s %s: src ist nicht der Strom von Open für %s", ziel.name, op, ref)
	}
}

// TestSatzWahlJeAufrufstelle (Z1): je Aufrufstelle (3) und Dienst (3) geht der Aufruf an den Codec des Dienstes und
// an keinen anderen; D0 ohne Satz nimmt die Codecs der Konfig.
func TestSatzWahlJeAufrufstelle(t *testing.T) {
	w, alle := satzWelt(t)
	w.start()
	faelle := 0
	for _, ziel := range alle {
		vor := satzStand(alle)
		if _, err := w.imp.Import(ziel.ctx(), ziel.anfrage()); err != nil {
			t.Fatalf("Import von %s: %v", ziel.name, err)
		}
		for _, op := range satzOps() {
			faelle++
			w.pruefeAufruf(t, alle, ziel, op, vor, 0)
		}
	}
	if faelle != 9 {
		t.Errorf("Menge %d, erwartet 9", faelle)
	}
}

// TestSatzFehlerJeAufrufstelle (Z2): ein Fehler des Codecs von D1 oder D2 wird an jeder Aufrufstelle (3) nach Nr. 17
// abgebildet, ein verpackter Kontextfehler kehrt unverändert und ohne Code zurück; jeder Strom ist genau einmal
// geschlossen, danach ruft der Dienst weder Blobs.Begin noch Imports.Publish noch einen weiteren Codec.
func TestSatzFehlerJeAufrufstelle(t *testing.T) {
	faelle := 0
	for i, op := range satzOps() {
		for _, wahl := range []int{1, 2} {
			for _, art := range []string{"Verstoß", "Kontextfehler"} {
				faelle++
				w, alle := satzWelt(t)
				ziel := alle[wahl]
				var fehler error
				verpackt := fmt.Errorf("verpackt: %w", context.Canceled)
				var mf *model.MetadatenFehler
				var df *model.DatenFehler
				switch {
				case art == "Kontextfehler":
					fehler = verpackt
				case op == "ParseMetadata":
					mf = &model.MetadatenFehler{Ursache: model.UrsacheStruktur, Ort: "Shop/" + ziel.name}
					fehler = mf
				default:
					df = &model.DatenFehler{Ursache: model.DatenWert, Ort: "/value/0/" + ziel.name}
					fehler = df
				}
				ziel.p.scheitere(op, 1, fehler)
				name := fmt.Sprintf("%s %s %s", op, ziel.name, art)
				vor := satzStand(alle)
				_, err := w.start().imp.Import(ziel.ctx(), ziel.anfrage())
				switch {
				case err == nil:
					t.Errorf("%s: kein Fehler", name)
				case art == "Kontextfehler":
					if !errors.Is(err, verpackt) || !ohneCode(err) {
						t.Errorf("%s: %v", name, err)
					}
				case mf != nil:
					var got *model.MetadatenFehler
					erwarteFehler(t, name, err, model.CodeE6004, 422)
					if !errors.As(err, &got) || got != mf {
						t.Errorf("%s: %v", name, err)
					}
				default:
					var got *model.DatenFehler
					erwarteFehler(t, name, err, model.CodeE6101, 422)
					if !errors.As(err, &got) || got != df {
						t.Errorf("%s: %v", name, err)
					}
				}
				strome := w.blobs.strome()
				if len(strome) != i+1 {
					t.Errorf("%s: %d Ströme geöffnet, erwartet %d", name, len(strome), i+1)
				}
				for k, st := range strome {
					if st.geschlossen() != 1 {
						t.Errorf("%s: Strom %d %d-mal geschlossen", name, k, st.geschlossen())
					}
				}
				for _, o := range []string{"Blobs.Begin", "Imports.Publish"} {
					if n := w.p.anzahl(o); n != 0 {
						t.Errorf("%s: %s %d-mal gerufen", name, o, n)
					}
				}
				for k, o := range satzOps() {
					want := 0
					if k <= i {
						want = 1
					}
					if got := ziel.p.anzahl(o) - vor[ziel][o]; got != want {
						t.Errorf("%s: %s %d-mal, erwartet %d", name, o, got, want)
					}
				}
			}
		}
	}
	if faelle != 12 {
		t.Errorf("Menge %d, erwartet 12", faelle)
	}
}

// satzDienste baut drei Dienste a, b, c mit je einem Satz der Fakes der Welt; fehler[k] ersetzt den Satz des Dienstes k.
func satzDienste(w *welt, ersatz map[int]*outbound.CodecSatz) []baselineimport.Dienst {
	var ds []baselineimport.Dienst
	for k, id := range []string{"a", "b", "c"} {
		satz := &outbound.CodecSatz{Version: "4.0", Metadaten: w.meta, Daten: w.daten}
		if e, ok := ersatz[k]; ok {
			satz = e
		}
		ds = append(ds, baselineimport.Dienst{ID: id, UpstreamURL: upstream, Codecs: satz})
	}
	return ds
}

// TestNewSatzTeilNil (Z3): ein Satz mit Metadaten oder Daten nil ist ErrUngueltig ohne Code und ohne Dienst, an jeder
// Stelle der Liste (3) in jeder Form (3); der Gutfall mit denselben Nachbarn besteht.
func TestNewSatzTeilNil(t *testing.T) {
	w := neu(t)
	if imp, err := baselineimport.New(func() baselineimport.Konfig { k := w.konfig; k.Dienste = satzDienste(w, nil); return k }()); err != nil || imp == nil {
		t.Fatalf("Gutfall: %v", err)
	}
	formen := map[string]*outbound.CodecSatz{
		"Metadaten nil": {Daten: w.daten},
		"Daten nil":     {Metadaten: w.meta},
		"beide nil":     {},
	}
	n := 0
	for pos := 0; pos < 3; pos++ {
		for name, satz := range formen {
			n++
			k := w.konfig
			k.Dienste = satzDienste(w, map[int]*outbound.CodecSatz{pos: satz})
			imp, err := baselineimport.New(k)
			if !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
				t.Errorf("Position %d, %s: %v, Dienst %v", pos, name, err, imp != nil)
			}
		}
	}
	// Satz teil-nil und leere ID am selben Dienst: welcher Verstoß gemeldet wird, ist nicht zugesagt.
	k := w.konfig
	k.Dienste = []baselineimport.Dienst{{UpstreamURL: upstream, Codecs: &outbound.CodecSatz{}}}
	if imp, err := baselineimport.New(k); !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
		t.Errorf("Satz teil-nil und leere ID: %v", err)
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

// TestNewKonfigCodecsPflicht (Z4): MetadatenCodec und DatenCodec der Konfig bleiben Pflicht, ob kein, jeder oder ein
// Teil der Dienste einen Satz trägt (2 Felder mal 3 Formen).
func TestNewKonfigCodecsPflicht(t *testing.T) {
	w := neu(t)
	mitSatz := baselineimport.Dienst{ID: "a", UpstreamURL: upstream, Codecs: &outbound.CodecSatz{Metadaten: w.meta, Daten: w.daten}}
	ohneSatz := baselineimport.Dienst{ID: "b", UpstreamURL: upstream}
	formen := map[string][]baselineimport.Dienst{
		"alle ohne Satz": {ohneSatz},
		"alle mit Satz":  {mitSatz},
		"gemischt":       {mitSatz, ohneSatz},
	}
	n := 0
	for feld, nimm := range map[string]func(*baselineimport.Konfig){
		"MetadatenCodec": func(k *baselineimport.Konfig) { k.MetadatenCodec = nil },
		"DatenCodec":     func(k *baselineimport.Konfig) { k.DatenCodec = nil },
	} {
		for name, ds := range formen {
			n++
			k := w.konfig
			k.Dienste = ds
			if imp, err := baselineimport.New(k); err != nil || imp == nil {
				t.Fatalf("%s %s: Gutfall %v", feld, name, err)
			}
			nimm(&k)
			if imp, err := baselineimport.New(k); !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
				t.Errorf("%s nil, %s: %v", feld, name, err)
			}
		}
	}
	if n != 6 {
		t.Errorf("Menge %d, erwartet 6", n)
	}
}

// TestSatzUebrigeFelder (Z5): Version und NurLesend ändern weder New noch Import; das Ergebnis gleicht byte für byte dem
// Import mit den Codecs der Konfig (Version 4.0 im Modell des Fakes, 4 Versionen mal 2 Werte); ein Satz darf von
// zwei Diensten geteilt werden.
func TestSatzUebrigeFelder(t *testing.T) {
	ref := neu(t)
	d0 := konfigSatz(ref, "D0", dienstID, sitzungID, "Products")
	ref.baue(d0)
	soll, err := ref.start().imp.Import(d0.ctx(), d0.anfrage())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, v := range []string{"", "4.0", "2.0", "x"} {
		for _, nur := range []bool{false, true} {
			n++
			w := neu(t)
			s := neuerSatz(t, "D0", dienstID, sitzungID, "Products")
			s.satz.Version, s.satz.NurLesend = v, nur
			w.baue(s)
			got, err := w.start().imp.Import(s.ctx(), s.anfrage())
			if err != nil {
				t.Errorf("Version %q, NurLesend %v: %v", v, nur, err)
				continue
			}
			if got.Import.ReportJSON != soll.Import.ReportJSON {
				t.Errorf("Version %q, NurLesend %v: ReportJSON weicht ab", v, nur)
			}
			if s.p.anzahl("ParseMetadata") != 1 || w.p.anzahl("ParseMetadata") != 0 {
				t.Errorf("Version %q, NurLesend %v: der Satz wurde nicht gewählt", v, nur)
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge %d, erwartet 8", n)
	}
	w := neu(t)
	a := neuerSatz(t, "A", "ga", "ses-ga", "Products")
	b := *a
	b.name, b.id, b.ses = "B", "gb", "ses-gb"
	w.baue(a)
	w.baue(&b)
	w.start()
	for _, s := range []*satzFake{a, &b} {
		if _, err := w.imp.Import(s.ctx(), s.anfrage()); err != nil {
			t.Errorf("geteilter Satz, Dienst %s: %v", s.id, err)
		}
	}
	if got := a.p.anzahl("ParseMetadata"); got != 2 {
		t.Errorf("geteilter Satz: %d Aufrufe von ParseMetadata, erwartet 2", got)
	}
}

// TestNormierungBehaeltSatz (Z6): der Satz bleibt bei seinem Dienst, auch wenn der Aufrufer die Liste nach New ändert
// (Satz und Kennung, 2 Dienste); eine doppelte Kennung ist mit verschiedenen Sätzen ungültig.
func TestNormierungBehaeltSatz(t *testing.T) {
	n := 0
	for _, wahl := range []int{1, 2} {
		n++
		w, alle := satzWelt(t)
		ziel := alle[wahl]
		w.start()
		for i := range w.konfig.Dienste {
			w.konfig.Dienste[i].Codecs = nil
			w.konfig.Dienste[i].ID += "-geaendert"
		}
		vor := satzStand(alle)
		if _, err := w.imp.Import(ziel.ctx(), ziel.anfrage()); err != nil {
			t.Fatalf("%s: %v", ziel.name, err)
		}
		for _, op := range satzOps() {
			w.pruefeAufruf(t, alle, ziel, op, vor, 0)
		}
	}
	w := neu(t)
	k := w.konfig
	k.Dienste = []baselineimport.Dienst{
		{ID: "a", UpstreamURL: upstream, Codecs: &outbound.CodecSatz{Metadaten: w.meta, Daten: w.daten}},
		{ID: "a", UpstreamURL: upstream, Codecs: &outbound.CodecSatz{Metadaten: w.meta, Daten: w.daten}},
	}
	if imp, err := baselineimport.New(k); !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
		t.Errorf("doppelte Kennung mit verschiedenen Sätzen: %v", err)
	}
	if n != 2 {
		t.Errorf("Menge %d, erwartet 2", n)
	}
}

// TestSatzGleichzeitig (Z7, Wirkungs-Fänger): die Imports von D0, D1 und D2 laufen gleichzeitig; eine Schranke hält
// jeden in ParseMetadata seines Codecs, bis alle drei dort sind, eine zweite in ParseCollection; die Antwort der
// Einzel-Entity liest danach jeder, nachdem alle das $metadata gelesen haben. Jeder Codec sieht nur Aufrufe seines
// Dienstes. Eine Sperre um den ganzen Import lässt die Frist der Schranke ablaufen.
func TestSatzGleichzeitig(t *testing.T) {
	w, alle := satzWelt(t)
	w.start()
	tor := neueSchranke(len(alle))
	tor2 := neueSchranke(len(alle))
	for _, s := range alle {
		s.p.setzeHaken("ParseMetadata", func(int) { tor.warte() })
		s.p.setzeHaken("ParseCollection", func(int) { tor2.warte() })
	}
	vor := satzStand(alle)
	var wg sync.WaitGroup
	ergebnis := make([]lauf, len(alle))
	for i, s := range alle {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ergebnis[i].r, ergebnis[i].err = w.imp.Import(s.ctx(), s.anfrage())
		}()
	}
	wg.Wait()
	tor.pruefe(t, "ParseMetadata")
	tor2.pruefe(t, "ParseCollection")
	for i, s := range alle {
		if ergebnis[i].err != nil || ergebnis[i].r.Import.ServiceID != s.id {
			t.Errorf("Import von %s: %v, Dienst %q", s.name, ergebnis[i].err, ergebnis[i].r.Import.ServiceID)
		}
		for _, op := range satzOps() {
			w.pruefeAufruf(t, alle, s, op, vor, 1)
		}
	}
}

// TestSatzKeinAufrufVorDemLesen (Z8): ein Import, der vor dem Lesen endet, und Get rufen keinen Codec, für jeden Dienst
// (3); die Wahl des Satzes liegt hinter der Prüfung der ServiceID (die unbekannte Kennung ist kein Panic).
func TestSatzKeinAufrufVorDemLesen(t *testing.T) {
	// vorher bereitet die Welt vor dem gezählten Aufruf vor, anfrage ändert die Anfrage des Dienstes; code 0 und
	// ungueltig falsch heißt: der Aufruf gelingt.
	type fall struct {
		name      string
		vorher    func(t *testing.T, w *welt, ziel *satzFake)
		anfrage   func(a *inbound.ImportAnfrage, alle []*satzFake, ziel *satzFake)
		code      model.Code
		http      int
		ungueltig bool
	}
	ersterImport := func(t *testing.T, w *welt, ziel *satzFake) {
		if _, err := w.imp.Import(ziel.ctx(), ziel.anfrage()); err != nil {
			t.Fatalf("erster Import: %v", err)
		}
	}
	faelle := []fall{
		{name: "Anfrage ungültig", ungueltig: true, anfrage: func(a *inbound.ImportAnfrage, _ []*satzFake, _ *satzFake) { a.ImportID = "" }},
		{name: "ServiceID unbekannt", code: model.CodeE5005, http: 404, anfrage: func(a *inbound.ImportAnfrage, _ []*satzFake, _ *satzFake) { a.ServiceID = "unbekannt" }},
		{name: "Wiederholung mit gleichem Hash", vorher: ersterImport},
		{name: "Wiederholung mit anderem Hash", vorher: ersterImport, code: model.CodeE6102, http: 409,
			anfrage: func(a *inbound.ImportAnfrage, _ []*satzFake, z *satzFake) { a.MetadataHash = z.metaRef }},
		{name: "Session unbekannt", code: model.CodeE7003, http: 404,
			anfrage: func(a *inbound.ImportAnfrage, _ []*satzFake, _ *satzFake) {
				a.SourceSessionIDs = []string{"ses-unbekannt"}
			}},
		{name: "Session eines anderen Service", code: model.CodeE5004, http: 409,
			anfrage: func(a *inbound.ImportAnfrage, alle []*satzFake, z *satzFake) {
				for _, s := range alle {
					if s != z {
						a.SourceSessionIDs = []string{s.ses}
						return
					}
				}
			}},
		{name: "Session offen", code: model.CodeE7009, http: 409,
			vorher: func(_ *testing.T, w *welt, z *satzFake) {
				w.sessions.m[z.ses] = model.Session{ID: z.ses, ServiceID: z.id, CreatedAt: t0(), Status: model.SessionOpen}
			}},
	}
	n := 0
	for _, f := range faelle {
		for k := 0; k < 3; k++ {
			n++
			w, alle := satzWelt(t)
			w.start()
			ziel := alle[k]
			if f.vorher != nil {
				f.vorher(t, w, ziel)
			}
			vor := satzStand(alle)
			a := ziel.anfrage()
			if f.anfrage != nil {
				f.anfrage(&a, alle, ziel)
			}
			_, err := w.imp.Import(ziel.ctx(), a)
			switch {
			case f.code != 0:
				erwarteFehler(t, f.name+" "+ziel.name, err, f.code, f.http)
			case f.ungueltig:
				if !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) {
					t.Errorf("%s %s: %v", f.name, ziel.name, err)
				}
			case err != nil:
				t.Errorf("%s %s: %v", f.name, ziel.name, err)
			}
			keinAufruf(t, f.name+" "+ziel.name, alle, vor)
		}
	}
	for k := 0; k < 3; k++ {
		n++
		w, alle := satzWelt(t)
		w.start()
		ziel := alle[k]
		res, err := w.imp.Import(ziel.ctx(), ziel.anfrage())
		if err != nil {
			t.Fatal(err)
		}
		vor := satzStand(alle)
		if got, err := w.imp.Get(ziel.ctx(), res.Import.ID); err != nil || got.ID != res.Import.ID {
			t.Errorf("Get %s: %v", ziel.name, err)
		}
		_, err = w.imp.Get(ziel.ctx(), "imp-unbekannt")
		erwarteFehler(t, "Get unbekannt "+ziel.name, err, model.CodeE7005, 404)
		keinAufruf(t, "Get "+ziel.name, alle, vor)
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// keinAufruf prüft, dass keine Aufrufstelle eines Satzes seit vor gezählt hat.
func keinAufruf(t *testing.T, name string, alle []*satzFake, vor map[*satzFake]map[string]int) {
	t.Helper()
	for _, s := range alle {
		for _, op := range satzOps() {
			if got := s.p.anzahl(op) - vor[s][op]; got != 0 {
				t.Errorf("%s: %s von %s %d-mal gerufen", name, op, s.name, got)
			}
		}
	}
}
