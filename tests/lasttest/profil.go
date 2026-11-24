package main

import "time"

// Profil ist ein fester Satz von Werten des Lasttests (SPEC-216 Nr. 5); es gibt genau zwei, abnahme und rauch. Seiten folgt aus
// Sets, Entities und SeitenGroesse; Exchanges ist 1 + Seiten + Lookups.
type Profil struct {
	Name          string
	Exchanges     int
	Sets          int
	Entities      int
	SeitenGroesse int
	Lookups       int
	Rate          int
	Warmup        time.Duration
	Messphase     time.Duration
	Laeufe        int
	FristAufbau   time.Duration
	Gesamtfrist   time.Duration
}

// abnahme trägt die Werte von LH-QA-01.a.
func abnahme() Profil {
	return Profil{
		Name: "abnahme", Exchanges: 10000, Sets: 5, Entities: 2000, SeitenGroesse: 50, Lookups: 9799, Rate: 20,
		Warmup: 60 * time.Second, Messphase: 600 * time.Second, Laeufe: 2, FristAufbau: 30 * time.Minute, Gesamtfrist: 120 * time.Minute,
	}
}

// rauch prüft das Zusammenspiel in wenigen Minuten und ist nie eine Abnahme.
func rauch() Profil {
	return Profil{
		Name: "rauch", Exchanges: 300, Sets: 5, Entities: 60, SeitenGroesse: 50, Lookups: 289, Rate: 20,
		Warmup: 3 * time.Second, Messphase: 10 * time.Second, Laeufe: 1, FristAufbau: 5 * time.Minute, Gesamtfrist: 15 * time.Minute,
	}
}

// SeitenJeSet ist die Zahl der Seiten eines Sets.
func (p Profil) SeitenJeSet() int { return (p.Entities + p.SeitenGroesse - 1) / p.SeitenGroesse }

// Seiten ist die Zahl der Seiten aller Sets.
func (p Profil) Seiten() int { return p.Sets * p.SeitenJeSet() }

// Anfragen ist die Zahl der Anfragen eines Laufs: Warm-up und Messphase.
func (p Profil) Anfragen() int { return p.Rate * int((p.Warmup+p.Messphase)/time.Second) }
