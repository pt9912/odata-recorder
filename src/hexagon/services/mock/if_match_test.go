package mock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Die Tests dieser Datei halten die Zusagen von SPEC-283 (Tabelle Zusage und Fänger des Plans, F1 bis F26). Die Literale der Header und
// der Codes stammen aus dem Vertrag, nie aus dem Code des Dienstes.

// anfrageMitBody ist der Request einer Mutation an den Eintrag id mit dem Body e und den Kopfzeilen kopf (die Wertlisten werden
// kopiert, eine nil-Liste bleibt nil, eine leere bleibt leer).
func (u *umg) anfrageMitBody(methode, id string, e model.EntityEingabe, kopf map[string][]string) inbound.Anfrage {
	u.setzeMutation(methode, id, e)
	a := anfrage(methode, dienstRoot+"/x")
	a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
	a.Header["Content-Type"] = []string{"application/json"}
	for n, w := range kopf {
		a.Header[n] = slices.Clone(w)
	}
	return a
}

// mutiereMit ruft die Mutation mit dem Standard-Body und den Kopfzeilen kopf.
func (u *umg) mutiereMit(methode, id string, kopf map[string][]string) (inbound.Antwort, error) {
	u.t.Helper()
	return u.rufe(u.anfrageMitBody(methode, id, eingabeIDName(id, "x"), kopf))
}

// zeile ist der Wert einer Zeile mit dem Tag der Version v: schwach oder stark.
func zeile(v int64, schwach bool) string {
	if schwach {
		return "W/" + `"` + strconv.FormatInt(v, 10) + `"`
	}
	return `"` + strconv.FormatInt(v, 10) + `"`
}

// kopfName sind die zwei Header der Vorbedingung; erfuellt und verletzt liefern je Header einen Wert der Version v, der die Bedingung
// erfüllt beziehungsweise nicht (If-Match: Treffer; If-None-Match: Nichttreffer).
type kopfName string

const (
	ifMatch     kopfName = "If-Match"
	ifNoneMatch kopfName = "If-None-Match"
)

func (k kopfName) erfuellt(v int64) []string {
	if k == ifMatch {
		return []string{zeile(v, true)}
	}
	return []string{zeile(v+1, true)}
}

func (k kopfName) verletzt(v int64) []string {
	if k == ifMatch {
		return []string{zeile(v+1, true)}
	}
	return []string{zeile(v, true)}
}

func (k kopfName) mit(w []string) map[string][]string { return map[string][]string{string(k): w} }

func beideKoepfe() []kopfName { return []kopfName{ifMatch, ifNoneMatch} }

// koerper liest den Body einer Antwort bis zum Ende; ein nil-Body ist leer.
func koerper(t *testing.T, a inbound.Antwort) []byte {
	t.Helper()
	if a.Body == nil {
		return nil
	}
	return lies(t, a)
}

// bedingteMethoden sind die Methoden, die die Header auswerten (SPEC-283 Nr. 1).
func bedingteMethoden() []string { return []string{"PATCH", "PUT", "DELETE"} }

// abgewiesen prüft die Aufrufe nach einer abgewiesenen Vorbedingung: Transaktion und Entities.Get je einmal, Workspaces.Get zweimal (einmal
// in Schritt (f) vor der Transaktion, einmal in f; SPEC-192 Nr. 8 c), nie eine Operation, die schreibt, die Coverage liest oder den
// Schreiber ruft.
func (u *umg) abgewiesen(kontext string) {
	u.t.Helper()
	for _, op := range u.p.ops() {
		for _, p := range []string{"Entities.Put", "Workspaces.Advance", "Coverage.", "Schreiber.", "Cursors.", "Snapshots.", "Entities.List", "Entities.ReplaceAll"} {
			if strings.HasPrefix(op, p) {
				u.t.Errorf("%s: %s gerufen", kontext, op)
			}
		}
	}
	for op, want := range map[string]int{"Transaktion.WithWorkspaceTransaction": 1, "Workspaces.Get": 2, "Entities.Get": 1} {
		if got := u.p.anzahl(op); got != want {
			u.t.Errorf("%s: %s %d Mal, erwartet %d", kontext, op, got, want)
		}
	}
}

// erwarte412 prüft den Fehler: ODR-E6010 mit Status 412, fester Text, keine Ursache und kein Wert der Header im Text.
func erwarte412(t *testing.T, kontext string, err error) {
	t.Helper()
	if f := erwarteFehler(t, kontext, err, model.CodeE6010); f != nil {
		if f.Ursache != nil || f.Status != 412 || strings.Contains(f.Error(), kanarienwert) {
			t.Errorf("%s: Ursache %v Status %d Text %q", kontext, f.Ursache, f.Status, f.Error())
		}
	}
}

