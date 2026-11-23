package admin_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// eintragImport ist der Eintrag, den ein Erfolg mit importA schreibt, alle sieben Felder als Literale (SPEC-297 Nr. 10 b).
func eintragImport(korr string) support.Eintrag {
	return support.Eintrag{
		Art: "import_veroeffentlicht", ServiceID: "svc-dienst", SessionID: "", WorkspaceID: "", ImportID: "imp-dienst", ErrorCode: 0, KorrelationsID: korr,
	}
}

// TestIM20Eintrag hält IM20 (SPEC-297 Nr. 10 a, b, d): nach einem Erfolg von Import ruft `POST` Append genau einmal, bei 201 und
// bei 200, mit der Art `import_veroeffentlicht`, `ServiceID` und `ImportID` aus dem Ergebnis des Dienstes (der Fake liefert
// andere Werte, als der Body nennt, und der Body nennt eine Session), der Korrelations-ID der Antwort und sonst leeren Feldern
// (das Literal trägt alle sieben). Im Haltepunkt von Append ist Import gelaufen, und die Route hat nichts geschrieben (kein
// WriteHeader, kein Write, kein Header außer der Korrelations-ID); der Haltepunkt lief genau einmal. Ein zweites `POST` schreibt
// einen zweiten Eintrag mit eigener Korrelations-ID; der Handler ohne Korrelations-ID schreibt einen mit leerer.
func TestIM20Eintrag(t *testing.T) {
	for _, neu := range []bool{true, false} {
		u, fake, routen := importsUmgebung(t)
		fake.ergebnis.Neu = neu
		status := map[bool]int{true: 201, false: 200}[neu]
		var e ereignisse
		rec := httptest.NewRecorder()
		var vorAufrufe, vorImport, vorBody int
		var vorEreignisse, vorKopf []string
		u.audit.Vor = func(support.Eintrag) {
			vorAufrufe++
			vorImport = fake.zahlen()[impImport]
			vorEreignisse, vorBody = e.lies(), rec.Body.Len()
			for k := range rec.Header() {
				vorKopf = append(vorKopf, k)
			}
		}
		u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, importPost("", nil, strings.NewReader(gutBody)))
		if vorAufrufe != 1 || vorImport != 1 || len(vorEreignisse) != 0 || vorBody != 0 || !slices.Equal(vorKopf, []string{"X-Correlation-Id"}) {
			t.Errorf("Neu %v: Haltepunkt %d Aufrufe, Import %d-mal gelaufen, Ereignisse %q, Body %d Bytes, Header %q", neu, vorAufrufe, vorImport, vorEreignisse, vorBody, vorKopf)
		}
		gleichAuf(t, "Ereignisse nach Append", e.lies(), []string{"kopf", "schreiben"})
		korr := rec.Header().Get("X-Correlation-Id")
		if n := u.audit.Aufrufe(); n != 1 || korr == "" || korr != u.ids.letzte() || rec.Code != status || rec.Body.String() != reportA {
			t.Fatalf("Neu %v: %d Aufrufe von Audit, Korrelations-ID %q (letzte vergebene %q), Antwort %d %q", neu, n, korr, u.ids.letzte(), rec.Code, rec.Body.String())
		}
		gleichAuf(t, "Eintrag", u.audit.Eintraege(), []support.Eintrag{eintragImport(korr)})

		u.audit.Vor = nil
		rec2 := u.rufe(importPost("", nil, strings.NewReader(gutBody)))
		if _, err := routen[0].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: strings.NewReader(gutBody)}); err != nil {
			t.Fatal(err)
		}
		eintraege := u.audit.Eintraege()
		if len(eintraege) != 3 || eintraege[1] != eintragImport(rec2.Header().Get("X-Correlation-Id")) || eintraege[1].KorrelationsID == korr || eintraege[2] != eintragImport("") {
			t.Errorf("Neu %v: zweites POST und Handler ohne Korrelations-ID: Einträge %+v", neu, eintraege)
		}
	}
}

