package admin_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// importForm bildet den Request von `POST /__admin/imports` zu einer Form des Bodys.
func importForm(b bForm) *http.Request {
	r := importPost("", b.kopf, b.leser())
	if b.laenge != 0 {
		r.ContentLength = b.laenge
	}
	return r
}

// importFelder sind die sechs Felder des Bodys, die einen Wert tragen: fünf Zeichenketten und ein Eintrag von
// `source_session_ids` (SPEC-297 Nr. 5 d).
func importFelder() []string {
	return []string{"import_id", "service_id", "source_session_ids", "metadata_hash", "conflict_policy", "target"}
}

// importFeld liefert zu einem Feld, dem Text einer Zeichenkette im Body (text, wie er im JSON steht) und dem dekodierten Wert das
// Literal des Gliedes im Body und die Anfrage, die der Dienst erhalten soll: gutAnfrage mit dem Wert im Feld (bei
// `source_session_ids` als einziger Eintrag).
func importFeld(feld, text, wert string) (literal string, soll inbound.ImportAnfrage) {
	soll, literal = gutAnfrage(), `"`+text+`"`
	switch feld {
	case "import_id":
		soll.ImportID = wert
	case "service_id":
		soll.ServiceID = wert
	case "source_session_ids":
		soll.SourceSessionIDs, literal = []string{wert}, `["`+text+`"]`
	case "metadata_hash":
		soll.MetadataHash = wert
	case "conflict_policy":
		soll.ConflictPolicy = wert
	case "target":
		soll.Target = wert
	}
	return literal, soll
}

// importFeldBody bildet den Body, in dem das Feld den Text einer Zeichenkette trägt, und die Anfrage, die der Dienst mit dem
// dekodierten Wert erhalten soll.
func importFeldBody(feld string, text, wert []byte) (body []byte, soll inbound.ImportAnfrage) {
	literal, soll := importFeld(feld, string(text), string(wert))
	return []byte(bodyMitGlied("imp-1", feld, literal)), soll
}

// iGueltig ist eine gültige Form des Bodys mit der Anfrage, die der Dienst erhalten soll.
type iGueltig struct {
	bForm
	soll inbound.ImportAnfrage
}

// importGueltig sind die 13 gültigen Formen von M-Body gültig (SPEC-297 Nr. 5 a bis g): 10 Formen mit `Content-Type:
// application/json` und 3 weitere `Content-Type` (keiner, `text/plain`, `application/x-www-form-urlencoded`) zur Form
// `nur Pflicht`; das Orakel sind die Argumente von Import als Literale.
func importGueltig() []iGueltig {
	ae, euro, smiley, ss := nichtASCII()
	json := map[string][]string{"Content-Type": {"application/json"}}
	voll := inbound.ImportAnfrage{
		ImportID: "imp-1", ServiceID: "svc-1", SourceSessionIDs: []string{"ses-1"}, MetadataHash: "sha256:h",
		ConflictPolicy: "reject", Target: "new_baseline",
	}
	umgekehrt := slices.Clone(alleGlieder("imp-1"))
	slices.Reverse(umgekehrt)
	formen := []iGueltig{
		{bForm{name: "alle sechs", body: []byte(objektAus(alleGlieder("imp-1")...))}, voll},
		{bForm{name: "nur Pflicht", body: []byte(gutBody)}, gutAnfrage()},
		{bForm{name: "umgekehrte Folge", body: []byte(objektAus(umgekehrt...))}, voll},
		{bForm{name: "optionale leer", body: []byte(objektAus(append(pflichtGlieder("imp-1"), `"metadata_hash":""`, `"conflict_policy":""`, `"target":""`)...))}, gutAnfrage()},
		{bForm{
			name: "Leerraum",
			body: []byte("\r\n\t " + `{ "import_id" : " imp-1 " ,` + "\n" + ` "service_id" : " svc-1 " , "source_session_ids" : [ "b" , " ses-1 " ],` +
				` "metadata_hash" : " sha256:h " , "conflict_policy" : " reject " , "target" : " new_baseline " }` + " \t\r\n"),
		}, inbound.ImportAnfrage{
			ImportID: " imp-1 ", ServiceID: " svc-1 ", SourceSessionIDs: []string{"b", " ses-1 "}, MetadataHash: " sha256:h ",
			ConflictPolicy: " reject ", Target: " new_baseline ",
		}},
		{bForm{
			name: "Escapes und Nicht-ASCII",
			body: []byte(`{"import_id":"a` + esc("0076") + rueck() + `nb","service_id":"` + ae + euro + smiley + `","source_session_ids":["` + ss + `"]}`),
		}, inbound.ImportAnfrage{ImportID: "av\nb", ServiceID: ae + euro + smiley, SourceSessionIDs: []string{ss}}},
		{bForm{name: "leere Liste", body: []byte(`{"import_id":"imp-1","service_id":"svc-1","source_session_ids":[]}`)},
			inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: "svc-1", SourceSessionIDs: []string{}}},
		{bForm{name: "gleiche Einträge", body: []byte(`{"import_id":"imp-1","service_id":"svc-1","source_session_ids":["a","a"]}`)},
			inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: "svc-1", SourceSessionIDs: []string{"a", "a"}}},
		{bForm{name: "leerer Eintrag", body: []byte(`{"import_id":"imp-1","service_id":"svc-1","source_session_ids":[""]}`)},
			inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: "svc-1", SourceSessionIDs: []string{""}}},
		{bForm{name: "1048576 Bytes", body: auffuellen(gutBody, bodyGrenze)}, gutAnfrage()},
	}
	for i := range formen {
		formen[i].achse, formen[i].kopf = "gültig", json
	}
	for _, ct := range [][]string{nil, {"text/plain"}, {"application/x-www-form-urlencoded"}} {
		f := iGueltig{bForm{achse: "Content-Type", name: "Content-Type " + strings.Join(ct, ""), body: []byte(gutBody)}, gutAnfrage()}
		if ct != nil {
			f.kopf = map[string][]string{"Content-Type": ct}
		}
		formen = append(formen, f)
	}
	return formen
}

