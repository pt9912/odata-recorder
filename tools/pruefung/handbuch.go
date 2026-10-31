package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

const (
	handbuchPfad = "docs/user/benutzerhandbuch.md"
	beispielVerz = "docs/user/beispiele"
	ausgabeKopf  = "pruefung handbuch: "
)

// Nummern des Vertrags (SPEC-200), nach denen die Befunde geordnet werden.
const (
	nrZeichen  = 3
	nrKopf     = 4
	nrZaun     = 5
	nrBeispiel = 6
	nrLauf     = 7
	nrPraefix  = 8
	nrMake     = 9
	nrPlanung  = 10
	nrChronik  = 11
	nrKennung  = 12
)

// befund ist eine Zeile der Ausgabe von `pruefung handbuch`. Die Ausgabe ordnet nach nr, dann
// nach zeile, dann nach name (SPEC-200 Nr. 2 d).
type befund struct {
	nr, zeile int
	name      string
	text      string
}

// pruefeHandbuch prüft das Handbuch unter wurzel gegen das Binary (SPEC-200): Eingabefehler
// (Exit 2) schließen jeden Befund aus und gehen dem ersten Start des Binary voraus.
func pruefeHandbuch(u Umgebung, binary, wurzel string, stdout, stderr io.Writer) int {
	var eingabe []string
	if fi, err := os.Stat(binary); err != nil || !fi.Mode().IsRegular() {
		eingabe = append(eingabe, "Binary fehlt: "+binary)
	}
	zeilen, fehler := liesHandbuch(filepath.Join(wurzel, handbuchPfad))
	if fehler != "" {
		eingabe = append(eingabe, fehler)
	}
	var befunde []befund
	if len(eingabe) == 0 {
		befunde, eingabe = pruefeMitArbeit(u, binary, wurzel, zeilen)
	}
	if len(eingabe) > 0 {
		for _, e := range eingabe {
			_, _ = fmt.Fprintf(stderr, "%sEingabe: %s\n", ausgabeKopf, e)
		}
		return exitAufruf
	}
	return meldeBefunde(befunde, stdout, stderr)
}

// pruefeMitArbeit legt das Arbeitsverzeichnis an und entfernt es am Ende (SPEC-200 Nr. 2 e).
func pruefeMitArbeit(u Umgebung, binary, wurzel string, zeilen []string) ([]befund, []string) {
	bin, err := filepath.Abs(binary)
	if err != nil {
		return nil, []string{"Binary nicht startbar: " + err.Error()}
	}
	arbeit, err := os.MkdirTemp("", "pruefung-handbuch-")
	if err != nil {
		return nil, []string{"Arbeitsverzeichnis nicht anlegbar: " + err.Error()}
	}
	defer func() { _ = os.RemoveAll(arbeit) }()
	return pruefeDokument(u, bin, wurzel, arbeit, zeilen)
}

// pruefeDokument sammelt die Befunde der Nummern 3 bis 12 und die Eingabefehler des Starts.
func pruefeDokument(u Umgebung, bin, wurzel, arbeit string, zeilen []string) ([]befund, []string) {
	d := zerlege(zeilen)
	befunde := zeichenBefunde(zeilen)
	k := leseKopf(d)
	befunde = append(befunde, k.befunde...)
	bindungen, b5 := pruefeZaeune(&d, wurzel)
	befunde = append(befunde, b5...)
	befunde = append(befunde, verwaiste(wurzel, bindungen)...)
	vl, err := fahre(bin, []string{"version"}, arbeit, u.Zeitgrenze)
	if err != nil {
		return nil, []string{eingabeText(err)}
	}
	befunde = append(befunde, versionBefunde(vl, k)...)
	bb, eingabe := pruefeBeispiele(u, bin, wurzel, arbeit, bindungen)
	if len(eingabe) > 0 {
		return nil, eingabe
	}
	befunde = append(befunde, bb...)
	befunde = append(befunde, praefixBefunde(wurzel)...)
	as := absaetze(d)
	befunde = append(befunde, makeBefunde(as, zielliste(wurzel))...)
	befunde = append(befunde, mengenBefunde(as)...)
	return befunde, nil
}

