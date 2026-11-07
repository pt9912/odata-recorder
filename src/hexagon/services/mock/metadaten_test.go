package mock_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// metadatenInhalte sind die Inhalte von Z14, aus Bytewerten gebaut (AGENTS.md §3.18): kein UTF-8, mit BOM, 1 MiB.
func metadatenInhalte() map[string][]byte {
	gross := make([]byte, 1<<20)
	for i := range gross {
		gross[i] = byte(i*31 + i>>8)
	}
	bom := append([]byte{0xEF, 0xBB, 0xBF}, []byte("<?xml?><edmx/>")...)
	return map[string][]byte{
		"kein UTF-8": {0xFF, 0xFE, 0x00, 0x80, 0xC3, 0x28, '<', '/', 0xF0, 0x28, 0x8C, 0x28, 0x0D, 0x0A},
		"mit BOM":    bom,
		"1 MiB":      gross,
	}
}

// TestMetadaten hält Z14: Status 200, der Body ist der Strom des Blobs byte-gleich, Content-Type application/xml, das
// Kennzeichen unverändert aus der Darstellung von SchreibeAnzahl(ctx, dst, 0) und sonst kein Header; Content-Type und Body
// der Probe erscheinen nie, MetadatenCodec wird nicht gerufen.
func TestMetadaten(t *testing.T) {
	inhalte := metadatenInhalte()
	if len(inhalte) != 3 {
		t.Fatal("Z14: drei Inhalte")
	}
	for name, inhalt := range inhalte {
		for _, mitSatz := range []bool{false, true} {
			u := neueUmg(t)
			u.bl.inhalt[metaRef] = inhalt
			kennz := model.Kopfzeile{Name: "X-Fake-Kennzeichen", Wert: "kennzeichen-1"}
			if mitSatz {
				c := neueCodecs(u.p, model.Kopfzeile{Name: "X-Satz-Kennzeichen", Wert: "satz-2"})
				u.k.Dienste[0].Codecs = c.satz(false, "")
				kennz = c.sch.kopf.Kennzeichen
			}
			antwort, err := u.rufe(metadatenAnfrage())
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			ct := antwort.Header["Content-Type"]
			kz := antwort.Header[kennz.Name]
			if antwort.Status != 200 || len(ct) != 1 || ct[0] != "application/xml" || len(kz) != 1 || kz[0] != kennz.Wert || len(antwort.Header) != 2 {
				t.Errorf("%s (Satz %v): Status %d Header %v", name, mitSatz, antwort.Status, antwort.Header)
			}
			if got := lies(t, antwort); !bytes.Equal(got, inhalt) {
				t.Errorf("%s (Satz %v): Body %d Bytes, erwartet %d byte-gleich", name, mitSatz, len(got), len(inhalt))
			}
			proben := u.p.von("Schreiber.SchreibeAnzahl")
			if len(proben) != 1 || proben[0].args[0] == nil || proben[0].args[1] != int64(0) || proben[0].ctx != u.ctx {
				t.Errorf("%s: Probe %v", name, proben)
			}
			opens := u.p.von("Blobs.Open")
			if len(opens) != 1 || opens[0].args[0] != metaRef || opens[0].ctx != u.ctx {
				t.Errorf("%s: Open %v", name, opens)
			}
			if n := u.p.anzahl("MetadatenCodec.ParseMetadata"); n != 0 {
				t.Errorf("%s: ParseMetadata %d", name, n)
			}
			if s := u.bl.strome(); len(s) != 1 || s[0].geschlossen() != 1 {
				t.Errorf("%s: Ströme %v", name, s)
			}
		}
	}
}

