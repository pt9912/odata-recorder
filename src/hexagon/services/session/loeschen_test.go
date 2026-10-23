package session_test

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestDelete: unbekannt ODR-E7003, offen ODR-E7009 (ohne Sessions.Delete), referenziert ODR-E7004,
// in dieser Rangfolge; eine offene Session wird nie gelöscht; die Kennung geht unverändert an den
// Port; der Bericht ist bei jedem Fehler der Nullwert (V11).
func TestDelete(t *testing.T) {
	get, del, bereinige := "Sessions.Get(s1)", "Sessions.Delete(s1)", "CleanupBlobs"
	gelungen := inbound.Loeschbericht{Entfernt: 2}
	faelle := []struct {
		name   string
		id     string
		status model.SessionStatus // leer: keine Session im Fake
		setup  func(*welt)
		ev     []string
		code   string // leer: gelungen
		http   int
	}{
		{name: "unbekannt", id: "nix", ev: []string{"Sessions.Get(nix)"}, code: "ODR-E7003", http: 404},
		{name: "leere Kennung", id: "", ev: []string{"Sessions.Get()"}, code: "ODR-E7003", http: 404},
		{name: "offen", status: model.SessionOpen, ev: []string{get}, code: "ODR-E7009", http: 409},
		{name: "offen und referenziert", status: model.SessionOpen, ev: []string{get}, code: "ODR-E7009", http: 409,
			setup: func(u *welt) { u.st.scheitere("Sessions.Delete", 0, verpackt(outbound.ErrReferenziert)) }},
		{name: "complete", status: model.SessionComplete, ev: []string{get, del, bereinige}},
		{name: "incomplete", status: model.SessionIncomplete, ev: []string{get, del, bereinige}},
		{name: "unbekannter Status", status: "zz", ev: []string{get, del, bereinige}},
		{name: "Kennung mit Leerzeichen", id: "Ab c", status: model.SessionComplete,
			ev: []string{"Sessions.Get(Ab c)", "Sessions.Delete(Ab c)", bereinige}},
		{name: "referenziert", status: model.SessionComplete, ev: []string{get, del}, code: "ODR-E7004", http: 409,
			setup: func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrReferenziert)) }},
		{name: "Delete: nicht gefunden", status: model.SessionComplete, ev: []string{get, del}, code: "ODR-E7003", http: 404,
			setup: func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrNichtGefunden)) }},
		{name: "Get beschädigt", status: model.SessionComplete, ev: []string{get}, code: "ODR-E3001", http: 500,
			setup: func(u *welt) { u.st.scheitere("Sessions.Get", 1, verpackt(outbound.ErrBeschaedigt)) }},
		{name: "Get Platte", status: model.SessionComplete, ev: []string{get}, code: "ODR-E9001", http: 500,
			setup: func(u *welt) { u.st.scheitere("Sessions.Get", 1, errDisk) }},
		{name: "Delete beschädigt", status: model.SessionComplete, ev: []string{get, del}, code: "ODR-E3001", http: 500,
			setup: func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrBeschaedigt)) }},
		{name: "Delete Platte", status: model.SessionComplete, ev: []string{get, del}, code: "ODR-E9001", http: 500,
			setup: func(u *welt) { u.st.scheitere("Sessions.Delete", 1, errDisk) }},
	}
	for _, f := range faelle {
		u := neu(t)
		id := f.id
		if id == "" && f.status != "" {
			id = "s1"
		}
		if f.status != "" {
			u.sessions.fuege(id, dienstA, t0(), f.status)
		}
		if f.setup != nil {
			f.setup(u)
		}
		b, err := u.del(context.Background(), id)
		if !slices.Equal(u.p.liste(), f.ev) {
			t.Errorf("%s: Ereignisse %v, erwartet %v", f.name, u.p.liste(), f.ev)
		}
		_, da := u.sessions.hat(id)
		if f.code == "" {
			if err != nil || b != gelungen || da {
				t.Errorf("%s: %+v, %v, Session da: %v", f.name, b, err, da)
			}
			continue
		}
		if c, st := codeVon(t, err); c != f.code || st != f.http || b != (inbound.Loeschbericht{}) {
			t.Errorf("%s: %s %d, Bericht %+v", f.name, c, st, b)
		}
		if f.status != "" && !da {
			t.Errorf("%s: die Session wurde gelöscht", f.name)
		}
		if n := u.st.anzahl(bereinige); n != 0 {
			t.Errorf("%s: Bereinigung %d-mal", f.name, n)
		}
	}

	// Wiederholung: das zweite Delete derselben Session ist unbekannt und ruft die Bereinigung nicht.
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	if _, err := u.del(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	_, err := u.del(context.Background(), "s1")
	if c, _ := codeVon(t, err); c != "ODR-E7003" {
		t.Errorf("zweites Delete: %v", err)
	}
	if n := u.st.anzahl(bereinige); n != 1 {
		t.Errorf("Bereinigung %d-mal, erwartet 1", n)
	}
}

