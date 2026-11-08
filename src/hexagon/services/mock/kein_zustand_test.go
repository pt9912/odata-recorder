package mock_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// TestGegenstandJeStelle hält den Gegenstand der Fehlerabbildung an den Aufrufstellen S1, S2, S3, S4 und S12 (SPEC-195
// Nr. 10, 11): 5 Stellen × 6 Fehler des Speichers, Orakel von Hand. Der Gegenstand entscheidet allein bei ErrNichtGefunden
// und ErrBereitsVorhanden (SPEC-109 Nr. 5); ein Sentinel des Speichers bleibt auch aus dem URL-Leser ODR-E3001 (SPEC-195 Nr. 6).
func TestGegenstandJeStelle(t *testing.T) {
	fehler := []error{outbound.ErrNichtGefunden, outbound.ErrBereitsVorhanden, outbound.ErrBeschaedigt, outbound.ErrBlobFehlt,
		model.ErrUngueltig, errors.New("roh " + kanarienwert)}
	e30, e90, e31, e50, e70, e77 := model.CodeE3001, model.CodeE9001, model.CodeE3001, model.CodeE5005, model.CodeE7001, model.CodeE7007
	stellen := []struct {
		name, op string
		soll     [6]model.Code
	}{
		{"S1", "URL.ParseURL", [6]model.Code{e90, e90, e30, e31, e90, e90}},
		{"S2", "Workspaces.Get", [6]model.Code{e70, e77, e30, e31, e90, e90}},
		{"S3", "Metadaten.Get", [6]model.Code{e50, e90, e30, e31, e90, e90}},
		{"S4", "Blobs.Open", [6]model.Code{e90, e90, e30, e31, e90, e90}},
		{"S12", "Schreiber.SchreibeAnzahl", [6]model.Code{e90, e90, e30, e31, e90, e90}},
	}
	n := 0
	for _, s := range stellen {
		for i, err := range fehler {
			n++
			u := neueUmg(t)
			u.p.setzeFehler(s.op, err)
			_, got := u.rufe(metadatenAnfrage())
			if f := erwarteFehler(t, s.name+" "+err.Error(), got, s.soll[i]); f != nil && !errors.Is(f, err) {
				t.Errorf("%s %v: Ursache nicht erreichbar", s.name, err)
			}
		}
	}
	if n != 30 {
		t.Errorf("%d Zellen, erwartet 30", n)
	}
}

