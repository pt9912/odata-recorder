package record_test

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// pruefeEintraege hält die Einträge gegen die erwarteten: je Eintrag Stufe ERROR, die festen
// Attribute und genau die zusätzlichen (SPEC-121 Nr. 12).
func pruefeEintraege(t *testing.T, name string, got []eintrag, soll []map[string]string) {
	t.Helper()
	if len(got) != len(soll) {
		t.Errorf("%s: %d Einträge, erwartet %d: %v", name, len(got), len(soll), got)
		return
	}
	for i, e := range got {
		w := map[string]string{"mode": "record", "service_id": "svc-a", "session_id": "s1",
			"exchange_id": kennung("exc", 1), "correlation_id": "cor-1"}
		maps.Copy(w, soll[i])
		if e.stufe != slog.LevelError || !maps.Equal(e.attrs, w) {
			t.Errorf("%s: Eintrag %d %v %v, erwartet %v", name, i, e.stufe, e.attrs, w)
		}
	}
}

// TestProtokoll: je Ereignis ein Eintrag mit genau den Attributen; keiner für Prüfungen aus Nr. 2
// und für einen Client-Abbruch vor dem Kopf (B23).
func TestProtokoll(t *testing.T) {
	u := neu(t)
	ant, err := u.record(context.Background(), mitBody(anfrage("POST", "/a/x"), "text/plain", "x"))
	if err != nil {
		t.Fatal(err)
	}
	schliesse(t, ant.Body)
	pruefeEintraege(t, "capture", u.log.alle(), []map[string]string{
		{"ereignis": "capture", "error_code": "ODR-E2401", "richtung": "request", "grund": "typ"},
		{"ereignis": "capture", "error_code": "ODR-E2401", "richtung": "response", "grund": "abbruch"},
	})

	for grund, f := range map[string]error{
		"nicht_erreichbar": outbound.ErrUpstreamNichtErreichbar, "zeit": outbound.ErrUpstreamZeit, "intern": errDisk,
	} {
		v := neu(t)
		v.upstream.plane(&plan{lies: -1, fehler: f})
		if _, err := v.record(context.Background(), anfrage("GET", "/a/x")); err == nil {
			t.Fatal("kein Fehler")
		}
		code := "ODR-E4001"
		if grund == "intern" {
			code = "ODR-E9001"
		}
		pruefeEintraege(t, grund, v.log.alle(), []map[string]string{{"ereignis": "upstream", "error_code": code, "grund": grund}})
	}
	d := neu(t)
	d.upstream.plane(&plan{lies: -1, warteAufCtx: true})
	res := d.recordNebenbei(context.Background(), anfrage("GET", "/a/x"))
	empfange(t, "Senden", d.upstream.imSenden)
	dr := d.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Record", res)
	empfange(t, "Drain", dr)
	pruefeEintraege(t, "drain", d.log.alle(), []map[string]string{{"ereignis": "upstream", "error_code": "ODR-E4001", "grund": "drain"}})

	a := neu(t)
	a.exchanges.appendFehler[0] = errDisk
	a.upstream.plane(leer())
	a.rundlauf(anfrage("GET", "/a/x"))
	pruefeEintraege(t, "ablage", a.log.alle(), []map[string]string{{"ereignis": "ablage", "error_code": "ODR-E9001"}})

	k := neu(t)
	if _, err := k.record(context.Background(), anfrage("GET", "/c/x")); err == nil {
		t.Fatal("ODR-E5005 erwartet")
	}
	k.sessions.setze("s1", "svc-a", model.SessionComplete)
	if _, err := k.record(context.Background(), anfrage("GET", "/a/x")); err == nil {
		t.Fatal("ODR-E5009 erwartet")
	}
	k.sessions.setze("s1", "svc-a", model.SessionOpen)
	k.upstream.plane(&plan{lies: -1, warteAufCtx: true})
	ctx, abbrechen := context.WithCancel(context.Background())
	res = k.recordNebenbei(ctx, anfrage("GET", "/a/x"))
	empfange(t, "Senden", k.upstream.imSenden)
	abbrechen()
	empfange(t, "Record", res)
	if l := k.log.alle(); len(l) != 0 || !slices.Equal(k.sessions.unerwartet, nil) {
		t.Errorf("Einträge ohne Ereignis: %v", l)
	}
}
