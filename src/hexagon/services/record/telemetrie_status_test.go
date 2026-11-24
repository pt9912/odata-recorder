package record_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// statusJeFall ist der Status am Ende je Fall von dienstFaelle als Literal (SPEC-295 Nr. 7 a): complete 3, incomplete 6, failed 7.
func statusJeFall() map[string]string {
	return map[string]string{
		"Gutfall": "complete", "Drain, beide Ergebnisse": "complete", "nicht angehängt": "complete",
		"Abbruch in der Response": "incomplete", "ohne Antwort, nicht erreichbar": "incomplete", "ohne Antwort, Zeit": "incomplete",
		"ohne Antwort, Client-Abbruch": "incomplete", "ohne Antwort, Drain": "incomplete", "Drain, Response läuft": "incomplete",
		"typ": "failed", "syntax": "failed", "zu_gross": "failed", "Abbruch im Request": "failed",
		"ohne Antwort, anderer Fehler mit Geheimnis im Text": "failed", "Drain, Request läuft": "failed",
		"Drain, Request verworfen, Response läuft": "failed",
	}
}

// TestStatusJeAusgang hält C5 (SPEC-295 Nr. 4 a, b, 7 a): Menge M-Status, jeder Fall von dienstFaelle (16) und ein Request mit multipart/mixed (1):
// die Kennung der Session steht bei StartSpan, der Platzhalter dort, genau ein Set trägt den Status der Aufnahme (auch wenn der Exchange nicht
// angehängt ist); die Session kommt aus dem Header (zwei offene) und aus der einen offenen.
func TestStatusJeAusgang(t *testing.T) {
	faelle := dienstFaelle()
	zaehl := map[string]int{}
	status := statusJeFall()
	if len(faelle) != 16 || len(status) != 16 {
		t.Fatalf("%d Fälle und %d Status, erwartet je 16", len(faelle), len(status))
	}
	for _, f := range faelle {
		want, ok := status[f.name]
		if !ok {
			t.Errorf("Fall %q hat keinen Status im Orakel", f.name)
			continue
		}
		zaehl[want]++
		u, j := neuMitJournal(t)
		f.lauf(t, u)
		if b := buecher(j.FakeTelemetry); len(b) != 1 {
			t.Errorf("%s: %d Bücher, erwartet 1", f.name, len(b))
		} else {
			pruefeBuch(t, f.name, b[0], "s1", want)
		}
	}
	if zaehl["complete"] != 3 || zaehl["incomplete"] != 6 || zaehl["failed"] != 7 {
		t.Errorf("Status je Fall %v, erwartet complete 3, incomplete 6, failed 7", zaehl)
	}
	u, j := neuMitJournal(t)
	u.upstream.plane(geheimeAntwort(koerper(resJSON)))
	normal(t, u, geheimeAnfrage("multipart/mixed; boundary=b", "--b--"))
	pruefeBuch(t, "multipart/mixed", buecher(j.FakeTelemetry)[0], "s1", "failed")
	for name, id := range map[string]string{"Header": "s2", "eine offene": "s1"} {
		v, k := neuMitJournal(t)
		a := anfrage("GET", "/a/x")
		if name == "Header" {
			v.sessions.setze("s2", "svc-a", model.SessionOpen)
			a = mitSession(a, sessionHeader, id)
		}
		v.rundlauf(a)
		pruefeBuch(t, name, buecher(k.FakeTelemetry)[0], id, "complete")
	}
}

