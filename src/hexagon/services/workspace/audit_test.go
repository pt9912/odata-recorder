package workspace_test

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
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

const (
	// zeichenKennung sind die 37 Zeichen einer Kennung des Workspace (SPEC-129 Nr. 3).
	zeichenKennung = "0123456789abcdefghijklmnopqrstuvwxyz-"
	geheimnis      = "GEHEIM-7f3a"
	zeitgrenze     = 5 * time.Second
)

// geloeschtWS ist der Eintrag, den Delete für den Workspace id schreibt.
func geloeschtWS(id string) support.Eintrag {
	return support.Eintrag{Art: support.ArtWorkspaceGeloescht, WorkspaceID: id}
}

// legeWorkspaceAn legt den Workspace id im Store ab und leert das Protokoll.
func legeWorkspaceAn(w *welt, id string) {
	z := w.st.stand()
	z.ws[id] = model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", CreatedAt: t0()}
	w.st.setze(z)
	w.p.leeren()
}

// TestAuditEintragUndKennungen (E05, E01 der Stelle, M5): bei Erfolg genau ein Eintrag mit Art und der
// id, ohne Service, mit leerer KorrelationsID, auch wenn CleanupBlobs Blobs entfernt. Menge: 37 Zeichen ×
// 2 Längen, 2 Berichte.
func TestAuditEintragUndKennungen(t *testing.T) {
	n := 0
	for _, c := range zeichenKennung {
		for _, l := range []int{1, 64} {
			id := strings.Repeat(string(c), l)
			w := neu(t)
			legeWorkspaceAn(w, id)
			if _, err := w.v.Delete(context.Background(), id); err != nil {
				t.Fatalf("id %q: %v", id, err)
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloeschtWS(id)}) {
				t.Errorf("id %q: %+v", id, got)
			}
			n++
		}
	}
	if n != 74 {
		t.Errorf("Kennungen %d, erwartet 74", n)
	}
	for _, entfernt := range []int{0, 5} {
		w := neu(t)
		legeWorkspaceAn(w, "ws-1")
		w.wartung.setze(inbound.Loeschbericht{Entfernt: entfernt}, nil)
		b, err := w.v.Delete(context.Background(), "ws-1")
		if err != nil || b.Entfernt != entfernt || !slices.Equal(w.audit.Eintraege(), []support.Eintrag{geloeschtWS("ws-1")}) {
			t.Errorf("Entfernt %d: %+v, %v", entfernt, b, err)
		}
	}
}

// TestAuditAblehnungenSchreibenNichts (E08, M2 des Workspace): bei ctx nil, ODR-E7001 und jedem Fehler
// des Ports wird Append nie gerufen; die Menge der Fehler ist die der Tabelle (tabellenFehlerListe).
func TestAuditAblehnungenSchreibenNichts(t *testing.T) {
	var keinCtx context.Context
	w := neu(t)
	if _, err := w.v.Delete(keinCtx, "ws-1"); err == nil || w.audit.Aufrufe() != 0 {
		t.Errorf("ctx nil: %v, Aufrufe %d", err, w.audit.Aufrufe())
	}
	w = neu(t)
	if _, err := w.v.Delete(context.Background(), "unbekannt"); err == nil || w.audit.Aufrufe() != 0 {
		t.Errorf("ODR-E7001: %v, Aufrufe %d", err, w.audit.Aufrufe())
	}
	liste := tabellenFehlerListe()
	if len(liste) == 0 {
		t.Fatal("keine Fehler der Tabelle")
	}
	for _, fe := range liste {
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.p.setzeFehler("Workspaces.Delete", fe.err)
		if _, err := w.v.Delete(context.Background(), "ws-1"); err == nil || w.audit.Aufrufe() != 0 {
			t.Errorf("%s: %v, Aufrufe %d", fe.name, err, w.audit.Aufrufe())
		}
	}
}

