package main

import (
	"fmt"
	"strings"
)

const (
	klasseNebenlaeufig = "nebenläufig"
	klasseAndere       = "andere"
	blockInfo          = "groessenmarke"
	prosaMarker        = "**Größenmarke**"
	anzahlZellen       = 5
	maxStellen         = 9
)

// kopfZellen sind die fünf Zellen der Kopfzeile der Tabelle (SPEC-128 Nr. 3 a).
func kopfZellen() []string {
	return []string{"Nr.", "Zusage", "Fänger", "Mutation", "Klasse"}
}

// zustand ist der Zustand eines Plans nach SPEC-128 Nr. 3 p.
type zustand int

const (
	keinBlock zustand = iota
	prosaForm
	markeAusstehend
	vollstaendig
)

// tabelle hält die Zahl der Zeilen je Klasse.
type tabelle struct {
	nebenlaeufig int64
	andere       int64
}

// eintrag ist `ueberschreitung` oder `ersetzt`: Obergrenze bzw. ältere Marke und Grund.
type eintrag struct {
	n     int64
	grund string
}

// block hält die Werte des Blocks `groessenmarke`; gesetzt nennt die Schlüssel, die er trägt.
type block struct {
	pfade           []string
	produktcode     int64
	fakes           int64
	basis           string
	schaetzung      int64
	marke           int64
	schnittGrund    string
	ersetzt         []eintrag
	ueberschreitung []eintrag
	gesetzt         map[string]bool
}

// plan ist das Ergebnis des Lesens: Zustand, Tabelle, Block und alle Eingabefehler.
type plan struct {
	zustand zustand
	tab     tabelle
	blk     block
	fehler  []string
}

// tabellenSchaetzung ist S aus der Tabelle, dem Produktcode und den Fakes des Blocks.
func (p plan) tabellenSchaetzung() int64 {
	return schaetzung(p.tab.nebenlaeufig, p.tab.andere, p.blk.fakes, p.blk.produktcode)
}

// lesePlan liest einen Plan (SPEC-128 Nr. 3). Ohne Block liest es die Tabelle nicht.
func lesePlan(text string) plan {
	aussen, bloecke := zerlege(text)
	var p plan
	if len(bloecke) == 0 {
		p.zustand = keinBlock
		for _, l := range aussen {
			if strings.HasPrefix(l, prosaMarker) {
				p.zustand = prosaForm
			}
		}
		return p
	}
	p.tab, p.fehler = leseTabelle(aussen)
	switch {
	case len(bloecke) > 1:
		p.fehler = append(p.fehler, "Block mehrdeutig")
	case !bloecke[0].geschlossen:
		p.fehler = append(p.fehler, "Block nicht geschlossen")
	default:
		var bf []string
		p.blk, bf = leseBlock(bloecke[0].zeilen)
		p.fehler = append(p.fehler, bf...)
	}
	p.zustand = markeAusstehend
	if p.blk.gesetzt["basis"] && p.blk.gesetzt["schaetzung"] && p.blk.gesetzt["marke"] {
		p.zustand = vollstaendig
	}
	return p
}

// rohBlock sind die Zeilen eines Blocks `groessenmarke` und ob sein Zaun schließt.
type rohBlock struct {
	zeilen      []string
	geschlossen bool
}

// zaun ist ein geöffneter umzäunter Block.
type zaun struct {
	zeichen byte
	laenge  int
	block   bool
}

// maxEinrueckung ist die größte Zahl Leerzeichen vor einem Zaun (SPEC-128 Nr. 3 q); vier
// Leerzeichen oder ein Tab davor machen die Zeile zu Text.
const maxEinrueckung = 3

