package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// impZustand ist einer der drei Zustände der Antworten: `POST` mit `Neu` wahr, `POST` mit `Neu` falsch (die Wiederholung) und
// `GET`; status ist der Status einer Erfolgsantwort.
type impZustand struct {
	name      string
	post, neu bool
	status    int
}

func impZustaende() []impZustand {
	return []impZustand{{"POST neu", true, true, 201}, {"POST wiederholt", true, false, 200}, {"GET", false, false, 200}}
}

// sende stellt den Dienst auf den Import und sendet den Request des Zustands; der Aufruf ändert die Zähler des Fakes genau um
// einen Aufruf der Operation der Route (je Operation des Ports, auch die, die die Route nie rufen darf) und ändert die
// Aufrufzahl von Audit um die des Orakels von SPEC-297 Nr. 10 a: jede 2xx-Antwort auf `POST` hat genau einen Eintrag, jede
// andere Antwort und jedes `GET` keinen (Aufrufzahl je Route × `Neu` × Ausgang).
func (z impZustand) sende(t *testing.T, u *umgebung, fake *importsFake, imp model.Import) *httptest.ResponseRecorder {
	t.Helper()
	vorher, auditVorher := fake.zahlen(), u.audit.Aufrufe()
	fake.ergebnis, fake.imp = inbound.ImportErgebnis{Import: imp, Neu: z.neu}, imp
	op, r := impGet, importGet("imp-pfad")
	if z.post {
		op, r = impImport, importPost("", nil, strings.NewReader(gutBody))
	}
	rec := u.rufe(r)
	for o, n := range fake.zahlen() {
		soll := 0
		if impOp(o) == op {
			soll = 1
		}
		if n-vorher[o] != soll {
			t.Errorf("%s: Operation %d wurde %d-mal gerufen, erwartet %d", z.name, o, n-vorher[o], soll)
		}
	}
	soll := 0
	if z.post && rec.Code/100 == 2 {
		soll = 1
	}
	if n := u.audit.Aufrufe() - auditVorher; n != soll {
		t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", z.name, n, soll)
	}
	return rec
}