// TestBereinigungErgebnis: Bericht und Fehler der Bereinigung ändern das Ergebnis des Löschens
// nicht; der Bericht kommt unverändert zurück (V14).
func TestBereinigungErgebnis(t *testing.T) {
	faelle := []struct {
		name    string
		bericht inbound.Loeschbericht
		fehler  error
	}{
		{"(a) Entfernt", inbound.Loeschbericht{Entfernt: 3}, nil},
		{"(b) zurückgestellt", inbound.Loeschbericht{Zurueckgestellt: true}, nil},
		{"(c) ODR-E3001", inbound.Loeschbericht{Entfernt: 2}, &inbound.Fehler{Code: model.CodeE3001, Status: 500, Ursache: errDisk}},
		{"(d) ODR-E9001", inbound.Loeschbericht{Entfernt: 1}, &inbound.Fehler{Code: model.CodeE9001, Status: 500, Ursache: errDisk}},
		{"(e) ohne Fehlerwert", inbound.Loeschbericht{Entfernt: 4}, errDisk},
		{"(f) Kontextfehler", inbound.Loeschbericht{Entfernt: 2}, context.Canceled},
	}
	for _, f := range faelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		u.wartung.bericht, u.wartung.fehler = f.bericht, f.fehler
		b, err := u.del(context.Background(), "s1")
		if err != nil || b != f.bericht {
			t.Errorf("%s: %+v, %v; erwartet %+v, nil", f.name, b, err, f.bericht)
		}
		if _, da := u.sessions.hat("s1"); da {
			t.Errorf("%s: die Session besteht noch", f.name)
		}
	}
}