// TestIfMatchPasst (F1): ein erfüllter Header lässt PATCH, PUT und DELETE zu (204); PATCH und PUT tragen den ETag der Version v + 1,
// DELETE keinen; Put und Advance je 1, die Revision steigt um 1. Methode 3 × Version 10 (0 bis 100 und die großen Versionen 2^32, 2^32+1, 2^62, 2^63-2; v + 1 muss darstellbar bleiben) × Form 2 × Header 2 = 120 Läufe.
func TestIfMatchPasst(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		for _, v := range append([]int64{0, 1, 9, 10, 99, 100}, versionenGross()[:4]...) {
			for _, schwach := range []bool{true, false} {
				for _, k := range beideKoepfe() {
					n++
					kontext := fmt.Sprintf("%s %s Version %d schwach=%v", m, k, v, schwach)
					u := etagUmg(t, true, v, v, v)
					wert := []string{zeile(v, schwach)}
					if k == ifNoneMatch {
						wert = []string{zeile(v+1, schwach)}
					}
					a, err := u.mutiereMit(m, "2", k.mit(wert))
					erwarteStatus(t, kontext, a, err, 204)
					if w, _ := etagWerte(a); m != "DELETE" && (len(w) != 1 || w[0] != etagLit(v+1)) || m == "DELETE" && len(w) != 0 {
						t.Errorf("%s: ETag %v, erwartet %s", kontext, w, etagLit(v+1))
					}
					got := u.zustand()["2"]
					if got.Version != v+1 || got.Deleted != (m == "DELETE") {
						t.Errorf("%s: Version %d Deleted %v", kontext, got.Version, got.Deleted)
					}
					for op, want := range map[string]int{"Entities.Put": 1, "Workspaces.Advance": 1, "Transaktion.WithWorkspaceTransaction": 1} {
						if u.p.anzahl(op) != want {
							t.Errorf("%s: %s %d Mal, erwartet %d", kontext, op, u.p.anzahl(op), want)
						}
					}
					if u.ws.revision(wsID) != 1 {
						t.Errorf("%s: Revision %d, erwartet 1", kontext, u.ws.revision(wsID))
					}
				}
			}
		}
	}
	if n != 3*10*2*2 {
		t.Errorf("Menge %d, erwartet %d", n, 3*10*2*2)
	}
}

// versionenGross sind die großen Versionen aus dem Wortlaut der Zusage (Versionen 0 bis 2^63-1): 2^32, 2^32+1, 2^62, 2^63-2, 2^63-1; die Fakes
// beginnen mit ihnen. Der Dienst reicht den Zähler als int64 weiter, ohne Kürzung der Breite, des Vorzeichens oder der Genauigkeit.
func versionenGross() []int64 {
	return []int64{1 << 32, 1<<32 + 1, 1 << 62, 1<<63 - 2, 1<<63 - 1}
}

// TestIfMatchVeraltet (F2, F21): ein Tag, der die Bedingung bricht (v - 1, v + 1, 0, "x", W/"x"; für If-None-Match der eigene Tag), ist
// ODR-E6010: Zustand, Revision und Zähler bleiben, nur Transaktion, Workspaces.Get und Entities.Get je einmal. Der Wert trägt eine
// Marke, die in keinem Text stehen darf. Methode 3 × Version 9 (0, 1, 7, 10 und die großen Versionen) × Tag je Header 5 × Header 2 = 270 Läufe.
func TestIfMatchVeraltet(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		for _, v := range append([]int64{0, 1, 7, 10}, versionenGross()...) {
			for _, k := range beideKoepfe() {
				var tags []string
				if k == ifMatch {
					tags = []string{zeile(v-1, true), zeile(v+1, true), zeile(v+2, false), `"` + kanarienwert + `"`, `W/"x"`}
				} else {
					tags = []string{zeile(v, true), zeile(v, false), `*`, zeile(v, true) + `,"` + kanarienwert + `"`, `"` + kanarienwert + `",` + zeile(v, true)}
				}
				for _, tag := range tags {
					n++
					kontext := fmt.Sprintf("%s %s Version %d %s", m, k, v, tag)
					u := etagUmg(t, true, v, v, v)
					vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
					a, err := u.mutiereMit(m, "2", k.mit([]string{tag}))
					erwarte412(t, kontext, err)
					if a.Status != 0 || a.Header != nil || a.Body != nil {
						t.Errorf("%s: Antwort %+v bei einem Fehler", kontext, a)
					}
					u.abgewiesen(kontext)
					u.unveraendert(vorher, rev, kontext)
				}
			}
		}
	}
	if n != 3*9*2*5 {
		t.Errorf("Menge %d, erwartet %d", n, 3*9*2*5)
	}
}

// partiell ersetzt den Eintrag 2 durch einen partiellen der Version v.
func partiell(u *umg, v int64) {
	e := teilEintrag("2", v)
	u.en.setze(setProducts, voll(1, "a"), e, voll(3, "c"))
}

