package maintenance_test

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
)

// TestWaechterDurchgereicht: der Wächter reicht jeden Aufruf unverändert durch: gleiche Argumente
// am Fake darunter, Ergebnisse und Fehler mit == zurück (W18).
func TestWaechterDurchgereicht(t *testing.T) {
	u := neu(t)
	ctx := context.Background()
	var wr outbound.BlobRepository = u.waechter
	if _, err := wr.Begin(ctx, 0); err != model.ErrUngueltig {
		t.Errorf("Begin(0): %v", err)
	}
	w, err := wr.Begin(ctx, 8)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := w.Write(make([]byte, 9)); n != 0 || err != outbound.ErrZuGross {
		t.Errorf("Write über die Grenze: %d, %v", n, err)
	}
	w2, err := wr.Begin(ctx, 8)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := w2.Write([]byte("abc")); n != 3 || err != nil {
		t.Errorf("Write: %d, %v", n, err)
	}
	ref, laenge, err := w2.Commit()
	if ref != refVon([]byte("abc")) || laenge != 3 || err != nil {
		t.Errorf("Commit: %q, %d, %v", ref, laenge, err)
	}
	if ref2, l2, err := w2.Commit(); ref2 != "" || l2 != 0 || err != outbound.ErrAbgeschlossen {
		t.Errorf("zweiter Commit: %q, %d, %v", ref2, l2, err)
	}
	if err := w2.Abort(); err != nil || !u.blobs.hat(ref) {
		t.Errorf("Abort nach Commit: %v, Datei vorhanden %v", err, u.blobs.hat(ref))
	}
	if err := w.Abort(); err != nil {
		t.Errorf("Abort: %v", err)
	}
	w4, err := wr.Begin(ctx, 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w4.Write([]byte("d")); err != nil {
		t.Fatal(err)
	}
	ref4, _, err := w4.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if err := w4.Abort(); err != nil || !u.blobs.hat(ref4) {
		t.Errorf("Abort unmittelbar nach Commit: %v, Datei vorhanden %v", err, u.blobs.hat(ref4))
	}
	if rd, err := wr.Open(ctx, ref); err != nil {
		t.Errorf("Open: %v", err)
	} else if b, _ := io.ReadAll(rd); string(b) != "abc" {
		t.Errorf("Open liest %q", b)
	}
	for _, f := range []struct {
		name string
		ruf  func(string) error
	}{
		{"Open", func(ref string) error { _, err := wr.Open(ctx, ref); return err }},
		{"Verify", func(ref string) error { return wr.Verify(ctx, ref) }},
	} {
		if err := f.ruf(r("x")); err != outbound.ErrBlobFehlt {
			t.Errorf("%s einer fehlenden Referenz: %v", f.name, err)
		}
		if err := f.ruf("x"); err != model.ErrUngueltig {
			t.Errorf("%s einer ungültigen Referenz: %v", f.name, err)
		}
	}
	if err := wr.Verify(ctx, ref); err != nil {
		t.Errorf("Verify: %v", err)
	}
	if got, err := wr.List(ctx); err != nil || !slices.Equal(got, slices.Sorted(slices.Values([]string{ref, ref4}))) {
		t.Errorf("List: %v, %v", got, err)
	}
	if err := wr.Remove(ctx, r("x")); err != nil {
		t.Errorf("Remove einer unbekannten Referenz: %v", err)
	}
	if err := wr.Remove(ctx, "x"); err != model.ErrUngueltig {
		t.Errorf("Remove einer ungültigen Referenz: %v", err)
	}
	abgebrochen, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := wr.Begin(abgebrochen, 8); err != context.Canceled {
		t.Errorf("Begin bei beendetem ctx: %v", err)
	}
	soll := map[string][]string{
		"Begin":  {"Begin(0)", "Begin(8)", "Begin(8)", "Begin(8)", "Begin(8)"},
		"Open":   {ev("Open", ref), ev("Open", r("x")), ev("Open", "x")},
		"Verify": {ev("Verify", r("x")), ev("Verify", "x"), ev("Verify", ref)},
		"Remove": {ev("Remove", r("x")), ev("Remove", "x")},
	}
	for op, folge := range soll {
		if got := u.p.mit(op); !slices.Equal(got, folge) {
			t.Errorf("%s am Fake: %v, erwartet %v", op, got, folge)
		}
	}
	pruefeKontextDurchgereicht(t)
}

// pruefeKontextDurchgereicht ruft jede Operation des Wächters, die einen ctx trägt, mit beendetem
// und mit lebendem ctx; der ctx trägt einen Wert unter ctxSchluessel, den ein context.Background()
// nicht trägt. Der Fake darunter liefert den Fehler des ctx wie ein echter Port und merkt den Wert
// (W18; Write, Commit und Abort haben keinen ctx).
func pruefeKontextDurchgereicht(t *testing.T) {
	t.Helper()
	ops := []struct {
		name string
		ruf  func(ctx context.Context, w outbound.BlobRepository) error
	}{
		{"Begin", func(ctx context.Context, w outbound.BlobRepository) error { _, err := w.Begin(ctx, 8); return err }},
		{"Open", func(ctx context.Context, w outbound.BlobRepository) error { _, err := w.Open(ctx, r("x")); return err }},
		{"Verify", func(ctx context.Context, w outbound.BlobRepository) error { return w.Verify(ctx, r("x")) }},
		{"List", func(ctx context.Context, w outbound.BlobRepository) error { _, err := w.List(ctx); return err }},
		{"Remove", func(ctx context.Context, w outbound.BlobRepository) error { return w.Remove(ctx, r("x")) }},
	}
	for _, o := range ops {
		for _, beendet := range []bool{true, false} {
			u := neu(t)
			u.blobs.lege(r("x"))
			wert := "wert-" + o.name
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, wert))
			var soll error
			if beendet {
				cancel()
				soll = context.Canceled
			}
			err := o.ruf(ctx, u.waechter)
			cancel()
			if err != soll {
				t.Errorf("%s, ctx beendet %v: %v, erwartet %v", o.name, beendet, err, soll)
			}
			if got := u.blobs.ctxWerte(o.name); !slices.Equal(got, []any{wert}) {
				t.Errorf("%s, ctx beendet %v: der Fake erhielt den Wert %v, erwartet %q", o.name, beendet, got, wert)
			}
		}
	}
}