// TestIM10Status hält IM10 (SPEC-297 Nr. 6, SPEC-238 Nr. 5): `POST` antwortet 201, wenn `Neu` wahr ist, sonst 200, `GET` 200;
// die Erfolgsantwort trägt `Content-Type`, `Content-Length` und `X-Correlation-Id` und sonst keinen Header (kein `Location`);
// der Status hängt allein von `Neu` ab: jedes der 6 Felder von model.Import mit eigenem Wert ändert in den 2 Zellen von `POST`
// (12 Fälle) Status und Body nicht (der Body folgt allein dem Feld `ReportJSON`). Zählprüfung: 3 + 12.
func TestIM10Status(t *testing.T) {
	for _, z := range impZustaende() {
		u, fake, _ := importsUmgebung(t)
		rec := z.sende(t, u, fake, importA())
		if rec.Code != z.status || rec.Body.String() != reportA {
			t.Errorf("%s: Antwort %d %q, erwartet %d %q", z.name, rec.Code, rec.Body.String(), z.status, reportA)
		}
		pruefeErfolgsHeader(t, z.name, rec)
	}
	gleichAuf(t, "Felder von model.Import", reflect.TypeOf(model.Import{}).NumField(), 6)
	faelle := 0
	for i := range reflect.TypeOf(model.Import{}).NumField() {
		for _, z := range impZustaende()[:2] {
			faelle++
			imp := importA()
			setzeFeld(t, reflect.ValueOf(&imp).Elem(), i, "anders-"+strconv.Itoa(i), zeitA().Add(time.Hour))
			soll := reportA
			if reflect.TypeOf(imp).Field(i).Name == "ReportJSON" {
				imp.ReportJSON, soll = `{"anders":`+strconv.Itoa(i)+`}`, `{"anders":`+strconv.Itoa(i)+`}`
			}
			u, fake, _ := importsUmgebung(t)
			rec := z.sende(t, u, fake, imp)
			if rec.Code != z.status || rec.Body.String() != soll {
				t.Errorf("Feld %d, %s: Antwort %d %q, erwartet %d %q", i, z.name, rec.Code, rec.Body.String(), z.status, soll)
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 12)
	// Der Nullwert jedes Feldes außer ReportJSON (5 Felder × 2 Zellen = 10 Zusatzfälle, die der Plan nicht nennt): der Status
	// hängt auch von keinem leeren Feld ab.
	nullwerte := 0
	for i := range reflect.TypeOf(model.Import{}).NumField() {
		if reflect.TypeOf(model.Import{}).Field(i).Name == "ReportJSON" {
			continue
		}
		for _, z := range impZustaende()[:2] {
			nullwerte++
			imp := importA()
			reflect.ValueOf(&imp).Elem().Field(i).SetZero()
			u, fake, _ := importsUmgebung(t)
			if rec := z.sende(t, u, fake, imp); rec.Code != z.status || rec.Body.String() != reportA {
				t.Errorf("Feld %d leer, %s: Antwort %d %q, erwartet %d %q", i, z.name, rec.Code, rec.Body.String(), z.status, reportA)
			}
		}
	}
	gleichAuf(t, "Nullwert-Fälle", nullwerte, 10)
}

// TestIM11Report hält IM11 (SPEC-297 Nr. 6): der Body von `POST` (201 und 200) und von `GET` ist der Text von `ReportJSON`
// Byte für Byte, ohne Neuserialisierung, `Content-Length` gleich der Länge: 15 gültige Formen × 3 Zustände = 45 und 7 Längen
// × 3 = 21, zusammen 66. Orakel: die Bytes als Literale aus Bytewerten, nie `encoding/json`. Prüfgrenze (akzeptiertes Negativ
// (1)): 4 MiB + 1 Bytes.
func TestIM11Report(t *testing.T) {
	var formen []reportForm
	for _, n := range reportsLaengen() {
		formen = append(formen, reportForm{"Länge " + strconv.Itoa(n), reportDerLaenge(n)})
	}
	gleichAuf(t, "Formen und Längen", []int{len(reportsGueltig()), len(formen), len(impZustaende())}, []int{15, 7, 3})
	faelle := 0
	for _, f := range append(reportsGueltig(), formen...) {
		for _, z := range impZustaende() {
			faelle++
			imp := importA()
			imp.ReportJSON = string(f.roh)
			u, fake, _ := importsUmgebung(t)
			rec := z.sende(t, u, fake, imp)
			name := f.name + " " + z.name
			if rec.Code != z.status || !bytes.Equal(rec.Body.Bytes(), f.roh) {
				t.Errorf("%s: Status %d, Body gleich dem Report: %v (%d von %d Bytes)", name, rec.Code, bytes.Equal(rec.Body.Bytes(), f.roh), rec.Body.Len(), len(f.roh))
			}
			if rec.Header().Get("Content-Length") != strconv.Itoa(len(f.roh)) {
				t.Errorf("%s: Content-Length %q, erwartet %d", name, rec.Header().Get("Content-Length"), len(f.roh))
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 66)
}

// TestIM11ZweiAufrufe hält IM11 für die Wiederholung: `POST` und `POST` sowie `POST` und `GET` mit demselben Report liefern
// gleiche Bytes (15 Formen × 2 Folgen).
func TestIM11ZweiAufrufe(t *testing.T) {
	for _, f := range reportsGueltig() {
		for _, zweiter := range impZustaende()[1:] {
			imp := importA()
			imp.ReportJSON = string(f.roh)
			u, fake, _ := importsUmgebung(t)
			erster := impZustaende()[0].sende(t, u, fake, imp)
			zweite := zweiter.sende(t, u, fake, imp)
			if !bytes.Equal(erster.Body.Bytes(), f.roh) || !bytes.Equal(zweite.Body.Bytes(), erster.Body.Bytes()) {
				t.Errorf("%s, POST dann %s: die Bodys sind nicht gleich dem Report", f.name, zweiter.name)
			}
		}
	}
}

// TestIM12UngueltigerReport hält IM12 (SPEC-297 Nr. 6): ein Report, der kein gültiges JSON in UTF-8 ist (7 Formen × 3
// Zustände = 21), ergibt 500 `ODR-E9001`, keine Teilantwort, genau 1 Zählen, bei `POST` keinen Eintrag; kein Byte des Reports
// steht im Fehlerobjekt; der Handler allein liefert einen Fehler ohne Code (nicht model.ErrUngueltig) ohne Wert im Text und keine
// Antwort.
func TestIM12UngueltigerReport(t *testing.T) {
	formen := reportsUngueltig(canary())
	gleichAuf(t, "Formen", len(formen), 7)
	faelle := 0
	for _, f := range formen {
		for _, z := range impZustaende() {
			faelle++
			name := f.name + " " + z.name
			imp := importA()
			imp.ReportJSON = string(f.roh)
			u, fake, routen := importsUmgebung(t)
			rec := z.sende(t, u, fake, imp)
			pruefeFehlerAntwort(t, name, rec, 500, objekt500, "")
			if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E9001"}) || andere != 0 {
				t.Errorf("%s: CountError %q, andere Aufrufe %d, erwartet genau einmal ODR-E9001", name, codes, andere)
			}
			if u.audit.Aufrufe() != 0 {
				t.Errorf("%s: %d Aufrufe von Audit, erwartet 0", name, u.audit.Aufrufe())
			}
			op := impGet
			if z.post {
				op = impImport
			}
			fake.nurDiagonale(t, name, op)

			i := 0
			if !z.post {
				i = 1
			}
			ant, err := routen[i].Handler(context.Background(), admin.Anfrage{
				Methode: "POST", Body: strings.NewReader(gutBody), Pfadwerte: map[string]string{"id": "imp-pfad"},
			})
			if err == nil || errors.Is(err, model.ErrUngueltig) || strings.Contains(err.Error(), canary()) || ant.Status != 0 || len(ant.Body) != 0 {
				t.Errorf("%s: Handler liefert %q und %v, erwartet einen Fehler ohne Code, ohne Wert und keine Antwort", name, ant.Body, err)
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 21)
}

// pruefeReport sendet den Report roh im Zustand z: gültig heißt der Status des Zustands, der Body Byte für Byte der Report und
// `Content-Length` seine Länge, sonst 500 mit dem Fehlerobjekt `ODR-E9001`; der Zustand zählt die Aufrufe (sende).
func pruefeReport(t *testing.T, name string, z impZustand, roh []byte, gueltig bool) {
	t.Helper()
	imp := importA()
	imp.ReportJSON = string(roh)
	u, fake, _ := importsUmgebung(t)
	rec := z.sende(t, u, fake, imp)
	switch {
	case gueltig && (rec.Code != z.status || !bytes.Equal(rec.Body.Bytes(), roh) || rec.Header().Get("Content-Length") != strconv.Itoa(len(roh))):
		t.Errorf("%s, %s: Status %d, Body gleich dem Report: %v (%d von %d Bytes)", name, z.name, rec.Code, bytes.Equal(rec.Body.Bytes(), roh), rec.Body.Len(), len(roh))
	case !gueltig && (rec.Code != 500 || rec.Body.String() != objekt500):
		t.Errorf("%s, %s: Status %d, erwartet 500 und das Fehlerobjekt von ODR-E9001", name, z.name, rec.Code)
	}
}

// TestIM11Zeichen hält IM11 und IM12 für die Menge Zeichen × Form × Stelle × Zustand (SPEC-297 Nr. 6): jedes der 303
// Elemente von importZeichenMenge im Text einer Zeichenkette des Reports, am Anfang, in der Mitte und am Ende, in jeder der 4
// Formen von reportMitNutzlast und in jedem der 3 Zustände = 10908 Fälle; die 259 gültigen kommen Byte für Byte an (auch rohes
// U+FFFD, NBSP und U+FEFF und der Escape eines einzelnen Surrogats, die ein Dienst schreiben kann), die 44 ungültigen (rohes
// Steuerzeichen, ungültiges UTF-8, ungültiger Escape) sind 500 `ODR-E9001`.
func TestIM11Zeichen(t *testing.T) {
	faelle := 0
	for _, z := range importZeichenMenge() {
		for _, st := range importZeichenStellen() {
			for form := range reportFormenZahl {
				f := reportMitNutzlast(form, st[0]+z.text+st[1])
				for _, zu := range impZustaende() {
					faelle++
					pruefeReport(t, z.name+" in "+f.name, zu, f.roh, z.gueltig)
				}
			}
		}
	}
	gleichAuf(t, "Fälle laut Plan", faelle, 303*3*reportFormenZahl*3)
}

// TestIM11LaengeForm hält IM11 für die Menge Länge × Form × Zustand (SPEC-297 Nr. 6): jede Länge von importLaengen(100,
// 4194305) (Zweier- und Zehnerpotenzen um 1 bis 4 MiB + 1 Bytes) in jeder der 4 Formen von reportDerForm (Leerraum am Rand und
// zwischen den Token, `<`, `>`, `&`, U+2028 roh, Escapes, Mitglieder in absteigender Folge, `1.0`, `-0`, `1e400`) und in jedem
// der 3 Zustände kommt Byte für Byte an; ein Neuserialisieren, Trimmen oder Kappen ab einer Länge ändert Bytes. Zählprüfung:
// Längen × 4 Formen × 3 Zustände. Prüfgrenze (akzeptiertes Negativ (1)): die Längen dazwischen.
func TestIM11LaengeForm(t *testing.T) {
	faelle, laengen := 0, importLaengen(100, 4194305)
	for form := range reportFormenZahl {
		for _, n := range laengen {
			roh := reportDerForm(form, n)
			if len(roh) != n {
				t.Fatalf("Orakel: Form %d, Länge %d liefert %d Bytes", form, n, len(roh))
			}
			for _, zu := range impZustaende() {
				faelle++
				pruefeReport(t, fmt.Sprintf("Form %d, %d Bytes", form, n), zu, roh, true)
			}
		}
	}
	gleichAuf(t, "Längen und Fälle", []int{len(laengen), faelle}, []int{63, 63 * reportFormenZahl * 3})
}

// TestIM12Stelle hält IM12 für die Stelle des ungültigen Bytes (SPEC-297 Nr. 6): drei Folgen (rohes NUL, 0xFF, die
// Folge 0xED 0xA0 0x80 eines Surrogats) hinter p Buchstaben in einer Zeichenkette, für jedes p von importLaengen(0, 1 MiB + 1)
// (bis hinter 64 KiB und hinter 1 MiB), in der Zeichenkette allein und im verschachtelten Objekt, in jedem der 3 Zustände, ist
// 500 `ODR-E9001` ohne Teilantwort; eine Prüfung der ersten N Bytes lässt den Report mit dem Byte hinter N durch.
func TestIM12Stelle(t *testing.T) {
	faelle, stellen := 0, importLaengen(0, bodyGrenze+1)
	for _, folge := range []string{"\x00", "\xff", "\xed\xa0\x80"} {
		for _, p := range stellen {
			for _, form := range []int{0, 3} {
				f := reportMitNutzlast(form, strings.Repeat("a", p)+folge+"b")
				for _, zu := range impZustaende() {
					faelle++
					pruefeReport(t, fmt.Sprintf("%q an Stelle %d, %s", folge, p, f.name), zu, f.roh, false)
				}
			}
		}
	}
	gleichAuf(t, "Stellen und Fälle", []int{len(stellen), faelle}, []int{78, 3 * 78 * 2 * 3})
}

// importFehlerZellen liefert die Zellen von M-Fehler je Route: die 8 allgemeinen Ausgänge und die Codes der Tabelle von
// SPEC-297 Nr. 7 (`POST` 9, `GET` 3) als Literale.
func importFehlerZellen() map[string][]fehlerFall {
	gemeinsam := []fehlerFall{codeFall(model.CodeE3001, 500, "ODR-E3001", objektE3001), codeFall(model.CodeE9001, 500, "ODR-E9001", objekt500)}
	post := append([]fehlerFall{
		codeFall(model.CodeE5005, 404, "ODR-E5005", objektE5005), codeFall(model.CodeE7003, 404, "ODR-E7003", objektE7003),
		codeFall(model.CodeE5004, 409, "ODR-E5004", objektE5004), codeFall(model.CodeE7009, 409, "ODR-E7009", objektE7009),
		codeFall(model.CodeE6004, 422, "ODR-E6004", objektE6004), codeFall(model.CodeE6101, 422, "ODR-E6101", objektE6101),
		codeFall(model.CodeE6102, 409, "ODR-E6102", objektE6102),
	}, gemeinsam...)
	get := append([]fehlerFall{codeFall(model.CodeE7005, 404, "ODR-E7005", objektE7005)}, gemeinsam...)
	return map[string][]fehlerFall{"POST": append(allgemeineAusgaenge(), post...), "GET": append(allgemeineAusgaenge(), get...)}
}

// lesefehler ist eine Art von M-Lesen: der Fehler, den der Reader des Bodys hinter den Bytes liefert, und was SPEC-297 Nr. 5 a
// aus ihm macht: `ende` (das Ende des Bodys, kein Fehler), `lese` (ein Lesefehler, der kein Kontextfehler ist: 400) oder
// `kontext` (ein Kontextfehler, auch verpackt: die Verbindung bricht ab).
type lesefehler struct {
	name string
	err  error
	art  string
}

// importLesefehler sind die 13 Arten von M-Lesen: das Ende, 7 Lesefehler und 5 Kontextfehler (auch in `errors.Join` und mit
// `%w` verpackt, in beiden Folgen).
func importLesefehler() []lesefehler {
	return []lesefehler{
		{"EOF", io.EOF, "ende"},
		{"ErrUnexpectedEOF", io.ErrUnexpectedEOF, "lese"}, {"ErrClosedPipe", io.ErrClosedPipe, "lese"}, {"beliebiger Fehler", errors.New("boom"), "lese"},
		{"Fehler des Netzes", &net.OpError{Op: "read", Net: "tcp", Err: errors.New("connection reset by peer")}, "lese"},
		{"Frist des Netzes", os.ErrDeadlineExceeded, "lese"}, {"MaxBytesError", &http.MaxBytesError{Limit: bodyGrenze}, "lese"},
		{"verpackter Lesefehler", fmt.Errorf("x: %w", io.ErrUnexpectedEOF), "lese"},
		{"Canceled", context.Canceled, "kontext"}, {"DeadlineExceeded", context.DeadlineExceeded, "kontext"},
		{"Canceled und Lesefehler", errors.Join(context.Canceled, io.ErrUnexpectedEOF), "kontext"},
		{"Lesefehler und DeadlineExceeded", errors.Join(io.ErrUnexpectedEOF, context.DeadlineExceeded), "kontext"},
		{"verpacktes Canceled", fmt.Errorf("x: %w", context.Canceled), "kontext"},
	}
}

// lesenDaten liefert die Bytes vor dem Fehler an einer der vier Stellen von M-Lesen: keine, das Objekt ohne sein letztes Byte,
// das vollständige Objekt, das Objekt und ein Leerzeichen; das Objekt hat n Bytes (Leerraum vor der schließenden Klammer).
func lesenDaten(stelle, n int) []byte {
	obj := auffuellen(gutBody[:len(gutBody)-1], n-1)
	obj = append(obj, '}')
	return [][]byte{nil, obj[:n-1], obj, append(slices.Clone(obj), ' ')}[stelle]
}

// lesenLaengen sind die vier Längen von M-Lesen: das Objekt klein, 1 MiB - 1, 1 MiB und 1 MiB + 1 Bytes.
func lesenLaengen() []int { return []int{len(gutBody), bodyGrenze - 1, bodyGrenze, bodyGrenze + 1} }

// importRequestFuer bildet den Request der Route name: `POST` mit gültigem Body, `GET` mit einem Pfadwert.
func importRequestFuer(name string) *http.Request {
	if name == "POST" {
		return importPost("", nil, strings.NewReader(gutBody))
	}
	return importGet("imp-pfad")
}

// TestIM13Fehler hält IM13 (SPEC-297 Nr. 5 a, 7; SPEC-238 Nr. 6, 15 a): M-Fehler (8 allgemeine Ausgänge je Route und die Codes
// der Tabelle: `POST` 9, `GET` 3; 28 Zellen) mit Status, Code und Body als Literale, genau einem Zählen je Fehlerantwort und
// keinem bei einem Kontextfehler (Verbindung abgebrochen, nichts geschrieben); in jeder Zelle neben einem Wert des Dienstes
// (der Fehler gilt), mit genau einem Aufruf der Operation der Route und keinem von Audit, bei `Neu` wahr und falsch (je
// Zelle zwei Läufe: der Fehler hängt nicht von `Neu` ab, und der Fake zählt jede Operation, auch die, die die Route nie ruft).
func TestIM13Fehler(t *testing.T) {
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op := map[string]impOp{"POST": impImport, "GET": impGet}[route]
		for _, f := range importFehlerZellen()[route] {
			for _, neu := range []bool{true, false} {
				zellen++
				name := fmt.Sprintf("%s %s mit Neu %v", route, f.name, neu)
				u, fake, _ := importsUmgebung(t)
				fake.err, fake.ergebnis.Neu = f.err, neu
				rec := httptest.NewRecorder()
				r := importRequestFuer(route)
				abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, r) })
				codes, andere := zaehlungenVon(u.tele)
				var soll []string
				if f.abbruch {
					if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || rec.Result().StatusCode != 200 {
						t.Errorf("%s: Panik %v, Body %q: erwartet den Abbruch der Verbindung ohne Antwort", name, abbruch, rec.Body.String())
					}
				} else {
					soll = []string{f.code}
					if abbruch != nil {
						t.Errorf("%s: Panik %v", name, abbruch)
					}
					pruefeFehlerAntwort(t, name, rec, f.status, f.objekt, "")
				}
				if !slices.Equal(codes, soll) || andere != 0 || u.audit.Aufrufe() != 0 {
					t.Errorf("%s: CountError %q, andere Aufrufe %d, Audit %d, erwartet %q, 0 und 0", name, codes, andere, u.audit.Aufrufe(), soll)
				}
				fake.nurDiagonale(t, name, op)
			}
		}
	}
	gleichAuf(t, "Zellen ohne die Achse Lesen × Neu", zellen, 2*(2*8+9+3))
}

// TestIM13Lesen hält IM13 für die Achse *Lesen* (SPEC-297 Nr. 5 a, M-Lesen): jede der 13 Arten von importLesefehler (das Ende,
// 7 Lesefehler, 5 Kontextfehler) × 4 Stellen, an denen der Fehler hinter den Bytes kommt (vor dem ersten Byte, vor dem letzten
// Byte des Objekts, hinter dem vollständigen Objekt, hinter einem Leerzeichen dahinter) × 4 Längen des Objekts (klein, 1 MiB - 1,
// 1 MiB, 1 MiB + 1) = 208 Zellen. Orakel: wer mehr als 1 MiB Bytes liefert, ist 400 (die Grenze des Rahmens, welche Art auch
// folgt); sonst bricht ein Kontextfehler die Verbindung ab (kein Fehlerobjekt, kein Zählen, Import nicht gerufen, der Handler
// liefert den Fehler unverändert und nicht model.ErrUngueltig), ein Lesefehler ist 400 (auch hinter einem vollständigen Objekt),
// und das Ende gilt: 201 mit dem Objekt, wenn es vollständig ist, sonst 400. Audit sieht genau einen Aufruf in der Zelle mit
// 201 und sonst keinen (SPEC-297 Nr. 10 a; die Schleife steht hier, IM21 liest sie nicht).
func TestIM13Lesen(t *testing.T) {
	zellen := 0
	for _, lf := range importLesefehler() {
		for stelle := range 4 {
			for _, n := range lesenLaengen() {
				zellen++
				daten, name := lesenDaten(stelle, n), fmt.Sprintf("%s, Stelle %d, Länge %d", lf.name, stelle, n)
				u, fake, routen := importsUmgebung(t)
				rec := httptest.NewRecorder()
				p := abbruchVon(func() { u.srv.ServeHTTP(rec, importPost("", nil, &bodyLeser{daten: daten, rest: lf.err})) })
				codes, andere := zaehlungenVon(u.tele)
				soll, eintraege := 0, u.audit.Aufrufe() // Einträge laut SPEC-297 Nr. 10 a: allein der Erfolg schreibt einen; die Aufrufe des Handlers unten zählen nicht
				switch {
				case len(daten) <= bodyGrenze && lf.art == "kontext":
					if p != http.ErrAbortHandler || rec.Body.Len() != 0 || len(codes) != 0 || andere != 0 || fake.gesamt() != 0 {
						t.Errorf("%s: Panik %v, Body %q, Zählungen %q, %d Aufrufe: erwartet den Abbruch ohne Antwort", name, p, rec.Body.String(), codes, fake.gesamt())
					}
					if _, err := routen[0].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: &bodyLeser{daten: daten, rest: lf.err}}); err != lf.err {
						t.Errorf("%s: Handler liefert %v, erwartet den Fehler unverändert", name, err)
					}
				case len(daten) <= bodyGrenze && lf.art == "ende" && stelle >= 2:
					soll = 1
					if p != nil || rec.Code != 201 || fake.zahlen() != [impAnzahl]int{1, 0} || !gleicheAnfragen(fake.anfrageListe(), []inbound.ImportAnfrage{gutAnfrage()}) {
						t.Errorf("%s: Panik %v, Status %d, Zähler %v: erwartet 201 und einen Aufruf von Import", name, p, rec.Code, fake.zahlen())
					}
				default:
					pruefeFehlerAntwort(t, name, rec, 400, objekt400, "")
					if p != nil || !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 || fake.gesamt() != 0 {
						t.Errorf("%s: Panik %v, Zählungen %q, %d Aufrufe: erwartet 400 ohne Aufruf", name, p, codes, fake.gesamt())
					}
					if _, err := routen[0].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: &bodyLeser{daten: daten, rest: lf.err}}); len(daten) <= bodyGrenze && !errors.Is(err, model.ErrUngueltig) {
						t.Errorf("%s: Handler liefert %v, erwartet model.ErrUngueltig", name, err)
					}
				}
				if eintraege != soll {
					t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", name, eintraege, soll)
				}
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 13*4*4)
}

