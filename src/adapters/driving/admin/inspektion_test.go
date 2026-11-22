package admin_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Die Bodies der Erfolgsantworten ohne Inhalt, als Literale des Vertrags (SPEC-293 Nr. 6 bis 8): ein Dienst ohne Eintrag.
const (
	inspLeereSeite      = `{"exchanges":[],"limit":0,"total":0,"next":null}`
	inspLeereCoverage   = `{"entries":[]}`
	inspLeereDiagnose   = `{"mode":"","settings":{},"errors":[]}`
	inspExchangesMuster = "/__admin/sessions/{id}/exchanges"
)

// TestI1Konstruktor hält I1 (SPEC-293 Nr. 1): der Gutfall liefert genau die drei Routen in der Folge des Vertrags, ein
// nil-Port ist model.ErrUngueltig ohne Routen, die Funktion hat einen Parameter vom Typ des Ports (weder AuditLog noch
// Telemetry), ruft keinen Port, und die Zahl der Goroutinen wächst nach 100 Aufrufen um höchstens 10; eine kurzlebige Goroutine fängt allein die Quelle
// (TestI17QuelleDerInspektion).
func TestI1Konstruktor(t *testing.T) {
	fake := &inspektionFake{}
	vorher := runtime.NumGoroutine()
	var routen []admin.Route
	var err error
	for range 100 {
		routen, err = admin.InspektionRouten(fake)
	}
	if nachher := runtime.NumGoroutine(); err != nil || nachher > vorher+10 {
		t.Errorf("InspektionRouten: %v, Goroutinen vorher %d, nachher %d", err, vorher, nachher)
	}
	got := [][2]string{}
	for _, r := range routen {
		if r.Handler == nil {
			t.Errorf("Route %s %s ohne Handler", r.Methode, r.Muster)
		}
		got = append(got, [2]string{r.Methode, r.Muster})
	}
	gleichAuf(t, "Routen in der Folge des Vertrags", got,
		[][2]string{{"GET", inspExchangesMuster}, {"GET", "/__admin/coverage"}, {"GET", "/__admin/diagnostics"}})
	if fake.gesamt() != 0 {
		t.Errorf("InspektionRouten rief %d-mal den Dienst, erwartet 0", fake.gesamt())
	}

	routen, err = admin.InspektionRouten(nil)
	if routen != nil || !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("nil: %v Routen, Fehler %v, erwartet keine und model.ErrUngueltig", routen, err)
	}
	if text := fmt.Sprintf("%v %+v %#v", err, err, err); strings.Contains(text, "<nil>") || strings.Contains(text, "%!") {
		t.Errorf("der Text des Fehlers nennt einen Wert: %q", text)
	}

	ft := reflect.TypeOf(admin.InspektionRouten)
	gleichAuf(t, "Zahl der Parameter und Ergebnisse", []int{ft.NumIn(), ft.NumOut()}, []int{1, 2})
	if ft.In(0) != reflect.TypeFor[inbound.InspectionUseCase]() {
		t.Errorf("der Parameter ist %v, erwartet der Port inbound.InspectionUseCase", ft.In(0))
	}
}

// inspAufrufFall ist ein Request an eine der drei Routen mit dem Orakel seiner Argumente und der Anfrage, die ein Test dem
// Handler direkt übergibt.
type inspAufrufFall struct {
	name, pfad, query string
	op                inspOp
	id                string
	seite             inbound.Seite
	anfrage           admin.Anfrage
}

func inspAufrufFaelle() []inspAufrufFall {
	return []inspAufrufFall{
		{"exchanges", exchangesZiel("ses-x"), "limit=3&after=5", inspOpExchanges, "ses-x", inbound.Seite{Limit: 3, Nach: 5},
			admin.Anfrage{RawQuery: "limit=1&after=2", Pfadwerte: map[string]string{"id": "ses-d"}}},
		{"coverage", inspCoverage, "baseline_id=bl-x", inspOpCoverage, "bl-x", inbound.Seite{},
			admin.Anfrage{RawQuery: "baseline_id=bl-d"}},
		{"diagnose", inspDiagnostics, "", inspOpDiagnose, "", inbound.Seite{}, admin.Anfrage{}},
	}
}

