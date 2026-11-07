package bootstrap_test

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Determinismus (SPEC-049 Nr. 5, LH-QA-02): 100 Läufe nach Reset im selben Workspace ergeben gleiche normalisierte Antworten und gleichen Stand.

const (
	abDetLaeufe = 100
	abDetFrist  = 120 * time.Second
)

// abMelder ist, was die Auswertung der Läufe vom Test braucht (testing.TB trägt einen privaten Teil und lässt sich nicht nachbilden).
type abMelder interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
}

// abDetErgebnis ist das Ergebnis der Läufe: gelaufene Läufe, die Zahl der abweichenden, ihre Beschreibungen und ein Fehler im Lauf.
type abDetErgebnis struct {
	Laeufe       int
	Abweichend   int
	Abweichungen []string
	Fatal        string
}

// abDetMelde meldet das Ergebnis: ein Fehler im Lauf ist Fatal, jede Abweichung Error, am Ende steht die Zeile `Läufe: 100, abweichend: <k>`.
func abDetMelde(m abMelder, e abDetErgebnis) {
	for _, a := range e.Abweichungen {
		m.Errorf("%s", a)
	}
	m.Logf("Läufe: %d, abweichend: %d", e.Laeufe, e.Abweichend)
	if e.Fatal != "" {
		m.Fatalf("%s", e.Fatal)
	}
	if e.Abweichend > 0 {
		m.Errorf("Läufe: %d, abweichend: %d", e.Laeufe, e.Abweichend)
	}
}

// abFolgeD ist die Folge der vierzehn Schritte (SPEC-049 Nr. 5 b): jeder Schritt wiederholt einen Schritt der Tabelle aus Nr. 4.
func abFolgeD(t *testing.T) []abSchritt {
	t.Helper()
	var folge []abSchritt
	for _, a := range [][2]string{{"ODV4-GET-01", "a"}, {"ODV4-GET-02", "a"}, {"ODV4-Q-12", "a"}, {"ODV4-CRUD-02", "a"}, {"ODV4-CRUD-02", "b"}, {"ODV4-CRUD-04", "a"},
		{"ODV4-CRUD-04", "b"}, {"ODV4-CRUD-06", "a"}, {"ODV4-GET-02", "c"}, {"ODV4-CRUD-04", "c"}, {"ODV4-CRUD-04", "d"}, {"ODV4-H-03", "a"}, {"ODV4-E-04", "b"},
		{"ODV4-GET-03", "a"}} {
		folge = append(folge, abSchrittAus(t, a[0], a[1]))
	}
	return folge
}

// abDetCursor ist der Cursor von Nr. 5 a mit der Position pos.
func abDetCursor(pos int64) model.ReplayCursor {
	var verbraucht []int64
	for i := int64(1); i <= pos; i++ {
		verbraucht = append(verbraucht, i)
	}
	return model.ReplayCursor{WorkspaceID: "ws-1", ReplaySessionID: "rep-1", RecordingSessionID: "ses-1", ServiceID: "svc", Strategy: model.StrategySequential,
		Position: pos, Consumed: verbraucht}
}

// abStandNachReset ist der Stand nach dem Reset von Lauf i: der Ausgangszustand, der Cursor auf Position 0, die Revision 3·(i−1)+1.
func abStandNachReset(i int) abStand {
	s := abAusgangsstand("ws-1", int64(3*(i-1)+1))
	c := abDetCursor(0)
	c.Consumed = nil
	s.Cursors = []model.ReplayCursor{c}
	return s
}

// abStandNachLauf ist der Stand nach Lauf i: Products 1 mit Version 1 und geändertem Namen und Bestand, Products 3 mit Version 1 und Tombstone, der
// Cursor wie nach dem Reset, die Revision 3·i.
func abStandNachLauf(i int) abStand {
	s := abStandNachReset(i)
	s.Revision = int64(3 * i)
	for j := range s.Eintraege {
		e := &s.Eintraege[j]
		switch {
		case e.EntitySet == "Products" && e.EntityKey == "1":
			e.Version, e.Payload = 1, abPayload(1, map[string]string{"Name": `"Anvil2"`, "Stock": `"10"`})
		case e.EntitySet == "Products" && e.EntityKey == "3":
			e.Version, e.Deleted = 1, true
		}
	}
	return s
}

