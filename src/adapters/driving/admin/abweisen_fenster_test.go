package admin_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Dieser Teil hält SPEC-301: höchstens 20 Einträge admin_abgewiesen je Fenster von 60 s ab der Abweisung, die es eröffnet;
// jede weitere Abweisung ist beantwortet und gezählt, ohne Eintrag. Grenze, Dauer, Status und Header stehen als Literale
// des Vertrags, nie aus dem Code.

// fensterT0 ist der Stand der Uhr, mit dem die Tests des Fensters beginnen: er trägt Nanosekunden und liegt auf keiner
// vollen Minute.
func fensterT0() time.Time { return time.Date(2026, 11, 1, 12, 0, 0, 999_999_999, time.UTC) }

// stellUhr ist eine Uhr, deren Stand der Test setzt, auch zurück oder außerhalb der Jahre 1 bis 9999. Sie zählt die
// Aufrufe von Now; vor läuft in jedem Aufruf vor dem Lesen und erhält dessen Nummer ab 1; schritt rückt den Stand je
// Aufruf vor.
type stellUhr struct {
	mu      sync.Mutex
	stand   time.Time
	schritt time.Duration
	aufrufe int
	vor     func(n int)
}

func (u *stellUhr) Now() time.Time {
	u.mu.Lock()
	u.aufrufe++
	n, vor := u.aufrufe, u.vor
	u.mu.Unlock()
	if vor != nil {
		vor(n)
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stand = u.stand.Add(u.schritt)
	return u.stand
}

func (u *stellUhr) setze(t time.Time) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stand = t
}

func (u *stellUhr) vorruecken(d time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stand = u.stand.Add(d)
}

func (u *stellUhr) zahl() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.aufrufe
}

// fensterRouten sind die Routen der Tests des Fensters: GET /ok (200), GET /e401 (Fehler der Route ODR-E2101 mit 401),
// GET /e400, GET /e503 und GET /e500 mit den Fehlern ihrer Antwort. Die Karte hält den Fake je Muster.
func fensterRouten() ([]admin.Route, map[string]*routeFake) {
	fehler := map[string]error{
		"/ok": nil, "/e401": fehlerE2101(nil), "/e400": model.ErrUngueltig,
		"/e503": &inbound.Fehler{Code: model.CodeE5010, Status: 503}, "/e500": errors.New("Ablage"),
	}
	var routen []admin.Route
	fakes := map[string]*routeFake{}
	for muster, err := range fehler {
		f := neueRoute()
		f.err = err
		fakes[muster] = f
		routen = append(routen, admin.Route{Methode: "GET", Muster: muster, Handler: f.handle})
	}
	return routen, fakes
}

// fensterServer baut einen Server mit der Uhr uhr und den Routen von fensterRouten.
func fensterServer(t *testing.T, uhr support.Clock) (*umgebung, map[string]*routeFake) {
	t.Helper()
	routen, fakes := fensterRouten()
	k, u := konfigMit(routen...)
	k.Uhr = uhr
	return neuMitKonfig(t, k, u), fakes
}

// weiseAb sendet n Abweisungen ohne Header Authorization und verlangt von jeder 401.
func weiseAb(t *testing.T, u *umgebung, n int) {
	t.Helper()
	for i := range n {
		if rec := u.rufe(ohneAuth("GET", "/ok", nil)); rec.Code != 401 {
			t.Fatalf("Abweisung %d: Status %d, erwartet 401", i+1, rec.Code)
		}
	}
}

// zahlEintraege liefert die Zahl der Einträge, die das Audit hält.
func zahlEintraege(u *umgebung) int { return len(u.audit.Eintraege()) }

// zugelassen sendet eine Abweisung und meldet, ob sie einen Eintrag schrieb.
func zugelassen(t *testing.T, u *umgebung) bool {
	t.Helper()
	vorher := zahlEintraege(u)
	weiseAb(t, u, 1)
	return zahlEintraege(u) == vorher+1
}

