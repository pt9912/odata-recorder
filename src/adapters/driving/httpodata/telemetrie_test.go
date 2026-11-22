package httpodata_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// kanarie bildet den Wert, den der Test an einer Stelle des Requests setzt, aus Bytewerten (AGENTS.md §3.18).
func kanarie(gross bool) string {
	b := []byte{0x63, 0x61, 0x6e, 0x61, 0x72, 0x79}
	if gross {
		return strings.ToUpper(string(b))
	}
	return string(b)
}

// TestNewTelemetrieUngueltig hält G1 (SPEC-294 Nr. 1): nil Telemetrie, leerer Modus und beide sind ErrUngueltig ohne Server, der Text nennt
// den Wert des Modus nicht, und New ruft keine Operation von Telemetrie; die gültige Konfiguration besteht.
func TestNewTelemetrieUngueltig(t *testing.T) {
	faelle := []struct {
		name  string
		tel   bool
		modus string
	}{
		{"nil Telemetrie", false, kanarie(false)},
		{"leerer Modus", true, ""},
		{"beide", false, ""},
	}
	if len(faelle) != 3 {
		t.Fatalf("%d Fälle, erwartet 3", len(faelle))
	}
	for _, f := range faelle {
		tel := supporttest.NewFakeTelemetry()
		lis := &zaehlListener{}
		k := httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: lis, Modus: f.modus}
		if f.tel {
			k.Telemetrie = tel
		}
		srv, err := httpodata.New(k)
		if !errors.Is(err, model.ErrUngueltig) || srv != nil {
			t.Errorf("%s: srv=%v err=%v, erwartet nil und ErrUngueltig", f.name, srv, err)
		}
		if err != nil && f.modus != "" && strings.Contains(err.Error(), f.modus) {
			t.Errorf("%s: der Fehlertext nennt den Wert des Modus: %v", f.name, err)
		}
		if n := len(tel.Aufrufe()); n != 0 || lis.aufrufe != 0 {
			t.Errorf("%s: New rief Telemetrie %d-mal und den Listener %d-mal", f.name, n, lis.aufrufe)
		}
	}
	tel := supporttest.NewFakeTelemetry()
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: &zaehlListener{}, Telemetrie: tel, Modus: "record"})
	if err != nil || srv == nil || len(tel.Aufrufe()) != 0 {
		t.Errorf("gültige Konfiguration: srv=%v err=%v, Aufrufe von Telemetrie %d", srv, err, len(tel.Aufrufe()))
	}
}

// TestGesamtspanJederWeg hält G2 (SPEC-294 Nr. 2, 7 a): auf jedem der 14 Wege des Handlers, mit jeder der 14 Methoden von M-Methoden (196
// Fälle), öffnet er den Span `odata.gateway.request` einmal und beendet ihn einmal; die Schlüssel von StartSpan und Set sind genau die
// Pflicht-Schlüssel von support.Spans() (ohne den Status bei den Wegen ohne WriteHeader), nie ein weiterer; die Zahl der Zähler folgt dem Weg.
func TestGesamtspanJederWeg(t *testing.T) {
	ws := wege()
	if len(ws) != 14 {
		t.Fatalf("%d Wege, erwartet 14", len(ws))
	}
	pflicht := support.Spans()[0].Pflicht
	if len(pflicht) != 3 {
		t.Fatalf("support.Spans() nennt %d Pflicht-Schlüssel für den Gesamtspan, erwartet 3", len(pflicht))
	}
	var faelle []weg
	for i := range ws {
		faelle = append(faelle, wegMitMethoden(i)...)
	}
	if len(faelle) != 14*14 {
		t.Fatalf("%d Fälle, erwartet 196 (14 Wege × 14 Methoden)", len(faelle))
	}
	for _, w := range faelle {
		w.name += " " + w.methode
		p, jw, abbruch := lauf(t, "record", w)
		b := buecher(p.tel.FakeTelemetry)
		if len(b) != 1 {
			t.Errorf("%s: %d Spans, erwartet 1", w.name, len(b))
			continue
		}
		if b[0].name != "odata.gateway.request" || b[0].ends != 1 {
			t.Errorf("%s: Span %q mit %d End, erwartet odata.gateway.request und 1", w.name, b[0].name, b[0].ends)
		}
		var schluessel []string
		for _, a := range b[0].start {
			schluessel = append(schluessel, a.Key)
		}
		for _, s := range b[0].sets {
			for _, a := range s {
				schluessel = append(schluessel, a.Key)
			}
		}
		soll := slices.Clone(pflicht)
		if w.status == 0 {
			soll = slices.DeleteFunc(soll, func(k string) bool { return k == "http.response.status_code" })
		}
		slices.Sort(schluessel)
		slices.Sort(soll)
		gleich(t, w.name+": Schlüssel", schluessel, soll)
		var sollZaehler []model.Code
		if w.zaehler != 0 {
			sollZaehler = []model.Code{w.zaehler}
		}
		gleich(t, w.name+": Zähler", zaehlerFolge(p.tel.FakeTelemetry), sollZaehler)
		if abbruch != w.abbruch || jw.status != w.status {
			t.Errorf("%s: Abbruch %v, Status %d; erwartet %v und %d", w.name, abbruch, jw.status, w.abbruch, w.status)
		}
	}
}

