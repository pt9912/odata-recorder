package admin_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// TestIM1Konstruktor hält IM1 (SPEC-297 Nr. 1): der Gutfall liefert genau die zwei Routen in der Folge des Vertrags, ein
// nil-Port ist model.ErrUngueltig ohne Routen und ohne Wert im Text (der andere Port trägt einen Canary, den ein Formatieren
// des Arguments in den Text brächte), die Funktion ruft keinen Port, und `New` nimmt die Routen der Sessions, der Imports und
// der Inspektion zusammen an (4 + 2 + 3 Routen ohne Überlappung); die Zahl der Goroutinen fängt eine Goroutine, die bleibt,
// eine kurzlebige fängt allein die Quelle (TestIM15QuelleDerRouten).
func TestIM1Konstruktor(t *testing.T) {
	audit := supporttest.NewFakeAudit()
	fake := &importsFake{}
	vorher := runtime.NumGoroutine()
	var routen []admin.Route
	var err error
	for range 100 {
		routen, err = admin.ImportRouten(fake, audit)
	}
	if nachher := runtime.NumGoroutine(); err != nil || nachher > vorher+10 {
		t.Errorf("ImportRouten: %v, Goroutinen vorher %d, nachher %d", err, vorher, nachher)
	}
	got := [][2]string{}
	for _, r := range routen {
		if r.Handler == nil {
			t.Errorf("Route %s %s ohne Handler", r.Methode, r.Muster)
		}
		got = append(got, [2]string{r.Methode, r.Muster})
	}
	gleichAuf(t, "Routen in der Folge des Vertrags", got, [][2]string{{"POST", "/__admin/imports"}, {"GET", "/__admin/imports/{id}"}})
	if fake.gesamt() != 0 || audit.Aufrufe() != 0 {
		t.Errorf("ImportRouten rief %d-mal den Dienst und %d-mal Audit, erwartet 0 und 0", fake.gesamt(), audit.Aufrufe())
	}

	mitCanary := supporttest.NewFakeAudit()
	_ = mitCanary.Append(support.Eintrag{ServiceID: canary()})
	faelle := []struct {
		name    string
		imports inbound.ImportUseCase
		audit   support.AuditLog
	}{
		{"imports nil", nil, mitCanary},
		{"audit nil", &importsFake{imp: model.Import{ID: canary()}}, nil},
	}
	for _, f := range faelle {
		routen, err := admin.ImportRouten(f.imports, f.audit)
		if routen != nil || !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v Routen, Fehler %v, erwartet keine und model.ErrUngueltig", f.name, routen, err)
			continue
		}
		if text := fmt.Sprintf("%v %+v %#v", err, err, err); strings.Contains(text, canary()) {
			t.Errorf("%s: der Text des Fehlers nennt einen Wert: %q", f.name, text)
		}
	}
	gleichAuf(t, "Zahl der Felder und Routen", []int{len(faelle), len(got)}, []int{2, 2})
	gleichAuf(t, "Aufrufe von Audit nach den nil-Fällen", mitCanary.Aufrufe(), 1)

	k, _ := konfigMit()
	sessionsRouten, err := admin.SessionRouten(&sessionsFake{}, audit)
	if err != nil {
		t.Fatal(err)
	}
	inspektionRouten, err := admin.InspektionRouten(&inspektionFake{})
	if err != nil {
		t.Fatal(err)
	}
	k.Routen = slices.Concat(sessionsRouten, routen, inspektionRouten)
	if srv, err := admin.New(k); err != nil || srv == nil || len(k.Routen) != 4+2+3 {
		t.Errorf("New mit den Routen der Sessions, der Imports und der Inspektion: %v, %d Routen, erwartet 9 ohne Fehler", err, len(k.Routen))
	}
}