// TestIfMatchStern (F3): If-Match * lässt PATCH, PUT und DELETE an einem vorhandenen Eintrag zu, ohne Versionsvergleich, für jede Version;
// PATCH an einem partiellen Eintrag ist ODR-E6003 (F14). Methode 3 × Eintrag 2 × Version 4 = 24 Läufe.
func TestIfMatchStern(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		for _, art := range []string{"vollständig", "partiell"} {
			for _, v := range []int64{0, 1, 7, 100} {
				n++
				kontext := fmt.Sprintf("%s %s Version %d", m, art, v)
				u := etagUmg(t, true, v, v, v)
				if art == "partiell" {
					partiell(u, v)
				}
				a, err := u.mutiereMit(m, "2", ifMatch.mit([]string{"*"}))
				if m == "PATCH" && art == "partiell" {
					pruefeFehler(t, kontext, err, e6003)
					continue
				}
				erwarteStatus(t, kontext, a, err, 204)
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// TestIfMatchRangZustand (F4): ein Tombstone und ein fehlender Eintrag (Coverage complete, partial, unknown) bleiben ODR-E6008 und
// ODR-E6003, nie ODR-E6010, mit jedem Wert beider Header; Put 0. Zustand 4 × Methode 3 × Header 2 × Wert 4 = 96 Läufe.
func TestIfMatchRangZustand(t *testing.T) {
	zustaende := []struct {
		name string
		bau  func(u *umg)
		id   string
		code model.Code
	}{
		{"Tombstone", func(u *umg) {
			e := voll(2, "b")
			e.Version, e.Deleted = 4, true
			u.en.setze(setProducts, voll(1, "a"), e, voll(3, "c"))
		}, "2", e6008},
		{"fehlend complete", func(u *umg) { u.co.setze(baselineID, setProducts, model.CoverageComplete) }, "99", e6008},
		{"fehlend partial", func(u *umg) { u.co.setze(baselineID, setProducts, model.CoveragePartial) }, "99", e6003},
		{"fehlend unknown", func(u *umg) { u.co.setze(baselineID, setProducts, model.CoverageUnknown) }, "99", e6003},
	}
	n := 0
	for _, z := range zustaende {
		for _, m := range bedingteMethoden() {
			for _, k := range beideKoepfe() {
				for wi, w := range [][]string{k.erfuellt(4), k.verletzt(4), {"*"}, {"x"}} {
					n++
					kontext := fmt.Sprintf("%s %s %s Wert %d", z.name, m, k, wi)
					u := etagUmg(t, true, 4, 4, 4)
					z.bau(u)
					_, err := u.mutiereMit(m, z.id, k.mit(w))
					pruefeFehler(t, kontext, err, z.code)
					u.ohneSchreiben(kontext)
				}
			}
		}
	}
	if n != 96 {
		t.Errorf("Menge %d, erwartet 96", n)
	}
}

// TestIfMatchListePosition (F5, F11): Länge 1 bis 4 der Liste × Position des treffenden Tags (oder keine) × Methode × Header: If-Match gelingt
// mit einem Treffer an jeder Position, If-None-Match mit keinem; sonst ODR-E6010. 14 Formen × 3 × 2 = 84 Läufe.
func TestIfMatchListePosition(t *testing.T) {
	n := 0
	for laenge := 1; laenge <= 4; laenge++ {
		for pos := 0; pos <= laenge; pos++ {
			var elemente []string
			for i := 0; i < laenge; i++ {
				elemente = append(elemente, zeile(int64(20+i), i%2 == 0))
			}
			if pos < laenge {
				elemente[pos] = zeile(5, pos%2 == 0)
			}
			for _, m := range bedingteMethoden() {
				for _, k := range beideKoepfe() {
					n++
					kontext := fmt.Sprintf("%s %s Liste %v", m, k, elemente)
					u := etagUmg(t, true, 5, 5, 5)
					a, err := u.mutiereMit(m, "2", k.mit([]string{strings.Join(elemente, ",")}))
					if treffer := pos < laenge; treffer == (k == ifMatch) {
						erwarteStatus(t, kontext, a, err, 204)
					} else {
						erwarte412(t, kontext, err)
					}
				}
			}
		}
	}
	if n != 84 {
		t.Errorf("Menge %d, erwartet 84", n)
	}
}

// TestIfMatchZeilen (F9): mehrere Zeilen sind eine Liste, der Name zählt ohne Schreibweise, mehrere Schreibweisen im selben Request werden
// gemeinsam gelesen; eine leere Zeile neben einer gültigen zählt nicht, nur leere Zeilen und ein Stern neben einem Tag sind ODR-E6010.
func TestIfMatchZeilen(t *testing.T) {
	schreibweisen := func(k kopfName) []string {
		if k == ifMatch {
			return []string{"If-Match", "if-match", "IF-MATCH", "If-match"}
		}
		return []string{"If-None-Match", "if-none-match", "IF-NONE-MATCH", "If-none-match"}
	}
	treffer := func(k kopfName, v int64) string { // der Tag, den der Header als erfüllt (If-Match) beziehungsweise verletzt (If-None-Match) liest
		if k == ifMatch {
			return zeile(v, true)
		}
		return zeile(v, false)
	}
	n := 0
	gelingt := func(k kopfName, hit bool) bool { return hit == (k == ifMatch) }
	// Eine Schreibweise × Zeile des Treffers 1 bis 3 (die Zeilen sind verschieden) × Header.
	for _, k := range beideKoepfe() {
		for _, name := range schreibweisen(k) {
			for hitZeile := 0; hitZeile < 3; hitZeile++ {
				for _, hit := range []bool{true, false} {
					n++
					zeilen := []string{`"x"`, `W/"y"`, `"z"`}
					if hit {
						zeilen[hitZeile] = treffer(k, 5)
					}
					u := etagUmg(t, true, 5, 5, 5)
					kontext := fmt.Sprintf("%s Zeile %d hit=%v", name, hitZeile, hit)
					a, err := u.mutiereMit("PATCH", "2", map[string][]string{name: zeilen})
					if gelingt(k, hit) {
						erwarteStatus(t, kontext, a, err, 204)
					} else {
						erwarte412(t, kontext, err)
					}
				}
			}
		}
	}
	// Zwei Schreibweisen im selben Request: jedes geordnete Paar × Träger des Treffers × Header.
	for _, k := range beideKoepfe() {
		namen := schreibweisen(k)
		for i, a1 := range namen {
			for j, a2 := range namen {
				if i == j {
					continue
				}
				for _, traeger := range []string{a1, a2} {
					n++
					kopf := map[string][]string{a1: {`"x"`}, a2: {`"y"`}}
					kopf[traeger] = []string{treffer(k, 5)}
					u := etagUmg(t, true, 5, 5, 5)
					kontext := fmt.Sprintf("%s und %s, Treffer in %s", a1, a2, traeger)
					a, err := u.mutiereMit("PATCH", "2", kopf)
					if gelingt(k, true) {
						erwarteStatus(t, kontext, a, err, 204)
					} else {
						erwarte412(t, kontext, err)
					}
				}
			}
		}
	}
	// Leere Zeilen, Stern neben einem Tag in einer anderen Zeile.
	for _, k := range beideKoepfe() {
		for _, name := range schreibweisen(k) {
			n += 4
			u := etagUmg(t, true, 5, 5, 5)
			_, err := u.mutiereMit("PATCH", "2", map[string][]string{name: {``}})
			erwarte412(t, name+" eine leere Zeile", err)
			u = etagUmg(t, true, 5, 5, 5)
			_, err = u.mutiereMit("PATCH", "2", map[string][]string{name: {``, ` `, `,`}})
			erwarte412(t, name+" nur leere Zeilen", err)
			u = etagUmg(t, true, 5, 5, 5)
			_, err = u.mutiereMit("PATCH", "2", map[string][]string{name: {`*`, `"x"`}})
			erwarte412(t, name+" Stern und Tag in zwei Zeilen", err)
			u = etagUmg(t, true, 5, 5, 5)
			a, err := u.mutiereMit("PATCH", "2", map[string][]string{name: {``, k.erfuellt(5)[0]}})
			erwarteStatus(t, name+" leere Zeile neben Tag", a, err, 204)
		}
	}
	if n != 2*4*3*2+2*12*2+2*4*4 {
		t.Errorf("Menge %d", n)
	}
}

// TestIfMatchNichtGesetzt (F10): kein Schlüssel, ein nil-Wert und eine leere Wertliste wirken wie ohne Header; eine leere Zeile ist gesetzt und
// ODR-E6010. Zustand 4 × Header 2 × Methode 3 = 24 Läufe.
func TestIfMatchNichtGesetzt(t *testing.T) {
	zustaende := []struct {
		name string
		kopf func(k kopfName) map[string][]string
		ok   bool
	}{
		{"kein Schlüssel", func(kopfName) map[string][]string { return map[string][]string{} }, true},
		{"nil", func(k kopfName) map[string][]string { return map[string][]string{string(k): nil} }, true},
		{"leere Wertliste", func(k kopfName) map[string][]string { return map[string][]string{string(k): {}} }, true},
		{"eine leere Zeile", func(k kopfName) map[string][]string { return map[string][]string{string(k): {""}} }, false},
	}
	n := 0
	for _, z := range zustaende {
		for _, k := range beideKoepfe() {
			for _, m := range bedingteMethoden() {
				n++
				kontext := fmt.Sprintf("%s %s %s", z.name, k, m)
				u := etagUmg(t, true, 5, 5, 5)
				a, err := u.mutiereMit(m, "2", z.kopf(k))
				if z.ok {
					erwarteStatus(t, kontext, a, err, 204)
				} else {
					erwarte412(t, kontext, err)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// TestIfMatchBeide (F12): mit beiden Headern gelingt die Mutation nur, wenn beide erfüllt sind. Methode 3 × If-Match 2 × If-None-Match 2 = 12 Läufe.
func TestIfMatchBeide(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		for _, imOK := range []bool{true, false} {
			for _, inmOK := range []bool{true, false} {
				n++
				kopf := map[string][]string{"If-Match": ifMatch.verletzt(5), "If-None-Match": ifNoneMatch.verletzt(5)}
				if imOK {
					kopf["If-Match"] = ifMatch.erfuellt(5)
				}
				if inmOK {
					kopf["If-None-Match"] = ifNoneMatch.erfuellt(5)
				}
				u := etagUmg(t, true, 5, 5, 5)
				kontext := fmt.Sprintf("%s IM=%v INM=%v", m, imOK, inmOK)
				a, err := u.mutiereMit(m, "2", kopf)
				if imOK && inmOK {
					erwarteStatus(t, kontext, a, err, 204)
				} else {
					erwarte412(t, kontext, err)
				}
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12", n)
	}
}

// verstoss ist ein Verstoß vor der Transaktion (SPEC-192 Nr. 2 (e) bis (j)): wie er den Aufbau und den Request verletzt, welcher Code folgt und
// für welche Methoden er gilt.
type verstoss struct {
	name    string
	code    model.Code
	bau     func(u *umg, a *inbound.Anfrage)
	methode func(m string) bool
}

func vorTransaktion() []verstoss {
	alle := func(string) bool { return true }
	mitBody := func(m string) bool { return m != "DELETE" }
	return []verstoss{
		{"Ressource Collection", e6002, func(u *umg, _ *inbound.Anfrage) { u.menge(model.Optionen{}) }, alle},
		{"$filter", e6002, func(u *umg, _ *inbound.Anfrage) {
			u.entity(model.Optionen{Filter: idGroesser("1")}, "2")
		}, alle},
		{"nur lesender Satz (V2)", e6006, func(u *umg, _ *inbound.Anfrage) { u.k.Dienste[0].Codecs = u.c.satz(true, "4.0") }, alle},
		{"kein Leser der Bodies", e6002, func(u *umg, _ *inbound.Anfrage) {
			s := u.c.satz(false, "4.0")
			s.Eingabe = nil
			u.k.Dienste[0].Codecs = s
		}, mitBody},
		{"Workspace unbekannt", e7001, func(_ *umg, a *inbound.Anfrage) { a.Header["X-Recorder-Workspace"] = []string{"unbekannt"} }, alle},
		{"Workspace fremd", model.CodeE5004, func(u *umg, _ *inbound.Anfrage) {
			u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer", BaselineSnapshotID: baselineID})
		}, alle},
		{"Service ohne Metadaten", model.CodeE6004, func(u *umg, _ *inbound.Anfrage) { u.md.setze(model.Service{ID: dienstID}) }, alle},
		{"Set unbekannt", e6008, func(u *umg, _ *inbound.Anfrage) {
			u.ressource(model.RessourceEntity, "Unbekannt", model.Optionen{}, zahl("", "2"))
		}, alle},
		{"Schlüssel der URL", e6005, func(u *umg, _ *inbound.Anfrage) {
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{}, zahl("Falsch", "2"))
		}, alle},
		{"Body ungültig", e6007, func(u *umg, _ *inbound.Anfrage) {
			u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeSyntax})
		}, mitBody},
		{"Body nicht unterstützt", e6002, func(u *umg, _ *inbound.Anfrage) {
			u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeNavigation})
		}, mitBody},
		{"Body Programmfehler", e9001, func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("Eingabe.LeseEntity", model.ErrUngueltig) }, mitBody},
		{"Body reservierter Text", e6007, func(u *umg, _ *inbound.Anfrage) {
			u.ei.ergebnis.Properties = append(u.ei.ergebnis.Properties, pw("Name", model.ArtString, model.Marker))
		}, mitBody},
		{"Body widerspricht dem Schlüssel", e6007, func(u *umg, _ *inbound.Anfrage) { u.ei.ergebnis = eingabeIDName("3", "x") }, mitBody},
	}
}

// TestIfMatchVorTransaktion (F13): jeder Verstoß vor der Transaktion geht der Vorbedingung vor, auch wenn der Header nicht erfüllt ist; der
// Fake zählt Transaktion 0. Ohne den Verstoß ist dieselbe Anfrage ODR-E6010. Verstoß 14 × Methode 3 (je nach Gültigkeit 8 + 14 + 14) × Header 2 = 72 Läufe.
func TestIfMatchVorTransaktion(t *testing.T) {
	n := 0
	for _, k := range beideKoepfe() {
		for _, m := range bedingteMethoden() {
			u := etagUmg(t, true, 5, 5, 5)
			_, err := u.mutiereMit(m, "2", k.mit(k.verletzt(5)))
			erwarte412(t, fmt.Sprintf("%s %s ohne Verstoß", m, k), err)
			for _, v := range vorTransaktion() {
				if !v.methode(m) {
					continue
				}
				n++
				kontext := fmt.Sprintf("%s %s %s", m, k, v.name)
				u := etagUmg(t, true, 5, 5, 5)
				a := u.anfrageMitBody(m, "2", eingabeIDName("2", "x"), k.mit(k.verletzt(5)))
				v.bau(u, &a)
				_, err := u.rufe(a)
				pruefeFehler(t, kontext, err, v.code)
				if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
					t.Errorf("%s: Transaktion gerufen", kontext)
				}
			}
		}
	}
	// Von den 14 Verstößen gelten 8 für DELETE (der Body wird nie gelesen), 14 für PATCH und für PUT.
	if want := 2 * (8 + 14 + 14); n != want {
		t.Errorf("Menge %d, erwartet %d", n, want)
	}
}

// TestIfMatchInTransaktion (F14): in der Transaktion gehen ODR-E7001 (Workspace nach Schritt (f) gelöscht), ODR-E6003 (PATCH an einem partiellen
// Eintrag) und ODR-E6007 (Typ) der Vorbedingung vor; PUT und DELETE an einem partiellen Eintrag werden verglichen (204 oder ODR-E6010).
func TestIfMatchInTransaktion(t *testing.T) {
	for _, k := range beideKoepfe() {
		for _, m := range bedingteMethoden() {
			u := etagUmg(t, true, 5, 5, 5)
			u.ws.nach = func(n int) {
				if n == 1 {
					u.ws.entferne(wsID)
				}
			}
			_, err := u.mutiereMit(m, "2", k.mit(k.verletzt(5)))
			pruefeFehler(t, fmt.Sprintf("%s %s Workspace gelöscht", m, k), err, e7001)
		}
		u := etagUmg(t, true, 5, 5, 5)
		partiell(u, 5)
		_, err := u.mutiereMit("PATCH", "2", k.mit(k.verletzt(5)))
		pruefeFehler(t, fmt.Sprintf("PATCH %s partiell", k), err, e6003)
		for _, m := range []string{"PATCH", "PUT"} {
			u = etagUmg(t, true, 5, 5, 5)
			u.setzeMutation(m, "2", model.EntityEingabe{Typ: "N.Anderer", Schluessel: "2", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "2")}})
			a := anfrage(m, dienstRoot+"/x")
			a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
			a.Header["Content-Type"] = []string{"application/json"}
			a.Header[string(k)] = slices.Clone(k.verletzt(5))
			_, err = u.rufe(a)
			pruefeFehler(t, fmt.Sprintf("%s %s Typ", m, k), err, e6007)
		}
		for _, m := range []string{"PUT", "DELETE"} {
			u = etagUmg(t, true, 5, 5, 5)
			partiell(u, 5)
			a, err := u.mutiereMit(m, "2", k.mit(k.erfuellt(5)))
			erwarteStatus(t, fmt.Sprintf("%s %s partiell erfüllt", m, k), a, err, 204)
			u = etagUmg(t, true, 5, 5, 5)
			partiell(u, 5)
			_, err = u.mutiereMit(m, "2", k.mit(k.verletzt(5)))
			erwarte412(t, fmt.Sprintf("%s %s partiell nicht erfüllt", m, k), err)
		}
	}
}

