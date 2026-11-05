package capture_test

import (
	"context"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// Z17 und Z13 (SPEC-119 Nr. 2 und Nr. 8): die Abweisung im Tee und das Durchreichen der Ziele.

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

// Z17: Tee weist ab, ohne ein Byte zu lesen und ohne Goroutine; gültige Kombinationen werden eröffnet.
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

// Z13: der Eingang des Tee reicht die Ziele an den Redaktor.
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

// Z13: gleichzeitige Aufnahmen mit und ohne Treffer teilen keine Bytes.
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
		roh := "wert-" + strconv.Itoa(i) + strings.Repeat("x", 1+i)
		a, err := p.Tee(context.Background(), quelle(roh), ein)
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
