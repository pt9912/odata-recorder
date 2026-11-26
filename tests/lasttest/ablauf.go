package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

// Die Variablen, die das Programm liest; der Treiber setzt jede mit dem wirksamen Wert (SPEC-216 Nr. 3 b).
const (
	varBetriebsart = "LASTTEST_BETRIEBSART"
	varProfil      = "LASTTEST_PROFIL"
	varZielMs      = "LASTTEST_ZIEL_MS"
	varStartwert   = "LASTTEST_STARTWERT"
	varKerne       = "LASTTEST_KERNE"

	betriebBeide = "beide"
)

// werte sind die Eingaben eines Aufrufs.
type werte struct {
	Profil      Profil
	Betriebsart string
	ZielMs      int64
	Startwert   uint64
	Kerne       int
}

// liesUmgebung liest die fünf Variablen über lookup; eine fehlende oder unlesbare ist ein Fehler, der ihren Namen nennt.
func liesUmgebung(lookup func(string) (string, bool)) (werte, error) {
	var w werte
	lies := func(name string, setze func(string) bool) error {
		v, ok := lookup(name)
		if !ok || !setze(v) {
			return errors.New("Umgebung: " + name + " fehlt oder ist ungültig")
		}
		return nil
	}
	err := errors.Join(
		lies(varBetriebsart, func(v string) bool {
			w.Betriebsart = v
			return v == betriebBeide || v == string(betriebReplay) || v == string(betriebMock)
		}),
		lies(varProfil, func(v string) bool {
			for _, p := range []Profil{abnahme(), rauch()} {
				if p.Name == v {
					w.Profil = p
				}
			}
			return w.Profil.Name != ""
		}),
		lies(varZielMs, func(v string) bool {
			n, err := strconv.ParseInt(v, 10, 64)
			w.ZielMs = n
			return err == nil && n > 0
		}),
		lies(varStartwert, func(v string) bool {
			n, err := strconv.ParseUint(v, 10, 64)
			w.Startwert = n
			return err == nil
		}),
		lies(varKerne, func(v string) bool {
			n, err := strconv.Atoi(v)
			w.Kerne = n
			return err == nil && n > 0
		}),
	)
	return w, err
}

// betriebsarten sind die Betriebsarten eines Aufrufs in der Folge der Läufe: replay vor mock.
func betriebsarten(wahl string) []Betriebsart {
	if wahl == betriebBeide {
		return []Betriebsart{betriebReplay, betriebMock}
	}
	return []Betriebsart{Betriebsart(wahl)}
}

// ablaufEingabe sind die Werte eines Aufrufs und die Stellen, an denen das Programm seine Umwelt berührt: Uhr, Last des Hosts, Aufbau,
// Wahl des Ports, Start des Binary, Verbindungsaufbau der Bereitschaft, Transport der Last (nil: je Lauf ein eigener), CPU-Zeit des
// Programms und die Ströme.
type ablaufEingabe struct {
	werte
	Uhr       Uhr
	Last      lastQuelle
	DataDir   string
	Aufbau    func(ctx context.Context, e aufbauEingabe) (aufbauErgebnis, error)
	Port      func(ctx context.Context) (string, error)
	Starte    func(ctx context.Context, env []string) (*prozess, error)
	Verbinde  func(ctx context.Context, adresse string) error
	Transport http.RoundTripper
	CPU       func() time.Duration
	Aus       io.Writer
	Err       io.Writer
}

// schreibe schreibt eine Zeile der Ausgabe; ein Fehler des Schreibens ändert den Ablauf nicht.
func (e ablaufEingabe) schreibe(z zeilenart) { _ = schreibe(e.Aus, z) }

// melde schreibt eine Meldung nach stderr.
func (e ablaufEingabe) melde(text string) { fmt.Fprintf(e.Err, "lasttest: %s\n", text) }

// unterbrochen meldet den Abbruch (signale ist beendet) vor der abgelaufenen Gesamtfrist (Ursache von ctx) und liefert den Exit.
func (e ablaufEingabe) unterbrochen(signale, ctx context.Context) (int, bool) {
	switch {
	case signale.Err() != nil:
		e.melde("abgebrochen")
		return exitCode(ausgang{Abbruch: true}), true
	case errors.Is(context.Cause(ctx), errFrist):
		e.melde("Gesamtfrist abgelaufen")
		return exitCode(ausgang{Frist: true}), true
	}
	return 0, false
}

// aufbauMeldung ist die Meldung eines gescheiterten Aufbaus (SPEC-216 Nr. 6 e, f und 11): das Token kommt über errors.As aus dem
// *aufbauFehler, nie aus dem Text des Fehlers.
func aufbauMeldung(err error) string {
	grund := grundAufbau
	var f *aufbauFehler
	switch {
	case errors.Is(err, errLast):
		grund = grundLast
	case errors.As(err, &f):
		if f.stelle == stelleDaten {
			return f.Error()
		}
		grund = f.grund
	}
	return "Aufbau gescheitert: " + grund
}

