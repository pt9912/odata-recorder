package admin_test

import (
	"context"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// importsZiel ist der Pfad der Routen für Imports (SPEC-297 Nr. 1).
const importsZiel = "/__admin/imports"

// impOp nennt die zwei Operationen von inbound.ImportUseCase in der Folge des Ports.
type impOp int

const (
	impImport impOp = iota
	impGet
	impAnzahl
)

// Die Fehlerobjekte der Codes, die ein Dienst für Imports liefert (SPEC-109 Nr. 2, SPEC-240), als Literale des Vertrags;
// E5005, E7003, E7009 und E3001 stehen in sessions_fake_test.go.
const (
	objektE5004 = `{"error":{"code":"ODR-E5004","message":"Assignment belongs to another service","target":"replay"}}`
	objektE6004 = `{"error":{"code":"ODR-E6004","message":"Metadata invalid or unsupported","target":"import"}}`
	objektE6101 = `{"error":{"code":"ODR-E6101","message":"Baseline import invalid or inconsistent","target":"import"}}`
	objektE6102 = `{"error":{"code":"ODR-E6102","message":"Import id reused with different content","target":"import"}}`
	objektE7005 = `{"error":{"code":"ODR-E7005","message":"Import unknown","target":"import"}}`
)

// reportA ist der Report des Dienstes in den Fällen, die den Inhalt nicht prüfen.
const reportA = `{"import_id":"imp-dienst","status":"ok"}`

// importA ist ein Import des Dienstes: jedes Feld hat einen eigenen Wert.
func importA() model.Import {
	return model.Import{
		ID: "imp-dienst", RequestHash: "rh-dienst", BaselineSnapshotID: "snap-dienst", ServiceID: "svc-dienst",
		CreatedAt: zeitA(), ReportJSON: reportA,
	}
}

// importsFake erfüllt inbound.ImportUseCase: jede der zwei Operationen zählt ihren Aufruf und hält ctx und jedes Argument
// fest (die Anfrage mit einer Kopie der Liste, die Kennung), in der Folge der Aufrufe; Ergebnis und Fehler sind einstellbar
// (ergebnis, imp, err; importiere und liest ersetzen sie). Der Fake gerät nie in Panik und ist für gleichzeitigen Aufruf
// sicher; die Einstellungen setzt der Test vor dem ersten Aufruf.
type importsFake struct {
	mu       sync.Mutex
	ops      []impOp
	ctxs     []context.Context
	anfragen []inbound.ImportAnfrage
	ids      []string
	zaehler  [impAnzahl]int

	ergebnis   inbound.ImportErgebnis
	imp        model.Import
	err        error
	importiere func(ctx context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error)
	liest      func(ctx context.Context, id string) (model.Import, error)
}

func (f *importsFake) notiere(ctx context.Context, op impOp) {
	f.ops = append(f.ops, op)
	f.ctxs = append(f.ctxs, ctx)
	f.zaehler[op]++
}

func (f *importsFake) Import(ctx context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
	f.mu.Lock()
	f.notiere(ctx, impImport)
	a.SourceSessionIDs = slices.Clone(a.SourceSessionIDs)
	f.anfragen = append(f.anfragen, a)
	f.mu.Unlock()
	if f.importiere != nil {
		return f.importiere(ctx, a)
	}
	return f.ergebnis, f.err
}

func (f *importsFake) Get(ctx context.Context, id string) (model.Import, error) {
	f.mu.Lock()
	f.notiere(ctx, impGet)
	f.ids = append(f.ids, id)
	f.mu.Unlock()
	if f.liest != nil {
		return f.liest(ctx, id)
	}
	return f.imp, f.err
}

// zahlen liefert die Zahl der Aufrufe je Operation.
func (f *importsFake) zahlen() [impAnzahl]int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler
}

// gesamt liefert die Zahl aller Aufrufe.
func (f *importsFake) gesamt() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.ops)
}

// anfrageListe liefert die Anfragen der Aufrufe von Import, kennungen die Kennungen der Aufrufe von Get, ctxListe den ctx
// jedes Aufrufs; alle als Kopie.
func (f *importsFake) anfrageListe() []inbound.ImportAnfrage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.anfragen)
}

func (f *importsFake) kennungen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ids)
}

func (f *importsFake) ctxListe() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

// nurDiagonale meldet jeden Zähler ungleich 1 bei der Operation op und ungleich 0 bei der anderen.
func (f *importsFake) nurDiagonale(t *testing.T, name string, op impOp) {
	t.Helper()
	for o, n := range f.zahlen() {
		soll := 0
		if impOp(o) == op {
			soll = 1
		}
		if n != soll {
			t.Errorf("%s: Operation %d wurde %d-mal gerufen, erwartet %d", name, o, n, soll)
		}
	}
}

