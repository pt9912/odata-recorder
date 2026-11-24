package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// wsEintragArt ist die Art des Eintrags der Route i (SPEC-298 Nr. 11, Tabelle *Wer schreibt was*) als Literal; die Routen ohne
// Eintrag liefern "".
func wsEintragArt(i int) string {
	return map[int]string{
		rCreate: "workspace_angelegt", rReset: "workspace_zurueckgesetzt", rSnCreate: "snapshot_erzeugt", rRestore: "workspace_wiederhergestellt",
	}[i]
}

// wsDienst stellt die Dienste auf Ergebnisse, die weder der Pfad noch der Body nennen: jedes Feld des Workspace und des
// Snapshot trägt einen eigenen Wert (`ServiceID` und `WorkspaceID` des Snapshots weichen vom Workspace ab).
func wsDienst(e *wsEnv) {
	e.wf.ws = model.Workspace{ID: "ws-dienst", ServiceID: "svc-dienst", BaselineSnapshotID: "base-dienst", Revision: 7, CreatedAt: zeitA()}
	e.sf.snap = model.WorkspaceSnapshot{ID: "snp-dienst", ServiceID: "svc-snap", WorkspaceID: "ws-snap", Name: "nm-dienst", CreatedAt: zeitA()}
}

// wsDienstEintrag ist der Eintrag, den ein Erfolg der Route i mit den Ergebnissen von wsDienst schreibt, alle sieben Felder als
// Literale (SPEC-298 Nr. 11 b).
func wsDienstEintrag(i int, korr string) support.Eintrag {
	if i == rSnCreate {
		return support.Eintrag{
			Art: "snapshot_erzeugt", ServiceID: "svc-snap", SessionID: "", WorkspaceID: "ws-snap", ImportID: "", ErrorCode: 0, KorrelationsID: korr,
		}
	}
	return support.Eintrag{
		Art: support.AuditArt(wsEintragArt(i)), ServiceID: "svc-dienst", SessionID: "", WorkspaceID: "ws-dienst", ImportID: "", ErrorCode: 0, KorrelationsID: korr,
	}
}