// TestGesamtspanNichtVorDemHandler hält G2 (SPEC-294 Nr. 2): ein Request, den net/http vor dem Handler abweist (ein Pfad mit ungültigem
// Escape, ein Kopf über der Grenze), trägt weder Span noch Zähler.
func TestGesamtspanNichtVorDemHandler(t *testing.T) {
	faelle := []struct {
		name  string
		bytes []byte
		kopf  string
	}{
		{"ungültiger Escape", zeile("GET", "/a%zz"), "HTTP/1.1 400"},
		{"Kopf über der Grenze", zeile("GET", "/p", "X-Gross: "+strings.Repeat("a", 70000)), "HTTP/1.1 431"},
	}
	for _, f := range faelle {
		u := neuUmgebung(t, "record", nil)
		z, _ := teile(u.roh(t, f.bytes))
		if len(z) == 0 || !strings.HasPrefix(z[0], f.kopf) {
			t.Errorf("%s: Antwort %v, erwartet %s von net/http", f.name, z, f.kopf)
		}
		if n, c, d := u.tel.Anzahl(supporttest.OpStartSpan), u.tel.Anzahl(supporttest.OpCountError), len(u.d.alle()); n != 0 || c != 0 || d != 0 {
			t.Errorf("%s: StartSpan %d, CountError %d, Aufrufe des Dienstes %d; erwartet 0", f.name, n, c, d)
		}
	}
}

// TestGesamtspanBeginnFolge hält G3 (SPEC-294 Nr. 3 a): der Beginn folgt auf KorrelationsID und geht dem Dienst voraus; auch die Antwort 503
// trägt den Span.
func TestGesamtspanBeginnFolge(t *testing.T) {
	p, _, _ := lauf(t, "record", wege()[0])
	gleich(t, "Gutfall", p.j.folge()[:3], []string{"KorrelationsID", "StartSpan", "Verarbeiten"})
	p, _, _ = lauf(t, "record", wege()[8])
	gleich(t, "nach Anhalten", p.j.folge()[:5], []string{"KorrelationsID", "StartSpan", "CountError:ODR-E5010", "Set", "End"})
}

// TestGesamtspanEndeFolge hält G4 (SPEC-294 Nr. 3 b): vor WriteHeader stehen in dieser Folge CountError (nur beim Fehlerobjekt), Set und End;
// Write und Flush liegen dahinter. 4 Wege mit WriteHeader (Antwort mit Body, Fehlerobjekt, 503, Bruch) × 2 Methoden gleich 8 Fälle.
func TestGesamtspanEndeFolge(t *testing.T) {
	kopf := []string{"KorrelationsID", "StartSpan"}
	faelle := []struct {
		weg int // Index in wege()
		get []string
		hd  []string
	}{
		{0, []string{"Verarbeiten", "Set", "End", "WriteHeader:200", "Write", "Flush"}, nil},
		{2, []string{"Verarbeiten", "CountError:ODR-E6008", "Set", "End", "WriteHeader:404", "Write"}, []string{"Verarbeiten", "CountError:ODR-E6008", "Set", "End", "WriteHeader:404"}},
		{8, []string{"CountError:ODR-E5010", "Set", "End", "WriteHeader:503", "Write"}, []string{"CountError:ODR-E5010", "Set", "End", "WriteHeader:503"}},
		{4, []string{"Verarbeiten", "CountError:ODR-E9001", "Set", "End", "WriteHeader:500", "Write"}, []string{"Verarbeiten", "CountError:ODR-E9001", "Set", "End", "WriteHeader:500"}},
	}
	n := 0
	for _, f := range faelle {
		for _, m := range []string{"GET", "HEAD"} {
			n++
			w := wege()[f.weg]
			w.methode = m
			p, _, _ := lauf(t, "record", w)
			soll := f.get
			if m == "HEAD" && f.hd != nil {
				soll = f.hd
			}
			gleich(t, w.name+" "+m, p.j.folge(), append(slices.Clone(kopf), soll...))
		}
	}
	if n != 8 {
		t.Errorf("%d Fälle, erwartet 8", n)
	}
}

// methodenSatz ist M-Methoden (G5): zehn Methoden mit Form, ein Token aus 16 und eines aus 17 `A` (beide Seiten der Grenze) und zwei
// Verstöße gegen die Form (`get`, `M-SEARCH`), 14 Elemente.
func methodenSatz() []string {
	return []string{
		"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE", "PROPFIND",
		strings.Repeat("A", 16), strings.Repeat("A", 17), "get", "M-SEARCH",
	}
}

// TestGesamtspanMethode hält G5 (SPEC-294 Nr. 4 a): `http.request.method` ist r.Method byte-gleich und steht in StartSpan, nie in Set.
func TestGesamtspanMethode(t *testing.T) {
	ms := methodenSatz()
	if len(ms) != 14 {
		t.Fatalf("%d Methoden, erwartet 14", len(ms))
	}
	for _, m := range ms {
		p := neuProbe(t, "record", nil)
		p.ruf(neuJWriter(p.j), jRequest(context.Background(), m, "/p"))
		b := buecher(p.tel.FakeTelemetry)
		if len(b) != 1 {
			t.Fatalf("%q: %d Spans", m, len(b))
		}
		got, da := attrText(b[0].start, "http.request.method")
		if !da || got != m {
			t.Errorf("%q: Attribut in StartSpan %q (da %v), erwartet die Methode unverändert", m, got, da)
		}
		for _, s := range b[0].sets {
			if _, da := attrText(s, "http.request.method"); da {
				t.Errorf("%q: Set trägt die Methode", m)
			}
		}
	}
}

