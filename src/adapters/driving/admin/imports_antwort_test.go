package admin_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

// sende stellt den Dienst auf den Import und sendet den Request des Zustands.
func (z impZustand) sende(u *umgebung, fake *importsFake, imp model.Import) *httptest.ResponseRecorder {
	fake.ergebnis, fake.imp = inbound.ImportErgebnis{Import: imp, Neu: z.neu}, imp
	if z.post {
		return u.rufe(importPost("", nil, strings.NewReader(gutBody)))
	}
	return u.rufe(importGet("imp-pfad"))
}

// TestIM10Status hält IM10 (SPEC-297 Nr. 6, SPEC-238 Nr. 5): `POST` antwortet 201, wenn `Neu` wahr ist, sonst 200, `GET` 200;
// die Erfolgsantwort trägt `Content-Type`, `Content-Length` und `X-Correlation-Id` und sonst keinen Header (kein `Location`);
// der Status hängt allein von `Neu` ab: jedes der 6 Felder von model.Import mit eigenem Wert ändert in den 2 Zellen von `POST`
// (12 Fälle) Status und Body nicht (der Body folgt allein dem Feld `ReportJSON`). Zählprüfung: 3 + 12.
func TestIM10Status(t *testing.T) {
	for _, z := range impZustaende() {
		u, fake, _ := importsUmgebung(t)
		rec := z.sende(u, fake, importA())
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
			rec := z.sende(u, fake, imp)
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
			if rec := z.sende(u, fake, imp); rec.Code != z.status || rec.Body.String() != reportA {
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
			rec := z.sende(u, fake, imp)
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
			erster := impZustaende()[0].sende(u, fake, imp)
			zweite := zweiter.sende(u, fake, imp)
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
			rec := z.sende(u, fake, imp)
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

// importLesefehler sind die 3 Zellen der Achse *Lesen* von M-Fehler (nur `POST`): der Reader des Bodys liefert den Kontextfehler.
func importLesefehler() []error {
	return []error{context.Canceled, context.DeadlineExceeded, errors.Join(context.Canceled, io.ErrUnexpectedEOF)}
}

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
// (der Fehler gilt), mit genau einem Aufruf der Operation der Route und keinem von Audit.
func TestIM13Fehler(t *testing.T) {
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op := map[string]impOp{"POST": impImport, "GET": impGet}[route]
		for _, f := range importFehlerZellen()[route] {
			zellen++
			name := route + " " + f.name
			u, fake, _ := importsUmgebung(t)
			fake.err = f.err
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
	gleichAuf(t, "Zellen ohne die Achse Lesen", zellen, 2*8+9+3)
}

// TestIM13Lesen hält IM13 für die Achse *Lesen* (SPEC-297 Nr. 5 a): ein Kontextfehler des Readers (3 Zellen) bricht die
// Verbindung ab, ohne Fehlerobjekt, ohne Zählen, ohne Aufruf von Import und ohne Eintrag; der Handler liefert den Fehler
// unverändert und nicht model.ErrUngueltig.
func TestIM13Lesen(t *testing.T) {
	gleichAuf(t, "Zellen der Achse Lesen", len(importLesefehler()), 3)
	for _, ce := range importLesefehler() {
		u, fake, routen := importsUmgebung(t)
		leser := &bodyLeser{daten: []byte(`{"import_id":"imp-1"`), rest: ce}
		_, err := routen[0].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: leser})
		if err != ce || errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Handler liefert %v, erwartet %v unverändert und nicht model.ErrUngueltig", err, ce)
		}
		rec := httptest.NewRecorder()
		leser = &bodyLeser{daten: []byte(`{"import_id":"imp-1"`), rest: ce}
		if p := abbruchVon(func() { u.srv.ServeHTTP(rec, importPost("", nil, leser)) }); p != http.ErrAbortHandler {
			t.Errorf("%v: ServeHTTP endete mit %v, erwartet den Abbruch der Verbindung", ce, p)
		}
		codes, andere := zaehlungenVon(u.tele)
		if fake.gesamt() != 0 || u.audit.Aufrufe() != 0 || len(codes) != 0 || andere != 0 || rec.Body.Len() != 0 {
			t.Errorf("%v: %d Aufrufe des Dienstes, %d von Audit, Zählungen %q, Body %q", ce, fake.gesamt(), u.audit.Aufrufe(), codes, rec.Body.String())
		}
	}
}

// TestIM13Zahlen hält M-Fehler gegen die Quelle: 31 Zellen gleich 2 × 8 + 9 + 3 + 3; jede der 10 Codezeilen der Tabelle von
// SPEC-297 Nr. 7, die ein Dienst liefert, hat eine Zelle (die Zeilen 1 und 2 stehen in IM7 bis IM9, die Zeile 3 in IM14), jeder
// Code steht in model.AlleCodes() mit dem Status des Literals, und die Literale tragen die Message der Tabelle der Codes.
func TestIM13Zahlen(t *testing.T) {
	z := importFehlerZellen()
	gleichAuf(t, "Zellen", len(z["POST"])+len(z["GET"])+len(importLesefehler()), 31)
	gleichAuf(t, "Zellen gleich 2 × 8 + 9 + 3 + 3", len(z["POST"])+len(z["GET"])+len(importLesefehler()), 2*8+9+3+3)
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
