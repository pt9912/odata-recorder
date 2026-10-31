package cli_test

import (
	"bytes"
	"go/build"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/cli"
)

// C1: Menge W = Aufrufmuster {nil, leere Liste, version, start} × Start {gesetzt, nil} = 8.
func TestKommandoWahl(t *testing.T) {
	muster := []struct {
		name string
		args []string
	}{
		{"nil", nil},
		{"leere Liste", []string{}},
		{"version", []string{"version"}},
		{"start", []string{"start"}},
	}
	n := 0
	for _, m := range muster {
		for _, gesetzt := range []bool{true, false} {
			n++
			t.Run(m.name+map[bool]string{true: "/Start gesetzt", false: "/Start nil"}[gesetzt], func(t *testing.T) {
				var start *FakeStart
				if gesetzt {
					start = NewFakeStart(t, 0)
				}
				var out, errw bytes.Buffer
				code := cli.Run(m.args, konfig("1.2.3", start, NewFakeSignale(), &out, &errw))
				wantCode, wantOut, wantErr, wantAufrufe := 0, "odata-recorder 1.2.3\n", "", 0
				if m.name == "start" {
					wantOut = ""
					if gesetzt {
						wantAufrufe = 1
					} else {
						wantCode, wantErr = 2, "Unbekanntes Kommando: start\n"
					}
				}
				aufrufe := 0
				if start != nil {
					aufrufe = start.Zustand().Aufrufe
				}
				if code != wantCode || out.String() != wantOut || errw.String() != wantErr || aufrufe != wantAufrufe {
					t.Errorf("Code %d, stdout %q, stderr %q, Aufrufe %d; erwartet %d, %q, %q, %d",
						code, out.String(), errw.String(), aufrufe, wantCode, wantOut, wantErr, wantAufrufe)
				}
			})
		}
	}
	if n != 8 {
		t.Fatalf("Menge W hat %d Elemente, erwartet 8", n)
	}
}

// C2: Menge A × P (24 Formen × 4 Positionen) und 4 Sonderfälle = 100.
func TestAblehnung(t *testing.T) {
	nbsp := string([]byte{0xC2, 0xA0})
	bom := string([]byte{0xEF, 0xBB, 0xBF})
	nul := string([]byte{0})
	formen := []string{
		"foo", "", " ", "\t", "Start", "START", "Version", "VERSION", "start ", " start", "version\n",
		"starts", "sta", "versio", "versionx", "--help", "-h", "help", "--version", "-v", "--",
		"version" + nbsp, bom + "start", "version" + nul,
	}
	if len(formen) != 24 {
		t.Fatalf("Menge A hat %d Formen, erwartet 24", len(formen))
	}
	type fall struct {
		name       string
		args       []string
		startNil   bool
		wantErster string
	}
	var faelle []fall
	for _, f := range formen {
		faelle = append(faelle,
			fall{"P1", []string{f}, false, f},
			fall{"P2", []string{f, "x"}, false, f},
			fall{"P3", []string{"version", f}, false, "version"},
			fall{"P4", []string{"start", f}, false, "start"},
		)
	}
	if len(faelle) != 96 {
		t.Fatalf("Menge A × P hat %d Elemente, erwartet 96", len(faelle))
	}
	faelle = append(faelle,
		fall{"S1", []string{"start"}, true, "start"},
		fall{"S2", []string{"start", "x"}, true, "start"},
		fall{"S3", []string{"version", "x", "y"}, false, "version"},
		fall{"S4", []string{"start", "x", "y"}, false, "start"},
	)
	if len(faelle) != 100 {
		t.Fatalf("Menge hat %d Elemente, erwartet 100", len(faelle))
	}
	for _, f := range faelle {
		t.Run(f.name+"/"+strings.Join(f.args, "|"), func(t *testing.T) {
			var start *FakeStart
			if !f.startNil {
				start = NewFakeStart(t, 0)
			}
			sig := NewFakeSignale()
			var out, errw bytes.Buffer
			code := cli.Run(f.args, konfig("1.2.3", start, sig, &out, &errw))
			if code != 2 {
				t.Errorf("Code %d, erwartet 2", code)
			}
			if out.Len() != 0 {
				t.Errorf("stdout %q, erwartet leer", out.String())
			}
			if want := "Unbekanntes Kommando: " + f.wantErster + "\n"; errw.String() != want {
				t.Errorf("stderr %q, erwartet %q", errw.String(), want)
			}
			if start != nil && start.Zustand().Aufrufe != 0 {
				t.Errorf("Start %d-mal gerufen, erwartet 0", start.Zustand().Aufrufe)
			}
			if sig.Anmeldungen() != 0 || sig.Abmeldungen() != 0 {
				t.Errorf("Anmeldungen %d, Abmeldungen %d, erwartet 0, 0", sig.Anmeldungen(), sig.Abmeldungen())
			}
		})
	}
}

