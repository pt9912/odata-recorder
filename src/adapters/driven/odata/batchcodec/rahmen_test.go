package batchcodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func mitBody(body string) model.BatchRequest {
	r := req("POST", "x")
	if body != "" {
		r.Body = []byte(body)
	}
	return r
}

func bodyTeil(body string) string {
	return block([]string{zeileCTH}, anfrage("POST x HTTP/1.1", nil, body))
}

// auffuellungen sind alle Folgen aus SP und HTAB der Länge 0 bis 3.
func auffuellungen() []string {
	out := []string{""}
	for i := 0; i < len(out); i++ {
		if len(out[i]) < 3 {
			out = append(out, out[i]+" ", out[i]+ht)
		}
	}
	return out
}

// grenzzeilenFall setzt die Zeile hinter trenner in den Body des ersten Requests der Ebene mit der Zeichenfolge g.
type grenzzeilenFall struct {
	trenner, zeile string
	art            int // 0 Inhalt, 1 öffnend, 2 schließend
}

func grenzzeilenFaelle(g string) []grenzzeilenFall {
	var out []grenzzeilenFall
	for _, p := range auffuellungen() {
		out = append(out,
			grenzzeilenFall{crlf, "--" + g + p, 1},
			grenzzeilenFall{crlf, "--" + g + "--" + p, 2},
			grenzzeilenFall{crlf, "--" + g + p + "x", 0},
			grenzzeilenFall{crlf, "--" + g + "--" + p + "x", 0},
			grenzzeilenFall{lf, "--" + g + p, 0},
			grenzzeilenFall{lf, "--" + g + "--" + p, 0},
			grenzzeilenFall{cr, "--" + g + p, 0},
			grenzzeilenFall{cr, "--" + g + "--" + p, 0},
			grenzzeilenFall{crlf, " --" + g + p, 0},
			grenzzeilenFall{crlf, ht + "--" + g + "--" + p, 0},
		)
	}
	for _, z := range []string{
		"--" + g + "x", "--" + g + "x--", "--" + g + "-", "--" + g + "- -", "--" + g + " --", "-" + g, "-" + g + "--", g, g + "--",
		"--" + strings.ToUpper(g), "--" + strings.ToUpper(g) + "--", "--" + g + lf, "--" + g + cr,
	} {
		out = append(out, grenzzeilenFall{crlf, z, 0})
	}
	return out
}

// TestBatchGrenzzeile hält Zeile 9: eine Grenzzeile nur am Anfang der Ebene oder hinter CRLF, bytegleich, wahlweise `--`, SP und
// HTAB, dann CRLF oder Ende; alles andere ist Inhalt. Je Fall × öffnend und schließend × beide Ebenen.
func TestBatchGrenzzeile(t *testing.T) {
	n := 0
	for _, imChangeset := range []bool{false, true} {
		g, ort := "grenze", "2"
		if imChangeset {
			g, ort = "inner", "1/2"
		}
		for _, f := range grenzzeilenFaelle(g) {
			teil := bodyTeil("X" + f.trenner + f.zeile + crlf + "Y")
			var body, ct string
			if imChangeset {
				body, ct = huelle(changeset(g, teil)), ctBatch
			} else {
				body, ct = ebene(g, teil)+crlf, "multipart/mixed; boundary="+g
			}
			got, err := leseMit(strings.NewReader(body), []string{ct}, grenzeAll)
			var want model.BatchRequest
			switch f.art {
			case 1:
				wantFehler(t, "öffnend "+f.zeile, got, err, model.BatchKopfzeile, ort)
				n++
				continue
			case 2:
				want = mitBody("X")
			default:
				want = mitBody("X" + f.trenner + f.zeile + crlf + "Y")
			}
			soll := ergebnis(einzel(want))
			if imChangeset {
				soll = ergebnis(menge(want))
			}
			wantOK(t, f.trenner+f.zeile, got, err, soll)
			n++
		}
	}
	if n != 2*(15*10+13) {
		t.Errorf("%d Fälle", n)
	}
	// Die Grenzzeile am Anfang der Ebene und am Ende ohne CRLF.
	got, err := lese("--b" + crlf + httpTeil("GET x HTTP/1.1") + crlf + "--b--")
	wantOK(t, "ohne CRLF am Ende", got, err, ergebnis(einzel(req("GET", "x"))))
	got, err = lese("--b")
	wantFehler(t, "öffnend am Ende", got, err, model.BatchRahmen, "")
	// Eine Zeichenfolge, die auf `--` endet: ihre Zeile ist öffnend.
	got, err = leseMit(strings.NewReader("--a--"+crlf+httpTeil("GET x HTTP/1.1")+crlf+"--a----"), []string{"multipart/mixed; boundary=a--"}, grenzeAll)
	wantOK(t, "Zeichenfolge auf --", got, err, ergebnis(einzel(req("GET", "x"))))
}

