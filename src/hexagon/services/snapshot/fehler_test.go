package snapshot_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// stelle ist eine Aufrufstelle von SnapshotUseCase (SPEC-129 Nr. 26): Nummer, Operation, Name der Operation des Fakes,
// Gegenstand der Abbildung und ob sie eine Fehlerrückgabe und ein ctx hat.
type stelle struct {
	nr, op, label string
	g             outbound.Gegenstand
	fehler, ctx   bool
}

func stellenSnapshot() []stelle {
	return []stelle{
		{"N1", "Create", "Workspaces.Get", outbound.GegenstandWorkspace, true, true},
		{"N2", "Create", "Snapshots.Get", outbound.GegenstandSnapshot, true, true},
		{"N3", "Create", "Entities.List", outbound.GegenstandWorkspace, true, true},
		{"N4", "Create", "Blobs.Begin", outbound.GegenstandSonstiges, true, true},
		{"N5", "Create", "Writer.Write", outbound.GegenstandSonstiges, true, false},
		{"N6", "Create", "Writer.Commit", outbound.GegenstandSonstiges, true, false},
		{"N7", "Create", "Writer.Abort", outbound.GegenstandSonstiges, true, false},
		{"N8", "Create", "IDs.SnapshotID", outbound.GegenstandSonstiges, false, false},
		{"N9", "Create", "Uhr.Now", outbound.GegenstandSonstiges, false, false},
		{"N10", "Create", "Snapshots.CreateWorkspace", outbound.GegenstandWorkspace, true, true},
		{"N11", "List", "Snapshots.ListWorkspace", outbound.GegenstandWorkspace, true, true},
	}
}

// m7Stellen sind die acht Stellen von M7 und M8: N1 bis N6, N10, N11.
func m7Stellen() []stelle {
	return slices.DeleteFunc(stellenSnapshot(), func(s stelle) bool { return !s.fehler || s.nr == "N7" })
}

func (s stelle) lauf(ctx context.Context, w *welt) error {
	if s.op == "List" {
		_, err := w.d.List(ctx, "ws-1")
		return err
	}
	_, err := w.d.Create(ctx, "ws-1", "n")
	return err
}

// m7Soll ist die Tabelle von SPEC-109 Nr. 5 von Hand: der Code für den Fehler fe am Gegenstand g.
func m7Soll(fe string, g outbound.Gegenstand) string {
	var c string
	switch fe {
	case "ErrNichtGefunden":
		c = map[outbound.Gegenstand]string{outbound.GegenstandWorkspace: "ODR-E7001", outbound.GegenstandSnapshot: "ODR-E7002"}[g]
	case "ErrBereitsVorhanden":
		c = map[outbound.Gegenstand]string{outbound.GegenstandWorkspace: "ODR-E7007", outbound.GegenstandSnapshotname: "ODR-E7006"}[g]
	case "ErrFremderService":
		c = "ODR-E5004"
	case "ErrFalscheArt":
		c = "ODR-E7008"
	case "ErrBlobFehlt", "ErrBeschaedigt":
		c = "ODR-E3001"
	}
	if c == "" {
		return "ODR-E9001"
	}
	return c
}

