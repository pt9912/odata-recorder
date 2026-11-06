package session_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

const (
	// zeichenKennung sind die 37 Zeichen der Kennung einer Session (SPEC-124 Nr. 4).
	zeichenKennung = "0123456789abcdefghijklmnopqrstuvwxyz-"
	// zeichenDienst sind die 65 Zeichen einer Kennung des Audits (SPEC-234 Nr. 3).
	zeichenDienst = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-"
	geheimnis     = "GEHEIM-7f3a"
)

// geloescht ist der Eintrag, den Delete für die Session id des Service svc schreibt.
func geloescht(svc, id string) support.Eintrag {
	return support.Eintrag{Art: support.ArtSessionGeloescht, ServiceID: svc, SessionID: id}
}

// loescheEine legt die Session id des Service svc an, löscht sie und liefert die Einträge.
func loescheEine(t *testing.T, svc, id string, bericht inbound.Loeschbericht) ([]support.Eintrag, error) {
	t.Helper()
	u := neu(t)
	u.sessions.fuege(id, svc, t0(), model.SessionComplete)
	u.wartung.bericht = bericht
	b, err := u.del(context.Background(), id)
	if b != bericht {
		t.Errorf("Bericht %+v, erwartet %+v", b, bericht)
	}
	return u.audit.Eintraege(), err
}

// TestAuditEintragFelderUndKennungen (E05, E01 der Stelle, M5): bei Erfolg genau ein Eintrag mit Art,
// ServiceID aus Sessions.Get und der id, leerer KorrelationsID, auch wenn CleanupBlobs Blobs entfernt;
// Menge: 37 Zeichen × 2 Längen für die id, 65 × 2 für den Service, 2 Berichte.
func TestAuditEintragFelderUndKennungen(t *testing.T) {
	n := 0
	for _, c := range zeichenKennung {
		for _, l := range []int{1, 64} {
			id := strings.Repeat(string(c), l)
			got, err := loescheEine(t, "svc-a", id, inbound.Loeschbericht{})
			if err != nil || !slices.Equal(got, []support.Eintrag{geloescht("svc-a", id)}) {
				t.Errorf("id %q: %+v, %v", id, got, err)
			}
			n++
		}
	}
	if n != 74 {
		t.Errorf("Kennungen %d, erwartet 74", n)
	}
	n = 0
	for _, c := range zeichenDienst {
		for _, l := range []int{1, 64} {
			svc := strings.Repeat(string(c), l)
			got, err := loescheEine(t, svc, "s1", inbound.Loeschbericht{})
			if err != nil || !slices.Equal(got, []support.Eintrag{geloescht(svc, "s1")}) {
				t.Errorf("Service %q: %+v, %v", svc, got, err)
			}
			n++
		}
	}
	if n != 130 {
		t.Errorf("Services %d, erwartet 130", n)
	}
	for _, entfernt := range []int{0, 5} {
		got, err := loescheEine(t, "Svc.A_b-1", "s1", inbound.Loeschbericht{Entfernt: entfernt})
		if err != nil || !slices.Equal(got, []support.Eintrag{geloescht("Svc.A_b-1", "s1")}) {
			t.Errorf("Entfernt %d: %+v, %v", entfernt, got, err)
		}
	}
}

// TestAuditAblehnungenSchreibenNichts (E06, M2 der Session): bei jeder der sieben Ablehnungen vor der
// Tatsache wird Append nie gerufen.
func TestAuditAblehnungenSchreibenNichts(t *testing.T) {
	var keinCtx context.Context
	faelle := []struct {
		name   string
		setup  func(*welt)
		ctxNil bool
	}{
		{"ctx nil", func(*welt) {}, true},
		{"Get: nicht gefunden", func(u *welt) { u.st.scheitere("Sessions.Get", 1, verpackt(outbound.ErrNichtGefunden)) }, false},
		{"Get: anderer Fehler", func(u *welt) { u.st.scheitere("Sessions.Get", 1, errDisk) }, false},
		{"offen", func(u *welt) { u.sessions.fuege("s1", dienstA, t0(), model.SessionOpen) }, false},
		{"referenziert", func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrReferenziert)) }, false},
		{"Delete: nicht gefunden", func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrNichtGefunden)) }, false},
		{"Delete: anderer Fehler", func(u *welt) { u.st.scheitere("Sessions.Delete", 1, errDisk) }, false},
	}
	for _, f := range faelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		f.setup(u)
		ctx := context.Background()
		if f.ctxNil {
			ctx = keinCtx
		}
		if _, err := u.del(ctx, "s1"); err == nil {
			t.Errorf("%s: kein Fehler", f.name)
		}
		if n := u.audit.Aufrufe(); n != 0 {
			t.Errorf("%s: %d Aufrufe des Audits", f.name, n)
		}
	}
	if len(faelle) != 7 {
		t.Errorf("Fälle %d, erwartet 7", len(faelle))
	}
}

