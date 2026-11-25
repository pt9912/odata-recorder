package main

import (
	"testing"
	"time"
)

// TestProfil: beide Profile tragen die Werte von SPEC-216 Nr. 5 (11 Werte je Profil), und die Formel geht auf: Exchanges gleich 1 + Seiten +
// Lookups, Lookups mindestens 1 und höchstens Sets mal Entities, Anfragen gleich Rate mal Warm-up und Messphase (Orakel: die Literale der
// Tabelle des Vertrags, nicht der Code des Profils).
func TestProfil(t *testing.T) {
	wache(t, 30*time.Second)
	zeilen := []struct {
		name           string
		abnahme, rauch any
		wert           func(Profil) any
	}{
		{"Exchanges", 10000, 300, func(p Profil) any { return p.Exchanges }},
		{"Sets mal Entities", [2]int{5, 2000}, [2]int{5, 60}, func(p Profil) any { return [2]int{p.Sets, p.Entities} }},
		{"Entities je Seite", 50, 50, func(p Profil) any { return p.SeitenGroesse }},
		{"Seiten", 200, 10, func(p Profil) any { return p.Seiten() }},
		{"Lookups", 9799, 289, func(p Profil) any { return p.Lookups }},
		{"Rate", 20, 20, func(p Profil) any { return p.Rate }},
		{"Warm-up", 60 * time.Second, 3 * time.Second, func(p Profil) any { return p.Warmup }},
		{"Messphase", 600 * time.Second, 10 * time.Second, func(p Profil) any { return p.Messphase }},
		{"Läufe", 2, 1, func(p Profil) any { return p.Laeufe }},
		{"Frist des Aufbaus", 30 * time.Minute, 5 * time.Minute, func(p Profil) any { return p.FristAufbau }},
		{"Gesamtfrist", 120 * time.Minute, 15 * time.Minute, func(p Profil) any { return p.Gesamtfrist }},
	}
	n := 0
	for _, z := range zeilen {
		for _, f := range []struct {
			p        Profil
			erwartet any
		}{{abnahme(), z.abnahme}, {rauch(), z.rauch}} {
			n++
			if got := z.wert(f.p); got != f.erwartet {
				t.Errorf("Profil %s, %s: %v, erwartet %v", f.p.Name, z.name, got, f.erwartet)
			}
		}
	}
	if len(zeilen) != 11 || n != 22 {
		t.Errorf("%d Werte, %d Vergleiche, erwartet 11 und 22", len(zeilen), n)
	}
	for _, f := range []struct {
		p                          Profil
		name                       string
		anfragen, seitenJeSet, sum int
	}{{abnahme(), "abnahme", 13200, 40, 10000}, {rauch(), "rauch", 260, 2, 300}, {kleinesProfil(), "klein", 260, 2, 17}} {
		p := f.p
		if p.Name != f.name || p.Anfragen() != f.anfragen || p.SeitenJeSet() != f.seitenJeSet {
			t.Errorf("Profil %s: Name %q, Anfragen %d, Seiten je Set %d, erwartet %q, %d und %d",
				f.name, p.Name, p.Anfragen(), p.SeitenJeSet(), f.name, f.anfragen, f.seitenJeSet)
		}
		if p.Exchanges != f.sum || p.Exchanges != 1+p.Seiten()+p.Lookups || p.Lookups < 1 || p.Lookups > p.Sets*p.Entities {
			t.Errorf("Profil %s: Exchanges %d, Seiten %d, Lookups %d, erwartet %d = 1 + Seiten + Lookups und 1 <= Lookups <= %d",
				f.name, p.Exchanges, p.Seiten(), p.Lookups, f.sum, p.Sets*p.Entities)
		}
	}
}