// importZeichen sind die 3 Zeichen der Achse *Zeichen im Wert* (SPEC-297 Nr. 5 e) mit dem Wert, den der Dienst erhält: U+FEFF
// roh, der Escape eines einzelnen hohen und der eines einzelnen tiefen Surrogats (beide als U+FFFD); alles aus Bytewerten.
func importZeichen() []struct{ name, roh, soll string } {
	bomZeichen, ersatz := string([]byte{0xEF, 0xBB, 0xBF}), string([]byte{0xEF, 0xBF, 0xBD})
	return []struct{ name, roh, soll string }{
		{"U+FEFF", "s" + bomZeichen + "x", "s" + bomZeichen + "x"},
		{"hohes Surrogat", "s" + esc("d800") + "x", "s" + ersatz + "x"},
		{"tiefes Surrogat", "s" + esc("dc00") + "x", "s" + ersatz + "x"},
	}
}

// TestIM6GueltigerBody hält IM6 (SPEC-297 Nr. 5 a bis g): jede der 13 gültigen Formen ergibt 201 und genau einen Aufruf von
// Import mit den Literalen des Orakels (jedes der sechs Felder hat in der ersten Form einen eigenen Wert).
func TestIM6GueltigerBody(t *testing.T) {
	formen := importGueltig()
	gleichAuf(t, "gültige Formen", len(formen), 13)
	for _, f := range formen {
		u, fake, _ := importsUmgebung(t)
		rec := u.rufe(importForm(f.bForm))
		if rec.Code != 201 || rec.Body.String() != reportA {
			t.Errorf("%s: Antwort %d %q, erwartet 201 und den Report", f.name, rec.Code, rec.Body.String())
		}
		fake.nurDiagonale(t, f.name, impImport)
		if !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{f.soll}) {
			t.Errorf("%s: Argumente von Import %+v, erwartet %+v", f.name, fake.anfrageListe(), f.soll)
		}
	}
}

// TestIM6ZeichenImWert hält IM6 für die Achse *Zeichen im Wert* (SPEC-297 Nr. 5 e): 3 Zeichen × 6 Felder = 18 Fälle, je 201
// und der Wert unverändert (ein Surrogat-Escape als U+FFFD) am Dienst. Zählprüfung: 13 + 18 = 31 mit TestIM6GueltigerBody.
func TestIM6ZeichenImWert(t *testing.T) {
	faelle := 0
	for _, z := range importZeichen() {
		for _, feld := range importFelder() {
			faelle++
			u, fake, _ := importsUmgebung(t)
			literal, soll := importFeld(feld, z.soll, z.soll)
			roh, _ := importFeld(feld, z.roh, z.soll)
			rec := u.rufe(importPost("", nil, strings.NewReader(bodyMitGlied("imp-1", feld, roh))))
			if rec.Code != 201 || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}) {
				t.Errorf("%s in %s: Status %d, Argumente %+v, erwartet 201 und %s", z.name, feld, rec.Code, fake.anfrageListe(), literal)
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 3*len(importFelder()))
	gleichAuf(t, "Felder gleich den Feldern von ImportAnfrage", len(importFelder()), reflect.TypeFor[inbound.ImportAnfrage]().NumField())
	gleichAuf(t, "Fälle gültiger Bodys laut Plan", faelle+len(importGueltig()), 31)
}

