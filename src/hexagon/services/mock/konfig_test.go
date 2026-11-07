package mock_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// TestNewKonfig hält Z01 (M1): jeder Verstoß der Konfig ist model.ErrUngueltig ohne Dienst, ohne Port und mit einem
// Text ohne Wert; die acht Felder der Ports kommen aus der Reflexion über Konfig, nicht aus einer Liste.
func TestNewKonfig(t *testing.T) {
	type fall struct {
		name    string
		aendere func(k *mock.Konfig)
	}
	var faelle []fall
	felder := 0
	typ := reflect.TypeOf(mock.Konfig{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type.Kind() != reflect.Interface {
			continue
		}
		felder++
		name := f.Name
		faelle = append(faelle, fall{"nil in " + name, func(k *mock.Konfig) { reflect.ValueOf(k).Elem().FieldByName(name).SetZero() }})
	}
	if felder != 8 {
		t.Fatalf("Konfig hat %d Felder mit einer Schnittstelle, die Menge M1 nennt 8; Z01 ist nachzuziehen", felder)
	}
	satzOhne := func(name string, aendere func(s *outbound.CodecSatz)) fall {
		return fall{"Satz mit nil in " + name, func(k *mock.Konfig) {
			s := *k.Dienste[1].Codecs
			aendere(&s)
			k.Dienste[1].Codecs = &s
		}}
	}
	faelle = append(faelle,
		fall{"kein Dienst", func(k *mock.Konfig) { k.Dienste = nil }},
		fall{"leere Kennung", func(k *mock.Konfig) { k.Dienste[1].ID = "" }},
		fall{"leerer Root", func(k *mock.Konfig) { k.Dienste[1].Root = "" }},
		fall{"leeres PublicBaseURL", func(k *mock.Konfig) { k.PublicBaseURL = "" }},
		fall{"MaxBytes 0", func(k *mock.Konfig) { k.MaxBytes = 0 }},
		fall{"MaxBytes -1", func(k *mock.Konfig) { k.MaxBytes = -1 }},
		satzOhne("Metadaten", func(s *outbound.CodecSatz) { s.Metadaten = nil }),
		satzOhne("URL", func(s *outbound.CodecSatz) { s.URL = nil }),
		satzOhne("Schreiber", func(s *outbound.CodecSatz) { s.Schreiber = nil }),
	)
	if len(faelle) != 17 {
		t.Fatalf("%d Fälle, die Menge M1 nennt 17", len(faelle))
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neueUmg(t)
			u.k.Dienste = []mock.Dienst{
				{ID: "id-" + kanarienwert, Root: "/r-" + kanarienwert},
				{ID: "zwei", Root: "/zwei", Codecs: u.c.satz(false, "4.0")},
			}
			f.aendere(&u.k)
			m, err := mock.New(u.k)
			if m != nil || !errors.Is(err, model.ErrUngueltig) {
				t.Fatalf("New = %v, %v; erwartet nil und ErrUngueltig", m, err)
			}
			var mf *inbound.Fehler
			if errors.As(err, &mf) {
				t.Errorf("Fehler trägt einen Code: %v", err)
			}
			if strings.Contains(err.Error(), kanarienwert) {
				t.Errorf("Text nennt einen Wert: %q", err.Error())
			}
			if ops := u.p.ops(); len(ops) != 0 {
				t.Errorf("New rief Ports: %v", ops)
			}
		})
	}
}

// TestNewAkzeptiert hält die Grenzen von Z01: was die Konfiguration prüft, prüft New nicht (Form und Überlappung der
// Roots, Form von PublicBaseURL), und ein Satz ohne Daten und Eingabe ist gültig; New ruft keinen Port.
func TestNewAkzeptiert(t *testing.T) {
	u := neueUmg(t)
	satz := u.c.satz(false, "")
	satz.Daten, satz.Eingabe = nil, nil
	u.k.Dienste = []mock.Dienst{{ID: "a", Root: "kein-schraegstrich"}, {ID: "b", Root: "kein-schraegstrich", Codecs: satz}}
	u.k.PublicBaseURL = "keine url"
	u.k.MaxBytes = 1
	if m, err := mock.New(u.k); m == nil || err != nil {
		t.Fatalf("New = %v, %v", m, err)
	}
	if ops := u.p.ops(); len(ops) != 0 {
		t.Errorf("New rief Ports: %v", ops)
	}
}

