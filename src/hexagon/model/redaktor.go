package model

import (
	"errors"
	"io"
	"math"
	"strings"
)

// ErrUnsicher ist die Wurzel jedes Fehlers von RedigiereBody, der einen Body als nicht sicher
// redigierbar meldet (SPEC-118 Nr. 7). Ein Fehler, der mit ErrUngueltig endet, ist ein Fehler
// des Aufrufs und kein Befund über den Body.
var ErrUnsicher = errors.New("model: Body nicht sicher redigierbar")

// UnsicherGrund nennt, warum ein Body nicht sicher redigiert werden kann (SPEC-118 Nr. 7).
type UnsicherGrund string

// Die Gründe aus SPEC-118 Nr. 7.
const (
	GrundTyp            UnsicherGrund = "typ"
	GrundMismatch       UnsicherGrund = "mismatch"
	GrundSyntax         UnsicherGrund = "syntax"
	GrundUTF8           UnsicherGrund = "utf8"
	GrundTiefe          UnsicherGrund = "tiefe"
	GrundNichtstring    UnsicherGrund = "nichtstring"
	GrundZuGross        UnsicherGrund = "zu_gross"
	GrundUnvollstaendig UnsicherGrund = "unvollstaendig"
	GrundEingabe        UnsicherGrund = "eingabe"
	GrundAusgabe        UnsicherGrund = "ausgabe"
)

// UnsicherFehler ist der Fehler von RedigiereBody bei Status unsafe: ODR-E2401 mit Grund. Sein
// Text nennt den Grund und kein Byte der Eingabe; die Ursache eines Fehlers von src oder dst
// bleibt über errors.Is erreichbar.
type UnsicherFehler struct {
	Grund   UnsicherGrund
	ursache error
}

func unsicher(g UnsicherGrund, ursache error) *UnsicherFehler {
	return &UnsicherFehler{Grund: g, ursache: ursache}
}

// Error nennt den Code und den Grund.
func (e *UnsicherFehler) Error() string {
	return "ODR-E2401: Body nicht sicher redigierbar: " + string(e.Grund)
}

// Code liefert ODR-E2401 (SPEC-017).
func (*UnsicherFehler) Code() Code { return CodeE2401 }

// Unwrap macht ErrUnsicher und die Ursache eines Fehlers von src oder dst erreichbar.
func (e *UnsicherFehler) Unwrap() []error {
	if e.ursache == nil {
		return []error{ErrUnsicher}
	}
	return []error{ErrUnsicher, e.ursache}
}

// BodyEingang beschreibt den Body, den RedigiereBody liest (SPEC-118): ContentType sind alle
// Werte des Headers Content-Type, Metadata ist wahr genau für die Antwort auf `$metadata`.
type BodyEingang struct {
	ContentType []string
	Metadata    bool
}

// BodyRegeln sind die wirksamen Pfadmuster und die Grenze MAX_CAPTURE_BODY (SPEC-118).
type BodyRegeln struct {
	Pfade    []JSONPfad
	MaxBytes int64
}

// Größe der festen Puffer je Richtung (SPEC-118 Nr. 6): sie hängen nicht vom Strom ab. Der
// Eingabepuffer ist bodyEinPuffer groß und jeder Read bietet höchstens so viel an; der
// Ausgabepuffer wird geleert, sobald er bodyAusSchwelle erreicht, und jeder Write an dst ist
// höchstens bodyAusSchwelle + bodyAusReserve lang. TestBodyPuffer hält beide Größen für JSON.
const (
	bodyEinPuffer   = 32 << 10
	bodyAusSchwelle = 32 << 10
	// bodyAusReserve ist die längste Ausgabe eines einzelnen Eingabebytes (der Marker samt
	// Anführungszeichen); der Ausgabepuffer hat diesen Spielraum über der Schwelle, weil er erst
	// nach dem Byte geleert wird, das sie erreicht.
	bodyAusReserve = len(Marker) + 2
)