// meldeBefunde schreibt ok oder die geordneten Befunde.
func meldeBefunde(befunde []befund, stdout, stderr io.Writer) int {
	if len(befunde) == 0 {
		_, _ = fmt.Fprintln(stdout, "pruefung handbuch: ok")
		return exitOK
	}
	sort.SliceStable(befunde, func(i, j int) bool {
		a, b := befunde[i], befunde[j]
		if a.nr != b.nr {
			return a.nr < b.nr
		}
		if a.zeile != b.zeile {
			return a.zeile < b.zeile
		}
		return a.name < b.name
	})
	for _, b := range befunde {
		_, _ = fmt.Fprintf(stderr, "%s%s\n", ausgabeKopf, b.text)
	}
	return exitBefund
}

// liesHandbuch liefert die Zeilen des Handbuchs oder den Text des Eingabefehlers (SPEC-200 Nr. 3 a, b).
func liesHandbuch(pfad string) ([]string, string) {
	b, err := os.ReadFile(pfad)
	grund := ""
	switch {
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, syscall.ENOTDIR):
		return nil, "Handbuch fehlt: " + pfad
	case err != nil:
		grund = err.Error()
	case !utf8.Valid(b):
		grund = "kein UTF-8"
	case bytes.IndexByte(b, 0) >= 0:
		grund = "NUL"
	case bytes.HasPrefix(b, []byte("\xEF\xBB\xBF")):
		grund = "BOM"
	}
	if grund != "" {
		return nil, "Handbuch nicht lesbar: " + grund
	}
	return zeilenVon(string(b)), ""
}

// zeilenVon trennt an \n; ein \r unmittelbar davor gehört nicht zur Zeile. Die letzte Zeile ohne
// folgendes \n behält ein \r am Ende (SPEC-200 Nr. 3 b).
func zeilenVon(s string) []string {
	if s == "" {
		return nil
	}
	abgeschlossen := strings.HasSuffix(s, "\n")
	z := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for i := range z {
		if i < len(z)-1 || abgeschlossen {
			z[i] = strings.TrimSuffix(z[i], "\r")
		}
	}
	return z
}

// zeichenBefunde meldet je Zeile und Zeichen einmal ein Zeichen der Kategorien Cc (außer Tabulator),
// Cf, Co, Cs, Mn und Me (SPEC-200 Nr. 3 c).
func zeichenBefunde(zeilen []string) []befund {
	var bf []befund
	for i, z := range zeilen {
		gesehen := map[rune]bool{}
		for _, r := range z {
			if gesehen[r] || r == '\t' || !unicode.In(r, unicode.Cc, unicode.Cf, unicode.Co, unicode.Cs, unicode.Mn, unicode.Me) {
				continue
			}
			gesehen[r] = true
			bf = append(bf, befund{nr: nrZeichen, zeile: i + 1, text: fmt.Sprintf("Zeile %d: Zeichen U+%04X nicht zulässig", i+1, r)})
		}
	}
	return bf
}

// zaun ist ein Zaun des Handbuchs (SPEC-200 Nr. 5 a); offen und schluss sind Zeilenindizes, schluss
// ist -1 ohne Schluss.
type zaun struct {
	offen, schluss int
	info           string
	gebunden       bool
}

// dokument ist das Handbuch als Zeilen mit seinen Zäunen; im[i] ist 1 plus der Index des
// geschlossenen Zaunes, zu dem Zeile i gehört (die Zaunzeilen eingeschlossen), sonst 0.
type dokument struct {
	zeilen []string
	zaeune []zaun
	im     []int
}

// zaunOffen ist wahr für eine Zeile aus genau drei Backticks ohne Einrückung, gefolgt von einer
// Info-Zeichenkette ohne Backtick.
func zaunOffen(z string) bool {
	rest, ok := strings.CutPrefix(z, "```")
	return ok && !strings.Contains(rest, "`")
}