// TestProtokoll: je Ereignis genau die Attribute aus SPEC-124 Nr. 9 und nie ein Fehlertext;
// Erfolg, Kontextfehler und die Fehler der Operationen selbst schreiben nichts (V15).
func TestProtokoll(t *testing.T) {
	fehl := func(c model.Code) *inbound.Fehler { return &inbound.Fehler{Code: c, Status: 500, Ursache: errDisk} }
	attrs := func(ereignis, code string) map[string]string {
		m := map[string]string{"ereignis": ereignis, "service_id": dienstA, "session_id": "s1"}
		if code != "" {
			m["error_code"] = code
		}
		return m
	}
	faelle := []struct {
		name    string
		bericht inbound.Loeschbericht
		fehler  error
		soll    []eintrag
	}{
		{"(a) ODR-E3001", inbound.Loeschbericht{Entfernt: 1}, fehl(model.CodeE3001), []eintrag{{slog.LevelError,
			"session: Bereinigung nach dem Löschen fehlgeschlagen", attrs("bereinigung_fehlgeschlagen", "ODR-E3001")}}},
		{"(b) ODR-E9001", inbound.Loeschbericht{}, fehl(model.CodeE9001), []eintrag{{slog.LevelError,
			"session: Bereinigung nach dem Löschen fehlgeschlagen", attrs("bereinigung_fehlgeschlagen", "ODR-E9001")}}},
		{"(c) ohne Fehlerwert", inbound.Loeschbericht{}, errDisk, []eintrag{{slog.LevelError,
			"session: Bereinigung nach dem Löschen fehlgeschlagen", attrs("bereinigung_fehlgeschlagen", "ODR-E9001")}}},
		{"(d) zurückgestellt", inbound.Loeschbericht{Zurueckgestellt: true}, nil, []eintrag{{slog.LevelInfo,
			"session: Bereinigung nach dem Löschen zurückgestellt", attrs("bereinigung_zurueckgestellt", "")}}},
		{"(e) gelungen", inbound.Loeschbericht{Entfernt: 2}, nil, nil},
		{"(f) Canceled", inbound.Loeschbericht{Entfernt: 1}, context.Canceled, nil},
		{"(f) DeadlineExceeded", inbound.Loeschbericht{Entfernt: 1}, context.DeadlineExceeded, nil},
		{"(f) Fehler neben Canceled", inbound.Loeschbericht{}, errors.Join(fehl(model.CodeE3001), context.Canceled), nil},
	}
	for _, f := range faelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		u.wartung.bericht, u.wartung.fehler = f.bericht, f.fehler
		if _, err := u.del(context.Background(), "s1"); err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		if got := u.log.alle(); !reflect.DeepEqual(got, f.soll) && (len(got) != 0 || len(f.soll) != 0) {
			t.Errorf("%s: Einträge %+v, erwartet %+v", f.name, got, f.soll)
		}
		keinMarker(t, f.name, u.log.alle())
	}

	// (g) Die Fehler der Operationen selbst erhält der Aufrufer als Code; ins Protokoll gelangt nichts.
	mit := func(op string, n int, e error) func(*welt) {
		return func(u *welt) { u.st.scheitere(op, n, e) }
	}
	fehlerfaelle := map[string]struct {
		setup func(*welt)
		op    func(*welt) error
	}{
		"ODR-E7004": {mit("Sessions.Delete", 1, verpackt(outbound.ErrReferenziert)), func(u *welt) error { _, err := u.del(context.Background(), "s1"); return err }},
		"ODR-E7009": {func(u *welt) { u.sessions.fuege("o1", dienstA, t0(), model.SessionOpen) }, func(u *welt) error { _, err := u.del(context.Background(), "o1"); return err }},
		"ODR-E7003": {nil, func(u *welt) error { _, err := u.del(context.Background(), "nix"); return err }},
		"ODR-E7010": {mit("Sessions.Create", 1, verpackt(outbound.ErrBereitsVorhanden)), func(u *welt) error { _, err := u.create(context.Background(), dienstA, "x1"); return err }},
		"ODR-E5005": {mit("Sessions.Create", 1, verpackt(outbound.ErrDienstUnbekannt)), func(u *welt) error { _, err := u.create(context.Background(), dienstA, ""); return err }},
		"ODR-E9001": {mit("Sessions.Get", 1, errDisk), func(u *welt) error { _, err := u.get(context.Background(), "s1"); return err }},
	}
	for code, o := range fehlerfaelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		if o.setup != nil {
			o.setup(u)
		}
		if c, _ := codeVon(t, o.op(u)); c != code {
			t.Errorf("%s: Code %s", code, c)
		}
		if got := u.log.alle(); len(got) != 0 {
			t.Errorf("%s: Einträge %+v", code, got)
		}
	}
}

// keinMarker prüft, dass der Marker in keinem Text und keinem Attribut eines Eintrags steht.
func keinMarker(t *testing.T, name string, eintraege []eintrag) {
	t.Helper()
	for _, e := range eintraege {
		alle := e.text
		for k, v := range e.attrs {
			alle += " " + k + "=" + v
		}
		if strings.Contains(alle, marker) {
			t.Errorf("%s: der Marker steht im Eintrag %q", name, alle)
		}
	}
}
