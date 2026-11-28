//go:build linux

package admin_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Dieser Teil hält H1 (SPEC-238 Nr. 13): das Token erreicht keinen der acht Ausgabewege, auch nicht als Teilstück. Die
// Schleife läuft über Weg × Anlass; ein Anlass sendet Requests oder ruft New, Serve und Schliessen und sammelt, was jeder
// Weg sah.

// Die acht Wege in der Folge der Buchstaben (a) bis (h) von SPEC-238 Nr. 13.
const (
	wegBody = iota
	wegHeader
	wegEintrag
	wegTelemetrie
	wegFehlertexte
	wegLog
	wegStd
	wegPanik
	wegZahl
)

// wegNamen liefert die Namen der acht Wege, wie Nr. 13 sie nennt.
func wegNamen() [wegZahl]string {
	return [wegZahl]string{
		"a Body einer Antwort", "b Header einer Antwort", "c Eintrag an AuditLog", "d Argumente von Telemetry",
		"e Text eines Fehlers von New, Serve, Schliessen", "f log.Default()", "g stdout und stderr", "h Panik und ErrorLog",
	}
}

// ausgabe sammelt je Weg, was ein Anlass dort hinterließ, und zählt die Antworten und Fehlertexte, die der Anlass sah.
type ausgabe struct {
	text      [wegZahl]strings.Builder
	antworten int
	fehler    int
}

func (a *ausgabe) add(weg int, s string) {
	a.text[weg].WriteString(s)
	a.text[weg].WriteString("\n")
}

// antwort nimmt eine Antwort auf: den Body (a) und Status und Header (b).
func (a *ausgabe) antwort(rec *httptest.ResponseRecorder) {
	a.antworten++
	res := rec.Result()
	a.add(wegBody, rec.Body.String())
	a.add(wegHeader, fmt.Sprintf("%d %v", res.StatusCode, res.Header))
}

// ports nimmt auf, was Audit (c) und Telemetrie (d) sahen.
func (a *ausgabe) ports(u *umgebung) {
	a.add(wegEintrag, fmt.Sprintf("%+v", u.audit.Eintraege()))
	a.add(wegTelemetrie, fmt.Sprintf("%+v", u.tele.Aufrufe()))
}

// fehlerText nimmt den Text eines Fehlers von New, Serve oder Schliessen auf (e).
func (a *ausgabe) fehlerText(err error) {
	if err != nil {
		a.fehler++
		a.add(wegFehlertexte, err.Error())
	}
}

// sendeUndNimmAuf sendet den Request, fängt eine Panik, die ServeHTTP verlässt (h), und nimmt die Antwort auf.
func (a *ausgabe) sendeUndNimmAuf(u *umgebung, r *http.Request) {
	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if p := recover(); p != nil {
				a.add(wegPanik, fmt.Sprint(p))
			}
		}()
		u.srv.ServeHTTP(rec, r)
	}()
	a.antwort(rec)
}

// h1Funde liefert die Wege, auf denen einer der Werte oder ein Teilstück ab 8 Bytes steht, je Weg mit dem Index des
// ersten Wertes.
func h1Funde(a *ausgabe, werte []string) map[int]int {
	out := map[int]int{}
	for w := range wegZahl {
		for i, wert := range werte {
			if enthaeltTeil(a.text[w].String(), wert) {
				out[w] = i
				break
			}
		}
	}
	return out
}

// h1Anlass ist ein Anlass, bei dem das Token oder ein Wert des Requests durch den Adapter läuft: werte sind die Werte,
// die auf keinem Weg stehen dürfen, fahre setzt den Anlass um; min sind die Zahlen der Antworten und der Fehlertexte,
// die der Anlass sehen muss (sonst lief er nicht).
type h1Anlass struct {
	name          string
	werte         []string
	fahre         func(t *testing.T, a *ausgabe)
	minAntworten  int
	minFehlerText int
}

// h1Server baut einen Server mit dem Token tokenVielfalt und der Route GET /r, die r liefert.
func h1Server(t *testing.T, r *routeFake) *umgebung {
	t.Helper()
	k, u := konfigMit(admin.Route{Methode: "GET", Muster: "/r", Handler: r.handle})
	k.Token = tokenVielfalt()
	return neuMitKonfig(t, k, u)
}

