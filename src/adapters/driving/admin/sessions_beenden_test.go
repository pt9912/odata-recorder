package admin_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// berichtNull ist der Body von `DELETE` bei einem Bericht ohne Blobs und ohne Zurückstellen.
const berichtNull = `{"blobs_removed":0,"cleanup_deferred":false}`

// routeMitPfad sind die zwei Routen mit dem Pfadwert `{id}` (SPEC-292 Nr. 4): Name, Operation, Index in der Liste der Routen,
// Request zu einem kodierten Pfadwert und die Erfolgsantwort als Literal bei einem Dienst, der sessionA und den Bericht null
// liefert.
func routenMitPfad() []struct {
	name   string
	op     sessionsOp
	index  int
	req    func(id string) *http.Request
	erfolg string
} {
	return []struct {
		name   string
		op     sessionsOp
		index  int
		req    func(id string) *http.Request
		erfolg string
	}{
		{"close", opBeenden, 2, closeReq, sessionAJSON},
		{"DELETE", opDelete, 3, deleteReq, berichtNull},
	}
}

// TestB2AufrufeUndKontext hält B2 (SPEC-292 Nr. 2): `close` ruft genau Beenden und `DELETE` genau Delete, mit der Kennung des
// Pfades und dem ctx des Requests (2 Routen × 7 Zähler = 14 Zellen, 1 auf der Diagonale, sonst 0); der Wert im ctx des
// Requests kommt an, der Handler gibt den ctx ohne Ableitung weiter (Gleichheit der Identität); von audit läuft allein Append
// (bei `close` einmal, bei `DELETE` nie).
func TestB2AufrufeUndKontext(t *testing.T) {
	zellen := 0
	for _, f := range routenMitPfad() {
		u, fake, routen := sessionsUmgebung(t)
		fake.session = sessionA()
		r := f.req("ses-pfad")
		r = r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, "wert-"+f.name))
		if rec := u.rufe(r); rec.Code != 200 {
			t.Fatalf("%s: Status %d", f.name, rec.Code)
		}
		zellen += len(fake.zahlen())
		fake.nurDiagonale(t, f.name, f.op)
		gleichAuf(t, f.name+": Argumente", fake.argumente(), [][2]string{{"ses-pfad", ""}})
		if got := fake.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-"+f.name {
			t.Errorf("%s: der Wert im ctx des Requests kam als %v an", f.name, got)
		}
		if n, soll := u.audit.Aufrufe(), map[string]int{"close": 1, "DELETE": 0}[f.name]; n != soll {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", f.name, n, soll)
		}

		ctx := context.WithValue(context.Background(), ctxSchluessel{}, "direkt")
		_, _ = routen[f.index].Handler(ctx, admin.Anfrage{Methode: "X", Pfadwerte: map[string]string{"id": "ses-direkt"}, KorrelationsID: "cor-1"})
		if got := fake.ctxListe(); len(got) != 2 || got[1] != ctx {
			t.Errorf("%s: der Handler reichte einen anderen ctx weiter als den, den er erhielt", f.name)
		}
		gleichAuf(t, f.name+": Argumente des Handlers", fake.argumente(), [][2]string{{"ses-pfad", ""}, {"ses-direkt", ""}})
	}
	gleichAuf(t, "Zellen (Routen × Zähler)", zellen, 2*int(opAnzahl))
}

// pfadwerte ist M-Pfadwert (SPEC-292 Nr. 4): 13 kodierte Werte mit den dekodierten Bytes als Literale aus Bytewerten.
func pfadwerte() []struct{ kodiert, soll string } {
	bytesAus := func(b ...byte) string { return string(b) }
	return []struct{ kodiert, soll string }{
		{"x", "x"},
		{"ses-" + strings.Repeat("0123456789", 3) + "01", "ses-" + strings.Repeat("0123456789", 3) + "01"},
		{"SES-1", "SES-1"},
		{"a.b", "a.b"},
		{".", "."},
		{"..", ".."},
		{"a%2Fb", "a/b"},
		{"%C3%A4", bytesAus(0xC3, 0xA4)},
		{"%FF", bytesAus(0xFF)},
		{"%00", bytesAus(0x00)},
		{strings.Repeat("a", 65), strings.Repeat("a", 65)},
		{"a+b", "a+b"},
		{"%20", bytesAus(0x20)},
	}
}

