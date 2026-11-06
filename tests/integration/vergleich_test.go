package integration_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Der Vergleich der Protokolltreue (LH-QA-07.a): Status, Body nach dem Entpacken und die relevanten Header
// zweier Antworten. Er liest nur, was ein Client sah; die Randformen stehen in LH-QA-07.a Nr. 1 bis 7.

// rohantwort ist eine vollständige Antwort, wie der Client sie las: Statuscode, Reason-Phrase, die Feldzeilen
// je Name (die Schreibweise des Namens gilt, wie die Quelle sie liefert) und die Rohbytes des Bodys.
type rohantwort struct {
	status int
	phrase string
	header map[string][]string
	body   []byte
}

// abweichung ist ein Unterschied: art ist Status, Header oder Body, name der kleingeschriebene Name eines Headers.
type abweichung struct{ art, name string }

// vergleichsregeln sind die Eingaben des Vergleichs: die wirksame Menge von VOLATILE_HEADERS und der Modus
// für ein Coding, das der Client nicht nannte (LH-QA-07.a Nr. 4): die Rohbytes und Content-Encoding zählen dann.
type vergleichsregeln struct {
	volatile []string
	roh      bool
}

// asciiKlein schreibt A bis Z klein, jedes andere Byte bleibt (RFC 9110 §5.1: Namen ohne Beachtung der ASCII-Schreibweise).
func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

// nachName ordnet die Feldzeilen nach dem kleingeschriebenen Namen; Zeilen verschiedener Schreibweisen
// desselben Namens folgen der Bytefolge der Schreibweisen, dann ihrer Reihenfolge.
func nachName(h map[string][]string) map[string][]string {
	namen := make([]string, 0, len(h))
	for n := range h {
		namen = append(namen, n)
	}
	slices.Sort(namen)
	out := map[string][]string{}
	for _, n := range namen {
		k := asciiKlein(n)
		out[k] = append(out[k], h[n]...)
	}
	return out
}

// hopByHopNamen ist die feste Liste der Hop-by-Hop-Header (RFC 9110 §7.6.1, SPEC-121 Nr. 4).
func hopByHopNamen() []string {
	return []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"TE", "Trailer", "Transfer-Encoding", "Upgrade"}
}

// teileKomma trennt einen Wert an Kommas und entfernt Leerzeichen und Tabulatoren am Rand jedes Teils.
func teileKomma(w string) []string {
	var out []string
	for _, t := range strings.Split(w, ",") {
		out = append(out, strings.Trim(t, " \t"))
	}
	return out
}

// verbindungsnamen sind die Namen, die ein Wert von Connection nennt, klein; leere Teile haben keine Wirkung.
func verbindungsnamen(h map[string][]string) []string {
	var out []string
	for _, v := range nachName(h)["connection"] {
		for _, t := range teileKomma(v) {
			if t != "" {
				out = append(out, asciiKlein(t))
			}
		}
	}
	return out
}

// ausgenommen ist die Menge der Namen, die der Vergleich auslässt (LH-QA-07.a Nr. 2): Hop-by-Hop, die Namen aus
// Connection beider Antworten, VOLATILE_HEADERS, Content-Length und, außer im Modus roh, Content-Encoding.
func (r vergleichsregeln) ausgenommen(direkt, proxy map[string][]string) map[string]bool {
	namen := slices.Concat(hopByHopNamen(), r.volatile, verbindungsnamen(direkt), verbindungsnamen(proxy), []string{"Content-Length"})
	if !r.roh {
		namen = append(namen, "Content-Encoding")
	}
	m := map[string]bool{}
	for _, n := range namen {
		m[asciiKlein(n)] = true
	}
	return m
}

// relevante liefert die Feldzeilen je kleingeschriebenem Namen ohne die ausgenommenen Namen.
func relevante(h map[string][]string, aus map[string]bool) map[string][]string {
	out := map[string][]string{}
	for n, v := range nachName(h) {
		if !aus[n] {
			out[n] = v
		}
	}
	return out
}

// relevanteNamen sind die Namen der relevanten Header einer Antwort, sortiert, klein.
func (r vergleichsregeln) relevanteNamen(a rohantwort) []string {
	var namen []string
	for n := range relevante(a.header, r.ausgenommen(a.header, nil)) {
		namen = append(namen, n)
	}
	slices.Sort(namen)
	return namen
}

