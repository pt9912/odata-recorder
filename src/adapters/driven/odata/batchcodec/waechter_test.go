package batchcodec_test

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/batchcodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// familie ist eine Eingabe des Wächters; stueck begrenzt die Reads, ursache "" ist ein Erfolg.
type familie struct {
	name    string
	body    string
	ct      string
	stueck  int
	ursache model.BatchUrsache
}

const mib = 1 << 20

func familien() []familie {
	voll := func(muster string, n int) string { return strings.Repeat(muster, n/len(muster)+1)[:n] }
	rumpf := func(body string) string {
		return huelle(block([]string{zeileCTH}, anfrage("POST x HTTP/1.1", nil, body)))
	}
	vieleKopf := block(append([]string{zeileCTH}, zeilen(99)...), anfrage("GET x HTTP/1.1", zeilen(100), ""))
	kurz := block(append([]string{zeileCTH}, wiederholt("a:", 99)...), anfrage("GET x HTTP/1.1", wiederholt("a:", 100), ""))
	return []familie{
		{name: "viele öffnende Grenzzeilen", body: voll("--b"+crlf, 8*mib) + crlf + "--b--", ursache: model.BatchZuViele},
		{name: "viele Zeilen der Präambel", body: voll(crlf, 8*mib) + crlf + "--b--"},
		{name: "viele Zeilen des Epilogs", body: "--b--" + voll(crlf+"--b", 8*mib)},
		{name: "Fast-Treffer", body: rumpf(voll(crlf+"--bx"+crlf+" --b", 8*mib) + "x")},
		{name: "langer Body", body: rumpf(voll("y", 8*mib))},
		{name: "lange Präambel", body: voll("p", 4*mib) + crlf + huelle()},
		{name: "viele Kopfzeilen", body: huelle(wiederholt(vieleKopf, 1000)...)},
		{name: "kurze Kopfzeilen", body: huelle(wiederholt(kurz, 1000)...)},
		{name: "viele Teile ohne Inhalt", body: voll("--b"+crlf, 8*mib), ursache: model.BatchRahmen},
		{name: "viele Parameter des Teils", body: huelle(block([]string{"Content-Type: multipart/mixed" + voll(";", mib)}, "")), ursache: model.BatchBoundary},
		{name: "viele SP in der Anfragezeile", body: huelle(block([]string{zeileCTH}, voll(" ", mib))), ursache: model.BatchAnfragezeile},
		{name: "kleine Reads", body: rumpf(voll("z", mib)), stueck: 64},
		{name: "lange Kopfzeile des Teils", body: huelle(block([]string{zeileCTH, "X: " + voll("v", mib)}, anfrage("GET x HTTP/1.1", nil, "")))},
		{name: "langer Name eines Teils", body: huelle(block([]string{zeileCTH, voll("N", mib) + ": v"}, anfrage("GET x HTTP/1.1", nil, "")))},
		{name: "lange Kopfzeile des Requests", body: huelle(block([]string{zeileCTH}, anfrage("GET x HTTP/1.1", []string{"X: " + voll("v", mib)}, "")))},
		{name: "lange Anfragezeile", body: huelle(httpTeil("GET " + voll("z", mib) + " HTTP/1.1"))},
		{name: "lange Füllung der Grenzzeile", body: "--b" + voll(" ", mib) + crlf + httpTeil("GET x HTTP/1.1") + crlf + "--b--" + voll(ht, mib)},
		{name: "langer Content-Type der Hülle", body: huelle(), ct: ctBatch + voll(";", mib)},
		{name: "leer", body: "", ursache: model.BatchRahmen},
		{name: "winzig", body: "--b--"},
	}
}