// C3: Menge Version {5} × {nil, version} = 10.
func TestVersionAusgabe(t *testing.T) {
	versionen := []string{"1.2.3", "dev", "", " x ", "a\tb"}
	n := 0
	for _, v := range versionen {
		for _, args := range [][]string{nil, {"version"}} {
			n++
			start := NewFakeStart(t, 0)
			sig := NewFakeSignale()
			var out, errw bytes.Buffer
			code := cli.Run(args, konfig(v, start, sig, &out, &errw))
			if want := "odata-recorder " + v + "\n"; code != 0 || out.String() != want || errw.Len() != 0 {
				t.Errorf("Version %q, args %v: Code %d, stdout %q, stderr %q; erwartet 0, %q, leer", v, args, code, out.String(), errw.String(), want)
			}
			if sig.Anmeldungen() != 0 || start.Zustand().Aufrufe != 0 {
				t.Errorf("Version %q, args %v: Anmeldungen %d, Aufrufe %d; erwartet 0, 0", v, args, sig.Anmeldungen(), start.Zustand().Aufrufe)
			}
		}
	}
	if n != 10 {
		t.Fatalf("Menge hat %d Elemente, erwartet 10", n)
	}
}

// C4: Writer {nil, Schreibfehler} × Seite {stdout, stderr, beide} × Kommando {version, foo, start} = 18.
func TestWriterRandformen(t *testing.T) {
	kommandos := []struct {
		name string
		args []string
		code int
	}{
		{"version", []string{"version"}, 0},
		{"foo", []string{"foo"}, 2},
		{"start", []string{"start"}, 0},
	}
	seiten := []struct {
		name         string
		stdout, errw bool
	}{{"stdout", true, false}, {"stderr", false, true}, {"beide", true, true}}
	n := 0
	for _, fehler := range []bool{false, true} {
		for _, s := range seiten {
			for _, kd := range kommandos {
				n++
				t.Run(map[bool]string{false: "nil", true: "Fehler"}[fehler]+"/"+s.name+"/"+kd.name, func(t *testing.T) {
					start := NewFakeStart(t, 0)
					var gutOut, gutErr bytes.Buffer
					var out, errw io.Writer = &gutOut, &gutErr
					fw := &fehlerWriter{}
					waehle := func(an bool, gut io.Writer) io.Writer {
						switch {
						case !an:
							return gut
						case fehler:
							return fw
						default:
							return nil
						}
					}
					out, errw = waehle(s.stdout, out), waehle(s.errw, errw)
					if code := cli.Run(kd.args, konfig("1.2.3", start, NewFakeSignale(), out, errw)); code != kd.code {
						t.Errorf("Code %d, erwartet %d", code, kd.code)
					}
					if kd.name != "start" {
						return
					}
					z := start.Zustand()
					if z.Aufrufe != 1 || z.Stdout[0] == nil || z.Stderr[0] == nil {
						t.Fatalf("Start: Aufrufe %d, stdout %v, stderr %v; erwartet 1, nicht nil", z.Aufrufe, z.Stdout, z.Stderr)
					}
					if fehler && s.stdout && z.Stdout[0] != io.Writer(fw) {
						t.Errorf("Start erhielt nicht den Fehler-Writer als stdout")
					}
				})
			}
		}
	}
	if n != 18 {
		t.Fatalf("Menge Z hat %d Elemente, erwartet 18", n)
	}
}

