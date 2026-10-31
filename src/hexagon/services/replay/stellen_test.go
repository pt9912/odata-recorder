package replay_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// stelle ist eine Aufrufstelle eines Ports mit fehlerfähiger Rückgabe (SPEC-135 Nr. 18); op ist der Name im Protokoll.
type stelle struct {
	name string
	op   string
	// gegenstand ist der Gegenstand der Fehlerabbildung.
	gegenstand outbound.Gegenstand
}

func stellen() []stelle {
	return []stelle{
		{"S1", "Sessions.Get", outbound.GegenstandSession},
		{"S3", "Exchanges.ListBySession", outbound.GegenstandSession},
		{"S4", "Transaktion.WithWorkspaceTransaction", outbound.GegenstandSonstiges},
		{"S5", "tx Workspaces.Get", outbound.GegenstandWorkspace},
		{"S6", "tx Cursors.Get", outbound.GegenstandSonstiges},
		{"S7", "Blobs.Open", outbound.GegenstandSonstiges},
		{"S8", "tx Cursors.Put", outbound.GegenstandSession},
	}
}

// erfolgsLauf richtet den Lauf ein, der jede Stelle S1 bis S8 und S10 erreicht: ein replayfähiger Exchange mit
// KeyRedacted und Body.
func erfolgsLauf(u *umg) inbound.Anfrage {
	e := u.austausch(1, u.keyA(), true)
	e.KeyRedacted = true
	u.lege(e)
	return anf()
}

// TestKontextJederStelle hält Z4: Menge 8 (S1, S3 bis S8, S10); jeder Aufruf erhält den ctx des Aufrufs.
func TestKontextJederStelle(t *testing.T) {
	u := neu(t)
	k := neuerKontext(context.Background())
	if _, err := u.replay(k, erfolgsLauf(u)); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	for _, s := range stellen() {
		evs := u.p.von(s.op)
		if len(evs) != 1 {
			t.Errorf("%s: %d Aufrufe von %s", s.name, len(evs), s.op)
			continue
		}
		if evs[0].ctx != ctxMethoden(k) {
			t.Errorf("%s: %s erhielt einen anderen ctx", s.name, s.op)
		}
	}
	logs := u.tel.Aufrufe()
	if len(logs) != 1 || logs[0].Ctx() != context.Context(k) {
		t.Errorf("S10: Telemetrie.Log erhielt einen anderen ctx (%d Aufrufe)", len(logs))
	}
}

// folgen sind die Operationen nach der Stelle in der Reihenfolge von SPEC-135 Nr. 3.
func folgen(i int) []string {
	var r []string
	for _, s := range stellen()[i+1:] {
		r = append(r, s.op)
	}
	return r
}

// TestKontextfehlerJederStelle hält Z5: Menge 28 (7 Stellen × 4 Fehler); der Kontextfehler kehrt unverändert und
// ohne Code zurück, auch neben einem Fehler der Tabelle; danach kein Port, ein geöffneter Strom wird geschlossen.
func TestKontextfehlerJederStelle(t *testing.T) {
	fehler := map[string]func() (error, error){
		"Canceled":         func() (error, error) { return context.Canceled, context.Canceled },
		"DeadlineExceeded": func() (error, error) { return context.DeadlineExceeded, context.DeadlineExceeded },
		"Join(Canceled, ErrNichtGefunden)": func() (error, error) {
			return errors.Join(context.Canceled, outbound.ErrNichtGefunden), context.Canceled
		},
		"Join(ErrNichtGefunden, DeadlineExceeded)": func() (error, error) {
			return errors.Join(outbound.ErrNichtGefunden, context.DeadlineExceeded), context.DeadlineExceeded
		},
	}
	n := 0
	for i, s := range stellen() {
		for name, f := range fehler {
			n++
			u := neu(t)
			a := erfolgsLauf(u)
			err, ursache := f()
			u.p.setzeFehler(s.op, err)
			_, got := u.replay(context.Background(), a)
			erwarteOhneCode(t, s.name+" "+name, got, ursache)
			for _, op := range folgen(i) {
				if c := u.p.anzahl(op); c != 0 {
					t.Errorf("%s %s: %s nach dem Kontextfehler %d-mal gerufen", s.name, name, op, c)
				}
			}
			if u.tel.Anzahl(supporttest.OpLog) != 0 {
				t.Errorf("%s %s: Eintrag nach dem Kontextfehler", s.name, name)
			}
			if s.name == "S8" {
				st := u.bl.strome()
				if len(st) != 1 || st[0].geschlossen() != 1 {
					t.Errorf("%s %s: Strom nicht genau einmal geschlossen", s.name, name)
				}
			}
		}
	}
	if n != 28 {
		t.Errorf("Menge: %d Elemente, erwartet 28", n)
	}
}