// zaunBeginn erkennt eine Zeile ab drei Backticks oder drei Tilden nach null bis drei Leerzeichen;
// die Info-Zeichenkette ist der getrimmte Rest hinter dem Zaun und gleich `groessenmarke` für den
// Block. Bei Backticks darf der Rest keinen Backtick tragen, sonst ist die Zeile Text.
func zaunBeginn(l string) (zaun, bool) {
	s := strings.TrimLeft(l, " ")
	if len(l)-len(s) > maxEinrueckung || len(s) < 3 || (s[0] != '`' && s[0] != '~') {
		return zaun{}, false
	}
	n := 0
	for n < len(s) && s[n] == s[0] {
		n++
	}
	rest := s[n:]
	if n < 3 || (s[0] == '`' && strings.Contains(rest, "`")) {
		return zaun{}, false
	}
	return zaun{zeichen: s[0], laenge: n, block: strings.TrimSpace(rest) == blockInfo}, true
}

// schliessendDurch ist wahr für eine Zeile, die nur aus dem Zeichen des Zauns besteht, mindestens
// so lang wie der öffnende.
func (z zaun) schliessendDurch(l string) bool {
	t := strings.TrimSpace(l)
	return len(t) >= z.laenge && strings.Trim(t, string(z.zeichen)) == ""
}

// zerlege trennt die Zeilen außerhalb jedes Zauns von den Blöcken `groessenmarke`; an der Stelle
// eines Zauns steht in den Zeilen außerhalb eine leere Zeile, damit sie keine Tabelle verbindet.
func zerlege(text string) ([]string, []rohBlock) {
	var aussen []string
	var bloecke []rohBlock
	var z zaun
	offen := false
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSuffix(l, "\r")
		switch {
		case offen && z.schliessendDurch(l):
			offen = false
			if z.block {
				bloecke[len(bloecke)-1].geschlossen = true
			}
		case offen:
			if z.block {
				b := &bloecke[len(bloecke)-1]
				b.zeilen = append(b.zeilen, l)
			}
		default:
			nz, ok := zaunBeginn(l)
			if !ok {
				aussen = append(aussen, l)
				continue
			}
			z, offen = nz, true
			aussen = append(aussen, "")
			if z.block {
				bloecke = append(bloecke, rohBlock{})
			}
		}
	}
	return aussen, bloecke
}

// zellen trennt eine Tabellenzeile an `|`, außer an `\|`; der Anfangsstrich fällt weg, ein
// Schlussstrich erzeugt keine leere Zelle.
func zellen(l string) []string {
	s := strings.TrimRight(l, " \t\r")
	s = strings.TrimPrefix(s, "|")
	var out []string
	var cur strings.Builder
	ende := false
	for i := 0; i < len(s); i++ {
		ende = false
		switch {
		case s[i] == '\\' && i+1 < len(s) && s[i+1] == '|':
			cur.WriteByte('|')
			i++
		case s[i] == '|':
			out = append(out, cur.String())
			cur.Reset()
			ende = true
		default:
			cur.WriteByte(s[i])
		}
	}
	if !ende {
		out = append(out, cur.String())
	}
	return out
}

// istKopf ist wahr, wenn die Zeile aus genau den fünf Zellen des Kopfs besteht, getrimmt und
// bytegleich verglichen.
func istKopf(l string) bool {
	c := zellen(l)
	k := kopfZellen()
	if len(c) != len(k) {
		return false
	}
	for i := range c {
		if strings.TrimSpace(c[i]) != k[i] {
			return false
		}
	}
	return true
}

// istTrennzeile ist wahr, wenn die Zeile genau fünf Zellen hat und jede aus Strichen besteht, mit
// einem Doppelpunkt an einem oder an beiden Enden.
func istTrennzeile(l string) bool {
	c := zellen(l)
	if len(c) != anzahlZellen {
		return false
	}
	for _, z := range c {
		t := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(z), ":"), ":")
		if t == "" || strings.Trim(t, "-") != "" {
			return false
		}
	}
	return true
}