// C5: vier Eigenschaften des einen Aufrufs von Start.
func TestStartAufruf(t *testing.T) {
	start := NewFakeStart(t, 0)
	var out, errw bytes.Buffer
	if code := cli.Run([]string{"start"}, konfig("1.2.3", start, NewFakeSignale(), &out, &errw)); code != 0 {
		t.Fatalf("Code %d, erwartet 0", code)
	}
	z := start.Zustand()
	t.Run("genau ein Aufruf", func(t *testing.T) {
		if z.Aufrufe != 1 || z.Rueckkehren != 1 {
			t.Errorf("Aufrufe %d, Rückkehren %d; erwartet 1, 1", z.Aufrufe, z.Rueckkehren)
		}
	})
	if z.Aufrufe != 1 {
		return
	}
	t.Run("Writer aus Konfig", func(t *testing.T) {
		if z.Stdout[0] != io.Writer(&out) || z.Stderr[0] != io.Writer(&errw) {
			t.Errorf("Start erhielt nicht die Writer aus Konfig (stdout %v, stderr %v)", z.Stdout[0], z.Stderr[0])
		}
	})
	t.Run("abbrechbarer, nicht abgebrochener Kontext", func(t *testing.T) {
		if z.Ctx[0].Done() == nil || z.ErrBeiAufruf[0] != nil {
			t.Errorf("Done %v, Err beim Aufruf %v; erwartet Done ungleich nil, Err gleich nil", z.Ctx[0].Done(), z.ErrBeiAufruf[0])
		}
	})
	t.Run("Adapter schreibt nichts", func(t *testing.T) {
		if out.Len() != 0 || errw.Len() != 0 {
			t.Errorf("stdout %q, stderr %q; erwartet leer", out.String(), errw.String())
		}
	})
}

// C6: Rückgaben {0, 1, 2, 3, 255, 256, 257, -1, MaxInt, MinInt} = 10.
func TestExitAbbildung(t *testing.T) {
	rueckgaben := []int{0, 1, 2, 3, 255, 256, 257, -1, math.MaxInt, math.MinInt}
	if len(rueckgaben) != 10 {
		t.Fatalf("Menge hat %d Elemente, erwartet 10", len(rueckgaben))
	}
	for _, r := range rueckgaben {
		want := 1
		if r == 0 {
			want = 0
		}
		start := NewFakeStart(t, r)
		if code := cli.Run([]string{"start"}, konfig("1.2.3", start, NewFakeSignale(), nil, nil)); code != want {
			t.Errorf("Rückgabe %d: Code %d, erwartet %d", r, code, want)
		}
	}
}

// C6 als Klasse: jede Rückgabe außer 0 ergibt 1. Orakel: 0 bleibt 0, sonst 1. Elemente: die Werte
// von -1024 bis 1024 (2049), je Zweierpotenz 2^k für k = 11 bis 62 die Werte ±(2^k-1), ±2^k, ±(2^k+1)
// (52 × 6 = 312) und MaxInt, MaxInt-1, MinInt, MinInt+1 (4): 2365.
func TestExitAbbildungKlasse(t *testing.T) {
	var werte []int
	for v := -1024; v <= 1024; v++ {
		werte = append(werte, v)
	}
	for k := 11; k <= 62; k++ {
		p := 1 << k
		werte = append(werte, p-1, p, p+1, -(p - 1), -p, -(p + 1))
	}
	werte = append(werte, math.MaxInt, math.MaxInt-1, math.MinInt, math.MinInt+1)
	if len(werte) != 2365 {
		t.Fatalf("Klasse hat %d Elemente, erwartet 2365", len(werte))
	}
	for _, r := range werte {
		want := 1
		if r == 0 {
			want = 0
		}
		start := NewFakeStart(t, r)
		if code := cli.Run([]string{"start"}, konfig("1.2.3", start, NewFakeSignale(), nil, nil)); code != want {
			t.Errorf("Rückgabe %d: Code %d, erwartet %d", r, code, want)
		}
	}
}

// C13: Das Paket importiert nur die Standardbibliothek. Menge: alle Importe der Nicht-Test-Dateien.
func TestNurStandardbibliothek(t *testing.T) {
	p, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("ImportDir: %v", err)
	}
	if len(p.Imports) < 5 {
		t.Fatalf("Menge der Importe hat %d Elemente, erwartet mindestens 5", len(p.Imports))
	}
	for _, imp := range p.Imports {
		erster, _, _ := strings.Cut(imp, "/")
		if strings.Contains(erster, ".") {
			t.Errorf("Import %q ist nicht aus der Standardbibliothek", imp)
		}
	}
}

