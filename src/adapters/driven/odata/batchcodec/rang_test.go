package batchcodec_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// bauplan ist ein Teil mit application/http, den die Verstöße verändern.
type bauplan struct {
	teilZeilen []string
	typ        string
	kodierung  string
	id         string
	anfrage    string
	kopf       []string
}

func (b bauplan) teil() string {
	zeilen := append([]string{}, b.teilZeilen...)
	if b.typ != "" {
		zeilen = append(zeilen, "Content-Type: "+b.typ)
	}
	if b.kodierung != "" {
		zeilen = append(zeilen, "Content-Transfer-Encoding: "+b.kodierung)
	}
	if b.id != "" {
		zeilen = append(zeilen, "Content-ID: "+b.id)
	}
	return block(zeilen, anfrage(b.anfrage, b.kopf, ""))
}

func gutPlan() bauplan { return bauplan{typ: "application/http", anfrage: "GET x HTTP/1.1"} }

// verstoss ist ein Verstoß eines Teils mit seiner Stufe in der Folge von SPEC-182 Nr. 5 g.
type verstoss struct {
	stufe int
	u     model.BatchUrsache
	tu    func(*bauplan)
}

func verstoesse() []verstoss {
	return []verstoss{
		{1, model.BatchKopfzeile, func(b *bauplan) { b.teilZeilen = []string{"ohne Doppelpunkt"} }},
		{1, model.BatchZuViele, func(b *bauplan) { b.teilZeilen = zeilen(101) }},
		{2, model.BatchContentType, func(b *bauplan) { b.typ = "" }},
		{2, model.BatchTeilTyp, func(b *bauplan) { b.typ = "text/plain" }},
		{3, model.BatchTransferKodierung, func(b *bauplan) { b.kodierung = "base64" }},
		{4, model.BatchContentID, func(b *bauplan) { b.id = "!" }},
		{5, model.BatchAnfragezeile, func(b *bauplan) { b.anfrage = "GET x HTTP/1.0" }},
		{6, model.BatchKopfzeile, func(b *bauplan) { b.kopf = append(b.kopf, "ohne Doppelpunkt") }},
		{6, model.BatchZuViele, func(b *bauplan) { b.kopf = append(b.kopf, zeilen(101)...) }},
		{7, model.BatchTransferKodierung, func(b *bauplan) { b.kopf = append(b.kopf, "Transfer-Encoding: chunked") }},
	}
}

// changesetVerstoesse sind die Verstöße eines Changeset als Teil der Hülle mit der Unterstelle des Orts.
func changesetVerstoesse() []struct {
	u     model.BatchUrsache
	unter string
	teil  string
} {
	offen := block([]string{"Content-Type: multipart/mixed; boundary=c"}, "--c"+crlf+httpTeil("PUT y HTTP/1.1"))
	return []struct {
		u     model.BatchUrsache
		unter string
		teil  string
	}{
		{model.BatchBoundary, "", block([]string{"Content-Type: multipart/mixed"}, ebene("c"))},
		{model.BatchRahmen, "", offen},
		{model.BatchVerschachtelt, "/1", changeset("c", changeset("d"))},
		{model.BatchAnfragezeile, "/1", changeset("c", httpTeil("PUT y HTTP/1.0"))},
		{model.BatchTransferKodierung, "", block([]string{"Content-Type: multipart/mixed; boundary=c", "Content-Transfer-Encoding: base64"}, ebene("c"))},
	}
}

type verstossTeil struct {
	u     model.BatchUrsache
	unter string
	teil  string
}

func alleVerstossTeile() []verstossTeil {
	var out []verstossTeil
	for _, v := range verstoesse() {
		p := gutPlan()
		v.tu(&p)
		out = append(out, verstossTeil{v.u, "", p.teil()})
	}
	for _, v := range changesetVerstoesse() {
		out = append(out, verstossTeil{v.u, v.unter, v.teil})
	}
	return out
}

// TestBatchRang hält Zeile 30: alle Paare von Verstößen an verschiedenen Teilen in beiden Reihenfolgen und an demselben Teil nach
// der Folge von SPEC-182 Nr. 5, ohne Vorrang einer Klasse; ein Verstoß der Prüfung vor einem Verstoß des Lesers ergibt den des
// Lesers.
func TestBatchRang(t *testing.T) {
	gut := httpTeil("GET x HTTP/1.1")
	teile := alleVerstossTeile()
	n := 0
	for _, a := range teile {
		for _, b := range teile {
			got, err := lese(huelle(a.teil, b.teil))
			wantFehler(t, "Teil 1 vor Teil 2", got, err, a.u, "1"+a.unter)
			got, err = lese(huelle(gut, a.teil, gut, b.teil))
			wantFehler(t, "Teil 2 vor Teil 4", got, err, a.u, "2"+a.unter)
			n++
		}
	}
	if n != 15*15 {
		t.Errorf("%d Paare an verschiedenen Teilen", n)
	}
	n = 0
	vs := verstoesse()
	for _, a := range vs {
		for _, b := range vs {
			if a.stufe >= b.stufe {
				continue
			}
			for _, reihe := range [][]verstoss{{a, b}, {b, a}} {
				p := gutPlan()
				reihe[0].tu(&p)
				reihe[1].tu(&p)
				got, err := lese(huelle(p.teil()))
				wantFehler(t, "Stufe "+strconv.Itoa(a.stufe)+" vor "+strconv.Itoa(b.stufe), got, err, a.u, "1")
				got, err = lese(huelle(changeset("c", gut, p.teil())))
				wantFehler(t, "im Changeset, Stufe "+strconv.Itoa(a.stufe)+" vor "+strconv.Itoa(b.stufe), got, err, a.u, "1/2")
				n++
			}
		}
	}
	// 10 Verstöße, 45 Paare, 3 davon auf derselben Stufe, in beiden Reihenfolgen des Anwendens.
	if n != (45-3)*2 {
		t.Errorf("%d Paare an demselben Teil", n)
	}
	rangZeilen(t)
	rangEbenen(t)
	rangPruefung(t)
}

