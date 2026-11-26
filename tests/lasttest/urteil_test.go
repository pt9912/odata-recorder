package main

import (
	"testing"
	"time"
)

func lauf(p95 int64) Messung { return Messung{Gueltig: true, Grund: grundOK, P95: p95} }

// TestZielvergleich: p95 ≤ Zielwert × 1000 µs besteht, gleich besteht: Ziel 100 mit 99,999, 100,000 und 100,001 ms, Ziel 1 und Ziel 9999 je an
// der Grenze (9 Fälle), über bestanden und über das Urteil einer Betriebsart.
func TestZielvergleich(t *testing.T) {
	wache(t, 30*time.Second)
	faelle := []struct {
		ziel, p95 int64
		want      bool
	}{
		{100, 99999, true}, {100, 100000, true}, {100, 100001, false},
		{1, 999, true}, {1, 1000, true}, {1, 1001, false},
		{9999, 9998999, true}, {9999, 9999000, true}, {9999, 9999001, false},
	}
	if len(faelle) != 9 {
		t.Fatalf("%d Fälle, erwartet 9", len(faelle))
	}
	for _, f := range faelle {
		want := urteilDurchgefallen
		if f.want {
			want = urteilBestanden
		}
		if got := bestanden(f.p95, f.ziel); got != f.want {
			t.Errorf("Ziel %d ms, p95 %d µs: %v, erwartet %v", f.ziel, f.p95, got, f.want)
		}
		if u := urteilBetriebsart("replay", []Messung{lauf(f.p95)}, f.ziel); u.Ergebnis != want || u.Grund != grundOK || u.ZielMs != f.ziel {
			t.Errorf("Ziel %d ms, p95 %d µs: Urteil %+v, erwartet %s", f.ziel, f.p95, u, want)
		}
	}
}

// TestAbnahme: Profil (2) × Ziel (99, 100, 101) × Betriebsart (3) × Urteile (die 3 des Plans: bestanden und bestanden, bestanden und
// durchgefallen, ungueltig und ungueltig; dazu durchgefallen und bestanden, ungueltig und bestanden, damit jede der zwei Stellen allein
// scheitert) gleich 90; genau ein Element ergibt ja, grund je Element aus dem Raster (Profil, Ziel, Betriebsart) und den Urteilen als Literal;
// dazu die Urteile in anderer Zahl (keines, eines).
func TestAbnahme(t *testing.T) {
	wache(t, 30*time.Second)
	grund := map[string][3][3]string{
		"abnahme": {{"ziel", "ziel", "ziel"}, {"betriebsart", "betriebsart", ""}, {"ziel", "ziel", "ziel"}},
		"rauch":   {{"profil", "profil", "profil"}, {"profil", "profil", "profil"}, {"profil", "profil", "profil"}},
	}
	ergebnisse := [][]string{
		{urteilBestanden, urteilBestanden}, {urteilBestanden, urteilDurchgefallen}, {urteilUngueltig, urteilUngueltig},
		{urteilDurchgefallen, urteilBestanden}, {urteilUngueltig, urteilBestanden},
	}
	grundErgebnisse := []string{grundOK, grundErgebnis, grundErgebnis, grundErgebnis, grundErgebnis}
	n, ja := 0, 0
	for profil, raster := range grund {
		for z, ziel := range []int64{99, 100, 101} {
			for b, betriebsart := range []string{"replay", "mock", "beide"} {
				for e, urteile := range ergebnisse {
					erwartet := raster[z][b]
					if erwartet == "" {
						erwartet = grundErgebnisse[e]
					}
					gilt, got := abnahmeGilt(profil, ziel, betriebsart, urteile...)
					if got != erwartet || gilt != (erwartet == grundOK) {
						t.Errorf("%s, Ziel %d, %s, %v: %v %s; erwartet grund %s", profil, ziel, betriebsart, urteile, gilt, got, erwartet)
					}
					n++
					if gilt {
						ja++
					}
				}
			}
		}
	}
	if n != 2*3*3*5 || ja != 1 {
		t.Errorf("%d Elemente und %d mal ja, erwartet 90 und 1", n, ja)
	}
	for _, urteile := range [][]string{nil, {urteilBestanden}} {
		if gilt, got := abnahmeGilt("abnahme", 100, "beide", urteile...); gilt || got != grundErgebnis {
			t.Errorf("Urteile %v: %v %s; erwartet nein und ergebnis", urteile, gilt, got)
		}
	}
}