// TestBatchPraeambelEpilog hält Zeile 10: Präambel und Epilog ohne Wirkung; die Ebene endet an der ersten schließenden Grenzzeile.
func TestBatchPraeambelEpilog(t *testing.T) {
	weitere := crlf + "--b" + crlf + httpTeil("PUT z HTTP/1.1") + crlf + "--b--" + crlf + "Rest" + crlf + "--b"
	praeambel := []string{"", "Text" + crlf, "--bx" + crlf + " --b" + crlf + "a" + lf + "--b" + crlf, crlf + crlf, nul + crlf}
	for _, p := range praeambel {
		for _, e := range []string{"", crlf, weitere, " " + crlf + "x", crlf + "--b--"} {
			got, err := lese(p + ebene("b", httpTeil("GET x HTTP/1.1")) + e)
			wantOK(t, "Hülle "+p+e, got, err, ergebnis(einzel(req("GET", "x"))))
		}
	}
	weitereCS := crlf + "--c" + crlf + httpTeil("PUT z HTTP/1.1") + crlf + "--c--" + crlf + "--c"
	for _, p := range []string{"", "Text" + crlf, "--cx" + crlf} {
		for _, e := range []string{"", crlf, weitereCS} {
			cs := block([]string{"Content-Type: multipart/mixed; boundary=c"}, p+ebene("c", httpTeil("POST y HTTP/1.1"))+e)
			got, err := lese(huelle(cs, httpTeil("GET x HTTP/1.1")))
			wantOK(t, "Changeset "+p+e, got, err, ergebnis(menge(req("POST", "y")), einzel(req("GET", "x"))))
		}
	}
}

// TestBatchRahmen hält Zeile 11: ohne schließende Grenzzeile `rahmen`: leerer Body, nur Präambel, abgeschnitten an jeder Stelle,
// Zeilenenden nur LF, auf beiden Ebenen mit Ort.
func TestBatchRahmen(t *testing.T) {
	b := beispiel()
	ende := strings.LastIndex(b, "--batch_1--") + len("--batch_1--")
	for k := range ende {
		got, err := leseMit(strings.NewReader(b[:k]), ctBeispiel(), grenzeAll)
		wantFehler(t, "Anfang des Beispiels bis "+b[:k], got, err, model.BatchRahmen, "")
	}
	got, err := leseMit(strings.NewReader(b[:ende]), ctBeispiel(), grenzeAll)
	wantOK(t, "bis zum Ende der schließenden Zeile", got, err, beispielErgebnis())
	got, err = leseMit(strings.NewReader(strings.ReplaceAll(b, crlf, lf)), ctBeispiel(), grenzeAll)
	wantFehler(t, "nur LF", got, err, model.BatchRahmen, "")
	for _, s := range []string{"", "Text", crlf, "--b" + crlf, "--c--"} {
		got, err := lese(s)
		wantFehler(t, "ohne Grenzzeile "+s, got, err, model.BatchRahmen, "")
	}
	inner := ebene("c", httpTeil("POST y HTTP/1.1"))
	endeCS := strings.LastIndex(inner, "--c--") + len("--c--")
	kopf := []string{"Content-Type: multipart/mixed; boundary=c"}
	for k := range endeCS {
		got, err := lese(huelle(httpTeil("GET x HTTP/1.1"), block(kopf, inner[:k])))
		wantFehler(t, "Changeset abgeschnitten bei "+inner[:k], got, err, model.BatchRahmen, "2")
	}
	got, err = lese(huelle(httpTeil("GET x HTTP/1.1"), block(kopf, strings.ReplaceAll(inner, crlf, lf))))
	wantFehler(t, "Changeset nur LF", got, err, model.BatchRahmen, "2")
	got, err = lese(huelle(kopf[0]))
	wantFehler(t, "Changeset ohne leere Zeile", got, err, model.BatchRahmen, "1")
}

// TestBatchLeer hält Zeile 12: die erste Grenzzeile schließend ist der leere Batch mit leeren, nicht nil Teilen.
func TestBatchLeer(t *testing.T) {
	for _, praeambel := range []string{"", "x" + crlf} {
		for _, ende := range []string{"", crlf, " " + ht, " " + crlf, crlf + "Epilog", crlf + "--b" + crlf} {
			got, err := lese(praeambel + "--b--" + ende)
			wantOK(t, praeambel+ende, got, err, ergebnis())
		}
	}
}

// TestBatchInhalt hält Zeile 13: das CRLF vor der Grenzzeile gehört nicht zum Inhalt, davor stehende bleiben; ein leerer Teil ist
// `content_type`.
func TestBatchInhalt(t *testing.T) {
	for _, rumpf := range []string{"x", "{}", nul + esB(0x80, 0xFF, 0x00), cr, lf, "a" + cr, "a" + lf, lf + cr} {
		for k := range 4 {
			body := rumpf + strings.Repeat(crlf, k)
			got, err := lese(huelle(bodyTeil(body), httpTeil("GET x HTTP/1.1")))
			wantOK(t, "Body "+body, got, err, ergebnis(einzel(mitBody(body)), einzel(req("GET", "x"))))
		}
	}
	for k := range 4 {
		got, err := lese(huelle(bodyTeil(strings.Repeat(crlf, k))))
		wantOK(t, "nur CRLF", got, err, ergebnis(einzel(mitBody(strings.Repeat(crlf, k)))))
	}
	for _, leer := range []string{
		"--b" + crlf + "--b--", "--b" + crlf + crlf + "--b--", "--b" + crlf + crlf + crlf + "--b--",
	} {
		got, err := lese(leer)
		wantFehler(t, "leerer Teil "+leer, got, err, model.BatchContentType, "1")
	}
}
