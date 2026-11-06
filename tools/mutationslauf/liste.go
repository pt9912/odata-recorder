package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	listeMax    = 1 << 20
	mutationMax = 200
	rateMax     = 200
	rateVorgabe = 50

	klasseNebenlaeufig = "nebenläufig"
	klasseAndere       = "andere"
)

// Mutation ist ein Eintrag der Liste (SPEC-210 Nr. 2) mit den Werten, die die Prüfung auflöst.
type Mutation struct {
	Name, Datei, Alt, Neu, Test, Klasse string
	NDetektor, NOhne                    int
	Modul                               Modul  // Modul von Datei
	Paket                               string // Verzeichnis des Pakets relativ zur Wurzel des Baums
	modulOK                             bool   // Modul ist aufgelöst
}

// Nebenlaeufig sagt, ob die Rolle die Mutation als nebenläufig geführt hat.
func (m Mutation) Nebenlaeufig() bool { return m.Klasse == klasseNebenlaeufig }

// eintrag hält die geparsten Rohwerte eines Listeneintrags.
type eintrag struct {
	text      map[string]string
	zahl      map[string]int
	da        map[string]bool
	unbekannt []string
}

func textFelder() []string { return []string{"name", "datei", "alt", "neu", "test", "klasse", "paket"} }
func zahlFelder() []string { return []string{"n_detektor", "n_ohne"} }

// leseEintrag liest die Felder eines Eintrags; ein falscher Typ ist ein Fehler der ganzen Liste.
func leseEintrag(n int, roh json.RawMessage) (eintrag, error) {
	e := eintrag{text: map[string]string{}, zahl: map[string]int{}, da: map[string]bool{}}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(roh, &obj); err != nil || obj == nil {
		return e, fmt.Errorf("Mutation %d: kein Objekt", n)
	}
	for k, v := range obj {
		var err error
		switch {
		case slices.Contains(textFelder(), k):
			var s string
			err = json.Unmarshal(v, &s)
			e.text[k] = s
		case slices.Contains(zahlFelder(), k):
			var z int
			err = json.Unmarshal(v, &z)
			e.zahl[k] = z
		default:
			e.unbekannt = append(e.unbekannt, k)
			continue
		}
		if err != nil || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return e, fmt.Errorf("Mutation %d Feld %s: falscher Typ", n, k)
		}
		e.da[k] = true
	}
	sort.Strings(e.unbekannt)
	return e, nil
}

// pruefer sammelt die Befunde der Liste (SPEC-210 Nr. 2 d).
type pruefer struct {
	baum    string
	gcc     bool
	namen   map[string]bool
	befunde []string
}

func (p *pruefer) befund(n int, name, text string) {
	p.befunde = append(p.befunde, fmt.Sprintf("Mutation %d (%s): %s", n, name, text))
}

// wurzelElemente liest das Wurzelobjekt; der zweite Wert ist der Grund, wenn die Liste nicht lesbar ist.
func wurzelElemente(roh []byte) ([]json.RawMessage, string) {
	switch {
	case len(roh) > listeMax:
		return nil, "größer als 1 MiB"
	case bytes.HasPrefix(roh, []byte("\xef\xbb\xbf")):
		return nil, "BOM"
	case !utf8.Valid(roh):
		return nil, "kein UTF-8"
	}
	var wurzel map[string]json.RawMessage
	if err := json.Unmarshal(roh, &wurzel); err != nil || wurzel == nil {
		return nil, "kein JSON-Objekt"
	}
	for k := range wurzel {
		if k != "mutationen" {
			return nil, "Feld " + k + " unbekannt"
		}
	}
	liste, da := wurzel["mutationen"]
	if !da {
		return nil, "Feld mutationen fehlt"
	}
	var elemente []json.RawMessage
	if err := json.Unmarshal(liste, &elemente); err != nil || elemente == nil {
		return nil, "Feld mutationen: falscher Typ"
	}
	return elemente, ""
}