// TestIM2Zuordnung hält IM2 (SPEC-297 Nr. 1, SPEC-238 Nr. 2): M-Route, zwei Pfadformen mal sieben Methoden, 14 Fälle:
// `POST /__admin/imports` und `GET /__admin/imports/{id}` treffen (2), die 12 anderen sind 405 mit dem `Allow` der Pfadform
// (`POST`; `GET`, als Literale) und rufen den Dienst nie. Zählprüfung: die Pfadformen sind die verschiedenen Muster der
// gelieferten Routen (2) und die Routen die zwei der Liste.
func TestIM2Zuordnung(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	formen := []struct {
		pfad, allow, treffer string
		op                   impOp
	}{
		{"/__admin/imports", "POST", "POST", impImport},
		{"/__admin/imports/imp-pfad", "GET", "GET", impGet},
	}
	faelle, treffer := 0, 0
	for _, f := range formen {
		for _, m := range methoden {
			faelle++
			u, fake, _ := importsUmgebung(t)
			rec := u.rufe(anfrageMit(m, f.pfad, nil, strings.NewReader(gutBody)))
			name := m + " " + f.pfad
			if m == f.treffer {
				treffer++
				if rec.Code/100 != 2 {
					t.Errorf("%s: Status %d", name, rec.Code)
				}
				fake.nurDiagonale(t, name, f.op)
				continue
			}
			pruefeFehlerAntwort(t, name, rec, 405, objekt404, f.allow)
			if codes, _ := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || fake.gesamt() != 0 || u.audit.Aufrufe() != 0 {
				t.Errorf("%s: CountError %q, %d Aufrufe des Dienstes, %d von Audit", name, codes, fake.gesamt(), u.audit.Aufrufe())
			}
		}
	}
	_, _, routen := importsUmgebung(t)
	muster := map[string]bool{}
	for _, r := range routen {
		muster[r.Muster] = true
	}
	gleichAuf(t, "Fälle, Treffer, Pfadformen, Routen", []int{faelle, treffer, len(muster), len(routen)}, []int{14, 2, len(formen), 2})
}

// TestIM3AufrufeUndKontext hält IM3 (SPEC-297 Nr. 2): jede der zwei Routen ruft genau eine Operation von imports (2 Routen ×
// 2 Werte von `Neu` × 2 Zähler = 8 Zellen, 1 auf der Diagonale, sonst 0) mit den Argumenten unverändert; der Wert im ctx des
// Requests kommt an, und der Handler gibt ctx ohne Ableitung weiter (Gleichheit der Identität); von audit läuft in diesem Teil
// nichts. Die Aufrufzahl jedes Erfolgs und jedes Fehlers zählen außerdem `impZustand.sende` (IM10 bis IM12) und TestIM13Fehler.
func TestIM3AufrufeUndKontext(t *testing.T) {
	gleichAuf(t, "Zähler gleich Zahl der Operationen des Ports", []int{int(impAnzahl), reflect.TypeFor[inbound.ImportUseCase]().NumMethod()}, []int{2, 2})
	faelle := []struct {
		name, methode, ziel, body string
		op                        impOp
		anfrage                   []inbound.ImportAnfrage
		ids                       []string
	}{
		{"POST", "POST", importsZiel, gutBody, impImport, []inbound.ImportAnfrage{gutAnfrage()}, nil},
		{"GET", "GET", importsZiel + "/imp-pfad", "", impGet, nil, []string{"imp-pfad"}},
	}
	zellen := 0
	for i, f := range faelle {
		for _, neu := range []bool{true, false} {
			name := fmt.Sprintf("%s mit Neu %v", f.name, neu)
			u, fake, routen := importsUmgebung(t)
			fake.ergebnis.Neu = neu
			r := anfrageMit(f.methode, f.ziel, nil, strings.NewReader(f.body))
			r = r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, "wert-"+f.name))
			if rec := u.rufe(r); rec.Code/100 != 2 {
				t.Fatalf("%s: Status %d", name, rec.Code)
			}
			zellen += len(fake.zahlen())
			fake.nurDiagonale(t, name, f.op)
			if !gleicheAnfragen(fake.anfrageListe(), f.anfrage) || !slices.Equal(fake.kennungen(), f.ids) {
				t.Errorf("%s: Argumente %+v und %q, erwartet %+v und %q", name, fake.anfrageListe(), fake.kennungen(), f.anfrage, f.ids)
			}
			if got := fake.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-"+f.name {
				t.Errorf("%s: der Wert im ctx des Requests kam als %v an", name, got)
			}
			if n := u.audit.Aufrufe(); n != 0 {
				t.Errorf("%s: %d Aufrufe von Audit, erwartet 0", name, n)
			}

			ctx := context.WithValue(context.Background(), ctxSchluessel{}, "direkt")
			_, _ = routen[i].Handler(ctx, admin.Anfrage{
				Methode: f.methode, Body: strings.NewReader(f.body), KorrelationsID: "cor-1", Pfadwerte: map[string]string{"id": "imp-pfad"},
			})
			if got := fake.ctxListe(); len(got) != 2 || got[1] != ctx {
				t.Errorf("%s: der Handler reichte einen anderen ctx weiter als den, den er erhielt", name)
			}
		}
	}
	gleichAuf(t, "Zellen (Routen × Neu × Zähler)", zellen, 2*2*2)
}

