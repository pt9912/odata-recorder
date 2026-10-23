package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// operation führt eine Operation des Dienstes aus und meldet, ob ihr Ergebnis der Nullwert ist.
type operation struct {
	name string
	f    func(context.Context, *welt) (leer bool, err error)
}

func operationen() []operation {
	return []operation{
		{"Create mit leerer id", func(ctx context.Context, u *welt) (bool, error) {
			s, err := u.v.Create(ctx, dienstA, "")
			return s == model.Session{}, err
		}},
		{"Create mit id", func(ctx context.Context, u *welt) (bool, error) {
			s, err := u.v.Create(ctx, dienstA, "x1")
			return s == model.Session{}, err
		}},
		{"OpenInitialSession", func(ctx context.Context, u *welt) (bool, error) {
			s, err := u.v.OpenInitialSession(ctx, dienstA)
			return s == model.Session{}, err
		}},
		{"Get", func(ctx context.Context, u *welt) (bool, error) {
			s, err := u.v.Get(ctx, "s1")
			return s == model.Session{}, err
		}},
		{"List", func(ctx context.Context, u *welt) (bool, error) {
			s, err := u.v.List(ctx)
			return s == nil, err
		}},
		{"Delete", func(ctx context.Context, u *welt) (bool, error) {
			b, err := u.v.Delete(ctx, "s1")
			return b.Entfernt == 0 && !b.Zurueckgestellt, err
		}},
	}
}

// bestand legt die Services a, b, c und die Session s1 (complete) an.
func (u *welt) bestand() {
	u.dienst(dienstA, dienstB, dienstC)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
}

// TestKontext: ein Kontextfehler kehrt unverändert und ohne Code zurück, beendet die Operation an
// der Stelle und gewinnt auch neben einem Fehler der Tabelle (V3).
func TestKontext(t *testing.T) {
	// (a) ctx schon beendet: je Operation der Kontextfehler selbst, der Nullwert als Ergebnis.
	ende, cancel := context.WithCancel(context.Background())
	cancel()
	for _, op := range operationen() {
		u := neu(t)
		u.bestand()
		leer, err := op.f(ende, u)
		if err != context.Canceled || !leer {
			t.Errorf("(a) %s: %v, Nullwert %v", op.name, err, leer)
		}
	}

	// (b) der Port liefert den Kontextfehler an jeder Stelle; danach folgt kein Aufruf mehr.
	stellen := []struct {
		name    string
		op      string
		n       int
		f       int
		zaehler map[string]int
	}{
		{"Sessions.Create", "Sessions.Create", 1, 1, map[string]int{"Sessions.Create": 1}},
		{"Sessions.Get der Operation Get", "Sessions.Get", 1, 3, map[string]int{"Sessions.Get": 1}},
		{"Dienste.List", "Dienste.List", 1, 4, map[string]int{"Dienste.List": 1, "Sessions.List": 0}},
		{"Sessions.List des zweiten", "Sessions.List", 2, 4, map[string]int{"Sessions.List": 2}},
		{"Sessions.Get von Delete", "Sessions.Get", 1, 5, map[string]int{"Sessions.Get": 1, "Sessions.Delete": 0, "CleanupBlobs": 0}},
		{"Sessions.Delete", "Sessions.Delete", 1, 5, map[string]int{"Sessions.Delete": 1, "CleanupBlobs": 0}},
	}
	ops := operationen()
	for _, s := range stellen {
		u := neu(t)
		u.bestand()
		u.st.scheitere(s.op, s.n, context.Canceled)
		leer, err := ops[s.f].f(context.Background(), u)
		if err != context.Canceled || !leer {
			t.Errorf("(b) %s: %v, Nullwert %v", s.name, err, leer)
		}
		for op, n := range s.zaehler {
			if got := u.st.anzahl(op); got != n {
				t.Errorf("(b) %s: %s %d-mal, erwartet %d", s.name, op, got, n)
			}
		}
	}

	// (c) neben einem Fehler der Tabelle: der Kontextfehler, ohne Code.
	for _, ctxErr := range []error{context.Canceled, context.DeadlineExceeded} {
		beides := errors.Join(outbound.ErrReferenziert, ctxErr)
		u := neu(t)
		u.bestand()
		u.st.scheitere("Sessions.Delete", 1, beides)
		_, err := u.v.Delete(context.Background(), "s1")
		if err != beides || !ohneCode(err) || !errors.Is(err, ctxErr) {
			t.Errorf("(c) Delete mit %v: %v", ctxErr, err)
		}
		if n := u.st.anzahl("CleanupBlobs"); n != 0 {
			t.Errorf("(c) Delete mit %v: Bereinigung %d-mal", ctxErr, n)
		}
		doppelt := errors.Join(outbound.ErrBereitsVorhanden, ctxErr)
		u = neu(t)
		u.st.scheitere("Sessions.Create", 1, doppelt)
		_, err = u.v.Create(context.Background(), dienstA, "x1")
		if err != doppelt || !ohneCode(err) || !errors.Is(err, ctxErr) {
			t.Errorf("(c) Create mit %v: %v", ctxErr, err)
		}
	}
}

