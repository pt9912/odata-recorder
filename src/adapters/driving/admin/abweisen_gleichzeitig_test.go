package admin_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Dieser Teil hält N1 (SPEC-238 Nr. 12): gleichzeitige Requests teilen nichts, und keine Sperre liegt über einem Aufruf von
// Handler, Append, CountError oder Vergleich.

// n1Art ist eine der fünf Requestarten von N1 mit der Erwartung als Literal des Vertrags; code ist der Code des Zählens
// (leer: die Antwort zählt nicht).
type n1Art struct {
	name   string
	bau    func() *http.Request
	status int
	body   string
	code   string
}

// n1Arten sind die fünf Requestarten: 401, 200, 404, Fehler der Route mit Code und Panik der Route.
func n1Arten() []n1Art {
	return []n1Art{
		{"401", func() *http.Request { return ohneAuth("GET", "/ok", nil) }, 401, objekt401, codeAbgewiesen},
		{"200", func() *http.Request { return anfrageMit("GET", "/ok", nil, nil) }, 200, routeBody, ""},
		{"404", func() *http.Request { return anfrageMit("GET", "/zzz", nil, nil) }, 404, objekt404, "ODR-E2102"},
		{"Fehler", func() *http.Request { return anfrageMit("GET", "/fehler", nil, nil) }, 404, objektE7003, "ODR-E7003"},
		{"Panik", func() *http.Request { return anfrageMit("GET", "/panik", nil, nil) }, 500, objekt500, "ODR-E9001"},
	}
}

// n1Umgebung baut den Server von N1 mit den Routen GET /ok (200), GET /fehler (Fehler ODR-E7003 mit 404), GET /panik
// (Panik) und GET /haengt (blockiert bis zur Freigabe); mod ändert die Konfiguration vor New. Der Server sieht das Token
// von tokenGueltig.
func n1Umgebung(t *testing.T, mod func(k *admin.Konfig)) (*umgebung, *routeFake) {
	t.Helper()
	ok, fehler, panik, haengt := neueRoute(), neueRoute(), neueRoute(), neueHaengRoute(t, wartetFrei)
	ok.antwort = admin.Antwort{Status: 200, Body: []byte(routeBody)}
	fehler.err = &inbound.Fehler{Code: model.CodeE7003, Status: 404}
	panik.aendere = func(admin.Anfrage) { panic("Panik der Route") }
	k, u := konfigMit(
		admin.Route{Methode: "GET", Muster: "/ok", Handler: ok.handle},
		admin.Route{Methode: "GET", Muster: "/fehler", Handler: fehler.handle},
		admin.Route{Methode: "GET", Muster: "/panik", Handler: panik.handle},
		admin.Route{Methode: "GET", Muster: "/haengt", Handler: haengt.handle})
	if mod != nil {
		mod(&k)
	}
	return neuMitKonfig(t, k, u), haengt
}

// n1Ergebnis ist die Antwort auf einen Request der Art art.
type n1Ergebnis struct {
	art int
	rec *httptest.ResponseRecorder
}

// parallel sendet n Requests zugleich (eine Schranke hält sie, bis alle bereit sind), die Art des i-ten ist i modulo der
// Zahl der Arten, und liefert die Antworten nach i; nach 5 s ohne Rückkehr bricht der Test ab.
func (u *umgebung) parallel(t *testing.T, n int) []n1Ergebnis {
	t.Helper()
	arten := n1Arten()
	out := make([]n1Ergebnis, n)
	schranke := neueSchranke(n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			art := i % len(arten)
			rec := httptest.NewRecorder()
			schranke.warte()
			u.srv.ServeHTTP(rec, arten[art].bau())
			out[i] = n1Ergebnis{art, rec}
		}()
	}
	fertig := make(chan struct{})
	go func() { wg.Wait(); close(fertig) }()
	select {
	case <-fertig:
	case <-time.After(5 * time.Second):
		t.Fatalf("Frist von 5 s abgelaufen: die %d gleichzeitigen Requests kehrten nicht zurück", n)
	}
	return out
}

// pruefeAntworten hält jede Antwort gegen Status und Body ihrer Art.
func pruefeAntworten(t *testing.T, name string, res []n1Ergebnis) {
	t.Helper()
	arten := n1Arten()
	for i, r := range res {
		a := arten[r.art]
		if r.rec.Code != a.status || r.rec.Body.String() != a.body {
			t.Errorf("%s, Request %d (%s): Antwort %d %q, erwartet %d %q", name, i, a.name, r.rec.Code, r.rec.Body.String(), a.status, a.body)
		}
	}
}