// TestGesamtspanModus hält G6 (SPEC-294 Nr. 4 b): `recorder.mode` ist Konfig.Modus unverändert und steht in StartSpan; ein Header des Requests
// ändert ihn nicht. 4 Werte und einer außerhalb der Form, 5 Fälle.
func TestGesamtspanModus(t *testing.T) {
	werte := []string{"record", "replay", "mock", "hybrid", "Mock"}
	for _, wert := range werte {
		p := neuProbe(t, wert, nil)
		r := jRequest(context.Background(), "GET", "/p")
		r.Header.Set("X-Recorder-Mode", kanarie(false))
		p.ruf(neuJWriter(p.j), r)
		b := buecher(p.tel.FakeTelemetry)
		if len(b) != 1 {
			t.Fatalf("%q: %d Spans", wert, len(b))
		}
		if got, da := attrText(b[0].start, "recorder.mode"); !da || got != wert {
			t.Errorf("%q: Attribut in StartSpan %q (da %v), erwartet den Modus der Konfiguration", wert, got, da)
		}
		for _, s := range b[0].sets {
			if _, da := attrText(s, "recorder.mode"); da {
				t.Errorf("%q: Set trägt den Modus", wert)
			}
		}
	}
	if len(werte) != 5 {
		t.Errorf("%d Werte, erwartet 5", len(werte))
	}
}

// TestGesamtspanStatus hält G7 (SPEC-294 Nr. 4 c): der Status im einen Set ist der, den der Adapter an WriteHeader übergibt: bei der Antwort
// des Dienstes 7 Stati mit beiden Rändern 200 und 599, bei jedem der 4 Brüche 500; genau ein Set mit einem Attribut der Art Int.
func TestGesamtspanStatus(t *testing.T) {
	faelle := []struct {
		name string
		tu   httpodata.Verarbeiter
		soll int
	}{
		{"200", antwortMit(200, leer()), 200}, {"201", antwortMit(201, leer()), 201}, {"204", antwortMit(204, leer()), 204},
		{"304", antwortMit(304, leer()), 304}, {"404", antwortMit(404, leer()), 404}, {"500", antwortMit(500, leer()), 500},
		{"599", antwortMit(599, leer()), 599},
		{"Bruch Body nil", bruchOhneBody(), 500}, {"Bruch 100", antwortMit(100, leer()), 500}, {"Bruch 199", antwortMit(199, leer()), 500},
		{"Bruch 600", antwortMit(600, leer()), 500},
	}
	if len(faelle) != 11 {
		t.Fatalf("%d Fälle, erwartet 11 (7 + 4)", len(faelle))
	}
	for _, f := range faelle {
		for _, m := range methodenSatz() {
			p := neuProbe(t, "record", f.tu)
			jw := neuJWriter(p.j)
			p.ruf(jw, jRequest(context.Background(), m, "/p"))
			b := buecher(p.tel.FakeTelemetry)
			if len(b) != 1 || len(b[0].sets) != 1 || len(b[0].sets[0]) != 1 {
				t.Errorf("%s %s: Buch %+v, erwartet 1 Span mit 1 Set mit 1 Attribut", f.name, m, b)
				continue
			}
			a := b[0].sets[0][0]
			if a.Key != "http.response.status_code" || a.Art != support.ArtInt || a.Ganz != int64(f.soll) || jw.status != f.soll {
				t.Errorf("%s %s: Set %+v, Status am WriteHeader %d; erwartet %d in beiden", f.name, m, a, jw.status, f.soll)
			}
		}
	}
}

// TestGesamtspanAbbruchOhneAntwort hält G8 (SPEC-294 Nr. 3 b, 4 c): ein Kontextfehler ohne Code bricht die Verbindung ab; der Span endet
// einmal, ohne Set, und nichts wird gezählt. 3 Formen über den echten Server.
func TestGesamtspanAbbruchOhneAntwort(t *testing.T) {
	formen := []struct {
		name string
		err  error
	}{
		{"Canceled", context.Canceled},
		{"DeadlineExceeded", context.DeadlineExceeded},
		{"Canceled mit Fehler", errors.Join(context.Canceled, &inbound.Fehler{Code: model.CodeE6008, Status: 404})},
	}
	if len(formen) != 3 {
		t.Fatalf("%d Formen, erwartet 3", len(formen))
	}
	for _, f := range formen {
		u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, f.err })
		resp, err := u.client().Get("http://" + u.addr + "/p")
		if err == nil {
			_ = resp.Body.Close()
			t.Errorf("%s: der Client sah die Antwort %d", f.name, resp.StatusCode)
		}
		warte(t, f.name+": End", func() bool { return u.tel.Anzahl(supporttest.OpEnd) >= 1 })
		if s, e, set, c := u.tel.Anzahl(supporttest.OpStartSpan), u.tel.Anzahl(supporttest.OpEnd), u.tel.Anzahl(supporttest.OpSet), u.tel.Anzahl(supporttest.OpCountError); s != 1 || e != 1 || set != 0 || c != 0 {
			t.Errorf("%s: StartSpan %d, End %d, Set %d, CountError %d; erwartet 1, 1, 0, 0", f.name, s, e, set, c)
		}
	}
}

// fehlerFall ist ein Element von M-Fehler (G9): der Fehler des Dienstes, Code und Status des geschriebenen Objekts.
type fehlerFall struct {
	name     string
	tu       httpodata.Verarbeiter
	anhalten bool
	code     model.Code
	status   int
}