// TestB3Pfadwert hält B3 (SPEC-292 Nr. 4): jeder der 13 Werte (Form, Länge, Schreibweise, `.`, `..`, Prozentfolgen, `+`) geht
// byte-gleich an den Dienst, bei `close` und `DELETE` (26 Fälle), die Antwort ist die Erfolgsantwort des Dienstes ohne den
// Wert, und der Fehler `ODR-E7003` steht ohne den Wert (jedes Byte der Antwort ist das Literal); ein Canary als Wert steht in
// keinem Byte von Status, Header und Body. Zählprüfung: 13 Werte und die zwei Routen mit `{id}`.
func TestB3Pfadwert(t *testing.T) {
	faelle := 0
	for _, f := range routenMitPfad() {
		for _, w := range pfadwerte() {
			faelle++
			name := f.name + " " + w.kodiert
			u, fake, _ := sessionsUmgebung(t)
			fake.session = sessionA()
			rec := u.rufe(f.req(w.kodiert))
			if rec.Code != 200 || rec.Body.String() != f.erfolg {
				t.Errorf("%s: Antwort %d %q, erwartet 200 %q", name, rec.Code, rec.Body.String(), f.erfolg)
			}
			gleichAuf(t, name+": Argumente", fake.argumente(), [][2]string{{w.soll, ""}})
			fake.nurDiagonale(t, name, f.op)

			u, fake, _ = sessionsUmgebung(t)
			fake.err = &inbound.Fehler{Code: model.CodeE7003, Status: 404}
			rec = u.rufe(f.req(w.kodiert))
			pruefeFehlerAntwort(t, name+" unbekannt", rec, 404, objektE7003, "")
			gleichAuf(t, name+": Argumente beim Fehler", fake.argumente(), [][2]string{{w.soll, ""}})
		}
		for _, fehler := range []error{nil, &inbound.Fehler{Code: model.CodeE7003, Status: 404}} {
			u, fake, _ := sessionsUmgebung(t)
			fake.session, fake.err = sessionA(), fehler
			rec := u.rufe(f.req(canary()))
			if strings.Contains(antwortText(rec), canary()) {
				t.Errorf("%s: der Canary steht in der Antwort: %s", f.name, antwortText(rec))
			}
			gleichAuf(t, f.name+": Argumente mit Canary", fake.argumente(), [][2]string{{canary(), ""}})
		}
	}
	gleichAuf(t, "Fälle, Werte, Routen mit {id}", []int{faelle, len(pfadwerte()), len(routenMitPfad())}, []int{26, 13, 2})
}