// TestWS19Eintrag hält WS19 (SPEC-298 Nr. 11 a, b, d): nach einem Erfolg ruft jede der vier Routen mit Eintrag Append genau
// einmal, bei 201 und bei 200, mit der Art ihrer Route, `ServiceID` und `WorkspaceID` aus dem Ergebnis des Dienstes (der Fake
// liefert andere Werte, als Pfad und Body nennen), der Korrelations-ID der Antwort und sonst leeren Feldern (das Literal trägt alle
// sieben). Im Haltepunkt von Append ist die Operation gelaufen, und die Route hat nichts geschrieben. Ein zweiter Request
// schreibt einen zweiten Eintrag mit eigener Korrelations-ID, der Handler ohne Korrelations-ID einen mit leerer. Zählprüfung: 4
// Routen, 4 Arten gleich den Zeilen 9 bis 12 von `support.Arten()`, 7 Felder gleich `NumField` von `support.Eintrag`.
func TestWS19Eintrag(t *testing.T) {
	wsFrist(t)
	routen, arten := 0, []string{}
	for i, r := range wsRouten() {
		if wsEintragArt(i) == "" {
			continue
		}
		routen++
		arten = append(arten, wsEintragArt(i))
		e := wsNeu(t)
		wsDienst(e)
		var ev ereignisse
		rec := httptest.NewRecorder()
		var vorAufrufe, vorOp, vorBody int
		var vorEreignisse, vorKopf []string
		e.audit.Vor = func(support.Eintrag) {
			vorAufrufe++
			vorOp = e.p.zahlen()[r.op]
			vorEreignisse, vorBody = ev.lies(), rec.Body.Len()
			for k := range rec.Header() {
				vorKopf = append(vorKopf, k)
			}
		}
		e.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &ev}, r.gut())
		if vorAufrufe != 1 || vorOp != 1 || len(vorEreignisse) != 0 || vorBody != 0 || !slices.Equal(vorKopf, []string{"X-Correlation-Id"}) {
			t.Errorf("%s: Haltepunkt %d Aufrufe, Operation %d-mal gelaufen, Ereignisse %q, Body %d Bytes, Header %q", r.name, vorAufrufe, vorOp, vorEreignisse, vorBody, vorKopf)
		}
		gleichAuf(t, r.name+": Ereignisse nach Append", ev.lies(), []string{"kopf", "schreiben"})
		korr := rec.Header().Get("X-Correlation-Id")
		if n := e.audit.Aufrufe(); n != 1 || korr == "" || korr != e.ids.letzte() || rec.Code != r.status {
			t.Fatalf("%s: %d Aufrufe von Audit, Korrelations-ID %q (letzte vergebene %q), Status %d, erwartet %d", r.name, n, korr, e.ids.letzte(), rec.Code, r.status)
		}
		gleichAuf(t, r.name+": Eintrag", e.audit.Eintraege(), []support.Eintrag{wsDienstEintrag(i, korr)})

		e.audit.Vor = nil
		rec2 := e.rufe(r.gut())
		if _, err := e.routen[i].Handler(context.Background(), admin.Anfrage{
			Methode: r.methode, Body: strings.NewReader(r.body), Pfadwerte: map[string]string{"id": wsPfadID},
		}); err != nil {
			t.Fatal(err)
		}
		eintraege := e.audit.Eintraege()
		if len(eintraege) != 3 || eintraege[1] != wsDienstEintrag(i, rec2.Header().Get("X-Correlation-Id")) || eintraege[1].KorrelationsID == korr ||
			eintraege[2] != wsDienstEintrag(i, "") {
			t.Errorf("%s: zweiter Request und Handler ohne Korrelations-ID: Einträge %+v", r.name, eintraege)
		}
	}
	zeilen := support.Arten()[8:]
	gleichAuf(t, "Routen mit Eintrag, Zeilen 9 bis 12, Felder", []int{routen, len(zeilen), reflect.TypeFor[support.Eintrag]().NumField()}, []int{4, 4, 7})
	for k, a := range arten {
		if string(zeilen[k].Art) != a {
			t.Errorf("Zeile %d von Arten() ist %q, erwartet %q", 9+k, zeilen[k].Art, a)
		}
	}
}

