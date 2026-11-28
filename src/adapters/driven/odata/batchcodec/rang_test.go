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
		{1, model.BatchKopfzeile, func(b *bauplan) { b.teilZeilen = doppelteZeile() }},
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

// doppelteZeile ist die Wiederholung einer Kopfzeile eines Teils mit Wirkung.
func doppelteZeile() []string { return []string{"Content-ID: 1", "content-id: 1"} }

// csPlan ist ein Teil mit multipart/mixed, den die Verstöße verändern: Zeilen vor dem Content-Type, die Grenzzeichenfolge, die
// Transfer-Kodierung, der erste Teil seiner Ebene, eine fehlende schließende Grenzzeile und 1000 weitere Teile.
type csPlan struct {
	teilZeilen []string
	grenze     string
	kodierung  string
	erster     string
	offen      bool
	viele      bool
}

func gutCS(grenze string) csPlan { return csPlan{grenze: grenze, erster: httpTeil("PUT y HTTP/1.1")} }

func (c csPlan) teil() string {
	kopf := append([]string{}, c.teilZeilen...)
	kopf = append(kopf, "Content-Type: multipart/mixed; boundary="+c.grenze)
	if c.kodierung != "" {
		kopf = append(kopf, "Content-Transfer-Encoding: "+c.kodierung)
	}
	teile := []string{c.erster}
	if c.viele {
		for range 1000 {
			teile = append(teile, httpTeil("PUT y HTTP/1.1"))
		}
	}
	inhalt := ebene(c.grenze, teile...)
	if c.offen {
		inhalt = strings.TrimSuffix(inhalt, "--"+c.grenze+"--")
	}
	return block(kopf, inhalt)
}

// csVerstoss ist ein Verstoß an einem Changeset-Teil mit seiner Stufe in der Folge von SPEC-182 Nr. 5 g und f: 1 die Kopfzeilen,
// 3 Content-Transfer-Encoding, 4 boundary, 5 rahmen, 6 die Zahl der Teile, 7 die Teile. Die Stufe 2 (Content-Type) hat an einem
// Changeset-Teil der Hülle keinen Verstoß.
type csVerstoss struct {
	stufe int
	u     model.BatchUrsache
	unter string
	tu    func(*csPlan)
}

// csVerstoesse sind die Verstöße eines Changeset-Teils; der Verstoß boundary behält eine zerlegbare Zeichenfolge, damit die
// Verstöße der Ebene neben ihm stehen.
func csVerstoesse() []csVerstoss {
	out := []csVerstoss{
		{1, model.BatchKopfzeile, "", func(c *csPlan) { c.teilZeilen = []string{"ohne Doppelpunkt"} }},
		{1, model.BatchZuViele, "", func(c *csPlan) { c.teilZeilen = zeilen(101) }},
		{1, model.BatchKopfzeile, "", func(c *csPlan) { c.teilZeilen = doppelteZeile() }},
		{3, model.BatchTransferKodierung, "", func(c *csPlan) { c.kodierung = "base64" }},
		{4, model.BatchBoundary, "", func(c *csPlan) { c.grenze += "!" }},
		{5, model.BatchRahmen, "", func(c *csPlan) { c.offen = true }},
		{6, model.BatchZuViele, "", func(c *csPlan) { c.viele = true }},
		{7, model.BatchVerschachtelt, "/1", func(c *csPlan) { c.erster = changeset("d") }},
	}
	for _, v := range verstoesse() {
		p := gutPlan()
		v.tu(&p)
		teil := p.teil()
		out = append(out, csVerstoss{7, v.u, "/1", func(c *csPlan) { c.erster = teil }})
	}
	return out
}

// rangChangeset: alle Paare von Verstößen verschiedener Stufen an demselben Changeset-Teil, in beiden Reihenfolgen des Anwendens,
// als erster und als zweiter Teil der Hülle; jeder Verstoß ist auch allein einer.
func rangChangeset(t *testing.T) {
	t.Helper()
	gut := httpTeil("GET x HTTP/1.1")
	pruefe := func(was string, p csPlan, u model.BatchUrsache, unter string) {
		t.Helper()
		got, err := lese(huelle(p.teil()))
		wantFehler(t, was, got, err, u, "1"+unter)
		got, err = lese(huelle(gut, p.teil()))
		wantFehler(t, was+", als Teil 2", got, err, u, "2"+unter)
	}
	got, err := lese(huelle(gutCS("c").teil()))
	if err != nil || len(got.Teile) != 1 || !got.Teile[0].Changeset {
		t.Fatalf("Ausgangs-Changeset: %v", err)
	}
	vs := csVerstoesse()
	n := 0
	for i, a := range vs {
		p := gutCS("c")
		a.tu(&p)
		pruefe("allein "+strconv.Itoa(i), p, a.u, a.unter)
		for j, b := range vs {
			if a.stufe >= b.stufe {
				continue
			}
			for _, reihe := range [][]csVerstoss{{a, b}, {b, a}} {
				p := gutCS("c")
				reihe[0].tu(&p)
				reihe[1].tu(&p)
				pruefe("Changeset-Teil, Verstoß "+strconv.Itoa(i)+" vor "+strconv.Itoa(j), p, a.u, a.unter)
				n++
			}
		}
	}
	// 19 Verstöße, 171 Paare, davon auf derselben Stufe 3 (Kopfzeilen) und 66 (12 Verstöße der Teile).
	if len(vs) != 19 || n != (171-3-66)*2 {
		t.Errorf("%d Verstöße, %d Paare an demselben Changeset-Teil", len(vs), n)
	}
}