// TestB3JedesByte hält B3 für die Zeichenmenge des Pfadwertes (der Plan führt 13 Werte, die Menge der Bytes ist die Klasse
// der Zusage „ohne Prüfung von Form“): jedes der 256 Bytes allein und zwischen zwei Buchstaben (`a<Byte>z`), als `%XX`
// kodiert, geht byte-gleich an den Dienst, bei `close` und `DELETE`; das Orakel ist die Zählschleife, nicht der Code der
// Route. Zählprüfung: 256 Bytes × 2 Formen × 2 Routen = 1024 Fälle.
func TestB3JedesByte(t *testing.T) {
	faelle := 0
	for _, f := range routenMitPfad() {
		for b := range 256 {
			for _, form := range []string{"%s", "a%sz"} {
				kodiert := strings.Replace(form, "%s", fmt.Sprintf("%%%02X", b), 1)
				soll := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
				u, fake, _ := sessionsUmgebung(t)
				fake.session = sessionA()
				rec := u.rufe(f.req(kodiert))
				faelle++
				if rec.Code != 200 || rec.Body.String() != f.erfolg {
					t.Errorf("%s Byte 0x%02X in der Form %q: Antwort %d %q, erwartet 200", f.name, b, form, rec.Code, rec.Body.String())
				}
				gleichAuf(t, fmt.Sprintf("%s Byte 0x%02X in der Form %q: Argumente", f.name, b, form), fake.argumente(), [][2]string{{soll, ""}})
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 256*2*len(routenMitPfad()))
	gleichAuf(t, "Fälle laut Plan", faelle, 1024)
}

// TestB4QueryUndBody hält B4 (SPEC-292 Nr. 3): weder `close` noch `DELETE` liest die Query, jede Form (6 Queries × 2 Routen =
// 12 Fälle) ist ohne Wirkung (Status 200, Antwort und Argumente wie ohne Query); beide lesen den Body nie (3 ungültige Bodies
// × 2 Routen = 6 Fälle, 0 Lesezugriffe). Zählprüfung: 18.
func TestB4QueryUndBody(t *testing.T) {
	queries := []string{"", "limit=x", "a=1&a=2", "%zz", "%", "a=1;b=2"}
	bodies := [][]byte{[]byte("{"), []byte("[1,2"), {0xFF, 0xFE, 0x00}}
	faelle := 0
	for _, f := range routenMitPfad() {
		pfad := map[string]string{"close": sessionsZiel + "/ses-pfad/close", "DELETE": sessionsZiel + "/ses-pfad"}[f.name]
		methode := map[string]string{"close": "POST", "DELETE": "DELETE"}[f.name]
		sende := func(ziel string, body io.Reader) (*httptest.ResponseRecorder, *sessionsFake) {
			u, fake, _ := sessionsUmgebung(t)
			fake.session = sessionA()
			return u.rufe(anfrageMit(methode, ziel, nil, body)), fake
		}
		for _, q := range queries {
			ziel := pfad
			if q != "" {
				ziel += "?" + q
			}
			rec, fake := sende(ziel, nil)
			if rec.Code != 200 || rec.Body.String() != f.erfolg {
				t.Errorf("%s mit Query %q: %d %q, erwartet 200 %q", f.name, q, rec.Code, rec.Body.String(), f.erfolg)
			}
			gleichAuf(t, f.name+" mit Query "+q+": Argumente", fake.argumente(), [][2]string{{"ses-pfad", ""}})
			fake.nurDiagonale(t, f.name+" mit Query "+q, f.op)
			faelle++
		}
		for i, b := range bodies {
			leser := &zaehlLeser{Reader: bytes.NewReader(b)}
			rec, fake := sende(pfad, leser)
			if rec.Code != 200 || rec.Body.String() != f.erfolg || leser.n.Load() != 0 || fake.gesamt() != 1 {
				t.Errorf("%s mit ungültigem Body Nr. %d: Status %d, %d Lesezugriffe, %d Aufrufe", f.name, i, rec.Code, leser.n.Load(), fake.gesamt())
			}
			faelle++
		}
	}
	gleichAuf(t, "Fälle", faelle, len(routenMitPfad())*(len(queries)+len(bodies)))
	gleichAuf(t, "Fälle laut Plan", faelle, 18)
}

// setzeFeld setzt das Feld i des Wertes v auf einen Wert ungleich dem Nullwert: eine Zeichenkette, eine Zeit, eine Zahl oder
// ein Wahrheitswert (nach dem Typ des Feldes).
func setzeFeld(t *testing.T, v reflect.Value, i int, text string, zeit time.Time) {
	t.Helper()
	f := v.Field(i)
	switch {
	case f.Type() == reflect.TypeOf(time.Time{}):
		f.Set(reflect.ValueOf(zeit))
	case f.Kind() == reflect.String:
		f.SetString(text)
	case f.Kind() == reflect.Int:
		f.SetInt(7)
	case f.Kind() == reflect.Bool:
		f.SetBool(true)
	default:
		t.Fatalf("Feld %d hat den Typ %v", i, f.Type())
	}
}

// TestB5FelderDerSession hält B5 (SPEC-292 Nr. 6, 7 b): `close` antwortet 200 mit den vier Feldern von model.Session in der
// Folge id, service_id, created_at, status, aus dem Ergebnis des Dienstes (nicht aus dem Pfad), jedes mit eigenem Wert, für
// `complete`, `incomplete` und einen unbekannten Status, mit den Headern der Erfolgsantwort; dazu M-Felder: je Feld von
// model.Session ein Lauf, in dem allein dieses Feld belegt ist (der Rest steht leer, auch die Zeit als `null`). Zählprüfung
// gegen reflect: 4 Felder.
func TestB5FelderDerSession(t *testing.T) {
	gleichAuf(t, "Felder von model.Session", reflect.TypeOf(model.Session{}).NumField(), 4)
	for _, status := range []string{"complete", "incomplete", "x"} {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = model.Session{ID: "ses-dienst", ServiceID: "svc-dienst", CreatedAt: zeitA(), Status: model.SessionStatus(status)}
		rec := u.rufe(closeReq("ses-pfad"))
		soll := `{"id":"ses-dienst","service_id":"svc-dienst","created_at":"2026-11-06T10:00:00.123456789Z","status":"` + status + `"}`
		if rec.Code != 200 || rec.Body.String() != soll {
			t.Errorf("Status %q: Antwort %d %q, erwartet 200 und %q", status, rec.Code, rec.Body.String(), soll)
		}
		pruefeErfolgsHeader(t, "Status "+status, rec)
	}
	soll := map[string]string{
		"ID":        `{"id":"w-id","service_id":"","created_at":null,"status":""}`,
		"ServiceID": `{"id":"","service_id":"w-id","created_at":null,"status":""}`,
		"CreatedAt": `{"id":"","service_id":"","created_at":"2026-11-06T10:00:00.123456789Z","status":""}`,
		"Status":    `{"id":"","service_id":"","created_at":null,"status":"w-id"}`,
	}
	felder := 0
	for i := range reflect.TypeOf(model.Session{}).NumField() {
		name := reflect.TypeOf(model.Session{}).Field(i).Name
		var s model.Session
		setzeFeld(t, reflect.ValueOf(&s).Elem(), i, "w-id", zeitA())
		u, fake, _ := sessionsUmgebung(t)
		fake.session = s
		if rec := u.rufe(closeReq("ses-pfad")); rec.Code != 200 || rec.Body.String() != soll[name] {
			t.Errorf("Feld %s: %d %q, erwartet 200 %q", name, rec.Code, rec.Body.String(), soll[name])
		}
		felder++
	}
	gleichAuf(t, "Felder und Literale", []int{felder, len(soll)}, []int{4, 4})
}

// TestB6WartetAufBeenden hält B6 (SPEC-292 Nr. 9 a): die Antwort von `close` entsteht erst nach der Rückkehr von Beenden.
// Wirkungs-Fänger mit Kanälen: während Beenden blockiert, ist ServeHTTP nicht zurückgekehrt (Frist 150 ms für das, was nicht
// geschehen darf), weder WriteHeader noch Write ist gelaufen, und der ctx trägt keine Frist; nach der Freigabe ist die
// Antwort 200 mit demselben Body wie ohne Wartezeit. Das fehlende `WriteTimeout` fängt `grenzen_test.go` des Rahmens.
func TestB6WartetAufBeenden(t *testing.T) {
	ohne, ohneFake, _ := sessionsUmgebung(t)
	ohneFake.session = sessionA()
	ohneRec := ohne.rufe(closeReq("ses-pfad"))
	u, fake, _ := sessionsUmgebung(t)
	fake.session = sessionA()
	eintritt, frei := make(chan struct{}), make(chan struct{})
	var mitFrist atomic.Bool
	fake.vor[opBeenden] = func(ctx context.Context) {
		if _, ok := ctx.Deadline(); ok {
			mitFrist.Store(true)
		}
		close(eintritt)
		select {
		case <-frei:
		case <-time.After(10 * time.Second):
		}
	}
	var e ereignisse
	rec := httptest.NewRecorder()
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, closeReq("ses-pfad"))
	}()
	select {
	case <-eintritt:
	case <-time.After(10 * time.Second):
		t.Fatal("Beenden wurde nicht gerufen")
	}
	select {
	case <-fertig:
		t.Error("ServeHTTP kehrte zurück, solange Beenden lief")
	case <-time.After(150 * time.Millisecond):
	}
	if ev := e.lies(); len(ev) != 0 || rec.Body.Len() != 0 {
		t.Errorf("während Beenden: Ereignisse %q, Body %d Bytes, erwartet nichts", ev, rec.Body.Len())
	}
	close(frei)
	select {
	case <-fertig:
	case <-time.After(10 * time.Second):
		t.Fatal("ServeHTTP kehrte nach der Freigabe nicht zurück")
	}
	if rec.Code != 200 || rec.Body.String() != sessionAJSON || rec.Body.String() != ohneRec.Body.String() {
		t.Errorf("nach der Freigabe: %d %q, erwartet 200 %q", rec.Code, rec.Body.String(), sessionAJSON)
	}
	gleichAuf(t, "Ereignisse nach der Freigabe", e.lies(), []string{"kopf", "schreiben"})
	if mitFrist.Load() {
		t.Error("der ctx von Beenden trägt eine Frist")
	}
}