// datenMiB ist die Summe der Größen der Dateien unter dir in ganzen MiB, abgeschnitten.
func datenMiB(dir string) int {
	var summe int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			if fi, err := d.Info(); err == nil {
				summe += fi.Size()
			}
		}
		return nil
	})
	return int(summe >> 20)
}

// baueAuf nimmt die Probe der Last vor dem Aufbau und baut auf, falls sie zulässig ist (SPEC-216 Nr. 11); die Zeile trägt Zahlen und Dauer.
func (e ablaufEingabe) baueAuf(ctx context.Context) (erg aufbauErgebnis, zeile aufbauZeile, err error) {
	err = vorAufbau(ctx, e.Last, e.Kerne, func(ctx context.Context) (err error) {
		start := e.Uhr.Jetzt()
		erg, err = e.Aufbau(ctx, aufbauEingabe{Profil: e.Profil, Startwert: e.Startwert, Uhr: e.Uhr, DataDir: e.DataDir})
		zeile.Dauer = e.Uhr.Jetzt().Sub(start)
		return err
	})
	zeile.Exchanges, zeile.Entities, zeile.DatenMiB = erg.Exchanges, erg.Entities, datenMiB(e.DataDir)
	return erg, zeile, err
}

// fahre ist der Ablauf eines Aufrufs (SPEC-216 Nr. 4 b und 13 d) und liefert den Exit: Start, Aufbau, je Betriebsart die Läufe mit ihren
// Zeilen, danach die Urteile und die Abnahme. Das Ende von signale bricht ab; die Gesamtfrist des Profils beginnt mit dem Aufruf. Beide
// enden ohne Urteil.
func (e ablaufEingabe) fahre(signale context.Context) int {
	ctx, cancel := mitFrist(signale, e.Uhr, e.Profil.Gesamtfrist)
	defer cancel()
	e.schreibe(startZeile{
		Profil: e.Profil.Name, Betriebsart: e.Betriebsart, ZielMs: e.ZielMs, Startwert: e.Startwert, Kerne: e.Kerne,
	})
	erg, zeile, err := e.baueAuf(ctx)
	if code, ende := e.unterbrochen(signale, ctx); ende {
		return code
	}
	if err != nil {
		e.melde(aufbauMeldung(err))
		return exitCode(ausgang{Aufbau: true})
	}
	e.schreibe(zeile)
	var urteile []urteil
	for _, b := range betriebsarten(e.Betriebsart) {
		var laeufe []Messung
		for nr := 1; nr <= e.Profil.Laeufe; nr++ {
			l := e.lauf(ctx, b, nr, erg)
			if code, ende := e.unterbrochen(signale, ctx); ende {
				return code
			}
			e.schreibe(l.messung)
			e.schreibe(l.last)
			if l.hinweis != nil {
				e.schreibe(*l.hinweis)
			}
			laeufe = append(laeufe, l.messung)
		}
		urteile = append(urteile, urteilBetriebsart(string(b), laeufe, e.ZielMs))
	}
	return e.urteile(urteile)
}

// urteile schreibt die Zeilen Urteil und Abnahme und liefert den Exit.
func (e ablaufEingabe) urteile(urteile []urteil) int {
	var a ausgang
	var ergebnisse []string
	for _, u := range urteile {
		e.schreibe(u)
		ergebnisse = append(ergebnisse, u.Ergebnis)
		if u.Betriebsart == string(betriebReplay) {
			a.Replay = u.Ergebnis
		} else {
			a.Mock = u.Ergebnis
		}
	}
	gilt, grund := abnahmeGilt(e.Profil.Name, e.ZielMs, e.Betriebsart, ergebnisse...)
	e.schreibe(abnahmeZeile{ZielMs: e.ZielMs, Abnahme: gilt, Grund: grund})
	return exitCode(a)
}

// liste ist die Liste der Anfragen eines Laufs (SPEC-216 Nr. 7).
func (e ablaufEingabe) liste(b Betriebsart, nr int, erg aufbauErgebnis) []Anfrage {
	if b == betriebReplay {
		return listeReplay(e.Profil, e.Startwert, nr, erg.Workspace, erg.Session)
	}
	return listeMock(e.Profil, e.Startwert, erg.Workspace)
}

// starte wählt den Port und startet das Binary der Betriebsart; nil heißt, der Start scheitert.
func (e ablaufEingabe) starte(ctx context.Context, b Betriebsart) (*prozess, string) {
	adresse, err := e.Port(ctx)
	if err != nil {
		return nil, ""
	}
	p, err := e.Starte(ctx, umgebung(b, adresse))
	if err != nil {
		return nil, ""
	}
	return p, adresse
}

// laufErgebnis sind die Zeilen eines Laufs; hinweis ist nil ohne Zeile Hinweis.
type laufErgebnis struct {
	messung Messung
	last    lastZeile
	hinweis *hinweis
}

