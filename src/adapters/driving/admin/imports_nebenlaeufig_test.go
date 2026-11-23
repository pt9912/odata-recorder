package admin_test

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// importBodyMit ist ein gültiger Body von `POST` mit der Kennung id, importReportMit der Report, den der Fake zu id liefert.
func importBodyMit(id string) string {
	return `{"import_id":"` + id + `","service_id":"svc-1","source_session_ids":["ses-1"]}`
}

func importReportMit(id string) string { return `{"import_id":"` + id + `"}` }

// TestIN1GleichzeitigeRequests hält IN1 (SPEC-297 Nr. 11): in 150 Runden je 32 `POST` und 32 `GET` zugleich, jeder Request mit
// eigener Kennung, die der Fake in `ServiceID`, `ID` und im Report zurückgibt; jede Antwort und jeder Eintrag gehört zum
// eigenen Request (der Eintrag trägt die Korrelations-ID der Antwort derselben Import-ID), kein `GET` schreibt einen. Alle 32
// Requests einer Route sind zugleich in jedem Fenster zwischen einem Aufruf eines Ports und dem nächsten Schritt der Route
// (Treffpunkt je Fenster, 6): bei `POST` am Ende des Lesens des Bodys, im Aufruf von Import, nach seiner Rückkehr und nach der
// Rückkehr von Append, bei `GET` im Aufruf von Get und nach seiner Rückkehr. Der Detektor von `make test-race` sieht die
// Zugriffe nach dem Treffpunkt ungeordnet; ein Zustand unter einer Sperre ohne Datenrennen, der in einem dieser Fenster
// schreibt und zurückliest, trifft die Antwort eines anderen Requests. Ein Fenster innerhalb der Route ohne Fuge (Schreiben und
// Zurücklesen hintereinander) hält kein Treffpunkt: Plan §3, akzeptiertes Negativ (4). Hält eine Sperre die Requests auseinander,
// endet der Test nach der ersten Frist mit Fatalf.
func TestIN1GleichzeitigeRequests(t *testing.T) {
	const n, runden = 32, 150
	u, fake, _ := importsUmgebung(t)
	fake.importiere = func(_ context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
		return inbound.ImportErgebnis{Import: model.Import{ID: a.ImportID, ServiceID: "s-" + a.ImportID, ReportJSON: importReportMit(a.ImportID)}, Neu: true}, nil
	}
	fake.liest = func(_ context.Context, id string) (model.Import, error) {
		return model.Import{ID: id, ServiceID: "s-" + id, ReportJSON: importReportMit(id)}, nil
	}
	korrZuID := map[string]string{}
	for r := range runden {
		fenster := map[string]*treffpunkt{
			"POST nach dem Lesen des Bodys": neuerTreffpunkt(n), "POST im Aufruf von Import": neuerTreffpunkt(n),
			"POST nach der Rückkehr von Import": neuerTreffpunkt(n), "POST nach der Rückkehr von Append": neuerTreffpunkt(n),
			"GET im Aufruf von Get": neuerTreffpunkt(n), "GET nach der Rückkehr von Get": neuerTreffpunkt(n),
		}
		fake.vor[impImport] = func(context.Context) { fenster["POST im Aufruf von Import"].halt() }
		fake.nach[impImport] = func(context.Context) { fenster["POST nach der Rückkehr von Import"].halt() }
		fake.vor[impGet] = func(context.Context) { fenster["GET im Aufruf von Get"].halt() }
		fake.nach[impGet] = func(context.Context) { fenster["GET nach der Rückkehr von Get"].halt() }
		u.nachAudit.nach = fenster["POST nach der Rückkehr von Append"].halt
		antworten := gleichzeitig(u, 2*n, func(i int) *http.Request {
			id := "g" + strconv.Itoa(r*n+i%n)
			if i < n {
				return importPost("", nil, &bodyLeser{daten: []byte(importBodyMit(id)), ende: fenster["POST nach dem Lesen des Bodys"].halt})
			}
			return importGet(id)
		})
		gleichAuf(t, "Zahl der Fenster", len(fenster), 6)
		for name, p := range fenster {
			if p.verpasst.Load() || p.n.Load() != n {
				t.Fatalf("Runde %d: nicht alle %d Requests waren zugleich im Fenster %s (%d eingetreten): eine Sperre hält sie auseinander", r, n, name, p.n.Load())
			}
		}
		for i := range n {
			id := "g" + strconv.Itoa(r*n+i)
			if rec := antworten[i]; rec.Code != 201 || rec.Body.String() != importReportMit(id) {
				t.Errorf("Runde %d POST %d: %d %q", r, i, rec.Code, rec.Body.String())
			}
			if rec := antworten[n+i]; rec.Code != 200 || rec.Body.String() != importReportMit(id) {
				t.Errorf("Runde %d GET %d: %d %q", r, i, rec.Code, rec.Body.String())
			}
			korrZuID[antworten[i].Header().Get("X-Correlation-Id")] = id
		}
	}
	gesehen := map[string]bool{}
	eintraege := u.audit.Eintraege()
	for _, e := range eintraege {
		if korrZuID[e.KorrelationsID] != e.ImportID || e.ServiceID != "s-"+e.ImportID || e.Art != "import_veroeffentlicht" {
			t.Errorf("Eintrag %+v gehört nicht zur Antwort mit der Korrelations-ID", e)
		}
		gesehen[e.ImportID] = true
	}
	gleichAuf(t, "Einträge, verschiedene Kennungen, Korrelations-IDs", []int{len(eintraege), len(gesehen), len(korrZuID)},
		[]int{runden * n, runden * n, runden * n})
}