// abVolatileDefault ist VOLATILE_HEADERS aus der Konfiguration des Modus replay mit gültiger Umgebung (SPEC-049 Nr. 5 d), keine Kopie.
func abVolatileDefault(t *testing.T) []string {
	t.Helper()
	k, err := bootstrap.LadeKonfiguration([]string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "r")})
	if err != nil {
		t.Fatalf("Konfiguration des Modus replay: %v", err)
	}
	return k.VolatileHeaders
}

// abNormDet ist die Antwort im Vergleich der Läufe: Status, Header ohne die Namen aus volatile (ohne Beachtung der Schreibweise, Namen
// kanonisiert), Body, bei einem Fehler Status und Code.
func abNormDet(i abIst, volatile []string) string {
	if i.F != nil {
		return fmt.Sprintf("Fehler %d %s", i.F.Status, i.F.Code)
	}
	if i.Err != nil {
		return "Fehler ohne Code: " + i.Err.Error()
	}
	var namen []string
	for n := range i.Header {
		if !slices.ContainsFunc(volatile, func(v string) bool { return http.CanonicalHeaderKey(v) == http.CanonicalHeaderKey(n) }) {
			namen = append(namen, n)
		}
	}
	slices.Sort(namen)
	var sb strings.Builder
	fmt.Fprintf(&sb, "Status %d\n", i.Status)
	for _, n := range namen {
		fmt.Fprintf(&sb, "%s: %q\n", n, i.Header[n])
	}
	sb.WriteString(i.Body)
	return sb.String()
}

// abVergleicheDet nennt die Art der Abweichung zweier Antworten ("" gleich): Status, Code, Header oder Body.
func abVergleicheDet(a, b abIst, volatile []string) string {
	if abNormDet(a, volatile) == abNormDet(b, volatile) {
		return ""
	}
	switch {
	case (a.F == nil) != (b.F == nil):
		return "Fehler"
	case a.F != nil && a.F.Status != b.F.Status:
		return "Status"
	case a.F != nil:
		return "Code"
	case a.Status != b.Status:
		return "Status"
	case a.Body != b.Body:
		return "Body"
	}
	return "Header"
}