func fehlerFaelle() []fehlerFall {
	var fs []fehlerFall
	for _, c := range model.AlleCodes() {
		for _, s := range c.Stati() {
			if s >= 400 {
				fs = append(fs, fehlerFall{name: c.String() + " mit " + http.StatusText(s), tu: fehlerMit(&inbound.Fehler{Code: c, Status: s}), code: c, status: s})
			}
		}
	}
	e9001 := func(name string, tu httpodata.Verarbeiter) fehlerFall {
		return fehlerFall{name: name, tu: tu, code: model.CodeE9001, status: 500}
	}
	// Die Achse „Stelle des typisierten nil-Zeigers im Baum“ (SPEC-136 Nr. 7 a): der nil-Zeiger gilt nicht und wird nicht entpackt; was neben
	// ihm steht, gilt wie ohne ihn (ein gültiger Fehler gilt, sonst ODR-E9001).
	var nz *inbound.Fehler
	gueltig := &inbound.Fehler{Code: model.CodeE6010, Status: 412}
	fs = append(fs,
		e9001("nil-Zeiger allein", fehlerMit(nz)),
		e9001("nil-Zeiger verpackt", fehlerMit(fmt.Errorf("Dienst: %w", nz))),
		fehlerFall{name: "nil-Zeiger als Ursache", tu: fehlerMit(&inbound.Fehler{Code: model.CodeE6010, Status: 412, Ursache: nz}), code: model.CodeE6010, status: 412},
		fehlerFall{name: "Join nil-Zeiger vor gültigem", tu: fehlerMit(errors.Join(nz, gueltig)), code: model.CodeE6010, status: 412},
		fehlerFall{name: "Join gültiger vor nil-Zeiger", tu: fehlerMit(errors.Join(gueltig, nz)), code: model.CodeE6010, status: 412},
		e9001("Join nil-Zeiger und anderer Fehler", fehlerMit(errors.Join(nz, errors.New("roh")))),
		e9001("ODR-E2001 ohne Status", fehlerMit(&inbound.Fehler{Code: model.CodeE2001, Status: 500})),
		e9001("ODR-E2401 ohne Status", fehlerMit(&inbound.Fehler{Code: model.CodeE2401, Status: 500})),
		e9001("Code 0", fehlerMit(&inbound.Fehler{Code: 0, Status: 500})),
		e9001("Code 9999", fehlerMit(&inbound.Fehler{Code: 9999, Status: 500})),
		e9001("Status 0", fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 0})),
		e9001("Status 399", fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 399})),
		e9001("Status 600", fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 600})),
		e9001("Status, den der Code nicht trägt", fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 500})),
		e9001("Fehler ohne Code", fehlerMit(errors.New("roh"))),
		e9001("ErrUngueltig", fehlerMit(model.ErrUngueltig)),
		e9001("Bruch Body nil", bruchOhneBody()),
		e9001("Bruch 100", antwortMit(100, leer())), e9001("Bruch 199", antwortMit(199, leer())), e9001("Bruch 600", antwortMit(600, leer())),
		fehlerFall{name: "Antwort 503", anhalten: true, code: model.CodeE5010, status: 503},
	)
	return fs
}

// TestFehlerzaehlerJederFall hält G9 (SPEC-294 Nr. 5): jedes Fehlerobjekt zählt genau einmal, mit dem Code des geschriebenen Objekts, in der
// Folge CountError, Set, End, WriteHeader. 42 Paare aus Code und Status und 21 weitere Fälle gleich 63 (darunter 6 Stellen eines typisierten
// nil-Zeigers `*inbound.Fehler` im Fehlerbaum), je mit jeder der 14 Methoden (882 Läufe; bei HEAD entfällt der Body); die 41 Codes zählt der Test
// gegen model.AlleCodes(), die 42 Paare sind Handkopie von SPEC-109 Nr. 2.
func TestFehlerzaehlerJederFall(t *testing.T) {
	fs := fehlerFaelle()
	if n := len(model.AlleCodes()); n != 41 {
		t.Fatalf("%d Codes, erwartet 41", n)
	}
	if len(fs) != 63 {
		t.Fatalf("%d Fälle, erwartet 63 (42 Paare und 21 weitere)", len(fs))
	}
	for _, f := range fs {
		for _, m := range methodenSatz() {
			p := neuProbe(t, "record", f.tu)
			if f.anhalten {
				p.srv.Anhalten()
			}
			jw := neuJWriter(p.j)
			p.ruf(jw, jRequest(context.Background(), m, "/p"))
			gleich(t, f.name+" "+m+": Zähler", zaehlerFolge(p.tel.FakeTelemetry), []model.Code{f.code})
			soll := []string{"CountError:" + f.code.String(), "Set", "End", "WriteHeader:" + strconv.Itoa(f.status), "Write"}
			if m == "HEAD" {
				soll = soll[:4]
			}
			folge := p.j.folge()
			gleich(t, f.name+" "+m+": Folge", folge[len(folge)-len(soll):], soll)
			var obj model.Fehlerobjekt
			if m == "HEAD" {
				if jw.body.Len() != 0 || jw.status != f.status {
					t.Errorf("%s HEAD: Status %d, Body %q; erwartet %d ohne Body", f.name, jw.status, jw.body.String(), f.status)
				}
			} else if err := json.Unmarshal(jw.body.Bytes(), &obj); err != nil || obj.Error.Code != f.code.String() || jw.status != f.status {
				t.Errorf("%s %s: Status %d, Body %q (%v); erwartet %d und den Code %s", f.name, m, jw.status, jw.body.String(), err, f.status, f.code)
			}
		}
	}
}