// C2 als Klasse: Rest vor und hinter dem Kommando (jedes Byte 0x00 bis 0xFF), jede Schreibweise mit
// einem umgeschalteten Buchstaben und jedes echte Präfix. Orakel: Verstoß, Meldung nennt das erste Argument.
// Elemente: 2 Kommandos × 2 Seiten × 256 Bytes = 1024, dazu 5 + 7 umgeschaltete Buchstaben und
// 4 + 6 echte Präfixe = 1046; dazu als Zusatzargument hinter dem Kommando jedes Byte,
// 2 Kommandos × 256 = 512; zusammen 1558.
func TestAblehnungKlasse(t *testing.T) {
	var aufrufe [][]string
	for _, w := range []string{"start", "version"} {
		for b := 0; b < 256; b++ {
			bs := string([]byte{byte(b)})
			aufrufe = append(aufrufe, []string{w + bs}, []string{bs + w}, []string{w, bs})
		}
		for i := range w {
			aufrufe = append(aufrufe, []string{w[:i] + strings.ToUpper(w[i:i+1]) + w[i+1:]})
			if i > 0 {
				aufrufe = append(aufrufe, []string{w[:i]})
			}
		}
	}
	if len(aufrufe) != 1558 {
		t.Fatalf("Klasse hat %d Elemente, erwartet 1558", len(aufrufe))
	}
	for _, args := range aufrufe {
		start := NewFakeStart(t, 0)
		sig := NewFakeSignale()
		var out, errw bytes.Buffer
		code := cli.Run(args, konfig("1.2.3", start, sig, &out, &errw))
		if code != 2 || out.Len() != 0 || errw.String() != "Unbekanntes Kommando: "+args[0]+"\n" ||
			start.Zustand().Aufrufe != 0 || sig.Anmeldungen() != 0 {
			t.Errorf("Aufruf %q: Code %d, stdout %q, stderr %q, Aufrufe %d, Anmeldungen %d",
				args, code, out.String(), errw.String(), start.Zustand().Aufrufe, sig.Anmeldungen())
		}
	}
}

// C3 als Klasse: Die Version erscheint unverändert, was immer sie enthält. Orakel: stdout gleich
// "odata-recorder " + Version + Zeilenende. Elemente: jedes Byte 0x00 bis 0xFF allein, davor, dahinter und in
// der Mitte (4 · 256 = 1024), dazu 5 Mehrbyte-Folgen (é, NBSP, BOM, U+FFFD, abgeschnittene Folge),
// 3 Längen (100, 1000, 100000 Zeichen) und 3 Formatfolgen (%s, %d, %!): 1035.
func TestVersionKlasse(t *testing.T) {
	var versionen []string
	for b := 0; b < 256; b++ {
		bs := string([]byte{byte(b)})
		versionen = append(versionen, bs, "1"+bs, bs+"1", "1"+bs+"2")
	}
	versionen = append(versionen,
		string([]byte{0xC3, 0xA9}), string([]byte{0xC2, 0xA0}), string([]byte{0xEF, 0xBB, 0xBF}),
		string([]byte{0xEF, 0xBF, 0xBD}), "a"+string([]byte{0xC3}),
		strings.Repeat("v", 100), strings.Repeat("v", 1000), strings.Repeat("v", 100000),
		"%s", "%d", "%!",
	)
	if len(versionen) != 1035 {
		t.Fatalf("Klasse hat %d Elemente, erwartet 1035", len(versionen))
	}
	for _, v := range versionen {
		var out, errw bytes.Buffer
		code := cli.Run([]string{"version"}, konfig(v, nil, nil, &out, &errw))
		if want := "odata-recorder " + v + "\n"; code != 0 || out.String() != want || errw.Len() != 0 {
			t.Errorf("Version %q: Code %d, stdout %q, stderr %q; erwartet 0, %q, leer", v, code, out.String(), errw.String(), want)
		}
	}
}
