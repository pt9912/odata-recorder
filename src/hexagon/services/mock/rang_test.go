package mock_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// stufe ist eine der acht Stufen von SPEC-195 Nr. 3, die dieser Dienst erreicht: wie ein Aufbau sie verletzt, welcher Code
// folgt (0 für ErrUngueltig ohne Code) und welche Aufrufe bis dahin stattgefunden haben.
type stufe struct {
	name    string
	code    model.Code
	bau     func(u *umg, a *inbound.Anfrage)
	ops     []string
	stufe   int
	keinCtx bool
}

// stufenM4 liefert die Formen der acht Stufen: Argumente vier (ctx nil, leere Methode, Laenge -2, Body nil bei Laenge 1),
// URL zwei (Fehler des Lesers, Art außerhalb der vier), jede andere Stufe eine; stufe ist die Nummer der Stufe.
func stufenM4() []stufe {
	url, ws, md := "URL.ParseURL", "Workspaces.Get", "Metadaten.Get"
	urlFehler := func(u *umg, _ *inbound.Anfrage) {
		u.c.url.f = func(string, string) (model.ODataURL, error) {
			return model.ODataURL{}, &model.URLFehler{Ursache: model.URLSegment}
		}
	}
	return []stufe{
		{name: "Argumente ctx nil", bau: func(*umg, *inbound.Anfrage) {}, stufe: 0, keinCtx: true},
		{name: "Argumente Methode leer", bau: func(_ *umg, a *inbound.Anfrage) { a.Methode = "" }, stufe: 0},
		{name: "Argumente Laenge -2", bau: func(_ *umg, a *inbound.Anfrage) { a.Laenge = -2 }, stufe: 0},
		{name: "Argumente Body nil", bau: func(_ *umg, a *inbound.Anfrage) { a.Laenge, a.Body = 1, nil }, stufe: 0},
		{name: "kein Dienst", code: model.CodeE5005, bau: func(_ *umg, a *inbound.Anfrage) { a.Pfad = "/nirgends" }, stufe: 1},
		{name: "Zuordnung", code: model.CodeE5003, bau: func(_ *umg, a *inbound.Anfrage) { delete(a.Header, headerName) }, stufe: 2},
		{name: "URL Fehler", code: model.CodeE6005, bau: urlFehler, ops: []string{url}, stufe: 3},
		{name: "URL Art 0", code: model.CodeE9001, bau: func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, nil }
		}, ops: []string{url}, stufe: 3},
		{name: "Methode", code: model.CodeE6002, bau: func(_ *umg, a *inbound.Anfrage) { a.Methode = "OPTIONS" }, ops: []string{url}, stufe: 4},
		{name: "Workspace unbekannt", code: model.CodeE7001, bau: func(u *umg, _ *inbound.Anfrage) { u.ws.entferne(wsID) },
			ops: []string{url, ws}, stufe: 5},
		{name: "Workspace fremd", code: model.CodeE5004, bau: func(u *umg, _ *inbound.Anfrage) {
			u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer", BaselineSnapshotID: "b"})
		}, ops: []string{url, ws}, stufe: 6},
		{name: "Service ohne Metadaten", code: model.CodeE6004, bau: func(u *umg, _ *inbound.Anfrage) { u.md.setze(model.Service{ID: dienstID}) },
			ops: []string{url, ws, md}, stufe: 7},
	}
}