// TestNewKopiert hält, dass der Dienst seine Konfig nicht teilt: eine Änderung von Dienste oder Satz nach New ändert den
// Dienst nicht (SPEC-195 Nr. 12).
func TestNewKopiert(t *testing.T) {
	u := neueUmg(t)
	satz := u.c.satz(false, "")
	u.k.Dienste = []mock.Dienst{{ID: dienstID, Root: dienstRoot, Codecs: satz}}
	m := u.mocker()
	u.k.Dienste[0].Root = "/anderswo"
	u.k.Dienste[0].Codecs = nil
	satz.Schreiber = nil
	satz.URL = nil
	a, err := m.Mock(u.ctx, metadatenAnfrage())
	if err != nil {
		t.Fatalf("Mock: %v", err)
	}
	lies(t, a)
}

// TestSatzWahl hält Z02: ein Dienst ohne Satz nutzt MetadatenCodec, URLCodec und Schreiber der Konfig, einer mit Satz
// nur dessen Fakes; die Version des Satzes ändert die Antwort nicht, und der Name des Kennzeichens kommt aus der Darstellung.
func TestSatzWahl(t *testing.T) {
	u := neueUmg(t)
	p2 := neuesProt()
	b := neueCodecs(p2, model.Kopfzeile{Name: "X-Fake-B", Wert: "kennzeichen-B"})
	u.k.Dienste = []mock.Dienst{{ID: dienstID, Root: dienstRoot}, {ID: "svc2", Root: "/mit", Codecs: b.satz(false, "")}}
	u.md.setze(model.Service{ID: "svc2", MetadataBlob: metaRef})
	u.ws.setze(model.Workspace{ID: "ws2", ServiceID: "svc2", BaselineSnapshotID: "b2"})
	m := u.mocker()

	a, err := m.Mock(u.ctx, metadatenAnfrage())
	if err != nil {
		t.Fatalf("ohne Satz: %v", err)
	}
	lies(t, a)
	if got := a.Header["X-Fake-Kennzeichen"]; len(got) != 1 || got[0] != "kennzeichen-1" {
		t.Errorf("Kennzeichen ohne Satz = %v", a.Header)
	}
	if u.p.anzahl("URL.ParseURL") != 1 || u.p.anzahl("Schreiber.SchreibeAnzahl") != 1 || len(p2.ops()) != 0 {
		t.Errorf("ohne Satz: Konfig %v, Satz %v", u.p.ops(), p2.ops())
	}

	u.p.leeren()
	req := anfrage("GET", "/mit/$metadata")
	req.Header["X-Recorder-Workspace"] = []string{"ws2"}
	a, err = m.Mock(u.ctx, req)
	if err != nil {
		t.Fatalf("mit Satz: %v", err)
	}
	lies(t, a)
	if got := a.Header["X-Fake-B"]; len(got) != 1 || got[0] != "kennzeichen-B" || a.Header["X-Fake-Kennzeichen"] != nil {
		t.Errorf("Kennzeichen mit Satz = %v", a.Header)
	}
	if p2.anzahl("URL.ParseURL") != 1 || p2.anzahl("Schreiber.SchreibeAnzahl") != 1 {
		t.Errorf("mit Satz: Satz %v", p2.ops())
	}
	for _, op := range u.p.ops() {
		if strings.HasPrefix(op, "URL.") || strings.HasPrefix(op, "Schreiber.") || strings.HasPrefix(op, "MetadatenCodec.") {
			t.Errorf("mit Satz rief der Dienst den Fake der Konfig: %s", op)
		}
	}

	var ref struct {
		status int
		header map[string][]string
		body   []byte
	}
	for i, v := range []string{"", "2.0", "4.0", "9.9", "OData-Version"} {
		u2 := neueUmg(t)
		u2.k.Dienste = []mock.Dienst{{ID: dienstID, Root: dienstRoot, Codecs: u2.c.satz(false, v)}}
		a, err := u2.rufe(metadatenAnfrage())
		if err != nil {
			t.Fatalf("Version %q: %v", v, err)
		}
		body := lies(t, a)
		if i == 0 {
			ref.status, ref.header, ref.body = a.Status, a.Header, body
			continue
		}
		if a.Status != ref.status || !reflect.DeepEqual(a.Header, ref.header) || !bytes.Equal(body, ref.body) {
			t.Errorf("Version %q ändert die Antwort: %d %v", v, a.Status, a.Header)
		}
	}
}