// kette liest die Kette von Content-Encoding (SPEC-119 Nr. 6): geteilt an Kommas über alle Zeilen, klein; ein leeres
// Glied, ein unbekanntes Glied und mehr als zwei Glieder sind ein Fehler des Tests.
func kette(h map[string][]string) ([]string, error) {
	var glieder []string
	for _, v := range nachName(h)["content-encoding"] {
		for _, t := range teileKomma(v) {
			g := asciiKlein(t)
			switch g {
			case "":
				return nil, errors.New("leeres Glied in Content-Encoding")
			case "gzip", "deflate", "identity":
				glieder = append(glieder, g)
			default:
				return nil, fmt.Errorf("coding %q nicht entpackbar", g)
			}
		}
	}
	if len(glieder) > 2 {
		return nil, fmt.Errorf("kette mit %d Gliedern nicht entpackbar", len(glieder))
	}
	return glieder, nil
}

// entpackeGlied nimmt eine Stufe ab: gzip (auch mehrere Glieder hintereinander) oder deflate als zlib-Format (RFC 1950).
func entpackeGlied(coding string, b []byte) ([]byte, error) {
	var r io.Reader
	var err error
	switch coding {
	case "gzip":
		r, err = gzip.NewReader(bytes.NewReader(b))
	case "deflate":
		r, err = zlib.NewReader(bytes.NewReader(b))
	default:
		return b, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", coding, err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", coding, err)
	}
	return out, nil
}

// entpacke liefert den Body nach dem Entpacken in umgekehrter Reihenfolge der Kette; ein Rohbody ohne ein Byte
// ist bei jeder bekannten Kette leer, ohne den Dekoder zu rufen (LH-QA-07.a Nr. 3).
func entpacke(h map[string][]string, roh []byte) ([]byte, error) {
	glieder, err := kette(h)
	if err != nil {
		return nil, err
	}
	if len(roh) == 0 {
		return nil, nil
	}
	b := roh
	for i := len(glieder) - 1; i >= 0; i-- {
		if b, err = entpackeGlied(glieder[i], b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// kodiere wendet die Glieder in der genannten Reihenfolge an (das erste zuerst, wie Content-Encoding sie nennt).
func kodiere(glieder []string, inhalt []byte) []byte {
	b := inhalt
	for _, g := range glieder {
		var buf bytes.Buffer
		switch g {
		case "gzip":
			w := gzip.NewWriter(&buf)
			_, _ = w.Write(b)
			_ = w.Close()
		case "deflate":
			w := zlib.NewWriter(&buf)
			_, _ = w.Write(b)
			_ = w.Close()
		default:
			buf.Write(b)
		}
		b = buf.Bytes()
	}
	return b
}

// vergleiche vergleicht zwei Antworten nach LH-QA-07.a: den Statuscode (nicht die Phrase), die relevanten Header
// als Liste der Feldzeilen je Name und die Bytes des Bodys nach dem Entpacken, je Seite mit ihrer eigenen Kette;
// im Modus roh die Rohbytes. Eine Kette, die der Test nicht entpackt, ist ein Fehler, nie ein Überspringen.
func vergleiche(r vergleichsregeln, direkt, proxy rohantwort) ([]abweichung, error) {
	var ab []abweichung
	if direkt.status != proxy.status {
		ab = append(ab, abweichung{"Status", ""})
	}
	aus := r.ausgenommen(direkt.header, proxy.header)
	d, p := relevante(direkt.header, aus), relevante(proxy.header, aus)
	namen := make([]string, 0, len(d)+len(p))
	for n := range d {
		namen = append(namen, n)
	}
	for n := range p {
		namen = append(namen, n)
	}
	slices.Sort(namen)
	for _, n := range slices.Compact(namen) {
		dv, dda := d[n]
		pv, pda := p[n]
		if dda != pda || !slices.Equal(dv, pv) {
			ab = append(ab, abweichung{"Header", n})
		}
	}
	db, pb := direkt.body, proxy.body
	if !r.roh {
		var err error
		if db, err = entpacke(direkt.header, direkt.body); err != nil {
			return nil, fmt.Errorf("direkte Antwort: %w", err)
		}
		if pb, err = entpacke(proxy.header, proxy.body); err != nil {
			return nil, fmt.Errorf("über den Proxy: %w", err)
		}
	}
	if !bytes.Equal(db, pb) {
		ab = append(ab, abweichung{"Body", ""})
	}
	return ab, nil
}
