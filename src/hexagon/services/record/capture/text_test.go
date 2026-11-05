package capture_test

import (
	"context"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// Die Abweisung im Tee und das Durchreichen der Ziele (SPEC-119 Nr. 2 und Nr. 8).

func textAntwort(ziel string, ct ...string) capture.Eingang {
	e := capture.Eingang{Richtung: capture.Response, ContentType: ct}
	switch ziel {
	case "zaehlung":
		e.Zaehlung = true
	case "wert":
		e.Wert, e.Name = true, "Name"
	}
	return e
}

// Tee weist ab (SPEC-119 Nr. 2), ohne ein Byte zu lesen und ohne Goroutine; gültige Kombinationen werden eröffnet.
func TestTeeZielKombinationen(t *testing.T) {
	zahl := 0
	for _, richtung := range []capture.Richtung{capture.Request, capture.Response} {
		for bits := 0; bits < 8; bits++ {
			for _, name := range []string{"", "n"} {
				ein := capture.Eingang{Richtung: richtung, ContentType: []string{"text/plain"},
					Metadata: bits&1 != 0, Zaehlung: bits&2 != 0, Wert: bits&4 != 0, Name: name}
				// Das Orakel: ein Request hat nie ein Ziel; eine Antwort genau eines von
				// Metadata, Zaehlung, Wert (Wert mit Namen) oder keines.
				gueltig := bits == 0 && name == ""
				if richtung == capture.Response {
					gueltig = gueltig || (bits == 1 && name == "") || (bits == 2 && name == "") || (bits == 4 && name == "n")
				}
				zahl++
				t.Run(strconv.Itoa(int(richtung))+"/"+strconv.Itoa(bits)+"/"+name, func(t *testing.T) {
					u := neu(t)
					p := u.pipe()
					var einmal sync.Once
					roh := p.AnhaltenAmStart()
					frei := func() { einmal.Do(roh) }
					defer frei()
					src := quelle("12")
					vorher := runtime.NumGoroutine()
					a, err := p.Tee(t.Context(), src, ein)
					if !gueltig {
						if !errorsIsUngueltig(err) || a != nil {
							t.Errorf("Tee = %v, %v; erwartet ErrUngueltig und keine Aufnahme", a, err)
						}
						if n := src.anzahlReads(); n != 0 {
							t.Errorf("src wurde %d-mal gelesen", n)
						}
						if st := u.blobs.stand(); st.begins != 0 {
							t.Errorf("Begin %d-mal gerufen", st.begins)
						}
						if nachher := runtime.NumGoroutine(); nachher > vorher {
							t.Errorf("Goroutinen %d -> %d: Tee startete eine Goroutine", vorher, nachher)
						}
						return
					}
					if err != nil || a == nil {
						t.Fatalf("Tee = %v, %v; erwartet eine Aufnahme", a, err)
					}
					a.Abbrechen()
					frei()
					_ = abwarten(t, a)
				})
			}
		}
	}
	if zahl != 32 {
		t.Errorf("%d Kombinationen", zahl)
	}
}

// Der Eingang des Tee reicht die Ziele an den Redaktor (SPEC-119 Nr. 8).
func TestTeeReichtZieleDurch(t *testing.T) {
	faelle := []struct {
		name   string
		ein    capture.Eingang
		roh    string
		red    model.RedactionStatus
		inhalt string
	}{
		{"Zaehlung", textAntwort("zaehlung", "text/plain"), "42", model.RedactionNone, "42"},
		{"Wert ohne Treffer", textAntwort("wert", "text/plain"), "Anvil", model.RedactionNone, "Anvil"},
		{"Wert mit Treffer", capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}, Wert: true, Name: "password"},
			"s3cr3t", model.RedactionApplied, model.Marker},
		{"Wert mit Treffer, Schreibweise", capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}, Wert: true, Name: "PASSWORD"},
			"s3cr3t", model.RedactionApplied, model.Marker},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			e := u.lauf(quelle(f.roh), f.ein)
			if got := vollstaendig(t, u, e, f.red); got != f.inhalt {
				t.Errorf("Blob %q, erwartet %q", got, f.inhalt)
			}
			// Dieselbe Antwort ohne Ziel ist unbekannt.
			ohne := f.ein
			ohne.Zaehlung, ohne.Wert, ohne.Name = false, false, ""
			u2 := neu(t)
			verworfen(t, u2.lauf(quelle(f.roh), ohne), capture.Response, capture.Verstoss, capture.GrundTyp)
		})
	}
	// Ein Zeichensatz außerhalb der Menge ist der Grund charset.
	u := neu(t)
	verworfen(t, u.lauf(quelle("42"), textAntwort("zaehlung", "text/plain; charset=latin1")), capture.Response, capture.Verstoss, capture.GrundCharset)
	// Eine Nicht-Ziffer ist syntax, auch bei einem Ziel, das die Aufnahme bekam.
	u = neu(t)
	verworfen(t, u.lauf(quelle("4a"), textAntwort("zaehlung", "text/plain")), capture.Response, capture.Verstoss, capture.GrundSyntax)
}

