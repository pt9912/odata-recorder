package maintenance_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
)

const (
	// zeichenKennung sind die 37 Zeichen einer Kennung der Session (SPEC-124 Nr. 4).
	zeichenKennung = "0123456789abcdefghijklmnopqrstuvwxyz-"
	// zeichenDienst sind die 65 Zeichen einer Kennung des Audits (SPEC-234 Nr. 3).
	zeichenDienst = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-"
	geheimnis     = "GEHEIM-7f3a"
)

// aufbewahrt ist der Eintrag, den die Aufbewahrung für die Session id des Service svc schreibt.
func aufbewahrt(svc, id string) support.Eintrag {
	return support.Eintrag{Art: support.ArtSessionAufbewahrung, ServiceID: svc, SessionID: id}
}

// alteSessions legt drei Sessions a, b, c des Service svc an, älter als MaxAge, in dieser Folge.
func alteSessions(u *welt, svc string) []string {
	u.dienst(svc)
	ids := []string{"a", "b", "c"}
	for i, id := range ids {
		u.sessions.fuege(id, svc, t0().Add(-time.Duration(40-i)*tag), model.SessionComplete)
	}
	return ids
}

// TestAuditFelderUndKennungen (E05, E01 der Stelle, M5): je gelöschter Session genau ein Eintrag mit
// Art, Service und Kennung der Session und leerer KorrelationsID; ein Blob, den die Stufe 4 entfernt,
// schreibt keinen. Menge: 37 Zeichen × 2 Längen für die Session, 65 × 2 für den Service.
func TestAuditFelderUndKennungen(t *testing.T) {
	lauf := func(svc, id string) {
		u := neu(t, mitMaxAge(30*tag))
		u.dienst(svc)
		u.sessions.fuege(id, svc, t0().Add(-40*tag), model.SessionComplete)
		u.blobs.lege(r("a"), r("b"), r("c"))
		b, err := u.startup(context.Background())
		if err != nil || b.SessionsGeloescht != 1 || b.BlobsEntfernt == 0 {
			t.Fatalf("Service %q, Session %q: %v %+v", svc, id, err, b)
		}
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{aufbewahrt(svc, id)}) {
			t.Errorf("Service %q, Session %q: %+v", svc, id, got)
		}
	}
	n := 0
	for _, c := range zeichenKennung {
		for _, l := range []int{1, 64} {
			lauf("svc-a", strings.Repeat(string(c), l))
			n++
		}
	}
	if n != 74 {
		t.Errorf("Kennungen %d, erwartet 74", n)
	}
	n = 0
	for _, c := range zeichenDienst {
		for _, l := range []int{1, 64} {
			lauf(strings.Repeat(string(c), l), "s1")
			n++
		}
	}
	if n != 130 {
		t.Errorf("Services %d, erwartet 130", n)
	}
	u := neu(t, mitMaxAge(30*tag))
	u.blobs.lege(r("a"))
	if c, err := u.cleanup(context.Background()); err != nil || c.Entfernt != 1 || u.audit.Aufrufe() != 0 {
		t.Errorf("CleanupBlobs: %v %+v, Aufrufe des Audits %d", err, c, u.audit.Aufrufe())
	}
}

