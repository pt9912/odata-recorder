package record_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// TestKonfigTelemetrie hält C1 (SPEC-295 Nr. 1): New liefert ErrUngueltig und keinen Dienst bei nil Telemetrie und bei jedem der 7 bisherigen
// nil-Felder, also 8 Fälle (Log bleibt Pflicht), und ruft Telemetrie in keinem Fall und im Gutfall nicht.
func TestKonfigTelemetrie(t *testing.T) {
	u := neu(t)
	faelle := map[string]func(*record.Konfig){
		"Telemetrie": func(k *record.Konfig) { k.Telemetrie = nil }, "Sessions": func(k *record.Konfig) { k.Sessions = nil },
		"Exchanges": func(k *record.Konfig) { k.Exchanges = nil }, "Upstream": func(k *record.Konfig) { k.Upstream = nil },
		"Pipeline": func(k *record.Konfig) { k.Pipeline = nil }, "Uhr": func(k *record.Konfig) { k.Uhr = nil },
		"IDs": func(k *record.Konfig) { k.IDs = nil }, "Log": func(k *record.Konfig) { k.Log = nil },
	}
	if len(faelle) != 8 {
		t.Fatalf("%d Fälle, erwartet 8", len(faelle))
	}
	for name, aendere := range faelle {
		k := u.konfig
		aendere(&k)
		if r, err := record.New(k); !errors.Is(err, model.ErrUngueltig) || r != nil {
			t.Errorf("%s nil: %v, Dienst %v", name, err, r != nil)
		}
	}
	if n := len(u.tel.Aufrufe()); n != 0 {
		t.Errorf("New rief Telemetrie %d-mal", n)
	}
}

// pruefeOhneTelemetrie hält, dass ein Aufruf Telemetrie nie rief.
func pruefeOhneTelemetrie(t *testing.T, was string, u *umgebung) {
	t.Helper()
	if a := u.tel.Aufrufe(); len(a) != 0 || len(u.upstream.gesehen()) != 0 {
		t.Errorf("%s: %d Aufrufe von Telemetrie, %d von Senden; erwartet 0 und 0", was, len(a), len(u.upstream.gesehen()))
	}
}

// vorFall ist ein Weg, der vor der Annahme endet (SPEC-295 Nr. 2); code ist der erwartete Code, 0 ein Fehler ohne Code.
type vorFall struct {
	name  string
	code  model.Code
	baue  []func(*record.Konfig)
	vorab func(u *umgebung)
	lauf  func(u *umgebung) error
}

func rufe(a inbound.Anfrage) func(u *umgebung) error {
	return func(u *umgebung) error { _, err := u.record(context.Background(), a); return err }
}