// pruefeImportUngueltig hält eine ungültige Form gegen IM7 und IM8: 400 mit dem Fehlerobjekt von ODR-E2102, genau ein
// Zählen, weder Import noch Get noch Append, kein Wert aus dem Body in der Antwort; der Handler allein liefert
// model.ErrUngueltig ohne Wert im Text und keine Antwort. Der Dienst würde ODR-E5005 liefern (Zähler 0: der Body zuerst).
func pruefeImportUngueltig(t *testing.T, f bForm) {
	t.Helper()
	name := f.achse + " " + f.name
	u, fake, routen := importsUmgebung(t)
	fake.err = &inbound.Fehler{Code: model.CodeE5005, Status: 404}
	rec := u.rufe(importForm(f))
	pruefeFehlerAntwort(t, name, rec, 400, objekt400, "")
	if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 {
		t.Errorf("%s: CountError %q, andere Aufrufe %d", name, codes, andere)
	}
	if fake.gesamt() != 0 || u.audit.Aufrufe() != 0 {
		t.Errorf("%s: %d Aufrufe des Dienstes, %d von Audit, erwartet 0 und 0", name, fake.gesamt(), u.audit.Aufrufe())
	}
	begrenzt := http.MaxBytesReader(nil, io.NopCloser(f.leser()), bodyGrenze)
	ant, err := routen[0].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: begrenzt})
	if !errors.Is(err, model.ErrUngueltig) || strings.Contains(err.Error(), canary()) || ant.Status != 0 || len(ant.Body) != 0 {
		t.Errorf("%s: Handler liefert %q, %v, erwartet model.ErrUngueltig ohne Wert und keine Antwort", name, ant.Body, err)
	}
	if fake.gesamt() != 0 {
		t.Errorf("%s: der Handler rief den Dienst", name)
	}
}

// TestIM7UngueltigForm hält IM7 (SPEC-297 Nr. 5 b, c, h): jede der 28 Formen der Achsen Wertform, Struktur, Schlüssel und
// Zeichen ist 400 `ODR-E2102` ohne Aufruf des Dienstes.
func TestIM7UngueltigForm(t *testing.T) {
	for _, f := range importUngueltigForm(canary()) {
		pruefeImportUngueltig(t, f)
	}
}

// TestIM7Rand hält IM7 für den Rand des Objekts: jeder Bytewert 0 bis 255 vor und hinter einem gültigen Objekt (512 Fälle);
// gültig sind genau Leerzeichen, Tabulator, LF und CR (4 × 2 Fälle, 201 und die Anfrage unverändert), jedes andere Byte (252
// × 2) ist 400 ohne Aufruf des Dienstes.
func TestIM7Rand(t *testing.T) {
	gueltig, ungueltig := 0, 0
	for _, f := range importRand() {
		u, fake, _ := importsUmgebung(t)
		rec := u.rufe(importForm(f))
		b := f.body[0]
		if strings.HasSuffix(f.name, "dahinter") {
			b = f.body[len(f.body)-1]
		}
		ws := b == 0x20 || b == 0x09 || b == 0x0a || b == 0x0d
		switch {
		case ws && rec.Code == 201 && gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{gutAnfrage()}):
			gueltig++
		case !ws && rec.Code == 400 && fake.gesamt() == 0:
			ungueltig++
		default:
			t.Errorf("%s: Status %d, %d Aufrufe von Import, Leerraum %v", f.name, rec.Code, fake.zahlen()[impImport], ws)
		}
	}
	gleichAuf(t, "gültige und ungültige Fälle", []int{gueltig, ungueltig}, []int{4 * 2, 252 * 2})
}