// RedigiereBody liest einen Body aus src, schreibt den redigierten Body nach dst und liefert
// den Status der Redaction (SPEC-118). Jeder Verstoß ist unsafe mit einem *UnsicherFehler; der
// Aufrufer verwirft dann dst (Nr. 9).
func RedigiereBody(dst io.Writer, src io.Reader, ein BodyEingang, regeln BodyRegeln) (RedactionStatus, error) {
	if err := pruefeBodyArgumente(dst, src, regeln); err != nil {
		return "", err
	}
	l := &bodyLauf{
		src:    src,
		max:    regeln.MaxBytes,
		aus:    &bodyAusgabe{dst: dst, puffer: make([]byte, 0, bodyAusSchwelle+bodyAusReserve)},
		puffer: make([]byte, bodyEinPuffer),
	}
	return l.lauf(ein, regeln.Pfade)
}

func pruefeBodyArgumente(dst io.Writer, src io.Reader, regeln BodyRegeln) error {
	switch {
	case regeln.MaxBytes < 1:
		return ungueltig("MaxBytes ist kleiner als 1")
	case dst == nil || src == nil:
		return ungueltig("dst oder src fehlt")
	}
	for _, p := range regeln.Pfade {
		if p == (JSONPfad{}) {
			return ungueltig("Pfadmuster ist der Nullwert")
		}
	}
	return nil
}

// bodyVerarbeiter ist der Zustandsautomat über den Bytes eines Bodys.
type bodyVerarbeiter interface {
	schreibe(p []byte) *UnsicherFehler
	ende() *UnsicherFehler
	status() RedactionStatus
}

// bodyAusgabe puffert die Ausgabe und schreibt sie in dst.
type bodyAusgabe struct {
	dst    io.Writer
	puffer []byte
}

func (a *bodyAusgabe) byte(c byte)   { a.puffer = append(a.puffer, c) }
func (a *bodyAusgabe) text(s string) { a.puffer = append(a.puffer, s...) }
func (a *bodyAusgabe) voll() bool    { return len(a.puffer) >= bodyAusSchwelle }

// leere schreibt den Puffer nach dst; ein kurzes Schreiben ohne Fehler ist io.ErrShortWrite.
func (a *bodyAusgabe) leere() error {
	if len(a.puffer) == 0 {
		return nil
	}
	n, err := a.dst.Write(a.puffer)
	if err == nil && n < len(a.puffer) {
		err = io.ErrShortWrite
	}
	a.puffer = a.puffer[:0]
	return err
}

// schreibeVoll leert den Puffer, wenn er die Schwelle erreicht hat, und meldet einen Fehler von dst.
func (a *bodyAusgabe) schreibeVoll() *UnsicherFehler {
	if !a.voll() {
		return nil
	}
	if err := a.leere(); err != nil {
		return unsicher(GrundAusgabe, err)
	}
	return nil
}

type bodyLauf struct {
	src     io.Reader
	max     int64
	gelesen int64
	aus     *bodyAusgabe
	puffer  []byte
	v       bodyVerarbeiter
}

// leseDaten liest bis zu einem Read mit Daten oder Fehler; leere Reads ändern nichts. Es liest
// nie über MaxBytes + 1 Bytes hinaus (bei MaxBytes = math.MaxInt64 nie über diesen Wert).
func (l *bodyLauf) leseDaten() (int, error) {
	for {
		p := l.puffer
		rest := l.max - l.gelesen
		if rest < math.MaxInt64 {
			rest++
		}
		if int64(len(p)) > rest {
			p = p[:rest]
		}
		n, err := l.src.Read(p)
		l.gelesen += int64(n)
		if n > 0 || err != nil {
			return n, err
		}
	}
}

func (l *bodyLauf) lauf(ein BodyEingang, pfade []JSONPfad) (RedactionStatus, error) {
	n, rerr := l.leseDaten()
	if n == 0 {
		if rerr == io.EOF {
			return RedactionNone, nil
		}
		return RedactionUnsafe, unsicher(GrundEingabe, rerr)
	}
	if f := l.waehleVerarbeiter(ein, pfade); f != nil {
		return RedactionUnsafe, f
	}
	var verarbeitet int64
	for {
		chunk := l.puffer[:n]
		zuGross := verarbeitet+int64(n) > l.max
		if zuGross {
			chunk = chunk[:l.max-verarbeitet]
		}
		verarbeitet += int64(len(chunk))
		if f := l.v.schreibe(chunk); f != nil {
			return l.scheitere(f)
		}
		if zuGross {
			return l.scheitere(unsicher(GrundZuGross, nil))
		}
		if err := l.aus.leere(); err != nil {
			return RedactionUnsafe, unsicher(GrundAusgabe, err)
		}
		switch {
		case rerr == io.EOF:
			return l.beende()
		case rerr != nil:
			return l.scheitere(unsicher(GrundEingabe, rerr))
		}
		n, rerr = l.leseDaten()
	}
}