// pruefeFensterVoll lässt ein neues Fenster erwarten: die nächste Abweisung und 19 weitere sind zugelassen, die 21. ist
// gedrosselt.
func pruefeFensterVoll(t *testing.T, name string, u *umgebung) {
	t.Helper()
	vorher := zahlEintraege(u)
	weiseAb(t, u, 20)
	if n := zahlEintraege(u) - vorher; n != 20 {
		t.Errorf("%s: %d von 20 Abweisungen zugelassen, erwartet 20", name, n)
	}
	if zugelassen(t, u) {
		t.Errorf("%s: die 21. Abweisung ist zugelassen", name)
	}
}

// TestNewUhr hält W01 (SPEC-301 Nr. 1): eine nil-Uhr ist model.ErrUngueltig ohne Server, der Text nennt weder das Token
// noch einen Wert; New ruft die Uhr nicht; Konfig führt genau diese acht Felder, keines für Zahl oder Dauer.
func TestNewUhr(t *testing.T) {
	k, _ := konfigMit()
	k.Token, k.Uhr = tokenVielfalt(), nil
	srv, err := admin.New(k)
	if !errors.Is(err, model.ErrUngueltig) || srv != nil {
		t.Errorf("nil-Uhr: Fehler %v, Server %v, erwartet ErrUngueltig ohne Server", err, srv)
	}
	if err != nil && enthaeltTeil(err.Error(), string(k.Token)) {
		t.Errorf("der Text nennt das Token: %q", err)
	}
	uhr := &stellUhr{stand: fensterT0()}
	k.Uhr = uhr
	if srv, err := admin.New(k); err != nil || srv == nil {
		t.Fatalf("mit Uhr: %v", err)
	}
	if n := uhr.zahl(); n != 0 {
		t.Errorf("New rief die Uhr %d-mal, erwartet 0", n)
	}
	var felder []string
	for _, f := range reflect.VisibleFields(reflect.TypeOf(admin.Konfig{})) {
		felder = append(felder, f.Name)
	}
	gleichAuf(t, "Felder von Konfig", felder, []string{"Listener", "Token", "IDs", "Audit", "Telemetrie", "Routen", "Vergleich", "Uhr"})
}

// TestFensterZahl hält W02 (Nr. 2 c, 5 a): bei stehender Uhr ergeben n Abweisungen min(n, 20) Einträge, n von 1 bis 25,
// mit den Zeilen 19 bis 22 als Literal; die Zahl des Fensters ist dieselbe und steigt nie über 20.
func TestFensterZahl(t *testing.T) {
	zeilen := map[int]int{19: 19, 20: 20, 21: 20, 22: 20}
	faelle := 0
	for n := 1; n <= 25; n++ {
		faelle++
		u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
		weiseAb(t, u, n)
		soll := min(n, 20)
		if z, da := zeilen[n]; da {
			soll = z
		}
		if got := zahlEintraege(u); got != soll {
			t.Errorf("%d Abweisungen: %d Einträge, erwartet %d", n, got, soll)
		}
		if got := u.srv.FensterZahlFuerTest(); got != soll {
			t.Errorf("%d Abweisungen: Zahl des Fensters %d, erwartet %d", n, got, soll)
		}
	}
	gleichAuf(t, "Zahl der Werte von n", faelle, 25)
}

// TestFensterGrenze hält W03 (Nr. 2 b, 5 b): nach 20 Zulassungen ist die nächste Abweisung bei 59 s − 1 ns, 59 s und
// 60 s − 1 ns gedrosselt, bei 60 s und 61 s zugelassen; an der Grenze beginnt ein neues Fenster mit 20 Zulassungen.
func TestFensterGrenze(t *testing.T) {
	faelle := []struct {
		name       string
		nach       time.Duration
		zugelassen bool
	}{
		{"59 s - 1 ns", 59*time.Second - time.Nanosecond, false},
		{"59 s", 59 * time.Second, false},
		{"60 s - 1 ns", 60*time.Second - time.Nanosecond, false},
		{"60 s", 60 * time.Second, true},
		{"61 s", 61 * time.Second, true},
	}
	for _, f := range faelle {
		uhr := &stellUhr{stand: fensterT0()}
		u, _ := fensterServer(t, uhr)
		weiseAb(t, u, 20)
		uhr.vorruecken(f.nach)
		if got := zugelassen(t, u); got != f.zugelassen {
			t.Errorf("%s nach dem Beginn: zugelassen %v, erwartet %v", f.name, got, f.zugelassen)
		}
	}
	gleichAuf(t, "Zahl der Abstände", len(faelle), 5)
	uhr := &stellUhr{stand: fensterT0()}
	u, _ := fensterServer(t, uhr)
	pruefeFensterVoll(t, "erstes Fenster", u)
	uhr.vorruecken(60*time.Second - time.Nanosecond)
	if zugelassen(t, u) {
		t.Error("60 s - 1 ns: zugelassen")
	}
	uhr.vorruecken(time.Nanosecond)
	pruefeFensterVoll(t, "Fenster an der Grenze", u)
}