// TestAbbildungJederStelle hält Z6: Menge 63 (7 Stellen × 9 Fehler), erwartet aus der Handtabelle von SPEC-109
// Nr. 5 und SPEC-135 Nr. 15, 18; ein Eintrag 0 der Tafel heißt Erfolg.
func TestAbbildungJederStelle(t *testing.T) {
	ng := outbound.ErrNichtGefunden
	type fall struct {
		name string
		err  error
	}
	faelle := []fall{
		{"ErrNichtGefunden", ng}, {"ErrBeschaedigt", outbound.ErrBeschaedigt}, {"ErrBlobFehlt", outbound.ErrBlobFehlt},
		{"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt}, {"ErrFremderService", outbound.ErrFremderService},
		{"ErrUngueltig", model.ErrUngueltig}, {"anderer", errors.New("anderer")},
		{"verpackt", fmt.Errorf("verpackt: %w", ng)}, {"Join(ErrNichtGefunden, ErrBeschaedigt)", errors.Join(ng, outbound.ErrBeschaedigt)},
	}
	// Spalten in der Reihenfolge der Fälle; 0 ist Erfolg (S6, fehlender Cursor).
	tafel := map[string][]model.Code{
		"S1": {model.CodeE7003, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE7003, model.CodeE3001},
		"S3": {model.CodeE7003, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE7003, model.CodeE3001},
		"S4": {model.CodeE9001, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE3001},
		"S5": {model.CodeE7001, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE7001, model.CodeE3001},
		"S6": {0, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, 0, model.CodeE3001},
		"S7": {model.CodeE9001, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE3001},
		"S8": {model.CodeE7003, model.CodeE3001, model.CodeE3001, model.CodeE3001, model.CodeE5004, model.CodeE9001, model.CodeE9001, model.CodeE7003, model.CodeE3001},
	}
	n := 0
	for _, s := range stellen() {
		for i, f := range faelle {
			n++
			u := neu(t)
			a := erfolgsLauf(u)
			u.p.setzeFehler(s.op, f.err)
			resp, err := u.replay(context.Background(), a)
			want := tafel[s.name][i]
			if want == 0 {
				if err != nil {
					t.Errorf("%s %s: Fehler %v, erwartet Erfolg", s.name, f.name, err)
				} else {
					_ = resp.Body.Close()
				}
				continue
			}
			erwarteCode(t, s.name+" "+f.name, err, want)
		}
	}
	if n != 63 {
		t.Errorf("Menge: %d Elemente, erwartet 63", n)
	}
}

// TestFehlerwert hält Z7: Code, Status des Codes, Ursache über Unwrap, Error() ohne Wert aus Request, Antwort
// oder Ursache; Menge 13.
func TestFehlerwert(t *testing.T) {
	if n := len(szenarien()); n != 13 {
		t.Fatalf("%d Szenarien, erwartet 13", n)
	}
	for _, sz := range szenarien() {
		u := neu(t)
		a := sz.baue(u, model.StrategySequential)
		// Pfad und Query gehen in den Key; der Canary steht dort nur, wo der Fehler vor der Auswahl endet.
		switch sz.code {
		case model.CodeE5003, model.CodeE5004, model.CodeE7001, model.CodeE7003, model.CodeE7009:
			a.Pfad += "/" + canary
			a.RawQuery += "q=" + canary
		}
		a.Header[canary] = []string{canary}
		_, err := u.replay(context.Background(), a)
		var f *inbound.Fehler
		if !errors.As(err, &f) {
			t.Errorf("%s: Fehler %v ohne Code", sz.code, err)
			continue
		}
		if f.Code != sz.code {
			t.Errorf("%s: Code %s", sz.code, f.Code)
		}
		if f.Status != sz.status {
			t.Errorf("%s: Status %d, erwartet %d", sz.code, f.Status, sz.status)
		}
		if want := sz.code.String() + ": " + sz.code.Message(); f.Error() != want {
			t.Errorf("%s: Error() %q, erwartet %q", sz.code, f.Error(), want)
		}
		if strings.Contains(err.Error(), canary) {
			t.Errorf("%s: Error() nennt einen Wert: %q", sz.code, err.Error())
		}
	}
	// Ursache über Unwrap: der Fehler des Ports bleibt erreichbar.
	for _, c := range []struct {
		op    string
		cause error
		code  model.Code
	}{
		{"Blobs.Open", outbound.ErrBlobFehlt, model.CodeE3001},
		{"Sessions.Get", outbound.ErrNichtGefunden, model.CodeE7003},
		{"tx Cursors.Put", outbound.ErrFremderService, model.CodeE5004},
	} {
		u := neu(t)
		a := erfolgsLauf(u)
		u.p.setzeFehler(c.op, errors.Join(c.cause, errors.New(canary)))
		_, err := u.replay(context.Background(), a)
		erwarteCode(t, c.op, err, c.code)
		if !errors.Is(err, c.cause) {
			t.Errorf("%s: Ursache nicht über Unwrap erreichbar", c.op)
		}
		if strings.Contains(err.Error(), canary) {
			t.Errorf("%s: Error() nennt den Text der Ursache", c.op)
		}
	}
}