// TestMetadatenOptionen hält Z15: die sieben Optionen der URL haben an $metadata keine Wirkung.
func TestMetadatenOptionen(t *testing.T) {
	optionen := []model.Optionen{
		{Filter: &model.FilterOption{Text: "a eq 1"}},
		{OrderBy: []model.Sortierung{{Name: "a", Absteigend: true}}},
		{Top: model.ZahlOption{Gesetzt: true, Wert: 0}},
		{Skip: model.ZahlOption{Gesetzt: true, Wert: 5}},
		{Count: model.WahrOption{Gesetzt: true, Wert: true}},
		{Select: &model.SelectOption{Namen: []string{"a"}}},
		{Sonstige: []model.SonstigeOption{{Name: "$format", Wert: "json"}}},
	}
	if len(optionen) != 7 {
		t.Fatal("Z15: sieben Optionen")
	}
	lauf := func(o model.Optionen) (int, map[string][]string, []byte, []string) {
		u := neueUmg(t)
		u.c.url.f = func(string, string) (model.ODataURL, error) {
			return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMetadaten}, Optionen: o}, nil
		}
		antwort, err := u.rufe(metadatenAnfrage())
		if err != nil {
			t.Fatalf("Optionen %+v: %v", o, err)
		}
		return antwort.Status, antwort.Header, lies(t, antwort), u.p.ops()
	}
	s0, h0, b0, o0 := lauf(model.Optionen{})
	for _, o := range optionen {
		s, h, b, ops := lauf(o)
		if s != s0 || !reflect.DeepEqual(h, h0) || !bytes.Equal(b, b0) || !reflect.DeepEqual(ops, o0) {
			t.Errorf("Option %+v ändert die Antwort: %d %v %d Bytes %v", o, s, h, len(b), ops)
		}
	}
}

// fehlerFall ist ein Fehlerfall von $metadata: wie er entsteht, welcher Code folgt und ob ein Strom geöffnet wurde.
type fehlerFall struct {
	name  string
	bau   func(u *umg)
	code  model.Code
	strom bool
}

// metadatenFehlerFaelle sind die fünf Fehler von Z16: drei von Open, zwei der Probe.
func metadatenFehlerFaelle() []fehlerFall {
	return []fehlerFall{
		{"Blob fehlt", func(u *umg) { u.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }, model.CodeE3001, false},
		{"Blob beschädigt", func(u *umg) { u.p.setzeFehler("Blobs.Open", outbound.ErrBlobBeschaedigt) }, model.CodeE3001, false},
		{"Open sonstiger", func(u *umg) { u.p.setzeFehler("Blobs.Open", errors.New("roh "+kanarienwert)) }, model.CodeE9001, false},
		{"Probe ErrUngueltig", func(u *umg) { u.p.setzeFehler("Schreiber.SchreibeAnzahl", model.ErrUngueltig) }, model.CodeE9001, true},
		{"Probe sonstiger", func(u *umg) { u.p.setzeFehler("Schreiber.SchreibeAnzahl", errors.New("roh "+kanarienwert)) }, model.CodeE9001, true},
	}
}

// kennzeichenFaelle sind die drei leeren Kennzeichen von Z16 (SPEC-198 Nr. 8): Name leer, Wert leer, beide leer; der Strom
// ist geöffnet, der Fehler ODR-E9001 hat keine Ursache.
func kennzeichenFaelle() []fehlerFall {
	leer := func(k model.Kopfzeile) func(u *umg) {
		return func(u *umg) { u.c.sch.kopf.Kennzeichen = k }
	}
	return []fehlerFall{
		{"Kennzeichen ohne Namen", leer(model.Kopfzeile{Wert: "w"}), model.CodeE9001, true},
		{"Kennzeichen ohne Wert", leer(model.Kopfzeile{Name: "X-Fake-Kennzeichen"}), model.CodeE9001, true},
		{"Kennzeichen leer", leer(model.Kopfzeile{}), model.CodeE9001, true},
	}
}

// TestKennzeichenLeer hält Z16 für ein leeres Kennzeichen der Probe: ODR-E9001/500 ohne Ursache, der Strom genau einmal
// geschlossen, GET wie HEAD, und kein Header mit leerem Namen.
func TestKennzeichenLeer(t *testing.T) {
	for _, f := range kennzeichenFaelle() {
		for _, methode := range []string{"GET", "HEAD"} {
			u := neueUmg(t)
			f.bau(u)
			antwort, err := u.rufe(anfrage(methode, dienstRoot+"/$metadata"))
			fe := erwarteFehler(t, f.name+" "+methode, err, f.code)
			if fe != nil && fe.Ursache != nil {
				t.Errorf("%s %s: Ursache %v", f.name, methode, fe.Ursache)
			}
			if st := u.bl.strome(); len(st) != 1 || st[0].geschlossen() != 1 {
				t.Errorf("%s %s: Ströme %v, erwartet einer, genau einmal geschlossen", f.name, methode, st)
			}
			if _, da := antwort.Header[""]; da || antwort.Body != nil {
				t.Errorf("%s %s: Antwort nicht der Nullwert: %+v", f.name, methode, antwort)
			}
		}
	}
}

