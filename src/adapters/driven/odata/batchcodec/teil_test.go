package batchcodec_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func wiederholt(teil string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = teil
	}
	return out
}

// TestBatchTeilzahl hält Zeile 17: 1000 Teile über beide Ebenen sind gültig, 1001 zu_viele mit dem Ort der Ebene; der Rahmen der
// Ebene geht vor.
func TestBatchTeilzahl(t *testing.T) {
	r := httpTeil("GET x HTTP/1.1")
	cs := func(n int) string { return changeset("c", wiederholt(r, n)...) }
	zehn := func(letzte int) []string { return append(wiederholt(cs(99), 9), cs(letzte)) }
	for _, f := range []struct {
		name  string
		teile []string
		ort   string // "" mit gueltig falsch: Hülle
		ok    bool
	}{
		{"1000 in der Hülle", wiederholt(r, 1000), "", true},
		{"1001 in der Hülle", wiederholt(r, 1001), "", false},
		{"999 im Changeset", []string{cs(999)}, "", true},
		{"1000 im Changeset", []string{cs(1000)}, "1", false},
		{"1001 im Changeset", []string{cs(1001)}, "1", false},
		{"zehn Changesets zu je 99", zehn(99), "", true},
		{"das zehnte mit 100", zehn(100), "10", false},
		{"Hülle und Changeset", append(wiederholt(r, 499), cs(500)), "", true},
		{"Hülle und Changeset darüber", append(wiederholt(r, 499), cs(501)), "500", false},
		{"Changeset vorn, Hülle zählt zuerst", append([]string{cs(1)}, wiederholt(r, 1000)...), "", false},
		{"leere Changesets zählen als Teil", wiederholt(cs(0), 1000), "", true},
		{"1001 leere Changesets", wiederholt(cs(0), 1001), "", false},
		{"Verstoß vor dem Überschreiten", append([]string{"x", cs(1000)}, r), "", false},
	} {
		got, err := lese(huelle(f.teile...))
		switch {
		case f.name == "Verstoß vor dem Überschreiten":
			wantFehler(t, f.name, got, err, model.BatchKopfzeile, "1")
		case f.ok:
			if err != nil || len(got.Teile) != len(f.teile) {
				t.Errorf("%s: %v, %d Teile", f.name, err, len(got.Teile))
			}
		default:
			wantFehler(t, f.name, got, err, model.BatchZuViele, f.ort)
		}
	}
	// Der Rahmen der Ebene geht der Zahl vor.
	offen := strings.Repeat("--b"+crlf+r+crlf, 1001)
	got, err := lese(offen)
	wantFehler(t, "1001 ohne schließende Zeile", got, err, model.BatchRahmen, "")
	offenCS := strings.Repeat("--c"+crlf+r+crlf, 1001)
	got, err = lese(huelle(block([]string{"Content-Type: multipart/mixed; boundary=c"}, offenCS)))
	wantFehler(t, "1001 im offenen Changeset", got, err, model.BatchRahmen, "1")
	// Die Zahl geht den Teilen vor.
	got, err = lese(huelle(wiederholt("x", 1001)...))
	wantFehler(t, "1001 fehlerhafte Teile", got, err, model.BatchZuViele, "")
}

func schreibweisenNamen(name string) []string {
	return []string{name, strings.ToLower(name), strings.ToUpper(name), strings.ToLower(name[:1]) + strings.ToUpper(name[1:])}
}

// TestBatchDoppelteKopfzeile hält Zeile 18: Content-Type, Content-Transfer-Encoding und Content-ID höchstens einmal je Teil in
// jeder Schreibweise, auf beiden Ebenen und im Teil eines Changeset; andere Namen dürfen sich wiederholen.
func TestBatchDoppelteKopfzeile(t *testing.T) {
	namen := []struct{ name, wert string }{{"Content-Type", "application/http"}, {"Content-Transfer-Encoding", "binary"}, {"Content-ID", "1"}}
	fremde := []string{"X-A: 1", "X-A: 1", "MIME-Version: 1.0"}
	n := 0
	for _, h := range namen {
		for _, zweiter := range schreibweisenNamen(h.name) {
			for abstand := range 3 {
				basis := []string{h.name + ": " + h.wert}
				if h.name != "Content-Type" {
					basis = append([]string{zeileCTH}, basis...)
				}
				kopf := append(append([]string{}, basis...), fremde[:abstand]...)
				kopf = append(kopf, zweiter+": "+h.wert)
				teil := block(kopf, anfrage("GET x HTTP/1.1", nil, ""))
				got, err := lese(huelle(teil))
				wantFehler(t, "Hülle "+zweiter, got, err, model.BatchKopfzeile, "1")
				got, err = lese(huelle(changeset("c", httpTeil("GET y HTTP/1.1"), teil)))
				wantFehler(t, "Changeset "+zweiter, got, err, model.BatchKopfzeile, "1/2")
				if h.name == "Content-Type" {
					kopf[0], kopf[len(kopf)-1] = "Content-Type: multipart/mixed; boundary=c", zweiter+": multipart/mixed; boundary=c"
				} else {
					kopf[0] = "Content-Type: multipart/mixed; boundary=c"
				}
				got, err = lese(huelle(block(kopf, ebene("c"))))
				wantFehler(t, "Teil des Changeset "+zweiter, got, err, model.BatchKopfzeile, "1")
				n += 3
			}
		}
	}
	if n != 3*4*3*3 {
		t.Errorf("%d Fälle", n)
	}
	for _, name := range schreibweisenNamen("Content-Length") {
		got, err := lese(huelle(block([]string{zeileCTH, name + ": 1", name + ": 2", "X: 1", "X: 2"}, anfrage("GET x HTTP/1.1", nil, ""))))
		wantOK(t, "fremder Name doppelt "+name, got, err, ergebnis(einzel(req("GET", "x"))))
	}
}