// h1Anlaesse sind die sieben Anlässe: das richtige Token; ein falsches Token mit Canary; Canary in Pfad, Query, Body und
// fremdem Header; das Token in der Ursache eines Fehlers der Route; das Token im Wert einer Panik der Route und eines
// Generators; die Fehler von New, Serve und Schliessen mit Frist; das richtige Token in den 20 ungültigen Formen des
// Headers.
func h1Anlaesse() []h1Anlass {
	tok := string(tokenVielfalt())
	o := neueOrte()
	kandidat := "Kd83Pq1Wv5Lm0Xe7"
	get := func(kopf map[string][]string) *http.Request { return ohneAuth("GET", "/r", kopf) }
	return []h1Anlass{
		{"richtiges Token", []string{tok}, func(t *testing.T, a *ausgabe) {
			route := neueRoute()
			route.antwort = admin.Antwort{Status: 200, Body: []byte(routeBody)}
			u := h1Server(t, route)
			a.sendeUndNimmAuf(u, get(bearer(tok)))
			a.sendeUndNimmAuf(u, ohneAuth("GET", "/healthz", bearer(tok)))
			a.ports(u)
		}, 2, 0},
		{"falsches Token mit Canary", []string{tok, kandidat}, func(t *testing.T, a *ausgabe) {
			u := h1Server(t, neueRoute())
			a.sendeUndNimmAuf(u, get(bearer(kandidat)))
			a.sendeUndNimmAuf(u, get(map[string][]string{"Authorization": {"Bearer " + tok, "Bearer " + kandidat}}))
			a.ports(u)
		}, 2, 0},
		{"Canary in Pfad, Query, Body und fremdem Header", append([]string{tok}, o.pfad, o.query, o.header, o.body), func(t *testing.T, a *ausgabe) {
			route := neueRoute()
			u := h1Server(t, route)
			for _, kand := range []string{"", tok} {
				for _, methode := range []string{"GET", "POST"} {
					a.sendeUndNimmAuf(u, o.request(methode, "/r", kand))
				}
			}
			r := httptest.NewRequest("GET", "/r?q="+o.query, strings.NewReader(o.body))
			r.Header.Set("X-Test", o.header)
			r.Header.Set("Authorization", "Bearer "+tok)
			a.sendeUndNimmAuf(u, r)
			a.ports(u)
		}, 5, 0},
		{"Token in der Ursache eines Fehlers der Route", []string{tok}, func(t *testing.T, a *ausgabe) {
			fehler := []error{
				&inbound.Fehler{Code: model.CodeE7003, Status: 404, Ursache: errors.New("Ursache " + tok)},
				fmt.Errorf("%w: %s", model.ErrUngueltig, tok),
				errors.New(tok),
				fehlerE2101(errors.New(tok)),
			}
			for _, err := range fehler {
				route := neueRoute()
				route.err = err
				u := h1Server(t, route)
				a.sendeUndNimmAuf(u, get(bearer(tok)))
				a.ports(u)
			}
		}, 4, 0},
		{"Token im Wert einer Panik", []string{tok, string(tokenGueltig())}, func(t *testing.T, a *ausgabe) {
			for _, wert := range []any{tok, errors.New(tok), &inbound.Fehler{Code: model.CodeE7003, Status: 404, Ursache: errors.New(tok)}} {
				route := neueRoute()
				route.aendere = func(admin.Anfrage) { panic(wert) }
				u := h1Server(t, route)
				a.sendeUndNimmAuf(u, get(bearer(tok)))
				a.ports(u)
			}
			// Der Generator der Korrelations-ID gerät in Panik (außerhalb der Route): net/http fängt sie und schreibt in
			// ErrorLog; der Wert ist das Token des Servers.
			lv := neuLiveMit(t, neueRoute(), liveOpt{vorKorr: func() { panic(string(tokenGueltig())) }})
			if x := erwarte(t, lv.async("GET", "/p"), "Panik des Generators"); x.err == nil {
				t.Errorf("Panik des Generators: der Client sah einen Status %d, erwartet einen Transportfehler", x.status)
			}
			a.antworten++
		}, 4, 0},
		{"Fehler von New, Serve und Schliessen", []string{tok, string(tokenGueltig()), canary()}, func(t *testing.T, a *ausgabe) {
			for _, f := range newFehlerFaelle() {
				_, err := admin.New(f.konfig())
				if err == nil {
					t.Errorf("New %s: kein Fehler, der Zweig wird nicht erreicht", f.name)
				}
				a.fehlerText(err)
			}
			srv := neuMitListener(t, neuerListener())
			a.fehlerText(srv.Serve())
			route := neueHaengRoute(t, wartetCtx)
			lv := neuLive(t, route)
			r := lv.async("GET", "/haengt")
			route.warteBegonnen(t, 1)
			lv.srv.Anhalten()
			_, err := schliesse(t, lv.srv, 20*time.Millisecond)
			a.fehlerText(err)
			_, err = schliesse(t, lv.srv, -time.Second)
			a.fehlerText(err)
			route.gibFrei()
			erwarte(t, r, "Antwort nach Schliessen")
			a.fehlerText(serveEnde(t, lv.serve, "Serve nach Schliessen"))
			a.portsLive(lv)
		}, 0, len(newFehlerFaelle()) + 2},
		{"richtiges Token in den 20 ungültigen Formen", []string{tok}, func(t *testing.T, a *ausgabe) {
			for _, f := range u1Formen(tok) {
				u := h1Server(t, neueRoute())
				a.sendeUndNimmAuf(u, get(f.kopf))
				a.ports(u)
			}
		}, 20, 0},
	}
}