// TestAuditPositionUndErgebnis (E07, E11, M2 und M7 der Aufbewahrung): drei Sessions, das Ergebnis von
// Delete an der ersten, mittleren oder letzten Position (nil, ErrReferenziert, ErrNichtGefunden, anderer
// Fehler; die anderen nil) = 12 Läufe; ein Eintrag je gelungenem Delete in der Folge der Sessions.
func TestAuditPositionUndErgebnis(t *testing.T) {
	ergebnisse := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"ErrReferenziert", outbound.ErrReferenziert},
		{"ErrNichtGefunden", outbound.ErrNichtGefunden},
		{"anderer Fehler", errDisk},
	}
	n := 0
	for pos := 1; pos <= 3; pos++ {
		for _, e := range ergebnisse {
			n++
			u := neu(t, mitMaxAge(30*tag))
			ids := alteSessions(u, dienstA)
			if e.err != nil {
				u.st.scheitere("Delete", pos, e.err)
			}
			b, err := u.startup(context.Background())
			var soll []support.Eintrag
			for i, id := range ids {
				if i+1 == pos && e.err != nil {
					if !errors.Is(e.err, outbound.ErrReferenziert) && !errors.Is(e.err, outbound.ErrNichtGefunden) {
						break
					}
					continue
				}
				soll = append(soll, aufbewahrt(dienstA, id))
			}
			if got := u.audit.Eintraege(); !slices.Equal(got, soll) {
				t.Errorf("Position %d, %s: Einträge %+v, erwartet %+v", pos, e.name, got, soll)
			}
			if e.name == "anderer Fehler" {
				if err == nil {
					t.Errorf("Position %d: kein Fehler", pos)
				}
			} else if err != nil || b.SessionsGeloescht != len(soll) {
				t.Errorf("Position %d, %s: %v, gelöscht %d, Einträge %d", pos, e.name, err, b.SessionsGeloescht, len(soll))
			}
		}
	}
	if n != 12 {
		t.Errorf("Läufe %d, erwartet 12", n)
	}
	// MaxAge 0 und ohne Session: kein Aufruf des Audits.
	u := neu(t)
	alteSessions(u, dienstA)
	if _, err := u.startup(context.Background()); err != nil || u.audit.Aufrufe() != 0 {
		t.Errorf("MaxAge 0: %v, Aufrufe %d", err, u.audit.Aufrufe())
	}
	v := neu(t, mitMaxAge(30*tag))
	v.dienst(dienstA)
	if _, err := v.startup(context.Background()); err != nil || v.audit.Aufrufe() != 0 {
		t.Errorf("ohne Session: %v, Aufrufe %d", err, v.audit.Aufrufe())
	}
	// M2, Elemente 6 bis 8: eine Session innerhalb der Frist, ein Fehler von Dienste.List und einer von
	// Sessions.List (SPEC-234 Nr. 4: kein Eintrag vor der Tatsache).
	j := neu(t, mitMaxAge(30*tag))
	alteSessions(j, dienstA)
	j.sessions.fuege("jung", dienstA, t0().Add(-tag), model.SessionComplete)
	if _, err := j.startup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := j.audit.Eintraege(); len(got) != 3 {
		t.Errorf("junge Session: Einträge %+v, erwartet 3 (die alten)", got)
	}
	if _, da := j.sessions.status("jung"); !da {
		t.Error("junge Session gelöscht")
	}
	for _, op := range []string{"Dienste.List", "Sessions.List"} {
		x := neu(t, mitMaxAge(30*tag))
		alteSessions(x, dienstA)
		x.st.scheitere(op, 1, errDisk)
		if _, err := x.startup(context.Background()); err == nil || x.audit.Aufrufe() != 0 {
			t.Errorf("%s: %v, Aufrufe des Audits %d", op, err, x.audit.Aufrufe())
		}
	}
}

// TestAuditVorDemNaechstenPort (E11, F3): im Delete der zweiten Session stehen die Einträge der ersten
// schon; scheitert die zweite mit einem anderen Fehler, bleibt der Eintrag der ersten.
func TestAuditVorDemNaechstenPort(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	ids := alteSessions(u, dienstA)
	var stand [3][]support.Eintrag
	for i := 1; i <= 3; i++ {
		u.st.bei("Delete", i, func() { stand[i-1] = u.audit.Eintraege() })
	}
	u.st.scheitere("Delete", 2, errDisk)
	u.audit.Vor = func(support.Eintrag) { u.p.melde("Audit.Append") }
	if _, err := u.startup(context.Background()); err == nil {
		t.Fatal("kein Fehler")
	}
	soll := []string{"Delete(a)", "Audit.Append", "Delete(b)"}
	var folge []string
	for _, ev := range u.p.liste() {
		if strings.HasPrefix(ev, "Delete(") || ev == "Audit.Append" {
			folge = append(folge, ev)
		}
	}
	if !slices.Equal(folge, soll) {
		t.Errorf("Folge %v, erwartet %v", folge, soll)
	}
	if len(stand[0]) != 0 || !slices.Equal(stand[1], []support.Eintrag{aufbewahrt(dienstA, ids[0])}) || stand[2] != nil {
		t.Errorf("Stand in Delete: %+v", stand)
	}
	if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{aufbewahrt(dienstA, "a")}) {
		t.Errorf("Einträge %+v", got)
	}
}