// TestWS20KeinEintrag hält WS20 (SPEC-298 Nr. 10, 11 c) als Menge M-Audit: die Zellen Route × Ausgang (je Route 1 Erfolg + 11
// allgemeine Ausgänge + die Codes von Nr. 8; 65 bei den vier Routen mit Eintrag, 59 bei den anderen, zusammen 124 gleich 8 × 12 +
// 28) mit dem Orakel `FakeAudit.Aufrufe()` gleich 1 in genau den 4 Erfolgszellen der Routen mit Eintrag, sonst 0 (auch bei dem
// Erfolg von `DELETE` und bei einem Fehler des Dienstes, der einen Wert mitliefert). Dasselbe Orakel lesen die 145 ungültigen Bodys
// (400) und die 9 Zellen der Achse Lesen (Abbruch) von Teil 1; die 140 Zellen von WS18 prüft `TestWS18Kontext`.
func TestWS20KeinEintrag(t *testing.T) {
	wsFrist(t)
	zellen, mitEintrag, jeRoute := 0, 0, map[int]int{}
	for r, rt := range wsRouten() {
		for _, z := range slices.Concat([]wsFall{{route: r, name: "Erfolg"}}, wsAllgemein(r), wsCodes(r)) {
			zellen++
			jeRoute[r]++
			name := rt.name + " " + z.name
			e := wsNeu(t)
			e.wf.err, e.sf.err = z.err, z.err
			rec, p := wsRahmenAntwort(e, rt.gut())
			soll := 0
			if z.err == nil && wsEintragArt(r) != "" {
				soll = 1
				mitEintrag++
			}
			if n := e.audit.Aufrufe(); n != soll || (z.err == nil) != (p == nil && rec.Code == rt.status) {
				t.Errorf("%s: %d Aufrufe von Audit, Status %d, Panik %v, erwartet %d Aufrufe", name, n, rec.Code, p, soll)
			}
			if es := e.audit.Eintraege(); soll == 1 && (len(es) != 1 || string(es[0].Art) != wsEintragArt(r)) {
				t.Errorf("%s: Einträge %+v, erwartet einen der Art %q", name, es, wsEintragArt(r))
			}
		}
	}
	gleichAuf(t, "Zellen, Zellen mit Eintrag, Zellen je Route", []int{zellen, mitEintrag, 8*12 + 28}, []int{124, 4, 124})
	gleichAuf(t, "Zellen je Route", jeRoute, map[int]int{rList: 14, rCreate: 16, rGet: 15, rDelete: 15, rReset: 15, rSnList: 15, rSnCreate: 16, rRestore: 18})

	bodys := wsUngueltigeBodys(canary())
	gleichAuf(t, "Zahl der ungültigen Bodys", len(bodys), 145)
	for _, f := range bodys {
		e := wsNeu(t)
		rec := e.rufe(f.request())
		if n := e.audit.Aufrufe(); n != 0 || rec.Code != 400 {
			t.Errorf("%s %s: %d Aufrufe von Audit, Status %d, erwartet 0 und 400", f.achse, f.name, n, rec.Code)
		}
	}
	lesen := wsLesen()
	gleichAuf(t, "Zellen der Achse Lesen", len(lesen), 9)
	for _, z := range lesen {
		e := wsNeu(t)
		rec, p := wsRahmenAntwort(e, wsRouten()[z.route].request(wsPfadID, "", nil, errLeser{z.err}))
		if n := e.audit.Aufrufe(); n != 0 || p != http.ErrAbortHandler || rec.Body.Len() != 0 {
			t.Errorf("%s: %d Aufrufe von Audit, Panik %v, erwartet 0 und den Abbruch", z.name, n, p)
		}
	}
}