// TestFensterBeginn hält W04 (Nr. 2 b, 5 c): der Beginn ist die Zeit der Abweisung, die das Fenster eröffnet, nicht die
// von New und kein Vielfaches von 60 s; der Beginn eines neuen Fensters gilt für dieses; eine Ruhe von einem Tag öffnet
// ein Fenster.
func TestFensterBeginn(t *testing.T) {
	uhr := &stellUhr{stand: fensterT0()}
	u, _ := fensterServer(t, uhr)
	t1 := fensterT0().Add(5*time.Minute + 17*time.Second)
	uhr.setze(t1)
	weiseAb(t, u, 20)
	if got := zahlEintraege(u); got != 20 {
		t.Fatalf("%d Einträge im ersten Fenster, erwartet 20", got)
	}
	uhr.setze(t1.Add(60*time.Second - time.Nanosecond))
	if zugelassen(t, u) {
		t.Error("59,999999999 s nach dem Beginn: zugelassen")
	}
	uhr.setze(t1.Add(60 * time.Second))
	pruefeFensterVoll(t, "60 s nach dem Beginn", u)
	uhr.vorruecken(24 * time.Hour)
	pruefeFensterVoll(t, "ein Tag später", u)

	uhr = &stellUhr{stand: fensterT0()}
	u, _ = fensterServer(t, uhr)
	uhr.setze(fensterT0().Add(30 * time.Second))
	weiseAb(t, u, 20)
	uhr.setze(fensterT0().Add(60 * time.Second))
	if zugelassen(t, u) {
		t.Error("60 s nach New, 30 s nach dem Beginn: zugelassen")
	}
	uhr.setze(fensterT0().Add(90 * time.Second))
	if !zugelassen(t, u) {
		t.Error("60 s nach dem Beginn: gedrosselt")
	}
}

// TestFensterBeginnBleibt hält W04 (Nr. 2 b, 2 c, 5 c): der Beginn ist allein die Zeit der eröffnenden Abweisung, keine
// spätere verschiebt ihn. Jede Abweisung eines Fensters liegt auf einem anderen Stand der Uhr: 20 zugelassene bei 0, 2
// bis 38 s, 5 gedrosselte bei 40 bis 48 s; danach ist 60 s − 1 ns gedrosselt und 60 s zugelassen, in zwei Fenstern
// nacheinander.
func TestFensterBeginnBleibt(t *testing.T) {
	uhr := &stellUhr{stand: fensterT0()}
	u, _ := fensterServer(t, uhr)
	beginn, stellen := fensterT0(), 0
	for fenster := range 2 {
		for i := range 25 {
			stellen++
			uhr.setze(beginn.Add(time.Duration(2*i) * time.Second))
			if got, soll := zugelassen(t, u), i < 20; got != soll {
				t.Errorf("Fenster %d, Abweisung %d bei %d s: zugelassen %v, erwartet %v", fenster+1, i+1, 2*i, got, soll)
			}
		}
		uhr.setze(beginn.Add(60*time.Second - time.Nanosecond))
		if zugelassen(t, u) {
			t.Errorf("Fenster %d, 60 s − 1 ns nach dem Beginn: zugelassen", fenster+1)
		}
		beginn = beginn.Add(60 * time.Second)
	}
	uhr.setze(beginn)
	if !zugelassen(t, u) {
		t.Error("60 s nach dem Beginn des zweiten Fensters: gedrosselt")
	}
	gleichAuf(t, "Zahl der Abweisungen auf eigenem Stand", stellen, 50)
}