// TestFehlerWert: jeder *Fehler trägt den einen Status seines Codes, Error() ist Code und message
// ohne den Text der Ursache, Unwrap erreicht die Ursache (V16). Alle Fälle der Operationen laufen
// außerdem durch pruefeFehlerWert der Helfer.
func TestFehlerWert(t *testing.T) {
	faelle := []struct {
		name   string
		setup  func(*welt, error)
		op     func(*welt) error
		ursach error
		code   string
		status int
		text   string
	}{
		{"E7010", func(u *welt, e error) { u.st.scheitere("Sessions.Create", 1, e) },
			func(u *welt) error { _, err := u.create(context.Background(), dienstA, "x1"); return err },
			verpackt(outbound.ErrBereitsVorhanden), "ODR-E7010", 409, "ODR-E7010: Session id already taken"},
		{"E5005", func(u *welt, e error) { u.st.scheitere("Sessions.Create", 1, e) },
			func(u *welt) error { _, err := u.create(context.Background(), dienstA, "x1"); return err },
			verpackt(outbound.ErrDienstUnbekannt), "ODR-E5005", 404, ""},
		{"E3001", func(u *welt, e error) { u.st.scheitere("Sessions.Get", 1, e) },
			func(u *welt) error { _, err := u.get(context.Background(), "s1"); return err },
			verpackt(outbound.ErrBeschaedigt), "ODR-E3001", 500, ""},
		{"E9001", func(u *welt, e error) { u.st.scheitere("Dienste.List", 1, e) },
			func(u *welt) error { _, err := u.list(context.Background()); return err },
			errDisk, "ODR-E9001", 500, ""},
		{"E7003", func(*welt, error) {},
			func(u *welt) error { _, err := u.get(context.Background(), "nix"); return err },
			outbound.ErrNichtGefunden, "ODR-E7003", 404, "ODR-E7003: Recording session unknown"},
		{"E7009", func(u *welt, _ error) { u.sessions.fuege("o1", dienstA, t0(), model.SessionOpen) },
			func(u *welt) error { _, err := u.del(context.Background(), "o1"); return err },
			nil, "ODR-E7009", 409, "ODR-E7009: Session is still open"},
		{"E7004", func(u *welt, e error) { u.st.scheitere("Sessions.Delete", 1, e) },
			func(u *welt) error { _, err := u.del(context.Background(), "s1"); return err },
			verpackt(outbound.ErrReferenziert), "ODR-E7004", 409, "ODR-E7004: Session is referenced by a baseline"},
	}
	for _, f := range faelle {
		u := neu(t)
		u.bestand()
		f.setup(u, f.ursach)
		err := f.op(u)
		c, st := codeVon(t, err)
		if c != f.code || st != f.status {
			t.Errorf("%s: %s %d", f.name, c, st)
		}
		if f.text != "" && err.Error() != f.text {
			t.Errorf("%s: Error() = %q, erwartet %q", f.name, err.Error(), f.text)
		}
		if f.ursach != nil && !errors.Is(err, f.ursach) {
			t.Errorf("%s: die Ursache ist nicht erreichbar", f.name)
		}
		if f.ursach == nil && errors.Unwrap(err) != nil {
			t.Errorf("%s: Unwrap = %v", f.name, errors.Unwrap(err))
		}
	}
}