// TestWS21AppendFehler hält WS21 (SPEC-298 Nr. 11 e) als Schleife über appendFehlerMenge (14 Arten × 4 Verpackungen = 56
// Fehler): jeder Fehler von Append ändert Status, Body und Header der Antwort der vier Routen mit Eintrag nicht (224 Zellen, gleich
// der Antwort ohne Fehler), löst keine Panik aus, Append lief einmal, es gab keine Zählung. Prüfgrenze: Arten und Verpackungen
// außerhalb der 14 und 4.
func TestWS21AppendFehler(t *testing.T) {
	wsFrist(t)
	menge := appendFehlerMenge()
	gleichAuf(t, "Zahl der Fehler", len(menge), 14*4)
	zellen := 0
	for i, r := range wsRouten() {
		if wsEintragArt(i) == "" {
			continue
		}
		sende := func(fehler error) (*wsEnv, *httptest.ResponseRecorder, any, bool) {
			e := wsNeu(t)
			wsDienst(e)
			e.audit.Fehler = fehler
			rec, wert, panikt := e.rufeMitPanik(r.gut())
			return e, rec, wert, panikt
		}
		_, ohne, _, _ := sende(nil)
		if ohne.Code != r.status || ohne.Body.Len() == 0 {
			t.Fatalf("%s ohne Fehler: Antwort %d %q, erwartet %d mit Body", r.name, ohne.Code, ohne.Body.String(), r.status)
		}
		for _, f := range menge {
			zellen++
			e, rec, wert, panikt := sende(f.err)
			if panikt || rec.Code != ohne.Code || rec.Body.String() != ohne.Body.String() || !reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
				t.Errorf("%s, Fehler %s: Panik %v (%v), Antwort %d %q %v, erwartet %d %q %v", r.name, f.name, panikt, wert, rec.Code, rec.Body.String(),
					rec.Result().Header, ohne.Code, ohne.Body.String(), ohne.Result().Header)
			}
			if n := e.audit.Aufrufe(); n != 1 {
				t.Errorf("%s, Fehler %s: %d Aufrufe von Audit, erwartet 1", r.name, f.name, n)
			}
			if codes, andere := zaehlungenVon(e.tele); len(codes) != 0 || andere != 0 {
				t.Errorf("%s, Fehler %s: Zählungen %q, andere Aufrufe %d", r.name, f.name, codes, andere)
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 4*14*4)
}

// wsKennungPaare sind die Paare (`WorkspaceID`, `ServiceID`) der Menge M-Kennung: 9 aus leer, Leerzeichen mit LF und 65 Bytes,
// jedes Byte 0 bis 255 allein und zwischen zwei Buchstaben in beiden Feldern (512), jede Länge von kennungLaengen und jede Folge
// von kennungFolgen in je einem Feld (2 × 350 und 2 × 324); das andere Feld ist gültig. Das Orakel ist die Eingabe selbst.
func wsKennungPaare() [][2]string {
	werte := []string{"", "a b\n", strings.Repeat("a", 65)}
	var out [][2]string
	for _, id := range werte {
		for _, svc := range werte {
			out = append(out, [2]string{id, svc})
		}
	}
	for b := range 256 {
		for _, form := range []string{"%s", "a%sz"} {
			wert := strings.Replace(form, "%s", string([]byte{byte(b)}), 1)
			out = append(out, [2]string{wert, wert})
		}
	}
	for _, n := range kennungLaengen() {
		out = append(out, [2]string{strings.Repeat("a", n), "svc-dienst"}, [2]string{"ws-dienst", strings.Repeat("a", n)})
	}
	for _, f := range kennungFolgen() {
		out = append(out, [2]string{f, "svc-dienst"}, [2]string{"ws-dienst", f})
	}
	return out
}

// TestWS19KennungenUnveraendert hält SPEC-298 Nr. 11 b (*Ungeprüft*) für die Zellen *leerer* und *fremder Wert*: liefert der
// Dienst Kennungen leer, außerhalb der Form von `SPEC-234` Nr. 3 oder in jeder Länge der Menge, trägt der Eintrag der vier Routen
// beide unverändert, und die Antwort ist 201 bzw. 200 (wsKennungPaare, 1869 Paare × 4 Routen = 7476 Zellen). Prüfgrenze
// (akzeptiertes Negativ): Längen über 1 MiB + 1, Folgen aus drei und mehr Zeichen.
func TestWS19KennungenUnveraendert(t *testing.T) {
	wsFrist(t)
	paare := wsKennungPaare()
	gleichAuf(t, "Paare, Längen, Folgen", []int{len(paare), len(kennungLaengen()), len(kennungFolgen())}, []int{9 + 2*256 + 2*350 + 2*324, 350, 324})
	zellen := 0
	for i, r := range wsRouten() {
		if wsEintragArt(i) == "" {
			continue
		}
		for _, p := range paare {
			zellen++
			e := wsNeu(t)
			e.wf.ws = model.Workspace{ID: p[0], ServiceID: p[1], BaselineSnapshotID: "base-dienst", Revision: 7, CreatedAt: zeitA()}
			e.sf.snap = model.WorkspaceSnapshot{ID: "snp-dienst", ServiceID: p[1], WorkspaceID: p[0], Name: "nm-dienst", CreatedAt: zeitA()}
			rec := e.rufe(r.gut())
			soll := support.Eintrag{Art: support.AuditArt(wsEintragArt(i)), ServiceID: p[1], WorkspaceID: p[0], KorrelationsID: rec.Header().Get("X-Correlation-Id")}
			if es := e.audit.Eintraege(); rec.Code != r.status || len(es) != 1 || es[0] != soll {
				t.Errorf("%s, WorkspaceID %.40q (%d Bytes), ServiceID %.40q (%d Bytes): Status %d, %d Einträge, erwartet %d und genau einen gleichen",
					r.name, p[0], len(p[0]), p[1], len(p[1]), rec.Code, len(es), r.status)
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 4*1869)
}