// TestAuditReihenfolge (E10, F2): Workspaces.Delete, dann Append, dann CleanupBlobs; im Aufruf von
// CleanupBlobs steht der Eintrag schon.
func TestAuditReihenfolge(t *testing.T) {
	w := neu(t)
	legeWorkspaceAn(w, "ws-1")
	w.audit.Vor = func(support.Eintrag) { _ = w.p.beginn("Audit.Append", nil) }
	var stand []support.Eintrag
	w.wartung.beiCleanup = func() { stand = w.audit.Eintraege() }
	if _, err := w.v.Delete(context.Background(), "ws-1"); err != nil {
		t.Fatal(err)
	}
	soll := []string{"Workspaces.Delete", "Audit.Append", "Wartung.CleanupBlobs"}
	if got := w.p.ops(); !slices.Equal(got, soll) {
		t.Errorf("Folge %v, erwartet %v", got, soll)
	}
	if !slices.Equal(stand, []support.Eintrag{geloeschtWS("ws-1")}) {
		t.Errorf("im Aufruf von CleanupBlobs: %+v", stand)
	}
}

// TestAuditFehlerAendertNichts (E12, E13, M1 des Workspace): ein Fehler von Append und ein im Port nach
// dem Commit beendeter ctx ändern Rückgabe und Bereinigung nicht, der Eintrag steht; Menge: Fehler
// (gesetzt, ungesetzt) × ctx (aktiv, im Port beendet) = 4.
func TestAuditFehlerAendertNichts(t *testing.T) {
	n := 0
	for _, beendet := range []bool{false, true} {
		for _, fehler := range []error{nil, errors.New("Ablage " + geheimnis)} {
			n++
			w := neu(t)
			legeWorkspaceAn(w, "ws-1")
			w.wartung.setze(inbound.Loeschbericht{Entfernt: 4}, nil)
			w.audit.Fehler = fehler
			ctx, cancel := context.WithCancel(context.Background())
			if beendet {
				w.p.nach = func(op string) {
					if op == "Workspaces.Delete" {
						cancel()
					}
				}
			}
			b, err := w.v.Delete(ctx, "ws-1")
			cancel()
			// Bei beendetem ctx liefert der Fake der Wartung den Kontextfehler mit dem Bericht; der Dienst
			// gibt ihn unverändert zurück und den Fehler nicht.
			if err != nil || b.Entfernt != 4 {
				t.Errorf("beendet %v, Fehler %v: %+v, %v", beendet, fehler, b, err)
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloeschtWS("ws-1")}) {
				t.Errorf("beendet %v, Fehler %v: Einträge %+v", beendet, fehler, got)
			}
			if c := w.p.anzahl("Wartung.CleanupBlobs"); c != 1 {
				t.Errorf("beendet %v, Fehler %v: CleanupBlobs %d-mal", beendet, fehler, c)
			}
		}
	}
	if n != 4 {
		t.Errorf("Menge %d, erwartet 4", n)
	}
}

// TestAuditWiederholung (E15): ein zweites Delete desselben Workspace ist ODR-E7001 und schreibt nichts.
func TestAuditWiederholung(t *testing.T) {
	w := neu(t)
	legeWorkspaceAn(w, "ws-1")
	if _, err := w.v.Delete(context.Background(), "ws-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.v.Delete(context.Background(), "ws-1"); err == nil {
		t.Fatal("zweites Delete ohne Fehler")
	}
	if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloeschtWS("ws-1")}) {
		t.Errorf("Einträge %+v", got)
	}
}

// TestAuditGleichzeitigesDelete (E16): zwei gleichzeitige Delete desselben Workspace schreiben genau einen
// Eintrag (der Fake des Ports löscht atomar). 32 Paare.
func TestAuditGleichzeitigesDelete(t *testing.T) {
	for p := range 32 {
		w := neu(t)
		id := fmt.Sprintf("ws-p%d", p)
		legeWorkspaceAn(w, id)
		var wg sync.WaitGroup
		var mu sync.Mutex
		gelungen := 0
		for range 2 {
			wg.Go(func() {
				if _, err := w.v.Delete(context.Background(), id); err == nil {
					mu.Lock()
					gelungen++
					mu.Unlock()
				}
			})
		}
		wg.Wait()
		if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloeschtWS(id)}) || gelungen != 1 {
			t.Errorf("Paar %d: Einträge %+v, gelungen %d", p, got, gelungen)
		}
	}
}

// TestAuditKonfig (E17): New ruft Append nie; das Feld Audit hat den Typ des Ports (nil ist in
// TestNewNilFeld abgedeckt).
func TestAuditKonfig(t *testing.T) {
	w := neu(t)
	if w.audit.Aufrufe() != 0 {
		t.Errorf("New rief den Audit-Port %d-mal", w.audit.Aufrufe())
	}
	if f, ok := reflect.TypeOf(w.konfig).FieldByName("Audit"); !ok || f.Type != reflect.TypeOf((*support.AuditLog)(nil)).Elem() {
		t.Errorf("Feld Audit: %v", f)
	}
}

