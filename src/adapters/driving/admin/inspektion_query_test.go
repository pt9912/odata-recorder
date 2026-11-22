package admin_test

import (
	"fmt"
	"math"
	"math/big"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// inspLauf ist ein Server mit den Routen der Inspektion über einem Fake, der bei jedem Aufruf den Fehler liefert, mit dem
// der Test ihn baute: nil für die Fälle, die der Dienst erreicht, `ODR-E7003` oder `ODR-E7002` für die Fälle, die er nie
// erreichen darf (mit ihm wäre die Antwort 404 und nicht 400).
type inspLauf struct {
	u *umgebung
	f *inspektionFake
}

func neuerInspLauf(t *testing.T, fehler error) *inspLauf {
	t.Helper()
	u, f, _ := inspUmgebung(t)
	f.err = fehler
	return &inspLauf{u: u, f: f}
}

// sende sendet `GET <pfad>?<query>` und liefert die Antwort, den letzten Aufruf dieses Requests und ob der Dienst gerufen wurde.
func (l *inspLauf) sende(pfad, query string) (rec *httptest.ResponseRecorder, aufruf inspAufruf, gerufen bool) {
	vorher := l.f.gesamt()
	rec = l.u.rufe(inspReq(pfad, query))
	if args := l.f.argumente(); len(args) > vorher {
		return rec, args[len(args)-1], true
	}
	return rec, inspAufruf{}, false
}

// abgewiesen meldet, wenn der Request nicht mit 400 `ODR-E2102` ohne Aufruf des Dienstes endete.
func (l *inspLauf) abgewiesen(t *testing.T, name, pfad, query string) {
	t.Helper()
	rec, _, gerufen := l.sende(pfad, query)
	if rec.Code != 400 || rec.Body.String() != objekt400 || gerufen {
		t.Errorf("%s: Antwort %d %q, Dienst gerufen: %v; erwartet 400 %q ohne Aufruf", name, rec.Code, rec.Body.String(), gerufen, objekt400)
	}
}

// angenommen meldet, wenn der Request nicht mit 200 und genau einem Aufruf endete, und liefert den Aufruf.
func (l *inspLauf) angenommen(t *testing.T, name, pfad, query string) inspAufruf {
	t.Helper()
	rec, aufruf, gerufen := l.sende(pfad, query)
	if rec.Code != 200 || !gerufen {
		t.Errorf("%s: Antwort %d %q, Dienst gerufen: %v; erwartet 200 mit einem Aufruf", name, rec.Code, rec.Body.String(), gerufen)
	}
	return aufruf
}

// nurAbweisungen meldet, wenn die Telemetrie etwas anderes sah als je einen Aufruf von CountError mit `ODR-E2102` für jede
// der n Abweisungen, oder wenn Audit einen Aufruf sah.
func (l *inspLauf) nurAbweisungen(t *testing.T, name string, n int) {
	t.Helper()
	codes, andere := zaehlungenVon(l.u.tele)
	if len(codes) != n || andere != 0 || l.u.audit.Aufrufe() != 0 || slices.ContainsFunc(codes, func(c string) bool { return c != "ODR-E2102" }) {
		t.Errorf("%s: %d Zählungen (erwartet %d, alle ODR-E2102), %d andere Aufrufe der Telemetrie, %d Aufrufe von Audit", name, len(codes), n, andere, l.u.audit.Aufrufe())
	}
}

// inspRoute ist eine der zwei Routen mit Query: Pfad, der Fehler, den ein Dienst lieferte (er darf nie ankommen), und die
// gültige Query der Route, die die Fälle der ungültigen Formen umgibt.
type inspRoute struct {
	name, pfad, gueltig string
	fehler              error
}

func inspRoutenMitQuery() []inspRoute {
	return []inspRoute{
		{"exchanges", exchangesZiel("ses-1"), "limit=3&after=5", inspE7003()},
		{"coverage", inspCoverage, "baseline_id=b", inspE7002()},
	}
}

// inspStellen setzt die Form q allein, hinter und vor die gültige Query g (die drei Stellen der Achse *Stelle der Form*).
func inspStellen(g, q string) []string { return []string{q, g + "&" + q, q + "&" + g} }

// inspUngueltigeFormen sind die sechs Formen von M-Query, die sich nicht zerlegen lassen (SPEC-293 Nr. 3).
func inspUngueltigeFormen() []string {
	return []string{"%zz", "%", "a=%", "a=1;b=2", "%2", "limit=1&%zz"}
}

// TestI5QueryZerlegbar hält I5 (SPEC-293 Nr. 2, 3): M-Query, die ganze Query muss zerlegbar sein: 6 ungültige Formen × 3
// Stellen (allein, hinter und vor der gültigen Query) × 2 Routen = 36 Fälle, 400 `ODR-E2102` ohne Aufruf, auch wenn der Dienst
// `ODR-E7003` oder `ODR-E7002` lieferte; dieselben 6 Formen × 3 Stellen an `diagnostics` (18 Fälle) sind ohne Wirkung (200 und
// ein Aufruf von Diagnose).
func TestI5QueryZerlegbar(t *testing.T) {
	abweisungen, ohneWirkung := 0, 0
	for _, r := range inspRoutenMitQuery() {
		l := neuerInspLauf(t, r.fehler)
		for _, form := range inspUngueltigeFormen() {
			for _, q := range inspStellen(r.gueltig, form) {
				l.abgewiesen(t, r.name+" "+q, r.pfad, q)
				abweisungen++
			}
		}
		l.nurAbweisungen(t, r.name, 18)
	}
	l := neuerInspLauf(t, nil)
	for _, form := range inspUngueltigeFormen() {
		for _, q := range inspStellen("limit=3", form) {
			if a := l.angenommen(t, "diagnostics "+q, inspDiagnostics, q); a.op != inspOpDiagnose {
				t.Errorf("diagnostics %s: Operation %d, erwartet Diagnose", q, a.op)
			}
			ohneWirkung++
		}
	}
	gleichAuf(t, "Abweisungen, Fälle ohne Wirkung", []int{abweisungen, ohneWirkung}, []int{36, 18})
}

// TestI5UnbekannteSchluessel hält I5 (SPEC-293 Nr. 3): ein unbekannter Schlüssel ist ohne Wirkung, auch mehrfach, auch leer, auch
// der bekannte Schlüssel der anderen Route und eine andere Schreibweise: 8 Formen bei `exchanges` und 8 bei `coverage`, je
// neben der gültigen Query (hinter und vor) und bei `exchanges` auch allein (die Query von `coverage` ohne `baseline_id` ist
// ungültig), 24 + 16 = 40 Fälle, mit den Argumenten der gültigen Query unverändert; ein bekannter Schlüssel in
// Prozentschreibweise (`%6Cimit`) ist bekannt.
func TestI5UnbekannteSchluessel(t *testing.T) {
	unbekannt := map[string][]string{
		"exchanges": {"x=1", "x=1&x=2", "Limit=1", "AFTER=2", "baseline_id=z", "x=", "x", "x=1&y=2&x=3"},
		"coverage":  {"x=1", "x=1&x=2", "Limit=1", "limit=7", "after=7", "Baseline_Id=z", "x=", "x"},
	}
	faelle := 0
	for _, r := range inspRoutenMitQuery() {
		l := neuerInspLauf(t, nil)
		soll := inspAufruf{op: inspOpExchanges, id: "ses-1", seite: inbound.Seite{Limit: 3, Nach: 5}}
		if r.name == "coverage" {
			soll = inspAufruf{op: inspOpCoverage, id: "b"}
		}
		for _, form := range unbekannt[r.name] {
			stellen := inspStellen(r.gueltig, form)
			if r.name == "coverage" {
				stellen = stellen[1:]
			}
			for _, q := range stellen {
				faelle++
				erwartet := soll
				if q == form {
					erwartet.seite = inbound.Seite{}
				}
				a := l.angenommen(t, r.name+" "+q, r.pfad, q)
				if a.op != erwartet.op || a.id != erwartet.id || a.seite != erwartet.seite {
					t.Errorf("%s %s: Aufruf %d %q %+v, erwartet %d %q %+v", r.name, q, a.op, a.id, a.seite, erwartet.op, erwartet.id, erwartet.seite)
				}
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 8*3+8*2)
	l := neuerInspLauf(t, nil)
	if a := l.angenommen(t, "%6Cimit", exchangesZiel("s"), "%6Cimit=3&%61fter=4"); a.seite != (inbound.Seite{Limit: 3, Nach: 4}) {
		t.Errorf("Schlüssel in Prozentschreibweise: Seite %+v, erwartet {3 4}", a.seite)
	}
}

// inspGueltigeZahlen sind die 8 Werte von M-Zahl gültig (SPEC-293 Nr. 4) mit dem Wert als Literal; die Zahl des größten int
// steht als Literal und wird gegen math.MaxInt gehalten.
func inspGueltigeZahlen() []struct {
	query string
	wert  int
} {
	return []struct {
		query string
		wert  int
	}{
		{"0", 0}, {"1", 1}, {"007", 7}, {"100", 100}, {"1000", 1000}, {"1001", 1001}, {"%31%30", 10}, {"9223372036854775807", math.MaxInt},
	}
}

// TestI6GueltigeZahl hält I6 (SPEC-293 Nr. 4): M-Zahl gültig, 8 Werte × 2 Parameter = 16 Fälle (der Wert steht allein im
// jeweiligen Feld von Seite), dazu M-Stellung: 5 Formen und kein Parameter = 6 (fehlt `limit`, ist Limit 0, fehlt `after`, ist
// Nach 0; Reihenfolge und Stellung ändern nichts). Orakel Literale; Zählprüfung: 16 + 6 = 22 Fälle, der größte int aus math.MaxInt.
func TestI6GueltigeZahl(t *testing.T) {
	gleichAuf(t, "größter int", strconv.Itoa(math.MaxInt), "9223372036854775807")
	faelle := 0
	l := neuerInspLauf(t, nil)
	for _, z := range inspGueltigeZahlen() {
		for _, param := range []string{"limit", "after"} {
			faelle++
			soll := inbound.Seite{Limit: z.wert}
			if param == "after" {
				soll = inbound.Seite{Nach: z.wert}
			}
			if a := l.angenommen(t, param+"="+z.query, exchangesZiel("s"), param+"="+z.query); a.seite != soll {
				t.Errorf("%s=%s: Seite %+v, erwartet %+v", param, z.query, a.seite, soll)
			}
		}
	}
	stellungen := []struct {
		query string
		soll  inbound.Seite
	}{
		{"limit=3", inbound.Seite{Limit: 3}}, {"after=5", inbound.Seite{Nach: 5}}, {"limit=3&after=5", inbound.Seite{Limit: 3, Nach: 5}},
		{"after=5&limit=3", inbound.Seite{Limit: 3, Nach: 5}}, {"x=1&limit=3&y=2&after=5", inbound.Seite{Limit: 3, Nach: 5}}, {"", inbound.Seite{}},
	}
	for _, s := range stellungen {
		faelle++
		if a := l.angenommen(t, "Stellung "+s.query, exchangesZiel("s"), s.query); a.seite != s.soll {
			t.Errorf("Stellung %q: Seite %+v, erwartet %+v", s.query, a.seite, s.soll)
		}
	}
	gleichAuf(t, "Fälle", faelle, 22)
}

// inspUngueltigeZahlen sind die 17 Werte von M-Zahl ungültig (SPEC-293 Nr. 4) in der Schreibweise der Query: Ziffern außerhalb
// von ASCII als Prozentfolgen ihrer UTF-8-Bytes (arabisch-indische Eins U+0661, vollbreite Eins U+FF11).
func inspUngueltigeZahlen() []string {
	return []string{
		"", "-1", "-0", "+1", "%2B1", "%201", "1%20", "1+0", "1.0", "1e3", "0x10", "1_0", "abc", "%D9%A1", "%EF%BC%91",
		"9223372036854775808", strings.Repeat("9", 25),
	}
}

// inspStellungen setzt `param=text` allein, hinter und vor den gültigen Wert des anderen Parameters (Achse *Stelle des
// Verstoßes*: der andere Parameter ist gültig und gibt dem ungültigen keinen Vorrang).
func inspStellungen(param, text string) []string {
	andere := map[string]string{"limit": "after=5", "after": "limit=3"}[param]
	q := param + "=" + text
	return []string{q, andere + "&" + q, q + "&" + andere}
}

// TestI7UngueltigeZahl hält I7 (SPEC-293 Nr. 4): M-Zahl ungültig, 17 Werte × 2 Parameter = 34 Fälle, je allein, hinter und vor dem
// gültigen Wert des anderen Parameters (102 Fälle), dazu 4 Paare (beide Parameter ungültig, beide Reihenfolgen, je mit `-1` und
// `1.0`) = 106: 400 `ODR-E2102` ohne Aufruf, auch wenn der Dienst `ODR-E7003` liefern würde. Zählprüfung gegen die Zahl der
// Werte, der größte int plus 1 aus math/big.
func TestI7UngueltigeZahl(t *testing.T) {
	gleichAuf(t, "größter int plus 1", new(big.Int).Add(big.NewInt(math.MaxInt), big.NewInt(1)).String(), inspUngueltigeZahlen()[15])
	l := neuerInspLauf(t, inspE7003())
	faelle := 0
	for _, w := range inspUngueltigeZahlen() {
		for _, param := range []string{"limit", "after"} {
			for _, q := range inspStellungen(param, w) {
				faelle++
				l.abgewiesen(t, q, exchangesZiel("s"), q)
			}
		}
	}
	paare := []string{"limit=-1&after=1.0", "after=1.0&limit=-1", "limit=1.0&after=-1", "after=-1&limit=1.0"}
	for _, q := range paare {
		faelle++
		l.abgewiesen(t, q, exchangesZiel("s"), q)
	}
	l.nurAbweisungen(t, "Zählungen", faelle)
	gleichAuf(t, "Fälle", []int{faelle, len(inspUngueltigeZahlen())}, []int{106, 17})
}

// inspOrakelZahl ist das Orakel der Grammatik von Nr. 4, ohne den Code des Lesers und ohne strconv: der Text besteht aus einer
// oder mehr ASCII-Ziffern, und der Wert liegt im Bereich von int (math/big).
func inspOrakelZahl(text string) (wert int, gueltig bool) {
	if text == "" || strings.Trim(text, "0123456789") != "" {
		return 0, false
	}
	n, _ := new(big.Int).SetString(text, 10)
	if n.Cmp(big.NewInt(math.MaxInt)) > 0 {
		return 0, false
	}
	return int(n.Int64()), true
}

// inspZahlTexte ist die Menge der Zahlen von I6 und I7 aus der Grammatik: jede Ziffer in jeder von fünf Formen (`d`, `1d`, `d1`,
// `0d`, `d0`), die Folge von Neunen der Länge 1 bis 40, die Zehnerpotenz 10^0 bis 10^25, null bis 30 führende Nullen vor
// `7` und vor dem größten int, und der größte int mit den Nachbarn minus 1, plus 1 und plus 10.
func inspZahlTexte() []string {
	var out []string
	for d := range 10 {
		c := strconv.Itoa(d)
		out = append(out, c, "1"+c, c+"1", "0"+c, c+"0")
	}
	for l := 1; l <= 40; l++ {
		out = append(out, strings.Repeat("9", l))
	}
	for l := range 26 {
		out = append(out, "1"+strings.Repeat("0", l))
	}
	for k := range 31 {
		out = append(out, strings.Repeat("0", k)+"7", strings.Repeat("0", k)+strconv.Itoa(math.MaxInt))
	}
	grenze := big.NewInt(math.MaxInt)
	for _, d := range []int64{-1, 0, 1, 10} {
		out = append(out, new(big.Int).Add(grenze, big.NewInt(d)).String())
	}
	return out
}

// TestI6I7ZahlenAusDerGrammatik hält I6 und I7 für die Klasse der Zusage „jede Zahl“: jeder Text aus inspZahlTexte, an `limit`
// und an `after` (je 2 Parameter), ist genau dann gültig, wenn das Orakel ihn als gültig nennt, und der Wert geht unverändert
// an den Dienst; sonst 400 ohne Aufruf. Zählprüfung: die Menge hat 50 + 40 + 26 + 62 + 4 Texte, davon gültig (Orakel) und
// ungültig zusammen gleich der Menge, und beide Seiten sind nicht leer.
func TestI6I7ZahlenAusDerGrammatik(t *testing.T) {
	texte := inspZahlTexte()
	gleichAuf(t, "Zahl der Texte", len(texte), 50+40+26+62+4)
	rein, raus := neuerInspLauf(t, nil), neuerInspLauf(t, inspE7003())
	gueltig, ungueltig := 0, 0
	for _, text := range texte {
		wert, ok := inspOrakelZahl(text)
		for _, param := range []string{"limit", "after"} {
			q := param + "=" + text
			if !ok {
				ungueltig++
				for _, st := range inspStellungen(param, text) {
					raus.abgewiesen(t, st, exchangesZiel("s"), st)
				}
				continue
			}
			gueltig++
			soll := inbound.Seite{Limit: wert}
			if param == "after" {
				soll = inbound.Seite{Nach: wert}
			}
			if a := rein.angenommen(t, q, exchangesZiel("s"), q); a.seite != soll {
				t.Errorf("%s: Seite %+v, erwartet %+v", q, a.seite, soll)
			}
		}
	}
	gleichAuf(t, "gültig plus ungültig gleich Menge × 2", gueltig+ungueltig, 2*len(texte))
	if gueltig < 100 || ungueltig < 20 {
		t.Errorf("gültig %d, ungültig %d: eine Seite der Grenze ist nicht geführt", gueltig, ungueltig)
	}
}

// TestI7ZeichenInDerZahl hält I7 für jedes Zeichen in jeder Stelle (die Klasse „nichts sonst als ASCII-Ziffern“): jedes der 256
// Bytes, als `%XX` kodiert, allein, hinter einer `1`, vor einer `1` und zwischen zwei `1`, an `limit` und an `after` (256 × 4 × 2
// = 2048 Fälle): gültig ist allein eine Ziffer `0` bis `9` (Orakel: die Ziffer, mit dem Wert aus der Stelle), jedes andere Byte ist
// 400 ohne Aufruf; dazu jede Ziffer außerhalb von ASCII (Klasse Nd von Unicode ab U+0080, als UTF-8 kodiert), allein und hinter
// einer `1`, an beiden Parametern: ungültig. Zählprüfung der Menge: 256 Bytes und die Zahl der Nd-Zeichen ab U+0080.
func TestI7ZeichenInDerZahl(t *testing.T) {
	rein, raus := neuerInspLauf(t, nil), neuerInspLauf(t, inspE7003())
	formen := []struct {
		vor, nach string
		wert      func(d int) int
	}{
		{"", "", func(d int) int { return d }}, {"1", "", func(d int) int { return 10 + d }},
		{"", "1", func(d int) int { return 10*d + 1 }}, {"1", "1", func(d int) int { return 100 + 10*d + 1 }},
	}
	faelle, nd := 0, 0
	for b := range 256 {
		for _, f := range formen {
			for _, param := range []string{"limit", "after"} {
				faelle++
				text := f.vor + fmt.Sprintf("%%%02X", b) + f.nach
				q := param + "=" + text
				if b < '0' || b > '9' {
					for _, st := range inspStellungen(param, text) {
						raus.abgewiesen(t, st, exchangesZiel("s"), st)
					}
					continue
				}
				soll := inbound.Seite{Limit: f.wert(b - '0')}
				if param == "after" {
					soll = inbound.Seite{Nach: f.wert(b - '0')}
				}
				if a := rein.angenommen(t, q, exchangesZiel("s"), q); a.seite != soll {
					t.Errorf("%s: Seite %+v, erwartet %+v", q, a.seite, soll)
				}
			}
		}
	}
	for r := rune(0x80); r <= utf8.MaxRune; r++ {
		if !unicode.Is(unicode.Nd, r) {
			continue
		}
		nd++
		kodiert := ""
		for _, b := range []byte(string(r)) {
			kodiert += fmt.Sprintf("%%%02X", b)
		}
		for _, text := range []string{kodiert, "1" + kodiert} {
			for _, param := range []string{"limit", "after"} {
				for _, st := range inspStellungen(param, text) {
					raus.abgewiesen(t, st, exchangesZiel("s"), st)
				}
			}
		}
	}
	gleichAuf(t, "Fälle der Bytes", faelle, 256*4*2)
	if nd < 600 {
		t.Errorf("%d Ziffern außerhalb von ASCII, erwartet mindestens 600 (Klasse Nd)", nd)
	}
}

// TestI8DoppeltUndLeer hält I8 (SPEC-293 Nr. 3): ein bekannter Schlüssel zweimal oder mit leerem Wert ist 400 ohne Aufruf, ein
// unbekannter doppelter nicht. 3 bekannte Schlüssel (`limit`, `after`, `baseline_id`) × 6 Formen (`k=1&k=2`, `k=1&k=1`,
// `k=1&k=`, `k=&k=1`, `k=1&k`, `k=1&x=2&k=3`) = 18 Fälle, leer (`k=`, `k`) 3 × 2 = 6, `x=1&x=2` an beiden Routen = 2 (200); jede
// Form auch neben der gültigen Query der Route (hinter und vor): 24 × 3 = 72 Abweisungen. Der Dienst würde `ODR-E7003` oder
// `ODR-E7002` liefern.
func TestI8DoppeltUndLeer(t *testing.T) {
	formen := []string{"k=1&k=2", "k=1&k=1", "k=1&k=", "k=&k=1", "k=1&k", "k=1&x=2&k=3", "k=", "k"}
	gleichAuf(t, "Formen", len(formen), 6+2)
	abweisungen, angenommen := 0, 0
	for _, r := range inspRoutenMitQuery() {
		l := neuerInspLauf(t, r.fehler)
		schluessel := []string{"limit", "after"}
		if r.name == "coverage" {
			schluessel = []string{"baseline_id"}
		}
		for _, k := range schluessel {
			for _, f := range formen {
				for _, q := range inspStellen(r.gueltig, strings.ReplaceAll(f, "k", k)) {
					abweisungen++
					l.abgewiesen(t, r.name+" "+q, r.pfad, q)
				}
			}
		}
		neuerInspLauf(t, nil).angenommen(t, r.name+" x=1&x=2", r.pfad, "x=1&x=2&"+r.gueltig)
		angenommen++
	}
	gleichAuf(t, "Abweisungen und Fälle mit unbekanntem doppelten Schlüssel", []int{abweisungen, angenommen}, []int{(2 + 1) * 8 * 3, 2})
}

// inspBaselineWerte sind die 9 Werte von M-baseline (SPEC-293 Nr. 5) mit den dekodierten Bytes als Literale aus Bytewerten.
func inspBaselineWerte() []struct{ query, soll string } {
	b := func(x ...byte) string { return string(x) }
	return []struct{ query, soll string }{
		{"x", "x"}, {"a%20b", "a b"}, {"a+b", "a b"}, {"%00", b(0)}, {"%FF", b(0xFF)}, {"a%2Fb", "a/b"}, {"A", "A"}, {"%20", " "},
		{strings.Repeat("a", 300), strings.Repeat("a", 300)},
	}
}

// TestI9Baseline hält I9 (SPEC-293 Nr. 5): `baseline_id` fehlt ist 400 ohne Aufruf (auch wenn der Dienst `ODR-E7002` lieferte), jeder
// andere dekodierte Wert geht byte-gleich an Coverage (9 Werte und fehlt = 10 Fälle); dazu die Klasse „jedes Byte“: jedes der
// 256 Bytes als `%XX` kodiert, an vier Stellen (allein, zwischen zwei Buchstaben, am Ende, am Anfang: 1024 Fälle), ein `;` als `%3B` (ein rohes `;` ist ein
// Verstoß gegen die Zerlegung), der Schlüssel in Prozentschreibweise, und der Schlüssel in anderer Schreibweise ist unbekannt
// (der Wert fehlt: 400).
func TestI9Baseline(t *testing.T) {
	rein, raus := neuerInspLauf(t, nil), neuerInspLauf(t, inspE7002())
	werte := inspBaselineWerte()
	for _, w := range werte {
		if a := rein.angenommen(t, w.query, inspCoverage, "baseline_id="+w.query); a.op != inspOpCoverage || a.id != w.soll {
			t.Errorf("baseline_id=%s: Aufruf %d %q, erwartet Coverage mit %q", w.query, a.op, a.id, w.soll)
		}
	}
	raus.abgewiesen(t, "fehlt", inspCoverage, "")
	for _, q := range []string{"x=1", "Baseline_Id=x", "BASELINE_ID=x", "baseline-id=x", "baseline_id=", "baseline_id", "baseline_id=&x=1"} {
		raus.abgewiesen(t, q, inspCoverage, q)
	}
	faelle := 0
	for b := range 256 {
		for _, form := range []string{"%s", "a%sz", "a%s", "%sz"} {
			faelle++
			kodiert := strings.Replace(form, "%s", fmt.Sprintf("%%%02X", b), 1)
			soll := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
			if a := rein.angenommen(t, kodiert, inspCoverage, "baseline_id="+kodiert); a.id != soll {
				t.Errorf("Byte 0x%02X in der Form %q: Aufruf mit %q, erwartet %q", b, form, a.id, soll)
			}
		}
	}
	if a := rein.angenommen(t, "%3B", inspCoverage, "baseline_id=a%3Bb"); a.id != "a;b" {
		t.Errorf("a%%3Bb: Aufruf mit %q", a.id)
	}
	raus.abgewiesen(t, "rohes Semikolon", inspCoverage, "baseline_id=a;b")
	if a := rein.angenommen(t, "%62aseline_id", inspCoverage, "%62aseline%5Fid=zz"); a.id != "zz" {
		t.Errorf("Schlüssel in Prozentschreibweise: Aufruf mit %q", a.id)
	}
	gleichAuf(t, "Werte, Fälle je Byte", []int{len(werte), faelle}, []int{9, 1024})
}