// TestIM22AppendFehler hält IM22 (SPEC-297 Nr. 10 e): jeder der 7 Fehler von Append ändert Status, Body und Header der Antwort
// von `POST` nicht (Byte für Byte gleich der Antwort ohne Fehler), bei 201 und bei 200 (14 Zellen); Append lief je Zelle
// einmal, und es gab keine Zählung.
func TestIM22AppendFehler(t *testing.T) {
	gleichAuf(t, "Zahl der Fehler", len(auditFehler()), 7)
	for _, status := range []int{201, 200} {
		sende := func(fehler error) (*umgebung, *httptest.ResponseRecorder) {
			u, fake, _ := importsUmgebung(t)
			fake.ergebnis.Neu = status == 201
			u.audit.Fehler = fehler
			return u, u.rufe(importPost("", nil, strings.NewReader(gutBody)))
		}
		_, ohne := sende(nil)
		for _, fehler := range auditFehler() {
			u, rec := sende(fehler)
			if rec.Code != status || rec.Body.String() != reportA || !reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
				t.Errorf("Status %d, Fehler %v: Antwort %d %q %v, erwartet %d %q %v", status, fehler, rec.Code, rec.Body.String(), rec.Result().Header,
					status, reportA, ohne.Result().Header)
			}
			if n := u.audit.Aufrufe(); n != 1 {
				t.Errorf("Status %d, Fehler %v: %d Aufrufe von Audit, erwartet 1", status, fehler, n)
			}
			if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 {
				t.Errorf("Status %d, Fehler %v: Zählungen %q, andere Aufrufe %d", status, fehler, codes, andere)
			}
		}
	}
}

// TestIM22KennungenUnveraendert hält die Zellen *leerer* und *fremder Wert* des Rasters (SPEC-297 Nr. 10 e, *Ungeprüft*): liefert
// der Dienst eine `ID` und einen `ServiceID` leer oder außerhalb der Form von `SPEC-234` Nr. 3 (ein Leerzeichen und LF, 65
// Bytes), 3 × 3 Paare, und jedes Byte 0 bis 255 allein und zwischen zwei Buchstaben in beiden Feldern zugleich (512 Paare),
// trägt der Eintrag beide unverändert, und die Antwort ist 201 mit dem Report; die Route prüft, ersetzt und kürzt nichts (der
// Adapter des Ports weist den Eintrag ab, IM22). Prüfgrenze (akzeptiertes Negativ): Längen über 65 Bytes und die Mischung von
// Bytes in einem Wert.
func TestIM22KennungenUnveraendert(t *testing.T) {
	werte := []string{"", "a b\n", strings.Repeat("a", 65)}
	var paare [][2]string
	for _, id := range werte {
		for _, svc := range werte {
			paare = append(paare, [2]string{id, svc})
		}
	}
	for b := range 256 {
		for _, form := range []string{"%s", "a%sz"} {
			wert := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
			paare = append(paare, [2]string{wert, wert})
		}
	}
	for _, p := range paare {
		u, fake, _ := importsUmgebung(t)
		fake.ergebnis.Import.ID, fake.ergebnis.Import.ServiceID = p[0], p[1]
		rec := u.rufe(importPost("", nil, strings.NewReader(gutBody)))
		soll := []support.Eintrag{{Art: "import_veroeffentlicht", ServiceID: p[1], ImportID: p[0], KorrelationsID: rec.Header().Get("X-Correlation-Id")}}
		if rec.Code != 201 || rec.Body.String() != reportA {
			t.Errorf("ID %q, ServiceID %q: Antwort %d %q", p[0], p[1], rec.Code, rec.Body.String())
		}
		gleichAuf(t, "Einträge", u.audit.Eintraege(), soll)
	}
	gleichAuf(t, "Paare", len(paare), 9+2*256)
}