// TestQuelltextOhneVersion hält Z02: im Quelltext des Pakets ohne Tests steht nichts über die Version des Satzes und
// den Namen eines Kennzeichens.
func TestQuelltextOhneVersion(t *testing.T) {
	dateien := quelltext(t)
	for _, d := range dateien {
		b, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range []string{"Version", "OData-Version", "DataServiceVersion"} {
			if strings.Contains(string(b), w) {
				t.Errorf("%s nennt %q", d, w)
			}
		}
	}
}

// quelltext liefert die Dateien des Pakets ohne Tests.
func quelltext(t *testing.T) []string {
	t.Helper()
	alle, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var r []string
	for _, d := range alle {
		if !strings.HasSuffix(d, "_test.go") {
			r = append(r, d)
		}
	}
	if len(r) < 4 {
		t.Fatalf("nur %d Quelldateien: %v", len(r), r)
	}
	return r
}

// TestFakesZaehlenAlles hält Z21: über die Methodenmengen der acht Ports (Reflexion) zählt jede Operation, hält ihre
// Argumente fest und speist einen Fehler ein, auch die, die der Dienst nie rufen darf.
func TestFakesZaehlenAlles(t *testing.T) {
	u := neueUmg(t)
	ports := []struct {
		praefix string
		impl    any
		typ     reflect.Type
		methode int
	}{
		{"Metadaten", u.md, reflect.TypeFor[outbound.MetadataRepository](), 4},
		{"Workspaces", u.ws, reflect.TypeFor[outbound.WorkspaceRepository](), 5},
		{"Entities", u.en, reflect.TypeFor[outbound.EntityRepository](), 5},
		{"Coverage", u.co, reflect.TypeFor[outbound.CoverageRepository](), 2},
		{"Blobs", u.bl, reflect.TypeFor[outbound.BlobRepository](), 5},
		{"MetadatenCodec", u.c.mc, reflect.TypeFor[outbound.MetadataCodec](), 1},
		{"URL", u.c.url, reflect.TypeFor[outbound.URLCodec](), 1},
		{"Schreiber", u.c.sch, reflect.TypeFor[outbound.ODataSchreiber](), 5},
	}
	ctxTyp := reflect.TypeFor[context.Context]()
	gesamt := 0
	for _, pt := range ports {
		if pt.typ.NumMethod() != pt.methode {
			t.Errorf("%s hat %d Methoden, Z21 nennt %d", pt.praefix, pt.typ.NumMethod(), pt.methode)
		}
		for i := 0; i < pt.typ.NumMethod(); i++ {
			name := pt.praefix + "." + pt.typ.Method(i).Name
			gesamt++
			mv := reflect.ValueOf(pt.impl).MethodByName(pt.typ.Method(i).Name)
			typ := mv.Type()
			var args []reflect.Value
			var erwartet []any
			hatCtx := typ.NumIn() > 0 && typ.In(0) == ctxTyp
			for j := 0; j < typ.NumIn(); j++ {
				v := reflect.Zero(typ.In(j))
				switch typ.In(j).Kind() {
				case reflect.String:
					v = reflect.ValueOf("arg-" + name + "-" + string(rune('a'+j))).Convert(typ.In(j))
				case reflect.Int64:
					v = reflect.ValueOf(int64(70 + j))
				}
				if j == 0 && hatCtx {
					v = reflect.ValueOf(u.ctx)
				}
				args = append(args, v)
				if !(j == 0 && hatCtx) {
					erwartet = append(erwartet, v.Interface())
				}
			}
			for _, einspeisen := range []error{nil, errors.New("eingespeist " + name)} {
				u.p.leeren()
				u.p.setzeFehler(name, einspeisen)
				out := mv.Call(args)
				evs := u.p.von(name)
				if len(evs) != 1 || len(u.p.ops()) != 1 {
					t.Errorf("%s: %d Ereignisse, Protokoll %v", name, len(evs), u.p.ops())
					continue
				}
				if !reflect.DeepEqual(evs[0].args, erwartet) {
					t.Errorf("%s: Argumente %v, erwartet %v", name, evs[0].args, erwartet)
				}
				if hatCtx && evs[0].ctx != u.ctx {
					t.Errorf("%s: ctx nicht festgehalten", name)
				}
				got, _ := out[len(out)-1].Interface().(error)
				if einspeisen != nil && got != einspeisen {
					t.Errorf("%s: Fehler %v, erwartet %v", name, got, einspeisen)
				}
			}
			u.p.setzeFehler(name, nil)
		}
	}
	if gesamt != 28 {
		t.Errorf("%d Operationen, Z21 nennt 28", gesamt)
	}
}
