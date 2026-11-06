package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// paketSchluessel identifiziert (Modul, Paket) einer Mutation.
type paketSchluessel struct{ modul, paket string }

// nullprobe übersetzt und fährt den unveränderten Stand je (Modul, Paket) der Liste (SPEC-210 Nr. 5 a)
// und liefert die Befunde.
func (w *Werk) nullprobe(ctx context.Context, muts []Mutation) []string {
	var (
		befunde []string
		reihe   []paketSchluessel
		je      = map[paketSchluessel][]int{}
	)
	for i, m := range muts {
		k := paketSchluessel{m.Modul.Dir, m.Paket}
		if _, da := je[k]; !da {
			reihe = append(reihe, k)
		}
		je[k] = append(je[k], i)
	}
	for j, k := range reihe {
		befunde = append(befunde, w.nullprobePaket(ctx, j, muts, je[k])...)
		if ctx.Err() != nil {
			break
		}
	}
	return befunde
}

// nullprobePaket fährt die Nullprobe eines Pakets, ohne Detektor und, wenn eine Mutation es braucht, mit.
func (w *Werk) nullprobePaket(ctx context.Context, j int, muts []Mutation, idx []int) []string {
	m0 := muts[idx[0]]
	kopie := filepath.Join(w.Ablage, "lauf", fmt.Sprintf("null%d", j))
	defer func() {
		for _, p := range []string{kopie, kopie + ".test", kopie + ".test.race"} {
			_ = w.D.Entfernen(p)
		}
	}()
	if err := w.D.Kopieren(w.Baum, m0.Modul, kopie); err != nil {
		return []string{fmt.Sprintf("Ausgangsstand rot: %s (Kopie: %v)", m0.Paket, err)}
	}
	befunde := w.nullprobeArt(ctx, kopie, muts, idx, false)
	for _, i := range idx {
		if muts[i].Nebenlaeufig() {
			return append(befunde, w.nullprobeArt(ctx, kopie, muts, idx, true)...)
		}
	}
	return befunde
}

// treffer zählt die Zeilen, die `-test.list` liefert; -1 heißt: der Aufruf scheiterte oder die Zeile ist fremd.
func treffer(l Ausgang, test string) int {
	if l.Code != 0 {
		return -1
	}
	k := 0
	for _, z := range strings.Split(l.Text, "\n") {
		if strings.TrimSpace(z) != "" {
			k++
		}
	}
	if k == 1 && strings.TrimSpace(l.Text) != test {
		return -1
	}
	return k
}

// nullprobeArt übersetzt den unveränderten Stand und prüft die Tests der Liste und das Paket.
func (w *Werk) nullprobeArt(ctx context.Context, kopie string, muts []Mutation, idx []int, detektor bool) []string {
	m0 := muts[idx[0]]
	bin := kopie + ".test"
	if detektor {
		bin += ".race"
	}
	rot := fmt.Sprintf("Ausgangsstand rot: %s", m0.Paket)
	a := w.P.Bauen(ctx, Bau{Dir: kopie, Paket: paketArg(m0), Ziel: bin, Detektor: detektor})
	if a.Code != 0 || a.Abgebrochen {
		return []string{rot}
	}
	var befunde []string
	dir := unter(kopie, relInModul(m0, m0.Paket))
	start := func(args ...string) Ausgang {
		return w.P.Starten(ctx, Start{Dir: dir, Binary: bin, Args: args, Detektor: detektor, Nullprobe: true, Frist: w.TestZeit + w.Gnade})
	}
	for _, i := range idx {
		m := muts[i]
		if detektor && !m.Nebenlaeufig() {
			continue
		}
		k := 0
		if !a.KeinBinary {
			k = treffer(start("-test.list", "^"+m.Test+"$"), m.Test)
		}
		if k != 1 {
			befunde = append(befunde, fmt.Sprintf("Eingabe: Mutation %d (%s): Test %s trifft %d Tests in %s", i+1, m.Name, m.Test, max(k, 0), m.Paket))
		}
	}
	if !a.KeinBinary {
		if l := start(w.testArgs("")...); l.Code != 0 || l.Abgebrochen {
			befunde = append(befunde, rot)
		}
	}
	return befunde
}