// importQueries sind die sechs Queries von M-Query: leer, ein Parameter, doppelte Parameter, zwei ungültige Prozentfolgen und
// ein Semikolon.
func importQueries() []string { return []string{"", "limit=x", "a=1&a=2", "%zz", "%", "a=1;b=2"} }

// TestIM4Query hält IM4 (SPEC-297 Nr. 3) für Query und Body von `GET`: keine der zwei Routen liest die Query, jede Form (6
// Queries × 2 Routen = 12 Fälle) ist ohne Wirkung gegenüber dem Request ohne Query; `GET` liest den Body nie (3 ungültige
// Bodies, 0 Lesezugriffe). Zählprüfung: 15.
func TestIM4Query(t *testing.T) {
	faelle := 0
	for _, q := range importQueries() {
		u, fake, _ := importsUmgebung(t)
		rec := u.rufe(importPost(q, nil, strings.NewReader(gutBody)))
		if rec.Code != 201 || rec.Body.String() != reportA {
			t.Errorf("POST mit Query %q: %d %q", q, rec.Code, rec.Body.String())
		}
		rec = u.rufe(anfrageMit("GET", importsZiel+"/imp-pfad?"+q, nil, nil))
		if rec.Code != 200 || rec.Body.String() != reportA {
			t.Errorf("GET mit Query %q: %d %q", q, rec.Code, rec.Body.String())
		}
		if !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{gutAnfrage()}) || !slices.Equal(fake.kennungen(), []string{"imp-pfad"}) {
			t.Errorf("Query %q: Argumente %+v und %q", q, fake.anfrageListe(), fake.kennungen())
		}
		gleichAuf(t, "Zähler mit Query "+q, fake.zahlen(), [impAnzahl]int{1, 1})
		faelle += 2
	}
	bodies := [][]byte{[]byte("{"), []byte("[1,2"), {0xFF, 0xFE, 0x00}}
	for i, b := range bodies {
		u, fake, _ := importsUmgebung(t)
		leser := &bodyLeser{daten: b}
		if rec := u.rufe(anfrageMit("GET", importsZiel+"/imp-pfad", nil, leser)); rec.Code != 200 || rec.Body.String() != reportA || leser.n.Load() != 0 || fake.gesamt() != 1 {
			t.Errorf("GET mit ungültigem Body Nr. %d: Status %d, %d Lesezugriffe, %d Aufrufe", i, rec.Code, leser.n.Load(), fake.gesamt())
		}
		faelle++
	}
	gleichAuf(t, "Fälle", faelle, len(importQueries())*2+len(bodies))
	gleichAuf(t, "Fälle laut Plan", faelle, 15)
}