// TestIM21KeinEintrag hält IM21 (SPEC-297 Nr. 10 c) als Menge M-Audit: die Zellen Route × Ausgang, `POST` 2 Erfolge (201, 200)
// + 8 allgemeine Ausgänge + 9 Codes und `GET` 1 + 8 + 3 (31 Zellen, importFehlerZellen), und Orakel `FakeAudit.Aufrufe()` gleich 1
// in genau den 2 Erfolgszellen von `POST`, sonst 0 (der Dienst liefert zum Fehler einen Wert, auch `import_fehlgeschlagen`
// schreibt die Route nie). Mit demselben Orakel lesen die Mengen von Teil 1: die 86 ungültigen Bodys (importBodyUngueltig), der
// Rand (512, gültig sind die Bodys, die mit Leerraum an den Rändern gleich gutBody sind: 8), importRandUnicode (24),
// importElementeUngueltig (305), die 44 ungültigen Elemente von importZeichenMenge × 3 Stellen × 6 Felder und die ungültigen
// Reports (reportsUngueltig 7 und die 44 ungültigen Elemente × 3 Stellen × 4 Formen von reportMitNutzlast) in den 3 Zuständen.
func TestIM21KeinEintrag(t *testing.T) {
	zellen, mitEintrag := 0, 0
	for _, route := range []string{"POST", "GET"} {
		faelle := importFehlerZellen()[route]
		for _, f := range faelle {
			zellen++
			u, fake, _ := importsUmgebung(t)
			fake.err = f.err
			rec := httptest.NewRecorder()
			abbruchVon(func() { u.srv.ServeHTTP(rec, importRequestFuer(route)) })
			if n := u.audit.Aufrufe(); n != 0 {
				t.Errorf("%s %s: %d Aufrufe von Audit, erwartet 0", route, f.name, n)
			}
		}
		for _, neu := range []bool{true, false} {
			if route == "GET" && !neu {
				continue
			}
			zellen++
			u, fake, _ := importsUmgebung(t)
			fake.ergebnis.Neu = neu
			rec := u.rufe(importRequestFuer(route))
			soll := 0
			if route == "POST" {
				soll = 1
				mitEintrag++
			}
			if n := u.audit.Aufrufe(); n != soll || rec.Code/100 != 2 {
				t.Errorf("%s Erfolg mit Neu %v: %d Aufrufe von Audit, Status %d, erwartet %d und 2xx", route, neu, n, rec.Code, soll)
			}
		}
	}
	gleichAuf(t, "Zellen, Zellen mit Eintrag, Zellen laut Plan", []int{zellen, mitEintrag, (2 + 8 + 9) + (1 + 8 + 3)}, []int{31, 2, 31})

	bodies := slices.Concat(importBodyUngueltig(canary()), importRand(), importRandUnicode(), importElementeUngueltig(canary()))
	gleichAuf(t, "Zahl der Bodys", len(bodies), 86+512+24+305)
	gueltig := 0
	for _, f := range bodies {
		u, _, _ := importsUmgebung(t)
		rec := u.rufe(importForm(f))
		soll := 0
		if bytes.Equal(bytes.Trim(f.body, " \t\r\n"), []byte(gutBody)) {
			soll, gueltig = 1, gueltig+1
		}
		if n := u.audit.Aufrufe(); n != soll || (rec.Code == 201) != (soll == 1) {
			t.Errorf("%s %s: %d Aufrufe von Audit und Status %d, erwartet %d", f.achse, f.name, n, rec.Code, soll)
		}
	}
	gleichAuf(t, "gültige Bodys", gueltig, 8)

	ungueltig := 0
	for _, z := range importZeichenMenge() {
		if z.gueltig {
			continue
		}
		ungueltig++
		for _, st := range importZeichenStellen() {
			for _, feld := range importFelder() {
				body, _ := importFeldBody(feld, []byte(st[0]+z.text+st[1]), []byte(st[0]+z.wert+st[1]))
				u, _, _ := importsUmgebung(t)
				if rec := u.rufe(importPost("", nil, bytesLeser(body))); rec.Code != 400 || u.audit.Aufrufe() != 0 {
					t.Errorf("%s in %s: Status %d, %d Aufrufe von Audit, erwartet 400 und 0", z.name, feld, rec.Code, u.audit.Aufrufe())
				}
			}
		}
	}
	gleichAuf(t, "ungültige Elemente von importZeichenMenge", ungueltig, 44)

	reports := reportsUngueltig(canary())
	for _, z := range importZeichenMenge() {
		for _, st := range importZeichenStellen() {
			for form := range reportFormenZahl {
				if !z.gueltig {
					reports = append(reports, reportMitNutzlast(form, st[0]+z.text+st[1]))
				}
			}
		}
	}
	gleichAuf(t, "ungültige Reports", len(reports), 7+44*3*reportFormenZahl)
	for _, f := range reports {
		for _, zu := range impZustaende() {
			u, fake, _ := importsUmgebung(t)
			imp := importA()
			imp.ReportJSON = string(f.roh)
			if rec := zu.sende(t, u, fake, imp); rec.Code != 500 || u.audit.Aufrufe() != 0 {
				t.Errorf("%s, %s: Status %d, %d Aufrufe von Audit, erwartet 500 und 0", f.name, zu.name, rec.Code, u.audit.Aufrufe())
			}
		}
	}
}
