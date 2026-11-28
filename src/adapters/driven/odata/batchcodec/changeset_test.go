package batchcodec_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// csMit ist ein Batch mit einem Einzelteil und einem Changeset an zweiter Stelle mit dem Content-Type ct und dem Rest.
func csMit(ct, rest string) (model.Batch, error) {
	return lese(huelle(httpTeil("GET x HTTP/1.1"), block([]string{"Content-Type: " + ct}, rest)))
}

func csBoundaryFall(t *testing.T, param, soll string) {
	t.Helper()
	rest := ebene("c")
	if soll != "" {
		rest = ebene(soll)
	}
	got, err := csMit("multipart/mixed; "+param, rest)
	if soll == "" {
		wantFehler(t, "Changeset "+strconv.Quote(param), got, err, model.BatchBoundary, "2")
		return
	}
	wantOK(t, "Changeset "+strconv.Quote(param), got, err, ergebnis(einzel(req("GET", "x")), menge()))
}

// TestBatchChangeset hält Zeile 25: die Grenzzeichenfolge eines Changeset nach den Regeln der Hülle (`boundary`, `rahmen` mit Ort
// des Changeset); die Hülle wird zuerst zerlegt. Die Bytes laufen über die Menge der Kopfzeilenwerte.
func TestBatchChangeset(t *testing.T) {
	n := 0
	for b := range 256 {
		if !wertErlaubt(b) {
			continue
		}
		c := string([]byte{byte(b)})
		erlaubt := strings.IndexByte(boundaryZeichen, byte(b)) >= 0
		for stelle, wert := range []string{c + "ab", "a" + c + "b", "ab" + c} {
			soll := ""
			if erlaubt && (c != " " || stelle != 2) {
				soll = wert
			}
			csBoundaryFall(t, `boundary="`+wert+`"`, soll)
			soll = ""
			switch {
			case (c == " " || c == ht) && stelle != 1:
				soll = "ab"
			case c == ";":
				soll = []string{"", "a", "ab"}[stelle]
			case erlaubt:
				soll = wert
			}
			csBoundaryFall(t, "boundary="+wert, soll)
			n += 2
		}
	}
	if n != (256-32)*3*2 {
		t.Errorf("%d Fälle", n)
	}
	for _, l := range []int{0, 1, 70, 71} {
		w, soll := strings.Repeat("a", l), ""
		if l >= 1 && l <= 70 {
			soll = w
		}
		csBoundaryFall(t, "boundary="+w, soll)
	}
	for _, param := range []string{"", "boundary", "boundary=c; boundary=c", "boundary=c; boundary=d", `boundary="`, "xboundary=c"} {
		csBoundaryFall(t, param, "")
	}
	for _, param := range []string{"BOUNDARY=c", "charset=utf-8; boundary=c", `boundary="c"; x=y`} {
		csBoundaryFall(t, param, "c")
	}
	// Gleiche Zeichenfolge wie die Hülle: ihre erste Grenzzeile beendet den Teil, das Changeset bleibt ohne Ende.
	got, err := csMit("multipart/mixed; boundary=b", ebene("b", httpTeil("POST y HTTP/1.1")))
	wantFehler(t, "gleiche Zeichenfolge", got, err, model.BatchRahmen, "2")
	// Eine Zeichenfolge als Anfang der anderen trennt sie nicht.
	for _, p := range []struct{ h, c string }{{"b", "bb"}, {"bb", "b"}, {"b", "b-x"}} {
		body := ebene(p.h, changeset(p.c, httpTeil("POST y HTTP/1.1"))) + crlf
		got, err := leseMit(strings.NewReader(body), []string{"multipart/mixed; boundary=" + p.h}, grenzeAll)
		wantOK(t, p.h+" in "+p.c, got, err, ergebnis(menge(req("POST", "y"))))
	}
	// Die Zeichenfolge `b--` im Changeset: ihre Grenzzeile ist die schließende der Hülle.
	got, err = lese(huelle(changeset("b--", httpTeil("POST y HTTP/1.1"))))
	wantFehler(t, "Zeichenfolge b--", got, err, model.BatchRahmen, "1")
	got, err = csMit("multipart/mixed; boundary=c", "--c"+crlf+httpTeil("POST y HTTP/1.1"))
	wantFehler(t, "ohne schließende Zeile", got, err, model.BatchRahmen, "2")
}

// TestBatchChangesetLeer hält Zeile 26: ein Changeset ohne Teil ist gültig, Requests leer und nicht nil.
func TestBatchChangesetLeer(t *testing.T) {
	for _, rest := range []string{"--c--", "--c--" + crlf, "x" + crlf + "--c--" + ht, "--c--" + crlf + "--c" + crlf} {
		got, err := csMit("multipart/mixed; boundary=c", rest)
		wantOK(t, strconv.Quote(rest), got, err, ergebnis(einzel(req("GET", "x")), menge()))
		if err == nil && got.Teile[1].Requests == nil {
			t.Errorf("%q: Requests ist nil", rest)
		}
	}
}

// TestBatchErgebnis hält Zeile 27: alle Folgen der Länge 0 bis 4 aus {Request, leeres Changeset, Changeset mit 1 und 2}; Teile und
// Requests in der Reihenfolge des Bodys, Slices nie nil; ein Fehler hinter gültigen Teilen ergibt den Nullwert.
func TestBatchErgebnis(t *testing.T) {
	folgen := [][]int{{}}
	for i := 0; i < len(folgen); i++ {
		if len(folgen[i]) < 4 {
			for art := range 4 {
				folgen = append(folgen, append(append([]int{}, folgen[i]...), art))
			}
		}
	}
	if len(folgen) != 341 {
		t.Fatalf("%d Folgen", len(folgen))
	}
	for _, folge := range folgen {
		var teile []string
		var soll []model.BatchTeil
		for i, art := range folge {
			ziel := func(k int) string { return "z" + strconv.Itoa(i) + "_" + strconv.Itoa(k) }
			switch art {
			case 0:
				teile = append(teile, httpTeil("GET "+ziel(0)+" HTTP/1.1"))
				soll = append(soll, einzel(req("GET", ziel(0))))
			default:
				var inner []string
				var requests []model.BatchRequest
				for k := range art - 1 {
					inner = append(inner, httpTeil("PUT "+ziel(k)+" HTTP/1.1"))
					requests = append(requests, req("PUT", ziel(k)))
				}
				teile = append(teile, changeset("c", inner...))
				soll = append(soll, menge(requests...))
			}
		}
		got, err := lese(huelle(teile...))
		wantOK(t, "Folge", got, err, ergebnis(soll...))
		if len(folge) > 0 {
			ort := strconv.Itoa(len(folge) + 1)
			got, err = lese(huelle(append(append([]string{}, teile...), httpTeil("GET x HTTP/1.0"))...))
			wantFehler(t, "Fehler hinter gültigen Teilen", got, err, model.BatchAnfragezeile, ort)
			got, err = lese(huelle(append(append([]string{}, teile...), changeset("c", httpTeil("PUT y HTTP/1.1"), "x"))...))
			wantFehler(t, "Fehler im Changeset hinter gültigen Teilen", got, err, model.BatchKopfzeile, ort+"/2")
		}
	}
}
