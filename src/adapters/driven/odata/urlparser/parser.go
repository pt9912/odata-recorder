// Package urlparser liest Pfad und Query einer OData-V4-URL in den Query-AST des Modells (SPEC-160 bis SPEC-162). Er ist der
// Driven Adapter des Ports outbound.URLCodec, kennt das Modell des Dienstes nicht und nutzt allein die Standardbibliothek.
package urlparser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Grenzen des Lesers (SPEC-160 Nr. 6); jede Zahl ist erlaubt, die nächste der Verstoß.
const (
	maxPfad          = 2048
	maxQuery         = 8192
	maxPfadSegmente  = 16
	maxQuerySegmente = 64
	maxTeile         = 16
	maxLiteral       = 512
	maxName          = 128
	maxSelect        = 256
	maxOrderBy       = 32
)

// FilterLeser liest den dekodierten Wert von `$filter` (SPEC-160 Nr. 8). Er muss zustandslos und nebenläufig aufrufbar sein.
type FilterLeser interface {
	LeseFilter(text string) (model.FilterAusdruck, error)
}

// Option konfiguriert New.
type Option func(*Parser)

// WithFilter hängt einen FilterLeser ein; die letzte Option gilt, nil ist kein Leser.
func WithFilter(l FilterLeser) Option {
	return func(p *Parser) { p.filter = l }
}

// Parser setzt outbound.URLCodec um. Nach New hat er keinen veränderlichen Zustand (SPEC-160 Nr. 7).
type Parser struct {
	filter FilterLeser
}

// New bildet einen Parser; ohne WithFilter ist `$filter` die Ursache `filter`.
func New(opts ...Option) *Parser {
	p := &Parser{}
	for _, o := range opts {
		o(p)
	}
	return p
}

// lauf ist der Zustand eines Aufrufs: der erste Verstoß der Klasse `nicht_unterstuetzt`, der das Lesen nicht beendet.
type lauf struct {
	p         *Parser
	erster501 *model.URLFehler
}

func fehler(u model.URLUrsache, ort string) *model.URLFehler {
	return &model.URLFehler{Ursache: u, Ort: ort}
}

func (l *lauf) nichtUnterstuetzt(u model.URLUrsache, ort string) {
	if l.erster501 == nil {
		l.erster501 = fehler(u, ort)
	}
}

func ortNr(praefix string, n int) string { return praefix + "/" + strconv.Itoa(n) }

// ParseURL liest pfad und rawQuery (SPEC-160 Nr. 1 bis 6): zuerst die Form der Argumente, dann die Gesamtform, dann der
// Pfad vor der Query; ein Verstoß der Klasse `ungueltig` beendet das Lesen, einer der Klasse `nicht_unterstuetzt` nicht.
func (p *Parser) ParseURL(pfad, rawQuery string) (model.ODataURL, error) {
	if err := pruefeArgumente(pfad, rawQuery); err != nil {
		return model.ODataURL{}, err
	}
	segs, f := gesamtform(pfad, rawQuery)
	if f != nil {
		return model.ODataURL{}, f
	}
	l := &lauf{p: p}
	var u model.ODataURL
	if err := l.lesePfad(pfad, &u.Ressource); err != nil {
		return model.ODataURL{}, err
	}
	if err := l.leseQuery(segs, &u.Optionen); err != nil {
		return model.ODataURL{}, err
	}
	if l.erster501 != nil {
		return model.ODataURL{}, l.erster501
	}
	return u, nil
}

// pruefeArgumente prüft die Form der Argumente (SPEC-160 Nr. 1); ein Verstoß ist ein Programmfehler des Aufrufers.
func pruefeArgumente(pfad, rawQuery string) error {
	if pfad == "" || pfad[0] != '/' {
		return fmt.Errorf("%w: urlparser: pfad ohne führenden Schrägstrich", model.ErrUngueltig)
	}
	for i := range len(pfad) {
		if c := pfad[i]; steuerzeichen(c) || c == '?' || c == '#' {
			return fmt.Errorf("%w: urlparser: pfad mit rohem Steuerzeichen, ? oder #", model.ErrUngueltig)
		}
	}
	for i := range len(rawQuery) {
		if steuerzeichen(rawQuery[i]) {
			return fmt.Errorf("%w: urlparser: rawQuery mit rohem Steuerzeichen", model.ErrUngueltig)
		}
	}
	return nil
}

func steuerzeichen(c byte) bool { return c < 0x20 || c == 0x7f }

// gesamtform prüft Stufe 0 der Rangfolge (SPEC-160 Nr. 5) und liefert die Segmente der Query.
func gesamtform(pfad, rawQuery string) ([]model.QuerySegment, *model.URLFehler) {
	switch {
	case len(pfad) > maxPfad:
		return nil, fehler(model.URLZuLang, "pfad")
	case len(rawQuery) > maxQuery:
		return nil, fehler(model.URLZuLang, "query")
	case strings.Count(pfad, "/") > maxPfadSegmente:
		return nil, fehler(model.URLZuViele, "pfad")
	}
	segs := model.QuerySegmente(rawQuery)
	if len(segs) > maxQuerySegmente {
		return nil, fehler(model.URLZuViele, "query")
	}
	return segs, nil
}