// lastStand ist, was die Last eines Laufs zählt.
type lastStand struct {
	takt      taktStand
	fehler    atomic.Int64
	bytes     atomic.Int64
	nachlauf  bool
	bereit    bereitschaft
	gestartet bool
}

// sende fährt Takt und Nachlauf gegen das bereite Binary unter adresse (SPEC-216 Nr. 8 und 10 e).
func (e ablaufEingabe) sende(ctx context.Context, adresse string, liste []Anfrage, s *sammler, st *lastStand) {
	client := lastClient(e.Transport)
	defer client.CloseIdleConnections()
	st.takt = fahreTakt(ctx, e.Uhr, e.Profil.Rate, len(liste), func(ctx context.Context, k int) {
		a := liste[k]
		n, err := holeLast(ctx, e.Uhr, client, "http://"+adresse+a.Ziel(), a.Header)
		st.bytes.Add(n)
		if err != nil {
			st.fehler.Add(1)
		}
	})
	st.nachlauf = s.warteAufSpans(ctx, e.Uhr, st.takt.Gestartet)
}

// lauf fährt einen Lauf (SPEC-216 Nr. 4 b): Binary starten, Probe, Takt, Nachlauf, Probe, Binary beenden, auswerten. Die Meldungen zum Binary
// stehen nur, solange ctx nicht beendet ist.
func (e ablaufEingabe) lauf(ctx context.Context, b Betriebsart, nr int, erg aufbauErgebnis) laufErgebnis {
	cpu := e.CPU()
	liste := e.liste(b, nr, erg)
	s := neuerSammler(string(b))
	p, adresse := e.starte(ctx, b)
	st := &lastStand{bereit: nichtBereit, nachlauf: true, gestartet: p != nil}
	stroeme := func() (int, bool) { return 0, false }
	if p != nil {
		stroeme = entleere(p, s.nimm)
		st.bereit = warteBereit(ctx, e.Uhr, e.Verbinde, adresse, p.Ende)
	}
	proben := starteProben(ctx, e.Uhr, e.Last, e.Kerne, e.Profil.Warmup+e.Profil.Messphase)
	if st.bereit == bereit {
		e.sende(ctx, adresse, liste, s, st)
	}
	tot := e.binaryZustand(ctx, p, st)
	last := proben.beende()
	l := laufErgebnis{last: lastZeile{Betriebsart: string(b), Lauf: nr, Ergebnis: last}}
	l.hinweis = e.beende(ctx, string(b), p)
	fremd, lesefehler := stroeme()
	if lesefehler {
		e.melde("Lesefehler auf stderr des Binary")
	}
	je := e.Profil.Rate
	m, v := s.auswerte(je*int(e.Profil.Warmup/time.Second), je*int(e.Profil.Messphase/time.Second))
	v.Fehler = st.fehler.Load() > 0
	v.Zaehlung = v.Zaehlung || !st.nachlauf || lesefehler
	v.Last = !last.Zulaessig
	if tot {
		v = verstoesse{Binary: true}
	}
	m.Betriebsart, m.Lauf, m.Anfragen = string(b), nr, st.takt.Gestartet
	m.Fehler, m.FremdeZeilen, m.Bytes = int(st.fehler.Load()), fremd, st.bytes.Load()
	m.GleichzeitigMax, m.VerspaetungMax = st.takt.GleichzeitigMax, st.takt.VerspaetungMax.Microseconds()
	m.Gueltig, m.Grund = gueltigkeit(v)
	m.TreiberCPU = e.CPU() - cpu
	l.messung = m
	return l
}

// binaryZustand meldet ein Binary, das nicht bereit wurde oder vor dem Ende des Nachlaufs endete (SPEC-216 Nr. 9 c), und liefert, ob der
// Lauf allein den Verstoß binary trägt.
func (e ablaufEingabe) binaryZustand(ctx context.Context, p *prozess, st *lastStand) bool {
	selbst := st.gestartet && beendet(p.Ende)
	if st.bereit == bereit && !selbst {
		return false
	}
	switch {
	case ctx.Err() != nil:
	case selbst || st.bereit == vorherBeendet:
		e.melde("Binary endete (Exit " + strconv.Itoa(p.Exit()) + ")")
	default:
		e.melde("Binary startet nicht")
	}
	return true
}

// beende beendet ein noch laufendes Binary und liefert die Zeile Hinweis (SPEC-216 Nr. 9 d und 13 c): binary_kill nach einem nötigen
// SIGKILL, sonst binary_exit bei einem Exit ungleich 0. Ein Binary, das selbst endete, hat keinen Hinweis.
func (e ablaufEingabe) beende(ctx context.Context, betriebsart string, p *prozess) *hinweis {
	if p == nil || beendet(p.Ende) {
		return nil
	}
	if beendeBinaer(ctx, e.Uhr, p) {
		return &hinweis{Betriebsart: betriebsart, Kill: true}
	}
	if code := p.Exit(); code != 0 {
		return &hinweis{Betriebsart: betriebsart, Exit: code}
	}
	return nil
}