// newFehlerFall ist eine Konfiguration, die genau einen der Zweige von New mit Fehler auslöst; Token und Werte der Routen
// tragen Werte, die im Text des Fehlers nicht stehen dürfen (das Token von tokenVielfalt, den Canary als Methode und Muster).
type newFehlerFall struct {
	name   string
	konfig func() admin.Konfig
}

// newFehlerFaelle sind die zehn Zweige von New, die einen Fehler mit einem Text bilden (SPEC-238 Nr. 1, 2): Listener, IDs,
// Audit, Telemetrie, Uhr fehlt, Token zu kurz (31 Bytes) und die vier Verstöße einer Route (Methode, Muster, Handler,
// Überlappung).
func newFehlerFaelle() []newFehlerFall {
	basis := func(routen ...admin.Route) admin.Konfig {
		k, _ := konfigMit(routen...)
		k.Token = tokenVielfalt()
		return k
	}
	handler := neueRoute().handle
	muster := "/" + canary()
	return []newFehlerFall{
		{"Listener fehlt", func() admin.Konfig { k := basis(); k.Listener = nil; return k }},
		{"IDs fehlt", func() admin.Konfig { k := basis(); k.IDs = nil; return k }},
		{"Audit fehlt", func() admin.Konfig { k := basis(); k.Audit = nil; return k }},
		{"Telemetrie fehlt", func() admin.Konfig { k := basis(); k.Telemetrie = nil; return k }},
		{"Uhr fehlt", func() admin.Konfig { k := basis(); k.Uhr = nil; return k }},
		{"Token zu kurz", func() admin.Konfig { k := basis(); k.Token = tokenVielfalt()[:31]; return k }},
		{"Route mit ungültiger Methode", func() admin.Konfig {
			return basis(admin.Route{Methode: canary(), Muster: muster, Handler: handler})
		}},
		{"Route mit ungültigem Muster", func() admin.Konfig {
			return basis(admin.Route{Methode: "GET", Muster: canary() + "/x", Handler: handler})
		}},
		{"Route ohne Handler", func() admin.Konfig {
			return basis(admin.Route{Methode: "GET", Muster: muster})
		}},
		{"Routen überlappen", func() admin.Konfig {
			return basis(admin.Route{Methode: "GET", Muster: muster, Handler: handler}, admin.Route{Methode: "GET", Muster: muster, Handler: handler})
		}},
	}
}

// zaehleUngueltig zählt die Aufrufe von `ungueltig(…)` in den Quellen des Pakets ohne _test.go, getrennt nach Datei, und
// ohne die Definition der Funktion.
func zaehleUngueltig(t *testing.T) map[string]int {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]int{}
	for _, e := range eintraege {
		n := e.Name()
		if !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(x ast.Node) bool {
			if c, ok := x.(*ast.CallExpr); ok {
				if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "ungueltig" {
					out[n]++
				}
			}
			return true
		})
	}
	return out
}

// TestNewFehlerFaelleGegenQuelle hält die Menge von H1 Weg e (jeder Fehlertext von New) gegen den Quelltext: jeder Aufruf von
// `ungueltig(…)` außer dem in `Schliessen` (den der Anlass mit negativer Frist fährt) hat einen Fall, jeder Fall löst einen Fehler aus,
// und die Texte der Fälle sind paarweise verschieden (jeder Fall erreicht seinen eigenen Zweig). Keine feste Zahl.
func TestNewFehlerFaelleGegenQuelle(t *testing.T) {
	quelle := zaehleUngueltig(t)
	gesamt := 0
	for _, n := range quelle {
		gesamt += n
	}
	if quelle["lebenszyklus.go"] != 1 {
		t.Errorf("Aufrufe von ungueltig in lebenszyklus.go: %d, erwartet 1 (Schliessen, negative Frist)", quelle["lebenszyklus.go"])
	}
	faelle := newFehlerFaelle()
	gleichAuf(t, "Aufrufe von ungueltig im Quelltext ohne Schliessen und Fälle von New", len(faelle), gesamt-quelle["lebenszyklus.go"])
	texte := map[string]string{}
	for _, f := range faelle {
		_, err := admin.New(f.konfig())
		if err == nil {
			t.Errorf("%s: New liefert keinen Fehler", f.name)
			continue
		}
		if !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v, erwartet model.ErrUngueltig", f.name, err)
		}
		if vorher, da := texte[err.Error()]; da {
			t.Errorf("%s und %s liefern denselben Text %q", f.name, vorher, err)
		}
		texte[err.Error()] = f.name
	}
}

