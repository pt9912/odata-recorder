package batchcodec_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// tchar sind die Zeichen eines Tokens nach RFC 9110 §5.6.2.
const tchar = "!#$%&'*+-.^_`|~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// wertErlaubt ist wahr für HTAB, 0x20 bis 0x7E und 0x80 bis 0xFF.
func wertErlaubt(b int) bool { return b == 0x09 || (b >= 0x20 && b <= 0x7E) || b >= 0x80 }

// mitKopfzeile setzt die Zeile in den Block des Teils oder des Requests.
func mitKopfzeile(zeile string, imRequest bool) (model.Batch, error) {
	if imRequest {
		return lese(huelle(block([]string{zeileCTH}, anfrage("GET x HTTP/1.1", []string{zeile}, ""))))
	}
	return lese(huelle(block([]string{zeileCTH, zeile}, anfrage("GET x HTTP/1.1", nil, ""))))
}

// kopfzeileFall verlangt, dass die Zeile gelesen wird (name != "") oder `kopfzeile` ist.
func kopfzeileFall(t *testing.T, zeile, name, wert string, imRequest bool) {
	t.Helper()
	got, err := mitKopfzeile(zeile, imRequest)
	if name == "" {
		wantFehler(t, strconv.Quote(zeile), got, err, model.BatchKopfzeile, "1")
		return
	}
	r := req("GET", "x")
	if imRequest {
		r.Header = []model.BatchKopf{{Name: name, Wert: wert}}
	}
	wantOK(t, strconv.Quote(zeile), got, err, ergebnis(einzel(r)))
}

// TestBatchKopfzeile hält Zeile 14: alle 256 Bytes im Namen und im Wert, am Anfang, in der Mitte und am Ende, im Block des Teils
// und des Requests; Fortsetzungszeile, Leerraum vor `:`, ohne `:` und leerer Name sind `kopfzeile`.
func TestBatchKopfzeile(t *testing.T) {
	n := 0
	for _, imRequest := range []bool{false, true} {
		for b := range 256 {
			c := string([]byte{byte(b)})
			istTchar := strings.IndexByte(tchar, byte(b)) >= 0
			for stelle, name := range []string{c + "Xy", "X" + c + "y", "Xy" + c} {
				switch {
				case c == ":" && stelle != 0:
					teil, _, _ := strings.Cut(name, ":")
					rest := name[len(teil)+1:]
					kopfzeileFall(t, name+": v", teil, strings.TrimLeft(rest+": v", " "+ht), imRequest)
				case istTchar:
					kopfzeileFall(t, name+": v", name, "v", imRequest)
				default:
					kopfzeileFall(t, name+": v", "", "", imRequest)
				}
				n++
			}
			for stelle, wert := range []string{c + "ac", "a" + c + "c", "ac" + c} {
				soll := wert
				if (c == " " || c == ht) && stelle != 1 {
					soll = "ac"
				}
				if wertErlaubt(b) {
					kopfzeileFall(t, "Xy: "+wert, "Xy", soll, imRequest)
				} else {
					kopfzeileFall(t, "Xy: "+wert, "", "", imRequest)
				}
				n++
			}
		}
		for _, z := range []string{
			"Xy v", "Xy", ":v", " Xy: v", ht + "Xy: v", " fortgesetzt", ht + "fortgesetzt", "Xy : v", "Xy" + ht + ": v", " : v",
			"Xy: a" + cr + "c", "Xy: a" + lf + "Z: b", geheim,
		} {
			kopfzeileFall(t, z, "", "", imRequest)
		}
		kopfzeileFall(t, "Xy:", "Xy", "", imRequest)
		kopfzeileFall(t, "Xy:"+" "+ht+" ", "Xy", "", imRequest)
		kopfzeileFall(t, "Xy:a: b", "Xy", "a: b", imRequest)
		// Eine Fortsetzungszeile hinter einer gültigen Zeile wird nicht angehängt.
		got, err := mitKopfzeile("Xy: v"+crlf+" w", imRequest)
		wantFehler(t, "Fortsetzung", got, err, model.BatchKopfzeile, "1")
	}
	if n != 2*256*6 {
		t.Errorf("%d Fälle", n)
	}
}