// rangZeilen: je Zeile des Blocks des Teils zuerst die Zahl, dann die Form, dann die Wiederholung.
func rangZeilen(t *testing.T) {
	t.Helper()
	for _, f := range []struct {
		zeilen []string
		u      model.BatchUrsache
	}{
		{[]string{"schlecht", zeileCTH, zeileCTH}, model.BatchKopfzeile},
		{[]string{zeileCTH, zeileCTH, "schlecht"}, model.BatchKopfzeile},
		{append(append([]string{zeileCTH}, zeilen(99)...), "schlecht"), model.BatchZuViele},
		{append(append([]string{zeileCTH}, zeilen(98)...), "schlecht", "x: 1"), model.BatchKopfzeile},
		{append(append([]string{zeileCTH, zeileCTH}, zeilen(100)...), "x: 1"), model.BatchKopfzeile},
	} {
		got, err := lese(huelle(block(f.zeilen, anfrage("GET x HTTP/1.1", nil, ""))))
		wantFehler(t, "Zeilenfolge", got, err, f.u, "1")
	}
}

// rangEbenen: je Ebene Rahmen, Zahl, Teile; ein Changeset wird ganz gelesen, bevor der nächste Teil der Hülle beginnt.
func rangEbenen(t *testing.T) {
	t.Helper()
	gut := httpTeil("GET x HTTP/1.1")
	schlecht := httpTeil("GET x HTTP/1.0")
	got, err := lese("--b" + crlf + schlecht + crlf + "--b" + crlf + "x")
	wantFehler(t, "Rahmen vor Teil", got, err, model.BatchRahmen, "")
	got, err = lese(huelle(gut, block([]string{"Content-Type: multipart/mixed; boundary=c"}, "--c"+crlf+schlecht)))
	wantFehler(t, "Rahmen des Changeset vor seinen Teilen", got, err, model.BatchRahmen, "2")
	got, err = lese(huelle(block([]string{"Content-Type: multipart/mixed"}, "--c"+crlf+schlecht)))
	wantFehler(t, "boundary des Changeset vor seiner Ebene", got, err, model.BatchBoundary, "1")
	got, err = lese(huelle(changeset("c", schlecht), "x"))
	wantFehler(t, "Changeset vor dem nächsten Teil", got, err, model.BatchAnfragezeile, "1/1")
	got, err = lese(huelle("x", changeset("c", schlecht)))
	wantFehler(t, "Teil vor dem Changeset", got, err, model.BatchKopfzeile, "1")
	q := neueQuelle(huelle("x"))
	got, err = leseMit(q, []string{"multipart/mixed; boundary=\"" + strings.Repeat("a", 71) + "\""}, grenzeAll)
	wantFehler(t, "Grenzzeichenfolge der Hülle vor jedem Teil", got, err, model.BatchBoundary, "")
	if q.reads != 0 {
		t.Errorf("src gelesen: %d", q.reads)
	}
	got, err = leseMit(strings.NewReader("x"), []string{"text/plain"}, 1)
	wantFehler(t, "Content-Type vor dem Strom", got, err, model.BatchContentType, "")
}

// rangPruefung: jeder Verstoß des Lesers geht jedem der Prüfung vor (SPEC-182 Nr. 5 h); der Leser meldet keinen davon selbst.
func rangPruefung(t *testing.T) {
	t.Helper()
	pruefung := []struct {
		u    model.BatchUrsache
		teil string
	}{
		{model.BatchMethode, changeset("c", httpTeil("GET x HTTP/1.1"))},
		{model.BatchContentIDDoppelt, changeset("c", idTeil("1", "POST a HTTP/1.1"), idTeil("1", "POST a HTTP/1.1"))},
		{model.BatchReferenz, changeset("c", idTeil("1", "PATCH $9 HTTP/1.1"))},
		{model.BatchMethodeNichtUnterstuetzt, changeset("c", httpTeil("MERGE x HTTP/1.1"))},
	}
	for _, p := range pruefung {
		got, err := lese(huelle(p.teil))
		if err != nil || len(got.Teile) != 1 {
			t.Errorf("%s: der Leser meldet selbst: %v", p.u, err)
		}
		for _, l := range alleVerstossTeile() {
			got, err := lese(huelle(p.teil, l.teil))
			wantFehler(t, "Prüfung "+string(p.u)+" vor Leser", got, err, l.u, "2"+l.unter)
		}
	}
}