// utfZeichen sind Zeichen von einem bis vier Bytes.
func utfZeichen() []string { return []string{"x", "\xC3\xA9", "\xE2\x82\xAC", "\xF0\x9F\x98\x80"} }

// Gleichzeitige Aufnahmen mit und ohne Treffer teilen keine Bytes (SPEC-119 Nr. 8): jede trägt
// Zeichen von einem bis vier Bytes, und jede Quelle liefert Reads zu drei Bytes, die Folgen teilen.
func TestTeeTextGleichzeitig(t *testing.T) {
	const n = 24
	u := neu(t)
	p := u.pipe()
	frei := p.AnhaltenAmStart()
	type lauf struct {
		a       *capture.Aufnahme
		inhalt  string
		treffer bool
	}
	laeufe := make([]lauf, n)
	var wg sync.WaitGroup
	for i := range laeufe {
		treffer := i%2 == 0
		ein := capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}, Wert: true, Name: "Name"}
		if treffer {
			ein.Name = "password"
		}
		zeichen := utfZeichen()
		roh := "wert-" + strconv.Itoa(i) + strings.Repeat(zeichen[i/2%len(zeichen)], 1+i)
		a, err := p.Tee(context.Background(), zerlegt([]byte(roh), 3), ein)
		if err != nil {
			t.Fatalf("Tee %d: %v", i, err)
		}
		laeufe[i] = lauf{a: a, inhalt: roh, treffer: treffer}
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = io.Copy(io.Discard, a)
		}()
	}
	frei()
	wg.Wait()
	for i, l := range laeufe {
		e := abwarten(t, l.a)
		soll, red := l.inhalt, model.RedactionNone
		if l.treffer {
			soll, red = model.Marker, model.RedactionApplied
		}
		if got := vollstaendig(t, u, e, red); got != soll {
			t.Errorf("Aufnahme %d: Blob %q, erwartet %q", i, got, soll)
		}
	}
}

// TestTeeTextMarkerZustand: der Stand des Vergleichs mit dem Marker (`mPos`, `mAbweich`) gehört der
// Aufnahme (SPEC-118 Nr. 11 b, Idempotenz). Die Aufnahme A trägt den Marker in drei Reads, ihr
// letzter Read kommt erst, wenn B (der Marker in einem Read) und C (ein anderer Inhalt) fertig sind.
// Teilen die Aufnahmen den Stand, ist er für jede Reihenfolge der Köpfe von A, B und C verändert:
// A oder B gelten dann als Änderung, obwohl sie der Marker sind; die Wirkung hängt weder vom Zufall
// noch vom Detektor ab.
func TestTeeTextMarkerZustand(t *testing.T) {
	u := neu(t)
	p := u.pipe()
	cFertig := make(chan struct{})
	var einmal sync.Once
	defer einmal.Do(func() { close(cFertig) })
	ein := capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}, Wert: true, Name: "password"}
	tee := func(src io.Reader) *capture.Aufnahme {
		a, err := p.Tee(context.Background(), src, ein)
		if err != nil {
			t.Fatalf("Tee: %v", err)
		}
		go func() { _, _ = io.Copy(io.Discard, a) }()
		return a
	}
	a := tee(quelle("[RE", "DAC").haltVor(cFertig, nil, "TED]", nil))
	b := tee(quelle(model.Marker))
	c := tee(quelle("x"))
	eB, eC := abwarten(t, b), abwarten(t, c)
	einmal.Do(func() { close(cFertig) })
	eA := abwarten(t, a)
	for _, f := range []struct {
		name string
		e    capture.Ergebnis
		red  model.RedactionStatus
	}{{"A", eA, model.RedactionNone}, {"B", eB, model.RedactionNone}, {"C", eC, model.RedactionApplied}} {
		if got := vollstaendig(t, u, f.e, f.red); got != model.Marker {
			t.Errorf("%s: Blob %q, erwartet %q", f.name, got, model.Marker)
		}
	}
}

// meldeBlobs meldet den ersten Write des ersten Writers; die Aufnahme, die als erste schreibt, ist
// damit in ihrem Strom angekommen.
type meldeBlobs struct {
	*fakeBlobs
	einmal sync.Once
	erste  func()
}

func (b *meldeBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	w, err := b.fakeBlobs.Begin(ctx, grenze)
	if err != nil {
		return nil, err
	}
	return &meldeWriter{CaptureWriter: w, b: b}, nil
}

type meldeWriter struct {
	outbound.CaptureWriter
	b *meldeBlobs
}

func (w *meldeWriter) Write(p []byte) (int, error) {
	w.b.einmal.Do(w.b.erste)
	return w.CaptureWriter.Write(p)
}

