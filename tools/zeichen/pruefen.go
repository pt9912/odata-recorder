package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	// binGrenze ist die Zahl der Bytes, in denen ein NUL eine Datei binär macht (SPEC-233 Nr. 6 a).
	binGrenze = 8000
	// ausgenommen ist der Teilbaum, den das Programm nicht betritt (SPEC-233 Nr. 5 a).
	ausgenommen = ".harness/baseline"
	// lesepuffer ist die Größe des Lesepuffers; eine Zeile darf länger sein.
	lesepuffer = 4096
)

type art int

const (
	artDatei art = iota
	artAnderes
	artVerzeichnisfehler
)

// eintrag ist eine Stelle des Baums in der Folge der Pfade.
type eintrag struct {
	pfad string
	art  art
	err  error
}

// pruefer prüft einen Baum und zählt (SPEC-233 Nr. 10 c, d); readDir und open sind die beiden
// Zugriffe auf das Dateisystem.
type pruefer struct {
	readDir func(string) ([]os.DirEntry, error)
	open    func(string) (io.ReadCloser, error)
	out     io.Writer
	err     io.Writer

	gelesen, binaer, andere int
	befunde, dateien        int
	eingabefehler           bool
}

func neuerPruefer(out, errw io.Writer) *pruefer {
	return &pruefer{
		readDir: os.ReadDir,
		open:    func(p string) (io.ReadCloser, error) { return os.Open(p) },
		out:     out,
		err:     errw,
	}
}

// eingabe meldet einen Eingabefehler (SPEC-233 Nr. 10 e).
func (p *pruefer) eingabe(text string) {
	p.eingabefehler = true
	_, _ = fmt.Fprintf(p.err, "zeichen: Eingabe: %s\n", text)
}

// grund ist der Text eines Fehlers ohne den Pfad, maskiert.
func grund(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		err = pe.Err
	}
	return maskiere(err.Error())
}

func (p *pruefer) nichtLesbar(pfad string, err error) {
	p.eingabe(maskiere(pfad) + ": nicht lesbar: " + grund(err))
}

// pruefen liest alle Dateien unter wurzel und liefert den Exit-Code (SPEC-233 Nr. 11).
func (p *pruefer) pruefen(wurzel string) int {
	if st, err := os.Stat(wurzel); err != nil || !st.IsDir() {
		p.eingabe("Wurzel fehlt oder ist kein Verzeichnis: " + maskiere(wurzel))
		return 2
	}
	var liste []eintrag
	p.sammle(wurzel, "", &liste)
	slices.SortFunc(liste, func(a, b eintrag) int { return strings.Compare(a.pfad, b.pfad) })
	for _, e := range liste {
		switch e.art {
		case artAnderes:
			p.andere++
		case artVerzeichnisfehler:
			p.nichtLesbar(e.pfad, e.err)
		case artDatei:
			p.datei(wurzel, e.pfad)
		}
	}
	if p.gelesen == 0 {
		p.eingabe("keine Textdatei geprüft")
	}
	if p.befunde > 0 {
		_, _ = fmt.Fprintf(p.err, "zeichen: %d Befund(e) in %d Datei(en)\n", p.befunde, p.dateien)
	}
	switch {
	case p.eingabefehler:
		return 2
	case p.befunde > 0:
		return 1
	}
	_, _ = fmt.Fprintf(p.out, "zeichen: ok, %d Datei(en) geprüft, %d Binärdatei(en) und %d andere Einträge übergangen\n",
		p.gelesen, p.binaer, p.andere)
	return 0
}