// TestAuditReihenfolge (E10, F2): Sessions.Delete, dann Append, dann CleanupBlobs; im Aufruf von
// CleanupBlobs steht der Eintrag schon.
func TestAuditReihenfolge(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	u.audit.Vor = func(support.Eintrag) { u.p.melde("Audit.Append") }
	var stand []support.Eintrag
	u.st.bei("CleanupBlobs", 1, func() { stand = u.audit.Eintraege() })
	if _, err := u.del(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	soll := []string{"Sessions.Get(s1)", "Sessions.Delete(s1)", "Audit.Append", "CleanupBlobs"}
	if got := u.p.liste(); !slices.Equal(got, soll) {
		t.Errorf("Folge %v, erwartet %v", got, soll)
	}
	if !slices.Equal(stand, []support.Eintrag{geloescht(dienstA, "s1")}) {
		t.Errorf("im Aufruf von CleanupBlobs: %+v", stand)
	}
}

// TestAuditFehlerAendertNichts (E12, E13, M1 der Session): ein Fehler von Append und ein im Port nach
// dem Commit beendeter ctx ändern Rückgabe und Bereinigung nicht, der Eintrag steht; Menge: Fehler
// (gesetzt, ungesetzt) × ctx (aktiv, im Port beendet) = 4.
func TestAuditFehlerAendertNichts(t *testing.T) {
	n := 0
	var ref [2]inbound.Loeschbericht
	for ci, beendet := range []bool{false, true} {
		for _, fehler := range []error{nil, errors.New("Ablage " + marker)} {
			u := neu(t)
			u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
			u.audit.Fehler = fehler
			ctx, cancel := context.WithCancel(context.Background())
			if beendet {
				u.sessions.nachDelete = cancel
			}
			b, err := u.del(ctx, "s1")
			cancel()
			if err != nil {
				t.Errorf("beendet %v, Fehler %v: Fehler %v", beendet, fehler, err)
			}
			if fehler == nil {
				ref[ci] = b
			} else if b != ref[ci] {
				t.Errorf("beendet %v: Bericht %+v mit Fehler, %+v ohne", beendet, b, ref[ci])
			}
			if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloescht(dienstA, "s1")}) {
				t.Errorf("beendet %v, Fehler %v: Einträge %+v", beendet, fehler, got)
			}
			if c := u.st.anzahl("CleanupBlobs"); c != 1 {
				t.Errorf("beendet %v, Fehler %v: CleanupBlobs %d-mal", beendet, fehler, c)
			}
			n++
		}
	}
	if n != 4 || ref[0].Entfernt != 2 {
		t.Errorf("Menge %d, Bericht %+v", n, ref)
	}
}

// TestAuditWiederholung (E15): ein zweites Delete derselben Session ist ODR-E7003 und schreibt nichts.
func TestAuditWiederholung(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	if _, err := u.del(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := u.del(context.Background(), "s1"); err == nil {
		t.Fatal("zweites Delete ohne Fehler")
	}
	if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloescht(dienstA, "s1")}) {
		t.Errorf("Einträge %+v", got)
	}
}

// TestAuditGleichzeitigesDelete (E16): zwei gleichzeitige Delete derselben Session schreiben genau einen
// Eintrag; die Barriere in Sessions.Get hält beide, bis beide gelesen haben, erst danach beginnt ein
// Delete. 32 Paare.
func TestAuditGleichzeitigesDelete(t *testing.T) {
	for p := range 32 {
		u := neu(t)
		id := fmt.Sprintf("s-%d", p)
		u.sessions.fuege(id, dienstA, t0(), model.SessionComplete)
		b := neueBarriere(2)
		u.sessions.nachGet = b.an
		var wg sync.WaitGroup
		var mu sync.Mutex
		gelungen := 0
		for range 2 {
			wg.Go(func() {
				if _, err := u.del(context.Background(), id); err == nil {
					mu.Lock()
					gelungen++
					mu.Unlock()
				}
			})
		}
		wg.Wait()
		if b.lief() {
			t.Fatalf("Paar %d: die Barriere lief ab", p)
		}
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloescht(dienstA, id)}) || gelungen != 1 {
			t.Errorf("Paar %d: Einträge %+v, gelungen %d", p, got, gelungen)
		}
	}
}