// TestWaechterZeit: der Wächter merkt die Zeit jedes gelungenen Commit, die letzte gilt, ein
// gescheiterter merkt nichts, und er hält nur die Referenzen der letzten Karenz (W19).
func TestWaechterZeit(t *testing.T) {
	u := neu(t)
	T := t0()
	ctx := context.Background()
	x := u.commit("x")
	if got := u.waechter.Gemerkt(); !maps.Equal(got, map[string]time.Time{x: T}) {
		t.Errorf("nach dem ersten Commit: %v", got)
	}
	u.uhr.Advance(time.Hour)
	u.commit("x")
	leer := u.commit("")
	if got := u.waechter.Gemerkt(); !maps.Equal(got, map[string]time.Time{x: T.Add(time.Hour), refVon(nil): T.Add(time.Hour)}) || leer != refVon(nil) {
		t.Errorf("nach dem zweiten Commit: %v", got)
	}

	vorher := u.waechter.Gemerkt()
	w, err := u.waechter.Begin(ctx, 16)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("zz")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Commit(); err != nil {
		t.Fatal(err)
	}
	mitZz := u.waechter.Gemerkt()
	if _, _, err := w.Commit(); err != outbound.ErrAbgeschlossen || !maps.Equal(u.waechter.Gemerkt(), mitZz) || len(mitZz) != len(vorher)+1 {
		t.Errorf("zweiter Commit: %v, %v", err, u.waechter.Gemerkt())
	}
	abgebrochen, cancel := context.WithCancel(ctx)
	w3, err := u.waechter.Begin(abgebrochen, 16)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w3.Write([]byte("yy")); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, _, err := w3.Commit(); err != context.Canceled || !maps.Equal(u.waechter.Gemerkt(), mitZz) {
		t.Errorf("Commit bei beendetem ctx: %v, %v", err, u.waechter.Gemerkt())
	}

	// Bereinigung bei jedem Commit: r1 bei T, r2 bei T + karenz - 1 ns, r3 bei T + karenz.
	v := neu(t)
	karenz := maintenance.StandardKarenz
	r1 := v.commit("1")
	v.uhr.Advance(karenz - time.Nanosecond)
	r2 := v.commit("2")
	if got := slices.Sorted(maps.Keys(v.waechter.Gemerkt())); !slices.Equal(got, slices.Sorted(slices.Values([]string{r1, r2}))) {
		t.Errorf("knapp unter der Karenz: %v", got)
	}
	v.uhr.Advance(time.Nanosecond)
	r3 := v.commit("3")
	if got := slices.Sorted(maps.Keys(v.waechter.Gemerkt())); !slices.Equal(got, slices.Sorted(slices.Values([]string{r2, r3}))) {
		t.Errorf("auf der Karenz: %v", got)
	}
}