// TestFensterEineLesung hält W09 (Nr. 2 a): jede Antwort 401 ruft Now() genau einmal: die eröffnende, eine zugelassene,
// eine gedrosselte, die eines neuen Fensters nach 60 s und eine bei einer Uhr vor dem Beginn, je aus der Authentisierung
// und aus der Route (10 Zellen).
func TestFensterEineLesung(t *testing.T) {
	wege := map[string]func() *http.Request{
		"Authentisierung": func() *http.Request { return ohneAuth("GET", "/ok", nil) },
		"Route":           func() *http.Request { return anfrageMit("GET", "/e401", nil, nil) },
	}
	zellen := 0
	for name, weg := range wege {
		uhr := &stellUhr{stand: fensterT0()}
		u, _ := fensterServer(t, uhr)
		pruefe := func(fall string, zulassung bool) {
			t.Helper()
			zellen++
			aufrufe, vorher := uhr.zahl(), zahlEintraege(u)
			if rec := u.rufe(weg()); rec.Code != 401 {
				t.Fatalf("%s, %s: Status %d, erwartet 401", name, fall, rec.Code)
			}
			if n := uhr.zahl() - aufrufe; n != 1 {
				t.Errorf("%s, %s: %d Aufrufe von Now, erwartet 1", name, fall, n)
			}
			if got := zahlEintraege(u) == vorher+1; got != zulassung {
				t.Errorf("%s, %s: zugelassen %v, erwartet %v", name, fall, got, zulassung)
			}
		}
		pruefe("Eröffnung", true)
		pruefe("Zulassung im Fenster", true)
		weiseAb(t, u, 18)
		pruefe("Drosselung", false)
		uhr.vorruecken(60 * time.Second)
		pruefe("neues Fenster", true)
		weiseAb(t, u, 19)
		uhr.setze(fensterT0())
		pruefe("Uhr vor dem Beginn", false)
	}
	gleichAuf(t, "Zahl der Zellen", zellen, 10)
}

// TestFensterUhrWerte hält W05 (Nr. 5 c): ein Wert vor dem Beginn (1 ns, 1 Jahr, die Nullzeit, 10000 Jahre davor und
// damit außerhalb der Jahre 1 bis 9999) lässt das Fenster offen, ohne Panik, und verschiebt den Beginn nicht (danach ist
// 60 s − 1 ns nach ihm gedrosselt, 60 s zugelassen); die Nullzeit als erste Zeit eröffnet genau ein Fenster; ein Wert
// nach dem Beginn (Jahr 9999) eröffnet ein neues.
func TestFensterUhrWerte(t *testing.T) {
	b := fensterT0()
	if y := b.AddDate(-10000, 0, 0).Year(); y >= 1 {
		t.Fatalf("10000 Jahre vor dem Beginn liegt im Jahr %d, erwartet vor Jahr 1", y)
	}
	faelle := []struct {
		name string
		zeit time.Time
	}{
		{"1 ns davor", b.Add(-time.Nanosecond)},
		{"1 Jahr davor", b.AddDate(-1, 0, 0)},
		{"Nullzeit", time.Time{}},
		{"10000 Jahre davor", b.AddDate(-10000, 0, 0)},
	}
	for _, f := range faelle {
		uhr := &stellUhr{stand: b}
		u, _ := fensterServer(t, uhr)
		weiseAb(t, u, 20)
		uhr.setze(f.zeit)
		rec := u.rufeSicher(t, f.name, ohneAuth("GET", "/ok", nil))
		if rec.Code != 401 || zahlEintraege(u) != 20 {
			t.Errorf("%s: Status %d, %d Einträge, erwartet 401 und 20", f.name, rec.Code, zahlEintraege(u))
		}
		uhr.setze(b.Add(60*time.Second - time.Nanosecond))
		if zugelassen(t, u) {
			t.Errorf("%s, dann 60 s − 1 ns nach dem Beginn: zugelassen", f.name)
		}
		uhr.setze(b.Add(60 * time.Second))
		if !zugelassen(t, u) {
			t.Errorf("%s, dann 60 s nach dem Beginn: gedrosselt", f.name)
		}
	}
	gleichAuf(t, "Zahl der Werte vor dem Beginn", len(faelle), 4)

	uhr := &stellUhr{}
	u, _ := fensterServer(t, uhr)
	pruefeFensterVoll(t, "Nullzeit als erste Zeit", u)
	uhr.setze(b)
	pruefeFensterVoll(t, "nach der Nullzeit eine Zeit von heute", u)

	uhr = &stellUhr{stand: b}
	u, _ = fensterServer(t, uhr)
	pruefeFensterVoll(t, "Beginn von heute", u)
	uhr.setze(time.Date(9999, 6, 1, 0, 0, 0, 0, time.UTC))
	pruefeFensterVoll(t, "Jahr 9999 nach dem Beginn", u)
}

