package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	ergRot       = "rot"
	ergRotPaket  = "rot-paket"
	ergGruen     = "gruen"
	ergFlattern  = "flatternd"
	ergUebersetz = "uebersetzungsfehler"
	ergAbbruch   = "abbruch"

	gleichzeitig = 4
	diagZeilen   = 40
)

// Konfig sind die Zahlen und Pfade eines Aufrufs (SPEC-210 Nr. 8, 9).
type Konfig struct {
	Baum, Ablage string
	TestZeit     time.Duration // -test.timeout der Test-Prozesse
	Gnade        time.Duration // Zeit über TestZeit, bis das Werkzeug den Prozess beendet
	CacheGrenze  int64
	CPUs         int
}

// Werk führt eine Liste von Mutationen aus.
type Werk struct {
	Konfig
	P     Prozesse
	D     Dateien
	Last  func() string
	Jetzt func() time.Time
	Aus   io.Writer
	Fehl  io.Writer
}

// Erg ist das Ergebnis einer Mutation (SPEC-210 Nr. 1 b).
type Erg struct {
	Name, E, Test, Paket string
	Prozesse             int
	Mit, Ohne            string
	Dauer                time.Duration
	Last                 string
	CPUs                 int
	Diag                 string
}

// ausgeben schreibt die Ergebniszeile und die Diagnose (SPEC-210 Nr. 1 b, c).
func (w *Werk) ausgeben(e Erg) {
	_, _ = fmt.Fprintf(w.Aus, "mutation %s ergebnis=%s test=%s paket=%s prozesse=%d mit-detektor=%s ohne-detektor=%s dauer=%.1f last=%s/%d\n",
		e.Name, e.E, e.Test, e.Paket, e.Prozesse, e.Mit, e.Ohne, e.Dauer.Seconds(), e.Last, e.CPUs)
	switch e.E {
	case ergUebersetz, ergRotPaket, ergFlattern, ergAbbruch:
		zeilen := strings.Split(strings.TrimRight(e.Diag, "\n"), "\n")
		for _, z := range zeilen[:min(len(zeilen), diagZeilen)] {
			_, _ = fmt.Fprintf(w.Fehl, "%s: %s\n", e.Name, z)
		}
	}
}

func dauerArg(d time.Duration) string {
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", d/time.Minute)
	}
	return d.String()
}

// relInModul liefert einen Pfad der Wurzel relativ zum Modul.
func relInModul(m Mutation, p string) string {
	if m.Modul.Dir == "" {
		return p
	}
	return strings.TrimPrefix(p, m.Modul.Dir+"/")
}

// paketArg ist das Paket als Argument von `go test -c`, relativ zur Wurzel des Moduls.
func paketArg(m Mutation) string {
	if m.Paket == m.Modul.Dir {
		return "."
	}
	return "./" + relInModul(m, m.Paket)
}

// wendeAn ersetzt `alt` durch `neu` in der Kopie (SPEC-210 Nr. 4 b).
func wendeAn(kopie string, m Mutation) error {
	p := unter(kopie, relInModul(m, m.Datei))
	i, err := os.Stat(p)
	if err != nil {
		return err
	}
	roh, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if strings.Count(string(roh), m.Alt) != 1 {
		return fmt.Errorf("alt trifft nicht genau einmal in %s", m.Datei)
	}
	return os.WriteFile(p, []byte(strings.Replace(string(roh), m.Alt, m.Neu, 1)), i.Mode().Perm())
}

func (w *Werk) testArgs(test string) []string {
	a := []string{"-test.count=1", "-test.timeout=" + dauerArg(w.TestZeit)}
	if test != "" {
		a = append([]string{"-test.run", "^" + test + "$"}, a...)
	}
	return a
}

// bauErgebnis liest das Ergebnis eines Baus der Mutation; leer heißt: gebaut.
func bauErgebnis(a Ausgang) string {
	switch {
	case a.Abgebrochen || a.KeinBinary:
		return ergAbbruch
	case a.Code != 0:
		return ergUebersetz
	}
	return ""
}

