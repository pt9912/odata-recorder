package admin_test

import (
	"context"
	"errors"
	"io"
	"net/http"
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

// importFeld liefert zu einem Feld und einem Wert das Literal des Gliedes im Body und die Anfrage, die der Dienst erhalten
// soll: gutAnfrage mit dem Wert im Feld (bei `source_session_ids` als einziger Eintrag).
func importFeld(feld, wert string) (literal string, soll inbound.ImportAnfrage) {
	soll, literal = gutAnfrage(), `"`+wert+`"`
	switch feld {
	case "import_id":
		soll.ImportID = wert
	case "service_id":
		soll.ServiceID = wert
	case "source_session_ids":
		soll.SourceSessionIDs, literal = []string{wert}, `["`+wert+`"]`
	case "metadata_hash":
		soll.MetadataHash = wert
	case "conflict_policy":
		soll.ConflictPolicy = wert
	case "target":
		soll.Target = wert
	}
	return literal, soll
}

// importFeldBody bildet den Body, in dem das Feld den Wert trägt, und die Anfrage, die der Dienst erhalten soll.
func importFeldBody(feld string, wert []byte) (body []byte, soll inbound.ImportAnfrage) {
	literal, soll := importFeld(feld, string(wert))
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
			literal, soll := importFeld(feld, z.soll)
			roh, _ := importFeld(feld, z.roh)
			rec := u.rufe(importPost("", nil, strings.NewReader(bodyMitGlied("imp-1", feld, roh))))
			if rec.Code != 201 || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}) {
				t.Errorf("%s in %s: Status %d, Argumente %+v, erwartet 201 und %s", z.name, feld, rec.Code, fake.anfrageListe(), literal)
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 3*len(importFelder()))
	gleichAuf(t, "Felder gleich 5 Zeichenketten und 1 Liste", len(importFelder()), 5+1)
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