// TestIM7ArrayMitPaaren hält SPEC-297 Nr. 5 b für Arrays, deren Elemente sich wie die Paare eines Objekts lesen ließen (die
// Klasse „ein Array“ hinter der Form `[]`; 3 Zusatzformen, die der Plan nicht nennt): jede ist 400 ohne Aufruf des Dienstes.
func TestIM7ArrayMitPaaren(t *testing.T) {
	c := canary()
	formen := []string{
		`["import_id","` + c + `","service_id","svc-1","source_session_ids",["ses-1"]]`,
		`["service_id","svc-1","import_id","` + c + `","source_session_ids",["ses-1"],"target","new_baseline"]`,
		`["source_session_ids",["ses-1"],"import_id","` + c + `","service_id","svc-1"]`,
	}
	for i, body := range formen {
		pruefeImportUngueltig(t, bForm{achse: "Array", name: strconv.Itoa(i), body: []byte(body)})
	}
	gleichAuf(t, "Zahl der Formen", len(formen), 3)
}

// TestIM8UngueltigRest hält IM8 (SPEC-297 Nr. 5 a, d, h): jede der 58 Formen der Achsen Pflicht, Werte, Lesen und Paare ist
// 400 `ODR-E2102` ohne Aufruf des Dienstes; ein Paar zweier Verstoßarten im selben Body ist dieselbe Antwort.
func TestIM8UngueltigRest(t *testing.T) {
	for _, f := range importUngueltigRest(canary()) {
		pruefeImportUngueltig(t, f)
	}
}

// TestIM7IM8Zahlen hält M-Body ungültig gegen die Quelle des Tests: 28 + 58 = 86 Formen, je Achse die Zahl des Plans (Werte 37
// gleich 10 + 5 + 7 + 15, Paare 15 gleich 6 über 2), die Namen je Achse und die Bodies einer Achse sind verschieden, der Rand
// führt 512 Formen. Die Funktion `importBodyUngueltig` liefert dieselben 86 Formen, die Teil 2 liest.
func TestIM7IM8Zahlen(t *testing.T) {
	a, b := importUngueltigForm("x"), importUngueltigRest("x")
	gleichAuf(t, "Formen von IM7 und IM8", []int{len(a), len(b), len(importBodyUngueltig("x")), len(importRand())}, []int{28, 58, 86, 512})
	gleichAuf(t, "Achsen von IM7", zaehleAchsen(a), map[string]int{"Wertform": 7, "Struktur": 7, "Schlüssel": 11, "Zeichen": 3})
	gleichAuf(t, "Achsen von IM8", zaehleAchsen(b), map[string]int{"Pflicht": 4, "Werte": 37, "Lesen": 2, "Paar": 15})
	gleichAuf(t, "Paare gleich 6 über 2 und Werte gleich 10 + 5 + 7 + 15", []int{6 * 5 / 2, 10 + 5 + 7 + 15}, []int{15, 37})
	namen, bodies := map[string]bool{}, map[string]bool{}
	for _, f := range importBodyUngueltig("x") {
		namen[f.achse+"/"+f.name] = true
		bodies[string(f.body)+"|"+strconv.FormatInt(f.laenge, 10)] = true
	}
	gleichAuf(t, "verschiedene Namen und Bodies", []int{len(namen), len(bodies)}, []int{86, 86})
}

// wertFeldFaelle prüft einen Wert in jedem der sechs Felder gegen den Dienst: text ist der Text der Zeichenkette im Body, wert
// der dekodierte Wert; gueltig heißt 201 und die Anfrage mit dem Wert unverändert, sonst 400 ohne Aufruf.
func wertFeldFaelle(t *testing.T, name string, text, wert []byte, gueltig bool) {
	t.Helper()
	for _, feld := range importFelder() {
		u, fake, _ := importsUmgebung(t)
		body, soll := importFeldBody(feld, text, wert)
		rec := u.rufe(importPost("", nil, bytesLeser(body)))
		if gueltig && (rec.Code != 201 || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll})) {
			t.Errorf("%s in %s: Status %d, Argumente gleich dem Wert: %v", name, feld, rec.Code, gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}))
		}
		if !gueltig && (rec.Code != 400 || fake.gesamt() != 0) {
			t.Errorf("%s in %s: Status %d, %d Aufrufe, erwartet 400 und 0", name, feld, rec.Code, fake.gesamt())
		}
	}
}

// bytesLeser liefert einen Body aus den Bytes.
func bytesLeser(b []byte) io.Reader { return strings.NewReader(string(b)) }