// TestRangfolge hält Z11 (M4): je zwei Verstöße der acht Stufen, der erste entscheidet und die Ports dahinter bleiben
// ungerufen; „Workspace unbekannt“ und „fremd“ schließen sich aus, bleiben 27 von 28 Paaren, dazu die acht einzelnen.
func TestRangfolge(t *testing.T) {
	stufen := stufenM4()
	if len(stufen) != 12 {
		t.Fatalf("%d Formen, M4 nennt 12 (8 Stufen)", len(stufen))
	}
	pruefe := func(name string, verletzt ...stufe) {
		u := neueUmg(t)
		a := metadatenAnfrage()
		// Der erste Verstoß entscheidet: die späteren Stufen zuerst setzen, damit die frühere eine gleiche Größe überschreibt.
		for k := len(verletzt) - 1; k >= 0; k-- {
			verletzt[k].bau(u, &a)
		}
		erste := verletzt[0]
		m := u.mocker()
		ctx := u.ctx
		if slices.ContainsFunc(verletzt, func(s stufe) bool { return s.keinCtx }) {
			ctx = keinKontext()
		}
		_, err := m.Mock(ctx, a)
		if erste.code == 0 {
			var f *inbound.Fehler
			if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &f) {
				t.Errorf("%s: Fehler %v, erwartet ErrUngueltig ohne Code", name, err)
			}
		} else {
			_ = erwarteFehler(t, name, err, erste.code)
		}
		if ops := u.p.ops(); !slices.Equal(ops, erste.ops) {
			t.Errorf("%s: Aufrufe %v, erwartet %v", name, ops, erste.ops)
		}
	}
	paare := 0
	for i, s := range stufen {
		pruefe(s.name, s)
		for j := i + 1; j < len(stufen); j++ {
			if s.stufe == stufen[j].stufe || s.stufe == 5 && stufen[j].stufe == 6 {
				continue
			}
			paare++
			pruefe(s.name+" vor "+stufen[j].name, s, stufen[j])
		}
	}
	if paare != 58 {
		t.Errorf("%d Paare, erwartet 58 (27 Paare der Stufen, aufgefächert nach den Formen)", paare)
	}
}

// fall ist ein Fehlerfall des Dienstes, an dem Z19 die Fehlerbildung und Z24 den Verzicht auf das Fehlerobjekt prüft.
type fall struct {
	name string
	bau  func(u *umg, a *inbound.Anfrage)
	code model.Code
}

// fehlerFaelle sind alle Fehlerzeilen von Z08 bis Z16; die Fehlerzeilen des Lesens stehen in lesen_fehler_test.go.
func fehlerFaelle() []fall {
	var r []fall
	klassen := urlKlassen()
	for _, ursache := range model.AlleURLUrsachen() {
		r = append(r, fall{"URL " + string(ursache), func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{}, &model.URLFehler{Ursache: ursache, Ort: kanarienwert}
			}
		}, klassen[ursache]})
	}
	for _, ursache := range model.AlleKompatUrsachen() {
		r = append(r, fall{"Kompat " + string(ursache), func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{}, &model.KompatFehler{Ursache: ursache, Ort: kanarienwert}
			}
		}, model.CodeE6006})
	}
	roh := errors.New("roh " + kanarienwert)
	r = append(r,
		fall{"URL ErrUngueltig", func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("URL.ParseURL", model.ErrUngueltig) }, model.CodeE9001},
		fall{"URL roh", func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("URL.ParseURL", roh) }, model.CodeE9001},
		fall{"Art 0", func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, nil }
		}, model.CodeE9001},
		fall{"kein Dienst", func(_ *umg, a *inbound.Anfrage) { a.Pfad = "/" + kanarienwert }, model.CodeE5005},
		fall{"Header fehlt", func(_ *umg, a *inbound.Anfrage) { delete(a.Header, headerName) }, model.CodeE5003},
		fall{"Methode POST", func(_ *umg, a *inbound.Anfrage) { a.Methode = "OPTIONS" }, model.CodeE6002},
		fall{"Methode nur lesend", func(u *umg, a *inbound.Anfrage) {
			a.Methode = "DELETE"
			u.k.Dienste[0].Codecs = u.c.satz(true, "")
		}, model.CodeE6006},
		fall{"Workspace unbekannt", func(u *umg, _ *inbound.Anfrage) { u.ws.entferne(wsID) }, model.CodeE7001},
		fall{"Workspace fremd", func(u *umg, _ *inbound.Anfrage) { u.ws.setze(model.Workspace{ID: wsID, ServiceID: "x"}) }, model.CodeE5004},
		fall{"Workspace beschädigt", func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("Workspaces.Get", outbound.ErrBeschaedigt) }, model.CodeE3001},
		fall{"Workspace roh", func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("Workspaces.Get", roh) }, model.CodeE9001},
		fall{"Service unbekannt", func(u *umg, _ *inbound.Anfrage) { u.md.entferne(dienstID) }, model.CodeE5005},
		fall{"Service ohne Blob", func(u *umg, _ *inbound.Anfrage) { u.md.setze(model.Service{ID: dienstID}) }, model.CodeE6004},
		fall{"Service roh", func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("Metadaten.Get", roh) }, model.CodeE9001},
	)
	for _, f := range append(metadatenFehlerFaelle(), kennzeichenFaelle()...) {
		r = append(r, fall{"Metadaten " + f.name, func(u *umg, _ *inbound.Anfrage) { f.bau(u) }, f.code})
	}
	return r
}