// TestIfMatchNeuanlage (F19): nach DELETE und POST an derselben Adresse treffen die ETags vor dem DELETE nie, der ETag der Neuanlage trifft.
// Version 3 × Methode (PATCH, PUT) × Header 2.
func TestIfMatchNeuanlage(t *testing.T) {
	for _, v := range []int64{1, 3, 9} {
		for _, m := range []string{"PATCH", "PUT"} {
			u := etagUmg(t, true, v, v, v)
			a, err := u.mutiereMit("DELETE", "2", ifMatch.mit([]string{zeile(v, true)}))
			erwarteStatus(t, "DELETE", a, err, 204)
			a, err = u.mutiere("POST", "2", eingabeIDName("2", "neu"))
			erwarteStatus(t, "POST", a, err, 201)
			for _, alt := range []int64{v, v + 1} {
				_, err = u.mutiereMit(m, "2", ifMatch.mit([]string{zeile(alt, true)}))
				erwarte412(t, fmt.Sprintf("%s mit dem ETag %d nach Neuanlage", m, alt), err)
			}
			a, err = u.mutiereMit(m, "2", ifMatch.mit([]string{zeile(v+2, true)}))
			erwarteStatus(t, m+" mit dem ETag der Neuanlage", a, err, 204)
		}
	}
}