// rangGeschachtelt: an einem Changeset-Teil in einem Changeset gehen die Kopfzeilen dem Medientyp vor und der Medientyp
// (verschachtelt) der Transfer-Kodierung und allem, was seine Grenzzeichenfolge und seine Ebene trügen.
func rangGeschachtelt(t *testing.T) {
	t.Helper()
	kopfzeilen := []csVerstoss{{u: model.BatchVerschachtelt, tu: func(*csPlan) {}}}
	inhalte := []func(*csPlan){func(*csPlan) {}}
	for _, v := range csVerstoesse() {
		switch {
		case v.stufe == 1:
			kopfzeilen = append(kopfzeilen, v)
		case v.stufe >= 4:
			inhalte = append(inhalte, v.tu)
		}
	}
	n := 0
	for i, k := range kopfzeilen {
		for _, kodierung := range []string{"", "binary", "base64"} {
			for j, inhalt := range inhalte {
				p := gutCS("d")
				k.tu(&p)
				p.kodierung = kodierung
				inhalt(&p)
				got, err := lese(huelle(changeset("c", p.teil())))
				wantFehler(t, "geschachtelt "+strconv.Itoa(i)+" "+kodierung+" "+strconv.Itoa(j), got, err, k.u, "1/1")
				n++
			}
		}
	}
	// 4 Zustände der Kopfzeilen, 3 der Transfer-Kodierung, 16 des Inhalts (unverändert, boundary, rahmen, Zahl, 12 Teile).
	if n != 4*3*16 {
		t.Errorf("%d Fälle am geschachtelten Teil", n)
	}
}

// rangImChangeset: von zwei Verstößen an verschiedenen Teilen eines Changeset gilt der des früheren.
func rangImChangeset(t *testing.T) {
	t.Helper()
	gut := httpTeil("PUT y HTTP/1.1")
	var teile []verstossTeil
	for _, v := range verstoesse() {
		p := gutPlan()
		v.tu(&p)
		teile = append(teile, verstossTeil{v.u, "", p.teil()})
	}
	teile = append(teile, verstossTeil{model.BatchVerschachtelt, "", changeset("d")})
	n := 0
	for _, a := range teile {
		for _, b := range teile {
			got, err := lese(huelle(changeset("c", a.teil, b.teil)))
			wantFehler(t, "im Changeset Teil 1 vor Teil 2", got, err, a.u, "1/1")
			got, err = lese(huelle(changeset("c", gut, a.teil, gut, b.teil)))
			wantFehler(t, "im Changeset Teil 2 vor Teil 4", got, err, a.u, "1/2")
			n++
		}
	}
	if n != 12*12 {
		t.Errorf("%d Paare an verschiedenen Teilen eines Changeset", n)
	}
}

// TestBatchRang hält Zeile 30: alle Paare von Verstößen an verschiedenen Teilen in beiden Reihenfolgen, in der Hülle und in einem
// Changeset, und an demselben Teil (Request, Changeset, geschachteltes Changeset) nach der Folge von SPEC-182 Nr. 5, ohne Vorrang
// einer Klasse; ein Verstoß der Prüfung vor einem Verstoß des Lesers ergibt den des Lesers.
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
	if n != 16*16 {
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
	// 11 Verstöße, 55 Paare, 5 davon auf derselben Stufe (3 in den Kopfzeilen des Teils, 1 beim Medientyp, 1 in den Kopfzeilen
	// des Requests), in beiden Reihenfolgen des Anwendens.
	if n != (55-5)*2 {
		t.Errorf("%d Paare an demselben Teil", n)
	}
	rangChangeset(t)
	rangGeschachtelt(t)
	rangImChangeset(t)
	rangZeilen(t)
	rangEbenen(t)
	rangPruefung(t)
}

// rangZeilen: je Zeile des Blocks eines Teils zuerst die Zahl, dann die Form, dann die Wiederholung, für jede Art von Teil
// (Request, Changeset, geschachteltes Changeset).
func rangZeilen(t *testing.T) {
	t.Helper()
	n := 0
	for _, art := range []struct {
		typ, rest, ort string
		body           func(teil string) string
	}{
		{zeileCTH, anfrage("GET x HTTP/1.1", nil, ""), "1", func(teil string) string { return huelle(teil) }},
		{"Content-Type: multipart/mixed; boundary=c", ebene("c"), "1", func(teil string) string { return huelle(teil) }},
		{"Content-Type: multipart/mixed; boundary=d", ebene("d"), "1/1", func(teil string) string { return huelle(changeset("c", teil)) }},
	} {
		for i, f := range []struct {
			zeilen []string
			u      model.BatchUrsache
		}{
			{[]string{"schlecht", art.typ, art.typ}, model.BatchKopfzeile},
			{[]string{art.typ, art.typ, "schlecht"}, model.BatchKopfzeile},
			{append(append([]string{art.typ}, zeilen(99)...), "schlecht"), model.BatchZuViele},
			{append(append([]string{art.typ}, zeilen(98)...), "schlecht", "x: 1"), model.BatchKopfzeile},
			{append(append([]string{art.typ, art.typ}, zeilen(100)...), "x: 1"), model.BatchKopfzeile},
		} {
			got, err := lese(art.body(block(f.zeilen, art.rest)))
			wantFehler(t, "Zeilenfolge "+strconv.Itoa(i)+" an "+art.typ, got, err, f.u, art.ort)
			n++
		}
	}
	if n != 3*5 {
		t.Errorf("%d Zeilenfolgen", n)
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