// TestIM9Laenge hält IM9 (SPEC-297 Nr. 5 e, „unverändert und ungekürzt“) für die Menge Länge × Feld × Zeichenklasse: die
// Klassen klein und Mehrbyte mit jeder Länge von importLaengen(1, 1048000) (Zweier- und Zehnerpotenzen um 1 bis nahe der Grenze
// des Bodys), die fünf anderen Klassen mit den 12 Längen von laengenDerWerte, je × 6 Felder, je 201 und der Wert byte-gleich am
// Dienst. Prüfgrenze (akzeptiertes Negativ (1)): die Längen dazwischen und die Mischungen der Klassen in einem Wert.
func TestIM9Laenge(t *testing.T) {
	faelle, soll := 0, 0
	for _, klasse := range klassenDerWerte() {
		laengen := laengenDerWerte()
		if klasse == "klein" || klasse == "mehrbyte" {
			laengen = importLaengen(1, 1048000)
		}
		soll += len(laengen) * len(importFelder())
		for _, n := range laengen {
			wert := wertDerLaenge(klasse, n)
			if len(wert) != n {
				t.Fatalf("Orakel: Klasse %s Länge %d liefert %d Bytes", klasse, n, len(wert))
			}
			wertFeldFaelle(t, klasse+" "+strconv.Itoa(n), []byte(wert), []byte(wert), true)
			faelle += len(importFelder())
		}
	}
	gleichAuf(t, "Fälle", faelle, soll)
	gleichAuf(t, "Fälle laut Plan", faelle, (2*74+5*12)*6)
}

// zeichenFolge ist ein Element von M-Zeichen: text steht im Text einer Zeichenkette des Bodys, wert ist der dekodierte Wert (das
// Orakel, als Literal); gueltig sagt, ob der Body gilt.
type zeichenFolge struct {
	name, text, wert string
	gueltig          bool
}

// importZeichenMenge sind die 303 Elemente von M-Zeichen (IM9, IM12): jedes der 128 Zeichen U+0000 bis U+007F als `\u`-Escape (128),
// roh die 94 ohne Anführungszeichen und Backslash (94, gültig) und die 32 unter 0x20 (ungültig), die 8 Kurz-Escapes (`\"`,
// `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`), 12 Zeichen jenseits von ASCII roh und als Escape (bei Zeichen jenseits der
// Grundebene als Paar von Surrogaten; 24), 5 Escapes einzelner Surrogate (als U+FFFD), 6 ungültige rohe Folgen und 6
// ungültige Escapes; alle Bytes aus Bytewerten oder `\x`-Folgen.
func importZeichenMenge() []zeichenFolge {
	var out []zeichenFolge
	add := func(name, text, wert string, gueltig bool) {
		out = append(out, zeichenFolge{name, text, wert, gueltig})
	}
	for c := range 0x80 {
		b, hex := string([]byte{byte(c)}), fmt.Sprintf("%04x", c)
		switch {
		case c < 0x20:
			add("roh "+hex, b, "", false)
		case c != '"' && c != 0x5c:
			add("roh "+hex, b, b, true)
		}
		add("Escape "+hex, esc(hex), b, true)
	}
	for _, k := range [][2]string{{`"`, `"`}, {rueck(), rueck()}, {"/", "/"}, {"b", "\b"}, {"f", "\f"}, {"n", "\n"}, {"r", "\r"}, {"t", "\t"}} {
		add("Kurz-Escape "+k[0], rueck()+k[0], k[1], true)
	}
	for _, z := range []struct{ name, utf8, escape string }{
		{"U+0080", "\xc2\x80", esc("0080")}, {"NBSP", "\xc2\xa0", esc("00a0")}, {"U+07FF", "\xdf\xbf", esc("07ff")},
		{"U+0800", "\xe0\xa0\x80", esc("0800")}, {"U+2028", "\xe2\x80\xa8", esc("2028")}, {"U+2029", "\xe2\x80\xa9", esc("2029")},
		{"U+FEFF", "\xef\xbb\xbf", esc("feff")}, {"U+FFFD", "\xef\xbf\xbd", esc("fffd")}, {"U+FFFF", "\xef\xbf\xbf", esc("ffff")},
		{"U+10000", "\xf0\x90\x80\x80", esc("d800") + esc("dc00")}, {"U+1F600", "\xf0\x9f\x98\x80", esc("d83d") + esc("de00")},
		{"U+10FFFF", "\xf4\x8f\xbf\xbf", esc("dbff") + esc("dfff")},
	} {
		add("roh "+z.name, z.utf8, z.utf8, true)
		add("Escape "+z.name, z.escape, z.utf8, true)
	}
	ersatz := "\xef\xbf\xbd"
	for _, e := range [][2]string{{esc("d800"), ersatz}, {esc("dbff"), ersatz}, {esc("dc00"), ersatz}, {esc("dfff"), ersatz}, {esc("dc00") + esc("d800"), ersatz + ersatz}} {
		add("einzelnes Surrogat "+e[0], e[0], e[1], true)
	}
	for _, f := range []string{"\x80", "\xc0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xff", "\xc3"} {
		add("ungültige Folge "+strconv.Quote(f), f, "", false)
	}
	for _, e := range []string{rueck() + "a", rueck() + "x41", esc("12"), esc("12g4"), rueck() + "U0041", rueck() + "0"} {
		add("ungültiger Escape "+e, e, "", false)
	}
	return out
}