// TestBatchWaechter hält Zeile 33: die Gesamtzuteilung eines Aufrufs ist höchstens 256 · n + 1 MiB, n die Zahl der aus src
// erhaltenen Bytes; die langen Zeilen der Familien werden gelesen (Zeile 36).
func TestBatchWaechter(t *testing.T) {
	leser := batchcodec.NewLeser()
	for _, f := range familien() {
		ct := f.ct
		if ct == "" {
			ct = ctBatch
		}
		q := neueQuelle(f.body)
		q.hoechstens = f.stueck
		q.angefordert, q.geliefert = make([]int, 0, 1<<16), make([]int, 0, 1<<16)
		var vor, nach runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&vor)
		_, err := leser.LeseBatch(context.Background(), q, []string{ct}, grenzeAll)
		runtime.ReadMemStats(&nach)
		zuteilung := nach.TotalAlloc - vor.TotalAlloc
		hoechstens := 256*uint64(len(f.body)) + mib
		t.Logf("%s: n %d, Zuteilung %d, Grenze %d", f.name, len(f.body), zuteilung, hoechstens)
		if zuteilung > hoechstens {
			t.Errorf("%s: Zuteilung %d über %d bei n %d", f.name, zuteilung, hoechstens, len(f.body))
		}
		bf, ok := err.(*model.BatchFehler)
		if f.ursache == "" && err != nil || f.ursache != "" && (!ok || bf.Ursache != f.ursache) {
			t.Errorf("%s: %v, erwartet %q", f.name, err, f.ursache)
		}
	}
}

// alphabetOrakel nennt ohne Code des Lesers die Ursache eines Bodys über dem Alphabet des Rahmens bis zur Prüfgrenze: eine Zeile
// ist, was zwischen zwei CRLF steht; ohne `:` ist jede nichtleere Kopfzeile falsch, und kein Teil des Alphabets nennt einen
// Content-Type.
func alphabetOrakel(body string) (model.BatchUrsache, string) {
	zeilen := strings.Split(body, crlf)
	grenzzeile := func(z string) (bool, bool) {
		if !strings.HasPrefix(z, "--b") {
			return false, false
		}
		rest := z[3:]
		schliesst := strings.HasPrefix(rest, "--")
		if schliesst {
			rest = rest[2:]
		}
		return strings.Trim(rest, " "+ht) == "", schliesst
	}
	var treffer []int
	geschlossen := false
	for i, z := range zeilen {
		if ok, schliesst := grenzzeile(z); ok {
			treffer = append(treffer, i)
			if schliesst {
				geschlossen = true
				break
			}
		}
	}
	switch {
	case !geschlossen:
		return model.BatchRahmen, ""
	case len(treffer) == 1:
		return "", ""
	}
	inhalt := zeilen[treffer[0]+1 : treffer[1]]
	if len(inhalt) == 0 || inhalt[0] == "" {
		return model.BatchContentType, "1"
	}
	return model.BatchKopfzeile, "1"
}

// TestBatchAlphabet hält Zeile 34: jede Folge bis Länge 6 über {`--b`, `--`, CRLF, CR, LF, SP, `x`} hat die Ursache des Orakels.
func TestBatchAlphabet(t *testing.T) {
	alphabet := []string{"--b", "--", crlf, cr, lf, " ", "x"}
	leser := batchcodec.NewLeser()
	n := 0
	var geh func(body string, laenge int)
	geh = func(body string, laenge int) {
		n++
		got, err := leser.LeseBatch(context.Background(), strings.NewReader(body), []string{ctBatch}, grenzeAll)
		u, ort := alphabetOrakel(body)
		switch f, ok := err.(*model.BatchFehler); {
		case u == "":
			if err != nil || got.Teile == nil || len(got.Teile) != 0 {
				t.Fatalf("%q: %v, erwartet den leeren Batch", body, err)
			}
		case !ok || f.Ursache != u || f.Ort != ort:
			t.Fatalf("%q: %v, erwartet %s %q", body, err, u, ort)
		}
		if laenge < 6 {
			for _, a := range alphabet {
				geh(body+a, laenge+1)
			}
		}
	}
	geh("", 0)
	if n != 1+7+49+343+2401+16807+117649 {
		t.Errorf("%d Folgen", n)
	}
}