// abDetLaufen führt n Läufe nach Reset aus (SPEC-049 Nr. 5 a): die Frist von ctx muss gesetzt und höchstens abDetFrist lang sein; ein Fehler im Lauf
// ist Fatal im Ergebnis und beendet die Läufe, eine Abweichung zählt und die übrigen Läufe laufen.
func abDetLaufen(ctx context.Context, u *abUmgebung, folge []abSchritt, n int, volatile []string) abDetErgebnis {
	var e abDetErgebnis
	if dl, ok := ctx.Deadline(); !ok || time.Until(dl) > abDetFrist {
		e.Fatal = "keine Frist von höchstens 120 s für die Läufe gesetzt"
		return e
	}
	st, err := abLeseStand(ctx, u.db, "ws-1")
	if err != nil || !reflect.DeepEqual(st.Cursors, []model.ReplayCursor{abDetCursor(5)}) {
		e.Fatal = fmt.Sprintf("Kontrollwert des Lesers: Cursor vor dem ersten Reset %+v (%v), erwartet Position 5", st.Cursors, err)
		return e
	}
	var lauf1 []abIst
	var stand1 abStand
	for i := 1; i <= n; i++ {
		if ctx.Err() != nil {
			e.Fatal = fmt.Sprintf("Lauf %d: Frist abgelaufen: %v", i, ctx.Err())
			return e
		}
		w, err := u.ws.Reset(ctx, "ws-1")
		if err != nil || w.Revision != int64(3*(i-1)+1) {
			e.Fatal = fmt.Sprintf("Lauf %d: Reset ergab Revision %d (%v), erwartet %d", i, w.Revision, err, 3*(i-1)+1)
			return e
		}
		abwich := false
		melde := func(schritt, art string) {
			abwich = true
			e.Abweichungen = append(e.Abweichungen, fmt.Sprintf("Lauf %d, %s: Abweichung %s", i, schritt, art))
		}
		nachReset, err := abLeseStand(ctx, u.db, "ws-1")
		if err != nil {
			e.Fatal = fmt.Sprintf("Lauf %d: Lesefehler nach dem Reset: %v", i, err)
			return e
		}
		if d := abStandGleich(nachReset, abStandNachReset(i)); len(d) != 0 {
			melde("Stand nach Reset", "Stand: "+strings.Join(d, "; "))
		}
		ist := make([]abIst, len(folge))
		for j, s := range folge {
			ist[j] = abSende(ctx, u.mock, "ws-1", s.Anf)
			if ctx.Err() != nil {
				e.Fatal = fmt.Sprintf("Lauf %d, Schritt D%02d: Frist abgelaufen: %v", i, j+1, ctx.Err())
				return e
			}
		}
		nachLauf, err := abLeseStand(ctx, u.db, "ws-1")
		if err != nil {
			e.Fatal = fmt.Sprintf("Lauf %d: Lesefehler nach dem Lauf: %v", i, err)
			return e
		}
		if i == 1 {
			lauf1, stand1 = ist, nachLauf
			for j, s := range folge {
				if d := abVergleiche(ist[j], s.Soll); len(d) != 0 {
					melde(fmt.Sprintf("Schritt D%02d", j+1), "gegen das Orakel: "+strings.Join(d, "; "))
				}
			}
			if d := abStandGleich(nachLauf, abStandNachLauf(1)); len(d) != 0 {
				melde("Stand nach dem Lauf", "Stand gegen das Orakel: "+strings.Join(d, "; "))
			}
		} else {
			for j := range folge {
				if art := abVergleicheDet(ist[j], lauf1[j], volatile); art != "" {
					melde(fmt.Sprintf("Schritt D%02d", j+1), art)
				}
			}
			erw := stand1
			erw.Revision = int64(3 * i)
			if d := abStandGleich(nachLauf, erw); len(d) != 0 {
				melde("Stand nach dem Lauf", "Stand: "+strings.Join(d, "; "))
			}
		}
		e.Laeufe++
		if abwich {
			e.Abweichend++
		}
	}
	return e
}

// abDetUmgebung ist die Umgebung der Läufe: die Ausgangsdaten und der Cursor von Nr. 5 a auf Position 5.
func abDetUmgebung(t *testing.T) *abUmgebung {
	t.Helper()
	u := abFrisch(t)
	u.cursor("ws-1", 5)
	return u
}

// TestAbnahmeDeterminismus (T44 bis T48, T50, T53): 100 Läufe nach Reset; Lauf 1 gleich dem Orakel, jeder Lauf ab 2 gleich Lauf 1, Stand nach
// Reset und nach dem Lauf, Cursor, nur die vier Namen der Antwort im Mock.
func TestAbnahmeDeterminismus(t *testing.T) {
	start := time.Now()
	volatile := abVolatileDefault(t)
	u := abDetUmgebung(t)
	ctx, cancel := context.WithTimeout(context.Background(), abDetFrist)
	defer cancel()
	folge := abFolgeD(t)
	e := abDetLaufen(ctx, u, folge, abDetLaeufe, volatile)
	t.Logf("Dauer der %d Läufe: %v", e.Laeufe, time.Since(start))
	abDetMelde(t, e)
	if e.Laeufe != abDetLaeufe {
		t.Errorf("%d Läufe, erwartet %d", e.Laeufe, abDetLaeufe)
	}
	// Die Namen der Antworten: allein die vier des Mock, keiner flüchtig (SPEC-049 Nr. 5 d).
	erlaubt := []string{"Content-Type", "Odata-Version", "X-Recorder-Coverage", "Location"}
	if len(erlaubt) != 4 {
		t.Fatalf("Menge der Namen %d, erwartet 4", len(erlaubt))
	}
	u2 := abFrisch(t)
	for j, s := range folge {
		ist := u2.sende(s.Anf)
		for n := range ist.Header {
			if !slices.Contains(erlaubt, n) {
				t.Errorf("Schritt D%02d: Header %q außerhalb der vier Namen des Mock", j+1, n)
			}
			if slices.ContainsFunc(volatile, func(v string) bool { return http.CanonicalHeaderKey(v) == n }) {
				t.Errorf("Schritt D%02d: Header %q steht in VOLATILE_HEADERS", j+1, n)
			}
		}
	}
	for _, v := range erlaubt {
		if slices.ContainsFunc(volatile, func(x string) bool { return http.CanonicalHeaderKey(x) == http.CanonicalHeaderKey(v) }) {
			t.Errorf("%s steht in VOLATILE_HEADERS", v)
		}
	}
	// Die Folge hat vierzehn Schritte.
	if len(folge) != 14 {
		t.Errorf("Folge hat %d Schritte, erwartet 14", len(folge))
	}
}

