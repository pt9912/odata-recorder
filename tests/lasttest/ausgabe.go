package main

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// errToken ist der Fehler einer Zeile, deren grund keines der 13 Tokens ist.
var errToken = errors.New("grund außerhalb der Tokens")

// tokens sind die 13 Werte des Schlüssels grund (SPEC-216 Nr. 13).
func tokens() []string {
	return []string{
		grundOK, grundFehler, grundZaehlung, grundModus, grundLast, grundBinary, grundStreuung, grundAufbau, grundFrist,
		grundProfil, grundZiel, grundBetriebsart, grundErgebnis,
	}
}

// zeilenart ist eine Zeile der Ausgabe (SPEC-216 Nr. 13); der Fehler ist errToken für einen grund außerhalb der Tokens.
type zeilenart interface {
	zeile() (string, error)
}

// schreibe schreibt eine Zeile mit Zeilenende; ein Fehler der Zeile schreibt nichts.
func schreibe(w io.Writer, z zeilenart) error {
	s, err := z.zeile()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s+"\n")
	return err
}

// bauen bildet `lasttest: <Art> <schluessel>=<wert> …` aus Paaren von Schlüssel und Wert.
func bauen(art string, paare ...string) string {
	var b strings.Builder
	b.WriteString("lasttest: " + art)
	for i := 0; i+1 < len(paare); i += 2 {
		b.WriteString(" " + paare[i] + "=" + paare[i+1])
	}
	return b.String()
}

// tausendstel bildet eine Zahl mit genau drei Nachkommastellen aus ganzen Tausendsteln: Mikrosekunden als Millisekunden, Millisekunden als Sekunden.
func tausendstel(n int64) string { return fmt.Sprintf("%d.%03d", n/1000, n%1000) }

// zehntel bildet den Anteil in Zehntelprozent mit einer Nachkommastelle.
func zehntel(n int) string { return fmt.Sprintf("%d.%d", n/10, n%10) }

// hundertstel bildet die Last mit zwei Nachkommastellen.
func hundertstel(n int) string { return fmt.Sprintf("%d.%02d", n/100, n%100) }

func ganz(n int) string { return strconv.Itoa(n) }

func janein(b bool) string {
	if b {
		return "ja"
	}
	return "nein"
}

// pruefeGrund liefert errToken, wenn g keines der Tokens ist.
func pruefeGrund(g string) error {
	if !slices.Contains(tokens(), g) {
		return errToken
	}
	return nil
}

// startZeile ist die Zeile `Start`.
type startZeile struct {
	Profil      string
	Betriebsart string
	ZielMs      int64
	Startwert   uint64
	Kerne       int
}

func (s startZeile) zeile() (string, error) {
	return bauen("Start", "profil", s.Profil, "betriebsart", s.Betriebsart, "ziel_ms", strconv.FormatInt(s.ZielMs, 10),
		"startwert", strconv.FormatUint(s.Startwert, 10), "kerne", ganz(s.Kerne)), nil
}

// aufbauZeile ist die Zeile `Aufbau`.
type aufbauZeile struct {
	Exchanges int
	Entities  int
	DatenMiB  int
	Dauer     time.Duration
}

func (a aufbauZeile) zeile() (string, error) {
	return bauen("Aufbau", "exchanges", ganz(a.Exchanges), "entities", ganz(a.Entities), "daten_mib", ganz(a.DatenMiB),
		"dauer_s", tausendstel(a.Dauer.Milliseconds())), nil
}

func (m Messung) zeile() (string, error) {
	if err := pruefeGrund(m.Grund); err != nil {
		return "", err
	}
	return bauen("Messung", "betriebsart", m.Betriebsart, "lauf", ganz(m.Lauf), "anfragen", ganz(m.Anfragen), "spans", ganz(m.Spans),
		"messphase", ganz(m.Messphase), "fehler", ganz(m.Fehler), "fremde_zeilen", ganz(m.FremdeZeilen), "p95_ms", tausendstel(m.P95),
		"max_ms", tausendstel(m.Max), "kind", ganz(m.Kind), "kind_p95_ms", tausendstel(m.KindP95), "kind_anteil", zehntel(m.KindAnteil),
		"ohne_kind", ganz(m.OhneKind), "kind_ohne_eltern", ganz(m.KindOhneEltern), "gleichzeitig_max", ganz(m.GleichzeitigMax),
		"verspaetung_max_ms", tausendstel(m.VerspaetungMax), "bytes", strconv.FormatInt(m.Bytes, 10),
		"treiber_cpu_s", tausendstel(m.TreiberCPU.Milliseconds()), "gueltig", janein(m.Gueltig), "grund", m.Grund), nil
}

// lastZeile ist die Zeile `Last` eines Laufs.
type lastZeile struct {
	Betriebsart string
	Lauf        int
	Ergebnis    lastErgebnis
}

func (l lastZeile) zeile() (string, error) {
	return bauen("Last", "betriebsart", l.Betriebsart, "lauf", ganz(l.Lauf), "proben", ganz(l.Ergebnis.Proben),
		"min", hundertstel(l.Ergebnis.Min), "max", hundertstel(l.Ergebnis.Max), "schwelle", hundertstel(l.Ergebnis.Schwelle)), nil
}

func (u urteil) zeile() (string, error) {
	if err := pruefeGrund(u.Grund); err != nil {
		return "", err
	}
	return bauen("Urteil", "betriebsart", u.Betriebsart, "laeufe", ganz(u.Laeufe), "ziel_ms", strconv.FormatInt(u.ZielMs, 10),
		"p95_ms", tausendstel(u.P95), "streuung_ms", tausendstel(u.Streuung), "toleranz_ms", tausendstel(u.Toleranz),
		"ergebnis", u.Ergebnis, "grund", u.Grund), nil
}

// abnahmeZeile ist die Zeile `Abnahme`.
type abnahmeZeile struct {
	ZielMs  int64
	Abnahme bool
	Grund   string
}

func (a abnahmeZeile) zeile() (string, error) {
	if err := pruefeGrund(a.Grund); err != nil {
		return "", err
	}
	return bauen("Abnahme", "ziel_ms", strconv.FormatInt(a.ZielMs, 10), "abnahmewert_ms", strconv.Itoa(abnahmewertMs),
		"abnahme", janein(a.Abnahme), "grund", a.Grund), nil
}

// hinweis ist die Zeile `Hinweis`: der Exit des Binary (Kill leer) oder ein nötiges SIGKILL (SPEC-216 Nr. 9 d).
type hinweis struct {
	Betriebsart string
	Exit        int
	Kill        bool
}

func (h hinweis) zeile() (string, error) {
	if h.Kill {
		return bauen("Hinweis", "betriebsart", h.Betriebsart, "binary_kill", "ja"), nil
	}
	return bauen("Hinweis", "betriebsart", h.Betriebsart, "binary_exit", ganz(h.Exit)), nil
}