// TestBatchTeilTyp hält Zeile 19: fehlend oder leer `content_type`; application/http ein Request; multipart/mixed in der Hülle ein
// Changeset, im Changeset `verschachtelt`; jeder andere `teil_typ`. Je Wert × Schreibweise × Parameter × beide Ebenen.
func TestBatchTeilTyp(t *testing.T) {
	type fall struct {
		wert          string
		huelle, innen model.BatchUrsache // "" gültig
	}
	faelle := []fall{
		{"application/http", "", ""},
		{"multipart/mixed", "", model.BatchVerschachtelt},
		{"application/json", model.BatchTeilTyp, model.BatchTeilTyp},
		{"text/plain", model.BatchTeilTyp, model.BatchTeilTyp},
		{"application/httpx", model.BatchTeilTyp, model.BatchTeilTyp},
		{"application /http", model.BatchTeilTyp, model.BatchTeilTyp},
		{"application/http x", model.BatchTeilTyp, model.BatchTeilTyp},
		{"multipart/related", model.BatchTeilTyp, model.BatchTeilTyp},
		{"", model.BatchTeilTyp, model.BatchTeilTyp},
		{"application/j" + esB(0xC5, 0xBF) + "on", model.BatchTeilTyp, model.BatchTeilTyp},
	}
	for _, f := range faelle {
		for _, wert := range []string{f.wert, strings.ToUpper(f.wert), strings.ToUpper(f.wert[:min(1, len(f.wert))]) + f.wert[min(1, len(f.wert)):]} {
			for _, param := range []string{"", ";", "; boundary=c", ht + "; charset=utf-8; boundary=c"} {
				zeile := "Content-Type: " + wert + param
				if f.wert == "" {
					zeile = "Content-Type: ;" + param
				}
				rest := anfrage("GET x HTTP/1.1", nil, "")
				if f.wert == "multipart/mixed" {
					rest = ebene("c")
				}
				hu := f.huelle
				if f.wert == "multipart/mixed" && !strings.Contains(param, "boundary") {
					hu = model.BatchBoundary
				}
				teil := block([]string{zeile}, rest)
				got, err := lese(huelle(teil))
				pruefeArt(t, "Hülle "+zeile, got, err, hu, "1")
				if f.innen == "" {
					got, err = lese(huelle(changeset("o", teil)))
					pruefeArt(t, "Changeset "+zeile, got, err, "", "")
				} else {
					got, err = lese(huelle(changeset("o", teil)))
					pruefeArt(t, "Changeset "+zeile, got, err, f.innen, "1/1")
				}
			}
		}
	}
	for _, kopf := range [][]string{nil, {"X: y"}, {"Content-Type:"}, {"Content-Type: " + ht + " "}} {
		got, err := lese(huelle(block(kopf, anfrage("GET x HTTP/1.1", nil, ""))))
		wantFehler(t, "ohne Typ "+strings.Join(kopf, "|"), got, err, model.BatchContentType, "1")
		got, err = lese(huelle(changeset("o", block(kopf, anfrage("GET x HTTP/1.1", nil, "")))))
		wantFehler(t, "ohne Typ im Changeset "+strings.Join(kopf, "|"), got, err, model.BatchContentType, "1/1")
	}
}

// pruefeArt verlangt die Ursache oder, ohne Ursache, einen Erfolg.
func pruefeArt(t *testing.T, was string, got model.Batch, err error, u model.BatchUrsache, ort string) {
	t.Helper()
	if u == "" {
		if err != nil || len(got.Teile) != 1 {
			t.Errorf("%s: %v", was, err)
		}
		return
	}
	wantFehler(t, was, got, err, u, ort)
}