// leseTabelle sucht außerhalb der Zäune die Tabelle mit dem Kopf der fünf Zellen (Nr. 3 a bis g).
func leseTabelle(aussen []string) (tabelle, []string) {
	var treffer [][]string
	for i := 0; i < len(aussen); i++ {
		if !strings.HasPrefix(aussen[i], "|") {
			continue
		}
		j := i
		for j < len(aussen) && strings.HasPrefix(aussen[j], "|") {
			j++
		}
		if istKopf(aussen[i]) {
			treffer = append(treffer, aussen[i:j])
		}
		i = j
	}
	switch {
	case len(treffer) == 0:
		return tabelle{}, []string{"Tabelle fehlt"}
	case len(treffer) > 1:
		return tabelle{}, []string{"Tabelle mehrdeutig"}
	}
	t := treffer[0]
	if len(t) < 2 || !istTrennzeile(t[1]) {
		return tabelle{}, []string{"Trennzeile fehlt"}
	}
	if len(t) == 2 {
		return tabelle{}, []string{"Tabelle ohne Zeilen"}
	}
	return leseZeilen(t[2:])
}

// leseZeilen prüft die Datenzeilen (Zellenzahl, Nr., Klasse) und zählt die Klassen.
func leseZeilen(zeilen []string) (tabelle, []string) {
	var t tabelle
	var fehler []string
	gesehen := map[string]bool{}
	for i, l := range zeilen {
		c := zellen(l)
		if len(c) != anzahlZellen {
			fehler = append(fehler, fmt.Sprintf("Zeile %d: %d Zellen statt %d", i+1, len(c), anzahlZellen))
			continue
		}
		nr := strings.TrimSpace(c[0])
		switch {
		case nr == "":
			fehler = append(fehler, "Nr. leer")
		case gesehen[nr]:
			fehler = append(fehler, "Nr. "+nr+" doppelt")
		}
		gesehen[nr] = true
		klasse := strings.TrimSpace(c[4])
		switch klasse {
		case klasseNebenlaeufig:
			t.nebenlaeufig++
		case klasseAndere:
			t.andere++
		default:
			fehler = append(fehler, fmt.Sprintf("Zeile %s: Klasse '%s' unbekannt", nr, klasse))
		}
	}
	return t, fehler
}

// zahl liest eine Zahl nach Nr. 3 j: ein bis neun ASCII-Ziffern ohne führende Null (die 0 gilt).
func zahl(s string) (int64, bool) {
	if s == "" || len(s) > maxStellen || (len(s) > 1 && s[0] == '0') {
		return 0, false
	}
	var n int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int64(s[i]-'0')
	}
	return n, true
}

// basisGueltig ist wahr für sieben bis vierzig Zeichen aus 0-9a-f.
func basisGueltig(s string) bool {
	return len(s) >= 7 && len(s) <= 40 && nurHex(s)
}

func nurHex(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && (s[i] < 'a' || s[i] > 'f') {
			return false
		}
	}
	return true
}

// pfadGueltig prüft einen Pfad nach Nr. 3 m, ohne die Doppelung. Ein führender Schrägstrich ist
// ein leeres erstes Segment.
func pfadGueltig(p string) bool {
	if p == "" || p[0] == ':' {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] < 0x20 || p[i] == 0x7f {
			return false
		}
	}
	n := strings.TrimSuffix(p, "/")
	for _, s := range strings.Split(n, "/") {
		if s == "" || s == "." || s == ".." {
			return false
		}
	}
	for _, v := range []string{"spec", "docs", "harness/vertraege"} {
		if n == v || strings.HasPrefix(n, v+"/") {
			return false
		}
	}
	return true
}

// leseEintrag liest `<N> | <Grund>`, getrennt am ersten Strich.
func leseEintrag(v string) (eintrag, bool) {
	n, grund, ok := strings.Cut(v, "|")
	if !ok {
		return eintrag{}, false
	}
	z, ok := zahl(strings.TrimSpace(n))
	return eintrag{n: z, grund: strings.TrimSpace(grund)}, ok
}

