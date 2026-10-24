package csdl

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

var errKodierung = errors.New("csdl: Kodierung nicht UTF-8")

func fehler(u model.MetadatenUrsache, ort string) error {
	return &model.MetadatenFehler{Ursache: u, Ort: ort}
}

func struktur(ort string) error { return fehler(model.UrsacheStruktur, ort) }

// lauf ist der Zustand eines Aufrufs; der Leser selbst hat keinen.
type lauf struct {
	dec      *xml.Decoder
	st       *strom
	stack    []rahmen
	tiefe    int
	gesehen  bool
	fertig   bool
	dataServ int
	d        dokument
	slots    []slot
}

func neuerLauf(ctx context.Context, src io.Reader, grenze int64) *lauf {
	st := &strom{kontext: ctx.Err, src: src, max: grenze}
	dec := xml.NewDecoder(st)
	dec.CharsetReader = func(string, io.Reader) (io.Reader, error) { return nil, errKodierung }
	return &lauf{dec: dec, st: st, d: neuesDokument()}
}

// lies liest das Dokument bis zum Ende des Stroms (Nr. 2 bis 8).
func (l *lauf) lies() error {
	for {
		von := l.dec.InputOffset()
		tok, err := l.dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tokenFehler(err)
		}
		if err := l.token(tok, von, l.dec.InputOffset()); err != nil {
			return err
		}
	}
	if !l.gesehen {
		return fehler(model.UrsacheSyntax, "")
	}
	return nil
}

func tokenFehler(err error) error {
	var sf *stromFehler
	switch {
	case errors.As(err, &sf):
		return stromZuFehler(sf)
	case errors.Is(err, io.ErrNoProgress):
		return err
	case errors.Is(err, errKodierung):
		return fehler(model.UrsacheKodierung, "")
	}
	return fehler(model.UrsacheSyntax, "")
}

func stromZuFehler(sf *stromFehler) error {
	switch sf.art {
	case stromKodierung:
		return fehler(model.UrsacheKodierung, "")
	case stromZuGross:
		return fehler(model.UrsacheZuGross, "")
	case stromSyntax:
		return fehler(model.UrsacheSyntax, "")
	}
	return sf.err
}

func (l *lauf) token(tok xml.Token, von, bis int64) error {
	switch t := tok.(type) {
	case xml.StartElement:
		if err := l.refs(von, bis, t); err != nil {
			return err
		}
		return l.start(t)
	case xml.EndElement:
		return l.ende()
	case xml.CharData:
		if len(l.stack) == 0 && len(bytes.Trim(t, " \t\r\n")) > 0 {
			return fehler(model.UrsacheSyntax, "")
		}
		return l.refs(von, bis, t)
	case xml.ProcInst:
		if t.Target == "xml" && von != 0 {
			return fehler(model.UrsacheSyntax, "")
		}
	case xml.Directive:
		if bytes.HasPrefix(t, []byte("DOCTYPE")) {
			return fehler(model.UrsacheDoctype, "")
		}
		return fehler(model.UrsacheSyntax, "")
	}
	return nil
}

// refs prüft die Zeichenreferenzen auf Ersatzzeichen (Nr. 4 b); der Decoder macht daraus U+FFFD,
// also sucht es nur dort in den Bytes des Tokens.
func (l *lauf) refs(von, bis int64, tok any) error {
	var ersetzt bool
	switch t := tok.(type) {
	case xml.CharData:
		ersetzt = bytes.Contains(t, []byte("�"))
	case xml.StartElement:
		for _, a := range t.Attr {
			ersetzt = ersetzt || strings.Contains(a.Value, "�")
		}
	}
	if !ersetzt {
		return nil
	}
	roh := l.st.alles[von:bis]
	if bytes.HasPrefix(roh, []byte("<![CDATA[")) || !hatErsatzRef(roh) {
		return nil
	}
	return fehler(model.UrsacheSyntax, "")
}

func hatErsatzRef(roh []byte) bool {
	for {
		i := bytes.Index(roh, []byte("&#"))
		if i < 0 {
			return false
		}
		zahl, rest, _ := bytes.Cut(roh[i+2:], []byte(";"))
		roh = rest
		z, basis := string(zahl), 10
		if strings.HasPrefix(z, "x") {
			z, basis = z[1:], 16
		}
		if n, err := strconv.ParseUint(z, basis, 32); err == nil && n >= 0xD800 && n <= 0xDFFF {
			return true
		}
	}
}

func (l *lauf) start(se xml.StartElement) error {
	if l.fertig {
		return fehler(model.UrsacheSyntax, "")
	}
	l.tiefe++
	if l.tiefe > maxTiefe {
		return fehler(model.UrsacheTiefe, "")
	}
	if doppelt(se.Attr) {
		return fehler(model.UrsacheSyntax, "")
	}
	if len(l.stack) == 0 {
		return l.wurzel(se)
	}
	p := l.stack[len(l.stack)-1]
	if p.art == kOpak {
		l.push(rahmen{art: kOpak})
		return nil
	}
	return l.kind(p, se)
}

func doppelt(attr []xml.Attr) bool {
	gesehen := make(map[xml.Name]bool, len(attr))
	for _, a := range attr {
		if gesehen[a.Name] {
			return true
		}
		gesehen[a.Name] = true
	}
	return false
}

func (l *lauf) push(r rahmen) { l.stack = append(l.stack, r) }

func (l *lauf) ende() error {
	r := l.stack[len(l.stack)-1]
	l.stack = l.stack[:len(l.stack)-1]
	l.tiefe--
	if r.art == kEdmx {
		l.fertig = true
		if l.dataServ == 0 {
			return fehler(model.UrsacheWurzel, "")
		}
	}
	return nil
}

// wurzel prüft Nr. 6 a und b am Start-Tag der Wurzel.
func (l *lauf) wurzel(se xml.StartElement) error {
	switch {
	case se.Name.Local == "Edmx" && se.Name.Space == nsEdmxV:
		return fehler(model.UrsacheVersion, "")
	case se.Name.Local != "Edmx" || se.Name.Space != nsEdmx:
		return fehler(model.UrsacheWurzel, "")
	}
	if v, ok := attr(se, "Version"); !ok || v != "4.0" {
		return fehler(model.UrsacheVersion, "")
	}
	l.gesehen = true
	l.push(rahmen{art: kEdmx})
	l.attributFunde(se, "", nil)
	return nil
}

// attr liefert den Wert des Attributs ohne Namensraum.
func attr(se xml.StartElement, name string) (string, bool) {
	for _, a := range se.Attr {
		if a.Name.Space == "" && a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}