// TestGleichzeitigeRequestsTeilenNichts hält N1 (Nr. 12): 32 Requests zugleich (401, 200, 404, Fehler der Route, Panik, in
// dieser Folge), 20 Läufe mit je eigenem Server unter dem Detektor: jede Antwort ist die ihrer Art, jede Korrelations-ID
// kommt einmal vor und ist die des Generators, die Einträge sind genau die der Antworten 401 (je einer mit der ID seiner
// Antwort), und die Zählung ist genau die Menge der Codes der Antworten.
func TestGleichzeitigeRequestsTeilenNichts(t *testing.T) {
	const n, laeufe = 32, 20
	arten := n1Arten()
	for lauf := range laeufe {
		u, _ := n1Umgebung(t, nil)
		res := u.parallel(t, n)
		name := "Lauf " + string(rune('A'+lauf))
		pruefeAntworten(t, name, res)
		var ids, ids401 []string
		var codes []string
		for _, r := range res {
			id := r.rec.Header().Get("X-Correlation-Id")
			ids = append(ids, id)
			if arten[r.art].status == 401 {
				ids401 = append(ids401, id)
			}
			if arten[r.art].code != "" {
				codes = append(codes, arten[r.art].code)
			}
		}
		slices.Sort(ids)
		if len(slices.Compact(slices.Clone(ids))) != n || slices.Contains(ids, "") {
			t.Errorf("%s: die IDs der Antworten sind nicht %d verschiedene nichtleere Werte: %q", name, n, ids)
		}
		if k, uebrige := u.ids.zahlen(); k != n || uebrige != 0 {
			t.Errorf("%s: KorrelationsID %d-mal, andere Operationen %d-mal, erwartet %d und 0", name, k, uebrige, n)
		}
		if a := u.audit.Aufrufe(); a != len(ids401) {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", name, a, len(ids401))
			continue
		}
		var eintragsIDs []string
		for _, e := range u.audit.Eintraege() {
			eintragsIDs = append(eintragsIDs, e.KorrelationsID)
			if e != eintragSoll(t, e.KorrelationsID) {
				t.Errorf("%s: Eintrag %+v, erwartet %+v", name, e, eintragSoll(t, e.KorrelationsID))
			}
		}
		slices.Sort(eintragsIDs)
		slices.Sort(ids401)
		if !slices.Equal(eintragsIDs, ids401) {
			t.Errorf("%s: die IDs der Einträge %q sind nicht die der Antworten 401 %q", name, eintragsIDs, ids401)
		}
		gezaehlt, andere := zaehlungenVon(u.tele)
		slices.Sort(gezaehlt)
		slices.Sort(codes)
		if !slices.Equal(gezaehlt, codes) || andere != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe %d, erwartet %q", name, gezaehlt, andere, codes)
		}
	}
	gleichAuf(t, "Zahl der Läufe", laeufe, 20)
	gleichAuf(t, "Zahl der Arten", len(arten), 5)
}

// blockTele ist die Telemetrie, deren erster Aufruf von CountError auf frei wartet, nachdem er begonnen meldete.
type blockTele struct {
	support.Telemetry
	zaehler  *atomic.Int64
	begonnen chan struct{}
	frei     chan struct{}
}

func (b blockTele) CountError(c model.Code) {
	if b.zaehler.Add(1) == 1 {
		b.begonnen <- struct{}{}
		<-b.frei
	}
	b.Telemetry.CountError(c)
}

// TestBlockierterAufrufHaeltDieUebrigenNicht hält N1 (Nr. 12): keine Sperre liegt über einem Aufruf von Handler, Append,
// CountError oder Vergleich. Der erste Request bleibt in einem der vier Aufrufe stehen; die 32 folgenden (alle fünf Arten)
// kehren trotzdem zurück, Anhalten kehrt zurück, und nach der Freigabe endet auch der erste Request mit seiner Antwort.
// Die vier Stellen sind die Aufzählung von Nr. 12.
func TestBlockierterAufrufHaeltDieUebrigenNicht(t *testing.T) {
	stellen := []string{"Handler", "Append", "CountError", "Vergleich"}
	for _, stelle := range stellen {
		frei, begonnen := make(chan struct{}), make(chan struct{}, 4)
		var einmal sync.Once
		gibFrei := func() { einmal.Do(func() { close(frei) }) }
		t.Cleanup(gibFrei)
		var zaehler atomic.Int64
		haelt := func() {
			if zaehler.Add(1) == 1 {
				begonnen <- struct{}{}
				<-frei
			}
		}
		u, haengt := n1Umgebung(t, func(k *admin.Konfig) {
			switch stelle {
			case "CountError":
				k.Telemetrie = blockTele{Telemetry: k.Telemetrie, zaehler: &zaehler, begonnen: begonnen, frei: frei}
			case "Vergleich":
				k.Vergleich = func(erwartet, gegeben []byte) bool { haelt(); return admin.Vergleiche(erwartet, gegeben) }
			}
		})
		if stelle == "Append" {
			u.audit.Vor = func(support.Eintrag) { haelt() }
		}
		erster := map[string]*http.Request{
			"Handler": anfrageMit("GET", "/haengt", nil, nil), "Append": ohneAuth("GET", "/ok", nil),
			"CountError": ohneAuth("GET", "/ok", nil), "Vergleich": anfrageMit("GET", "/ok", nil, nil),
		}[stelle]
		wantStatus := map[string]int{"Handler": 200, "Append": 401, "CountError": 401, "Vergleich": 200}[stelle]
		antwort := make(chan *httptest.ResponseRecorder, 1)
		go func() {
			rec := httptest.NewRecorder()
			u.srv.ServeHTTP(rec, erster)
			antwort <- rec
		}()
		if stelle == "Handler" {
			haengt.warteBegonnen(t, 1)
		} else {
			select {
			case <-begonnen:
			case <-time.After(5 * time.Second):
				t.Fatalf("%s: der erste Request erreichte den Aufruf nicht", stelle)
			}
		}
		pruefeAntworten(t, stelle, u.parallel(t, 32))
		if !innerhalb(2*time.Second, u.srv.Anhalten) {
			t.Errorf("%s: Anhalten kehrte in 2 s nicht zurück", stelle)
		}
		gibFrei()
		haengt.gibFrei()
		select {
		case rec := <-antwort:
			if rec.Code != wantStatus {
				t.Errorf("%s: der erste Request endete mit %d, erwartet %d", stelle, rec.Code, wantStatus)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: der erste Request endete nach der Freigabe nicht", stelle)
		}
	}
	gleichAuf(t, "Zahl der Stellen", len(stellen), 4)
}