// TestKeinZaehlerFuerKeinFehlerobjekt hält G10 (SPEC-294 Nr. 5, 7 e): was kein Fehlerobjekt des Adapters ist, zählt nicht: der Kontextfehler
// ohne Code (3), daneben ein nil-Zeiger `*inbound.Fehler` vor und hinter dem Kontextfehler im Join (2; der nil-Zeiger ändert die Wertung nicht,
// SPEC-136 Nr. 7 a), der Lesefehler und der Schreibfehler nach dem Kopf (2) und die Antwort des Dienstes mit Status 404, 500 und 599 (3), 10
// Fälle mit jeder der 14 Methoden (140 Läufe): kein Zähler, bei den fünf Kontextfällen auch kein Fehlerobjekt (kein WriteHeader).
func TestKeinZaehlerFuerKeinFehlerobjekt(t *testing.T) {
	var faelle []weg
	for _, i := range []int{9, 10, 11, 12, 13} {
		faelle = append(faelle, wegMitMethoden(i)...)
	}
	var nz *inbound.Fehler
	for _, m := range methodenSatz() {
		faelle = append(faelle,
			weg{name: "Join nil-Zeiger vor Canceled", methode: m, tu: fehlerMit(errors.Join(nz, context.Canceled)), abbruch: true},
			weg{name: "Join Canceled vor nil-Zeiger", methode: m, tu: fehlerMit(errors.Join(context.Canceled, nz)), abbruch: true})
	}
	for _, s := range []int{404, 500, 599} {
		for _, m := range methodenSatz() {
			faelle = append(faelle, weg{name: "Antwort " + strconv.Itoa(s), methode: m, tu: antwortMit(s, leer()), status: s})
		}
	}
	if len(faelle) != 10*14 {
		t.Fatalf("%d Fälle, erwartet 140 (10 × 14 Methoden)", len(faelle))
	}
	for _, f := range faelle {
		p, jw, abbruch := lauf(t, "record", f)
		if got := zaehlerFolge(p.tel.FakeTelemetry); len(got) != 0 {
			t.Errorf("%s %s: CountError %v, erwartet keinen Aufruf", f.name, f.methode, got)
		}
		if abbruch != f.abbruch || jw.status != f.status {
			t.Errorf("%s %s: Abbruch %v, Status %d; erwartet %v und %d", f.name, f.methode, abbruch, jw.status, f.abbruch, f.status)
		}
	}
}

// isFehler ist ein Fehler, der nach seiner Methode Is der Kontextfehler context.DeadlineExceeded ist.
type isFehler struct{}

func (isFehler) Error() string { return "Is" }

func (isFehler) Is(ziel error) bool { return ziel == context.DeadlineExceeded }

// asFehler ist ein Fehler, dessen Methode As den *inbound.Fehler f setzt (ein nil-Zeiger gilt nicht) oder, mit nein, nichts setzt.
type asFehler struct {
	f    *inbound.Fehler
	nein bool
}

func (asFehler) Error() string { return "As" }

func (a asFehler) As(ziel any) bool {
	p, ok := ziel.(**inbound.Fehler)
	if !ok || a.nein {
		return false
	}
	*p = a.f
	return true
}

// TestFehlerbaumFolgtIsUndAs hält SPEC-136 Nr. 7 a (die Suche folgt errors.Is und errors.As): die Methode Is eines Fehlers macht ihn zum
// Kontextfehler (Abbruch, kein Zähler), die Methode As liefert den *inbound.Fehler (gilt), liefert einen nil-Zeiger oder verneint (dann gilt
// der Fall ohne ihn: ODR-E9001); je mit jeder der 14 Methoden.
func TestFehlerbaumFolgtIsUndAs(t *testing.T) {
	gueltig := &inbound.Fehler{Code: model.CodeE6010, Status: 412}
	faelle := []struct {
		name    string
		err     error
		abbruch bool
		code    model.Code
		status  int
	}{
		{"Is liefert den Kontextfehler", fmt.Errorf("Dienst: %w", isFehler{}), true, 0, 0},
		{"As liefert den gültigen Fehler", asFehler{f: gueltig}, false, model.CodeE6010, 412},
		{"As liefert einen nil-Zeiger", asFehler{f: nil}, false, model.CodeE9001, 500},
		{"As verneint", asFehler{nein: true}, false, model.CodeE9001, 500},
	}
	for _, f := range faelle {
		for _, m := range methodenSatz() {
			p := neuProbe(t, "record", fehlerMit(f.err))
			jw := neuJWriter(p.j)
			abbruch := p.ruf(jw, jRequest(context.Background(), m, "/p"))
			var soll []model.Code
			if f.code != 0 {
				soll = []model.Code{f.code}
			}
			gleich(t, f.name+" "+m+": Zähler", zaehlerFolge(p.tel.FakeTelemetry), soll)
			if abbruch != f.abbruch || jw.status != f.status {
				t.Errorf("%s %s: Abbruch %v, Status %d; erwartet %v und %d", f.name, m, abbruch, jw.status, f.abbruch, f.status)
			}
		}
	}
}