// anlassFall ist ein Anlass einer Antwort 401 für W06.
type anlassFall struct {
	name string
	bau  func() *http.Request
}

// anlaesse401 sind die fünf Anlässe von W06: kein Header, leerer Wert, falsches Token, Schema Basic und der Fehler einer
// Route mit ODR-E2101.
func anlaesse401() []anlassFall {
	return []anlassFall{
		{"kein Header", func() *http.Request { return ohneAuth("GET", "/ok", nil) }},
		{"leerer Wert", func() *http.Request { return ohneAuth("GET", "/ok", map[string][]string{"Authorization": {""}}) }},
		{"falsches Token", func() *http.Request { return ohneAuth("GET", "/ok", bearer("falsch")) }},
		{"Schema Basic", func() *http.Request {
			return ohneAuth("GET", "/ok", map[string][]string{"Authorization": {"Basic " + string(tokenGueltig())}})
		}},
		{"Fehler der Route", func() *http.Request { return anfrageMit("GET", "/e401", nil, nil) }},
	}
}

// pruefe401OhneKennung hält eine Antwort gegen das 401 des Vertrags: Status, Body und genau die Header Content-Type,
// Content-Length, WWW-Authenticate und X-Correlation-Id (nichtleer); sie liefert die Header ohne X-Correlation-Id.
func pruefe401OhneKennung(t *testing.T, name string, rec *httptest.ResponseRecorder) (int, string, http.Header) {
	t.Helper()
	const body = `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`
	kopf := rec.Result().Header.Clone()
	id := kopf.Get("X-Correlation-Id")
	kopf.Del("X-Correlation-Id")
	soll := http.Header{"Content-Type": {"application/json"}, "Content-Length": {strconv.Itoa(len(body))}, "Www-Authenticate": {"Bearer"}}
	if rec.Code != 401 || rec.Body.String() != body || id == "" || !reflect.DeepEqual(kopf, soll) {
		t.Errorf("%s: Status %d, Body %q, X-Correlation-Id %q, Header %v, erwartet 401 und %v", name, rec.Code, rec.Body.String(), id, kopf, soll)
	}
	return rec.Code, rec.Body.String(), kopf
}

// TestGedrosseltAntwort hält W06 (Nr. 3): die gedrosselte Antwort gleicht der zugelassenen in Status, Body und Headern bis
// auf X-Correlation-Id, ohne Retry-After und ohne weiteren Header; 5 Anlässe × (zugelassen, gedrosselt) = 10 Zellen.
func TestGedrosseltAntwort(t *testing.T) {
	zellen := 0
	for _, a := range anlaesse401() {
		u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
		erste := u.rufe(a.bau())
		weiseAb(t, u, 19)
		if n := zahlEintraege(u); n != 20 {
			t.Fatalf("%s: %d Einträge vor der gedrosselten, erwartet 20", a.name, n)
		}
		zweite := u.rufe(a.bau())
		if n := zahlEintraege(u); n != 20 {
			t.Errorf("%s: die 21. Abweisung ist zugelassen (%d Einträge)", a.name, n)
		}
		s1, b1, k1 := pruefe401OhneKennung(t, a.name+" zugelassen", erste)
		s2, b2, k2 := pruefe401OhneKennung(t, a.name+" gedrosselt", zweite)
		zellen += 2
		if s1 != s2 || b1 != b2 || !reflect.DeepEqual(k1, k2) {
			t.Errorf("%s: zugelassen %d %q %v, gedrosselt %d %q %v", a.name, s1, b1, k1, s2, b2, k2)
		}
	}
	gleichAuf(t, "Zahl der Zellen", zellen, 10)
}