// TestStreuung: zwei gültige Läufe: Toleranz das Größere aus ⌊kleineres p95 ÷ 10⌋ und 2000 µs; d über der Toleranz ist streuung, sonst gilt
// das größere p95 (in beiden Folgen der Läufe); Grenzen bei 50000 (Toleranz 5000), 10000 (Boden), 19999 (⌊1999,9⌋ unter dem Boden), 20010 (2001)
// 20015 (⌊2001,5⌋), 20009, 20019, 20020 und 123456 (⌊12345,6⌋); Maximum gegen das Ziel; ein ungültiger Lauf (2 Formen) macht das Urteil ungültig mit seinem Grund; ein Lauf allein
// hat keine Streuung.
func TestStreuung(t *testing.T) {
	wache(t, 30*time.Second)
	faelle := []struct {
		klein, gross int64
		tol          int64
		gueltig      bool
	}{
		{50000, 55000, 5000, true}, {50000, 55001, 5000, false}, {10000, 12000, 2000, true}, {10000, 12001, 2000, false},
		{10000, 11500, 2000, true}, {19999, 21999, 2000, true}, {19999, 22000, 2000, false}, {20010, 22011, 2001, true},
		{20010, 22012, 2001, false}, {20015, 22016, 2001, true}, {20015, 22017, 2001, false},
		{20009, 22009, 2000, true}, {20009, 22010, 2000, false}, {20019, 22020, 2001, true}, {20019, 22021, 2001, false},
		{20020, 22022, 2002, true}, {20020, 22023, 2002, false}, {123456, 135801, 12345, true}, {123456, 135802, 12345, false},
	}
	n := 0
	for _, f := range faelle {
		for _, lauefe := range [][]Messung{{lauf(f.klein), lauf(f.gross)}, {lauf(f.gross), lauf(f.klein)}} {
			n++
			u := urteilBetriebsart("mock", lauefe, 1000)
			if u.Streuung != f.gross-f.klein || u.Toleranz != f.tol || u.P95 != f.gross {
				t.Errorf("%d und %d: d %d, Toleranz %d, p95 %d; erwartet %d, %d und %d", f.klein, f.gross, u.Streuung, u.Toleranz, u.P95, f.gross-f.klein, f.tol, f.gross)
			}
			if got := u.Ergebnis == urteilBestanden && u.Grund == grundOK || u.Ergebnis == urteilUngueltig && u.Grund == grundStreuung; !got || (u.Ergebnis == urteilBestanden) != f.gueltig {
				t.Errorf("%d und %d: Urteil %s, grund %s; erwartet gültig %v", f.klein, f.gross, u.Ergebnis, u.Grund, f.gueltig)
			}
		}
	}
	for _, f := range []struct {
		a, b, d, tol int64
		want         string
	}{{98000, 101000, 3000, 9800, urteilDurchgefallen}, {90000, 95000, 5000, 9000, urteilBestanden}, {100000, 100000, 0, 10000, urteilBestanden}} {
		u := urteilBetriebsart("replay", []Messung{lauf(f.a), lauf(f.b)}, 100)
		if u.Ergebnis != f.want || u.Grund != grundOK || u.P95 != f.b || u.Streuung != f.d || u.Toleranz != f.tol {
			t.Errorf("%d und %d bei Ziel 100: %+v; erwartet %s mit dem größeren p95, d %d und Toleranz %d", f.a, f.b, u, f.want, f.d, f.tol)
		}
		n++
	}
	ungueltig := Messung{Gueltig: false, Grund: grundLast, P95: 70000}
	for _, laeufe := range [][]Messung{{ungueltig, lauf(50000)}, {lauf(50000), ungueltig}} {
		u := urteilBetriebsart("replay", laeufe, 100)
		if u != (urteil{Betriebsart: "replay", Laeufe: 2, ZielMs: 100, P95: 50000, Ergebnis: urteilUngueltig, Grund: grundLast}) {
			t.Errorf("ein ungültiger Lauf: %+v; erwartet ungueltig, last, das p95 des gültigen Laufs, Streuung und Toleranz 0", u)
		}
		n++
	}
	zwei := []Messung{{Gueltig: false, Grund: grundLast, P95: 40000}, {Gueltig: false, Grund: grundBinary, P95: 90000}}
	for _, laeufe := range [][]Messung{zwei, {zwei[1], zwei[0]}} {
		u := urteilBetriebsart("replay", laeufe, 100)
		if u != (urteil{Betriebsart: "replay", Laeufe: 2, ZielMs: 100, Ergebnis: urteilUngueltig, Grund: laeufe[0].Grund}) {
			t.Errorf("zwei ungültige Läufe %s und %s: %+v; erwartet den Grund des ersten, p95, Streuung und Toleranz 0", laeufe[0].Grund, laeufe[1].Grund, u)
		}
		n++
	}
	for name, laeufe := range map[string][]Messung{"leere Liste": {}, "nil": nil} {
		if u := urteilBetriebsart("mock", laeufe, 80); u != (urteil{Betriebsart: "mock", ZielMs: 80, Ergebnis: urteilUngueltig, Grund: grundZaehlung}) {
			t.Errorf("%s ohne Lauf: %+v; erwartet ungueltig mit zaehlung, 0 Läufe und Dauern 0", name, u)
		}
		n++
	}
	u := urteilBetriebsart("replay", []Messung{lauf(120000)}, 100)
	if u != (urteil{Betriebsart: "replay", Laeufe: 1, ZielMs: 100, P95: 120000, Ergebnis: urteilDurchgefallen, Grund: grundOK}) {
		t.Errorf("ein Lauf allein: %+v; erwartet durchgefallen ohne Streuung", u)
	}
	n++
	if n != 48 {
		t.Errorf("%d Fälle, erwartet 48", n)
	}
}