// spec195Nr10 ist eine Handkopie der Liste der erreichbaren Codes in SPEC-195 Nr. 10 (15 Codes); kein Lauf liest die
// Spezifikation, eine Änderung der Liste dort fängt dieser Test nicht.
func spec195Nr10() []string {
	return []string{"ODR-E3001", "ODR-E5003", "ODR-E5004", "ODR-E5005", "ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E6004",
		"ODR-E6005", "ODR-E6006", "ODR-E6007", "ODR-E6008", "ODR-E6009", "ODR-E7001", "ODR-E9001"}
}

// laufeFall baut einen Aufbau, verletzt ihn nach dem Fall und ruft den Dienst; Request und Fehler der Ports tragen den
// Kanarienwert.
func laufeFall(t *testing.T, f fall) (*umg, error) {
	t.Helper()
	u := neueUmg(t)
	a := metadatenAnfrage()
	a.RawQuery = "c=" + kanarienwert
	a.Header["X-Canary"] = []string{kanarienwert}
	f.bau(u, &a)
	_, err := u.rufe(a)
	return u, err
}

// TestFehlerWert hält Z19 (M5): über alle Fehlerzeilen ist jeder Fehler mit Code ein *inbound.Fehler mit dem einen Status des
// Codes und einem Text ohne Wert; die Menge der Codes ist genau die der elf Codes dieses Teils von SPEC-195 Nr. 10.
func TestFehlerWert(t *testing.T) {
	erwartet := map[string]bool{}
	for _, c := range spec195Nr10() {
		if c != "ODR-E6003" && c != "ODR-E6008" && c != "ODR-E6007" && c != "ODR-E6009" {
			erwartet[c] = true
		}
	}
	if len(spec195Nr10()) != 15 || len(erwartet) != 11 {
		t.Fatalf("Liste: %d Codes, davon %d in diesem Teil; erwartet 15 und 11", len(spec195Nr10()), len(erwartet))
	}
	alle := model.AlleCodes()
	for c := range erwartet {
		code, ok := model.ParseCode(c)
		if !ok || !slices.Contains(alle, code) || len(code.Stati()) == 0 {
			t.Errorf("%s ist kein Code des Katalogs", c)
		}
	}
	gesehen := map[string]bool{}
	faelle := fehlerFaelle()
	for _, f := range faelle {
		_, err := laufeFall(t, f)
		fe := erwarteFehler(t, f.name, err, f.code)
		if fe == nil {
			continue
		}
		gesehen[fe.Code.String()] = true
		// Die Codes, die der Dienst ohne Tabelle bildet, tragen keine Ursache (SPEC-195 Nr. 10).
		ohne := slices.Contains([]string{"kein Dienst", "Header fehlt", "Methode POST", "Workspace fremd", "Service ohne Blob", "Art 0"}, f.name) ||
			strings.HasPrefix(f.name, "Metadaten Kennzeichen")
		if ohne && fe.Ursache != nil {
			t.Errorf("%s: Ursache %v, erwartet keine", f.name, fe.Ursache)
		}
		if !fe.Code.StatusErlaubt(fe.Status) || fe.Status != fe.Code.Stati()[0] {
			t.Errorf("%s: Status %d", f.name, fe.Status)
		}
		if strings.Contains(fe.Error(), kanarienwert) {
			t.Errorf("%s: Text nennt einen Wert: %q", f.name, fe.Error())
		}
	}
	if !reflect.DeepEqual(gesehen, erwartet) {
		t.Errorf("erreichte Codes %v, erwartet %v", gesehen, erwartet)
	}
	if len(faelle) < 40 {
		t.Errorf("nur %d Fehlerzeilen", len(faelle))
	}
}