// TestStellenSnapshot (T74, M11): SnapshotUseCase hat 11 Stellen und 11 Paare, 9 mit Fehlerrückgabe, 6 mit ctx; die
// Operationen mit ctx tragen in einem Lauf von Create und List das ctx des Aufrufs, keine andere; die Aufrufstellen im
// Quelltext stehen in TestKonfigSnapshot.
func TestStellenSnapshot(t *testing.T) {
	st := stellenSnapshot()
	fehler, ctxn := 0, 0
	for _, s := range st {
		if s.fehler {
			fehler++
		}
		if s.ctx {
			ctxn++
		}
	}
	if len(st) != 11 || fehler != 9 || ctxn != 6 {
		t.Errorf("Stellen %d, mit Fehler %d, mit ctx %d", len(st), fehler, ctxn)
	}
	w := neu(t)
	ctx := neuerCtx()
	if _, err := w.d.Create(ctx, "ws-1", "n"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.d.List(ctx, "ws-1"); err != nil {
		t.Fatal(err)
	}
	var mit []string
	for _, e := range w.p.ev {
		if e.ctx != nil {
			mit = append(mit, e.op)
			if e.ctx != context.Context(ctx) {
				t.Errorf("%s: anderes ctx", e.op)
			}
		}
	}
	slices.Sort(mit)
	want := []string{"Blobs.Begin", "Entities.List", "Snapshots.CreateWorkspace", "Snapshots.Get", "Snapshots.ListWorkspace", "Workspaces.Get"}
	if !slices.Equal(mit, want) {
		t.Errorf("Aufrufe mit ctx %v, erwartet %v", mit, want)
	}
}

// TestCodesSnapshot (T20): die vier Codes als Literale aus dem Vertrag (SPEC-025, SPEC-065, SPEC-094, SPEC-096): Code,
// Status, Ziel und Text, 4 × 4 = 16 Vergleiche gegen das Modell.
func TestCodesSnapshot(t *testing.T) {
	soll := []struct {
		code, ziel, text string
		status           int
	}{
		{"ODR-E7002", "workspace", "Snapshot unknown", 404},
		{"ODR-E5004", "replay", "Assignment belongs to another service", 409},
		{"ODR-E7006", "workspace", "Snapshot name already taken", 409},
		{"ODR-E7008", "workspace", "Snapshot kind does not fit the operation", 422},
	}
	n := 0
	for _, s := range soll {
		c, ok := model.ParseCode(s.code)
		if !ok {
			t.Fatalf("%s fehlt im Modell", s.code)
		}
		if c.String() != s.code {
			t.Errorf("%s: Code %s", s.code, c.String())
		}
		if st := c.Stati(); len(st) != 1 || st[0] != s.status {
			t.Errorf("%s: Status %v", s.code, st)
		}
		if string(c.Zielbereich()) != s.ziel || c.Message() != s.text {
			t.Errorf("%s: Ziel %s Text %q", s.code, c.Zielbereich(), c.Message())
		}
		n += 4
	}
	if n != 16 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestPortFehlerSnapshot (T21, M7): Stelle 8 × Fehler 8 × Form 2 = 128, Orakel die Tabelle von Hand mit dem Gegenstand je
// Stelle; ein Fehler von Abort (N7) ändert nichts (zwei Fälle).
func TestPortFehlerSnapshot(t *testing.T) {
	n := 0
	for _, s := range m7Stellen() {
		for _, fe := range m7Fehler() {
			for form := 0; form < 2; form++ {
				w := neu(t)
				err := fe.err
				if form == 1 {
					err = fmt.Errorf("verpackt: %w", fe.err)
				}
				w.p.setzeFehler(s.label, err)
				got := s.lauf(context.Background(), w)
				g := s.g
				if s.nr == "N10" && fe.name == "ErrBereitsVorhanden" {
					g = outbound.GegenstandSnapshotname
				}
				soll := m7Soll(fe.name, g)
				if c, ok := codeVon(got); !ok || c.Code.String() != soll || c.Status != c.Code.Stati()[0] || !errors.Is(got, fe.err) {
					t.Errorf("%s %s Form %d: %v, erwartet %s", s.nr, fe.name, form, got, soll)
				}
				n++
			}
		}
	}
	if n != 128 {
		t.Errorf("M7: %d Elemente", n)
	}
	for _, op := range []string{"Writer.Write", "Writer.Commit"} {
		w := neu(t)
		w.p.setzeFehler(op, errPlatte)
		w.blobs.abortFehler = errors.New("Abort scheitert")
		_, err := w.d.Create(context.Background(), "ws-1", "n")
		pruefeCode(t, err, "ODR-E9001", errPlatte)
		if w.p.anzahl("Writer.Abort") != 1 {
			t.Errorf("%s: Abort %d", op, w.p.anzahl("Writer.Abort"))
		}
	}
}

// TestKontextSnapshot (T22, M8): ein Kontextfehler in vier Formen an jeder der acht Stellen kehrt unverändert und ohne Code
// zurück, auch neben ErrNichtGefunden, und danach wird kein weiterer Port gerufen (Abort nach Write und Commit ausgenommen);
// das ctx des Aufrufs erreicht jeden Port mit ctx per Identität (32 Fälle).
func TestKontextSnapshot(t *testing.T) {
	formen := map[string]func(error) error{
		"nackt":       func(e error) error { return e },
		"%w":          func(e error) error { return fmt.Errorf("x: %w", e) },
		"Join vorn":   func(e error) error { return errors.Join(e, outbound.ErrNichtGefunden) },
		"Join hinten": func(e error) error { return errors.Join(outbound.ErrNichtGefunden, e) },
	}
	n := 0
	for _, s := range m7Stellen() {
		for name, f := range formen {
			w := neu(t)
			inj := f(context.Canceled)
			w.p.setzeFehler(s.label, inj)
			ctx := neuerCtx()
			err := s.lauf(ctx, w)
			if err != inj {
				t.Errorf("%s %s: %v", s.nr, name, err)
			}
			if _, ok := codeVon(err); ok {
				t.Errorf("%s %s: Code", s.nr, name)
			}
			ops := slices.DeleteFunc(w.p.ops(), func(op string) bool { return op == "Writer.Abort" })
			if ops[len(ops)-1] != s.label {
				t.Errorf("%s %s: letzter Aufruf %s", s.nr, name, ops[len(ops)-1])
			}
			for _, e := range w.p.ev {
				if e.ctx != nil && e.ctx != context.Context(ctx) {
					t.Errorf("%s %s: %s mit anderem ctx", s.nr, name, e.op)
				}
			}
			n++
		}
	}
	if n != 32 {
		t.Errorf("M8: %d Elemente", n)
	}
}

// erreichbar sind die Codes, die die Matrix M7 über Create und List erreicht (die Abbildung eingeschlossen); nie die
// übrigen. Beide Listen stehen als Literale im Test.
func erreichbar() []string {
	return []string{"ODR-E3001", "ODR-E5004", "ODR-E7001", "ODR-E7002", "ODR-E7006", "ODR-E7007", "ODR-E7008", "ODR-E9001"}
}

func nie() []string {
	return []string{"ODR-E2001", "ODR-E2101", "ODR-E2401", "ODR-E4001", "ODR-E5001", "ODR-E5002", "ODR-E5003", "ODR-E5005", "ODR-E5006",
		"ODR-E5007", "ODR-E5008", "ODR-E5009", "ODR-E5010", "ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E6004", "ODR-E6005", "ODR-E6006",
		"ODR-E6007", "ODR-E6008", "ODR-E6009", "ODR-E6101", "ODR-E6102", "ODR-E7003", "ODR-E7004", "ODR-E7005", "ODR-E7009", "ODR-E7010",
		"ODR-E7011", "ODR-E7012"}
}

// TestRueckgabenSnapshot (T75, M9): die im Ablauf und über die Abbildung erreichbaren Codes kommen in den Läufen von M7 vor,
// die übrigen nie; die beiden Listen des Tests teilen model.AlleCodes() ohne Rest und ohne Überschneidung.
func TestRueckgabenSnapshot(t *testing.T) {
	gesehen := map[string]bool{}
	for _, s := range m7Stellen() {
		for _, fe := range m7Fehler() {
			w := neu(t)
			w.p.setzeFehler(s.label, fe.err)
			if f, ok := codeVon(s.lauf(context.Background(), w)); ok {
				gesehen[f.Code.String()] = true
			}
		}
	}
	for _, c := range erreichbar() {
		if !gesehen[c] {
			t.Errorf("%s kommt in keinem Lauf vor", c)
		}
	}
	for _, c := range nie() {
		if gesehen[c] {
			t.Errorf("%s kommt vor", c)
		}
	}
	if len(gesehen) != len(erreichbar()) {
		t.Errorf("gesehen %v", gesehen)
	}
	alle := model.AlleCodes()
	if len(erreichbar())+len(nie()) != len(alle) {
		t.Errorf("Listen %d + %d, Codes %d", len(erreichbar()), len(nie()), len(alle))
	}
	for _, c := range alle {
		if in := slices.Contains(erreichbar(), c.String()); in == slices.Contains(nie(), c.String()) {
			t.Errorf("%s steht in keiner oder in beiden Listen", c.String())
		}
	}
}