// wertFeldFaelle prüft einen Wert in jedem der sechs Felder gegen den Dienst: gueltig heißt 201 und die Anfrage unverändert,
// sonst 400 ohne Aufruf.
func wertFeldFaelle(t *testing.T, name string, wert []byte, gueltig bool) {
	t.Helper()
	for _, feld := range importFelder() {
		u, fake, _ := importsUmgebung(t)
		body, soll := importFeldBody(feld, wert)
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

// TestIM9Laenge hält IM9 (SPEC-297 Nr. 5 e, „unverändert und ungekürzt“) für die Menge Länge × Feld × Zeichenklasse: 12 Längen
// × 6 Felder × 7 Klassen = 504 Fälle, je 201 und der Wert byte-gleich am Dienst. Prüfgrenze (akzeptiertes Negativ (1)): die
// Längen dazwischen und die Mischungen der Klassen in einem Wert.
func TestIM9Laenge(t *testing.T) {
	faelle := 0
	for _, klasse := range klassenDerWerte() {
		for _, n := range laengenDerWerte() {
			wert := wertDerLaenge(klasse, n)
			if len(wert) != n {
				t.Fatalf("Orakel: Klasse %s Länge %d liefert %d Bytes", klasse, n, len(wert))
			}
			wertFeldFaelle(t, klasse+" "+strconv.Itoa(n), []byte(wert), true)
			faelle += len(importFelder())
		}
	}
	gleichAuf(t, "Fälle", faelle, len(laengenDerWerte())*len(importFelder())*len(klassenDerWerte()))
	gleichAuf(t, "Fälle laut Plan", faelle, 12*6*7)
}

// importZeichenMenge sind die 139 Zeichen und Folgen von M-Zeichen (IM9): 32 Steuerzeichen unter 0x20 roh (ungültig), 94
// ASCII-Zeichen von 0x20 bis 0x7F außer Anführungszeichen und Backslash, 7 gültige Folgen und 6 ungültige Folgen; die Folge
// steht in `s<Folge>x`, die Folge 0xC3 vor dem Ende des Wertes in `s<0xC3>`.
func importZeichenMenge() (gueltig, ungueltig [][]byte) {
	umhuellt := func(folge ...byte) []byte { return append(append([]byte("s"), folge...), 'x') }
	for c := range byte(0x20) {
		ungueltig = append(ungueltig, umhuellt(c))
	}
	for c := byte(0x20); c <= 0x7F; c++ {
		if c != '"' && c != 0x5c {
			gueltig = append(gueltig, umhuellt(c))
		}
	}
	for _, f := range [][]byte{{0xC2, 0x80}, {0xDF, 0xBF}, {0xE0, 0xA0, 0x80}, {0xEF, 0xBF, 0xBF}, {0xF0, 0x90, 0x80, 0x80}, {0xF4, 0x8F, 0xBF, 0xBF}, {0xEF, 0xBB, 0xBF}} {
		gueltig = append(gueltig, umhuellt(f...))
	}
	for _, f := range [][]byte{{0x80}, {0xC0, 0x80}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0xFF}} {
		ungueltig = append(ungueltig, umhuellt(f...))
	}
	return gueltig, append(ungueltig, []byte{'s', 0xC3})
}

// TestIM9Zeichen hält IM9 für die Menge Zeichen × Feld: 139 Zeichen und Folgen roh in einer Zeichenkette × 6 Felder = 834
// Fälle; die 101 gültigen (94 + 7) gehen unverändert an den Dienst, die 38 ungültigen (32 + 6) sind 400 ohne Aufruf.
func TestIM9Zeichen(t *testing.T) {
	gueltig, ungueltig := importZeichenMenge()
	gleichAuf(t, "Zeichen und Folgen", []int{len(gueltig), len(ungueltig), len(gueltig) + len(ungueltig)}, []int{101, 38, 139})
	for _, w := range gueltig {
		wertFeldFaelle(t, "gültige Folge "+strconv.Quote(string(w)), w, true)
	}
	for _, w := range ungueltig {
		wertFeldFaelle(t, "ungültige Folge "+strconv.Quote(string(w)), w, false)
	}
	gleichAuf(t, "Fälle laut Plan", (len(gueltig)+len(ungueltig))*len(importFelder()), 834)
}

// TestIM9Stelle hält IM9 für die Stelle eines ungültigen Bytes: 0xFF an jeder Stelle p von 0, 1, 63, 64, 65, 255, 256, 4095,
// 4096, 65535, 65536 und 1048000 des Wertes, in jedem der 6 Felder (72 Fälle), ist 400 ohne Aufruf des Dienstes.
func TestIM9Stelle(t *testing.T) {
	stellen := []int{0, 1, 63, 64, 65, 255, 256, 4095, 4096, 65535, 65536, 1048000}
	for _, p := range stellen {
		wertFeldFaelle(t, "0xFF an Stelle "+strconv.Itoa(p), []byte(strings.Repeat("a", p)+string([]byte{0xFF})+"bb"), false)
	}
	gleichAuf(t, "Fälle laut Plan", len(stellen)*len(importFelder()), 72)
}

// TestIM9Liste hält IM9 für die Zahl der Einträge von `source_session_ids` (SPEC-297 Nr. 5 e): 0, 1, 2, 3, 100 und 40000
// Einträge gehen mit allen Einträgen in ihrer Folge an den Dienst (6 Fälle). Prüfgrenze (akzeptiertes Negativ (1)): 40000.
func TestIM9Liste(t *testing.T) {
	zahlen := []int{0, 1, 2, 3, 100, 40000}
	for _, n := range zahlen {
		liste := make([]string, n)
		var literal strings.Builder
		literal.WriteString("[")
		for i := range liste {
			liste[i] = "e-" + strconv.Itoa(i)
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
		if rec.Code != 201 || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}) {
			t.Errorf("%d Einträge: Status %d, Argumente gleich: %v", n, rec.Code, gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{soll}))
		}
	}
	gleichAuf(t, "Fälle", len(zahlen), 6)
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

// TestIM9Zahlen hält M-Wert gegen die Quelle: 504 + 834 + 72 + 6 + 10 = 1426 Fälle, die Längen und Klassen sind die des Plans.
func TestIM9Zahlen(t *testing.T) {
	g, u := importZeichenMenge()
	gleichAuf(t, "Längen, Klassen, Felder", []int{len(laengenDerWerte()), len(klassenDerWerte()), len(importFelder())}, []int{12, 7, 6})
	gesamt := len(laengenDerWerte())*len(klassenDerWerte())*len(importFelder()) + (len(g)+len(u))*len(importFelder()) + 12*len(importFelder()) + 6 + 10
	gleichAuf(t, "Fälle von M-Wert", gesamt, 1426)
}