// importKopf sind die 26 Paare aus Name und Wert von M-Header (20 Namen): `Content-Type` mit 7 Werten und je ein Wert für 19
// weitere Header, die eine Route lesen könnte.
func importKopf() [][2]string {
	var out [][2]string
	for _, ct := range []string{
		"application/json", "application/json; charset=utf-8", "application/json; charset=latin1", "text/plain",
		"application/x-www-form-urlencoded", "multipart/form-data; boundary=x", "",
	} {
		out = append(out, [2]string{"Content-Type", ct})
	}
	for _, n := range []string{
		"Content-Encoding", "Content-Language", "Content-Disposition", "Accept", "Accept-Encoding", "Accept-Language",
		"Transfer-Encoding", "If-Match", "If-None-Match", "If-Modified-Since", "Cookie", "Origin", "Referer", "User-Agent",
		"X-Correlation-Id", "X-Forwarded-For", "X-Recorder-Recording-Session", "Prefer", "Expect",
	} {
		out = append(out, [2]string{n, "wert-" + strings.ToLower(n)})
	}
	return out
}

// TestIM4Header hält IM4 (SPEC-297 Nr. 3) für Header und Länge: keine Route liest einen Header oder `Laenge`, jedes der 26
// Paare (20 Namen) und jede der 7 Längen (-1, 0, 1, 20, 1 MiB, 1 MiB + 1, größter int64) ist ohne Wirkung, bei `POST` mit
// gültigem Body (201) und bei `GET` (200), die Argumente am Dienst wie ohne. Zählprüfung: 33 × 2 = 66.
// Prüfgrenze (akzeptiertes Negativ (2)): ein Header außerhalb der 20 Namen fängt allein TestIM15QuelleDerRouten.
func TestIM4Header(t *testing.T) {
	gleichAuf(t, "Paare und Längen", []int{len(importKopf()), len(laengen())}, []int{26, 7})
	namen := map[string]bool{}
	for _, p := range importKopf() {
		namen[p[0]] = true
	}
	gleichAuf(t, "Namen", len(namen), 20)
	faelle := 0
	pruefe := func(name string, kopf map[string][]string, mitLaenge bool, laenge int64) {
		for _, post := range []bool{true, false} {
			u, fake, _ := importsUmgebung(t)
			var r = anfrageMit("GET", importsZiel+"/imp-pfad", kopf, nil)
			status, op := 200, impGet
			if post {
				r, status, op = anfrageMit("POST", importsZiel, kopf, strings.NewReader(gutBody)), 201, impImport
			}
			if mitLaenge {
				r.ContentLength = laenge
			}
			if rec := u.rufe(r); rec.Code != status || rec.Body.String() != reportA {
				t.Errorf("%s (POST %v): %d %q, erwartet %d %q", name, post, rec.Code, rec.Body.String(), status, reportA)
			}
			fake.nurDiagonale(t, name, op)
			if post && !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{gutAnfrage()}) || !post && !slices.Equal(fake.kennungen(), []string{"imp-pfad"}) {
				t.Errorf("%s (POST %v): Argumente %+v und %q", name, post, fake.anfrageListe(), fake.kennungen())
			}
			faelle++
		}
	}
	for _, p := range importKopf() {
		pruefe("Header "+p[0]+": "+p[1], map[string][]string{p[0]: {p[1]}}, false, 0)
	}
	for _, l := range laengen() {
		pruefe(fmt.Sprintf("Länge %d", l), nil, true, l)
	}
	gleichAuf(t, "Fälle", faelle, 66)
}

// importPfadwerte ist M-Pfadwert, die 13 kodierten Werte von `{id}` mit den dekodierten Bytes als Literale aus Bytewerten.
func importPfadwerte() []struct{ kodiert, soll string } {
	bytesAus := func(b ...byte) string { return string(b) }
	return []struct{ kodiert, soll string }{
		{"x", "x"}, {"imp-1", "imp-1"}, {"IMP-1", "IMP-1"}, {"a.b", "a.b"}, {".", "."}, {"..", ".."}, {"a%2Fb", "a/b"},
		{"%C3%A4", bytesAus(0xC3, 0xA4)}, {"%FF", bytesAus(0xFF)}, {"%00", bytesAus(0x00)},
		{strings.Repeat("a", 65), strings.Repeat("a", 65)}, {"a+b", "a+b"}, {"%20", bytesAus(0x20)},
	}
}