// vorFaelle ist die Menge M-VorAnnahme: 4 + 1 + 2 + 3 + 1 + 1 + 1 + 2 + 2 + 1 gleich 18 Wege.
func vorFaelle() []vorFall {
	var keinCtx context.Context
	beendet, abbrechen := context.WithCancel(context.Background())
	abbrechen()
	sess := func(werte ...string) inbound.Anfrage {
		return mitSession(anfrage("GET", "/a/x"), sessionHeader, werte...)
	}
	offen := func(id, dienst string, s model.SessionStatus) func(*umgebung) {
		return func(u *umgebung) { u.sessions.setze(id, dienst, s) }
	}
	mitBodyNil := anfrage("POST", "/a/x")
	mitBodyNil.Laenge = 1
	laenge := mitBody(anfrage("POST", "/a/x"), "application/json", "{}")
	laenge.Laenge = -2
	wurzel := func(k *record.Konfig) {
		k.Dienste = []record.Dienst{{ID: "svc-a", Root: "/", UpstreamURL: "http://a"}}
	}
	query := anfrage("GET", "/a/x")
	query.RawQuery = "a=%zz"
	return []vorFall{
		{name: "ctx nil", lauf: func(u *umgebung) error { _, err := u.r.Record(keinCtx, anfrage("GET", "/a/x")); return err }},
		{name: "Methode leer", lauf: rufe(anfrage("", "/a/x"))},
		{name: "Laenge -2", lauf: rufe(laenge)},
		{name: "Body nil bei Laenge 1", lauf: rufe(mitBodyNil)},
		{name: "ODR-E5005", code: model.CodeE5005, lauf: rufe(anfrage("GET", "/c/x"))},
		{name: "ODR-E5006", code: model.CodeE5006, baue: []func(*record.Konfig){reject}, lauf: rufe(query)},
		{name: "ODR-E9001", code: model.CodeE9001, baue: []func(*record.Konfig){wurzel}, lauf: rufe(anfrage("GET", "/a#b"))},
		{name: "ODR-E5003 leerer Wert", code: model.CodeE5003, lauf: rufe(sess(""))},
		{name: "ODR-E5003 zwei Werte", code: model.CodeE5003, lauf: rufe(sess("s1", "s1"))},
		{name: "ODR-E5003 mehrere offene", code: model.CodeE5003, vorab: offen("s2", "svc-a", model.SessionOpen), lauf: rufe(anfrage("GET", "/a/x"))},
		{name: "ODR-E7003", code: model.CodeE7003, lauf: rufe(sess("nix"))},
		{name: "ODR-E5004", code: model.CodeE5004, vorab: offen("sb", "svc-b", model.SessionOpen), lauf: rufe(sess("sb"))},
		{name: "ODR-E5009 nicht open", code: model.CodeE5009, vorab: offen("s1", "svc-a", model.SessionComplete), lauf: rufe(sess("s1"))},
		{name: "Ablage Get", vorab: func(u *umgebung) { u.sessions.getFehler = errDisk }, lauf: rufe(sess("s1"))},
		{name: "Ablage List", vorab: func(u *umgebung) { u.sessions.listFehler = errDisk }, lauf: rufe(anfrage("GET", "/a/x"))},
		{name: "Kontext Get", lauf: func(u *umgebung) error { _, err := u.r.Record(beendet, sess("s1")); return err }},
		{name: "Kontext List", lauf: func(u *umgebung) error { _, err := u.r.Record(beendet, anfrage("GET", "/a/x")); return err }},
		{name: "ODR-E5009 im Drain", code: model.CodeE5009, vorab: func(u *umgebung) {
			if _, err := u.r.Drain(context.Background(), "s1", 0); err != nil {
				u.t.Fatal(err)
			}
		}, lauf: rufe(anfrage("GET", "/a/x"))},
	}
}

// fahre baut die Umgebung des Wegs, führt ihn aus und liefert Umgebung und Fehler.
func (f vorFall) fahre(t *testing.T) (*umgebung, error) {
	t.Helper()
	u, _ := neuMitJournal(t, f.baue...)
	if f.vorab != nil {
		f.vorab(u)
	}
	return u, f.lauf(u)
}

// TestKeinSpanVorAnnahme hält C2 (SPEC-295 Nr. 2): jeder der 18 Wege endet mit seinem Fehler, ohne Aufruf von Telemetrie und ohne Senden.
func TestKeinSpanVorAnnahme(t *testing.T) {
	faelle := vorFaelle()
	if len(faelle) != 18 {
		t.Fatalf("%d Wege, erwartet 18", len(faelle))
	}
	for _, f := range faelle {
		u, err := f.fahre(t)
		var fe *inbound.Fehler
		switch {
		case err == nil:
			t.Errorf("%s: kein Fehler", f.name)
		case f.code != 0 && (!errors.As(err, &fe) || fe.Code != f.code):
			t.Errorf("%s: Fehler %v, erwartet %v", f.name, err, f.code)
		}
		pruefeOhneTelemetrie(t, f.name, u)
	}
}

// TestSpanBeginn hält C3 (SPEC-295 Nr. 3 a, 6 R1): der Beginn liegt nach Lesen der Session, Kennung und Zeit des Exchange und vor Senden, in
// beiden Wegen der Wahl (Header und die eine offene Session).
func TestSpanBeginn(t *testing.T) {
	faelle := []struct {
		name  string
		a     inbound.Anfrage
		folge []string
	}{
		{"Header", mitSession(anfrage("GET", "/a/x"), sessionHeader, "s1"), []string{"Sessions.Get", "ExchangeID", "Now", "StartSpan", "Senden"}},
		{"eine offene", anfrage("GET", "/a/x"), []string{"Sessions.List", "ExchangeID", "Now", "StartSpan", "Senden"}},
	}
	for _, f := range faelle {
		u, j := neuMitJournal(t)
		u.rundlauf(f.a)
		got := j.prot.liste()
		if len(got) < 5 || !slices.Equal(got[:5], f.folge) {
			t.Errorf("%s: Folge %v, erwartet Anfang %v", f.name, got, f.folge)
		}
	}
}