// TestFehlerMitAntwort hält SPEC-136 Nr. 7 b mit SPEC-294 Nr. 3 b und 5: liefert der Dienst zugleich einen Fehler und eine Antwort mit Body
// (nicht nil), gilt der Fehler; der Adapter schließt den Body genau einmal und liest ihn nie, auch wenn Close einen Fehler meldet; der Span
// endet genau einmal (beim Fehlerobjekt in der Folge CountError, Set, End, WriteHeader, beim Abbruch ohne Set und ohne WriteHeader), und
// gezählt wird allein das Fehlerobjekt, nie der Body. 8 Fehlerarten × 6 Formen der Antwort (Status 0, 200 und 600, je mit Close ohne und
// mit Fehler) × 14 Methoden gleich 672 Fälle.
func TestFehlerMitAntwort(t *testing.T) {
	var nz *inbound.Fehler
	gueltig := &inbound.Fehler{Code: model.CodeE6008, Status: 404}
	arten := []struct {
		name    string
		err     error
		abbruch bool
		code    model.Code
		status  int
	}{
		{"Fehlerobjekt", gueltig, false, model.CodeE6008, 404},
		{"Fehler ohne Code", errors.New("roh"), false, model.CodeE9001, 500},
		{"ErrUngueltig", model.ErrUngueltig, false, model.CodeE9001, 500},
		{"nil-Zeiger", nz, false, model.CodeE9001, 500},
		{"Bruch Status 0", &inbound.Fehler{Code: model.CodeE6008, Status: 0}, false, model.CodeE9001, 500},
		{"Canceled", context.Canceled, true, 0, 0},
		{"DeadlineExceeded", context.DeadlineExceeded, true, 0, 0},
		{"Canceled neben Fehler", errors.Join(gueltig, context.Canceled), true, 0, 0},
	}
	stati := []int{0, 200, 600}
	closeFehler := []error{nil, errors.New("Close")}
	if len(arten) != 8 || len(stati)*len(closeFehler) != 6 {
		t.Fatalf("%d Fehlerarten und %d Formen der Antwort, erwartet 8 und 6", len(arten), len(stati)*len(closeFehler))
	}
	n := 0
	for _, a := range arten {
		for _, s := range stati {
			for _, ce := range closeFehler {
				for _, m := range methodenSatz() {
					n++
					k := ausBytes([]byte("abc"))
					k.closeErr = ce
					p := neuProbe(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
						return inbound.Antwort{Status: s, Body: k}, a.err
					})
					k.beiClose = func() { p.j.add("Close") }
					jw := neuJWriter(p.j)
					abbruch := p.ruf(jw, jRequest(context.Background(), m, "/p"))
					name := fmt.Sprintf("%s mit Antwort %d, Close %v, %s", a.name, s, ce, m)
					if closes, nachClose, _, reads := k.stand(); closes != 1 || nachClose != 0 || reads != 0 {
						t.Errorf("%s: Close %d-mal, Read nach Close %d, Read %d; erwartet 1, 0, 0", name, closes, nachClose, reads)
					}
					var folge, soll []string
					for _, e := range p.j.folge() {
						if e != "Close" {
							folge = append(folge, e)
						}
					}
					soll = []string{"KorrelationsID", "StartSpan", "Verarbeiten", "End"}
					var codes []model.Code
					if !a.abbruch {
						soll = []string{"KorrelationsID", "StartSpan", "Verarbeiten", "CountError:" + a.code.String(), "Set", "End", "WriteHeader:" + strconv.Itoa(a.status), "Write"}
						if m == "HEAD" {
							soll = soll[:len(soll)-1]
						}
						codes = []model.Code{a.code}
					}
					gleich(t, name+": Folge", folge, soll)
					gleich(t, name+": Zähler", zaehlerFolge(p.tel.FakeTelemetry), codes)
					if i, v := slices.Index(p.j.folge(), "Close"), slices.Index(p.j.folge(), "Verarbeiten"); i < v || v < 0 {
						t.Errorf("%s: Close steht bei %d, Verarbeiten bei %d; erwartet Close nach Verarbeiten", name, i, v)
					}
					if abbruch != a.abbruch || jw.status != a.status {
						t.Errorf("%s: Abbruch %v, Status %d; erwartet %v und %d", name, abbruch, jw.status, a.abbruch, a.status)
					}
					if b := buecher(p.tel.FakeTelemetry); len(b) != 1 || b[0].ends != 1 {
						t.Errorf("%s: Buch %+v, erwartet 1 Span mit 1 End", name, b)
					}
				}
			}
		}
	}
	if n != 8*6*14 {
		t.Errorf("%d Fälle, erwartet 672", n)
	}
}

// TestPaketNenntOperationenJeEinmal hält G11 (SPEC-294 Nr. 5, 6): im Quelltext des Pakets (ohne _test.go) steht je ein Aufruf von
// NeuesFehlerobjekt, CountError, StartSpan, Set (am Span) und End, keiner von Log; das Set mit einem Namen als erstem Argument ist der Header.
func TestPaketNenntOperationenJeEinmal(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("Quelltext: %v, %d Dateien", err, len(dateien))
	}
	zaehler := map[string]int{}
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := ce.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if lit, ok := firstArg(ce).(*ast.BasicLit); sel.Sel.Name == "Set" && ok && lit.Kind == token.STRING {
				return true
			}
			zaehler[sel.Sel.Name]++
			return true
		})
	}
	soll := map[string]int{"NeuesFehlerobjekt": 1, "CountError": 1, "StartSpan": 1, "Set": 1, "End": 1, "Log": 0}
	for name, n := range soll {
		if zaehler[name] != n {
			t.Errorf("%s im Quelltext %d-mal, erwartet %d", name, zaehler[name], n)
		}
	}
}