// TestMetadatenFehler hält Z16: ErrBlobFehlt und ErrBlobBeschaedigt sind ODR-E3001, jeder andere Fehler ODR-E9001; bei einem
// Fehler der Probe wird der geöffnete Strom genau einmal geschlossen, bei einem Fehler von Open gibt es keinen.
func TestMetadatenFehler(t *testing.T) {
	faelle := metadatenFehlerFaelle()
	if len(faelle) != 5 {
		t.Fatal("Z16: fünf Fehler")
	}
	for _, f := range faelle {
		for _, methode := range []string{"GET", "HEAD"} {
			u := neueUmg(t)
			f.bau(u)
			_, err := u.rufe(anfrage(methode, dienstRoot+"/$metadata"))
			fe := erwarteFehler(t, f.name+" "+methode, err, f.code)
			if fe != nil && (fe.Ursache == nil || strings.Contains(fe.Error(), kanarienwert)) {
				t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
			}
			stroeme := u.bl.strome()
			switch {
			case f.strom && (len(stroeme) != 1 || stroeme[0].geschlossen() != 1):
				t.Errorf("%s %s: Ströme %v, erwartet einer, genau einmal geschlossen", f.name, methode, stroeme)
			case !f.strom && len(stroeme) != 0:
				t.Errorf("%s %s: Strom geöffnet", f.name, methode)
			}
		}
	}
	u := neueUmg(t)
	antwort, err := u.rufe(metadatenAnfrage())
	if err != nil {
		t.Fatal(err)
	}
	if s := u.bl.strome(); len(s) != 1 || s[0].geschlossen() != 0 {
		t.Errorf("Erfolg: der Dienst schloss den Strom: %v", s)
	}
	lies(t, antwort)
}

// TestHeadMetadaten hält Z17: HEAD liefert Status und Header von GET und einen leeren Body (nicht nil), schließt den
// geöffneten Strom und meldet dieselben Fehler.
func TestHeadMetadaten(t *testing.T) {
	faelle := append([]fehlerFall{{"Erfolg", func(*umg) {}, 0, true}}, append(metadatenFehlerFaelle(), kennzeichenFaelle()...)...)
	for _, f := range faelle {
		type ergebnis struct {
			status int
			header map[string][]string
			code   model.Code
			body   []byte
			zu     int
		}
		var erg [2]ergebnis
		for i, methode := range []string{"GET", "HEAD"} {
			u := neueUmg(t)
			f.bau(u)
			antwort, err := u.rufe(anfrage(methode, dienstRoot+"/$metadata"))
			e := ergebnis{status: antwort.Status, header: antwort.Header}
			if err != nil {
				e.code = fehlerMitCode(t, err).Code
			} else {
				zu := 0
				if s := u.bl.strome(); len(s) == 1 {
					zu = s[0].geschlossen()
				}
				e.zu = zu
				e.body = lies(t, antwort)
			}
			erg[i] = e
		}
		get, head := erg[0], erg[1]
		if get.status != head.status || !reflect.DeepEqual(get.header, head.header) || get.code != head.code {
			t.Errorf("%s: HEAD %+v ungleich GET %+v", f.name, head, get)
		}
		if f.code == 0 {
			if len(get.body) == 0 || len(head.body) != 0 || get.zu != 0 || head.zu != 1 {
				t.Errorf("%s: Body GET %d Bytes, HEAD %d Bytes, geschlossen bei Rückkehr GET %d, HEAD %d", f.name, len(get.body), len(head.body), get.zu, head.zu)
			}
		}
	}
}