// TestAbnahmeDeterminismusVergleich (T49): der Vergleich nimmt allein die Namen aus VOLATILE_HEADERS aus, ohne Beachtung der Schreibweise, und
// vergleicht Status, Body, Namen, Zahl und Folge der Werte.
func TestAbnahmeDeterminismusVergleich(t *testing.T) {
	volatile := abVolatileDefault(t)
	if len(volatile) == 0 {
		t.Fatal("VOLATILE_HEADERS ist leer")
	}
	basis := abIst{Status: 200, Header: map[string][]string{"Content-Type": {"a", "b"}, "Content-Length": {"5"}}, Body: "x"}
	mit := func(f func(i *abIst)) abIst {
		c := basis
		c.Header = map[string][]string{}
		for k, v := range basis.Header {
			c.Header[k] = append([]string(nil), v...)
		}
		f(&c)
		return c
	}
	for _, v := range volatile {
		for _, schreibweise := range []string{http.CanonicalHeaderKey(v), strings.ToLower(v), strings.ToUpper(v)} {
			a := mit(func(i *abIst) { i.Header[schreibweise] = []string{"x"} })
			// Der Name der Antwort ist kanonisch (abSende); eine andere Schreibweise im Rohwert zählt trotzdem als dieselbe Ausnahme.
			if art := abVergleicheDet(a, basis, volatile); art != "" {
				t.Errorf("flüchtiger Header %q (%q) ändert den Vergleich: %s", v, schreibweise, art)
			}
		}
	}
	for name, abw := range map[string]abIst{
		"Content-Length ist nicht flüchtig": mit(func(i *abIst) { i.Header["Content-Length"] = []string{"6"} }),
		"Status":                            mit(func(i *abIst) { i.Status = 201 }),
		"Body":                              mit(func(i *abIst) { i.Body = "y" }),
		"weiterer Name":                     mit(func(i *abIst) { i.Header["X-Neu"] = []string{"1"} }),
		"Name fehlt":                        mit(func(i *abIst) { delete(i.Header, "Content-Length") }),
		"Zahl der Werte":                    mit(func(i *abIst) { i.Header["Content-Type"] = []string{"a"} }),
		"Folge der Werte":                   mit(func(i *abIst) { i.Header["Content-Type"] = []string{"b", "a"} }),
	} {
		if abVergleicheDet(abw, basis, volatile) == "" {
			t.Errorf("%s: kein Unterschied gemeldet", name)
		}
	}
	if abVergleicheDet(basis, basis, volatile) != "" {
		t.Error("gleiche Antworten ungleich")
	}
	if abVergleicheDet(abFehlerIst(404, model.CodeE6008), abFehlerIst(404, model.CodeE6003), volatile) == "" ||
		abVergleicheDet(abFehlerIst(404, model.CodeE6008), abFehlerIst(400, model.CodeE6008), volatile) == "" {
		t.Error("Fehler mit anderem Code oder Status gleich")
	}
}

// abTestMelder zeichnet die Meldungen auf.
type abTestMelder struct{ fehler, fatal, log []string }