// TestIM5Pfadwert hält IM5 (SPEC-297 Nr. 4): jeder der 13 Werte (Form, Länge, Schreibweise, `.`, `..`, Prozentfolgen, `+`) geht
// byte-gleich an `Get`, die Antwort ist der Report ohne den Wert, und der Fehler `ODR-E7005` steht ohne den Wert (jedes Byte
// der Antwort ist das Literal); ein Canary als Wert steht in keinem Byte von Status, Header und Body. Zählprüfung: 13 Werte.
// Die Länge des Wertes führt TestIM5Laenge.
func TestIM5Pfadwert(t *testing.T) {
	faelle := 0
	for _, w := range importPfadwerte() {
		faelle++
		u, fake, _ := importsUmgebung(t)
		rec := u.rufe(importGet(w.kodiert))
		if rec.Code != 200 || rec.Body.String() != reportA {
			t.Errorf("%s: Antwort %d %q, erwartet 200 %q", w.kodiert, rec.Code, rec.Body.String(), reportA)
		}
		gleichAuf(t, w.kodiert+": Argumente", fake.kennungen(), []string{w.soll})
		fake.nurDiagonale(t, w.kodiert, impGet)

		u, fake, _ = importsUmgebung(t)
		fake.err = &inbound.Fehler{Code: model.CodeE7005, Status: 404}
		rec = u.rufe(importGet(w.kodiert))
		pruefeFehlerAntwort(t, w.kodiert+" unbekannt", rec, 404, objektE7005, "")
		gleichAuf(t, w.kodiert+": Argumente beim Fehler", fake.kennungen(), []string{w.soll})
	}
	for _, fehler := range []error{nil, &inbound.Fehler{Code: model.CodeE7005, Status: 404}} {
		u, fake, _ := importsUmgebung(t)
		fake.err = fehler
		rec := u.rufe(importGet(canary()))
		if strings.Contains(antwortText(rec), canary()) {
			t.Errorf("der Canary steht in der Antwort: %s", antwortText(rec))
		}
		gleichAuf(t, "Argumente mit Canary", fake.kennungen(), []string{canary()})
	}
	gleichAuf(t, "Fälle", faelle, 13)
}

// TestIM5JedesByte hält IM5 für die Zeichenmenge des Pfadwertes: jedes der 256 Bytes allein und zwischen zwei Buchstaben
// (`a<Byte>z`), als `%XX` kodiert, geht byte-gleich an `Get`; das Orakel ist die Zählschleife, nicht der Code der Route.
// Zählprüfung: 256 Bytes × 2 Formen = 512 Fälle.
func TestIM5JedesByte(t *testing.T) {
	faelle := 0
	for b := range 256 {
		for _, form := range []string{"%s", "a%sz"} {
			kodiert := strings.Replace(form, "%s", fmt.Sprintf("%%%02X", b), 1)
			soll := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
			u, fake, _ := importsUmgebung(t)
			rec := u.rufe(importGet(kodiert))
			faelle++
			if rec.Code != 200 || rec.Body.String() != reportA {
				t.Errorf("Byte 0x%02X in der Form %q: Antwort %d %q, erwartet 200", b, form, rec.Code, rec.Body.String())
			}
			gleichAuf(t, fmt.Sprintf("Byte 0x%02X in der Form %q: Argumente", b, form), fake.kennungen(), []string{soll})
		}
	}
	gleichAuf(t, "Fälle laut Plan", faelle, 512)
}