// TestIM13Zahlen hält M-Fehler und M-Lesen gegen die Quelle: 28 Zellen gleich 2 × 8 + 9 + 3 und 13 Arten des Lesens (das
// Ende, 7 Lesefehler, 5 Kontextfehler); jede der 10 Codezeilen der Tabelle von SPEC-297 Nr. 7, die ein Dienst liefert, hat eine
// Zelle (die Zeilen 1 und 2 stehen in IM7 bis IM9, die Zeile 3 in IM14), jeder Code steht in model.AlleCodes() mit dem Status
// des Literals, und die Literale tragen die Message der Tabelle der Codes.
func TestIM13Zahlen(t *testing.T) {
	z := importFehlerZellen()
	gleichAuf(t, "Zellen", len(z["POST"])+len(z["GET"]), 2*8+9+3)
	arten := map[string]int{}
	for _, lf := range importLesefehler() {
		arten[lf.art]++
	}
	gleichAuf(t, "Arten des Lesens", arten, map[string]int{"ende": 1, "lese": 7, "kontext": 5})
	alle := model.AlleCodes()
	gesehen := map[string]bool{}
	for route, fs := range z {
		for _, f := range fs {
			fe, ok := f.err.(*inbound.Fehler)
			if !ok || f.abbruch || f.mitWert || strings.HasPrefix(f.name, "verpackter") {
				continue
			}
			gesehen[f.code] = true
			c, bekannt := model.ParseCode(f.code)
			if !bekannt || !slices.Contains(alle, c) || c != fe.Code || !c.StatusErlaubt(f.status) || fe.Status != f.status {
				t.Errorf("%s %s: Code %q, Status %d passt nicht zur Tabelle der Codes", route, f.name, f.code, f.status)
			}
			if !strings.Contains(f.objekt, `"message":"`+c.Message()+`"`) {
				t.Errorf("%s %s: das Literal %q nennt die message %q nicht", route, f.name, f.objekt, c.Message())
			}
		}
	}
	for _, code := range []string{"ODR-E5005", "ODR-E7003", "ODR-E5004", "ODR-E7009", "ODR-E6004", "ODR-E6101", "ODR-E6102", "ODR-E3001", "ODR-E9001", "ODR-E7005"} {
		if !gesehen[code] {
			t.Errorf("die Tabelle von Nr. 7 nennt %s, M-Fehler hat keine Zelle dafür", code)
		}
	}
}