// TestAuditFehlerAendertNichts (E12, E13, M1): ein Fehler von Append und ein im Port nach dem Commit
// beendeter ctx ändern Bericht und Fehler nicht, und die Stufe schreibt den Eintrag der nächsten
// Session; Menge: Fehler (gesetzt, ungesetzt) × ctx (aktiv, im Port beendet) = 4.
func TestAuditFehlerAendertNichts(t *testing.T) {
	n := 0
	for _, beendet := range []bool{false, true} {
		type lauf struct {
			err  error
			b    string
			eint []support.Eintrag
		}
		var ref lauf
		for _, fehler := range []error{nil, errors.New("Ablage " + marker)} {
			n++
			u := neu(t, mitMaxAge(30*tag))
			alteSessions(u, dienstA)
			u.audit.Fehler = fehler
			ctx, cancel := context.WithCancel(context.Background())
			if beendet {
				u.sessions.nachDelete = func(id string) {
					if id == "a" {
						cancel()
					}
				}
			}
			b, err := u.startup(ctx)
			cancel()
			got := lauf{err, fmt.Sprintf("%+v", b), u.audit.Eintraege()}
			if fehler == nil {
				ref = got
			} else if !reflect.DeepEqual(got.b, ref.b) || errors.Is(got.err, context.Canceled) != errors.Is(ref.err, context.Canceled) ||
				(got.err == nil) != (ref.err == nil) || !slices.Equal(got.eint, ref.eint) {
				t.Errorf("beendet %v: mit Fehler %+v, ohne %+v", beendet, got, ref)
			}
			soll := []support.Eintrag{aufbewahrt(dienstA, "a")}
			if !beendet {
				soll = append(soll, aufbewahrt(dienstA, "b"), aufbewahrt(dienstA, "c"))
			}
			if !slices.Equal(got.eint, soll) {
				t.Errorf("beendet %v, Fehler %v: Einträge %+v, erwartet %+v", beendet, fehler, got.eint, soll)
			}
			if beendet != errors.Is(err, context.Canceled) {
				t.Errorf("beendet %v, Fehler %v: Fehler %v", beendet, fehler, err)
			}
		}
	}
	if n != 4 {
		t.Errorf("Menge %d, erwartet 4", n)
	}
}

// TestAuditWiederholung (E15): ein zweites Startup ohne neue Daten schreibt nichts.
func TestAuditWiederholung(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	alteSessions(u, dienstA)
	if _, err := u.startup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := u.startup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := u.audit.Eintraege(); len(got) != 3 {
		t.Errorf("Einträge %+v, erwartet 3", got)
	}
}

// TestAuditKonfig (E17): ein nil-Audit ist ErrUngueltig ohne Wartung; New ruft Append nie.
func TestAuditKonfig(t *testing.T) {
	u := neu(t)
	k := u.konfig
	k.Audit = nil
	if w, err := maintenance.New(k); !errors.Is(err, model.ErrUngueltig) || w != nil {
		t.Errorf("Audit nil: %v, Wartung %v", err, w != nil)
	}
	if u.audit.Aufrufe() != 0 {
		t.Errorf("New rief den Audit-Port %d-mal", u.audit.Aufrufe())
	}
	if f, ok := reflect.TypeOf(maintenance.Konfig{}).FieldByName("Audit"); !ok || f.Type != reflect.TypeOf((*support.AuditLog)(nil)).Elem() {
		t.Errorf("Feld Audit: %v", f)
	}
}

// TestAuditKeinGeheimnis (E18, M6 der Aufbewahrung): kein Feld des Eintrags trägt Bytes aus dem
// Fehlertext der Ports oder aus den Feldern des Service; die sieben Felder werden einzeln durchsucht.
func TestAuditKeinGeheimnis(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	u.dienste.dienste = []model.Service{{ID: dienstA, ServiceRoot: "https://" + geheimnis + "@host/", MetadataBlob: geheimnis}}
	u.sessions.fuege("a", dienstA, t0().Add(-40*tag), model.SessionComplete)
	u.sessions.fuege("b", dienstA, t0().Add(-39*tag), model.SessionComplete)
	u.st.scheitere("Delete", 2, errors.New("Delete "+geheimnis))
	u.audit.Fehler = errors.New("Ablage " + geheimnis)
	_, _ = u.startup(context.Background())
	got := u.audit.Eintraege()
	for _, e := range got {
		v := reflect.ValueOf(e)
		if v.NumField() != 7 {
			t.Fatalf("Felder %d", v.NumField())
		}
		for i := range v.NumField() {
			if s := fmt.Sprint(v.Field(i).Interface()); strings.Contains(s, geheimnis) {
				t.Errorf("Feld %s trägt das Geheimnis: %q", v.Type().Field(i).Name, s)
			}
		}
	}
	if len(got) != 1 {
		t.Errorf("Einträge %+v", got)
	}
}