func (m *abTestMelder) Errorf(f string, a ...any) { m.fehler = append(m.fehler, fmt.Sprintf(f, a...)) }
func (m *abTestMelder) Fatalf(f string, a ...any) { m.fatal = append(m.fatal, fmt.Sprintf(f, a...)) }
func (m *abTestMelder) Logf(f string, a ...any)   { m.log = append(m.log, fmt.Sprintf(f, a...)) }

// TestAbnahmeDeterminismusAuswertung (T51): eine Abweichung macht den Test rot, mit der Zeile `Läufe: 100, abweichend: <k>`; ohne Abweichung bleibt er grün.
func TestAbnahmeDeterminismusAuswertung(t *testing.T) {
	gruen := &abTestMelder{}
	abDetMelde(gruen, abDetErgebnis{Laeufe: 100})
	if len(gruen.fehler) != 0 || len(gruen.fatal) != 0 || !slices.Equal(gruen.log, []string{"Läufe: 100, abweichend: 0"}) {
		t.Errorf("ohne Abweichung: %+v", gruen)
	}
	rot := &abTestMelder{}
	abDetMelde(rot, abDetErgebnis{Laeufe: 100, Abweichend: 3, Abweichungen: []string{"Lauf 7, Schritt D05: Body"}})
	if len(rot.fehler) == 0 || !slices.Contains(rot.fehler, "Lauf 7, Schritt D05: Body") || !slices.Contains(rot.log, "Läufe: 100, abweichend: 3") || len(rot.fatal) != 0 {
		t.Errorf("mit Abweichung: %+v", rot)
	}
	fatal := &abTestMelder{}
	abDetMelde(fatal, abDetErgebnis{Laeufe: 4, Fatal: "Lauf 5: Reset"})
	if len(fatal.fatal) != 1 {
		t.Errorf("Fehler im Lauf ist nicht Fatal: %+v", fatal)
	}
	// Eine Abweichung eines wirklichen Laufs: ein Orakel, das der Mock nicht erfüllt, ergibt Abweichungen und die Zeile.
	u := abDetUmgebung(t)
	folge := abFolgeD(t)
	folge[4].Soll.Body += "x"
	ctx, cancel := context.WithTimeout(context.Background(), abDetFrist)
	defer cancel()
	e := abDetLaufen(ctx, u, folge, 3, abVolatileDefault(t))
	if e.Fatal != "" || e.Laeufe != 3 || e.Abweichend != 1 {
		t.Errorf("Orakel verletzt: %+v, erwartet drei Läufe und genau Lauf 1 abweichend", e)
	}
}

// TestAbnahmeDeterminismusFrist (T52): ohne Frist, mit zu langer, abgelaufener und mit Frist im Lauf abgelaufener ist das Ergebnis Fatal.
func TestAbnahmeDeterminismusFrist(t *testing.T) {
	if abDetFrist != 120*time.Second {
		t.Fatalf("Frist %v, erwartet 120 s", abDetFrist)
	}
	u := abDetUmgebung(t)
	folge := abFolgeD(t)
	volatile := abVolatileDefault(t)
	if e := abDetLaufen(context.Background(), u, folge, 2, volatile); e.Fatal == "" || e.Laeufe != 0 {
		t.Errorf("ohne Frist: %+v", e)
	}
	lang, c1 := context.WithTimeout(context.Background(), abDetFrist+time.Minute)
	defer c1()
	if e := abDetLaufen(lang, u, folge, 2, volatile); e.Fatal == "" || e.Laeufe != 0 {
		t.Errorf("Frist über 120 s: %+v", e)
	}
	abgelaufen, c2 := context.WithTimeout(context.Background(), time.Nanosecond)
	defer c2()
	<-abgelaufen.Done()
	if e := abDetLaufen(abgelaufen, u, folge, 2, volatile); e.Fatal == "" || e.Laeufe != 0 {
		t.Errorf("abgelaufene Frist: %+v", e)
	}
	m := &abTestMelder{}
	abDetMelde(m, abDetLaufen(abgelaufen, u, folge, 2, volatile))
	if len(m.fatal) != 1 {
		t.Errorf("abgelaufene Frist ist nicht Fatal: %+v", m)
	}
}