// TestIM5Prozentfolgen hält IM5 für die Struktur im dekodierten Wert (die Route dekodiert nichts nach): das Zeichen `%` an
// jeder Stelle des Wertes (allein, nach einem Buchstaben, vor einem, zwischen zweien, 4 Stellungen) vor 0, 1, 2 und 3
// weiteren Zeichen. Folgen: 0 Zeichen (1), 1 Zeichen aus 25 (die 22 Hexziffern `0-9`, `a-f`, `A-F`, dazu `g`, `%` und `+`),
// 2 Zeichen (alle 625 Paare dieses Alphabets), 3 Zeichen aus 6 (`0`, `F`, `a`, `g`, `%`, `+`: 216) = 867 Folgen; jedes Byte
// des dekodierten Wertes ist als `%XX` kodiert, die Route sieht also ein `%` im Wert, das nur eine zweite Dekodierung auflöst.
// Zählprüfung: 867 × 4 = 3468 Fälle.
func TestIM5Prozentfolgen(t *testing.T) {
	zeichen, klein := "0123456789abcdefABCDEFg%+", "0Fag%+"
	gleichAuf(t, "Alphabet", len(zeichen), 25)
	folgen := []string{""}
	for _, a := range zeichen {
		folgen = append(folgen, string(a))
		for _, b := range zeichen {
			folgen = append(folgen, string(a)+string(b))
		}
	}
	for _, a := range klein {
		for _, b := range klein {
			for _, c := range klein {
				folgen = append(folgen, string(a)+string(b)+string(c))
			}
		}
	}
	gleichAuf(t, "Folgen", len(folgen), 1+25+625+216)
	u, fake, _ := importsUmgebung(t)
	faelle := 0
	for _, folge := range folgen {
		for _, vor := range []string{"", "a"} {
			for _, nach := range []string{"", "z"} {
				wert := vor + "%" + folge + nach
				var kodiert strings.Builder
				for _, b := range []byte(wert) {
					fmt.Fprintf(&kodiert, "%%%02X", b)
				}
				rec := u.rufe(importGet(kodiert.String()))
				faelle++
				ids := fake.kennungen()
				letzte := ""
				if len(ids) > 0 {
					letzte = ids[len(ids)-1]
				}
				if rec.Code != 200 || len(ids) != faelle || letzte != wert {
					t.Fatalf("Wert %q: Antwort %d, %d Aufrufe, letzte Kennung %q; erwartet 200 und den Wert unverändert", wert, rec.Code, len(ids), letzte)
				}
			}
		}
	}
	gleichAuf(t, "Fälle laut Plan", faelle, 867*4)
	gleichAuf(t, "Fälle gleich 13 + 512 + 3468 mit den zwei anderen Tests", 13+512+faelle, 3993)
}