// TestBatchBlockEnde hält Zeile 15: ein Block endet an der leeren Zeile oder am Ende des Inhalts, mit und ohne CRLF hinter der
// letzten Zeile.
func TestBatchBlockEnde(t *testing.T) {
	kopf := req("GET", "x")
	kopf.Header = []model.BatchKopf{{Name: "A", Wert: "1"}}
	teilKopf := zeileCTH + crlf + crlf
	for _, ende := range []string{"", crlf, crlf + crlf} {
		got, err := lese(huelle(teilKopf + "GET x HTTP/1.1" + crlf + "A: 1" + ende))
		wantOK(t, "Request mit Kopfzeile und "+strconv.Quote(ende), got, err, ergebnis(einzel(kopf)))
		got, err = lese(huelle(teilKopf + "GET x HTTP/1.1" + ende))
		wantOK(t, "Request ohne Kopfzeile und "+strconv.Quote(ende), got, err, ergebnis(einzel(req("GET", "x"))))
		// Der Block des Teils ohne Request: die Anfragezeile fehlt.
		got, err = lese(huelle(zeileCTH + ende))
		wantFehler(t, "Block des Teils "+strconv.Quote(ende), got, err, model.BatchAnfragezeile, "1")
		// Die Art entscheidet der Block, auch ohne leere Zeile.
		got, err = lese(huelle("Content-Type: text/plain" + ende))
		wantFehler(t, "anderer Typ "+strconv.Quote(ende), got, err, model.BatchTeilTyp, "1")
		got, err = lese(huelle("Content-Type: multipart/mixed; boundary=c" + ende))
		wantFehler(t, "Changeset "+strconv.Quote(ende), got, err, model.BatchRahmen, "1")
	}
	// Ein leerer Block: die erste Zeile ist leer.
	got, err := lese(huelle(crlf + zeileCTH))
	wantFehler(t, "leere erste Zeile", got, err, model.BatchContentType, "1")
}

// zeilen liefert n Kopfzeilen.
func zeilen(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "X" + strconv.Itoa(i) + ": v"
	}
	return out
}

// blockZeilen liefert einen Teil mit nTeil Zeilen im Block des Teils (die erste Content-Type) und nReq im Block des Requests.
func blockZeilen(nTeil, nReq int) string {
	return block(append([]string{zeileCTH}, zeilen(nTeil-1)...), anfrage("GET x HTTP/1.1", zeilen(nReq), ""))
}

// TestBatchZeilenzahl hält Zeile 16: 100 Zeilen je Block sind gültig, die 101. ist zu_viele, vor ihrer Form; je Block gezählt.
func TestBatchZeilenzahl(t *testing.T) {
	for _, nTeil := range []int{1, 99, 100, 101} {
		for _, nReq := range []int{0, 99, 100, 101} {
			got, err := lese(huelle(blockZeilen(nTeil, nReq)))
			if nTeil > 100 || nReq > 100 {
				wantFehler(t, "Zeilen "+strconv.Itoa(nTeil)+"/"+strconv.Itoa(nReq), got, err, model.BatchZuViele, "1")
				continue
			}
			if err != nil || len(got.Teile[0].Requests[0].Header) != nReq {
				t.Errorf("%d/%d: %v", nTeil, nReq, err)
			}
		}
	}
	teil := append([]string{zeileCTH}, zeilen(99)...)
	anf := zeilen(100)
	for _, letzte := range []string{"ohne Doppelpunkt", "Content-Type: x", " Fortsetzung", "A: " + nul} {
		// Die 101. Zeile ist zu_viele, gleich ihrer Form.
		got, err := lese(huelle(block(append(append([]string{}, teil...), letzte), anfrage("GET x HTTP/1.1", nil, ""))))
		wantFehler(t, "101. Zeile des Teils "+letzte, got, err, model.BatchZuViele, "1")
		got, err = lese(huelle(block([]string{zeileCTH}, anfrage("GET x HTTP/1.1", append(append([]string{}, anf...), letzte), ""))))
		wantFehler(t, "101. Zeile des Requests "+letzte, got, err, model.BatchZuViele, "1")
	}
	for _, letzte := range []string{"ohne Doppelpunkt", "Content-Type: x"} {
		// Die 100. Zeile wird auf ihre Form und Wiederholung geprüft.
		zt := append(append([]string{}, teil[:99]...), letzte)
		got, err := lese(huelle(block(zt, anfrage("GET x HTTP/1.1", nil, ""))))
		wantFehler(t, "100. Zeile "+letzte, got, err, model.BatchKopfzeile, "1")
	}
	// Die letzte Zeile ohne CRLF zählt.
	got, err := lese(huelle(block([]string{zeileCTH}, "GET x HTTP/1.1"+crlf+strings.Join(zeilen(101), crlf))))
	wantFehler(t, "101 Zeilen ohne CRLF am Ende", got, err, model.BatchZuViele, "1")
	// Die Zählung ist je Teil, auch im Changeset.
	got, err = lese(huelle(changeset("c", blockZeilen(100, 100), blockZeilen(100, 100)), blockZeilen(100, 100)))
	if err != nil || len(got.Teile) != 2 {
		t.Errorf("drei volle Teile: %v", err)
	}
	got, err = lese(huelle(changeset("c", blockZeilen(100, 100), blockZeilen(101, 0))))
	wantFehler(t, "101 im Changeset", got, err, model.BatchZuViele, "1/2")
}