// TestGedrosseltZaehlt hält W07 (Nr. 3, 4): jede der 25 Abweisungen, auch jede gedrosselte, ruft CountError(ODR-E2101)
// genau einmal und sonst keine Operation der Telemetrie, und KorrelationsID genau einmal; die ID der n-ten Antwort ist die
// des n-ten Aufrufs.
func TestGedrosseltZaehlt(t *testing.T) {
	u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	var ids []string
	for i := range 25 {
		rec := u.rufe(abweisungsRequest(i))
		if rec.Code != 401 {
			t.Fatalf("Abweisung %d: Status %d", i+1, rec.Code)
		}
		ids = append(ids, rec.Header().Get("X-Correlation-Id"))
	}
	if n := zahlEintraege(u); n != 20 {
		t.Errorf("%d Einträge, erwartet 20", n)
	}
	codes, andere := zaehlungenVon(u.tele)
	gleichAuf(t, "Aufrufe von CountError", codes, slices.Repeat([]string{"ODR-E2101"}, 25))
	gleichAuf(t, "andere Aufrufe der Telemetrie", andere, 0)
	if k, uebrige := u.ids.zahlen(); k != 25 || uebrige != 0 {
		t.Errorf("KorrelationsID %d-mal, andere %d-mal, erwartet 25 und 0", k, uebrige)
	}
	gleichAuf(t, "IDs der Antworten", ids, u.ids.vergeben)
}

// TestGedrosseltOhnePort hält W08 (Nr. 3): nach 20 Zulassungen ergeben 5 weitere Abweisungen keinen Append und keinen
// Aufruf eines anderen Ports außer KorrelationsID und CountError; keine Route läuft; die 20 Einträge tragen Art, Code und
// die ID ihrer Antwort und sonst kein Feld.
func TestGedrosseltOhnePort(t *testing.T) {
	u, fakes := fensterServer(t, &stellUhr{stand: fensterT0()})
	var ids []string
	for i := range 25 {
		rec := u.rufe(abweisungsRequest(i % 3))
		if rec.Code != 401 {
			t.Fatalf("Abweisung %d: Status %d", i+1, rec.Code)
		}
		ids = append(ids, rec.Header().Get("X-Correlation-Id"))
	}
	gleichAuf(t, "Aufrufe von Audit", u.audit.Aufrufe(), 20)
	var soll []support.Eintrag
	for _, id := range ids[:20] {
		soll = append(soll, eintragSoll(t, id))
	}
	gleichAuf(t, "Einträge", u.audit.Eintraege(), soll)
	for muster, f := range fakes {
		if n := f.aufrufe(); n != 0 {
			t.Errorf("Route %s: %d Aufrufe, erwartet 0", muster, n)
		}
	}
	if _, andere := zaehlungenVon(u.tele); andere != 0 {
		t.Errorf("andere Aufrufe der Telemetrie: %d", andere)
	}
	if a, c, ad := u.lst.zahlen(); a+c+ad != 0 {
		t.Errorf("Listener: %d/%d/%d Aufrufe", a, c, ad)
	}
}

// antwortenOhne401 sind die sieben Antworten, die keine Zulassung holen, mit ihrem Status (SPEC-301 Nr. 5 h).
func antwortenOhne401() []struct {
	name   string
	bau    func() *http.Request
	status int
} {
	return []struct {
		name   string
		bau    func() *http.Request
		status int
	}{
		{"404", func() *http.Request { return anfrageMit("GET", "/zzz", nil, nil) }, 404},
		{"405", func() *http.Request { return anfrageMit("POST", "/ok", nil, nil) }, 405},
		{"400", func() *http.Request { return anfrageMit("GET", "/e400", nil, nil) }, 400},
		{"503", func() *http.Request { return anfrageMit("GET", "/e503", nil, nil) }, 503},
		{"500", func() *http.Request { return anfrageMit("GET", "/e500", nil, nil) }, 500},
		{"200", func() *http.Request { return anfrageMit("GET", "/ok", nil, nil) }, 200},
		{"GET /healthz", func() *http.Request { return ohneAuth("GET", "/healthz", nil) }, 200},
	}
}