// sammle trägt die Einträge unter wurzel/rel ein (SPEC-233 Nr. 5 a bis c).
func (p *pruefer) sammle(wurzel, rel string, liste *[]eintrag) {
	dir := wurzel
	if rel != "" {
		dir = filepath.Join(wurzel, filepath.FromSlash(rel))
	}
	es, err := p.readDir(dir)
	if err != nil {
		pfad := rel
		if pfad == "" {
			pfad = "."
		}
		*liste = append(*liste, eintrag{pfad: pfad, art: artVerzeichnisfehler, err: err})
		return
	}
	for _, e := range es {
		pfad := e.Name()
		if rel != "" {
			pfad = rel + "/" + pfad
		}
		switch t := e.Type(); {
		case t.IsDir() && pfad == ausgenommen:
		case t.IsDir():
			p.sammle(wurzel, pfad, liste)
		case t.IsRegular():
			*liste = append(*liste, eintrag{pfad: pfad, art: artDatei})
		default:
			*liste = append(*liste, eintrag{pfad: pfad, art: artAnderes})
		}
	}
}

// fehlerLeser liefert immer denselben Fehler.
type fehlerLeser struct{ err error }

func (f fehlerLeser) Read([]byte) (int, error) { return 0, f.err }

func utf16Bom(kopf []byte) bool {
	return len(kopf) >= 2 && (kopf[0] == 0xFF && kopf[1] == 0xFE || kopf[0] == 0xFE && kopf[1] == 0xFF)
}

// datei prüft eine reguläre Datei: UTF-16-BOM, binär oder Text (SPEC-233 Nr. 6).
func (p *pruefer) datei(wurzel, rel string) {
	f, err := p.open(filepath.Join(wurzel, filepath.FromSlash(rel)))
	if err != nil {
		p.nichtLesbar(rel, err)
		return
	}
	defer func() { _ = f.Close() }()
	kopf := make([]byte, binGrenze)
	n, herr := io.ReadFull(f, kopf)
	kopf = kopf[:n]
	var rest io.Reader = fehlerLeser{herr}
	if herr == nil {
		rest = f
	} else if errors.Is(herr, io.EOF) || errors.Is(herr, io.ErrUnexpectedEOF) {
		rest = bytes.NewReader(nil)
	}
	switch {
	case utf16Bom(kopf):
		p.gelesen++
		p.befund(rel, fmt.Sprintf("1:1: UTF-16-BOM (Bytes 0x%02X 0x%02X)", kopf[0], kopf[1]))
		p.dateien++
	case bytes.IndexByte(kopf, 0) >= 0:
		p.binaer++
	default:
		p.gelesen++
		if p.scan(rel, io.MultiReader(bytes.NewReader(kopf), rest)) > 0 {
			p.dateien++
		}
	}
}

// befund schreibt eine Fundstelle: Pfad maskiert, dann der Text mit Ort.
func (p *pruefer) befund(rel, text string) {
	p.befunde++
	_, _ = fmt.Fprintf(p.err, "zeichen: %s:%s\n", maskiere(rel), text)
}

// scan liest r strömend und meldet jedes Zeichen von Z und jedes ungültige Byte (SPEC-233 Nr. 7 bis 9);
// es liefert die Zahl der Fundstellen. Ein Lesefehler ist ein Eingabefehler, die Fundstellen davor bleiben.
func (p *pruefer) scan(rel string, r io.Reader) int {
	br := bufio.NewReaderSize(r, lesepuffer)
	zeile, spalte, zahl := 1, 1, 0
	for {
		b, err := br.Peek(1)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.nichtLesbar(rel, err)
			}
			return zahl
		}
		erstes := b[0]
		c, groesse, _ := br.ReadRune()
		switch kat := kategorie(c); {
		case c == utf8.RuneError && groesse == 1:
			p.befund(rel, fmt.Sprintf("%d:%d: ungültiges UTF-8 (Byte 0x%02X)", zeile, spalte, erstes))
			zahl++
		case kat != "":
			p.befund(rel, fmt.Sprintf("%d:%d: %s", zeile, spalte, befundText(c, kat)))
			zahl++
		}
		if c == '\n' {
			zeile, spalte = zeile+1, 1
		} else {
			spalte++
		}
	}
}