// TestAuditKeinGeheimnis (E18, M6 des Workspace): kein Feld des Eintrags trägt Bytes aus dem Fehlertext
// der Ports; die sieben Felder werden einzeln durchsucht.
func TestAuditKeinGeheimnis(t *testing.T) {
	w := neu(t)
	legeWorkspaceAn(w, "ws-1")
	w.wartung.setze(inbound.Loeschbericht{Entfernt: 1}, errors.New("Wartung "+geheimnis))
	w.audit.Fehler = errors.New("Ablage " + geheimnis)
	if _, err := w.v.Delete(context.Background(), "ws-1"); err != nil {
		t.Fatal(err)
	}
	got := w.audit.Eintraege()
	if len(got) != 1 {
		t.Fatalf("Einträge %+v", got)
	}
	v := reflect.ValueOf(got[0])
	if v.NumField() != 7 {
		t.Fatalf("Felder %d", v.NumField())
	}
	for i := range v.NumField() {
		if s := fmt.Sprint(v.Field(i).Interface()); strings.Contains(s, geheimnis) {
			t.Errorf("Feld %s trägt das Geheimnis: %q", v.Type().Field(i).Name, s)
		}
	}
}

// TestAuditSynchronOhneFrist (E19, F1, M1 des Workspace): Append läuft auf der Goroutine des Dienstes, der
// Dienst setzt erst nach seiner Rückkehr fort, auch bei im Halt beendetem ctx; im Halt steht die Löschung,
// der Eintrag fehlt, CleanupBlobs ist nicht gerufen, und Delete ist nach 100 ms nicht zurückgekehrt.
// Nach der Freigabe steht der Eintrag, das Ergebnis ist das des Laufs ohne Halt.
func TestAuditSynchronOhneFrist(t *testing.T) {
	type ergebnis struct {
		b   inbound.Loeschbericht
		err error
	}
	n := 0
	for _, beendet := range []bool{false, true} {
		n++
		ref := neu(t)
		legeWorkspaceAn(ref, "ws-1")
		refCtx, refCancel := context.WithCancel(context.Background())
		ref.audit.Vor = func(support.Eintrag) {
			if beendet {
				refCancel()
			}
		}
		rb, rerr := ref.v.Delete(refCtx, "ws-1")
		refCancel()

		w := neu(t)
		legeWorkspaceAn(w, "ws-1")
		ctx, cancel := context.WithCancel(context.Background())
		erreicht, frei := make(chan struct{}), make(chan struct{})
		var stand []support.Eintrag
		var folge []string
		w.audit.Vor = func(support.Eintrag) {
			if beendet {
				cancel()
			}
			stand, folge = w.audit.Eintraege(), w.p.ops()
			close(erreicht)
			<-frei
		}
		fertig := make(chan ergebnis, 1)
		go func() {
			b, err := w.v.Delete(ctx, "ws-1")
			fertig <- ergebnis{b, err}
		}()
		select {
		case <-erreicht:
		case <-time.After(zeitgrenze):
			t.Fatalf("beendet %v: Append nicht erreicht", beendet)
		}
		select {
		case <-fertig:
			t.Errorf("beendet %v: Delete kehrte im Halt von Append zurück", beendet)
		case <-time.After(100 * time.Millisecond):
		}
		soll := []string{"Workspaces.Delete"}
		if !slices.Equal(folge, soll) || len(stand) != 0 || !slices.Equal(w.p.ops(), soll) {
			t.Errorf("beendet %v: im Halt Folge %v, danach %v, Einträge %+v", beendet, folge, w.p.ops(), stand)
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
		if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloeschtWS("ws-1")}) {
			t.Errorf("beendet %v: Einträge %+v", beendet, got)
		}
		if got := w.p.ops(); !slices.Equal(got, []string{"Workspaces.Delete", "Wartung.CleanupBlobs"}) {
			t.Errorf("beendet %v: Folge %v", beendet, got)
		}
	}
	if n != 2 {
		t.Errorf("Menge %d, erwartet 2", n)
	}
}
