package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
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

// jederFehler ist ein Fehler, der jedem Ziel von errors.Is und errors.As gleicht: eine Prüfung der Route auf einen Fehler,
// den sie nennt (ein Wert von errors.Is, ein Zeiger von errors.As), trifft ihn.
type jederFehler struct{}

func (jederFehler) Error() string { return "jeder Fehler" }
func (jederFehler) Is(error) bool { return true }
func (jederFehler) As(any) bool   { return true }

// appendFehlerMenge sind die Fehler, die Append liefern kann und die nichts ändern dürfen (SPEC-297 Nr. 10 e): 14 Arten
// (ErrUngueltig, ein Fehler ohne Code, io.ErrShortWrite, io.EOF, Canceled, DeadlineExceeded, os.ErrClosed, ein *fs.PathError, ein
// *inbound.Fehler mit Code 404 und mit Code 500, ein *inbound.Fehler ohne Code und Status, ein nil-Zeiger *inbound.Fehler,
// model.ErrKonflikt und ein Fehler, der jedem Ziel von errors.Is und errors.As gleicht) × 4 Verpackungen (bloß, mit `%w`, Join mit
// ihm zuerst, Join mit ihm zuletzt) = 56 Formen; die Arten der 7 Fehler von auditFehler sind darin enthalten.
func appendFehlerMenge() []fehlerFall {
	var out []fehlerFall
	for _, art := range []struct {
		name string
		err  error
	}{
		{"ErrUngueltig", model.ErrUngueltig}, {"Fehler ohne Code", errors.New("Ablage")}, {"io.ErrShortWrite", io.ErrShortWrite}, {"io.EOF", io.EOF},
		{"Canceled", context.Canceled}, {"DeadlineExceeded", context.DeadlineExceeded}, {"os.ErrClosed", os.ErrClosed},
		{"*fs.PathError", &fs.PathError{Op: "write", Path: "audit.jsonl", Err: syscall.ENOSPC}},
		{"*inbound.Fehler 404", fehlerE7003()}, {"*inbound.Fehler 500", &inbound.Fehler{Code: model.CodeE9001, Status: 500}},
		{"*inbound.Fehler ohne Code", &inbound.Fehler{}}, {"nil-Zeiger *inbound.Fehler", (*inbound.Fehler)(nil)},
		{"model.ErrKonflikt", model.ErrKonflikt}, {"Fehler für jedes Ziel", jederFehler{}},
	} {
		andere := errors.New("andere")
		for _, v := range []struct {
			name string
			err  error
		}{
			{"bloß", art.err}, {"mit %w", fmt.Errorf("x: %w", art.err)},
			{"Join, zuerst", errors.Join(art.err, andere)}, {"Join, zuletzt", errors.Join(andere, art.err)},
		} {
			out = append(out, fehlerFall{name: art.name + " " + v.name, err: v.err})
		}
	}
	return out
}