// armErg zählt die Prozesse eines Arms (SPEC-210 Nr. 7 b).
type armErg struct {
	gestartet, rot int
	abbruch        bool
	text           string
}

// arm startet n Prozesse, höchstens vier zugleich, alle bis zum Ende.
func (w *Werk) arm(ctx context.Context, n int, s Start) armErg {
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		erg armErg
	)
	sem := make(chan struct{}, gleichzeitig)
	for range n {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		erg.gestartet++
		wg.Add(1)
		go func() {
			defer wg.Done()
			a := w.P.Starten(ctx, s)
			<-sem
			mu.Lock()
			defer mu.Unlock()
			switch {
			case a.Abgebrochen:
				erg.abbruch, erg.text = true, a.Text
			case a.Code != 0:
				erg.rot++
				if erg.text == "" {
					erg.text = a.Text
				}
			}
		}()
	}
	wg.Wait()
	return erg
}

// stufe2 fährt das Paket einmal ohne Detektor (SPEC-210 Nr. 6 a).
func (w *Werk) stufe2(ctx context.Context, e *Erg, s Start) (string, string) {
	s.Args = w.testArgs("")
	a := w.P.Starten(ctx, s)
	e.Prozesse++
	switch {
	case a.Abgebrochen:
		return ergAbbruch, a.Text
	case a.Code != 0:
		return ergRotPaket, a.Text
	}
	return ergGruen, ""
}

// einfach fährt die Klasse andere: ein Prozess je Stufe (SPEC-210 Nr. 7 a).
func (w *Werk) einfach(ctx context.Context, e *Erg, m Mutation, s Start) (string, string) {
	s.Args = w.testArgs(m.Test)
	a := w.P.Starten(ctx, s)
	e.Prozesse++
	switch {
	case a.Abgebrochen:
		return ergAbbruch, a.Text
	case a.Code != 0:
		return ergRot, ""
	}
	return w.stufe2(ctx, e, s)
}

// rate fährt die Klasse nebenläufig: Arm mit Detektor, Arm ohne, dann Stufe 2 (SPEC-210 Nr. 7 b, c).
func (w *Werk) rate(ctx context.Context, e *Erg, m Mutation, s, sRace Start) (string, string) {
	sRace.Args, s.Args = w.testArgs(m.Test), w.testArgs(m.Test)
	mit := w.arm(ctx, m.NDetektor, sRace)
	ohne := w.arm(ctx, m.NOhne, s)
	e.Mit = fmt.Sprintf("%d/%d", mit.rot, mit.gestartet)
	e.Ohne = fmt.Sprintf("%d/%d", ohne.rot, ohne.gestartet)
	e.Prozesse += mit.gestartet + ohne.gestartet
	rot, gesamt := mit.rot+ohne.rot, mit.gestartet+ohne.gestartet
	switch {
	case mit.abbruch || ohne.abbruch || ctx.Err() != nil:
		return ergAbbruch, mit.text + ohne.text
	case rot == gesamt:
		return ergRot, ""
	case rot > 0:
		return ergFlattern, mit.text + ohne.text
	}
	return w.stufe2(ctx, e, s)
}

// fahre legt die Kopie an, wendet die Mutation an, baut und fährt die Stufen.
func (w *Werk) fahre(ctx context.Context, e *Erg, n int, m Mutation) (string, string) {
	lauf := filepath.Join(w.Ablage, "lauf")
	kopie, bin := filepath.Join(lauf, fmt.Sprint(n)), filepath.Join(lauf, fmt.Sprint(n)+".test")
	if err := w.D.Kopieren(w.Baum, m.Modul, kopie); err != nil {
		return ergAbbruch, err.Error()
	}
	if err := wendeAn(kopie, m); err != nil {
		return ergAbbruch, err.Error()
	}
	b := Bau{Dir: kopie, Paket: paketArg(m), Ziel: bin}
	if a := w.P.Bauen(ctx, b); bauErgebnis(a) != "" {
		return bauErgebnis(a), a.Text
	}
	s := Start{Dir: unter(kopie, relInModul(m, m.Paket)), Binary: bin, Frist: w.TestZeit + w.Gnade}
	if !m.Nebenlaeufig() {
		return w.einfach(ctx, e, m, s)
	}
	b.Ziel, b.Detektor = bin+".race", true
	if m.NDetektor > 0 {
		if a := w.P.Bauen(ctx, b); bauErgebnis(a) != "" {
			return bauErgebnis(a), a.Text
		}
	}
	sRace := s
	sRace.Binary, sRace.Detektor = b.Ziel, true
	return w.rate(ctx, e, m, s, sRace)
}