// TestB7KontextEndetWaehrendBeenden hält B7 (SPEC-292 Nr. 9 c, 11 a): endet der ctx des Requests während Beenden, sieht der
// Dienst das Ende (kein WithoutCancel: ctx.Done() ist geschlossen), liefert er die Session, so antwortet die Route 200 und
// schreibt einen Eintrag; liefert er einen Kontextfehler (Canceled oder DeadlineExceeded, 2 Läufe), bricht die Verbindung ab,
// ohne Antwort und ohne Eintrag. Der Wert im ctx kommt an.
func TestB7KontextEndetWaehrendBeenden(t *testing.T) {
	start := func() (*umgebung, *sessionsFake, *http.Request, *atomic.Bool) {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "wert-b7"))
		sah := &atomic.Bool{}
		fake.vor[opBeenden] = func(c context.Context) {
			cancel()
			select {
			case <-c.Done():
				sah.Store(true)
			case <-time.After(5 * time.Second):
			}
		}
		return u, fake, closeReq("ses-pfad").WithContext(ctx), sah
	}
	u, fake, r, sah := start()
	rec := u.rufe(r)
	if rec.Code != 200 || rec.Body.String() != sessionAJSON || !sah.Load() || u.audit.Aufrufe() != 1 {
		t.Errorf("Ende des ctx, Erfolg: %d %q, Dienst sah das Ende: %v, %d Einträge; erwartet 200, ja, 1", rec.Code, rec.Body.String(), sah.Load(), u.audit.Aufrufe())
	}
	if got := fake.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-b7" {
		t.Errorf("der Wert im ctx kam als %v an", got)
	}
	for _, ce := range []error{context.Canceled, context.DeadlineExceeded} {
		u, fake, r, _ := start()
		fake.err = ce
		rec := httptest.NewRecorder()
		abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, r) })
		if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || u.audit.Aufrufe() != 0 {
			t.Errorf("Kontextfehler %v: Panik %v, Body %q, %d Einträge; erwartet den Abbruch ohne Antwort und ohne Eintrag", ce, abbruch, rec.Body.String(), u.audit.Aufrufe())
		}
	}
}