// TestAuditSynchronOhneFrist (E19, F1, M1 der Aufbewahrung): Append der Session an jeder der drei Positionen
// läuft auf der Goroutine des Dienstes, Startup setzt erst nach seiner Rückkehr fort, auch bei im Halt
// beendetem ctx (6 Läufe); im Halt steht die Löschung, der Eintrag fehlt, das nächste Delete ist nicht gerufen, und Startup
// ist nach 100 ms nicht zurückgekehrt. Nach der Freigabe ist das Ergebnis das des Laufs ohne Halt.
func TestAuditSynchronOhneFrist(t *testing.T) {
	n := 0
	for _, kombi := range [][2]int{{1, 0}, {1, 1}, {2, 0}, {2, 1}, {3, 0}, {3, 1}} {
		pos, beendet := kombi[0], kombi[1] == 1
		n++
		// lauf baut die Welt; im Aufruf von Append der zweiten Session endet der ctx (wenn beendet) und
		// laeuft vor, falls gesetzt.
		lauf := func(vor func()) (*welt, context.Context, context.CancelFunc) {
			u := neu(t, mitMaxAge(30*tag))
			alteSessions(u, dienstA)
			ctx, cancel := context.WithCancel(context.Background())
			zaehler := 0
			u.audit.Vor = func(support.Eintrag) {
				zaehler++
				if zaehler != pos {
					return
				}
				if beendet {
					cancel()
				}
				if vor != nil {
					vor()
				}
			}
			return u, ctx, cancel
		}
		ref, refCtx, refCancel := lauf(nil)
		rb, rerr := ref.startup(refCtx)
		refCancel()

		erreicht, frei := make(chan struct{}), make(chan struct{})
		var u *welt
		var stand []support.Eintrag
		var folge []string
		u, ctx, cancel := lauf(func() {
			stand, folge = u.audit.Eintraege(), u.p.liste()
			close(erreicht)
			<-frei
		})
		type ergebnis struct {
			b   string
			err error
		}
		fertig := make(chan ergebnis, 1)
		go func() {
			b, err := u.startup(ctx)
			fertig <- ergebnis{fmt.Sprintf("%+v", b), err}
		}()
		warte(t, erreicht, "Append der Session erreicht")
		ids := []string{"a", "b", "c"}
		select {
		case <-fertig:
			t.Errorf("beendet %v: Startup kehrte im Halt von Append zurück", beendet)
		case <-time.After(100 * time.Millisecond):
		}
		var solld []string
		var solle []support.Eintrag
		for i, id := range ids[:pos] {
			solld = append(solld, "Delete("+id+")")
			if i < pos-1 {
				solle = append(solle, aufbewahrt(dienstA, id))
			}
		}
		if !slices.Equal(u.p.mit("Delete("), solld) || !slices.Equal(stand, solle) {
			t.Errorf("Position %d, beendet %v: im Halt Delete %v, Einträge %+v", pos, beendet, u.p.mit("Delete("), stand)
		}
		if len(u.p.liste()) != len(folge) {
			t.Errorf("beendet %v: im Halt tat der Dienst weiter: %v nach %v", beendet, u.p.liste(), folge)
		}
		close(frei)
		select {
		case r := <-fertig:
			if r.b != fmt.Sprintf("%+v", rb) || (r.err == nil) != (rerr == nil) || errors.Is(r.err, context.Canceled) != errors.Is(rerr, context.Canceled) {
				t.Errorf("beendet %v: Ergebnis %v, %v; ohne Halt %+v, %v", beendet, r.b, r.err, rb, rerr)
			}
		case <-time.After(zeitgrenze):
			t.Fatalf("beendet %v: Startup kehrte nach der Freigabe nicht zurück", beendet)
		}
		cancel()
		if got, soll := u.audit.Eintraege(), ref.audit.Eintraege(); !slices.Equal(got, soll) {
			t.Errorf("beendet %v: Einträge %+v, ohne Halt %+v", beendet, got, soll)
		}
	}
	if n != 6 {
		t.Errorf("Menge %d, erwartet 6", n)
	}
}