// TestKeinZustand hält SPEC-195 Nr. 12 an jeder Lesestelle: derselbe Dienst liest nach einer Änderung des Zustands zwischen
// zwei Aufrufen neu (Inhalt des Blobs, Kennzeichen des Schreibers, Ergebnis des URL-Lesers, Workspace, Service).
func TestKeinZustand(t *testing.T) {
	u := neueUmg(t)
	m := u.mocker()
	const ref2 = "sha256:0000000000000000000000000000000000000000000000000000000000000002"
	u.bl.setzeInhalt(ref2, []byte("zweiter Inhalt"))
	inhaltB := []byte("geaenderter Inhalt")
	schritte := []struct {
		name    string
		aendere func()
		code    model.Code
		body    []byte
		kennz   string
	}{
		{"Ausgang", func() {}, 0, metaInhalt(), "kennzeichen-1"},
		{"Inhalt des Blobs", func() { u.bl.setzeInhalt(metaRef, inhaltB) }, 0, inhaltB, "kennzeichen-1"},
		{"Kennzeichen", func() { u.c.sch.kopf.Kennzeichen.Wert = "kennzeichen-2" }, 0, inhaltB, "kennzeichen-2"},
		{"URL-Leser liefert eine Menge", func() {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge}}, nil
			}
		}, model.CodeE6008, nil, ""},
		{"URL-Leser wieder $metadata", func() { u.c.url.f = nil }, 0, inhaltB, "kennzeichen-2"},
		{"Workspace fremd", func() { u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer"}) }, model.CodeE5004, nil, ""},
		{"Workspace entfernt", func() { u.ws.entferne(wsID) }, model.CodeE7001, nil, ""},
		{"Workspace wieder eigen", func() { u.ws.setze(model.Workspace{ID: wsID, ServiceID: dienstID}) }, 0, inhaltB, "kennzeichen-2"},
		{"Service ohne Blob", func() { u.md.setze(model.Service{ID: dienstID}) }, model.CodeE6004, nil, ""},
		{"Service entfernt", func() { u.md.entferne(dienstID) }, model.CodeE5005, nil, ""},
		{"Service mit anderem Blob", func() { u.md.setze(model.Service{ID: dienstID, MetadataBlob: ref2}) }, 0, []byte("zweiter Inhalt"), "kennzeichen-2"},
		{"Service wieder der erste Blob", func() { u.md.setze(model.Service{ID: dienstID, MetadataBlob: metaRef}) }, 0, inhaltB, "kennzeichen-2"},
	}
	for _, s := range schritte {
		s.aendere()
		antwort, err := m.Mock(u.ctx, metadatenAnfrage())
		if s.code != 0 {
			_ = erwarteFehler(t, s.name, err, s.code)
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", s.name, err)
			continue
		}
		kz := antwort.Header["X-Fake-Kennzeichen"]
		if got := lies(t, antwort); !bytes.Equal(got, s.body) || len(kz) != 1 || kz[0] != s.kennz {
			t.Errorf("%s: Body %q Kennzeichen %v, erwartet %q %q", s.name, got, kz, s.body, s.kennz)
		}
	}
}

// TestMockerFelder hält, dass der Dienst nur seine Konfig trägt (SPEC-195 Nr. 12): Mocker hat das eine Feld k, und die Felder
// von Konfig sind die vierzehn des Vertrags (SPEC-195 Nr. 1, SPEC-192 Nr. 1) in Arten ohne Zustand (Schnittstelle, Liste der Dienste, Text, Wahrheitswert, Zahl).
func TestMockerFelder(t *testing.T) {
	typ := reflect.TypeFor[mock.Mocker]()
	if typ.NumField() != 1 || typ.Field(0).Name != "k" || typ.Field(0).Type != reflect.TypeFor[mock.Konfig]() {
		t.Fatalf("Mocker trägt mehr als die Konfig: %d Felder", typ.NumField())
	}
	soll := map[string]reflect.Kind{"Dienste": reflect.Slice, "Metadaten": reflect.Interface, "Workspaces": reflect.Interface,
		"Entities": reflect.Interface, "Coverage": reflect.Interface, "Blobs": reflect.Interface, "MetadatenCodec": reflect.Interface,
		"URLCodec": reflect.Interface, "Schreiber": reflect.Interface, "PublicBaseURL": reflect.String, "StrictCoverage": reflect.Bool,
		"MaxBytes": reflect.Int64, "Transaktion": reflect.Interface, "Eingabe": reflect.Interface}
	k := reflect.TypeFor[mock.Konfig]()
	if k.NumField() != len(soll) {
		t.Fatalf("Konfig hat %d Felder, erwartet %d", k.NumField(), len(soll))
	}
	for i := 0; i < k.NumField(); i++ {
		if kind, da := soll[k.Field(i).Name]; !da || k.Field(i).Type.Kind() != kind {
			t.Errorf("Feld %s der Konfig ist nicht vorgesehen oder von anderer Art", k.Field(i).Name)
		}
	}
	d := reflect.TypeFor[mock.Dienst]()
	if d.NumField() != 4 || d.Field(2).Type != reflect.TypeFor[*outbound.CodecSatz]() || d.Field(3).Name != "Etag" || d.Field(3).Type.Kind() != reflect.Bool {
		t.Errorf("Dienst trägt mehr als ID, Root, Codecs und Etag (SPEC-280 Nr. 2)")
	}
}