// TestI2AufrufeUndKontext hält I2 (SPEC-293 Nr. 2): jede der drei Routen ruft genau eine Operation von inspektion (3 Routen ×
// 3 Zähler = 9 Zellen, 1 auf der Diagonale, sonst 0) mit den Argumenten unverändert (auch Seite); der Wert im ctx des Requests
// kommt an, und der Handler gibt ctx ohne Ableitung weiter (Gleichheit der Identität); die Route schreibt keinen Eintrag.
func TestI2AufrufeUndKontext(t *testing.T) {
	zellen := 0
	for i, f := range inspAufrufFaelle() {
		u, fake, routen := inspUmgebung(t)
		r := inspReq(f.pfad, f.query)
		r = r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, "wert-"+f.name))
		if rec := u.rufe(r); rec.Code != 200 {
			t.Fatalf("%s: Status %d", f.name, rec.Code)
		}
		zellen += len(fake.zahlen())
		fake.nurDiagonale(t, f.name, f.op)
		args := fake.argumente()
		if len(args) != 1 || args[0].op != f.op || args[0].id != f.id || args[0].seite != f.seite {
			t.Errorf("%s: Argumente %+v, erwartet Operation %d, id %q, Seite %+v", f.name, args, f.op, f.id, f.seite)
		}
		if got := fake.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-"+f.name {
			t.Errorf("%s: der Wert im ctx des Requests kam als %v an", f.name, got)
		}
		if n := u.audit.Aufrufe(); n != 0 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 0", f.name, n)
		}

		ctx := context.WithValue(context.Background(), ctxSchluessel{}, "direkt")
		_, _ = routen[i].Handler(ctx, f.anfrage)
		if got := fake.ctxListe(); len(got) != 2 || got[1] != ctx {
			t.Errorf("%s: der Handler reichte einen anderen ctx weiter als den, den er erhielt", f.name)
		}
	}
	gleichAuf(t, "Zellen (Routen × Zähler)", zellen, 3*int(inspOpAnzahl))
	gleichAuf(t, "Zähler gleich Zahl der Operationen des Ports", int(inspOpAnzahl), reflect.TypeFor[inbound.InspectionUseCase]().NumMethod())
}

// inspLeser zählt jeden Aufruf von Read und liefert einen Lesefehler.
type inspLeser struct{ n int }

func (l *inspLeser) Read([]byte) (int, error) {
	l.n++
	return 0, errors.New("Lesefehler")
}

// TestI2BodyUngelesen hält SPEC-293 Nr. 2 („der Body der drei Routen wird nicht gelesen und nicht geprüft“): ein Body, der bei jedem Lesen
// scheitert, und ein fremder `Content-Type` ändern nichts, kein Aufruf von Read (3 Routen, 3 Fälle).
func TestI2BodyUngelesen(t *testing.T) {
	for _, f := range inspAufrufFaelle() {
		u, fake, _ := inspUmgebung(t)
		body := &inspLeser{}
		ziel := f.pfad
		if f.query != "" {
			ziel += "?" + f.query
		}
		rec := u.rufe(anfrageMit("GET", ziel, map[string][]string{"Content-Type": {"text/x-fremd"}}, body))
		if rec.Code != 200 || body.n != 0 || fake.gesamt() != 1 {
			t.Errorf("%s: Status %d, %d Lesezugriffe, %d Aufrufe des Dienstes, erwartet 200, 0 und 1", f.name, rec.Code, body.n, fake.gesamt())
		}
	}
}

// TestI3Zuordnung hält I3 (SPEC-293 Nr. 1, SPEC-238 Nr. 2): M-Route der Inspektion, 3 Pfadformen × 7 Methoden = 21 Fälle:
// `GET` trifft (3), die 6 anderen Methoden sind je Pfadform 405 mit `Allow: GET` (18) und rufen den Dienst nie, ohne Eintrag.
func TestI3Zuordnung(t *testing.T) {
	pfade := []string{exchangesZiel("ses-1"), inspCoverage + "?baseline_id=b", inspDiagnostics}
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	faelle, treffer := 0, 0
	for _, p := range pfade {
		for _, m := range methoden {
			faelle++
			u, fake, _ := inspUmgebung(t)
			rec := u.rufe(anfrageMit(m, p, nil, nil))
			if m == "GET" {
				treffer++
				if rec.Code != 200 || fake.gesamt() != 1 {
					t.Errorf("%s %s: Status %d, %d Aufrufe des Dienstes", m, p, rec.Code, fake.gesamt())
				}
				continue
			}
			pruefeFehlerAntwort(t, m+" "+p, rec, 405, objekt404, "GET")
			if codes, _ := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || fake.gesamt() != 0 || u.audit.Aufrufe() != 0 {
				t.Errorf("%s %s: CountError %q, %d Aufrufe des Dienstes, %d von Audit", m, p, codes, fake.gesamt(), u.audit.Aufrufe())
			}
		}
	}
	_, _, routen := inspUmgebung(t)
	gleichAuf(t, "Fälle, Treffer, Pfadformen, Routen", []int{faelle, treffer, len(pfade), len(routen)}, []int{21, 3, 3, 3})
}

// TestI3GesamtzahlDerPfadformen hält die Zählprüfung von M-Route gegen die Muster der beiden Konstruktoren: die Routen von
// SessionRouten und InspektionRouten überlappen einander nicht (New nimmt sie zusammen an, SPEC-238 Nr. 2 d), und die
// Pfadformen mal sieben Methoden sind die Fälle von I3 über alle Teile, die Treffer die Paare aus Methode und Muster. Stand mit
// den Routen aller drei Teile: 6 Pfadformen, 42 Fälle, 7 Treffer, 7 Routen.
func TestI3GesamtzahlDerPfadformen(t *testing.T) {
	k, u := konfigMit()
	sessionen, err := admin.SessionRouten(&sessionsFake{}, u.audit)
	if err != nil {
		t.Fatal(err)
	}
	inspektion, err := admin.InspektionRouten(&inspektionFake{})
	if err != nil {
		t.Fatal(err)
	}
	alle := slices.Concat(sessionen, inspektion)
	k.Routen = alle
	neuMitKonfig(t, k, u)
	muster, paare := map[string]bool{}, map[[2]string]bool{}
	for _, r := range alle {
		muster[r.Muster] = true
		paare[[2]string{r.Methode, r.Muster}] = true
	}
	gleichAuf(t, "Fälle, Treffer, Pfadformen, Routen", []int{len(muster) * 7, len(paare), len(muster), len(alle)}, []int{42, 7, 6, 7})
}