// leseListe parst und prüft die ganze Liste; ohne Befund sind die Mutationen aufgelöst.
func leseListe(roh []byte, baum string, gcc bool) ([]Mutation, []string) {
	nl := func(grund string) ([]Mutation, []string) { return nil, []string{"Liste nicht lesbar: " + grund} }
	elemente, grund := wurzelElemente(roh)
	if grund != "" {
		return nl(grund)
	}
	if len(elemente) == 0 {
		return nil, []string{"Liste leer"}
	}
	if len(elemente) > mutationMax {
		return nil, []string{"Liste zu lang"}
	}
	eintraege := make([]eintrag, len(elemente))
	for i, r := range elemente {
		e, err := leseEintrag(i+1, r)
		if err != nil {
			return nl(err.Error())
		}
		eintraege[i] = e
	}
	p := &pruefer{baum: baum, gcc: gcc, namen: map[string]bool{}}
	muts := make([]Mutation, len(eintraege))
	for i, e := range eintraege {
		muts[i] = p.pruefe(i+1, e)
	}
	if len(p.befunde) > 0 {
		return nil, p.befunde
	}
	return muts, nil
}

// pfadGueltig prüft einen relativen Pfad in Schrägstrich-Schreibweise (SPEC-210 Nr. 2 b).
func pfadGueltig(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.ContainsRune(p, 0) || strings.Contains(p, "\\") {
		return false
	}
	for _, s := range strings.Split(p, "/") {
		if s == "" || s == "." || s == ".." {
			return false
		}
	}
	return true
}

func zeichenMenge(s string, erlaubt func(r rune) bool, lo, hi int) bool {
	if n := utf8.RuneCountInString(s); n < lo || n > hi {
		return false
	}
	for _, r := range s {
		if !erlaubt(r) {
			return false
		}
	}
	return true
}

func nameZeichen(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-'
}

func testZeichen(r rune) bool { return nameZeichen(r) && r != '.' && r != '-' }

// modulVon liefert das Modul mit der nächsten go.mod oberhalb von relDir (SPEC-210 Nr. 4 a).
func modulVon(baum, relDir string) (Modul, bool) {
	for d := relDir; ; d = path.Dir(d) {
		if d == "." {
			d = ""
		}
		if fi, err := os.Lstat(unter(baum, path.Join(d, "go.mod"))); err == nil && fi.Mode().IsRegular() {
			return Modul{Dir: d}, true
		}
		if d == "" {
			return Modul{}, false
		}
	}
}

// pruefe prüft einen Eintrag in der Reihenfolge der Felder (SPEC-210 Nr. 2 a, d).
func (p *pruefer) pruefe(n int, e eintrag) Mutation {
	m := Mutation{Name: e.text["name"], Datei: e.text["datei"], Alt: e.text["alt"], Neu: e.text["neu"],
		Test: e.text["test"], Klasse: e.text["klasse"], NDetektor: rateVorgabe, NOhne: rateVorgabe}
	b := func(text string) { p.befund(n, m.Name, text) }
	fehlt := func(f string) bool {
		if !e.da[f] {
			b("Feld " + f + " fehlt")
		}
		return !e.da[f]
	}
	// name
	if !fehlt("name") {
		switch {
		case !zeichenMenge(m.Name, nameZeichen, 1, 64):
			b("name ungültig")
		case p.namen[m.Name]:
			b("Name doppelt")
		}
		p.namen[m.Name] = true
	}
	p.pruefeDatei(n, &m, fehlt)
	if !fehlt("test") && (!strings.HasPrefix(m.Test, "Test") || !zeichenMenge(m.Test, testZeichen, 1, 128)) {
		b("test ungültig")
	}
	if !fehlt("klasse") {
		switch {
		case m.Klasse != klasseNebenlaeufig && m.Klasse != klasseAndere:
			b("klasse unbekannt")
		case m.Nebenlaeufig() && !p.gcc:
			b("Detektor nicht verfügbar")
		}
	}
	p.pruefePaket(n, &m, e)
	p.pruefeRate(n, &m, e)
	for _, k := range e.unbekannt {
		b("Feld " + k + " unbekannt")
	}
	return m
}