// TestIN1ZweiGleichzeitigeGleicheImportID hält IN1 (SPEC-297 Nr. 11) für zwei `POST` zugleich mit derselben `import_id` und
// gleicher Anfrage, in 40 Runden mit je eigener Kennung: ein Fake mit Zustand (eigene Sperre) liefert dem zweiten Import
// `Neu` falsch; beide sind zugleich im Dienst (Treffpunkt). Orakel je Runde: einmal 201 und einmal 200, die Bodys sind
// byte-gleich dem Report, und es entstehen zwei Einträge mit der Import-ID und den Korrelations-IDs der beiden Antworten.
func TestIN1ZweiGleichzeitigeGleicheImportID(t *testing.T) {
	const runden = 40
	var mu sync.Mutex
	bekannt := map[string]bool{}
	u, fake, _ := importsUmgebung(t)
	fake.importiere = func(_ context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
		mu.Lock()
		defer mu.Unlock()
		neu := !bekannt[a.ImportID]
		bekannt[a.ImportID] = true
		return inbound.ImportErgebnis{Import: model.Import{ID: a.ImportID, ServiceID: "svc-1", ReportJSON: importReportMit(a.ImportID)}, Neu: neu}, nil
	}
	for r := range runden {
		id := "z" + strconv.Itoa(r)
		p := neuerTreffpunkt(2)
		fake.vor[impImport] = func(context.Context) { p.halt() }
		antworten := gleichzeitig(u, 2, func(int) *http.Request { return importPost("", nil, strings.NewReader(importBodyMit(id))) })
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Fatalf("Runde %d: beide Requests waren nicht zugleich im Dienst (%d Aufrufe): eine Sperre hält sie auseinander", r, p.n.Load())
		}
		codes := []int{antworten[0].Code, antworten[1].Code}
		slices.Sort(codes)
		if !slices.Equal(codes, []int{200, 201}) || antworten[0].Body.String() != importReportMit(id) || antworten[1].Body.String() != importReportMit(id) {
			t.Errorf("Runde %d: Status %v, Bodys %q und %q, erwartet 200 und 201 und den Report", r, codes, antworten[0].Body.String(), antworten[1].Body.String())
		}
		korr := []string{antworten[0].Header().Get("X-Correlation-Id"), antworten[1].Header().Get("X-Correlation-Id")}
		slices.Sort(korr)
		eintraege := u.audit.Eintraege()
		if len(eintraege) != 2*(r+1) {
			t.Fatalf("Runde %d: %d Einträge, erwartet %d", r, len(eintraege), 2*(r+1))
		}
		jetzt := eintraege[2*r:]
		if jetzt[0].ImportID != id || jetzt[1].ImportID != id || korr[0] == korr[1] ||
			!slices.Equal(korr, []string{min(jetzt[0].KorrelationsID, jetzt[1].KorrelationsID), max(jetzt[0].KorrelationsID, jetzt[1].KorrelationsID)}) {
			t.Errorf("Runde %d: Einträge %+v, Korrelations-IDs der Antworten %q", r, jetzt, korr)
		}
	}
}

// TestIN2KeineSperreUeberDemAufruf hält IN2 (SPEC-297 Nr. 11): zwei Requests derselben Route mit verschiedenen Kennungen sind
// zugleich im Aufruf von Import (`POST`), von Get (`GET`) und in Append (`POST`); der Aufruf kehrt erst zurück, wenn der zweite
// eingetreten ist (Wirkungs-Fänger mit Kanal; die Frist von 10 s ist nur der Abbruch im Fehlschlag).
func TestIN2KeineSperreUeberDemAufruf(t *testing.T) {
	stellen := []struct{ name, route string }{{"Import", "POST"}, {"Get", "GET"}, {"Append", "POST"}}
	for _, s := range stellen {
		u, fake, _ := importsUmgebung(t)
		p := neuerTreffpunkt(2)
		switch s.name {
		case "Import":
			fake.vor[impImport] = func(context.Context) { p.halt() }
		case "Get":
			fake.vor[impGet] = func(context.Context) { p.halt() }
		case "Append":
			u.audit.Vor = func(support.Eintrag) { p.halt() }
		}
		for i, rec := range gleichzeitig(u, 2, func(i int) *http.Request {
			if s.route == "POST" {
				return importPost("", nil, strings.NewReader(importBodyMit("imp-"+strconv.Itoa(i))))
			}
			return importGet("imp-" + strconv.Itoa(i))
		}) {
			if want := map[string]int{"POST": 201, "GET": 200}[s.route]; rec.Code != want {
				t.Errorf("%s: Request %d endete mit %d, erwartet %d", s.name, i, rec.Code, want)
			}
		}
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Errorf("%s: der zweite Aufruf trat nicht ein, solange der erste im Aufruf war (%d Aufrufe)", s.name, p.n.Load())
		}
	}
	gleichAuf(t, "Zahl der Stellen", len(stellen), 3)
}