// portsLive nimmt auf, was Audit (c) und Telemetrie (d) des laufenden Servers sahen.
func (a *ausgabe) portsLive(u *live) {
	a.add(wegEintrag, fmt.Sprintf("%+v", u.audit.Eintraege()))
	a.add(wegTelemetrie, fmt.Sprintf("%+v", u.tele.Aufrufe()))
}

// TestKeinTokenAufAchtWegen hält H1 (Nr. 13): Schleife Weg × Anlass, 8 × 7 = 56 Zellen: auf keinem der acht Wege steht
// ein Wert des Anlasses (das Token, der Canary) oder ein Teilstück ab 8 Bytes. Die Wege f und g (log.Default(), stdout und
// stderr) liest die Erfassung des Prozesses (Dateideskriptor 1 und 2 und das Standard-Log); der Weg h trägt den Wert einer
// Panik, die ServeHTTP verließ, und die Panik des Generators, die net/http an ErrorLog meldet (verworfen, sonst stünde sie
// auf f oder g). Die Zahl der Wege ist die der Buchstaben (a) bis (h) von Nr. 13: die Spezifikation liegt im Container
// der Tests nicht vor, die Zahl ist die der Aufzählung dort.
func TestKeinTokenAufAchtWegen(t *testing.T) {
	anlaesse := h1Anlaesse()
	namen := wegNamen()
	gleichAuf(t, "Zahl der Wege", len(namen[:]), 8)
	gleichAuf(t, "Zahl der Anlässe", len(anlaesse), 7)
	zellen := 0
	for _, an := range anlaesse {
		a := &ausgabe{}
		aus, fehl, logs := erfasse(t, func() { an.fahre(t, a) })
		a.add(wegLog, logs)
		a.add(wegStd, aus+fehl)
		if a.antworten < an.minAntworten || a.fehler < an.minFehlerText {
			t.Errorf("%s: %d Antworten und %d Fehlertexte gesehen, erwartet mindestens %d und %d", an.name, a.antworten, a.fehler, an.minAntworten, an.minFehlerText)
		}
		funde := h1Funde(a, an.werte)
		for w := range wegZahl {
			zellen++
			if i, da := funde[w]; da {
				t.Errorf("Anlass %q, Weg %s: der Wert Nr. %d (%q) oder ein Teilstück steht dort: %q", an.name, wegNamen()[w], i, an.werte[i], a.text[w].String())
			}
		}
	}
	gleichAuf(t, "Zahl der Zellen", zellen, 8*7)
}

// TestH1ProbeSiehtJedenWeg hält den Fänger von H1 gegen eine Probe: ein Wert, ein Teilstück von 8 Bytes am Anfang, in der
// Mitte und am Ende, auf genau einem der acht Wege, wird dort gefunden und nirgends sonst; ein Teilstück von 7 Bytes
// nicht (Prüfgrenze: ab 8 Bytes, Folge im Betrieb: ein Fehler, der nur 7 Bytes des Tokens ausgäbe, bliebe grün).
func TestH1ProbeSiehtJedenWeg(t *testing.T) {
	tok := string(tokenVielfalt())
	teile := map[string]string{"ganz": tok, "Anfang": tok[:8], "Mitte": tok[14:22], "Ende": tok[len(tok)-8:]}
	gleichAuf(t, "Zahl der Teile", len(teile), 4)
	for name, teil := range teile {
		for w := range wegZahl {
			a := &ausgabe{}
			a.add(w, "vorher "+teil+" nachher")
			funde := h1Funde(a, []string{tok})
			if _, da := funde[w]; !da || len(funde) != 1 {
				t.Errorf("Teil %s auf Weg %s: Funde %v, erwartet genau diesen Weg", name, wegNamen()[w], funde)
			}
		}
	}
	a := &ausgabe{}
	a.add(wegBody, tok[:7])
	if funde := h1Funde(a, []string{tok}); len(funde) != 0 {
		t.Errorf("7 Bytes: Funde %v, erwartet keine", funde)
	}
}