// TestIfMatchWiederholung (F20): dieselbe Anfrage ein zweites Mal ist ODR-E6010, auch nach PATCH mit {}; die abgewiesene ändert nichts.
// Methode 2 × Body 3.
func TestIfMatchWiederholung(t *testing.T) {
	for _, m := range []string{"PATCH", "PUT"} {
		leer := model.EntityEingabe{Typ: typProduct, Schluessel: "2"} // PATCH mit {}: keine Property; PUT trägt allein den Schlüssel
		if m == "PUT" {
			leer.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "2")}
		}
		for i, e := range []model.EntityEingabe{eingabeIDName("2", "neu"), leer, eingabeIDName("2", "b")} {
			u := etagUmg(t, true, 4, 4, 4)
			kontext := fmt.Sprintf("%s Body %d", m, i)
			kopf := ifMatch.mit([]string{zeile(4, true)})
			a, err := u.rufe(u.anfrageMitBody(m, "2", e, kopf))
			erwarteStatus(t, kontext+" erste", a, err, 204)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			for wdh := 0; wdh < 2; wdh++ {
				_, err = u.rufe(u.anfrageMitBody(m, "2", e, kopf))
				erwarte412(t, kontext+" Wiederholung", err)
				u.unveraendert(vorher, rev, kontext)
			}
			if got := u.zustand()["2"].Version; got != 5 {
				t.Errorf("%s: Version %d, erwartet 5", kontext, got)
			}
		}
	}
}

