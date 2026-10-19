package outboundtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// BlobAblage ist eine frische, leere Ablage mit den Hilfen, die der Vertrag braucht.
type BlobAblage struct {
	Blobs outbound.BlobRepository
	// Dateien liefert die Namen aller Dateien der Ablage, auch offener Writer, sortiert.
	Dateien func(t *testing.T) []string
	// Ueberschreibe ersetzt den Inhalt der Datei einer Referenz an Ort und Stelle.
	Ueberschreibe func(t *testing.T, ref string, inhalt []byte)
	// ErsetzeDurchVerzeichnis ersetzt die Datei einer Referenz durch ein Verzeichnis.
	ErsetzeDurchVerzeichnis func(t *testing.T, ref string)
	// LegeFremdesAb legt eine Datei ab, die keine Referenz ist.
	LegeFremdesAb func(t *testing.T)
}

// NeueAblage liefert je Aufruf eine frische Ablage.
type NeueAblage func(t *testing.T) BlobAblage

func refVon(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}

func lege(t *testing.T, a BlobAblage, inhalt []byte) string {
	t.Helper()
	w, err := a.Blobs.Begin(context.Background(), 1<<20)
	ok(t, err)
	_, err = w.Write(inhalt)
	ok(t, err)
	ref, n, err := w.Commit()
	ok(t, err)
	if ref != refVon(inhalt) || n != int64(len(inhalt)) {
		t.Fatalf("Commit lieferte %s, %d für %q", ref, n, inhalt)
	}
	return ref
}

func leer(t *testing.T, a BlobAblage) {
	t.Helper()
	if d := a.Dateien(t); len(d) != 0 {
		t.Fatalf("Dateien %v, erwartet keine", d)
	}
}

// RunBlobVertrag prüft SPEC-106 Nr. 1 bis 8 und den Kopf des Vertrags.
func RunBlobVertrag(t *testing.T, neu NeueAblage) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, NeueAblage)
	}{
		{"ungültige Referenzen", blobFall1},
		{"Begin mit Grenze 0 und -1", blobFall2},
		{"Grenze genau erreicht und um eins überschritten", blobFall3},
		{"Commit liefert Hash und Länge, leerer Body", blobFall4},
		{"gleicher Inhalt teilt eine Datei, auch gleichzeitig", blobFall5},
		{"Commit ersetzt eine beschädigte Datei", blobFall6},
		{"genau ein Abschluss", blobFall7},
		{"Abort räumt auf und ist wiederholbar", blobFall8},
		{"Abbruch des Kontexts beendet Write und Commit", blobFall9},
		{"Datei erst nach Commit sichtbar", blobFall10},
		{"Open und Verify prüfen den Hash", blobFall11},
		{"List sortiert und ignoriert Fremdes", blobFall12},
		{"Remove, unbekannt ist kein Fehler", blobFall13},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", blobFall14},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func blobFall1(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	gut := refVon([]byte("x"))
	for _, ref := range map[string]string{
		"leer":          "",
		"ohne Präfix":   gut[len("sha256:"):],
		"Großbuchstabe": "sha256:" + gut[len("sha256:"):len("sha256:")+63] + "A",
		"Pfadteil":      "sha256:../" + gut[len("sha256:")+3:],
		"kurz":          gut[:len(gut)-1],
		"lang":          gut + "0",
		"anderer Hash":  "sha1:" + gut[len("sha256:"):],
	} {
		_, err := a.Blobs.Open(ctx, ref)
		erwarte(t, err, model.ErrUngueltig)
		erwarte(t, a.Blobs.Verify(ctx, ref), model.ErrUngueltig)
		erwarte(t, a.Blobs.Remove(ctx, ref), model.ErrUngueltig)
	}
	// Die Prüfung der Referenz geht der des Kontexts voraus.
	abgebrochen, cancel := context.WithCancel(ctx)
	cancel()
	_, err := a.Blobs.Open(abgebrochen, "x")
	erwarte(t, err, model.ErrUngueltig)
}

func blobFall2(t *testing.T, neu NeueAblage) {
	a := neu(t)
	for _, g := range []int64{0, -1} {
		_, err := a.Blobs.Begin(context.Background(), g)
		erwarte(t, err, model.ErrUngueltig)
	}
	abgebrochen, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Blobs.Begin(abgebrochen, 0)
	erwarte(t, err, model.ErrUngueltig)
	leer(t, a)
}