func (l *bodyLauf) beende() (RedactionStatus, error) {
	if f := l.v.ende(); f != nil {
		return l.scheitere(f)
	}
	if err := l.aus.leere(); err != nil {
		return RedactionUnsafe, unsicher(GrundAusgabe, err)
	}
	return l.v.status(), nil
}

// scheitere gibt den bis zum Verstoß redigierten Teil an dst, außer dst selbst ist der Verstoß.
// Das Präfix ist kein Body (SPEC-118 Nr. 9).
func (l *bodyLauf) scheitere(f *UnsicherFehler) (RedactionStatus, error) {
	if f.Grund != GrundAusgabe {
		_ = l.aus.leere()
	}
	return RedactionUnsafe, f
}

// waehleVerarbeiter wählt nach dem Typ (SPEC-118 Nr. 3) den Automaten.
func (l *bodyLauf) waehleVerarbeiter(ein BodyEingang, pfade []JSONPfad) *UnsicherFehler {
	switch klassifiziereBody(ein) {
	case bodyJSON:
		l.v = neuerJSONRedaktor(pfade, l.aus)
	case bodyEDMX:
		l.v = &edmxPruefer{aus: l.aus}
	default:
		return unsicher(GrundTyp, nil)
	}
	return nil
}

type bodyTyp uint8

const (
	bodyUnbekannt bodyTyp = iota
	bodyJSON
	bodyEDMX
)

// klassifiziereBody liest genau einen Content-Type: den Media-Type bis zum ersten `;`, ohne
// Leerzeichen und Tabulatoren am Rand und ohne Beachtung der ASCII-Schreibweise.
func klassifiziereBody(ein BodyEingang) bodyTyp {
	if len(ein.ContentType) != 1 {
		return bodyUnbekannt
	}
	mt, _, _ := strings.Cut(ein.ContentType[0], ";")
	mt = asciiKlein(strings.Trim(mt, " \t"))
	const anwendung, json = "application/", "+json"
	switch {
	case mt == "application/json":
		return bodyJSON
	case strings.HasPrefix(mt, anwendung) && strings.HasSuffix(mt, json) && len(mt) > len(anwendung)+len(json):
		return bodyJSON
	case mt == "application/xml" && ein.Metadata:
		return bodyEDMX
	}
	return bodyUnbekannt
}

func istLeerraum(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// edmxPruefer reicht EDMX byte-gleich durch und verlangt `<` als erstes Zeichen nach einem
// einmaligen BOM am Anfang und Leerraum (SPEC-118 Nr. 3).
type edmxPruefer struct {
	aus *bodyAusgabe
	z   edmxZustand
}

type edmxZustand uint8

const (
	edmxStart edmxZustand = iota
	edmxBom1
	edmxBom2
	edmxVor
	edmxFrei
)

func (e *edmxPruefer) schreibe(p []byte) *UnsicherFehler {
	for _, c := range p {
		if e.z != edmxFrei && !e.vorspann(c) {
			return unsicher(GrundMismatch, nil)
		}
		e.aus.byte(c)
		if f := e.aus.schreibeVoll(); f != nil {
			return f
		}
	}
	return nil
}

// vorspann prüft ein Byte vor dem ersten `<`; false ist ein Verstoß.
func (e *edmxPruefer) vorspann(c byte) bool {
	switch {
	case e.z == edmxBom1:
		e.z = edmxBom2
		return c == 0xBB
	case e.z == edmxBom2:
		e.z = edmxVor
		return c == 0xBF
	case e.z == edmxStart && c == 0xEF:
		e.z = edmxBom1
		return true
	case c == '<':
		e.z = edmxFrei
		return true
	}
	e.z = edmxVor
	return istLeerraum(c)
}

func (e *edmxPruefer) ende() *UnsicherFehler {
	if e.z != edmxFrei {
		return unsicher(GrundMismatch, nil)
	}
	return nil
}

func (*edmxPruefer) status() RedactionStatus { return RedactionNone }