// TestB8Loeschbericht hält B8 (SPEC-292 Nr. 6, 10): `DELETE` antwortet 200 mit `{"blobs_removed":<Entfernt>,
// "cleanup_deferred":<Zurueckgestellt>}` in dieser Folge: M-Bericht, 6 Werte als Literale (der größte `int` darunter), mit den
// Headern der Erfolgsantwort; dazu M-Felder: je Feld von inbound.Loeschbericht ein Lauf, in dem allein dieses Feld belegt ist
// (fängt vertauschte Felder). Zählprüfung gegen reflect: 2 Felder.
func TestB8Loeschbericht(t *testing.T) {
	gleichAuf(t, "Felder von inbound.Loeschbericht", reflect.TypeOf(inbound.Loeschbericht{}).NumField(), 2)
	groesst := strconv.Itoa(math.MaxInt)
	werte := []struct {
		b    inbound.Loeschbericht
		soll string
	}{
		{inbound.Loeschbericht{}, berichtNull},
		{inbound.Loeschbericht{Entfernt: 3}, `{"blobs_removed":3,"cleanup_deferred":false}`},
		{inbound.Loeschbericht{Zurueckgestellt: true}, `{"blobs_removed":0,"cleanup_deferred":true}`},
		{inbound.Loeschbericht{Entfernt: 7, Zurueckgestellt: true}, `{"blobs_removed":7,"cleanup_deferred":true}`},
		{inbound.Loeschbericht{Entfernt: math.MaxInt}, `{"blobs_removed":` + groesst + `,"cleanup_deferred":false}`},
		{inbound.Loeschbericht{Entfernt: 1, Zurueckgestellt: true}, `{"blobs_removed":1,"cleanup_deferred":true}`},
	}
	for _, w := range werte {
		u, fake, _ := sessionsUmgebung(t)
		fake.bericht = w.b
		rec := u.rufe(deleteReq("ses-pfad"))
		if rec.Code != 200 || rec.Body.String() != w.soll {
			t.Errorf("Bericht %+v: %d %q, erwartet 200 %q", w.b, rec.Code, rec.Body.String(), w.soll)
		}
		pruefeErfolgsHeader(t, "Bericht", rec)
	}
	gleichAuf(t, "Zahl der Werte", len(werte), 6)
	soll := map[string]string{
		"Entfernt":        `{"blobs_removed":7,"cleanup_deferred":false}`,
		"Zurueckgestellt": `{"blobs_removed":0,"cleanup_deferred":true}`,
	}
	for i := range reflect.TypeOf(inbound.Loeschbericht{}).NumField() {
		name := reflect.TypeOf(inbound.Loeschbericht{}).Field(i).Name
		var b inbound.Loeschbericht
		setzeFeld(t, reflect.ValueOf(&b).Elem(), i, "", time.Time{})
		u, fake, _ := sessionsUmgebung(t)
		fake.bericht = b
		if rec := u.rufe(deleteReq("ses-pfad")); rec.Code != 200 || rec.Body.String() != soll[name] {
			t.Errorf("Feld %s: %d %q, erwartet 200 %q", name, rec.Code, rec.Body.String(), soll[name])
		}
	}
}