// TestFensterNurAbweisung hält W09 (Nr. 2, 5 g, 5 h): keine der 7 Antworten ohne 401 ruft die Uhr oder ändert das Fenster,
// weder bei leerem noch bei vollem Fenster (14 Zellen); die Authentisierung und die Route teilen ein Fenster: je 10
// ergeben 20 Zulassungen, danach ist die 21. aus jedem Weg gedrosselt.
func TestFensterNurAbweisung(t *testing.T) {
	zellen := 0
	for _, a := range antwortenOhne401() {
		for _, voll := range []bool{false, true} {
			zellen++
			name := a.name + ", Fenster leer"
			uhr := &stellUhr{stand: fensterT0()}
			u, _ := fensterServer(t, uhr)
			if voll {
				name = a.name + ", Fenster voll"
				weiseAb(t, u, 20)
				uhr.vorruecken(30 * time.Second)
			}
			vorher, aufrufe := zahlEintraege(u), uhr.zahl()
			if rec := u.rufe(a.bau()); rec.Code != a.status {
				t.Fatalf("%s: Status %d, erwartet %d", name, rec.Code, a.status)
			}
			if uhr.zahl() != aufrufe || zahlEintraege(u) != vorher {
				t.Errorf("%s: %d Aufrufe der Uhr und %d Einträge durch die Antwort, erwartet 0 und 0", name, uhr.zahl()-aufrufe, zahlEintraege(u)-vorher)
			}
			if !voll {
				pruefeFensterVoll(t, name, u)
				continue
			}
			if zugelassen(t, u) {
				t.Errorf("%s: die 21. Abweisung ist zugelassen", name)
			}
		}
	}
	gleichAuf(t, "Zahl der Zellen", zellen, 14)

	u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	authentisierung := func() *http.Request { return ohneAuth("GET", "/ok", nil) }
	route := func() *http.Request { return anfrageMit("GET", "/e401", nil, nil) }
	for _, weg := range []func() *http.Request{authentisierung, route} {
		for range 10 {
			if rec := u.rufe(weg()); rec.Code != 401 {
				t.Fatalf("Status %d, erwartet 401", rec.Code)
			}
		}
	}
	gleichAuf(t, "Einträge aus beiden Wegen", zahlEintraege(u), 20)
	for name, weg := range map[string]func() *http.Request{"Authentisierung": authentisierung, "Route": route} {
		if rec := u.rufe(weg()); rec.Code != 401 || zahlEintraege(u) != 20 {
			t.Errorf("21. aus %s: Status %d, %d Einträge, erwartet 401 und 20", name, rec.Code, zahlEintraege(u))
		}
	}
}

// TestFensterAppendFehler hält W13 (Nr. 5 e): eine erteilte Zulassung ist verbraucht, auch wenn Append scheitert: 25
// Abweisungen ergeben 20 Aufrufe von Append, jede Antwort ist 401, und CountError läuft 25-mal.
func TestFensterAppendFehler(t *testing.T) {
	u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	u.audit.Fehler = errors.New("Ablage")
	for i := range 25 {
		pruefe401OhneKennung(t, "Abweisung "+strconv.Itoa(i+1), u.rufe(ohneAuth("GET", "/ok", nil)))
	}
	gleichAuf(t, "Aufrufe von Append", u.audit.Aufrufe(), 20)
	if codes, _ := zaehlungenVon(u.tele); len(codes) != 25 {
		t.Errorf("CountError %d-mal, erwartet 25", len(codes))
	}
}

// TestFensterJeServer hält W14 (Nr. 5 f): das Fenster gehört der Instanz: zwei Server teilen keinen Zustand, und ein neuer
// Server beginnt ohne Fenster, auch nachdem ein anderer geschlossen ist.
func TestFensterJeServer(t *testing.T) {
	a, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	pruefeFensterVoll(t, "A", a)
	b, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	pruefeFensterVoll(t, "B neben dem vollen A", b)
	if err := a.srv.Schliessen(time.Second); err != nil {
		t.Fatalf("Schliessen von A: %v", err)
	}
	c, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	pruefeFensterVoll(t, "C nach dem Schliessen von A", c)
}