// TestKeinFehlerobjekt hält Z24: über alle Fehlerzeilen ruft der Dienst SchreibeFehler nie.
func TestKeinFehlerobjekt(t *testing.T) {
	for _, f := range fehlerFaelle() {
		u, err := laufeFall(t, f)
		if err == nil {
			t.Errorf("%s: kein Fehler", f.name)
		}
		if n := u.p.anzahl("Schreiber.SchreibeFehler"); n != 0 {
			t.Errorf("%s: SchreibeFehler %d", f.name, n)
		}
	}
}

// TestKontext hält Z20 (M6): an den Stellen S2, S3, S4 und S12 kehrt ein Kontextfehler in vier Formen unverändert und ohne
// Code zurück, danach wird kein weiterer Port gerufen und der Strom geschlossen; jeder Aufruf erhält den ctx des Aufrufs.
func TestKontext(t *testing.T) {
	stellen := []struct {
		name  string
		op    string
		mitCd error
		ops   []string
		strom bool
	}{
		{"S2", "Workspaces.Get", outbound.ErrNichtGefunden, []string{"URL.ParseURL", "Workspaces.Get"}, false},
		{"S3", "Metadaten.Get", outbound.ErrNichtGefunden, []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get"}, false},
		{"S4", "Blobs.Open", outbound.ErrBlobFehlt, []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open"}, false},
		{"S12", "Schreiber.SchreibeAnzahl", model.ErrUngueltig, []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open", "Schreiber.SchreibeAnzahl"}, true},
	}
	formen := []struct {
		name string
		bau  func(codeFehler error) error
	}{
		{"Canceled", func(error) error { return context.Canceled }},
		{"DeadlineExceeded", func(error) error { return context.DeadlineExceeded }},
		{"verpackt", func(error) error { return fmt.Errorf("port %s: %w", kanarienwert, context.Canceled) }},
		{"Join neben Code", func(c error) error { return errors.Join(c, context.DeadlineExceeded) }},
	}
	n := 0
	for _, s := range stellen {
		for _, f := range formen {
			n++
			u := neueUmg(t)
			injiziert := f.bau(s.mitCd)
			u.p.setzeFehler(s.op, injiziert)
			_, err := u.rufe(metadatenAnfrage())
			if err == nil || err != injiziert {
				t.Errorf("%s %s: Fehler %v, erwartet den unveränderten Kontextfehler", s.name, f.name, err)
			}
			var fe *inbound.Fehler
			if errors.As(err, &fe) {
				t.Errorf("%s %s: Fehler trägt einen Code", s.name, f.name)
			}
			if ops := u.p.ops(); !slices.Equal(ops, s.ops) {
				t.Errorf("%s %s: Aufrufe %v, erwartet %v", s.name, f.name, ops, s.ops)
			}
			stroeme := u.bl.strome()
			if s.strom && (len(stroeme) != 1 || stroeme[0].geschlossen() != 1) {
				t.Errorf("%s %s: Strom nicht genau einmal geschlossen: %v", s.name, f.name, stroeme)
			}
		}
	}
	if n != 16 {
		t.Errorf("%d Fälle, M6 nennt 16", n)
	}

	u := neueUmg(t)
	antwort, err := u.rufe(metadatenAnfrage())
	if err != nil {
		t.Fatal(err)
	}
	lies(t, antwort)
	for _, s := range stellen {
		evs := u.p.von(s.op)
		if len(evs) != 1 || evs[0].ctx != u.ctx || evs[0].ctx.Value(schluesselKtx{}) != "ctx-des-aufrufs" {
			t.Errorf("%s: ctx nicht unverändert weitergereicht: %v", s.name, evs)
		}
	}
}