// dateiInhalt prüft datei und liefert den Inhalt, wenn sie als reguläre .go-Datei in einem Modul des Baums liegt.
func (p *pruefer) dateiInhalt(n int, m *Mutation) ([]byte, bool) {
	if !pfadGueltig(m.Datei) || !strings.HasSuffix(m.Datei, ".go") {
		p.befund(n, m.Name, "datei ungültig")
		return nil, false
	}
	fi, err := os.Lstat(unter(p.baum, m.Datei))
	if err != nil || !fi.Mode().IsRegular() {
		p.befund(n, m.Name, "Datei fehlt")
		return nil, false
	}
	inhalt, err := os.ReadFile(unter(p.baum, m.Datei))
	if err != nil {
		p.befund(n, m.Name, "Datei fehlt")
		return nil, false
	}
	m.Modul, m.modulOK = modulVon(p.baum, path.Dir(m.Datei))
	if !m.modulOK {
		p.befund(n, m.Name, "datei ungültig")
	}
	return inhalt, m.modulOK
}

// pruefeDatei prüft datei, alt und neu, die am Inhalt der Datei hängen.
func (p *pruefer) pruefeDatei(n int, m *Mutation, fehlt func(string) bool) {
	b := func(text string) { p.befund(n, m.Name, text) }
	var inhalt []byte
	lesbar := false
	if !fehlt("datei") {
		inhalt, lesbar = p.dateiInhalt(n, m)
	}
	if !fehlt("alt") {
		switch c := bytes.Count(inhalt, []byte(m.Alt)); {
		case m.Alt == "":
			b("alt leer")
		case lesbar && c == 0:
			b("alt trifft nicht")
		case lesbar && c > 1:
			b("alt trifft mehrfach")
		}
	}
	if !fehlt("neu") && m.Neu == m.Alt && m.Alt != "" {
		b("neu ändert nichts")
	}
}

// pruefePaket löst paket auf (SPEC-210 Nr. 2 b): Vorgabe ist das Verzeichnis von datei.
func (p *pruefer) pruefePaket(n int, m *Mutation, e eintrag) {
	m.Paket = path.Dir(m.Datei)
	if !e.da["paket"] {
		return
	}
	pk := e.text["paket"]
	if !pfadGueltig(pk) {
		p.befund(n, m.Name, "paket ungültig")
		return
	}
	m.Paket = pk
	if fi, err := os.Lstat(unter(p.baum, pk)); err != nil || !fi.IsDir() {
		p.befund(n, m.Name, "Paket fehlt")
	} else if mod, ok := modulVon(p.baum, pk); ok && m.modulOK && mod != m.Modul {
		p.befund(n, m.Name, "paket außerhalb des Moduls")
	}
}

// pruefeRate prüft n_detektor und n_ohne gegen die Klasse (SPEC-210 Nr. 2 b).
func (p *pruefer) pruefeRate(n int, m *Mutation, e eintrag) {
	felder := []struct {
		name string
		ziel *int
	}{{"n_detektor", &m.NDetektor}, {"n_ohne", &m.NOhne}}
	for _, f := range felder {
		if !e.da[f.name] {
			continue
		}
		z := e.zahl[f.name]
		switch {
		case m.Klasse == klasseAndere:
			p.befund(n, m.Name, "Rate bei Klasse andere")
		case z < 0 || z > rateMax:
			p.befund(n, m.Name, f.name+" ungültig")
		default:
			*f.ziel = z
		}
	}
	if m.Nebenlaeufig() && m.NDetektor == 0 && m.NOhne == 0 {
		p.befund(n, m.Name, "Rate ohne Prozess")
	}
}