// gleicheAnfragen meldet, ob zwei Listen von Anfragen in jedem Feld und in jedem Eintrag der Liste gleich sind.
func gleicheAnfragen(a, b []inbound.ImportAnfrage) bool {
	return slices.EqualFunc(a, b, func(x, y inbound.ImportAnfrage) bool {
		return x.ImportID == y.ImportID && x.ServiceID == y.ServiceID && x.MetadataHash == y.MetadataHash &&
			x.ConflictPolicy == y.ConflictPolicy && x.Target == y.Target && slices.Equal(x.SourceSessionIDs, y.SourceSessionIDs)
	})
}

// importsUmgebung baut einen Server mit den Routen der Imports über dem Fake und dem FakeAudit der Umgebung; der Dienst
// liefert importA mit Neu wahr.
func importsUmgebung(t *testing.T) (*umgebung, *importsFake, []admin.Route) {
	t.Helper()
	k, u := konfigMit()
	f := &importsFake{ergebnis: inbound.ImportErgebnis{Import: importA(), Neu: true}, imp: importA()}
	routen, err := admin.ImportRouten(f, u.audit)
	if err != nil || len(routen) != 2 {
		t.Fatalf("ImportRouten: %d Routen, %v", len(routen), err)
	}
	k.Routen = routen
	return neuMitKonfig(t, k, u), f, routen
}

// importPost bildet `POST /__admin/imports`, query steht hinter dem `?` (leer: ohne), kopf setzt Header.
func importPost(query string, kopf map[string][]string, body io.Reader) *http.Request {
	ziel := importsZiel
	if query != "" {
		ziel += "?" + query
	}
	return anfrageMit("POST", ziel, kopf, body)
}

// importGet bildet `GET /__admin/imports/{id}`; id steht unverändert im Pfad, also kodiert.
func importGet(id string) *http.Request { return anfrageMit("GET", importsZiel+"/"+id, nil, nil) }

// gutBody ist ein gültiger Body von `POST /__admin/imports` mit den drei Pflichtschlüsseln; gutAnfrage ist, was der Dienst
// erhält.
const gutBody = `{"import_id":"imp-1","service_id":"svc-1","source_session_ids":["ses-1"]}`

func gutAnfrage() inbound.ImportAnfrage {
	return inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: "svc-1", SourceSessionIDs: []string{"ses-1"}}
}

// bodyLeser liefert die Bytes und danach den Fehler rest (nil: das Ende) und zählt jeden Lesezugriff.
type bodyLeser struct {
	daten []byte
	rest  error
	pos   int
	n     atomic.Int64
}

func (b *bodyLeser) Read(p []byte) (int, error) {
	b.n.Add(1)
	if b.pos < len(b.daten) {
		n := copy(p, b.daten[b.pos:])
		b.pos += n
		return n, nil
	}
	if b.rest != nil {
		return 0, b.rest
	}
	return 0, io.EOF
}

// objektAus fügt Glieder zu einem Objekt.
func objektAus(glieder ...string) string { return "{" + strings.Join(glieder, ",") + "}" }

// pflichtGlieder sind die drei Pflichtschlüssel mit s als Wert von `import_id`; optionaleGlieder die drei übrigen.
func pflichtGlieder(s string) []string {
	return []string{`"import_id":"` + s + `"`, `"service_id":"svc-1"`, `"source_session_ids":["ses-1"]`}
}

func optionaleGlieder() []string {
	return []string{`"metadata_hash":"sha256:h"`, `"conflict_policy":"reject"`, `"target":"new_baseline"`}
}

// alleGlieder sind die sechs Schlüssel mit s als Wert von `import_id`.
func alleGlieder(s string) []string { return append(pflichtGlieder(s), optionaleGlieder()...) }