// TestIM22AppendFehler hält IM22 (SPEC-297 Nr. 10 e) als Schleife über appendFehlerMenge: jeder der 56 Fehler von Append ändert
// Status, Body und Header der Antwort von `POST` nicht (Byte für Byte gleich der Antwort ohne Fehler) und löst weder eine Panik
// noch den Abbruch der Verbindung aus, bei 201 und bei 200 (112 Zellen); Append lief je Zelle einmal, und es gab keine Zählung.
// Prüfgrenze (akzeptiertes Negativ): Arten von Fehlern außerhalb der 14 und Verpackungen außerhalb der 4.
func TestIM22AppendFehler(t *testing.T) {
	menge := appendFehlerMenge()
	gleichAuf(t, "Zahl der Fehler", len(menge), 14*4)
	zellen := 0
	for _, status := range []int{201, 200} {
		sende := func(fehler error) (*umgebung, *httptest.ResponseRecorder, any, bool) {
			u, fake, _ := importsUmgebung(t)
			fake.ergebnis.Neu = status == 201
			u.audit.Fehler = fehler
			rec, wert, panikt := u.rufeMitPanik(importPost("", nil, strings.NewReader(gutBody)))
			return u, rec, wert, panikt
		}
		_, ohne, _, _ := sende(nil)
		for _, f := range menge {
			zellen++
			u, rec, wert, panikt := sende(f.err)
			if panikt || rec.Code != status || rec.Body.String() != reportA || !reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
				t.Errorf("Status %d, Fehler %s: Panik %v (%v), Antwort %d %q %v, erwartet %d %q %v", status, f.name, panikt, wert, rec.Code, rec.Body.String(),
					rec.Result().Header, status, reportA, ohne.Result().Header)
			}
			if n := u.audit.Aufrufe(); n != 1 {
				t.Errorf("Status %d, Fehler %s: %d Aufrufe von Audit, erwartet 1", status, f.name, n)
			}
			if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 {
				t.Errorf("Status %d, Fehler %s: Zählungen %q, andere Aufrufe %d", status, f.name, codes, andere)
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 2*14*4)
}

// kennungLaengen sind die Längen der Menge M-Kennung (IM22): jede Länge 0 bis 300 und die Längen von importLaengen(0, 1 MiB + 1)
// (um Zweier- und Zehnerpotenzen bis 1 MiB + 1), aufsteigend und ohne Doppelte.
func kennungLaengen() []int {
	menge := map[int]bool{}
	for n := range 301 {
		menge[n] = true
	}
	for _, n := range importLaengen(0, bodyGrenze+1) {
		menge[n] = true
	}
	return slices.Sorted(maps.Keys(menge))
}

// kennungFolgen sind die Teilfolgen der Menge M-Kennung (IM22): jede Folge xy aus 9 Zeichen (Kleinbuchstabe, Großbuchstabe, Ziffer,
// `.`, `_`, `-`, Leerzeichen, NUL, 0xFF; 81 Folgen), allein, vor `z`, hinter `a` und zwischen `a` und `z` (4 Stellen) = 324
// Werte; die Folgen `--` und `..` und jede andere Zweierfolge der Form sind darin.
func kennungFolgen() []string {
	zeichen := []string{"a", "Z", "0", ".", "_", "-", " ", "\x00", "\xff"}
	var out []string
	for _, x := range zeichen {
		for _, y := range zeichen {
			for _, st := range [][2]string{{"", ""}, {"", "z"}, {"a", ""}, {"a", "z"}} {
				out = append(out, st[0]+x+y+st[1])
			}
		}
	}
	return out
}

// kennungZelle lässt den Dienst die Kennungen id und svc liefern und prüft: die Antwort ist 201 mit dem Report, und genau ein
// Eintrag trägt beide unverändert (Art, Korrelations-ID der Antwort, übrige Felder leer).
func kennungZelle(t *testing.T, name, id, svc string) {
	t.Helper()
	u, fake, _ := importsUmgebung(t)
	fake.ergebnis.Import.ID, fake.ergebnis.Import.ServiceID = id, svc
	rec := u.rufe(importPost("", nil, strings.NewReader(gutBody)))
	soll := support.Eintrag{Art: "import_veroeffentlicht", ServiceID: svc, ImportID: id, KorrelationsID: rec.Header().Get("X-Correlation-Id")}
	if rec.Code != 201 || rec.Body.String() != reportA {
		t.Errorf("%s: Antwort %d %q", name, rec.Code, rec.Body.String())
	}
	e := u.audit.Eintraege()
	if len(e) != 1 {
		t.Errorf("%s: %d Einträge, erwartet 1", name, len(e))
	} else if e[0] != soll {
		t.Errorf("%s: Eintrag mit ID der Länge %d (erwartet %d), ServiceID der Länge %d (erwartet %d), Art %q, Korrelations-ID %q", name,
			len(e[0].ImportID), len(id), len(e[0].ServiceID), len(svc), e[0].Art, e[0].KorrelationsID)
	}
}

// TestIM22KennungenUnveraendert hält die Zellen *leerer* und *fremder Wert* des Rasters (SPEC-297 Nr. 10 e, *Ungeprüft*): liefert
// der Dienst eine `ID` und einen `ServiceID` leer, außerhalb der Form von `SPEC-234` Nr. 3 oder in jeder Länge der Menge, trägt
// der Eintrag beide unverändert, und die Antwort ist 201 mit dem Report; die Route prüft, ersetzt und kürzt nichts (der Adapter des
// Ports weist den Eintrag ab, IM22). Menge M-Kennung, Orakel die Eingabe selbst: 9 Paare aus leer, Leerzeichen mit LF und 65 Bytes;
// jedes Byte 0 bis 255 allein und zwischen zwei Buchstaben in beiden Feldern zugleich (512 Paare); jede Länge von kennungLaengen
// (350: 0 bis 300 und die Längen um Zweier- und Zehnerpotenzen bis 1 MiB + 1) und jede Teilfolge von kennungFolgen (324) in je
// einem Feld, das andere ist gültig (2 × 350 + 2 × 324 Zellen). Prüfgrenze (akzeptiertes Negativ, Plan §3 (5)): Längen über
// 1 MiB + 1, Folgen aus drei und mehr Zeichen, die Mischung von Bytes in einem Wert über die Zweierfolgen hinaus.
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
		kennungZelle(t, fmt.Sprintf("ID %q, ServiceID %q", p[0], p[1]), p[0], p[1])
	}
	laengen, folgen, zellen := kennungLaengen(), kennungFolgen(), 0
	for _, n := range laengen {
		kennungZelle(t, fmt.Sprintf("ID der Länge %d", n), strings.Repeat("a", n), "svc-dienst")
		kennungZelle(t, fmt.Sprintf("ServiceID der Länge %d", n), "imp-dienst", strings.Repeat("a", n))
		zellen += 2
	}
	for _, f := range folgen {
		kennungZelle(t, fmt.Sprintf("ID %q", f), f, "svc-dienst")
		kennungZelle(t, fmt.Sprintf("ServiceID %q", f), "imp-dienst", f)
		zellen += 2
	}
	gleichAuf(t, "Paare, Längen, Folgen, Zellen von Längen und Folgen", []int{len(paare), len(laengen), len(folgen), zellen},
		[]int{9 + 2*256, 350, 4 * 81, 2*350 + 2*324})
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