// TestJedeSchreibstelleBeendetDenSpanVorWriteHeader hält G4 gegen die Quelle (SPEC-294 Nr. 3 b): jeder Aufruf von WriteHeader im Quelltext des
// Pakets steht unmittelbar hinter einem Aufruf von vorKopf (Set und End), und die Funktionen mit einem solchen Aufruf sind genau die Stellen,
// denen die Wege von G4 zugeordnet sind: schreibeAntwort (Antwort des Dienstes) und schreibeFehlerobjekt (Fehlerobjekt, Antwort 503, Bruch).
func TestJedeSchreibstelleBeendetDenSpanVorWriteHeader(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("Quelltext: %v, %d Dateien", err, len(dateien))
	}
	stellen := map[string]int{}
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				b, ok := n.(*ast.BlockStmt)
				if !ok {
					return true
				}
				for i, st := range b.List {
					if aufrufName(st) != "WriteHeader" {
						continue
					}
					stellen[fn.Name.Name]++
					if i == 0 || aufrufName(b.List[i-1]) != "vorKopf" {
						t.Errorf("%s: WriteHeader steht nicht unmittelbar hinter vorKopf", fn.Name.Name)
					}
				}
				return true
			})
		}
	}
	gleich(t, "Funktionen mit WriteHeader", stellen, map[string]int{"schreibeAntwort": 1, "schreibeFehlerobjekt": 1})
}

// aufrufName liefert den Namen der Methode, wenn die Anweisung ein Aufruf x.Name(...) ist, sonst den leeren Text.
func aufrufName(st ast.Stmt) string {
	es, ok := st.(*ast.ExprStmt)
	if !ok {
		return ""
	}
	ce, ok := es.X.(*ast.CallExpr)
	if !ok {
		return ""
	}
	if sel, ok := ce.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name
	}
	return ""
}

func firstArg(ce *ast.CallExpr) ast.Expr {
	if len(ce.Args) == 0 {
		return nil
	}
	return ce.Args[0]
}

// TestVerarbeitenErhaeltDasCtxDesSpans hält G12 (SPEC-294 Nr. 3 d, e): StartSpan erhält r.Context(); der Dienst erhält das gelieferte ctx,
// das r.Context() ableitet (Wert bleibt, Span des Gesamtspans darin, keine eigene Frist: ohne Frist des Requests hat das ctx keine, mit einer
// ist es dieselbe); der Gesamtspan hat keine Eltern, ein Span des Dienstes hat ihn als Eltern. Beide Formen des Requests (ohne und mit Frist).
func TestVerarbeitenErhaeltDasCtxDesSpans(t *testing.T) {
	type schluessel struct{}
	frist := time.Now().Add(3 * time.Hour)
	ohne := context.WithValue(context.Background(), schluessel{}, "wert")
	mit, aufheben := context.WithDeadline(ohne, frist)
	defer aufheben()
	for _, f := range []struct {
		name  string
		mit   bool
		frist time.Time
		da    bool
	}{{"ohne Frist", false, time.Time{}, false}, {"mit Frist", true, frist, true}} {
		eltern := ohne
		if f.mit {
			eltern = mit
		}
		var p *probe
		gesehen := false
		var dl time.Time
		var dlDa bool
		p = neuProbe(t, "record", func(ctx context.Context, _ inbound.Anfrage) (inbound.Antwort, error) {
			gesehen = ctx.Value(schluessel{}) == "wert"
			dl, dlDa = ctx.Deadline()
			_, kind := p.tel.StartSpan(ctx, support.SpanMockQuery)
			kind.End()
			return inbound.Antwort{Status: 200, Body: leer()}, nil
		})
		r := jRequest(eltern, "GET", "/p")
		p.ruf(neuJWriter(p.j), r)
		aufrufe := p.tel.Aufrufe()
		var start []supporttest.Aufruf
		for _, a := range aufrufe {
			if a.Op == supporttest.OpStartSpan {
				start = append(start, a)
			}
		}
		if len(start) != 2 {
			t.Fatalf("%s: %d StartSpan, erwartet 2", f.name, len(start))
		}
		if start[0].Ctx() != eltern || start[0].Eltern != nil {
			t.Errorf("%s: StartSpan des Gesamtspans erhielt nicht r.Context() oder trägt Eltern", f.name)
		}
		if !gesehen || start[1].Eltern != start[0].Span || start[1].Eltern == nil {
			t.Errorf("%s: Dienst: Wert aus r.Context() da %v, Eltern des Kindes %v, erwartet das ctx des Gesamtspans", f.name, gesehen, start[1].Eltern)
		}
		if dlDa != f.da || !dl.Equal(f.frist) {
			t.Errorf("%s: Frist des ctx des Dienstes %v (da %v), erwartet %v (da %v)", f.name, dl, dlDa, f.frist, f.da)
		}
	}
}

// TestCtxDesSpansEndetMitAbbruchUndSchliessen hält G12 (SPEC-294 Nr. 3 d, SPEC-136 Nr. 4g, 14): das ctx, das der Dienst erhält, endet mit dem
// Abbruch des Clients und mit dem Ablauf der Frist von Schliessen.
func TestCtxDesSpansEndetMitAbbruchUndSchliessen(t *testing.T) {
	for _, art := range []string{"Abbruch des Clients", "Schliessen"} {
		begonnen, ende := make(chan struct{}), make(chan error, 1)
		u := neuUmgebung(t, "record", func(ctx context.Context, _ inbound.Anfrage) (inbound.Antwort, error) {
			close(begonnen)
			select {
			case <-ctx.Done():
				ende <- ctx.Err()
			case <-time.After(5 * time.Second):
				ende <- errors.New("ctx nicht beendet")
			}
			return inbound.Antwort{}, context.Canceled
		})
		cctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequestWithContext(cctx, "GET", "http://"+u.addr+"/p", nil)
		fertig := make(chan struct{})
		go func() {
			defer close(fertig)
			if resp, err := u.client().Do(req); err == nil {
				_ = resp.Body.Close()
			}
		}()
		<-begonnen
		if art == "Abbruch des Clients" {
			cancel()
		} else {
			_ = u.srv.Schliessen(20 * time.Millisecond)
		}
		select {
		case err := <-ende:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("%s: ctx.Err() = %v, erwartet context.Canceled", art, err)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("%s: der Dienst kehrte nicht zurück", art)
		}
		cancel()
		<-fertig
	}
}