// importZeichenStellen sind die drei Stellen von M-Zeichen: das Element am Anfang, in der Mitte und am Ende eines Wertes `s`.
func importZeichenStellen() [][2]string { return [][2]string{{"", "s"}, {"s", "s"}, {"s", ""}} }

// TestIM9Zeichen hält IM9 für die Menge Zeichen × Darstellung × Stelle × Feld: die 303 Elemente von importZeichenMenge (roh,
// als `\u`-Escape, als Kurz-Escape) × 3 Stellen (Anfang, Mitte, Ende des Wertes) × 6 Felder = 5454 Fälle; die 259 gültigen
// gehen als der Wert des Orakels (dekodiert, unverändert) an den Dienst, die 44 ungültigen sind 400 ohne Aufruf.
func TestIM9Zeichen(t *testing.T) {
	menge, gueltig := importZeichenMenge(), 0
	for _, z := range menge {
		if z.gueltig {
			gueltig++
		}
		for _, st := range importZeichenStellen() {
			wertFeldFaelle(t, z.name+" an Stelle "+st[0]+"|"+st[1], []byte(st[0]+z.text+st[1]), []byte(st[0]+z.wert+st[1]), z.gueltig)
		}
	}
	gleichAuf(t, "gültige und ungültige Elemente", []int{gueltig, len(menge) - gueltig, len(menge)}, []int{259, 44, 303})
	gleichAuf(t, "Fälle laut Plan", len(menge)*len(importZeichenStellen())*len(importFelder()), 5454)
}

// TestIM6Wertformen hält IM6 und IM8 für die Raster-Zelle *leerer Wert* und ihre Nachbarn (SPEC-297 Nr. 5 d, e), je Feld der
// sechs: die Wertformen fehlend, leer, Leerzeichen, vier Zeichen Leerraum, Randleerraum, NBSP und NBSP am Rand gehen unverändert
// an den Dienst (201), außer dem fehlenden Pflichtschlüssel (400 ohne Aufruf; ein fehlender optionaler ist leer); `null` ist
// bei jedem Feld 400. Die Route prüft keinen Wert: auch die leere `import_id` geht an den Dienst. 8 Formen × 6 Felder = 48.
func TestIM6Wertformen(t *testing.T) {
	nbsp := "\xc2\xa0"
	formen := []struct{ name, text, wert string }{
		{"leer", "", ""}, {"Leerzeichen", " ", " "}, {"vier Zeichen Leerraum", " " + rueck() + "t" + rueck() + "r" + rueck() + "n", " \t\r\n"},
		{"Randleerraum", " x ", " x "}, {"NBSP", nbsp, nbsp}, {"NBSP am Rand", nbsp + "x" + nbsp, nbsp + "x" + nbsp},
	}
	for _, f := range formen {
		wertFeldFaelle(t, f.name, []byte(f.text), []byte(f.wert), true)
	}
	faelle := len(formen) * len(importFelder())
	for _, feld := range importFelder() {
		var ohne []string
		for _, g := range pflichtGlieder("imp-1") {
			if !strings.HasPrefix(g, `"`+feld+`"`) {
				ohne = append(ohne, g)
			}
		}
		pflicht := len(ohne) < 3
		for _, f := range []struct{ name, body string }{{"fehlend", objektAus(ohne...)}, {"null", bodyMitGlied("imp-1", feld, "null")}} {
			u, fake, _ := importsUmgebung(t)
			rec := u.rufe(importPost("", nil, strings.NewReader(f.body)))
			faelle++
			ok := rec.Code == 400 && fake.gesamt() == 0
			if f.name == "fehlend" && !pflicht {
				ok = rec.Code == 201 && gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{gutAnfrage()})
			}
			if !ok {
				t.Errorf("%s %s: Status %d, %d Aufrufe, Argumente %+v", f.name, feld, rec.Code, fake.gesamt(), fake.anfrageListe())
			}
		}
	}
	gleichAuf(t, "Fälle laut Plan", faelle, 8*6)
}