func blobFall3(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	w, err := a.Blobs.Begin(ctx, 4)
	ok(t, err)
	_, err = w.Write([]byte("abcd"))
	ok(t, err)
	_, _, err = w.Commit()
	ok(t, err)

	w, err = a.Blobs.Begin(ctx, 4)
	ok(t, err)
	_, err = w.Write([]byte("abc"))
	ok(t, err)
	_, err = w.Write([]byte("de"))
	erwarte(t, err, outbound.ErrZuGross)
	_, err = w.Write([]byte("a"))
	erwarte(t, err, outbound.ErrAbgeschlossen)
	_, _, err = w.Commit()
	erwarte(t, err, outbound.ErrAbgeschlossen)
	if d := a.Dateien(t); len(d) != 1 {
		t.Fatalf("Dateien %v, erwartet nur den ersten Blob", d)
	}
}

func blobFall4(t *testing.T, neu NeueAblage) {
	a := neu(t)
	lege(t, a, []byte("{\"a\":1}"))
	ref := lege(t, a, nil)
	r, err := a.Blobs.Open(context.Background(), ref)
	ok(t, err)
	defer func() { _ = r.Close() }()
	b, err := io.ReadAll(r)
	ok(t, err)
	if len(b) != 0 {
		t.Fatalf("leerer Body las %q", b)
	}
}

func blobFall5(t *testing.T, neu NeueAblage) {
	a := neu(t)
	ref := lege(t, a, []byte("gleich"))
	if lege(t, a, []byte("gleich")) != ref {
		t.Fatal("andere Referenz für gleiche Bytes")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w, err := a.Blobs.Begin(context.Background(), 100)
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = w.Write([]byte("gleich"))
			if r, _, err := w.Commit(); err != nil || r != ref {
				t.Errorf("Commit %s, %v", r, err)
			}
		}()
	}
	wg.Wait()
	refs, err := a.Blobs.List(context.Background())
	ok(t, err)
	if len(refs) != 1 || refs[0] != ref || len(a.Dateien(t)) != 1 {
		t.Fatalf("Referenzen %v, Dateien %v, erwartet eine", refs, a.Dateien(t))
	}
}

func blobFall6(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	ref := lege(t, a, []byte("heil"))
	a.Ueberschreibe(t, ref, []byte("kaputt"))
	_, err := a.Blobs.Open(ctx, ref)
	erwarte(t, err, outbound.ErrBlobBeschaedigt)
	lege(t, a, []byte("heil"))
	ok(t, a.Blobs.Verify(ctx, ref))
}

func blobFall7(t *testing.T, neu NeueAblage) {
	a := neu(t)
	w, err := a.Blobs.Begin(context.Background(), 10)
	ok(t, err)
	_, err = w.Write([]byte("a"))
	ok(t, err)
	_, _, err = w.Commit()
	ok(t, err)
	_, err = w.Write([]byte("b"))
	erwarte(t, err, outbound.ErrAbgeschlossen)
	_, _, err = w.Commit()
	erwarte(t, err, outbound.ErrAbgeschlossen)
	ok(t, w.Abort())
	if d := a.Dateien(t); len(d) != 1 {
		t.Fatalf("Dateien %v, Abort nach Commit ließ den Blob nicht bestehen", d)
	}

	w, err = a.Blobs.Begin(context.Background(), 10)
	ok(t, err)
	ok(t, w.Abort())
	_, err = w.Write([]byte("b"))
	erwarte(t, err, outbound.ErrAbgeschlossen)
	_, _, err = w.Commit()
	erwarte(t, err, outbound.ErrAbgeschlossen)
}

func blobFall8(t *testing.T, neu NeueAblage) {
	a := neu(t)
	w, err := a.Blobs.Begin(context.Background(), 10)
	ok(t, err)
	_, err = w.Write([]byte("a"))
	ok(t, err)
	ok(t, w.Abort())
	ok(t, w.Abort())
	leer(t, a)
}

func blobFall9(t *testing.T, neu NeueAblage) {
	a := neu(t)
	ctx, cancel := context.WithCancel(context.Background())
	w, err := a.Blobs.Begin(ctx, 10)
	ok(t, err)
	_, err = w.Write([]byte("a"))
	ok(t, err)
	cancel()
	_, err = w.Write([]byte("b"))
	erwarte(t, err, context.Canceled)
	leer(t, a)

	ctx, cancel = context.WithCancel(context.Background())
	w, err = a.Blobs.Begin(ctx, 10)
	ok(t, err)
	_, err = w.Write([]byte("a"))
	ok(t, err)
	cancel()
	_, _, err = w.Commit()
	erwarte(t, err, context.Canceled)
	leer(t, a)
}