// TestIM13Unveraendert hält SPEC-297 Nr. 7 an der Route allein: der Handler gibt jeden Fehler des Dienstes unverändert zurück
// (dasselbe Objekt, nicht verpackt) und keine Antwort daneben, auch wenn der Dienst einen Wert und den Fehler liefert; die
// Allaussage führt jede Zelle von M-Fehler und dazu den nil-Zeiger `*inbound.Fehler` (2 Zusatzzellen, die der Plan nicht
// nennt: der Rahmen macht daraus 500 `ODR-E9001`, auch über den Server).
func TestIM13Unveraendert(t *testing.T) {
	var nilFehler *inbound.Fehler
	zellen := 0
	for i, route := range []string{"POST", "GET"} {
		faelle := append(importFehlerZellen()[route], fehlerFall{name: "nil-Zeiger", err: nilFehler, status: 500, code: "ODR-E9001", objekt: objekt500})
		for _, f := range faelle {
			zellen++
			u, fake, routen := importsUmgebung(t)
			fake.err = f.err
			ant, err := routen[i].Handler(context.Background(), admin.Anfrage{
				Methode: route, Body: strings.NewReader(gutBody), Pfadwerte: map[string]string{"id": "imp-pfad"},
			})
			if err != f.err || ant.Status != 0 || len(ant.Body) != 0 {
				t.Errorf("%s %s: Handler liefert %v und Antwort %d %q, erwartet denselben Fehler und keine Antwort", route, f.name, err, ant.Status, ant.Body)
			}
			if f.name == "nil-Zeiger" {
				rec := u.rufe(importRequestFuer(route))
				pruefeFehlerAntwort(t, route+" "+f.name, rec, 500, objekt500, "")
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 28+2)
}