// TestIM5Laenge hält IM5 für die Länge des Pfadwertes (SPEC-297 Nr. 4, „ohne Prüfung von Form, Länge oder Schreibweise“): jede
// Länge von importLaengen(1, bodyGrenze+1) (Zweier- und Zehnerpotenzen um 1, bis 1 MiB + 1 Bytes) × 7 Zeichenklassen des Wertes
// (die Bytes der Mehrbyteklasse als `%XX` kodiert) geht byte-gleich an Get; Orakel ist wertDerLaenge, nicht der Code der Route.
// Zählprüfung: Längen × 7 Klassen. Prüfgrenze (akzeptiertes Negativ (1)): die Längen dazwischen und die Mischung der Klassen.
func TestIM5Laenge(t *testing.T) {
	laengen := importLaengen(1, bodyGrenze+1)
	faelle := 0
	for _, klasse := range klassenDerWerte() {
		for _, n := range laengen {
			wert := wertDerLaenge(klasse, n)
			if len(wert) != n {
				t.Fatalf("Orakel: Klasse %s Länge %d liefert %d Bytes", klasse, n, len(wert))
			}
			u, fake, _ := importsUmgebung(t)
			rec := u.rufe(importGet(url.PathEscape(wert)))
			faelle++
			if rec.Code != 200 || !slices.Equal(fake.kennungen(), []string{wert}) {
				t.Errorf("Klasse %s, %d Bytes: Status %d, Argumente gleich: %v", klasse, n, rec.Code, slices.Equal(fake.kennungen(), []string{wert}))
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, len(laengen)*len(klassenDerWerte()))
}

// impFolgeSchritt ist ein Request von M-Folge und das, was der Dienst ihm liefert, mit der Antwort als Literal des Vertrags.
type impFolgeSchritt struct {
	name, route string
	ergebnis    inbound.ImportErgebnis
	imp         model.Import
	err         error
	body        string
	status      int
	antwort     string
}

// reportB und reportC sind Reports, die sich von reportA und voneinander unterscheiden.
const (
	reportB = `{"import_id":"imp-b","status":"wiederholt"}`
	reportC = `{"import_id":"imp-c","status":"gelesen"}`
)

// impFolgeSchritte sind die 10 Antworten von M-Folge (IM17).
func impFolgeSchritte() []impFolgeSchritt {
	mitReport := func(r string) model.Import { i := importA(); i.ReportJSON = r; return i }
	return []impFolgeSchritt{
		{name: "POST 201", route: "POST", ergebnis: inbound.ImportErgebnis{Import: mitReport(reportA), Neu: true}, body: gutBody, status: 201, antwort: reportA},
		{name: "POST 200", route: "POST", ergebnis: inbound.ImportErgebnis{Import: mitReport(reportB)}, body: gutBody, status: 200, antwort: reportB},
		{name: "POST 400", route: "POST", body: "{", status: 400, antwort: objekt400},
		{name: "POST 404", route: "POST", err: &inbound.Fehler{Code: model.CodeE5005, Status: 404}, body: gutBody, status: 404, antwort: objektE5005},
		{name: "POST 409", route: "POST", err: &inbound.Fehler{Code: model.CodeE6102, Status: 409}, body: gutBody, status: 409, antwort: objektE6102},
		{name: "POST 422", route: "POST", err: &inbound.Fehler{Code: model.CodeE6004, Status: 422}, body: gutBody, status: 422, antwort: objektE6004},
		{name: "POST 500 Report", route: "POST", ergebnis: inbound.ImportErgebnis{Import: mitReport("{"), Neu: true}, body: gutBody, status: 500, antwort: objekt500},
		{name: "GET 200", route: "GET", imp: mitReport(reportC), status: 200, antwort: reportC},
		{name: "GET 404", route: "GET", err: &inbound.Fehler{Code: model.CodeE7005, Status: 404}, status: 404, antwort: objektE7005},
		{name: "GET 500", route: "GET", err: &inbound.Fehler{Code: model.CodeE3001, Status: 500}, status: 500, antwort: objektE3001},
	}
}

// TestIM17FolgeOhneZustand hält IM17 (SPEC-297 Nr. 11): M-Folge ist jedes geordnete Paar aus den 10 Schritten (100 Paare,
// auch gleiche) auf demselben Server, der Fake je Schritt neu eingestellt. Jede Antwort ist die zu ihrem Schritt, nie die
// eines früheren (Orakel: die Literale, gleich der Antwort eines frischen Servers).
func TestIM17FolgeOhneZustand(t *testing.T) {
	alle := impFolgeSchritte()
	gleichAuf(t, "Zahl der Schritte", len(alle), 10)
	paare := 0
	for _, x := range alle {
		for _, y := range alle {
			paare++
			u, fake, _ := importsUmgebung(t)
			for _, s := range []impFolgeSchritt{x, y} {
				fake.ergebnis, fake.imp, fake.err = s.ergebnis, s.imp, s.err
				r := importGet("imp-pfad")
				if s.route == "POST" {
					r = importPost("", nil, strings.NewReader(s.body))
				}
				rec := u.rufe(r)
				if rec.Code != s.status || rec.Body.String() != s.antwort {
					t.Errorf("%s dann %s, Schritt %s: %d %q, erwartet %d %q", x.name, y.name, s.name, rec.Code, rec.Body.String(), s.status, s.antwort)
				}
			}
		}
	}
	gleichAuf(t, "Zahl der Paare", paare, 10*10)
}