// TestIfMatchAus (F24): bei ausgeschaltetem Etag haben beide Header keine Wirkung: Antwort und Zustand wie ohne Header. Methode 3 × Header 2 ×
// Wert 5 (treffend, veraltet, *, unlesbar, leer) = 30 Läufe.
func TestIfMatchAus(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		ohne := etagUmg(t, false, 4, 6, 9)
		aOhne, err := ohne.mutiereMit(m, "2", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range beideKoepfe() {
			for _, w := range []string{etagLit(6), etagLit(99), "*", "x", ""} {
				n++
				kontext := fmt.Sprintf("%s %s %q", m, k, w)
				u := etagUmg(t, false, 4, 6, 9)
				a, err := u.mutiereMit(m, "2", k.mit([]string{w}))
				erwarteStatus(t, kontext, a, err, aOhne.Status)
				if fmt.Sprint(a.Header) != fmt.Sprint(aOhne.Header) {
					t.Errorf("%s: Header %v, ohne Header %v", kontext, a.Header, aOhne.Header)
				}
				if !slices.EqualFunc(u.zustandsliste(), ohne.zustandsliste(), func(a, b model.EntityState) bool {
					return a.EntityKey == b.EntityKey && a.Version == b.Version && a.Deleted == b.Deleted && a.Payload == b.Payload
				}) {
					t.Errorf("%s: Zustand verschieden von dem ohne Header", kontext)
				}
			}
		}
	}
	if n != 30 {
		t.Errorf("Menge %d, erwartet 30", n)
	}
}