// TestIM9Stelle hält IM9 für die Stelle eines ungültigen Bytes: 0xFF an jeder Stelle p von importLaengen(0, 1048000) (Abstand
// vom Anfang des Wertes: 0 und jede Zweier- und Zehnerpotenz um 1 bis 1048000), in jedem der 6 Felder, ist 400 ohne Aufruf des
// Dienstes. Zählprüfung: Stellen × 6 Felder.
func TestIM9Stelle(t *testing.T) {
	stellen := importLaengen(0, 1048000)
	for _, p := range stellen {
		w := []byte(strings.Repeat("a", p) + string([]byte{0xFF}) + "bb")
		wertFeldFaelle(t, "0xFF an Stelle "+strconv.Itoa(p), w, w, false)
	}
	gleichAuf(t, "Fälle laut Plan", len(stellen)*len(importFelder()), 75*6)
}

// TestIM9Liste hält IM9 für die Zahl der Einträge von `source_session_ids` (SPEC-297 Nr. 5 e): jede Zahl von
// importLaengen(0, bis) (0 und jede Zweier- und Zehnerpotenz um 1 bis zur größten Zahl, die in einen Body von 1 MiB passt) geht
// mit allen Einträgen in ihrer Folge an den Dienst; zwei Muster der Einträge, `"", "b", "a"` (leere, gleiche, absteigend
// geordnete: kürzen, entdoppeln, sortieren und Auslassen ändern die Liste) und lauter `""` (das Muster mit den meisten
// Einträgen je Byte). Zählprüfung: Zahlen × 2 Muster. Prüfgrenze (akzeptiertes Negativ (1)): die Zahlen dazwischen.
func TestIM9Liste(t *testing.T) {
	faelle := 0
	for _, muster := range [][]string{{"", "b", "a"}, {""}} {
		bytesJeMuster := 0
		for _, e := range muster {
			bytesJeMuster += len(e) + 3
		}
		for _, n := range importLaengen(0, (bodyGrenze-200)*len(muster)/bytesJeMuster) {
			liste := make([]string, n)
			var literal strings.Builder
			literal.WriteString("[")
			for i := range liste {
				liste[i] = muster[i%len(muster)]
				if i > 0 {
					literal.WriteString(",")
				}
				literal.WriteString(`"` + liste[i] + `"`)
			}
			literal.WriteString("]")
			u, fake, _ := importsUmgebung(t)
			rec := u.rufe(importPost("", nil, strings.NewReader(bodyMitGlied("imp-1", "source_session_ids", literal.String()))))
			soll := gutAnfrage()
			soll.SourceSessionIDs = liste
			faelle++
			if rec.Code != 201 || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}) {
				t.Errorf("Muster %q, %d Einträge: Status %d, Argumente gleich: %v", muster, n, rec.Code, gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}))
			}
		}
	}
	gleichAuf(t, "Fälle laut Plan", faelle, 69*2)
}

// TestIM8ElementStelle hält SPEC-297 Nr. 5 d für die Stelle eines Elements von `source_session_ids`, das keine Zeichenkette ist:
// jede der 5 Arten (`null`, `1`, `true`, `[]`, `{}`) an jeder Stelle p von importLaengen(0, 100000), hinter p und vor einem
// weiteren Eintrag, ist 400 ohne Aufruf des Dienstes. Zählprüfung: Stellen × 5 Arten.
func TestIM8ElementStelle(t *testing.T) {
	for _, f := range importElementeUngueltig(canary()) {
		pruefeImportUngueltig(t, f)
	}
	gleichAuf(t, "Stellen und Formen", []int{len(importLaengen(0, 100000)), len(importElementeUngueltig("x"))}, []int{61, 5 * 61})
}

// TestIM7RandUnicode hält IM7 für den Rand des Objekts über Einzelbytes hinaus: jede der 24 Formen von importRandUnicode
// (12 Zeichen, die wie Leerraum aussehen, vor und hinter dem Objekt) ist 400 ohne Aufruf des Dienstes.
func TestIM7RandUnicode(t *testing.T) {
	for _, f := range importRandUnicode() {
		pruefeImportUngueltig(t, f)
	}
	gleichAuf(t, "Formen", len(importRandUnicode()), 24)
}