// Die drei Wege des Exchange zum Ende (SPEC-295 Nr. 3 b).
const (
	wegEOF         = "Close nach EOF"
	wegOhneEOF     = "Close ohne EOF"
	wegOhneAntwort = "ohne Antwort"
)

// laufeWeg führt einen Exchange auf dem Weg weg bis zu seinem Ende (nach der Rückkehr von Close oder Record).
func laufeWeg(t *testing.T, u *umgebung, weg string) {
	t.Helper()
	switch weg {
	case wegEOF:
		u.rundlauf(anfrage("GET", "/a/x"))
	case wegOhneEOF:
		ant, err := u.record(context.Background(), anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		schliesse(t, ant.Body)
	default:
		u.upstream.plane(&plan{lies: -1, fehler: outbound.ErrUpstreamNichtErreichbar})
		if _, err := u.record(context.Background(), anfrage("GET", "/a/x")); err == nil {
			t.Fatal("Record ohne Fehler")
		}
	}
}

// TestEndeJeAusgang hält C4 (SPEC-295 Nr. 3 b, 7 a): Menge M-Ausgänge, 3 Wege × 3 Weisen des Anhängens + 1 gleich 10; je Fall ein StartSpan und ein
// End, das End steht nach dem Anhängen und dem Eintrag ablage; im zehnten Fall kehrt Drain nicht vor dem End zurück (das End braucht Zeit).
func TestEndeJeAusgang(t *testing.T) {
	ablage := "Log record: Exchange nicht angeh" + string([]byte{0xc3, 0xa4}) + "ngt"
	soll := [][]string{
		{"ListBySession", "Append", "End"},
		{"ListBySession", "Append", ablage, "End"},
		{"ListBySession", ablage, "End"},
	}
	n := 0
	for _, weg := range []string{wegEOF, wegOhneEOF, wegOhneAntwort} {
		for i, anh := range []string{"gelungen", "Append scheitert", "ListBySession scheitert"} {
			n++
			u, j := neuMitJournal(t)
			u.exchanges.appendFehler[0], u.exchanges.listFehler = nil, nil
			switch i {
			case 1:
				u.exchanges.appendFehler[0] = errDisk
			case 2:
				u.exchanges.listFehler = errDisk
			}
			laufeWeg(t, u, weg)
			if got := j.folge("ListBySession", "Append", ablage, "End"); !slices.Equal(got, soll[i]) {
				t.Errorf("%s, %s: Folge %v, erwartet %v", weg, anh, got, soll[i])
			}
			if s, e := j.Anzahl(supporttest.OpStartSpan), j.Anzahl(supporttest.OpEnd); s != 1 || e != 1 {
				t.Errorf("%s, %s: %d StartSpan und %d End, erwartet 1 und 1", weg, anh, s, e)
			}
		}
	}
	n++
	u, j := neuMitJournal(t)
	j.vorEnd = func() { time.Sleep(50 * time.Millisecond) }
	u.upstream.plane(&plan{lies: -1, warteAufCtx: true})
	res := u.recordNebenbei(context.Background(), anfrage("GET", "/a/x"))
	empfange(t, "Senden", u.upstream.imSenden)
	d := u.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Drain", d)
	if s, e := j.Anzahl(supporttest.OpStartSpan), j.Anzahl(supporttest.OpEnd); s != 1 || e != 1 {
		t.Errorf("Drain: bei der Rückkehr %d StartSpan und %d End, erwartet 1 und 1", s, e)
	}
	empfange(t, "Record", res)
	if n != 10 {
		t.Errorf("%d Fälle, erwartet 10", n)
	}
}
