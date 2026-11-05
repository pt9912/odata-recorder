package model

import (
	"net/url"
	"strings"
)

const (
	pfadZaehlung = "/$count"
	pfadWert     = "/$value"
)

// TextZiel leitet aus dem OData-Pfad (nach dem Service-Root, mit seinen Escapes, ohne Query) die
// Ziele des Typs Text ab (SPEC-118 Nr. 11 d): zaehlung bei `/$count` mit einem Gegenstand davor,
// wert bei `/<Name>/$value`, dann ist name der einmal %XX-dekodierte Name. Beide sind nie zugleich
// wahr. Die Funktion ist rein und für jeden Text definiert.
func TextZiel(pfad string) (zaehlung, wert bool, name string) {
	if rest, ok := strings.CutSuffix(pfad, pfadZaehlung); ok {
		return rest != "", false, ""
	}
	rest, ok := strings.CutSuffix(pfad, pfadWert)
	if !ok {
		return false, false, ""
	}
	i := strings.LastIndex(rest, "/")
	if i < 0 {
		return false, false, ""
	}
	roh := rest[i+1:]
	if roh == "" || roh[0] == '$' || strings.ContainsAny(roh, "()'") {
		return false, false, ""
	}
	if dekodiert, err := url.PathUnescape(roh); err == nil {
		return false, true, dekodiert
	}
	return false, true, roh
}

// textTyp meldet, ob der Media-Type `text/<x>` mit nichtleerem <x> ist (SPEC-118 Nr. 3); die
// Bytes von <x> werden nicht geprüft.
func textTyp(mt string) bool {
	const text = "text/"
	return strings.HasPrefix(mt, text) && len(mt) > len(text)
}

// zeichensatzErlaubt prüft jeden Parameter charset des ersten Content-Type (SPEC-118 Nr. 11 c):
// nur utf-8 und us-ascii sind erlaubt, ein fehlender Parameter gilt als UTF-8.
func zeichensatzErlaubt(contentType string) bool {
	_, parameter, _ := strings.Cut(contentType, ";")
	for p := range strings.SplitSeq(parameter, ";") {
		n, w, _ := strings.Cut(p, "=")
		if asciiKlein(strings.Trim(n, " \t")) != "charset" {
			continue
		}
		w = strings.Trim(w, " \t")
		if len(w) >= 2 && w[0] == '"' && w[len(w)-1] == '"' {
			w = w[1 : len(w)-1]
		}
		if w = asciiKlein(w); w != "utf-8" && w != "us-ascii" {
			return false
		}
	}
	return true
}

// zaehlungPruefer reicht eine Dezimalzahl byte-gleich durch (SPEC-118 Nr. 11 a): jedes Byte außer
// 0 bis 9 ist syntax. Die leere Eingabe ist vorher entschieden (Nr. 2).
type zaehlungPruefer struct{ aus *bodyAusgabe }

func (z *zaehlungPruefer) schreibe(p []byte) *UnsicherFehler {
	for _, c := range p {
		if c < '0' || c > '9' {
			return unsicher(GrundSyntax, nil)
		}
		z.aus.byte(c)
		if f := z.aus.schreibeVoll(); f != nil {
			return f
		}
	}
	return nil
}

func (*zaehlungPruefer) ende() *UnsicherFehler { return nil }

func (*zaehlungPruefer) status() RedactionStatus { return RedactionNone }

// wertRedaktor prüft den Inhalt eines `$value` als UTF-8 und ersetzt ihn bei einem treffenden
// Namen als Ganzes durch den Marker (SPEC-118 Nr. 11 b). Bei einem Treffer erreicht kein Byte
// des Inhalts den Puffer; der Marker steht erst am Ende des Stroms.
type wertRedaktor struct {
	aus      *bodyAusgabe
	trifft   bool
	rest     uint8 // fehlende Folgebytes einer Mehrbyte-Folge
	lo, hi   byte  // zulässiger Bereich des nächsten Folgebytes
	mPos     int   // so viele Rohbytes des Inhalts gleichen dem Anfang von Marker
	mAbweich bool
}

func neuerWertRedaktor(name string, pfade []JSONPfad, aus *bodyAusgabe) *wertRedaktor {
	w := &wertRedaktor{aus: aus}
	for _, p := range pfade {
		if p.Passt(name) {
			w.trifft = true
			break
		}
	}
	return w
}

func (w *wertRedaktor) schreibe(p []byte) *UnsicherFehler {
	for _, c := range p {
		if !w.utf8Byte(c) {
			return unsicher(GrundUTF8, nil)
		}
		if !w.trifft {
			w.aus.byte(c)
			if f := w.aus.schreibeVoll(); f != nil {
				return f
			}
			continue
		}
		if !w.mAbweich && w.mPos < len(Marker) && c == Marker[w.mPos] {
			w.mPos++
		} else {
			w.mAbweich = true
		}
	}
	return nil
}

// utf8Byte nimmt ein Byte in die Prüfung auf (Regeln von Go: kein überlanger Code, kein Surrogat,
// höchstens U+10FFFF); false ist ein Verstoß.
func (w *wertRedaktor) utf8Byte(c byte) bool {
	if w.rest > 0 {
		if c < w.lo || c > w.hi {
			return false
		}
		w.rest--
		w.lo, w.hi = 0x80, 0xBF
		return true
	}
	if c < 0x80 {
		return true
	}
	var ok bool
	w.rest, w.lo, w.hi, ok = utf8Anfang(c)
	return ok
}

// utf8Anfang liefert für ein Anfangsbyte ab 0x80 die Zahl der Folgebytes und den Bereich des ersten
// von ihnen; ok ist falsch für ein Byte, das keine Folge beginnen darf.
func utf8Anfang(c byte) (rest uint8, lo, hi byte, ok bool) {
	switch {
	case c >= 0xC2 && c <= 0xDF:
		return 1, 0x80, 0xBF, true
	case c == 0xE0:
		return 2, 0xA0, 0xBF, true
	case c == 0xED:
		return 2, 0x80, 0x9F, true
	case c >= 0xE1 && c <= 0xEF:
		return 2, 0x80, 0xBF, true
	case c == 0xF0:
		return 3, 0x90, 0xBF, true
	case c == 0xF4:
		return 3, 0x80, 0x8F, true
	case c >= 0xF1 && c <= 0xF3:
		return 3, 0x80, 0xBF, true
	}
	return 0, 0, 0, false
}

// ende urteilt über das Ende des Stroms: eine halbe Mehrbyte-Folge ist kein gültiges UTF-8; bei
// einem Treffer schreibt erst jetzt der Marker.
func (w *wertRedaktor) ende() *UnsicherFehler {
	if w.rest > 0 {
		return unsicher(GrundUTF8, nil)
	}
	if w.trifft {
		w.aus.text(Marker)
	}
	return nil
}

func (w *wertRedaktor) status() RedactionStatus {
	if w.trifft && (w.mAbweich || w.mPos != len(Marker)) {
		return RedactionApplied
	}
	return RedactionNone
}