// TestFensterAngehalten hält W15 (Nr. 5 g): ein angehaltener Server antwortet 503, ruft die Uhr nicht und schreibt nichts;
// eine vor Anhalten zugelassene Abweisung, die in Append hängt, schreibt ihren Eintrag nach Anhalten.
func TestFensterAngehalten(t *testing.T) {
	uhr := &stellUhr{stand: fensterT0()}
	u, _ := fensterServer(t, uhr)
	u.srv.Anhalten()
	for range 3 {
		if rec := u.rufe(ohneAuth("GET", "/ok", nil)); rec.Code != 503 || rec.Body.String() != objekt503 {
			t.Errorf("angehalten: %d %q", rec.Code, rec.Body.String())
		}
	}
	if uhr.zahl() != 0 || u.audit.Aufrufe() != 0 {
		t.Errorf("angehalten: %d Aufrufe der Uhr, %d von Audit, erwartet 0 und 0", uhr.zahl(), u.audit.Aufrufe())
	}

	u, _ = fensterServer(t, &stellUhr{stand: fensterT0()})
	begonnen, frei, fertig := make(chan struct{}), make(chan struct{}), make(chan *httptest.ResponseRecorder, 1)
	u.audit.Vor = func(support.Eintrag) {
		close(begonnen)
		<-frei
	}
	go func() { fertig <- u.rufe(ohneAuth("GET", "/ok", nil)) }()
	select {
	case <-begonnen:
	case <-time.After(5 * time.Second):
		t.Fatal("Append begann nicht in 5 s")
	}
	u.srv.Anhalten()
	close(frei)
	select {
	case rec := <-fertig:
		pruefe401OhneKennung(t, "vor Anhalten zugelassen", rec)
		gleichAuf(t, "Einträge", u.audit.Eintraege(), []support.Eintrag{eintragSoll(t, rec.Header().Get("X-Correlation-Id"))})
	case <-time.After(5 * time.Second):
		t.Fatal("der Request endete nicht in 5 s")
	}
}

// TestFensterOhneZeitgeber hält am Quelltext von abweisen_fenster.go (Nr. 1, 5 g): keine Anweisung go, keine Nennung von
// time.AfterFunc, NewTimer, NewTicker, After, Tick oder Sleep und keine der Systemuhr (time.Now, Since, Until). Die Probe
// hält den Zähler gegen eine Quelle mit je einer Stelle.
func TestFensterOhneZeitgeber(t *testing.T) {
	namen := []string{"AfterFunc", "NewTimer", "NewTicker", "After", "Tick", "Sleep", "Now", "Since", "Until"}
	probe := "package x\nimport \"time\"\nfunc f() { go f(); time.AfterFunc(); time.NewTimer(); time.NewTicker(); time.After(); " +
		"time.Tick(); time.Sleep(); time.Now(); time.Since(); time.Until() }"
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 1)
	for _, n := range namen {
		gleichAuf(t, "Probe time."+n, zaehleSelektoren(t, "probe.go", probe, "time", n), 1)
	}
	gleichAuf(t, "Zahl der Namen", len(namen), 9)
	src, err := os.ReadFile("abweisen_fenster.go")
	if err != nil {
		t.Fatal(err)
	}
	gleichAuf(t, "Anweisungen go", zaehleGoAnweisungen(t, "abweisen_fenster.go", src), 0)
	for _, n := range namen {
		if z := zaehleSelektoren(t, "abweisen_fenster.go", src, "time", n); z != 0 {
			t.Errorf("%d Nennungen von time.%s", z, n)
		}
	}
}

// TestFensterKeineNeuenArten hält W16 (Nr. 8): keine Art des Audit, kein Feld von Eintrag und kein Code kommt dazu: 12
// Arten, 7 Felder, 41 Codes.
func TestFensterKeineNeuenArten(t *testing.T) {
	var arten []string
	for _, a := range support.Arten() {
		arten = append(arten, string(a.Art))
	}
	if len(arten) != 12 || slices.ContainsFunc(arten, func(a string) bool { return strings.Contains(a, "gedrossel") }) {
		t.Errorf("Arten %v, erwartet 12 ohne eine für gedrosselte Abweisungen", arten)
	}
	gleichAuf(t, "Felder von Eintrag", reflect.TypeOf(support.Eintrag{}).NumField(), 7)
	gleichAuf(t, "Codes", len(model.AlleCodes()), 41)
}