// TestIfMatchPostIgnoriert (F25): POST wertet die Header nie aus: 201, bei vorhandenem Schlüssel ODR-E6009. Header 2 × Wert 5 × Eintrag 2 = 20 Läufe.
func TestIfMatchPostIgnoriert(t *testing.T) {
	n := 0
	for _, k := range beideKoepfe() {
		for _, w := range []string{etagLit(1), etagLit(99), "*", "x", ""} {
			for _, vorhanden := range []bool{false, true} {
				n++
				id := map[bool]string{false: "7", true: "2"}[vorhanden]
				u := etagUmg(t, true, 5, 5, 5)
				kontext := fmt.Sprintf("POST %s %q vorhanden=%v", k, w, vorhanden)
				a, err := u.rufe(u.anfrageMitBody("POST", id, eingabeIDName(id, "neu"), k.mit([]string{w})))
				if vorhanden {
					pruefeFehler(t, kontext, err, e6009)
				} else {
					erwarteStatus(t, kontext, a, err, 201)
				}
			}
		}
	}
	if n != 20 {
		t.Errorf("Menge %d, erwartet 20", n)
	}
}

// TestIfMatchLesenIgnoriert (F26): GET und HEAD ignorieren die Header: kein 304, kein 412, die Antwort gleicht byte-gleich der ohne Header, bei
// aus und bei an. Methode und Ressource 4 (GET Entity, Collection, Anzahl; HEAD Entity) × Header 2 × Wert 5 × Schalter 2 = 80 Läufe.
func TestIfMatchLesenIgnoriert(t *testing.T) {
	arten := map[string]func(u *umg){
		"Entity":     func(u *umg) { u.entity(model.Optionen{}, "2") },
		"Collection": func(u *umg) { u.menge(model.Optionen{}) },
		"Anzahl":     func(u *umg) { u.anzahl(model.Optionen{}) },
	}
	n := 0
	for _, an := range []bool{false, true} {
		for _, m := range []string{"GET", "HEAD"} {
			for name, bau := range arten {
				if m == "HEAD" && name != "Entity" {
					continue
				}
				ohne := etagUmg(t, an, 4, 6, 9)
				bau(ohne)
				aOhne, err := ohne.rufe(anfrage(m, dienstRoot+"/x"))
				if err != nil {
					t.Fatalf("%s %s ohne Header: %v", m, name, err)
				}
				bOhne := koerper(t, aOhne)
				for _, k := range beideKoepfe() {
					for _, w := range []string{etagLit(6), etagLit(99), "*", "x", ""} {
						n++
						u := etagUmg(t, an, 4, 6, 9)
						bau(u)
						a := anfrage(m, dienstRoot+"/x")
						a.Header[string(k)] = []string{w}
						got, err := u.rufe(a)
						kontext := fmt.Sprintf("an=%v %s %s %s %q", an, m, name, k, w)
						if err != nil || got.Status != 200 || got.Status != aOhne.Status || fmt.Sprint(got.Header) != fmt.Sprint(aOhne.Header) || !slices.Equal(koerper(t, got), bOhne) {
							t.Errorf("%s: Status %d Fehler %v, ohne Header %d", kontext, got.Status, err, aOhne.Status)
						}
					}
				}
			}
		}
	}
	// Methode × Ressource: GET 3 + HEAD 1 = 4, × Header 2 × Wert 5 × Schalter 2.
	if n != 4*2*5*2 {
		t.Errorf("Menge %d, erwartet %d", n, 4*2*5*2)
	}
}

// versionenAnSternGrenze sind die Versionen für If-None-Match: * an einem vorhandenen Eintrag (SPEC-283 Nr. 5 b): die Grenze 0/1/2, Zehnerpotenzen und große Zähler.
func versionenAnSternGrenze() []int64 {
	return []int64{0, 1, 2, 3, 9, 10, 99, 100, 4294967296, 9223372036854775806}
}