func blobFall10(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	w, err := a.Blobs.Begin(ctx, 10)
	ok(t, err)
	_, err = w.Write([]byte("a"))
	ok(t, err)
	refs, err := a.Blobs.List(ctx)
	ok(t, err)
	if len(refs) != 0 {
		t.Fatalf("offener Writer sichtbar: %v", refs)
	}
	_, err = a.Blobs.Open(ctx, refVon([]byte("a")))
	erwarte(t, err, outbound.ErrBlobFehlt)
	ok(t, w.Abort())
}

func blobFall11(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	ref := lege(t, a, []byte("inhalt"))
	r, err := a.Blobs.Open(ctx, ref)
	ok(t, err)
	b, err := io.ReadAll(r)
	ok(t, err)
	ok(t, r.Close())
	if !bytes.Equal(b, []byte("inhalt")) {
		t.Fatalf("gelesen %q", b)
	}
	ok(t, a.Blobs.Verify(ctx, ref))

	fehlt := refVon([]byte("fehlt"))
	_, err = a.Blobs.Open(ctx, fehlt)
	erwarte(t, err, outbound.ErrBlobFehlt)
	erwarte(t, a.Blobs.Verify(ctx, fehlt), outbound.ErrBlobFehlt)

	for name, inhalt := range map[string][]byte{"verändert": []byte("inhAlt"), "gekürzt": []byte("inha"), "leer": nil, "verlängert": []byte("inhalt!")} {
		a.Ueberschreibe(t, ref, inhalt)
		_, err = a.Blobs.Open(ctx, ref)
		if err == nil {
			t.Fatalf("%s: Open lieferte einen Leser", name)
		}
		erwarte(t, err, outbound.ErrBlobBeschaedigt)
		erwarte(t, a.Blobs.Verify(ctx, ref), outbound.ErrBlobBeschaedigt)
	}

	ref2 := lege(t, a, []byte("zwei"))
	a.ErsetzeDurchVerzeichnis(t, ref2)
	_, err = a.Blobs.Open(ctx, ref2)
	erwarte(t, err, outbound.ErrBlobBeschaedigt)
}

func blobFall12(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	refs, err := a.Blobs.List(ctx)
	ok(t, err)
	if refs == nil || len(refs) != 0 {
		t.Fatalf("leere Ablage lieferte %#v", refs)
	}
	var soll []string
	for _, s := range []string{"c", "a", "b", "d"} {
		soll = append(soll, lege(t, a, []byte(s)))
	}
	sort.Strings(soll)
	a.LegeFremdesAb(t)
	w, err := a.Blobs.Begin(ctx, 10)
	ok(t, err)
	defer func() { _ = w.Abort() }()
	refs, err = a.Blobs.List(ctx)
	ok(t, err)
	if len(refs) != len(soll) {
		t.Fatalf("Referenzen %v, erwartet %v", refs, soll)
	}
	for i := range soll {
		if refs[i] != soll[i] {
			t.Fatalf("Referenzen %v, erwartet %v", refs, soll)
		}
	}
}

func blobFall13(t *testing.T, neu NeueAblage) {
	ctx := context.Background()
	a := neu(t)
	ref := lege(t, a, []byte("weg"))
	andere := lege(t, a, []byte("bleibt"))
	ok(t, a.Blobs.Remove(ctx, ref))
	_, err := a.Blobs.Open(ctx, ref)
	erwarte(t, err, outbound.ErrBlobFehlt)
	ok(t, a.Blobs.Remove(ctx, ref))
	ok(t, a.Blobs.Remove(ctx, refVon([]byte("nie da"))))
	ok(t, a.Blobs.Verify(ctx, andere))
}

func blobFall14(t *testing.T, neu NeueAblage) {
	a := neu(t)
	ref := lege(t, a, []byte("da"))
	vorher := a.Dateien(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Blobs.Begin(ctx, 10)
	erwarte(t, err, context.Canceled)
	_, err = a.Blobs.Open(ctx, ref)
	erwarte(t, err, context.Canceled)
	erwarte(t, a.Blobs.Verify(ctx, ref), context.Canceled)
	_, err = a.Blobs.List(ctx)
	erwarte(t, err, context.Canceled)
	erwarte(t, a.Blobs.Remove(ctx, ref), context.Canceled)
	if nachher := a.Dateien(t); len(nachher) != len(vorher) || nachher[0] != vorher[0] {
		t.Fatalf("Dateien %v, vorher %v", nachher, vorher)
	}
	ok(t, a.Blobs.Verify(context.Background(), ref))
}