// importUngueltigForm liefert die 28 Formen von IM7 (Form, Schlüssel, Zeichen und Rand der Achsen Wertform, Struktur,
// Schlüssel und Zeichen); s steht als Wert von `import_id` in jeder Form, die einen Wert nennt.
func importUngueltigForm(s string) []bForm {
	var out []bForm
	add := func(achse, name, body string) { out = append(out, bForm{achse: achse, name: name, body: []byte(body)}) }
	p := pflichtGlieder(s)
	voll := alleGlieder(s)
	for _, w := range [][2]string{{"leer", ""}, {"Leerraum", " \t\r\n"}, {"null", "null"}, {"Array", "[]"}, {"Zeichenkette", `"x"`}, {"Zahl", "1"}, {"Wahrheitswert", "true"}} {
		add("Wertform", w[0], w[1])
	}
	for _, w := range [][2]string{
		{"offen", "{"}, {"nicht geschlossen", "{" + strings.Join(p, ",")}, {"Komma vor }", "{" + strings.Join(p, ",") + ",}"},
		{"einfache Anführungszeichen", `{'import_id':'` + s + `','service_id':'svc-1','source_session_ids':['ses-1']}`},
		{"Text dahinter", objektAus(p...) + " x"}, {"Objekt dahinter", objektAus(p...) + "{}"},
		{"Doppelpunkt fehlt", `{"import_id" "` + s + `","service_id":"svc-1","source_session_ids":["ses-1"]}`},
	} {
		add("Struktur", w[0], w[1])
	}
	for _, k := range []string{"extra", "Import_Id", "idempotency_key"} {
		add("Schlüssel", "fremd "+k, objektAus(append(slices.Clone(voll), `"`+k+`":"x"`)...))
	}
	for _, g := range voll {
		add("Schlüssel", "doppelt gleich "+strings.SplitN(g, `"`, 3)[1], objektAus(append(slices.Clone(voll), g)...))
	}
	add("Schlüssel", "doppelt anders", objektAus(append(slices.Clone(voll), `"import_id":"anders"`)...))
	add("Schlüssel", "doppelt über Escape", objektAus(append(slices.Clone(voll), `"import`+esc("005f")+`id":"`+s+`"`)...))
	bomBytes := string([]byte{0xEF, 0xBB, 0xBF})
	add("Zeichen", "BOM", bomBytes+objektAus(p...))
	add("Zeichen", "ungültiges UTF-8", objektAus(`"import_id":"`+s+string([]byte{0xFF})+`"`, p[1], p[2]))
	add("Zeichen", "rohes NUL", objektAus(`"import_id":"`+s+string([]byte{0x00})+`"`, p[1], p[2]))
	return out
}

// importRand liefert den Rand von IM7: jeder Bytewert 0 bis 255 vor und hinter einem gültigen Objekt, 512 Formen; gültig sind
// genau die vier Zeichen des Leerraums.
func importRand() []bForm {
	var out []bForm
	obj := []byte(gutBody)
	for b := range 256 {
		name := "Byte " + strconv.Itoa(b)
		out = append(out,
			bForm{achse: "Rand", name: name + " davor", body: append([]byte{byte(b)}, obj...)},
			bForm{achse: "Rand", name: name + " dahinter", body: append(slices.Clone(obj), byte(b))})
	}
	return out
}

// bodyMitGlied liefert ein Objekt aus den drei Pflichtgliedern (s als `import_id`), in dem das Glied des Schlüssels den
// Wert literal trägt; ein optionaler Schlüssel wird angehängt.
func bodyMitGlied(s, schluessel, literal string) string {
	glieder := pflichtGlieder(s)
	neu := `"` + schluessel + `":` + literal
	for i, g := range glieder {
		if strings.HasPrefix(g, `"`+schluessel+`"`) {
			glieder[i] = neu
			return objektAus(glieder...)
		}
	}
	return objektAus(append(glieder, neu)...)
}

// importUngueltigRest liefert die 58 Formen von IM8 (Pflicht, Typen, Lesen und die 15 Paare).
func importUngueltigRest(s string) []bForm {
	var out []bForm
	add := func(achse, name, body string) { out = append(out, bForm{achse: achse, name: name, body: []byte(body)}) }
	p := pflichtGlieder(s)
	add("Pflicht", "leeres Objekt", "{}")
	for i, k := range []string{"import_id", "service_id", "source_session_ids"} {
		add("Pflicht", "ohne "+k, objektAus(slices.Delete(slices.Clone(p), i, i+1)...))
	}
	nichtText := []string{"null", "1", "true", "[]", "{}"}
	for _, k := range []string{"import_id", "service_id"} {
		for _, v := range nichtText {
			add("Werte", k+" "+v, bodyMitGlied(s, k, v))
		}
	}
	for _, v := range []string{"null", `"a"`, "1", "true", "{}"} {
		add("Werte", "source_session_ids "+v, bodyMitGlied(s, "source_session_ids", v))
	}
	for _, v := range []string{"[null]", "[1]", "[true]", "[[]]", "[{}]", `["a",1]`, `[1,"a"]`} {
		add("Werte", "Elemente "+v, bodyMitGlied(s, "source_session_ids", v))
	}
	for _, k := range []string{"metadata_hash", "conflict_policy", "target"} {
		for _, v := range nichtText {
			add("Werte", k+" "+v, bodyMitGlied(s, k, v))
		}
	}
	out = append(out,
		bForm{achse: "Lesen", name: "1048577 Bytes", body: auffuellen(objektAus(p...), bodyGrenze+1)},
		bForm{achse: "Lesen", name: "kürzer als Content-Length", body: []byte(`{"import_id`), rest: io.ErrUnexpectedEOF, laenge: 100})
	return append(out, importPaare(s)...)
}

