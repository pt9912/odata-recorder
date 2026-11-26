package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

// laufFolge sind die Aufrufe eines Laufs: Port, Start, Bereitschaft, Probe vor dem Warm-up, Probe nach dem Nachlauf, SIGTERM.
func laufFolge(modus string) []string {
	return []string{"port", "start MODE=" + modus, "verbinde", "last", "last", "signal terminated"}
}

// TestAblauf (SPEC-216 Nr. 4 b): Probe, Aufbau und je Betriebsart, replay vor mock, ihre Läufe, für beide, replay, mock und das Profil rauch.
func TestAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		name, betriebsart string
		laeufe            int
		folge             []string
	}{
		{"abnahme", "beide", 2, []string{"replay", "replay", "mock", "mock"}},
		{"abnahme", "replay", 2, []string{"replay", "replay"}},
		{"abnahme", "mock", 2, []string{"mock", "mock"}},
		{"rauch", "beide", 1, []string{"replay", "mock"}},
	} {
		s := neuerProbestand(t, kurzesProfil(f.name, f.laeufe), f.betriebsart)
		if code := warteAuf(t, s.fahre(context.Background()), 30*time.Second); code != 0 {
			t.Fatalf("%s %s: Exit %d, %q", f.name, f.betriebsart, code, s.fehler.String())
		}
		want := []string{"last", "aufbau"}
		var laeufe []string
		for i, modus := range f.folge {
			want = append(want, laufFolge(modus)...)
			laeufe = append(laeufe, modus+" "+[]string{"1", "2"}[i%f.laeufe])
		}
		if got := s.e.liste(); !slices.Equal(got, want) {
			t.Errorf("%s %s: Aufrufe %v, erwartet %v", f.name, f.betriebsart, got, want)
		}
		var got []string
		for _, z := range s.zeilenVon("Messung") {
			got = append(got, feld(z, "betriebsart")+" "+feld(z, "lauf"))
		}
		if !slices.Equal(got, laeufe) {
			t.Errorf("%s %s: Läufe %v, erwartet %v", f.name, f.betriebsart, got, laeufe)
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// TestAbbruch (SPEC-216 Nr. 4 d): der Abbruch im Aufbau, im Warm-up, in der Messphase und im Nachlauf sendet dem laufenden Binary genau ein
// SIGTERM (im Aufbau läuft keines), schreibt keine Zeile des Laufs, kein Urteil und keine Abnahme, meldet `abgebrochen` und endet mit 2.
func TestAbbruch(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		name    string
		bei     int
		starts  int
		zeilen  []string
		signale []os.Signal
	}{
		{"Aufbau", -1, 0, []string{"Start"}, nil},
		{"Warm-up", 5, 1, []string{"Start", "Aufbau"}, []os.Signal{syscall.SIGTERM}},
		{"Messphase", 25, 1, []string{"Start", "Aufbau"}, []os.Signal{syscall.SIGTERM}},
		{"Nachlauf", 39, 1, []string{"Start", "Aufbau"}, []os.Signal{syscall.SIGTERM}},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		s := neuerProbestand(t, kurzesProfil("klein", 2), "beide")
		switch f.name {
		case "Aufbau":
			s.ein.Aufbau = func(c context.Context, _ aufbauEingabe) (aufbauErgebnis, error) {
				cancel()
				<-c.Done()
				return aufbauErgebnis{}, &aufbauFehler{grund: grundAufbau, stelle: stelleAufnahme, ursache: c.Err()}
			}
		case "Nachlauf":
			s.dienst.stumm = func(_, k int) bool { return k == f.bei }
		default:
			s.dienst.halte = func(_, k int, _ *http.Request) {
				if k == f.bei {
					cancel()
				}
			}
		}
		exit := s.fahre(ctx)
		if f.name == "Nachlauf" {
			bisWahr(func() bool { b := s.dienst.aktuell(); return b != nil && b.fertig.Load() == 40 })
			pause(20 * time.Millisecond)
			cancel()
		}
		code := warteAuf(t, exit, 30*time.Second)
		cancel()
		if code != 2 || s.fehler.String() != "lasttest: abgebrochen\n" || !slices.Equal(s.arten(), f.zeilen) || s.dienst.zahl() != f.starts {
			t.Errorf("%s: Exit %d, Meldungen %q, Zeilen %v, %d Starts", f.name, code, s.fehler.String(), s.arten(), s.dienst.zahl())
		}
		if s.dienst.zahl() == 1 && !slices.Equal(s.dienst.lauf(0).gesendet(), f.signale) {
			t.Errorf("%s: Signale %v, erwartet %v", f.name, s.dienst.lauf(0).gesendet(), f.signale)
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Phasen, erwartet 4", n)
	}
}

// TestGesamtfrist (SPEC-216 Nr. 4 c): je Gesamtfrist der Profile beendet die Uhr bei Beginn + Frist das laufende Binary, meldet es und endet
// mit 3; bei Beginn + Frist − 1 ns endet der Lauf mit seinen Urteilen. Der Beginn ist der Aufruf, nicht das Ende des Aufbaus.
func TestGesamtfrist(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, frist := range []time.Duration{abnahme().Gesamtfrist, rauch().Gesamtfrist} {
		for _, versatz := range []time.Duration{0, -time.Nanosecond} {
			p := kurzesProfil("klein", 1)
			p.Gesamtfrist = frist
			s := neuerProbestand(t, p, "replay")
			start := s.uhr.Jetzt()
			aufbau := s.ein.Aufbau
			s.ein.Aufbau = func(c context.Context, a aufbauEingabe) (aufbauErgebnis, error) {
				s.uhr.Vorstellen(time.Minute)
				return aufbau(c, a)
			}
			var lauf <-chan struct{}
			starte := s.ein.Starte
			s.ein.Starte = func(c context.Context, env []string) (*prozess, error) {
				lauf = c.Done()
				return starte(c, env)
			}
			s.dienst.halte = func(_, k int, _ *http.Request) {
				if k != 39 {
					return
				}
				s.uhr.Vorstellen(start.Add(frist + versatz).Sub(s.uhr.Jetzt()))
				if versatz == 0 {
					bisWahr(func() bool { return beendet(lauf) })
				}
			}
			code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
			want, zeilen, meldung := 0, []string{"Start", "Aufbau", "Messung", "Last", "Urteil", "Abnahme"}, ""
			if versatz == 0 {
				want, zeilen, meldung = 3, []string{"Start", "Aufbau"}, "lasttest: Gesamtfrist abgelaufen\n"
			}
			if code != want || !slices.Equal(s.arten(), zeilen) || s.fehler.String() != meldung ||
				!slices.Equal(s.dienst.lauf(0).gesendet(), []os.Signal{syscall.SIGTERM}) {
				t.Errorf("Frist %v, Versatz %v: Exit %d, Zeilen %v, Meldungen %q, Signale %v", frist, versatz, code, s.arten(), s.fehler.String(),
					s.dienst.lauf(0).gesendet())
			}
			n++
		}
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// TestProbenImAblauf (SPEC-216 Nr. 11): eine Probe über der Schwelle vor dem Aufbau verhindert ihn; eine allein nach dem Nachlauf macht den
// Lauf mit grund=last ungültig.
func TestProbenImAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
	s.last.inhalt = func(int) string { return "10.01 1 1" }
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	if code != 3 || len(s.aufbaue) != 0 || s.fehler.String() != "lasttest: Aufbau gescheitert: last\n" || !slices.Equal(s.arten(), []string{"Start"}) {
		t.Errorf("vor dem Aufbau: Exit %d, %d Aufrufe des Aufbaus, Meldungen %q, Zeilen %v", code, len(s.aufbaue), s.fehler.String(), s.arten())
	}
	s = neuerProbestand(t, kurzesProfil("klein", 1), "replay")
	// Über der Schwelle ist allein eine Probe, die nach der letzten beendeten Anfrage genommen wird.
	s.last.inhalt = func(int) string {
		if b := s.dienst.aktuell(); b != nil && b.fertig.Load() == 40 {
			return "12.00 1 1"
		}
		return "1.00 1 1"
	}
	code = warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	m, l := s.zeilenVon("Messung"), s.zeilenVon("Last")
	if code != 3 || feld(m[0], "grund") != "last" || feld(l[0], "proben") != "2" || feld(l[0], "max") != "12.00" {
		t.Errorf("nach dem Nachlauf: Exit %d, %q, %q", code, m, l)
	}
}

// TestProbenDauer (SPEC-216 Nr. 11): die Proben eines Laufs reichen über Warm-up und Messphase: bei 40 s und 60 s sind es fünf (Beginn,
// +30, +60, +90 s, nach dem Nachlauf); die Uhr geht über einen Tick erst hinaus, wenn die Quelle gelesen ist.
func TestProbenDauer(t *testing.T) {
	wache(t, 60*time.Second)
	p := kurzesProfil("klein", 1)
	p.Warmup, p.Messphase = 40*time.Second, 60*time.Second
	s := neuerProbestand(t, p, "replay")
	vorher, aufgegeben := s.uhr.vorher, false
	s.uhr.vorher = func(d time.Duration, ziel time.Time) {
		if b := s.dienst.aktuell(); b != nil && !aufgegeben {
			want := int64(2 + s.uhr.Jetzt().Sub(b.start)/(30*time.Second))
			aufgegeben = !bisWahr(func() bool { return s.last.n.Load() >= want })
		}
		vorher(d, ziel)
	}
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	if l := s.zeilenVon("Last"); code != 0 || len(l) != 1 || feld(l[0], "proben") != "5" {
		t.Errorf("Exit %d, %q; erwartet proben=5", code, l)
	}
}

// TestSpaeteSpans (SPEC-216 Nr. 10 e): kommt der letzte Span erst nach der Frist des Nachlaufs und vor dem Ende des Binary, trägt der Lauf
// bei voller Zahl grund=zaehlung.
func TestSpaeteSpans(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
	s.dienst.stumm = func(_, k int) bool { return k == 39 }
	s.dienst.signal = func(b *ersatzBinaer, _ os.Signal) {
		b.zeile(kindZeile("replay", "spaet", 100))
		b.zeile(gesamtZeile("replay", "spaet", 1000))
	}
	exit := s.fahre(context.Background())
	treibeBis(t, s.uhr, 10*time.Second, s.dienst.alleBeendet(40), s.dienst.signaleBei(1))
	code := warteAuf(t, exit, 30*time.Second)
	if m := s.zeilenVon("Messung"); code != 3 || len(m) != 1 || feld(m[0], "spans") != "40" || feld(m[0], "grund") != "zaehlung" ||
		feld(m[0], "fehler") != "0" {
		t.Errorf("Exit %d, %q", code, m)
	}
}

// TestAufbauMeldung (SPEC-216 Nr. 6 e, f): die Meldung nennt das Token aus dem *aufbauFehler, auch wenn ein Fehler von Herunterfahren dazu
// kam; der Exit ist 3, und es folgt keine weitere Zeile.
func TestAufbauMeldung(t *testing.T) {
	wache(t, 60*time.Second)
	schritt := &aufbauFehler{grund: grundAufbau, stelle: stelleStatus, ursache: errors.New("Status 500")}
	n := 0
	for _, f := range []struct {
		err     error
		meldung string
	}{
		{errors.Join(schritt, errors.New("Herunterfahren: kaputt")), "Aufbau gescheitert: aufbau"},
		{errors.Join(&aufbauFehler{grund: grundFrist, stelle: stelleFrist, ursache: errFrist}, errors.New("x")), "Aufbau gescheitert: frist"},
		{&aufbauFehler{grund: grundAufbau, stelle: stelleDaten, ursache: errors.New("y")}, "Aufbau: DATA_DIR nicht leer"},
		{errors.New("ohne Token"), "Aufbau gescheitert: aufbau"},
	} {
		s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
		s.ein.Aufbau = func(context.Context, aufbauEingabe) (aufbauErgebnis, error) { return aufbauErgebnis{}, f.err }
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		if code != 3 || s.fehler.String() != "lasttest: "+f.meldung+"\n" || !slices.Equal(s.arten(), []string{"Start"}) || s.dienst.zahl() != 0 {
			t.Errorf("%v: Exit %d, Meldungen %q, Zeilen %v", f.err, code, s.fehler.String(), s.arten())
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// umgebungMit liefert den lookup über eine Karte der fünf Variablen, in der name den Wert wert hat; fehlt nimmt name heraus.
func umgebungMit(name, wert string, fehlt bool) func(string) (string, bool) {
	m := map[string]string{
		varBetriebsart: "beide", varProfil: "abnahme", varZielMs: "100", varStartwert: "1", varKerne: "20",
	}
	if name != "" {
		m[name] = wert
	}
	if fehlt {
		delete(m, name)
	}
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

// TestLiesUmgebung: jede der fünf Variablen mit ihren Werten; fehlend, leer und ungültig sind je ein Fehler, der sie nennt.
func TestLiesUmgebung(t *testing.T) {
	wache(t, 30*time.Second)
	w, err := liesUmgebung(umgebungMit("", "", false))
	if err != nil || w != (werte{Profil: abnahme(), Betriebsart: "beide", ZielMs: 100, Startwert: 1, Kerne: 20}) {
		t.Fatalf("%+v, %v", w, err)
	}
	for _, f := range []struct {
		name, wert string
		pruefe     func(werte) bool
	}{
		{varBetriebsart, "replay", func(w werte) bool { return w.Betriebsart == "replay" }},
		{varBetriebsart, "mock", func(w werte) bool { return w.Betriebsart == "mock" }},
		{varProfil, "rauch", func(w werte) bool { return w.Profil == rauch() }},
		{varZielMs, "9999", func(w werte) bool { return w.ZielMs == 9999 }},
		{varStartwert, "999999999999999999", func(w werte) bool { return w.Startwert == 999999999999999999 }},
		{varStartwert, "0", func(w werte) bool { return w.Startwert == 0 }},
		{varKerne, "1", func(w werte) bool { return w.Kerne == 1 }},
	} {
		if w, err := liesUmgebung(umgebungMit(f.name, f.wert, false)); err != nil || !f.pruefe(w) {
			t.Errorf("%s=%s: %+v, %v", f.name, f.wert, w, err)
		}
	}
	n := 0
	for name, falsch := range map[string][]string{
		varBetriebsart: {"alle", ""}, varProfil: {"voll", ""}, varZielMs: {"0", "x", ""}, varStartwert: {"-1", ""}, varKerne: {"0", "x", ""},
	} {
		for _, wert := range falsch {
			if _, err := liesUmgebung(umgebungMit(name, wert, false)); err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("%s=%q: %v", name, wert, err)
			}
			n++
		}
		if _, err := liesUmgebung(umgebungMit(name, "", true)); err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s fehlt: %v", name, err)
		}
		n++
	}
	if n != 17 {
		t.Errorf("%d Fälle, erwartet 17", n)
	}
}

// TestStartwertBisZumAufbau (SPEC-216 Nr. 6 c): der Startwert aus der Umgebung geht unverändert an den Aufbau und in die Liste der Anfragen.
func TestStartwertBisZumAufbau(t *testing.T) {
	wache(t, 60*time.Second)
	w, err := liesUmgebung(umgebungMit(varStartwert, "999999999999999999", false))
	if err != nil {
		t.Fatal(err)
	}
	p := kurzesProfil("klein", 1)
	s := neuerProbestand(t, p, "replay")
	w.Profil, w.Betriebsart = p, "replay"
	s.ein.werte = w
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	a := s.aufbaue
	if code != 0 || len(a) != 1 || a[0].Startwert != 999999999999999999 || a[0].Profil != p || a[0].DataDir != s.ein.DataDir || a[0].Uhr != Uhr(s.uhr) {
		t.Fatalf("Exit %d, Aufbau %+v", code, a)
	}
	if want := erwarteteZeilen(listeReplay(p, 999999999999999999, 1, "w-1", "s-1")); !slices.Equal(s.dienst.liste(), want) {
		t.Errorf("Anfragen weichen von der Liste des Startwerts ab: %q", s.dienst.liste()[:3])
	}
	if z := s.zeilenVon("Start"); feld(z[0], "startwert") != "999999999999999999" {
		t.Errorf("%q", z)
	}
}

// TestDatenMiB: die Summe der Dateien unter dem Verzeichnis, auch in Unterverzeichnissen, in ganzen MiB abgeschnitten.
func TestDatenMiB(t *testing.T) {
	wache(t, 30*time.Second)
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "u"), 0o700); err != nil {
		t.Fatal(err)
	}
	schreibe := func(name string, n int) {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	schreibe("a", 3<<19)
	schreibe("u/b", 3<<19-1)
	if got := datenMiB(dir); got != 2 {
		t.Errorf("%d MiB, erwartet 2", got)
	}
	schreibe("c", 1)
	if got := datenMiB(dir); got != 3 {
		t.Errorf("%d MiB, erwartet 3", got)
	}
	if got := datenMiB(filepath.Join(dir, "fehlt")); got != 0 {
		t.Errorf("fehlendes Verzeichnis: %d", got)
	}
}

// TestProzessCPU: die CPU-Zeit des Prozesses wächst unter Rechenarbeit und fällt nicht.
func TestProzessCPU(t *testing.T) {
	wache(t, 30*time.Second)
	a := prozessCPU()
	var x uint64
	ok := bisWahr(func() bool {
		for i := range 100000 {
			x = x*31 + uint64(i)
		}
		return prozessCPU() > a+5*time.Millisecond
	})
	if !ok {
		t.Errorf("CPU-Zeit wuchs nicht: %v, %d", prozessCPU()-a, x%2)
	}
}

// TestAblaufAusgabeFolge (SPEC-216 Nr. 13 d): die Folge der Zeilenarten für beide, replay, mock und rauch.
func TestAblaufAusgabeFolge(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		betriebsart string
		laeufe      int
		folge       string
	}{
		{"beide", 2, "Start Aufbau Messung Last Messung Last Messung Last Messung Last Urteil Urteil Abnahme"},
		{"replay", 2, "Start Aufbau Messung Last Messung Last Urteil Abnahme"},
		{"mock", 2, "Start Aufbau Messung Last Messung Last Urteil Abnahme"},
		{"beide", 1, "Start Aufbau Messung Last Messung Last Urteil Urteil Abnahme"},
	} {
		s := neuerProbestand(t, kurzesProfil("klein", f.laeufe), f.betriebsart)
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		if got := strings.Join(s.arten(), " "); code != 0 || got != f.folge {
			t.Errorf("%s mit %d Läufen: Exit %d, %s; erwartet %s", f.betriebsart, f.laeufe, code, got, f.folge)
		}
		var urteile []string
		for _, z := range s.zeilenVon("Urteil") {
			urteile = append(urteile, feld(z, "betriebsart"))
		}
		want := []string{"replay", "mock"}
		if f.betriebsart != "beide" {
			want = []string{f.betriebsart}
		}
		abn := s.zeilenVon("Abnahme")
		if !slices.Equal(urteile, want) || feld(abn[0], "grund") != "profil" || feld(abn[0], "ziel_ms") != "100" {
			t.Errorf("%s: Urteile %v, Abnahme %q", f.betriebsart, urteile, abn)
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// TestAbnahmeImAblauf (SPEC-216 Nr. 12 b): das Profil abnahme mit Ziel 100 und beiden Betriebsarten, beide bestanden, trägt abnahme=ja;
// mit nur einer Betriebsart grund=betriebsart; ein p95 über dem Ziel endet mit 1.
func TestAbnahmeImAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	for _, f := range []struct {
		betriebsart string
		ziel        int64
		exit        int
		abnahme     string
		grund       string
	}{
		{"beide", 100, 0, "ja", "ok"}, {"replay", 100, 0, "nein", "betriebsart"}, {"beide", 200, 0, "nein", "ziel"},
	} {
		s := neuerProbestand(t, kurzesProfil("abnahme", 2), f.betriebsart)
		s.ein.ZielMs = f.ziel
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		z := s.zeilenVon("Abnahme")
		if code != f.exit || feld(z[0], "abnahme") != f.abnahme || feld(z[0], "grund") != f.grund {
			t.Errorf("%s Ziel %d: Exit %d, %q", f.betriebsart, f.ziel, code, z)
		}
	}
	s := neuerProbestand(t, kurzesProfil("abnahme", 2), "beide")
	s.dienst.dauer = func(int, int) (int, int) { return 100001, 100 }
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	u := s.zeilenVon("Urteil")
	if code != 1 || feld(u[0], "ergebnis") != "durchgefallen" || feld(u[0], "p95_ms") != "100.001" || feld(s.zeilenVon("Abnahme")[0], "grund") != "ergebnis" {
		t.Errorf("über dem Ziel: Exit %d, %q", code, u)
	}
}

// TestWegeOhneUrteil (SPEC-216 Nr. 13 d): die vier Wege ohne Urteil; bei der Gesamtfrist und beim Abbruch steht der beendete Lauf davor.
func TestWegeOhneUrteil(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		name  string
		folge string
		exit  int
	}{
		{"Probe vor dem Aufbau", "Start", 3}, {"Aufbau gescheitert", "Start", 3},
		{"Gesamtfrist", "Start Aufbau Messung Last", 3}, {"Abbruch", "Start Aufbau Messung Last", 2},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		s := neuerProbestand(t, kurzesProfil("klein", 1), "beide")
		start := s.uhr.Jetzt()
		var lauf <-chan struct{}
		starte := s.ein.Starte
		s.ein.Starte = func(c context.Context, env []string) (*prozess, error) {
			lauf = c.Done()
			return starte(c, env)
		}
		switch f.name {
		case "Probe vor dem Aufbau":
			s.last.inhalt = func(int) string { return "99 1 1" }
		case "Aufbau gescheitert":
			s.ein.Aufbau = func(context.Context, aufbauEingabe) (aufbauErgebnis, error) {
				return aufbauErgebnis{}, &aufbauFehler{grund: grundAufbau, stelle: stelleZahl, ursache: errors.New("x")}
			}
		default:
			s.dienst.halte = func(nr, k int, _ *http.Request) {
				if nr != 1 || k != 10 {
					return
				}
				if f.name == "Abbruch" {
					cancel()
					return
				}
				s.uhr.Vorstellen(start.Add(s.ein.Profil.Gesamtfrist).Sub(s.uhr.Jetzt()))
				bisWahr(func() bool { return beendet(lauf) })
			}
		}
		code := warteAuf(t, s.fahre(ctx), 30*time.Second)
		cancel()
		if got := strings.Join(s.arten(), " "); code != f.exit || got != f.folge {
			t.Errorf("%s: Exit %d, Zeilen %s; erwartet %d und %s", f.name, code, got, f.exit, f.folge)
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Wege, erwartet 4", n)
	}
}

// TestHinweisImAblauf (SPEC-216 Nr. 9 d, 13 c): kein Hinweis bei Exit 0 und bei einem selbst beendeten Binary, binary_exit bei Exit 1,
// allein binary_kill, wenn SIGKILL nötig war (dessen Exit 137 ist).
func TestHinweisImAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		name    string
		term    int
		stur    bool
		selbst  bool
		hinweis string
	}{
		{"Exit 0", 0, false, false, ""},
		{"Exit 1", 1, false, false, "lasttest: Hinweis betriebsart=replay binary_exit=1"},
		{"SIGKILL", 0, true, false, "lasttest: Hinweis betriebsart=replay binary_kill=ja"},
		{"selbst beendet", 1, false, true, ""},
	} {
		s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
		s.dienst.termExit, s.dienst.stur = f.term, f.stur
		if f.selbst {
			s.dienst.selbst = s.selbstNachAllen(1, 39)
		}
		exit := s.fahre(context.Background())
		if f.stur {
			treibeBis(t, s.uhr, 15*time.Second, s.dienst.signaleBei(1), s.dienst.signaleBei(2))
		}
		warteAuf(t, exit, 30*time.Second)
		if got := strings.Join(s.zeilenVon("Hinweis"), "|"); got != f.hinweis {
			t.Errorf("%s: %q, erwartet %q", f.name, got, f.hinweis)
		}
		if f.hinweis != "" && !slices.Equal(s.arten(), []string{"Start", "Aufbau", "Messung", "Last", "Hinweis", "Urteil", "Abnahme"}) {
			t.Errorf("%s: Folge %v", f.name, s.arten())
		}
		n++
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// TestFelderDerMessung (SPEC-216 Nr. 13): jedes Feld, das der Ablauf setzt, trägt in einem Lauf einen eigenen Wert: Lauf 2 der Betriebsart
// mock mit drei Fehlern, sieben fremden Zeilen, fünf gleichzeitigen Anfragen, 13 ms Verspätung, 4321 Bytes, 2,5 s CPU-Zeit.
func TestFelderDerMessung(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerProbestand(t, kurzesProfil("klein", 2), "mock")
	s.cpu = []time.Duration{0, 0, time.Second, 3500 * time.Millisecond}
	s.dienst.status = func(nr, k int) int {
		if nr == 1 && (k == 1 || k == 22 || k == 38) {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}
	s.dienst.body = func(nr, k int) string {
		n := 108
		if k == 7 {
			n++
		}
		return strings.Repeat("b", n)
	}
	s.dienst.dauer = func(nr, k int) (int, int) {
		kind := 2000
		if k == 2 || k == 25 || k == 30 {
			kind = -1
		}
		return (k + 1) * 1000, kind
	}
	s.dienst.beiStart = func(b *ersatzBinaer) {
		for range 7 {
			b.zeile("Textzeile des Binary")
		}
	}
	// Hinter dem Nachlauf: sechs Kinder ohne Eltern und ein Gesamtspan außerhalb der Messphase.
	s.dienst.signal = func(b *ersatzBinaer, _ os.Signal) {
		for i := range 6 {
			b.zeile(kindZeile("mock", "ohne-"+strings.Repeat("x", i+1), 5000))
		}
		b.zeile(gesamtZeile("mock", "extra", 1000))
	}
	alle := make(chan struct{})
	s.dienst.halte = func(nr, k int, _ *http.Request) {
		switch {
		case nr == 1 && k == 39:
			close(alle)
		case nr == 1 && k >= 35:
			<-alle
		}
	}
	vorher := s.uhr.vorher
	// Bis auf die gehaltenen Anfragen 35 bis 39 ist jede fällige beendet, bevor die nächste startet, damit allein jene gleichzeitig offen sind.
	s.uhr.vorher = func(d time.Duration, ziel time.Time) {
		vorher(d, ziel)
		if b := s.dienst.aktuell(); b != nil && b.nr == 1 {
			n := int32(min(ziel.Sub(b.start)/(50*time.Millisecond), 35))
			bisWahr(func() bool { return b.fertig.Load() >= n })
			pause(time.Millisecond)
		}
	}
	s.uhr.extra = func(ziel time.Time) time.Duration {
		if b := s.dienst.aktuell(); b != nil && b.nr == 1 && ziel.Sub(b.start) == 7*50*time.Millisecond {
			return 13 * time.Millisecond
		}
		return 0
	}
	s.dienst.termExit = 1
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	m := s.zeilenVon("Messung")
	want := "lasttest: Messung betriebsart=mock lauf=2 anfragen=40 spans=41 messphase=20 fehler=3 fremde_zeilen=7 p95_ms=39.000 " +
		"max_ms=40.000 kind=18 kind_p95_ms=2.000 kind_anteil=6.5 ohne_kind=4 kind_ohne_eltern=6 gleichzeitig_max=5 " +
		"verspaetung_max_ms=13.000 bytes=4321 treiber_cpu_s=2.500 gueltig=nein grund=fehler"
	if code != 3 || len(m) != 2 || m[1] != want {
		t.Errorf("Exit %d, Zeile %q, erwartet %q", code, m, want)
	}
	if h := s.zeilenVon("Hinweis"); len(h) != 2 || h[1] != "lasttest: Hinweis betriebsart=mock binary_exit=1" {
		t.Errorf("Hinweise %q", h)
	}
	if l := s.zeilenVon("Last"); len(l) != 2 || l[1] != "lasttest: Last betriebsart=mock lauf=2 proben=2 min=1.00 max=1.00 schwelle=10.00" {
		t.Errorf("Last %q", l)
	}
	if a := s.zeilenVon("Aufbau"); a[0] != "lasttest: Aufbau exchanges=17 entities=40 daten_mib=0 dauer_s=0.000" {
		t.Errorf("%q", a)
	}
	if z := s.zeilenVon("Start"); z[0] != "lasttest: Start profil=klein betriebsart=mock ziel_ms=100 startwert=7 kerne=20" {
		t.Errorf("%q", z)
	}
}

// TestAufbauZeile: die Zeile Aufbau trägt die Dauer der Uhr und die Größe der Daten.
func TestAufbauZeile(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
	s.ein.Aufbau = func(context.Context, aufbauEingabe) (aufbauErgebnis, error) {
		s.uhr.Vorstellen(1234567 * time.Microsecond)
		if err := os.WriteFile(filepath.Join(s.ein.DataDir, "db"), make([]byte, 5<<20+1), 0o600); err != nil {
			t.Error(err)
		}
		return aufbauErgebnis{Session: "s-1", Workspace: "w-1", Exchanges: 17, Entities: 40}, nil
	}
	warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	if a := s.zeilenVon("Aufbau"); len(a) != 1 || a[0] != "lasttest: Aufbau exchanges=17 entities=40 daten_mib=5 dauer_s=1.234" {
		t.Errorf("%q", a)
	}
}

// TestVerstoesseImAblauf (SPEC-216 Nr. 10 h, 9 c): ein Lauf mit allein einem Fehler, allein einer Probe über der Schwelle und allein einem
// Binary, das nach allen Spans selbst endet, trägt je diesen Grund.
func TestVerstoesseImAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, f := range []struct {
		grund, meldung string
		setze          func(*probestand)
	}{
		{"fehler", "", func(s *probestand) {
			s.dienst.status = func(_, k int) int {
				if k == 3 {
					return http.StatusNotFound
				}
				return http.StatusOK
			}
		}},
		{"last", "", func(s *probestand) {
			s.last.inhalt = func(i int) string {
				if i == 2 {
					return "20 1 1"
				}
				return "1 1 1"
			}
		}},
		{"binary", "lasttest: Binary endete (Exit 0)\n", func(s *probestand) {
			s.dienst.selbst = s.selbstNachAllen(0, 39)
		}},
	} {
		s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
		f.setze(s)
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		m := s.zeilenVon("Messung")
		if code != 3 || feld(m[0], "grund") != f.grund || feld(m[0], "spans") != "40" || s.fehler.String() != f.meldung {
			t.Errorf("%s: Exit %d, %q, Meldungen %q", f.grund, code, m, s.fehler.String())
		}
		n++
	}
	if n != 3 {
		t.Errorf("%d Verstöße, erwartet 3", n)
	}
}

// TestStartScheitertImAblauf (SPEC-216 Nr. 9 c): ein Port oder ein Start, der scheitert, ist ein Lauf mit grund=binary ohne Anfrage; die
// folgenden Läufe fahren.
func TestStartScheitertImAblauf(t *testing.T) {
	wache(t, 60*time.Second)
	for _, was := range []string{"Port", "Start"} {
		s := neuerProbestand(t, kurzesProfil("klein", 2), "replay")
		if was == "Port" {
			s.ein.Port = func(context.Context) (string, error) { return "", errors.New("kein Port") }
		} else {
			s.ein.Starte = func(context.Context, []string) (*prozess, error) { return nil, errors.New("fehlt") }
		}
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		m := s.zeilenVon("Messung")
		if code != 3 || len(m) != 2 || feld(m[0], "grund") != "binary" || feld(m[1], "grund") != "binary" || feld(m[0], "anfragen") != "0" ||
			len(s.zeilenVon("Last")) != 2 || len(s.zeilenVon("Hinweis")) != 0 || len(s.dienst.liste()) != 0 ||
			s.fehler.String() != strings.Repeat("lasttest: Binary startet nicht\n", 2) {
			t.Errorf("%s: Exit %d, %q, Meldungen %q", was, code, s.zeilen(), s.fehler.String())
		}
	}
}