// stelleDesRequests ist eine der 8 Stellen von SPEC-294 Nr. 8, an der der Test den Wert setzt.
type stelleDesRequests struct {
	name  string
	setze func(r *http.Request, k string)
}

// stellenDesRequests liefert die 8 Stellen; die Stelle "Text der Ursache" liegt nicht im Request, sondern im Fehler des Dienstes.
func stellenDesRequests() []stelleDesRequests {
	return []stelleDesRequests{
		{"Pfad", func(r *http.Request, k string) { r.RequestURI, r.URL.Path = "/"+k, "/"+k }},
		{"Name der Query", func(r *http.Request, k string) { r.RequestURI, r.URL.RawQuery = r.RequestURI+"?"+k+"=1", k+"=1" }},
		{"Wert der Query", func(r *http.Request, k string) { r.RequestURI, r.URL.RawQuery = r.RequestURI+"?x="+k, "x="+k }},
		{"Name eines Headers", func(r *http.Request, k string) { r.Header.Set("X-"+k, "1") }},
		{"Wert eines Headers", func(r *http.Request, k string) { r.Header.Set("X-Wert", k) }},
		{"Body", func(r *http.Request, k string) {
			r.Body, r.ContentLength = io.NopCloser(strings.NewReader(k)), int64(len(k))
		}},
		{"Host", func(r *http.Request, k string) { r.Host = k }},
		{"Text der Ursache", func(*http.Request, string) {}},
	}
}

// TestKeinRequestWertImGesamtspan hält G13 (SPEC-294 Nr. 8): ein Wert an einer der 8 Stellen des Requests erreicht in keiner der 3 Arten der
// Antwort (Antwort des Dienstes, Fehlerobjekt, 503) ein Attribut, das der Fake sah, und keinen Aufruf von Log; der Wert steht in zwei Formen
// (Token, Methode). 8 Stellen × 3 Arten × 2 Formen, 24 Fälle in zwei Formen; die Stelle der Ursache führt beim Fehlerobjekt zwei Formen des
// Fehlers (`*inbound.Fehler` mit Ursache, Fehler ohne Code mit dem Text als Meldung).
func TestKeinRequestWertImGesamtspan(t *testing.T) {
	stellen := stellenDesRequests()
	arten := []string{"Antwort", "Fehlerobjekt", "503"}
	if len(stellen) != 8 || len(arten) != 3 {
		t.Fatalf("%d Stellen und %d Arten, erwartet 8 und 3", len(stellen), len(arten))
	}
	n := 0
	for _, s := range stellen {
		for _, art := range arten {
			n++
			for _, gross := range []bool{false, true} {
				k := kanarie(gross)
				ursache := ""
				if s.name == "Text der Ursache" {
					ursache = k
				}
				tu := antwortMit(200, leer())
				if art == "Fehlerobjekt" {
					tu = fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 404, Ursache: errors.New(ursache)})
				}
				p := neuProbe(t, "record", tu)
				if art == "503" {
					p.srv.Anhalten()
				}
				r := jRequest(context.Background(), "GET", "/p")
				s.setze(r, k)
				p.ruf(neuJWriter(p.j), r)
				for _, b := range buecher(p.tel.FakeTelemetry) {
					for _, a := range append(slices.Clone(b.start), slices.Concat(b.sets...)...) {
						if strings.Contains(strings.ToLower(a.Text), kanarie(false)) {
							t.Errorf("%s %s %v: das Attribut %s trägt den Wert des Requests: %q", s.name, art, gross, a.Key, a.Text)
						}
					}
				}
				if p.tel.Anzahl(supporttest.OpLog) != 0 {
					t.Errorf("%s %s: der Adapter rief Log", s.name, art)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Fälle, erwartet 24", n)
	}
}

// TestKeinFehlertextImGesamtspan hält G13 (SPEC-294 Nr. 8) für die zweite Form des Fehlers der Stelle der Ursache: ein Fehler ohne Code,
// dessen Meldung der Wert ist (`err.Error()` trägt ihn), erreicht kein Attribut des Fakes; beide Formen des Werts.
func TestKeinFehlertextImGesamtspan(t *testing.T) {
	for _, gross := range []bool{false, true} {
		p := neuProbe(t, "record", fehlerMit(errors.New(kanarie(gross))))
		p.ruf(neuJWriter(p.j), jRequest(context.Background(), "GET", "/p"))
		b := buecher(p.tel.FakeTelemetry)
		if len(b) != 1 {
			t.Fatalf("Buch %+v, erwartet 1 Span", b)
		}
		for _, a := range append(slices.Clone(b[0].start), slices.Concat(b[0].sets...)...) {
			if strings.Contains(strings.ToLower(a.Text), kanarie(false)) {
				t.Errorf("%v: das Attribut %s trägt den Text des Fehlers: %q", gross, a.Key, a.Text)
			}
		}
	}
}