// TestIfMatchSternBeiJederVersion (F11, Klasse Version 0 bis 2^63-1): If-None-Match: * ist bei vorhandenem Eintrag nicht erfüllt, auch an einem
// unveränderten Baseline-Eintrag (Version 0); If-Match: * ist bei jeder Version erfüllt. Methode 3 × Version 10 × Header 2 = 60 Läufe.
func TestIfMatchSternBeiJederVersion(t *testing.T) {
	n := 0
	for _, m := range bedingteMethoden() {
		for _, v := range versionenAnSternGrenze() {
			for _, k := range beideKoepfe() {
				n++
				kontext := fmt.Sprintf("%s %s Version %d", m, k, v)
				u := etagUmg(t, true, v, v, v)
				a, err := u.mutiereMit(m, "2", k.mit([]string{"*"}))
				if k == ifMatch {
					erwarteStatus(t, kontext, a, err, 204)
					continue
				}
				erwarte412(t, kontext, err)
				u.abgewiesen(kontext)
			}
		}
	}
	if n != 3*10*2 {
		t.Errorf("Menge %d, erwartet %d", n, 3*10*2)
	}
}

// TestIfMatchAbbruchEntitiesGet (Raster Abbruch, SPEC-283 Nr. 9 d): ein Fehler an Entities.Get geht der Vorbedingung vor, auch bei nicht erfülltem
// Header: ein Port-Fehler ist sein Code, ein Kontextfehler bleibt ohne Code und ist errors.Is gleich; nichts wird geschrieben.
// Methode 3 × Header 2 × Fehlerart 3 = 18 Läufe.
func TestIfMatchAbbruchEntitiesGet(t *testing.T) {
	arten := []struct {
		name string
		err  error
		code model.Code
	}{
		{"Port-Fehler", outbound.ErrBeschaedigt, model.CodeE3001},
		{"context.Canceled", context.Canceled, 0},
		{"context.DeadlineExceeded", context.DeadlineExceeded, 0},
	}
	n := 0
	for _, m := range bedingteMethoden() {
		for _, k := range beideKoepfe() {
			for _, a := range arten {
				n++
				kontext := fmt.Sprintf("%s %s %s", m, k, a.name)
				u := etagUmg(t, true, 5, 5, 5)
				u.p.setzeFehler("Entities.Get", a.err)
				_, err := u.mutiereMit(m, "2", k.mit(k.verletzt(5)))
				if a.code != 0 {
					pruefeFehler(t, kontext, err, a.code)
				} else if !errors.Is(err, a.err) || fehlercode(err) != 0 {
					t.Errorf("%s: Fehler %v, erwartet %v ohne Code", kontext, err, a.err)
				}
				u.ohneSchreiben(kontext)
				if u.p.anzahl("Entities.Get") != 1 || u.p.anzahl("Coverage.Get") != 0 || u.p.anzahl("Schreiber.SchreibeAnzahl") != 0 {
					t.Errorf("%s: Entities.Get %d, Coverage.Get %d", kontext, u.p.anzahl("Entities.Get"), u.p.anzahl("Coverage.Get"))
				}
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestIfMatchBeendeterKontext (SPEC-283 Nr. 9 d zweite Hälfte "prüft keinen Kontext"): bei einem vor dem Aufruf beendeten ctx (abgebrochen, abgelaufen)
// und Ports, die ihn nicht beachten, wertet der Dienst die Vorbedingung wie sonst aus: ein erfüllter Header lässt die Mutation zu (204, Put einmal),
// ein nicht erfüllter ist ODR-E6010 und nicht der Kontextfehler. Methode 3 × Header 2 × Kontext 2 × Header erfüllt oder nicht 2 = 24 Läufe.
func TestIfMatchBeendeterKontext(t *testing.T) {
	kontexte := map[string]func() (context.Context, context.CancelFunc){
		"abgebrochen": func() (context.Context, context.CancelFunc) {
			c, cancel := context.WithCancel(context.Background())
			cancel()
			return c, cancel
		},
		"abgelaufen": func() (context.Context, context.CancelFunc) {
			return context.WithDeadline(context.Background(), time.Unix(0, 0))
		},
	}
	n := 0
	for name, neu := range kontexte {
		for _, m := range bedingteMethoden() {
			for _, k := range beideKoepfe() {
				for _, erfuellt := range []bool{true, false} {
					n++
					kontext := fmt.Sprintf("%s %s %s erfüllt=%v", name, m, k, erfuellt)
					u := etagUmg(t, true, 5, 5, 5)
					c, cancel := neu()
					u.ctx = c
					wert := k.verletzt(5)
					if erfuellt {
						wert = k.erfuellt(5)
					}
					a, err := u.mutiereMit(m, "2", k.mit(wert))
					cancel()
					if erfuellt {
						erwarteStatus(t, kontext, a, err, 204)
						if u.p.anzahl("Entities.Put") != 1 {
							t.Errorf("%s: Put %d Mal, erwartet 1", kontext, u.p.anzahl("Entities.Put"))
						}
						continue
					}
					erwarte412(t, kontext, err)
					u.ohneSchreiben(kontext)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}