// TestExit: Urteil von replay und mock je aus {bestanden, durchgefallen, ungueltig, nicht gefahren} ohne das Paar „beide nicht gefahren“ (15,
// Exit als Literal) und die drei Ausgänge Aufbau, Frist und Abbruch mit jedem der 15 Paare (45): 0 alle bestanden, 1 durchgefallen ohne
// ungültig, 3 ungültig, Aufbau oder Frist, 2 Abbruch; Vorrang 2 vor 3 vor 1 vor 0.
func TestExit(t *testing.T) {
	wache(t, 30*time.Second)
	b, d, u, k := urteilBestanden, urteilDurchgefallen, urteilUngueltig, urteilKein
	paare := []struct {
		replay, mock string
		exit         int
	}{
		{b, b, 0}, {b, d, 1}, {b, u, 3}, {b, k, 0}, {d, b, 1}, {d, d, 1}, {d, u, 3}, {d, k, 1},
		{u, b, 3}, {u, d, 3}, {u, u, 3}, {u, k, 3}, {k, b, 0}, {k, d, 1}, {k, u, 3},
	}
	if len(paare) != 15 {
		t.Fatalf("%d Paare, erwartet 15", len(paare))
	}
	n := 0
	for _, p := range paare {
		if got := exitCode(ausgang{Replay: p.replay, Mock: p.mock}); got != p.exit {
			t.Errorf("replay %q, mock %q: Exit %d, erwartet %d", p.replay, p.mock, got, p.exit)
		}
		n++
		for name, a := range map[string]ausgang{
			"Aufbau": {Replay: p.replay, Mock: p.mock, Aufbau: true}, "Frist": {Replay: p.replay, Mock: p.mock, Frist: true},
			"Abbruch": {Replay: p.replay, Mock: p.mock, Abbruch: true}, "Abbruch und Aufbau": {Replay: p.replay, Mock: p.mock, Abbruch: true, Aufbau: true},
		} {
			want := 3
			if a.Abbruch {
				want = 2
			}
			if got := exitCode(a); got != want {
				t.Errorf("%s, replay %q, mock %q: Exit %d, erwartet %d", name, p.replay, p.mock, got, want)
			}
			n++
		}
	}
	if n != 15*5 {
		t.Errorf("%d Fälle, erwartet 75", n)
	}
}