// importPaare liefert die 15 Paare aus den sechs Verstoßarten Lesen, Form, Schlüssel, Pflicht, Typ und Zeichen: jeder Body
// trägt genau zwei Arten.
func importPaare(s string) []bForm {
	arten := []string{"Lesen", "Form", "Schlüssel", "Pflicht", "Typ", "Zeichen"}
	var out []bForm
	for i := range arten {
		for j := i + 1; j < len(arten); j++ {
			an := func(a string) bool { return arten[i] == a || arten[j] == a }
			glieder := pflichtGlieder(s)
			f := bForm{achse: "Paar", name: arten[i] + "+" + arten[j]}
			if an("Zeichen") {
				glieder[0] = `"import_id":"` + s + string([]byte{0x00}) + `"`
			}
			if an("Pflicht") {
				glieder = glieder[:2]
			}
			if an("Schlüssel") {
				glieder = append(glieder, `"extra":"x"`)
			}
			if an("Typ") {
				glieder = append(glieder, `"metadata_hash":1`)
			}
			body := objektAus(glieder...)
			if an("Form") {
				body += " x"
			}
			if an("Lesen") {
				f.rest, f.laenge = io.ErrUnexpectedEOF, 1000
			}
			f.body = []byte(body)
			out = append(out, f)
		}
	}
	return out
}

// importBodyUngueltig sind die 86 ungültigen Bodys von IM7 und IM8 ohne den Rand (28 + 58); Teil 2 liest sie in IM21.
func importBodyUngueltig(s string) []bForm {
	return append(importUngueltigForm(s), importUngueltigRest(s)...)
}

// reportForm ist ein Report des Dienstes mit seinem Namen; roh sind seine Bytes.
type reportForm struct {
	name string
	roh  []byte
}

// reportsGueltig sind die 15 gültigen Formen von M-Report (SPEC-297 Nr. 6); jede Escape-Folge entsteht aus Bytewerten.
func reportsGueltig() []reportForm {
	ae, euro, smiley, _ := nichtASCII()
	ls, ps := string([]byte{0xE2, 0x80, 0xA8}), string([]byte{0xE2, 0x80, 0xA9})
	return []reportForm{
		{"kompakt", []byte(`{"import_id":"imp-1","coverage":[1,2],"ok":true}`)},
		{"Leerraum zwischen den Token", []byte("{ \"a\" :\r\n 1 ,\n\t\"b\" : [ 1 , 2 ] }")},
		{"Leerraum am Rand", []byte(" \n\t{\"a\":1}\r\n ")},
		{"spitze Klammern und Und", []byte(`{"a":"<b>&</b>"}`)},
		{"U+2028 und U+2029 roh", []byte(`{"a":"` + ls + ps + `"}`)},
		{"Escape für < und U+2028", []byte(`{"a":"` + esc("003c") + esc("2028") + `"}`)},
		{"Mehrbyte roh", []byte(`{"a":"` + ae + euro + smiley + `"}`)},
		{"Mitglieder absteigend", []byte(`{"z":1,"b":2,"a":3}`)},
		{"doppeltes Mitglied", []byte(`{"a":1,"a":2}`)},
		{"Zahlen", []byte(`{"a":1e400,"b":-0,"c":1.0}`)},
		{"Schrägstrich-Escape", []byte(`{"a":"` + rueck() + `/"}`)},
		{"null", []byte("null")},
		{"0", []byte("0")},
		{"leeres Array", []byte("[]")},
		{"Zeichenkette", []byte(`"x"`)},
	}
}

// reportsLaengen sind die sieben Längen von M-Report: unter, auf und über 64 KiB, 1 MiB und 4 MiB.
func reportsLaengen() []int { return []int{2, 1000, 65535, 65536, 1048576, 1048577, 4194305} }

// reportDerLaenge liefert eine Zeichenkette als JSON-Wert von genau n Bytes, reihum aus einem Alphabet gebildet.
func reportDerLaenge(n int) []byte { return []byte(`"` + wertDerLaenge("klein", n-2) + `"`) }

// reportsUngueltig sind die 7 ungültigen Formen von M-Report (SPEC-297 Nr. 6); s steht in jeder Form, die einen Wert nennt.
func reportsUngueltig(s string) []reportForm {
	return []reportForm{
		{"leer", nil},
		{"abgeschnitten", []byte(`{"a":"` + s + `"`)},
		{"Text dahinter", []byte(`{"a":"` + s + `"} x`)},
		{"zwei Werte", []byte(`{"a":"` + s + `"}{}`)},
		{"rohes 0xFF", []byte(`{"a":"` + s + string([]byte{0xFF}) + `"}`)},
		{"BOM davor", []byte(string([]byte{0xEF, 0xBB, 0xBF}) + `{"a":"` + s + `"}`)},
		{"rohes NUL", []byte(`{"a":"` + s + string([]byte{0x00}) + `"}`)},
	}
}