// mutation fährt eine Mutation und räumt ihre Kopie und Binaries ab (SPEC-210 Nr. 4 b).
func (w *Werk) mutation(ctx context.Context, n int, m Mutation) (Erg, error) {
	t0 := w.Jetzt()
	e := Erg{Name: m.Name, Test: m.Test, Paket: m.Paket, Mit: "-", Ohne: "-", Last: w.Last(), CPUs: w.CPUs}
	if m.Nebenlaeufig() {
		e.Mit, e.Ohne = "0/0", "0/0"
	}
	e.E, e.Diag = w.fahre(ctx, &e, n, m)
	e.Dauer = w.Jetzt().Sub(t0)
	kopie := filepath.Join(w.Ablage, "lauf", fmt.Sprint(n))
	for _, p := range []string{kopie, kopie + ".test", kopie + ".test.race"} {
		if err := w.D.Entfernen(p); err != nil {
			return e, fmt.Errorf("Kopie nicht entfernt: %s", p)
		}
	}
	return e, nil
}

// zaehler zählt die Ergebnisse für die Zusammenfassung.
type zaehler map[string]int

// Lauf prüft die Nullprobe, fährt die Mutationen nacheinander und gibt den Exit-Code zurück
// (SPEC-210 Nr. 1 d).
func (w *Werk) Lauf(ctx context.Context, muts []Mutation) int {
	t0 := w.Jetzt()
	defer func() { _ = w.D.Entfernen(filepath.Join(w.Ablage, "lauf")) }()
	if befunde := w.nullprobe(ctx, muts); len(befunde) > 0 && ctx.Err() == nil {
		for _, b := range befunde {
			_, _ = fmt.Fprintf(w.Fehl, "mutationslauf: %s\n", b)
		}
		return 2
	}
	z, fehler := zaehler{}, error(nil)
	fertig := 0
	for i, m := range muts {
		if ctx.Err() != nil || fehler != nil {
			break
		}
		var e Erg
		e, fehler = w.mutation(ctx, i+1, m)
		w.ausgeben(e)
		z[e.E]++
		fertig = i + 1
		w.cachePruefen(ctx)
	}
	for _, m := range muts[fertig:] {
		e := Erg{Name: m.Name, E: ergAbbruch, Test: m.Test, Paket: m.Paket, Mit: "-", Ohne: "-", Last: w.Last(), CPUs: w.CPUs}
		w.ausgeben(e)
		z[ergAbbruch]++
	}
	_, _ = fmt.Fprintf(w.Aus, "zusammenfassung mutationen=%d rot=%d rot-paket=%d gruen=%d flatternd=%d uebersetzungsfehler=%d abbruch=%d dauer=%.1f\n",
		len(muts), z[ergRot], z[ergRotPaket], z[ergGruen], z[ergFlattern], z[ergUebersetz], z[ergAbbruch], w.Jetzt().Sub(t0).Seconds())
	switch {
	case ctx.Err() != nil:
		return 3
	case fehler != nil:
		_, _ = fmt.Fprintf(w.Fehl, "mutationslauf: %v\n", fehler)
		return 1
	case z[ergUebersetz]+z[ergAbbruch] > 0:
		return 1
	}
	return 0
}

// cachePruefen leert den Build-Cache über der Größengrenze (SPEC-210 Nr. 8 d).
func (w *Werk) cachePruefen(ctx context.Context) {
	if w.P.CacheGroesse() > w.CacheGrenze {
		_ = w.P.CacheLeeren(context.WithoutCancel(ctx))
		_, _ = fmt.Fprintln(w.Fehl, "cache geleert")
	}
}