// TestBatchTransferKodierung hält Zeile 20: fehlend, binary, 7bit, 8bit gültig, sonst `transfer_kodierung`; je Wert × Schreibweise
// × Teilart.
func TestBatchTransferKodierung(t *testing.T) {
	werte := []struct {
		wert string
		ok   bool
	}{{"binary", true}, {"7bit", true}, {"8bit", true}, {"", false}, {"base64", false}, {"quoted-printable", false},
		{"binary; x=1", false}, {"7 bit", false}, {"16bit", false}, {"binaryx", false}, {geheim, false}}
	n := 0
	for _, w := range werte {
		for _, wert := range []string{w.wert, strings.ToUpper(w.wert), " " + w.wert + ht} {
			for _, name := range schreibweisenNamen("Content-Transfer-Encoding") {
				kopf := name + ": " + wert
				request := block([]string{zeileCTH, kopf}, anfrage("GET x HTTP/1.1", nil, ""))
				cs := block([]string{"Content-Type: multipart/mixed; boundary=c", kopf}, ebene("c", httpTeil("POST y HTTP/1.1")))
				for _, f := range []struct{ teil, ort string }{
					{request, "1"}, {cs, "1"}, {changeset("o", httpTeil("POST y HTTP/1.1"), request), "1/2"},
				} {
					got, err := lese(huelle(f.teil))
					u := model.BatchUrsache("")
					if !w.ok {
						u = model.BatchTransferKodierung
					}
					pruefeArt(t, kopf, got, err, u, f.ort)
					n++
				}
			}
		}
	}
	if n != 11*3*4*3 {
		t.Errorf("%d Fälle", n)
	}
}

// contentIDFall liest ein Beispiel mit der Content-ID an einem der drei Orte.
func contentIDFall(zeile string, ort int) (model.Batch, error) {
	switch ort {
	case 0:
		return lese(huelle(block([]string{zeileCTH, zeile}, anfrage("GET x HTTP/1.1", nil, ""))))
	case 1:
		return lese(huelle(block([]string{"Content-Type: multipart/mixed; boundary=c", zeile}, ebene("c", httpTeil("POST y HTTP/1.1")))))
	default:
		return lese(huelle(block([]string{zeileCTH}, anfrage("GET x HTTP/1.1", []string{zeile}, ""))))
	}
}

// TestBatchContentID hält Zeile 21: alle 256 Bytes, die Längen 0, 1, 128, 129 und drei Orte; Wirkung allein im Block eines Teils
// mit application/http.
func TestBatchContentID(t *testing.T) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	pruefe := func(wert string, gueltig, form bool) {
		zeile := "Content-ID: " + wert
		for ort := range 3 {
			got, err := contentIDFall(zeile, ort)
			was := strconv.Quote(wert) + " Ort " + strconv.Itoa(ort)
			switch {
			case !form:
				wantFehler(t, was, got, err, model.BatchKopfzeile, "1")
			case ort == 0 && !gueltig:
				wantFehler(t, was, got, err, model.BatchContentID, "1")
			case err != nil:
				t.Errorf("%s: %v", was, err)
			case ort == 0 && got.Teile[0].Requests[0].ContentID != strings.Trim(wert, " "+ht):
				t.Errorf("%s: ContentID %q", was, got.Teile[0].Requests[0].ContentID)
			case ort != 0 && got.Teile[0].Requests[0].ContentID != "":
				t.Errorf("%s: wirkt außerhalb des Teils: %q", was, got.Teile[0].Requests[0].ContentID)
			}
		}
	}
	n := 0
	for b := range 256 {
		c := string([]byte{byte(b)})
		for stelle, wert := range []string{c + "ac", "a" + c + "c", "ac" + c} {
			gueltig := strings.IndexByte(alphabet, byte(b)) >= 0 || (c == " " || c == ht) && stelle != 1
			pruefe(wert, gueltig, wertErlaubt(b))
			n++
		}
	}
	if n != 256*3 {
		t.Errorf("%d Fälle", n)
	}
	for _, f := range []struct {
		l  int
		ok bool
	}{{0, false}, {1, true}, {128, true}, {129, false}} {
		pruefe(strings.Repeat("a", f.l), f.ok, true)
	}
	got, err := lese(huelle(httpTeil("GET x HTTP/1.1")))
	if err != nil || got.Teile[0].Requests[0].ContentID != "" {
		t.Errorf("ohne Content-ID: %v", err)
	}
	got, err = lese(huelle(changeset("c", idTeil("a.1", "POST y HTTP/1.1"), idTeil("a.1", "POST y HTTP/1.1"))))
	if err != nil || got.Teile[0].Requests[1].ContentID != "a.1" {
		t.Errorf("Content-ID im Changeset: %v", err)
	}
}