func zerlege(zeilen []string) dokument {
	d := dokument{zeilen: zeilen, im: make([]int, len(zeilen))}
	for i := 0; i < len(zeilen); i++ {
		if !zaunOffen(zeilen[i]) {
			continue
		}
		z := zaun{offen: i, schluss: -1, info: strings.TrimSpace(zeilen[i][3:])}
		for j := i + 1; j < len(zeilen); j++ {
			if strings.TrimRightFunc(zeilen[j], unicode.IsSpace) == "```" {
				z.schluss = j
				break
			}
		}
		d.zaeune = append(d.zaeune, z)
		if z.schluss < 0 {
			break
		}
		for j := i; j <= z.schluss; j++ {
			d.im[j] = len(d.zaeune)
		}
		i = z.schluss
	}
	return d
}

// blockText ist der Inhalt des Zaunes, jede Zeile mit einem \n abgeschlossen.
func (d dokument) blockText(z zaun) string {
	var b strings.Builder
	for _, l := range d.zeilen[z.offen+1 : z.schluss] {
		b.WriteString(l + "\n")
	}
	return b.String()
}

// pruefeZaeune prüft Info-Zeichenkette, Bindungszeile und Block gegen die Datei (SPEC-200 Nr. 5 a bis e)
// und setzt gebunden; es liefert die Namen aller gültigen Bindungen (Pfad P ab beispiele/).
func pruefeZaeune(d *dokument, wurzel string) ([]string, []befund) {
	var bf []befund
	var bindungen []string
	offen := map[int]bool{}
	for _, z := range d.zaeune {
		offen[z.offen] = true
	}
	for i, z := range d.zaeune {
		n := z.offen + 1
		if z.schluss < 0 {
			bf = append(bf, befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: Zaun nicht geschlossen", n)})
			continue
		}
		if z.info != "" {
			bf = append(bf, befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: Info-Zeichenkette %s nicht zulässig", n, z.info)})
		}
		p, b := bindung(*d, z)
		if b != nil {
			bf = append(bf, *b)
			continue
		}
		bindungen = append(bindungen, p)
		if b := vergleicheBlock(*d, z, p, wurzel); b != nil {
			bf = append(bf, *b)
			continue
		}
		d.zaeune[i].gebunden = true
	}
	for i, l := range d.zeilen {
		if d.im[i] == 0 && strings.HasPrefix(l, "Datei: [") && !offen[i+1] {
			bf = append(bf, befund{nr: nrZaun, zeile: i + 1, text: fmt.Sprintf("Zeile %d: Bindungszeile ohne Zaun", i+1)})
		}
	}
	return bindungen, bf
}

// beispielName ist der Name einer Beispiel-Datei: Kleinbuchstaben und Ziffern, durch einzelne
// Striche gegliedert, und .txt.
const beispielName = `[a-z0-9]+(?:-[a-z0-9]+)*\.txt`

// bindung liefert den Pfad P der Bindungszeile unmittelbar vor dem Zaun oder den Befund.
func bindung(d dokument, z zaun) (string, *befund) {
	n := z.offen + 1
	vor := ""
	if z.offen > 0 {
		vor = d.zeilen[z.offen-1]
	}
	if !strings.HasPrefix(vor, "Datei: [") {
		return "", &befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: Beispiel ohne Bindungszeile", n)}
	}
	m := regexpBindung().FindStringSubmatch(vor)
	if m == nil || m[1] != m[2] {
		return "", &befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: Bindung ungültig: %s", n, vor)}
	}
	return m[1], nil
}

// vergleicheBlock hält den Block gleich der Datei P unter docs/user (SPEC-200 Nr. 5 d).
func vergleicheBlock(d dokument, z zaun, p, wurzel string) *befund {
	n := z.offen + 1
	inhalt, err := os.ReadFile(filepath.Join(wurzel, "docs/user", p))
	if err != nil {
		return &befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: %s fehlt", n, p)}
	}
	if d.blockText(z) != string(inhalt) {
		return &befund{nr: nrZaun, zeile: n, text: fmt.Sprintf("Zeile %d: Beispiel weicht von %s ab", n, p)}
	}
	return nil
}

// verwaiste meldet Einträge des Verzeichnisses der Beispiele, die keine Beispiel-Datei sind oder
// keine gültige Bindung tragen (SPEC-200 Nr. 5 f).
func verwaiste(wurzel string, bindungen []string) []befund {
	eintraege, err := os.ReadDir(filepath.Join(wurzel, beispielVerz))
	if err != nil {
		return nil
	}
	genannt := map[string]bool{}
	for _, p := range bindungen {
		genannt[p] = true
	}
	var bf []befund
	for _, e := range eintraege {
		name := e.Name()
		fi, statErr := os.Stat(filepath.Join(wurzel, beispielVerz, name))
		switch {
		case statErr != nil || !fi.Mode().IsRegular() || !regexpName().MatchString(name):
			bf = append(bf, befund{nr: nrZaun, name: name, text: fmt.Sprintf("beispiele: %s keine Beispiel-Datei", name)})
		case !genannt["beispiele/"+name]:
			bf = append(bf, befund{nr: nrZaun, name: name, text: fmt.Sprintf("beispiele: %s ohne Bindung", name)})
		}
	}
	return bf
}

// absatz ist eine Folge nicht leerer Zeilen außerhalb gebundener Zäune oder der Inhalt eines
// gebundenen Zaunes (zitat); zeile ist die erste Zeile, im Zitat die des Zaunes (SPEC-200 Nr. 13 a, c, d).
type absatz struct {
	zeile  int
	zeilen []string
	zitat  bool
}

func absaetze(d dokument) []absatz {
	var out []absatz
	var cur []string
	start := 0
	flush := func() {
		if len(cur) > 0 {
			out = append(out, absatz{zeile: start + 1, zeilen: cur})
			cur = nil
		}
	}
	for i := 0; i < len(d.zeilen); i++ {
		if k := d.im[i]; k > 0 && d.zaeune[k-1].gebunden {
			flush()
			z := d.zaeune[k-1]
			out = append(out, absatz{zeile: z.offen + 1, zeilen: d.zeilen[z.offen+1 : z.schluss], zitat: true})
			i = z.schluss
			continue
		}
		if strings.TrimSpace(d.zeilen[i]) == "" {
			flush()
			continue
		}
		if cur == nil {
			start = i
		}
		cur = append(cur, d.zeilen[i])
	}
	flush()
	return out
}

// praefixBefunde prüft, dass der Bau das Aufrufpräfix der Beispiele deckt (SPEC-200 Nr. 8).
func praefixBefunde(wurzel string) []befund {
	eintraege, err := os.ReadDir(filepath.Join(wurzel, beispielVerz))
	if err != nil || len(eintraege) == 0 {
		return nil
	}
	var bf []befund
	if !enthaeltEine(wurzel, makeDateien(wurzel, false), "-t odata-recorder:build") {
		bf = append(bf, befund{nr: nrPraefix, text: "Aufrufpräfix nicht gedeckt: Image"})
	}
	if !enthaeltEine(wurzel, []string{"Dockerfile"}, "-o /out/odata-recorder") {
		bf = append(bf, befund{nr: nrPraefix, text: "Aufrufpräfix nicht gedeckt: Pfad"})
	}
	return bf
}

// makeDateien nennt die Dateien, aus denen make Ziele liest: Makefile und harness/mk/*.mk
// (nicht rekursiv), mit allZiele zusätzlich d-check.mk und repo.mk.
func makeDateien(wurzel string, allZiele bool) []string {
	dateien := []string{"Makefile"}
	eintraege, _ := os.ReadDir(filepath.Join(wurzel, "harness/mk"))
	for _, e := range eintraege {
		if fi, err := os.Stat(filepath.Join(wurzel, "harness/mk", e.Name())); err == nil && fi.Mode().IsRegular() && strings.HasSuffix(e.Name(), ".mk") {
			dateien = append(dateien, path.Join("harness/mk", e.Name()))
		}
	}
	if allZiele {
		dateien = append(dateien, "d-check.mk", "repo.mk")
	}
	return dateien
}

// lies liefert den Inhalt von datei unter wurzel, leer ohne Datei.
func lies(wurzel, datei string) string {
	b, err := os.ReadFile(filepath.Join(wurzel, datei))
	if err != nil {
		return ""
	}
	return string(b)
}

func enthaeltEine(wurzel string, dateien []string, text string) bool {
	for _, d := range dateien {
		if b, err := os.ReadFile(filepath.Join(wurzel, d)); err == nil && bytes.Contains(b, []byte(text)) {
			return true
		}
	}
	return false
}