// inspPfadwerte ist M-Pfadwert (SPEC-293 Nr. 5, SPEC-292 Nr. 4): 13 kodierte Werte mit den dekodierten Bytes als Literale aus
// Bytewerten, dazu drei mit einem `%` im dekodierten Wert (die Route dekodiert nichts nach).
func inspPfadwerte() []struct{ kodiert, soll string } {
	b := func(x ...byte) string { return string(x) }
	return []struct{ kodiert, soll string }{
		{"x", "x"}, {"ses-" + strings.Repeat("0123456789", 3) + "01", "ses-" + strings.Repeat("0123456789", 3) + "01"},
		{"SES-1", "SES-1"}, {"a.b", "a.b"}, {".", "."}, {"..", ".."}, {"a%2Fb", "a/b"}, {"%C3%A4", b(0xC3, 0xA4)},
		{"%FF", b(0xFF)}, {"%00", b(0)}, {strings.Repeat("a", 65), strings.Repeat("a", 65)}, {"a+b", "a+b"}, {"%20", b(0x20)},
		{"%2541", "%41"}, {"%25zz", "%zz"}, {"a%25", "a%"},
	}
}

// TestI4Pfadwert hält I4 (SPEC-293 Nr. 5): jeder der 16 Werte (13 des Vertrags und drei mit `%` im Wert) geht byte-gleich an
// Exchanges, ohne Prüfung; der Fehler `ODR-E7003` steht ohne den Wert, und ein Canary als Wert steht in keinem Byte von
// Status, Header und Body. Dazu die Klasse der Zusage „ohne Prüfung von Form“: jedes der 256 Bytes als `%XX` kodiert an vier
// Stellen (allein, zwischen zwei Buchstaben, am Ende, am Anfang: 1024 Fälle). Zählprüfung: 13 + 3 Werte und die eine Route mit `{id}`.
func TestI4Pfadwert(t *testing.T) {
	werte := inspPfadwerte()
	for _, w := range werte {
		u, fake, _ := inspUmgebung(t)
		rec := u.rufe(inspReq(exchangesZiel(w.kodiert), ""))
		if rec.Code != 200 || rec.Body.String() != inspLeereSeite {
			t.Errorf("%s: Antwort %d %q, erwartet 200 %q", w.kodiert, rec.Code, rec.Body.String(), inspLeereSeite)
		}
		if got := fake.argumente(); len(got) != 1 || got[0].id != w.soll || got[0].seite != (inbound.Seite{}) {
			t.Errorf("%s: Argumente %+v, erwartet id %q ohne Seite", w.kodiert, got, w.soll)
		}
		fake.nurDiagonale(t, w.kodiert, inspOpExchanges)

		u, fake, _ = inspUmgebung(t)
		fake.err = inspE7003()
		pruefeFehlerAntwort(t, w.kodiert+" unbekannt", u.rufe(inspReq(exchangesZiel(w.kodiert), "")), 404, objektE7003, "")
		if got := fake.argumente(); len(got) != 1 || got[0].id != w.soll {
			t.Errorf("%s: Argumente beim Fehler %+v", w.kodiert, got)
		}
	}
	for _, fehler := range []error{nil, inspE7003()} {
		u, fake, _ := inspUmgebung(t)
		fake.err = fehler
		if rec := u.rufe(inspReq(exchangesZiel(canary()), "")); strings.Contains(antwortText(rec), canary()) {
			t.Errorf("der Canary steht in der Antwort: %s", antwortText(rec))
		}
	}
	faelle := 0
	for b := range 256 {
		for _, form := range []string{"%s", "a%sz", "a%s", "%sz"} {
			kodiert := strings.Replace(form, "%s", fmt.Sprintf("%%%02X", b), 1)
			soll := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
			u, fake, _ := inspUmgebung(t)
			faelle++
			if rec := u.rufe(inspReq(exchangesZiel(kodiert), "")); rec.Code != 200 {
				t.Errorf("Byte 0x%02X in der Form %q: Status %d", b, form, rec.Code)
			}
			if got := fake.argumente(); len(got) != 1 || got[0].id != soll {
				t.Errorf("Byte 0x%02X in der Form %q: Argumente %+v, erwartet %q", b, form, got, soll)
			}
		}
	}
	gleichAuf(t, "Werte, Fälle je Byte", []int{len(werte), faelle}, []int{16, 1024})
	_, _, routen := inspUmgebung(t)
	gleichAuf(t, "Routen mit {id}", strings.Count(routen[0].Muster, "{"), 1)
}