// leseBlock liest die Zeilen eines Blocks (Nr. 3 h bis o).
func leseBlock(zeilen []string) (block, []string) {
	b := block{gesetzt: map[string]bool{}}
	var fehler []string
	for _, roh := range zeilen {
		l := strings.TrimSpace(roh)
		if l == "" {
			continue
		}
		k, v, ok := strings.Cut(l, ":")
		if !ok {
			fehler = append(fehler, "Zeile im Block ohne Schlüssel: "+l)
			continue
		}
		fehler = append(fehler, b.setze(strings.TrimSpace(k), strings.TrimSpace(v))...)
	}
	fehler = append(fehler, b.pruefePfade()...)
	return b, append(fehler, b.fehlende()...)
}

func wiederholbar(k string) bool {
	return k == "pfad" || k == "ueberschreitung"
}

func bekannt(k string) bool {
	switch k {
	case "pfad", "produktcode", "fakes", "basis", "schaetzung", "marke", "schnitt-grund", "ersetzt", "ueberschreitung":
		return true
	}
	return false
}

// setze trägt einen Schlüssel ein und meldet unbekannte, doppelte und ungültige Werte.
func (b *block) setze(k, v string) []string {
	if !bekannt(k) {
		return []string{"Schlüssel " + k + " unbekannt"}
	}
	if b.gesetzt[k] && !wiederholbar(k) {
		return []string{"Schlüssel " + k + " doppelt"}
	}
	b.gesetzt[k] = true
	switch k {
	case "pfad":
		b.pfade = append(b.pfade, v)
	case "basis":
		b.basis = v
		if !basisGueltig(v) {
			return []string{"Basis ungültig: " + v}
		}
	case "schnitt-grund":
		b.schnittGrund = v
	case "ersetzt", "ueberschreitung":
		e, ok := leseEintrag(v)
		if !ok {
			return []string{"Eintrag ungültig: " + k}
		}
		if k == "ersetzt" {
			b.ersetzt = append(b.ersetzt, e)
		} else {
			b.ueberschreitung = append(b.ueberschreitung, e)
		}
	default:
		n, ok := zahl(v)
		if !ok {
			return []string{"Zahl ungültig: " + k}
		}
		b.setzeZahl(k, n)
	}
	return nil
}

func (b *block) setzeZahl(k string, n int64) {
	switch k {
	case "produktcode":
		b.produktcode = n
	case "fakes":
		b.fakes = n
	case "schaetzung":
		b.schaetzung = n
	case "marke":
		b.marke = n
	}
}

// pruefePfade meldet ungültige und doppelte Pfade (Nr. 3 m).
func (b *block) pruefePfade() []string {
	var fehler []string
	gesehen := map[string]bool{}
	for _, p := range b.pfade {
		switch {
		case !pfadGueltig(p):
			fehler = append(fehler, "Pfad ungültig: "+p)
		case gesehen[p]:
			fehler = append(fehler, "Pfad doppelt: "+p)
		}
		gesehen[p] = true
	}
	return fehler
}

// fehlende nennt alle fehlenden Schlüssel: der Teil des Architect vollständig, der Teil des
// Planner ganz oder gar nicht (Nr. 3 k).
func (b *block) fehlende() []string {
	var fehlt []string
	for _, k := range []string{"pfad", "produktcode", "fakes"} {
		if !b.gesetzt[k] {
			fehlt = append(fehlt, k)
		}
	}
	planner := []string{"basis", "schaetzung", "marke"}
	gesetzt := 0
	for _, k := range planner {
		if b.gesetzt[k] {
			gesetzt++
		}
	}
	if gesetzt > 0 {
		for _, k := range planner {
			if !b.gesetzt[k] {
				fehlt = append(fehlt, k)
			}
		}
	}
	if len(fehlt) == 0 {
		return nil
	}
	return []string{"Block unvollständig: es fehlen " + strings.Join(fehlt, ", ")}
}