// TestCommitMerktUnterSperre: Commit merkt die Zeit unter der Lesersperre; im Halt von Now ist die
// Datei da und die Schreibersperre nicht zu nehmen (W25).
func TestCommitMerktUnterSperre(t *testing.T) {
	u := neu(t)
	rn := refVon([]byte("K"))
	h := u.st.halte("Now", 1)
	defer h.gib()
	fertig := make(chan error, 1)
	go func() {
		w, err := u.waechter.Begin(context.Background(), 16)
		if err == nil {
			_, err = w.Write([]byte("K"))
		}
		if err == nil {
			_, _, err = w.Commit()
		}
		fertig <- err
	}()
	warte(t, h.erreicht, "Now im Commit")
	if !u.blobs.hat(rn) {
		t.Error("im Halt besteht die Datei nicht")
	}
	if u.waechter.TryLock() {
		u.waechter.Unlock()
		t.Error("im Halt ist die Schreibersperre frei")
	}
	// Zusatzprüfung zur Wirkung, kein Fänger: sie wird bei einer Mutation nur mit einer
	// Wahrscheinlichkeit rot, denn der Test wartet nicht, bis CleanupBlobs die Sperre erreicht hat.
	// Fänger der Zusage ist TryLock im Halt.
	ch := u.cleanupAsync()
	h.gib()
	if err := <-fertig; err != nil {
		t.Errorf("Commit: %v", err)
	}
	if e := <-ch; e.err != nil || e.b.Entfernt != 0 || !u.blobs.hat(rn) {
		t.Errorf("CleanupBlobs: %v %+v, Datei vorhanden %v", e.err, e.b, u.blobs.hat(rn))
	}
}

// TestKarenzUhr: CleanupBlobs bezieht Uhr.Now() genau einmal, nach dem Nehmen der Sperre und nach
// dem Zurückstellen (W30).
func TestKarenzUhr(t *testing.T) {
	// (a) einmal für den Aufruf, nicht je Kandidat: drei Kandidaten, einer davon vom Wächter committet.
	u := neu(t)
	ra := u.commit("a")
	u.blobs.lege(r("b"), r("c"))
	vor := u.st.anzahl("Now")
	c, err := u.cleanup(context.Background())
	if n := u.st.anzahl("Now") - vor; n != 1 || err != nil || c.Entfernt != 2 || !u.blobs.hat(ra) {
		t.Errorf("(a) %d-mal Now: %v %+v, Datei des Wächters vorhanden %v", n, err, c, u.blobs.hat(ra))
	}

	// (b) Now hält an: die Sperre ist genommen und FindOpen schon gerufen.
	v := neu(t)
	v.blobs.lege(r("b"))
	h := v.st.halte("Now", 1)
	t.Cleanup(h.gib)
	ch := v.cleanupAsync()
	warte(t, h.erreicht, "Now in CleanupBlobs")
	sperreHalten(t, v, "Now in CleanupBlobs")
	if got := v.p.mit("FindOpen"); len(got) != 1 {
		t.Errorf("(b) im Halt von Now steht FindOpen nicht im Protokoll: %v", v.p.liste())
	}
	h.gib()
	if e := <-ch; e.err != nil || e.b.Entfernt != 1 {
		t.Errorf("(b) %v %+v", e.err, e.b)
	}

	// (c) Zurückgestellt: Now wird nicht gerufen.
	x := neu(t)
	x.blobs.lege(r("b"))
	x.sessions.fuege("s1", dienstA, t0().Add(-time.Hour), model.SessionOpen)
	if c, err := x.cleanup(context.Background()); err != nil || !c.Zurueckgestellt || x.st.anzahl("Now") != 0 {
		t.Errorf("(c) %v %+v, Now %d-mal", err, c, x.st.anzahl("Now"))
	}
}

// TestWaechterGleichzeitig: der Zustand des Wächters liegt unter einer eigenen Sperre, mehrere
// Commit halten die Lesersperre zugleich (W27). Ein Schreiben der Map ohne die eigene Sperre
// beendet den Prozess mit `concurrent map writes` und meldet unter make test-race ein Datenrennen
// (SPEC-122 Nr. 8a: der Detektor sieht nur die Zugriffe, die dieser Test ausführt).
func TestWaechterGleichzeitig(t *testing.T) {
	u := neu(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var refs []string
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 50 {
				inhalt := fmt.Sprintf("g%d-i%d", g, i)
				if i%5 == 0 {
					inhalt = fmt.Sprintf("gleich-%d", i%3)
				}
				w, err := u.waechter.Begin(context.Background(), 64)
				if err == nil {
					_, err = w.Write([]byte(inhalt))
				}
				var ref string
				if err == nil {
					ref, _, err = w.Commit()
				}
				if err != nil {
					t.Error(err)
					return
				}
				mu.Lock()
				refs = append(refs, ref)
				mu.Unlock()
			}
		}()
	}
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				if _, err := u.w.CleanupBlobs(context.Background()); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 100 {
			u.uhr.Advance(time.Millisecond)
		}
	}()
	wg.Wait()
	if len(refs) != 800 {
		t.Fatalf("%d Commit gelungen, erwartet 800", len(refs))
	}
	for _, ref := range refs {
		if !u.blobs.hat(ref) {
			t.Fatalf("Datei %s fehlt", ref)
		}
	}
}