// TestCtxJeAufrufstelle hält C6 (SPEC-295 Nr. 3 d, 6): StartSpan erhält das ctx von Record unverändert, und keiner der 5 Ports hinter dem Beginn
// (Senden, Tee des Requests, Tee der Antwort, ListBySession, Append) sieht das ctx von StartSpan; jeder sieht den Wert des Aufrufs.
func TestCtxJeAufrufstelle(t *testing.T) {
	u, j := neuMitJournal(t)
	ctx := context.WithValue(context.Background(), ctxAufruf{}, "aufruf")
	u.upstream.plane(antwort(200, map[string][]string{"Content-Type": {"application/json"}}, koerper(`{"a":1}`)))
	ant, err := u.record(ctx, mitBody(anfrage("POST", "/a/x"), "application/json", `{"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	liesUndSchliesse(t, ant.Body)
	if len(u.upstream.gesehen()) != 1 || len(u.blobs.ctxs) != 2 || len(u.exchanges.listCtx) != 1 || len(u.exchanges.appendCtx) != 1 {
		t.Fatalf("Aufrufe: Senden %d, Begin %d, ListBySession %d, Append %d; erwartet 1, 2, 1, 1",
			len(u.upstream.gesehen()), len(u.blobs.ctxs), len(u.exchanges.listCtx), len(u.exchanges.appendCtx))
	}
	stellen := map[string]context.Context{
		"Senden": u.upstream.gesehen()[0].ctx(), "Tee des Requests": u.blobs.ctxs[0], "Tee der Antwort": u.blobs.ctxs[1],
		"ListBySession": u.exchanges.listCtx[0], "Append": u.exchanges.appendCtx[0],
	}
	if len(stellen) != 5 {
		t.Fatalf("%d Stellen, erwartet 5", len(stellen))
	}
	for name, c := range stellen {
		if c.Value(ctxAufruf{}) != "aufruf" || c.Value(spanMarke{}) != nil {
			t.Errorf("%s: Wert des Aufrufs %v, Marke von StartSpan %v; erwartet aufruf und nil", name, c.Value(ctxAufruf{}), c.Value(spanMarke{}))
		}
	}
	if b := buecher(j.FakeTelemetry); len(b) != 1 || b[0].ctx() != ctx {
		t.Errorf("StartSpan erhielt nicht das ctx des Aufrufs: %d Bücher", len(b))
	}
}

// TestKeinCountErrorKeinLog hält C8 (SPEC-295 Nr. 1, 5, 6): in jedem der 18 Wege vor der Annahme, in E4001 mit 502 und 504 und in den drei
// Einträgen capture, upstream und ablage ruft der Dienst weder CountError noch Log des Ports; die Läufe tragen die 8 Codes und die 3 Ereignisse.
func TestKeinCountErrorKeinLog(t *testing.T) {
	codes, status, ereignisse := map[model.Code]bool{}, map[int]bool{}, map[string]bool{}
	pruefe := func(was string, u *umgebung, err error) {
		t.Helper()
		var fe *inbound.Fehler
		if errors.As(err, &fe) {
			codes[fe.Code], status[fe.Status] = true, true
		}
		if n, l := u.tel.Anzahl(supporttest.OpCountError), u.tel.Anzahl(supporttest.OpLog); n != 0 || l != 0 {
			t.Errorf("%s: %d CountError und %d Log, erwartet 0 und 0", was, n, l)
		}
		for _, e := range u.log.alle() {
			ereignisse[e.attrs["ereignis"]] = true
		}
	}
	for _, f := range vorFaelle() {
		u, err := f.fahre(t)
		pruefe(f.name, u, err)
	}
	for name, fehler := range map[string]error{"502": outbound.ErrUpstreamNichtErreichbar, "504": outbound.ErrUpstreamZeit} {
		u, _ := neuMitJournal(t)
		u.upstream.plane(&plan{lies: -1, fehler: fehler})
		_, err := u.record(context.Background(), anfrage("GET", "/a/x"))
		pruefe("E4001 "+name, u, err)
	}
	u, _ := neuMitJournal(t)
	normal(t, u, geheimeAnfrage("text/plain", "x"))
	pruefe("capture", u, nil)
	v, _ := neuMitJournal(t)
	v.exchanges.appendFehler[0] = errDisk
	v.rundlauf(anfrage("GET", "/a/x"))
	pruefe("ablage", v, nil)
	want := []model.Code{model.CodeE5003, model.CodeE5004, model.CodeE5005, model.CodeE5006, model.CodeE5009, model.CodeE7003, model.CodeE4001, model.CodeE9001}
	if len(codes) != len(want) || slices.ContainsFunc(want, func(c model.Code) bool { return !codes[c] }) {
		t.Errorf("Läufe tragen die Codes %v, erwartet %v", slices.Collect(maps.Keys(codes)), want)
	}
	if !ereignisse["capture"] || !ereignisse["upstream"] || !ereignisse["ablage"] || !status[502] || !status[504] {
		t.Errorf("Ereignisse %v, Status %v; erwartet capture, upstream, ablage und 502, 504", ereignisse, status)
	}
}

// TestQuelltextRuftDreiOperationen hält C8 gegen den Quelltext (SPEC-295 Nr. 6): das Paket nennt StartSpan, Set und End je einmal, weder CountError
// noch Snapshot noch Log an Telemetrie.
func TestQuelltextRuftDreiOperationen(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) < 8 {
		t.Fatalf("Dateien %v: %v", dateien, err)
	}
	zaehl := map[string]int{}
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			x, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			zaehl[x.Sel.Name]++
			if in, da := x.X.(*ast.SelectorExpr); da && x.Sel.Name == "Log" && in.Sel.Name == "Telemetrie" {
				t.Errorf("%s: Log an Telemetrie", d)
			}
			return true
		})
	}
	if zaehl["StartSpan"] != 1 || zaehl["Set"] != 1 || zaehl["End"] != 1 || zaehl["CountError"] != 0 || zaehl["Snapshot"] != 0 {
		t.Errorf("Selektoren StartSpan %d, Set %d, End %d, CountError %d, Snapshot %d; erwartet 1, 1, 1, 0, 0",
			zaehl["StartSpan"], zaehl["Set"], zaehl["End"], zaehl["CountError"], zaehl["Snapshot"])
	}
}

// lage ist ein Lauf mit den Werten, die ein Canary aufnimmt.
type lage struct {
	a         inbound.Anfrage
	antHeader map[string][]string
	antBody   string
	ursache   error
}

// kanarie ist ein wohlgeformter Wert (Buchstaben und Ziffern): ihn lässt die Form eines Attributs durch, so dass nur der Dienst ihn fernhält
// (SPEC-232 Nr. 2, *Ungeprüft*).
const kanarie = "KanarieQ7f"

// quelle ist eine der neun Stellen, an denen ein Wert des Requests oder der Antwort liegt (SPEC-295 Nr. 8); bestehtNur ist die Art, in der sie
// besteht: 0 beide, 1 angehängt, 2 ohne Antwort.
type quelle struct {
	name       string
	setze      func(l *lage)
	bestehtNur int
}

func neueLage() lage {
	return lage{a: anfrage("GET", "/a/x"), antHeader: map[string][]string{"Content-Type": {"application/json"}}, antBody: `{"a":1}`}
}

func quellen() []quelle {
	return []quelle{
		{"Pfad", func(l *lage) { l.a.Pfad = "/a/" + kanarie }, 0},
		{"Name der Query", func(l *lage) { l.a.RawQuery = kanarie + "=1" }, 0},
		{"Wert der Query", func(l *lage) { l.a.RawQuery = "q=" + kanarie }, 0},
		{"Name des Headers", func(l *lage) { l.a.Header["X-"+kanarie] = []string{"1"} }, 0},
		{"Wert des Headers", func(l *lage) { l.a.Header["X-Info"] = []string{kanarie} }, 0},
		{"Body des Requests", func(l *lage) { l.a = mitBody(l.a, "application/json", `{"x":"`+kanarie+`"}`) }, 0},
		{"Header der Antwort", func(l *lage) { l.antHeader["X-Out"] = []string{kanarie} }, 1},
		{"Body der Antwort", func(l *lage) { l.antBody = `{"y":"` + kanarie + `"}` }, 1},
		{"Text der Ursache von Senden", func(l *lage) { l.ursache = errors.New(kanarie) }, 2},
	}
}

// TestEintraegeOhneRequestWert hält E3 für den Record-Dienst (SPEC-231 Nr. 8): Menge M-Quellen-Einträge, 3 Stellen (capture, upstream, ablage) × 9
// Quellen gleich 27 Paare; an jeder Quelle liegt der Canary, die Stelle löst ihren Eintrag aus (die Antwort ist text/plain: capture; Senden
// scheitert: upstream; Append scheitert: ablage), und kein Text und kein Attribut eines Eintrags trägt ihn. Der Text der Ursache liegt bei
// upstream im Fehler von Senden und bei ablage im Fehler von Append. Drei Paare tragen den Canary nicht an die Stelle, weil die Stelle den Wert nicht
// lesen kann (`leer`: bei capture der Fehler von Senden, bei upstream Header und Body der Antwort, die es dort nicht gibt); sie laufen mit und sind
// gezählt. Die Zahl der Stellen gleicht der Zahl der Aufrufe von `protokolliere` im Quelltext des Pakets.
func TestEintraegeOhneRequestWert(t *testing.T) {
	stellen := []string{"capture", "upstream", "ablage"}
	leer := map[string]bool{"capture/Text der Ursache von Senden": true, "upstream/Header der Antwort": true, "upstream/Body der Antwort": true}
	n, nLeer := 0, 0
	for _, stelle := range stellen {
		for _, q := range quellen() {
			n++
			if leer[stelle+"/"+q.name] {
				nLeer++
			}
			u := neu(t)
			l := neueLage()
			q.setze(&l)
			ursache := l.ursache
			if ursache == nil {
				ursache = errDisk
			}
			switch stelle {
			case "capture":
				l.antHeader["Content-Type"] = []string{"text/plain"}
				u.upstream.plane(antwort(200, l.antHeader, koerper(l.antBody)))
				u.rundlaufMit(l.a)
			case "upstream":
				u.upstream.plane(&plan{lies: -1, fehler: ursache})
				_, _ = u.record(context.Background(), l.a)
			default:
				u.exchanges.appendFehler[0] = ursache
				u.upstream.plane(antwort(200, l.antHeader, koerper(l.antBody)))
				u.rundlaufMit(l.a)
			}
			gefunden := 0
			for _, e := range u.log.alle() {
				if e.attrs["ereignis"] == stelle {
					gefunden++
				}
				alle := e.text
				for k, v := range e.attrs {
					alle += " " + k + "=" + v
				}
				if strings.Contains(alle, kanarie) {
					t.Errorf("%s, %s: der Eintrag trägt den Canary: %s", stelle, q.name, alle)
				}
			}
			if gefunden == 0 {
				t.Errorf("%s, %s: kein Eintrag %s", stelle, q.name, stelle)
			}
		}
	}
	if n != len(stellen)*len(quellen()) || nLeer != len(leer) {
		t.Errorf("%d Paare, %d leere; erwartet %d Paare (Stellen × Quellen) und %d leere", n, nLeer, len(stellen)*len(quellen()), len(leer))
	}
	if q := zaehleAufrufe(t, ".", func(name string) bool { return name == "protokolliere" }); q != len(stellen) {
		t.Errorf("%d Aufrufe von protokolliere im Quelltext, %d Stellen im Test", q, len(stellen))
	}
}

// zaehleAufrufe zählt im Quelltext des Pakets in verz (ohne Tests) die Aufrufe, deren Selektor passt.
func zaehleAufrufe(t *testing.T, verz string, passt func(name string) bool) int {
	t.Helper()
	dateien, err := filepath.Glob(filepath.Join(verz, "*.go"))
	if err != nil || len(dateien) == 0 {
		t.Fatalf("Dateien in %s: %v, %d", verz, err, len(dateien))
	}
	n := 0
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(x ast.Node) bool {
			if c, ok := x.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok && passt(s.Sel.Name) {
					n++
				}
			}
			return true
		})
	}
	return n
}

// rundlaufMit führt den Request mit der gesetzten Antwort aus; liefert Record einen Fehler, bleibt es dabei.
func (u *umgebung) rundlaufMit(a inbound.Anfrage) {
	u.t.Helper()
	if ant, err := u.record(context.Background(), a); err == nil {
		liesUndSchliesse(u.t, ant.Body)
	}
}

// TestKeinRequestWertImSpan hält C9 (SPEC-295 Nr. 8): Menge M-Stellen-Record, 9 Stellen × 2 Arten (angehängt, ohne Antwort) gleich 18; an jeder
// Stelle liegt der Canary, und kein Schlüssel und kein Wert eines Aufrufs von Telemetrie trägt ihn. Wo die Stelle in der Art besteht, hat der
// Canary Upstream, Exchange oder Fehler erreicht (Gegenprobe).
func TestKeinRequestWertImSpan(t *testing.T) {
	n := 0
	for _, s := range quellen() {
		for art := 1; art <= 2; art++ {
			n++
			u := neu(t)
			l := neueLage()
			s.setze(&l)
			var err error
			if art == 1 {
				u.upstream.plane(antwort(200, l.antHeader, koerper(l.antBody)))
				var ant inbound.Antwort
				if ant, err = u.record(context.Background(), l.a); err == nil {
					liesUndSchliesse(t, ant.Body)
				}
			} else {
				if l.ursache == nil {
					l.ursache = outbound.ErrUpstreamNichtErreichbar
				}
				u.upstream.plane(&plan{lies: -1, fehler: l.ursache})
				_, err = u.record(context.Background(), l.a)
			}
			was := fmt.Sprintf("%s, Art %d", s.name, art)
			erreicht := fmt.Sprintf("%+v %+v %s %v", u.exchanges.angehaengt(), u.upstream.gesehen(), u.blobs.alleAngeboten(), errors.Unwrap(err))
			if (s.bestehtNur == 0 || s.bestehtNur == art) && !strings.Contains(erreicht, kanarie) {
				t.Errorf("%s: der Canary erreichte weder Upstream noch Exchange noch Fehler", was)
			}
			aufrufe := u.tel.Aufrufe()
			if u.tel.Anzahl(supporttest.OpStartSpan) != 1 {
				t.Errorf("%s: %d StartSpan, erwartet 1", was, u.tel.Anzahl(supporttest.OpStartSpan))
			}
			for _, a := range aufrufe {
				for _, at := range a.Attrs {
					if strings.Contains(at.Key+at.Text, kanarie) {
						t.Errorf("%s: Attribut %s=%q trägt den Canary", was, at.Key, at.Text)
					}
				}
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Paare, erwartet 18", n)
	}
}

// TestTexteDerEintraege hält E1 (SPEC-121 Nr. 12): die drei Einträge tragen den Text der Spezifikation als Literal; ä entsteht aus Bytewerten.
func TestTexteDerEintraege(t *testing.T) {
	ae := string([]byte{0xc3, 0xa4})
	soll := map[string]string{
		"capture":  "record: Aufzeichnung eines Bodys verworfen",
		"upstream": "record: Upstream ohne Antwort",
		"ablage":   "record: Exchange nicht angeh" + ae + "ngt",
	}
	laeufe := map[string]func(u *umgebung){
		"capture":  func(u *umgebung) { normal(t, u, geheimeAnfrage("text/plain", "x")) },
		"upstream": func(u *umgebung) { laufeWeg(t, u, wegOhneAntwort) },
		"ablage": func(u *umgebung) {
			u.exchanges.appendFehler[0] = errDisk
			u.rundlauf(anfrage("GET", "/a/x"))
		},
	}
	if len(soll) != 3 || len(laeufe) != 3 {
		t.Fatalf("%d Texte, %d Läufe, erwartet je 3", len(soll), len(laeufe))
	}
	for ereignis, text := range soll {
		u := neu(t)
		laeufe[ereignis](u)
		var got []string
		for _, e := range u.log.alle() {
			if e.attrs["ereignis"] == ereignis {
				got = append(got, e.text)
			}
		}
		if len(got) == 0 || got[0] != text {
			t.Errorf("%s: Texte %q, erwartet %q", ereignis, got, text)
		}
	}
}