// TestAuditKonfig (E17): ein nil-Audit ist ErrUngueltig ohne Dienst; New ruft Append nie.
func TestAuditKonfig(t *testing.T) {
	u := neu(t)
	k := u.konfig
	k.Audit = nil
	if v, err := session.New(k); !errors.Is(err, model.ErrUngueltig) || v != nil || !ohneCode(err) {
		t.Errorf("Audit nil: %v, Dienst %v", err, v != nil)
	}
	if _, err := session.New(u.konfig); err != nil {
		t.Fatal(err)
	}
	if n := u.audit.Aufrufe(); n != 0 {
		t.Errorf("New rief den Audit-Port %d-mal", n)
	}
	if f, ok := reflect.TypeOf(session.Konfig{}).FieldByName("Audit"); !ok || f.Type != reflect.TypeOf((*support.AuditLog)(nil)).Elem() {
		t.Errorf("Feld Audit: %v", f)
	}
}

// TestAuditKeinGeheimnis (E18, M6 der Session): kein Feld des Eintrags trägt Bytes aus dem Fehlertext der
// Ports oder aus den Feldern des Service; die sieben Felder werden einzeln durchsucht.
func TestAuditKeinGeheimnis(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	u.dienste.dienste = []model.Service{{ID: dienstA, ServiceRoot: "https://" + geheimnis + "@host/", MetadataBlob: geheimnis}}
	u.wartung.fehler = errors.New("Wartung " + geheimnis)
	if _, err := u.del(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
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

// TestAuditSynchronOhneFrist (E19, F1, M1 der Session): Append läuft auf der Goroutine des Dienstes, der
// Dienst setzt erst nach seiner Rückkehr fort, auch bei im Halt beendetem ctx; im Halt steht die
// Löschung, der Eintrag fehlt, CleanupBlobs ist nicht gerufen, und Delete ist nach 100 ms nicht
// zurückgekehrt. Nach der Freigabe steht der Eintrag, das Ergebnis ist das des Laufs ohne Halt.
func TestAuditSynchronOhneFrist(t *testing.T) {
	type ergebnis struct {
		b   inbound.Loeschbericht
		err error
	}
	n := 0
	for _, beendet := range []bool{false, true} {
		n++
		ref := neu(t)
		ref.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		refCtx, refCancel := context.WithCancel(context.Background())
		ref.audit.Vor = func(support.Eintrag) {
			if beendet {
				refCancel()
			}
		}
		rb, rerr := ref.del(refCtx, "s1")
		refCancel()

		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		ctx, cancel := context.WithCancel(context.Background())
		erreicht, frei := make(chan struct{}), make(chan struct{})
		var stand []support.Eintrag
		var folge []string
		u.audit.Vor = func(support.Eintrag) {
			if beendet {
				cancel()
			}
			stand, folge = u.audit.Eintraege(), u.p.liste()
			close(erreicht)
			<-frei
		}
		fertig := make(chan ergebnis, 1)
		go func() {
			b, err := u.del(ctx, "s1")
			fertig <- ergebnis{b, err}
		}()
		warte(t, erreicht, "Append erreicht")
		select {
		case <-fertig:
			t.Errorf("beendet %v: Delete kehrte im Halt von Append zurück", beendet)
		case <-time.After(100 * time.Millisecond):
		}
		soll := []string{"Sessions.Get(s1)", "Sessions.Delete(s1)"}
		if !slices.Equal(folge, soll) || len(stand) != 0 {
			t.Errorf("beendet %v: im Halt Folge %v, Einträge %+v", beendet, folge, stand)
		}
		if got := u.p.liste(); !slices.Equal(got, soll) {
			t.Errorf("beendet %v: im Halt tat der Dienst weiter: %v", beendet, got)
		}
		close(frei)
		select {
		case r := <-fertig:
			if r.b != rb || (r.err == nil) != (rerr == nil) {
				t.Errorf("beendet %v: Ergebnis %+v, %v; ohne Halt %+v, %v", beendet, r.b, r.err, rb, rerr)
			}
		case <-time.After(zeitgrenze):
			t.Fatalf("beendet %v: Delete kehrte nach der Freigabe nicht zurück", beendet)
		}
		cancel()
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloescht(dienstA, "s1")}) {
			t.Errorf("beendet %v: Einträge %+v", beendet, got)
		}
		if got := u.p.liste(); !slices.Equal(got, append(slices.Clone(soll), "CleanupBlobs")) {
			t.Errorf("beendet %v: Folge %v", beendet, got)
		}
	}
	if n != 2 {
		t.Errorf("Menge %d, erwartet 2", n)
	}
}