// TestTeeTextUTF8Zustand: der Zustand der UTF-8-Prüfung (offene Folgebytes und Bereich des nächsten)
// gehört der Aufnahme (SPEC-119 Nr. 8, SPEC-118 Nr. 11 b). Die Aufnahme A liefert ein Zeichen bis in
// die Folge hinein und wartet; erst wenn ihr Teil geschrieben ist, läuft die Aufnahme B ganz durch,
// und erst wenn B fertig ist, kommt der Rest von A. Teilen A und B einen Teil des Zustands, ist er
// nach B verändert: die offenen Folgebytes sind verbraucht, der Bereich ist zurückgesetzt. Eine
// gültige Folge von A bricht dann, eine ungültige, die allein der enge Bereich ihres Anfangsbytes
// abweist, geht durch. Die Reihenfolge ist durch Kanäle bestimmt: die Wirkung hängt weder vom Zufall
// noch vom Detektor ab. A trägt eines der vier Anfangsbytes mit engem Bereich (E0, ED, F0, F4), je
// mit einem Folgebyte an jeder Grenze des Bereichs und je mit einem knapp außerhalb; B ist ein
// Zeichen von zwei, drei oder vier Bytes.
func TestTeeTextUTF8Zustand(t *testing.T) {
	type el struct {
		name, kopf, rest string
		gueltig          bool
	}
	gruppe := []el{
		{"E0 A0", "a\xE0", "\xA0\x80z", true}, {"E0 BF", "a\xE0", "\xBF\xBFz", true},
		{"ED 80", "a\xED", "\x80\x80z", true}, {"ED 9F", "a\xED", "\x9F\xBFz", true},
		{"F0 90", "a\xF0", "\x90\x80\x80z", true}, {"F0 BF", "a\xF0", "\xBF\xBF\xBFz", true},
		{"F4 80", "a\xF4", "\x80\x80\x80z", true}, {"F4 8F", "a\xF4", "\x8F\xBF\xBFz", true},
		{"E0 9F", "a\xE0", "\x9F\x80z", false}, {"ED A0", "a\xED", "\xA0\x80z", false},
		{"F0 8F", "a\xF0", "\x8F\x80\x80z", false}, {"F4 90", "a\xF4", "\x90\x80\x80z", false},
	}
	if len(gruppe) != 12 {
		t.Fatalf("%d Aufnahmen", len(gruppe))
	}
	zeichen := utfZeichen()
	bs := []string{"a" + zeichen[1] + "z", "a" + zeichen[2] + "z", "a" + zeichen[3] + "z"}
	ein := capture.Eingang{Richtung: capture.Response, ContentType: []string{"text/plain"}, Wert: true, Name: "Name"}
	zahl := 0
	for _, g := range gruppe {
		if utf8.ValidString(g.kopf+g.rest) != g.gueltig {
			t.Fatalf("%s: utf8.ValidString sagt %v", g.name, !g.gueltig)
		}
		for _, b := range bs {
			zahl++
			t.Run(g.name+"/"+strconv.Itoa(len(b)-2), func(t *testing.T) {
				aDa, bFertig := make(chan struct{}), make(chan struct{})
				var einmal, einmalA sync.Once
				defer einmal.Do(func() { close(bFertig) })
				angekommen := func() { einmalA.Do(func() { close(aDa) }) }
				u := neu(t, func(k *capture.Konfig) {
					k.Blobs = &meldeBlobs{fakeBlobs: k.Blobs.(*fakeBlobs), erste: angekommen}
				})
				p := u.pipe()
				a, err := p.Tee(context.Background(), quelle(g.kopf).haltVor(bFertig, nil, g.rest, nil), ein)
				if err != nil {
					t.Fatalf("Tee A: %v", err)
				}
				go func() { _, _ = io.Copy(io.Discard, a) }()
				// Endet A vor ihrem ersten Write (eine Verletzung des Vertrags), wartet B nicht für immer.
				go func() { _ = a.Ergebnis(); angekommen() }()
				bq := (&skript{}).haltVor(aDa, nil, b, nil)
				bAuf, err := p.Tee(context.Background(), bq, ein)
				if err != nil {
					t.Fatalf("Tee B: %v", err)
				}
				go func() { _, _ = io.Copy(io.Discard, bAuf) }()
				eB := abwarten(t, bAuf)
				einmal.Do(func() { close(bFertig) })
				eA := abwarten(t, a)
				if got := vollstaendig(t, u, eB, model.RedactionNone); got != b {
					t.Errorf("B: Blob %q, erwartet %q", got, b)
				}
				if !g.gueltig {
					verworfen(t, eA, capture.Response, capture.Verstoss, capture.GrundUTF8)
					return
				}
				if eA.Ausgang != capture.Vollstaendig {
					t.Fatalf("A: %v/%q, erwartet Vollstaendig", eA.Ausgang, eA.Grund)
				}
				if got := vollstaendig(t, u, eA, model.RedactionNone); got != g.kopf+g.rest {
					t.Errorf("A: Blob %q, erwartet %q", got, g.kopf+g.rest)
				}
			})
		}
	}
	if zahl != 36 {
		t.Errorf("%d Paare", zahl)
	}
}