// TestIM9Dienst hält IM9 für `ErrUngueltig` des Dienstes (SPEC-297 Nr. 5 e, 7): der Fake liefert model.ErrUngueltig für 10
// Werte, die SPEC-127 Nr. 3 abweist; die Antwort ist 400 `ODR-E2102`, genau 1 Aufruf mit den Werten unverändert, ein Zählen
// und kein Eintrag der Route.
func TestIM9Dienst(t *testing.T) {
	faelle := []struct {
		name, feld, literal string
		aendere             func(a *inbound.ImportAnfrage)
	}{
		{"merge", "conflict_policy", `"merge"`, func(a *inbound.ImportAnfrage) { a.ConflictPolicy = "merge" }},
		{"target x", "target", `"x"`, func(a *inbound.ImportAnfrage) { a.Target = "x" }},
		{"Hash x", "metadata_hash", `"x"`, func(a *inbound.ImportAnfrage) { a.MetadataHash = "x" }},
		{"import_id mit 65 Bytes", "import_id", `"` + strings.Repeat("a", 65) + `"`, func(a *inbound.ImportAnfrage) { a.ImportID = strings.Repeat("a", 65) }},
		{"import_id A", "import_id", `"A"`, func(a *inbound.ImportAnfrage) { a.ImportID = "A" }},
		{"import_id _", "import_id", `"_"`, func(a *inbound.ImportAnfrage) { a.ImportID = "_" }},
		{"service_id leer", "service_id", `""`, func(a *inbound.ImportAnfrage) { a.ServiceID = "" }},
		{"keine Sessions", "source_session_ids", `[]`, func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = []string{} }},
		{"zwei Sessions", "source_session_ids", `["a","b"]`, func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = []string{"a", "b"} }},
		{"leere Session", "source_session_ids", `[""]`, func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = []string{""} }},
	}
	for _, f := range faelle {
		u, fake, _ := importsUmgebung(t)
		fake.err = model.ErrUngueltig
		rec := u.rufe(importPost("", nil, strings.NewReader(bodyMitGlied("imp-1", f.feld, f.literal))))
		pruefeFehlerAntwort(t, f.name, rec, 400, objekt400, "")
		soll := gutAnfrage()
		f.aendere(&soll)
		if !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}) || fake.zahlen() != [impAnzahl]int{1, 0} {
			t.Errorf("%s: Argumente %+v, Zähler %v, erwartet %+v und genau 1 Aufruf von Import", f.name, fake.anfrageListe(), fake.zahlen(), soll)
		}
		if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 || u.audit.Aufrufe() != 0 {
			t.Errorf("%s: CountError %q, andere %d, Audit %d", f.name, codes, andere, u.audit.Aufrufe())
		}
	}
	gleichAuf(t, "Fälle laut Plan", len(faelle), 10)
}

// TestIM9Zahlen hält M-Länge und M-Wert gegen die Quelle des Tests: die Längen von importLaengen steigen echt, enden an der
// Grenze und führen zu jeder Schwelle, die ein Leser setzen könnte (Zweier- und Zehnerpotenzen und die 255 unter 256), die Länge
// davor, auf ihr und dahinter; die Klassen sind 7, die Felder gleich denen von ImportAnfrage.
func TestIM9Zahlen(t *testing.T) {
	laengen := importLaengen(1, bodyGrenze+1)
	for i := 1; i < len(laengen); i++ {
		if laengen[i] <= laengen[i-1] {
			t.Errorf("die Längen steigen nicht echt an Stelle %d: %d nach %d", i, laengen[i], laengen[i-1])
		}
	}
	for _, schwelle := range []int{64, 128, 256, 1000, 1024, 4096, 8192, 10000, 65536, 100000, 262144, 1000000, bodyGrenze} {
		for _, n := range []int{schwelle - 1, schwelle, schwelle + 1} {
			if !slices.Contains(laengen, n) {
				t.Errorf("die Längen führen %d nicht (Schwelle %d)", n, schwelle)
			}
		}
	}
	gleichAuf(t, "Längen, Ende, Klassen, Felder", []int{len(laengen), laengen[len(laengen)-1], len(klassenDerWerte()), len(importFelder())},
		[]int{importLaengenZahl, bodyGrenze + 1, 7, reflect.TypeFor[inbound.ImportAnfrage]().NumField()})
}
