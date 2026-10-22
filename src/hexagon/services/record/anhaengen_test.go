package record_test

import (
	"context"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

func (u *umgebung) bericht(id string) inbound.DrainBericht {
	u.t.Helper()
	b, err := u.r.Drain(context.Background(), id, 0)
	if err != nil {
		u.t.Fatal(err)
	}
	return b
}

// TestSequenz: sequence_no in der Reihenfolge des Endes, lückenlos je Session, die Zahl beim
// ersten Anhängen aus ListBySession (B18).
func TestSequenz(t *testing.T) {
	u := neu(t)
	antA, err := u.record(context.Background(), anfrage("GET", "/a/A"))
	if err != nil {
		t.Fatal(err)
	}
	u.uhr.Advance(time.Second)
	antB, err := u.record(context.Background(), anfrage("GET", "/a/B"))
	if err != nil {
		t.Fatal(err)
	}
	liesUndSchliesse(t, antB.Body)
	liesUndSchliesse(t, antA.Body)
	l := u.exchanges.angehaengt()
	if len(l) != 2 {
		t.Fatalf("A und B: %d angehängt", len(l))
	}
	if l[0].Path != "/B" || l[0].SequenceNo != 1 || l[1].Path != "/A" || l[1].SequenceNo != 2 || l[1].StartedAt.After(l[0].StartedAt) {
		t.Errorf("Reihenfolge: %s %d, %s %d", l[0].Path, l[0].SequenceNo, l[1].Path, l[1].SequenceNo)
	}

	v := neu(t)
	v.exchanges.vorhanden["s1"] = make([]model.Exchange, 3)
	v.rundlauf(anfrage("GET", "/a/x"))
	v.rundlauf(anfrage("GET", "/a/x"))
	var nummern []int
	for _, e := range v.exchanges.angehaengt() {
		nummern = append(nummern, e.SequenceNo)
	}
	if !slices.Equal(nummern, []int{4, 5}) || v.exchanges.listAufrufe != 1 {
		t.Errorf("drei vorhanden: %v, ListBySession %d", nummern, v.exchanges.listAufrufe)
	}

	w := neu(t)
	w.sessions.setze("s2", "svc-a", model.SessionOpen)
	for _, s := range []string{"s1", "s2", "s1", "s2", "s2"} {
		w.rundlauf(mitSession(anfrage("GET", "/a/x"), sessionHeader, s))
	}
	soll := map[string][]int{"s1": {1, 2}, "s2": {1, 2, 3}}
	got := map[string][]int{}
	for _, e := range w.exchanges.angehaengt() {
		got[e.SessionID] = append(got[e.SessionID], e.SequenceNo)
	}
	if !slices.Equal(got["s1"], soll["s1"]) || !slices.Equal(got["s2"], soll["s2"]) {
		t.Errorf("zwei Sessions: %v", got)
	}
	if l := w.exchanges.angehaengt(); len(l) != 5 || l[0].ID == l[2].ID {
		t.Errorf("%d angehängt, erwartet fünf Exchanges mit eigener Kennung", len(l))
	}
}

// TestSequenzGleichzeitig: 20 Exchanges einer Session enden zugleich; die Anhänge laufen
// nacheinander (B19). Mutation: die Sperre der Folge entfernt (Lock und Unlock in haengeAn);
// gemessen je Prozess in 100 Läufen ist der Test rot in 100, mit dem Detektor von make test-race
// (Datenrennen) wie ohne ihn (zwei Append zugleich).
func TestSequenzGleichzeitig(t *testing.T) {
	u := neu(t)
	u.exchanges.pause = true
	const n = 20
	var bodies []io.ReadCloser
	for range n {
		ant, err := u.record(context.Background(), anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadAll(ant.Body); err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, ant.Body)
	}
	los := make(chan struct{})
	var wg sync.WaitGroup
	for _, b := range bodies {
		wg.Go(func() {
			<-los
			_ = b.Close()
		})
	}
	close(los)
	mitZeitgrenze(t, "20 Close", wg.Wait)
	l := u.exchanges.angehaengt()
	if len(l) != n || u.exchanges.hoch != 1 {
		t.Errorf("%d angehängt, höchstens %d Append zugleich", len(l), u.exchanges.hoch)
	}
	for i, e := range l {
		if e.SequenceNo != i+1 {
			t.Errorf("Nummer %d an Stelle %d", e.SequenceNo, i+1)
		}
	}
	if b := u.bericht("s1"); b.NichtAngehaengt != 0 {
		t.Errorf("Bericht %+v", b)
	}
}

// TestAnhaengeUnabhaengig: Anhänge an verschiedene Sessions laufen unabhängig: während Append eines
// Exchange von s1 angehalten ist, nimmt der Dienst einen Exchange von s2 an, und dessen Close kehrt
// angehängt zurück; erst danach endet der von s1 (B19; SPEC-121 Nr. 10, 15). Ein Warten an einer
// Sperre endet an der Zeitgrenze.
func TestAnhaengeUnabhaengig(t *testing.T) {
	u := neu(t)
	u.sessions.setze("s2", "svc-a", model.SessionOpen)
	frei := make(chan struct{})
	u.exchanges.haltAppend, u.exchanges.imAppend, u.exchanges.haltNur = frei, make(chan struct{}, 1), "s1"
	ant, err := u.record(context.Background(), mitSession(anfrage("GET", "/a/x"), sessionHeader, "s1"))
	if err != nil {
		t.Fatal(err)
	}
	s1 := nebenbei(func() error {
		_, _ = io.ReadAll(ant.Body)
		return ant.Body.Close()
	})
	empfange(t, "Append von s1", u.exchanges.imAppend)
	s2 := nebenbei(func() error {
		a2, err := u.r.Record(context.Background(), mitSession(anfrage("GET", "/a/y"), sessionHeader, "s2"))
		if err != nil {
			return err
		}
		_, _ = io.ReadAll(a2.Body)
		return a2.Body.Close()
	})
	if err := empfange(t, "Exchange von s2, während Append von s1 hält,", s2); err != nil {
		t.Fatal(err)
	}
	if l := u.exchanges.angehaengt(); len(l) != 1 || l[0].SessionID != "s2" || l[0].SequenceNo != 1 {
		t.Fatalf("während Append von s1 hält: %d angehängt, erwartet s2 mit 1", len(l))
	}
	close(frei)
	if err := empfange(t, "Close von s1", s1); err != nil {
		t.Fatal(err)
	}
	if l := u.exchanges.angehaengt(); len(l) != 2 || l[1].SessionID != "s1" || l[1].SequenceNo != 1 {
		t.Errorf("nach der Freigabe: %d angehängt, erwartet s1 mit 1 an zweiter Stelle", len(l))
	}
}

// TestNichtAngehaengt: jeder Fehler von Append und vom Lesen der Zahl, auch Validate: Close nil,
// Client las alles, Eintrag ablage mit dem Code zum Gegenstand Sonstiges, Bericht, Zahl und Blob
// bleiben; nach ErrUngueltigeSequenz und nach einem Fehler von ListBySession wird die Zahl neu
// gelesen (B20). Die Session trägt drei vorhandene Exchanges, ListBySession einmal keinen.
func TestNichtAngehaengt(t *testing.T) {
	json := map[string][]string{"Content-Type": {"application/json"}}
	ungueltigerWert := string([]byte{'a', 0xff})
	faelle := []struct {
		name      string
		vorher    func(u *umgebung)
		code      string
		capture   bool
		vorhanden int
	}{
		{"errDisk", func(u *umgebung) { u.exchanges.appendFehler[0] = errDisk }, "ODR-E9001", false, 3},
		{"ErrSessionGeschlossen", func(u *umgebung) { u.exchanges.appendFehler[0] = outbound.ErrSessionGeschlossen }, "ODR-E5009", false, 3},
		{"ErrUngueltigeSequenz", func(u *umgebung) { u.exchanges.appendFehler[0] = outbound.ErrUngueltigeSequenz }, "ODR-E9001", false, 3},
		// Beim Gegenstand Sonstiges tragen beide ODR-E9001, beim Gegenstand Session ODR-E7003 und
		// ODR-E7010 (SPEC-109 Nr. 5).
		{"ErrNichtGefunden", func(u *umgebung) { u.exchanges.appendFehler[0] = outbound.ErrNichtGefunden }, "ODR-E9001", false, 3},
		{"ErrBereitsVorhanden", func(u *umgebung) { u.exchanges.appendFehler[0] = outbound.ErrBereitsVorhanden }, "ODR-E9001", false, 3},
		{"ListBySession", func(u *umgebung) { u.exchanges.listFehler = errDisk }, "ODR-E9001", false, 3},
		{"ListBySession, keiner vorhanden", func(u *umgebung) { u.exchanges.listFehler = errDisk }, "ODR-E9001", false, 0},
		{"Header 0xFF", func(u *umgebung) {
			u.upstream.plane(antwort(200, map[string][]string{"Content-Type": {"application/json"}, "X-Wert": {ungueltigerWert}}, koerper(`{"a":1}`)))
		}, "ODR-E9001", false, 3},
		{"Status 600", func(u *umgebung) { u.upstream.plane(antwort(600, json, koerper(`{"a":1}`))) }, "ODR-E9001", false, 3},
		{"typ und errDisk", func(u *umgebung) {
			u.upstream.plane(antwort(200, map[string][]string{"Content-Type": {"text/plain"}}, koerper(`{"a":1}`)))
			u.exchanges.appendFehler[0] = errDisk
		}, "ODR-E9001", true, 3},
	}
	for _, f := range faelle {
		u := neu(t)
		u.exchanges.vorhanden["s1"] = make([]model.Exchange, f.vorhanden)
		f.vorher(u)
		_, gelesen := u.rundlauf(anfrage("GET", "/a/x"))
		u.exchanges.mu.Lock()
		u.exchanges.listFehler = nil
		u.exchanges.mu.Unlock()
		if gelesen != `{"a":1}` || len(u.exchanges.angehaengt()) != 0 {
			t.Errorf("%s: gelesen %q, %d angehängt", f.name, gelesen, len(u.exchanges.angehaengt()))
		}
		var ablage, capture int
		for _, e := range u.log.alle() {
			switch e.attrs["ereignis"] {
			case "ablage":
				ablage++
				if e.attrs["error_code"] != f.code {
					t.Errorf("%s: error_code %s, erwartet %s", f.name, e.attrs["error_code"], f.code)
				}
			case "capture":
				capture++
			}
		}
		if ablage != 1 || (capture == 1) != f.capture {
			t.Errorf("%s: %d ablage, %d capture", f.name, ablage, capture)
		}
		if f.name == "errDisk" {
			if _, da := u.blobs.inhalt(refVon([]byte(`{"a":1}`))); !da {
				t.Errorf("%s: Blob nicht im Fake", f.name)
			}
		}
		u.rundlauf(anfrage("GET", "/a/x"))
		if l := u.exchanges.angehaengt(); len(l) != 1 || l[0].SequenceNo != f.vorhanden+1 {
			t.Errorf("%s: danach %d angehängt, erwartet einer mit der Nummer %d", f.name, len(l), f.vorhanden+1)
		}
		lesungen := 1
		if f.name == "ErrUngueltigeSequenz" || strings.HasPrefix(f.name, "ListBySession") {
			lesungen = 2
		}
		if u.exchanges.listAufrufe != lesungen {
			t.Errorf("%s: ListBySession %d-mal, erwartet %d", f.name, u.exchanges.listAufrufe, lesungen)
		}
		if b := u.bericht("s1"); b.NichtAngehaengt != 1 {
			t.Errorf("%s: Bericht %+v", f.name, b)
		}
	}
}

// TestAnhaengenNachAbbruch: angehängt wird unter einem Kontext, den weder Client noch Drain beendet (B21).
func TestAnhaengenNachAbbruch(t *testing.T) {
	u := neu(t)
	u.upstream.plane(jsonAntwort(koerper(`{"a":`, `1}`)))
	ctx, abbrechen := context.WithCancel(context.Background())
	ant, err := u.record(ctx, anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	abbrechen()
	schliesse(t, ant.Body)
	if e := u.letzter(); e.Capture != model.CaptureIncomplete || u.exchanges.ctxFehler[0] != nil {
		t.Errorf("Client-Abbruch: %s, Kontext beim Append %v", e.Capture, u.exchanges.ctxFehler)
	}

	d := neu(t)
	d.upstream.plane(jsonAntwort(koerper(`{"a":`, `1}`)))
	ant, err = d.record(context.Background(), anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	dr := d.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Ende des Kontexts", d.upstream.gesehen()[0].done)
	schliesse(t, ant.Body)
	if b := empfange(t, "Drain", dr); b.b.Abgebrochen != 1 {
		t.Errorf("Bericht %+v", b.b)
	}
	if e := d.letzter(); e.Capture != model.CaptureIncomplete || d.exchanges.ctxFehler[0] != nil {
		t.Errorf("Drain: %s, Kontext beim Append %v", e.Capture, d.exchanges.ctxFehler)
	}
}
