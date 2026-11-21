package admin_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// bodyGrenze ist die Grenze des Bodys in Bytes (SPEC-238 Nr. 10, SPEC-292 Nr. 5 a).
const bodyGrenze = 1048576

// rueck liefert den Backslash aus seinem Bytewert (AGENTS.md §3.18: kein Escape im Quelltext des Tests, den das
// Schreibwerkzeug umsetzen könnte).
func rueck() string { return string([]byte{0x5c}) }

// esc liefert die JSON-Escape-Folge `\uXXXX` aus Bytewerten.
func esc(hex string) string { return string([]byte{0x5c, 0x75}) + hex }

// bForm ist eine Form des Bodys von `POST /__admin/sessions`: die Bytes, die Header, der Fehler, den der Leser hinter den
// Bytes liefert (nil: Ende), die angekündigte Länge (0: die der Bytes) und bei einer gültigen Form das Orakel der
// Argumente von Create.
type bForm struct {
	achse, name string
	body        []byte
	kopf        map[string][]string
	rest        error
	laenge      int64
	svc, id     string
}

// leser liefert den Body der Form als Leser, hinter den Bytes den Fehler rest oder das Ende.
func (b bForm) leser() io.Reader {
	if b.rest == nil {
		return bytes.NewReader(b.body)
	}
	return io.MultiReader(bytes.NewReader(b.body), errLeser{b.rest})
}

// request bildet den Request der Form.
func (b bForm) request() *http.Request {
	r := sessionsReq("POST", "", b.kopf, b.leser())
	if b.laenge != 0 {
		r.ContentLength = b.laenge
	}
	return r
}

// errLeser liefert bei jedem Lesen den Fehler.
type errLeser struct{ err error }

func (e errLeser) Read([]byte) (int, error) { return 0, e.err }

// auffuellen liefert body, mit Leerzeichen auf n Bytes gefüllt.
func auffuellen(body string, n int) []byte {
	return append([]byte(body), bytes.Repeat([]byte{' '}, n-len(body))...)
}

// nichtASCII liefert vier Zeichen außerhalb von ASCII aus Bytewerten: ä, €, ein Zeichen außerhalb der Basisebene, ß.
func nichtASCII() (ae, euro, smiley, ss string) {
	return string([]byte{0xC3, 0xA4}), string([]byte{0xE2, 0x82, 0xAC}), string([]byte{0xF0, 0x9F, 0x98, 0x80}),
		string([]byte{0xC3, 0x9F})
}

// bGueltig sind die 11 gültigen Formen von M-Body (SPEC-292 Nr. 5 a bis f): 8 Formen mit `Content-Type: application/json`
// und die erste Form mit keinem und zwei fremden `Content-Type`; das Orakel sind die Argumente von Create als Literale.
func bGueltig() []bForm {
	json := map[string][]string{"Content-Type": {"application/json"}}
	ae, euro, smiley, ss := nichtASCII()
	formen := []bForm{
		{name: "nur service_id", body: []byte(`{"service_id":"svc"}`), svc: "svc", id: ""},
		{name: "mit id", body: []byte(`{"service_id":"Svc-1","id":"abc"}`), svc: "Svc-1", id: "abc"},
		{name: "id zuerst", body: []byte(`{"id":"abc","service_id":"svc"}`), svc: "svc", id: "abc"},
		{name: "id leer", body: []byte(`{"service_id":"svc","id":""}`), svc: "svc", id: ""},
		{name: "Leerraum", body: []byte("\r\n\t " + `{ "service_id" : " svc " ,` + "\n" + ` "id" : " a " }` + " \t\r\n"), svc: " svc ", id: " a "},
		{
			name: "Escapes",
			body: []byte(`{"service_id":"a` + esc("0076") + rueck() + `nb","id":"x` + rueck() + `"y` + rueck() + rueck() + `z"}`),
			svc:  "av\nb", id: `x"y` + rueck() + "z",
		},
		{
			name: "Nicht-ASCII", body: []byte(`{"service_id":"` + ae + euro + smiley + `","id":"` + ss + `"}`),
			svc: ae + euro + smiley, id: ss,
		},
		{name: "1048576 Bytes", body: auffuellen(`{"service_id":"svc"}`, bodyGrenze), svc: "svc", id: ""},
	}
	for i := range formen {
		formen[i].achse, formen[i].kopf = "gültig", json
	}
	for _, ct := range [][]string{nil, {"text/plain"}, {"application/x-www-form-urlencoded"}} {
		f := bForm{achse: "Content-Type", name: "Content-Type " + strings.Join(ct, ""), body: formen[0].body, svc: "svc"}
		if ct != nil {
			f.kopf = map[string][]string{"Content-Type": ct}
		}
		formen = append(formen, f)
	}
	return formen
}

// bUngueltig sind die 48 ungültigen Formen von M-Body (SPEC-292 Nr. 5 a bis d, g) in sieben Achsen; s steht anstelle von
// `svc` in jeder Form, die einen Wert nennt.
func bUngueltig(s string) []bForm {
	svc := `"service_id":"` + s + `"`
	var out []bForm
	add := func(achse, name string, body string) {
		out = append(out, bForm{achse: achse, name: name, body: []byte(body)})
	}
	for name, body := range map[string]string{
		"leer": "", "Leerraum": " \t\r\n", "null": "null", "Array": "[]", "Zeichenkette": `"x"`, "Zahl": "1", "Wahrheitswert": "true",
	} {
		add("Wertform", name, body)
	}
	for name, body := range map[string]string{
		"offen": "{", "nicht geschlossen": "{" + svc, "Komma vor }": "{" + svc + ",}", "einfache Anführungszeichen": `{'service_id':'` + s + `'}`,
		"Text dahinter": "{" + svc + "} x", "Objekt dahinter": "{" + svc + "}{}", "Doppelpunkt fehlt": `{"service_id" "` + s + `"}`,
	} {
		add("Struktur", name, body)
	}
	for name, body := range map[string]string{
		"leeres Objekt": "{}", "nur id": `{"id":"` + s + `"}`, "fremder Schlüssel": "{" + svc + `,"extra":"x"}`,
		"Schreibweise": `{"Service_Id":"` + s + `"}`, "doppelt gleich": "{" + svc + "," + svc + "}",
		"doppelt anders": "{" + svc + `,"service_id":"b"}`, "id doppelt": "{" + svc + `,"id":"a","id":"a"}`,
		"doppelt über Escape": "{" + svc + `,"service` + esc("005f") + `id":"` + s + `"}`,
	} {
		add("Schlüssel", name, body)
	}
	for name, wert := range map[string]string{
		"BOM": string([]byte{0xEF, 0xBB, 0xBF}) + "{" + svc + "}", "ungültiges UTF-8": "{" + svc[:len(svc)-1] + string([]byte{0xFF}) + `"}`,
		"rohes NUL": "{" + svc[:len(svc)-1] + string([]byte{0x00}) + `"}`,
	} {
		add("Zeichen", name, wert)
	}
	for _, v := range []string{`""`, "null", "1", "true", "[]", "{}"} {
		add("Werte", "service_id "+v, `{"service_id":`+v+"}")
	}
	for _, v := range []string{"null", "1", "true", "[]", "{}"} {
		add("Werte", "id "+v, "{"+svc+`,"id":`+v+"}")
	}
	out = append(out,
		bForm{achse: "Lesen", name: "1048577 Bytes", body: auffuellen("{"+svc+"}", bodyGrenze+1)},
		bForm{achse: "Lesen", name: "kürzer als Content-Length", body: []byte(`{"service_id`), rest: io.ErrUnexpectedEOF, laenge: 100})
	return append(out, bPaare(s)...)
}

// bPaare sind die 10 Paare aus den fünf Verstoßarten unbekannter Schlüssel, Wert von service_id, Wert von id, doppelter
// Schlüssel und Zeichen: jeder Body trägt genau zwei Arten (SPEC-292 Nr. 5 g).
func bPaare(s string) []bForm {
	arten := []string{"Schlüssel", "service_id", "id", "doppelt", "Zeichen"}
	var out []bForm
	for i := range arten {
		for j := i + 1; j < len(arten); j++ {
			an := func(a string) bool { return arten[i] == a || arten[j] == a }
			wert := s
			if an("service_id") {
				wert = ""
			}
			if an("Zeichen") {
				wert += string([]byte{0x00})
			}
			sid := `"service_id":"` + wert + `"`
			glieder := []string{sid}
			if an("doppelt") {
				glieder = append(glieder, sid)
			}
			if an("Schlüssel") {
				glieder = append(glieder, `"extra":"x"`)
			}
			if an("id") {
				glieder = append(glieder, `"id":1`)
			}
			out = append(out, bForm{achse: "Paar", name: arten[i] + "+" + arten[j], body: []byte("{" + strings.Join(glieder, ",") + "}")})
		}
	}
	return out
}

// zaehleAchsen liefert die Zahl der Formen je Achse.
func zaehleAchsen(formen []bForm) map[string]int {
	out := map[string]int{}
	for _, f := range formen {
		out[f.achse]++
	}
	return out
}

// TestS567Zahlen hält M-Body gegen die Quelle des Tests (SPEC-292 Nr. 5): 11 gültige und 48 ungültige Formen, je Achse die
// Zahl des Plans; die Namen einer Achse sind verschieden, die Bodies der Paare auch.
func TestS567Zahlen(t *testing.T) {
	gleichAuf(t, "gültige Formen", len(bGueltig()), 11)
	gleichAuf(t, "Achsen der gültigen Formen", zaehleAchsen(bGueltig()), map[string]int{"gültig": 8, "Content-Type": 3})
	u := bUngueltig("x")
	gleichAuf(t, "ungültige Formen", len(u), 48)
	gleichAuf(t, "Achsen der ungültigen Formen", zaehleAchsen(u),
		map[string]int{"Wertform": 7, "Struktur": 7, "Schlüssel": 8, "Zeichen": 3, "Werte": 11, "Lesen": 2, "Paar": 10})
	namen, bodies := map[string]bool{}, map[string]bool{}
	for _, f := range u {
		namen[f.achse+"/"+f.name] = true
		bodies[string(f.body)+f.achse] = true
	}
	gleichAuf(t, "verschiedene Namen", len(namen), 48)
	gleichAuf(t, "verschiedene Bodies", len(bodies), 48)
}

// TestS5GueltigerBody hält S5 (SPEC-292 Nr. 5 a bis f): jede der 11 gültigen Formen ergibt 201 und genau einen Aufruf von
// Create mit den Literalen des Orakels.
func TestS5GueltigerBody(t *testing.T) {
	for _, f := range bGueltig() {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		rec := u.rufe(f.request())
		if rec.Code != 201 || rec.Body.String() != sessionAJSON {
			t.Errorf("%s: Antwort %d %q, erwartet 201 und die Session", f.name, rec.Code, rec.Body.String())
		}
		fake.nurDiagonale(t, f.name, opCreate)
		gleichAuf(t, f.name+": Argumente von Create", fake.argumente(), [][2]string{{f.svc, f.id}})
	}
}

// TestS5Leerraum hält S5 und S6 für die Klasse des Leerraums vor und hinter dem Objekt (SPEC-292 Nr. 5 b): Schleife über
// alle 256 Bytewerte an beiden Stellen; gültig sind genau Leerzeichen, Tabulator, LF und CR (4 × 2 Fälle), jedes andere Byte
// (248 × 2) ist 400 ohne Aufruf von Create.
func TestS5Leerraum(t *testing.T) {
	obj := []byte(`{"service_id":"svc"}`)
	gueltig, ungueltig := 0, 0
	for b := range 256 {
		for _, davor := range []bool{true, false} {
			body := append(slices.Clone(obj), byte(b))
			if davor {
				body = append([]byte{byte(b)}, obj...)
			}
			u, fake, _ := sessionsUmgebung(t)
			fake.session = sessionA()
			rec := u.rufe(postReq(body))
			ws := b == 0x20 || b == 0x09 || b == 0x0a || b == 0x0d
			name := "Byte " + strconv.Itoa(b) + " davor " + strconv.FormatBool(davor)
			switch {
			case ws && rec.Code == 201 && fake.zahlen()[opCreate] == 1:
				gueltig++
			case !ws && rec.Code == 400 && fake.gesamt() == 0:
				ungueltig++
			default:
				t.Errorf("%s: Status %d, %d Aufrufe von Create, Leerraum %v", name, rec.Code, fake.zahlen()[opCreate], ws)
			}
		}
	}
	gleichAuf(t, "gültige und ungültige Fälle", []int{gueltig, ungueltig}, []int{4 * 2, 252 * 2})
}

// zeichenWert ist ein Wert von service_id aus einem Zeichen: `s` + Zeichen + `x`.
func zeichenWert(c byte) string { return "s" + string([]byte{c}) + "x" }

// TestS5ZeichenImWert hält S5 und S6 für die Klasse des Zeichens in einer Zeichenkette (SPEC-292 Nr. 5 b, d): jedes
// Steuerzeichen unter 0x20 roh ist 400 (32 Fälle), jedes Byte von 0x20 bis 0x7f außer `"` und Backslash kommt unverändert
// an Create (94 Fälle), jede der 7 Folgen für ein Zeichen außerhalb von ASCII ebenso (Orakel: die Bytes), jede der 6 ungültigen
// UTF-8-Folgen ist 400.
func TestS5ZeichenImWert(t *testing.T) {
	versuche := func(wert []byte) (int, [][2]string) {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		rec := u.rufe(postReq([]byte(`{"service_id":"` + string(wert) + `"}`)))
		return rec.Code, fake.argumente()
	}
	steuer, ascii, gueltig, kaputt := 0, 0, 0, 0
	for c := range byte(0x80) {
		code, args := versuche([]byte(zeichenWert(c)))
		switch {
		case c < 0x20 && code == 400 && len(args) == 0:
			steuer++
		case c >= 0x20 && c != '"' && c != 0x5c && code == 201 && slices.Equal(args, [][2]string{{zeichenWert(c), ""}}):
			ascii++
		case c == '"' || c == 0x5c:
		default:
			t.Errorf("Zeichen 0x%02x: Status %d, Argumente %q", c, code, args)
		}
	}
	for _, rohe := range [][]byte{{0xC2, 0x80}, {0xDF, 0xBF}, {0xE0, 0xA0, 0x80}, {0xEF, 0xBF, 0xBD}, {0xEF, 0xBF, 0xBF}, {0xF0, 0x90, 0x80, 0x80}, {0xF4, 0x8F, 0xBF, 0xBF}} {
		if code, args := versuche(rohe); code != 201 || !slices.Equal(args, [][2]string{{string(rohe), ""}}) {
			t.Errorf("Zeichenfolge % x: Status %d, Argumente %q", rohe, code, args)
		}
		gueltig++
	}
	for _, rohe := range [][]byte{{0xFF}, {0xC0, 0x80}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0x80}, {0xE2, 0x82}} {
		if code, args := versuche(rohe); code != 400 || len(args) != 0 {
			t.Errorf("ungültige Folge % x: Status %d, Argumente %q", rohe, code, args)
		}
		kaputt++
	}
	gleichAuf(t, "Steuerzeichen, ASCII, gültige und ungültige Folgen", []int{steuer, ascii, gueltig, kaputt}, []int{32, 94, 7, 6})
}

// pruefeUngueltig hält eine ungültige Form gegen S6 und S7: 400 mit dem Fehlerobjekt von ODR-E2102, genau ein Zählen,
// weder Create noch Append, kein Wert aus dem Body in der Antwort; der Handler allein liefert model.ErrUngueltig ohne
// Wert im Text und keine Antwort.
func pruefeUngueltig(t *testing.T, f bForm) {
	t.Helper()
	name := f.achse + " " + f.name
	u, fake, routen := sessionsUmgebung(t)
	fake.err = &inbound.Fehler{Code: model.CodeE5005, Status: 404}
	rec := u.rufe(f.request())
	pruefeFehlerAntwort(t, name, rec, 400, objekt400, "")
	if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || andere != 0 {
		t.Errorf("%s: CountError %q, andere Aufrufe %d", name, codes, andere)
	}
	if fake.gesamt() != 0 || u.audit.Aufrufe() != 0 {
		t.Errorf("%s: %d Aufrufe des Dienstes, %d von Audit, erwartet 0 und 0", name, fake.gesamt(), u.audit.Aufrufe())
	}
	begrenzt := http.MaxBytesReader(nil, io.NopCloser(f.leser()), bodyGrenze)
	ant, err := routen[1].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: begrenzt})
	if !errors.Is(err, model.ErrUngueltig) || strings.Contains(err.Error(), canary()) || ant.Status != 0 || len(ant.Body) != 0 {
		t.Errorf("%s: Handler liefert %q, %v, erwartet model.ErrUngueltig ohne Wert und keine Antwort", name, ant.Body, err)
	}
	if fake.gesamt() != 0 {
		t.Errorf("%s: der Handler rief den Dienst", name)
	}
}

// TestS6S7UngueltigerBody hält S6 und S7 (SPEC-292 Nr. 5 a bis d, g): jede der 48 ungültigen Formen von M-Body ist 400
// `ODR-E2102`; der Dienst würde `ODR-E5005` liefern (Zähler 0: der Body zuerst).
func TestS6S7UngueltigerBody(t *testing.T) {
	for _, f := range bUngueltig(canary()) {
		pruefeUngueltig(t, f)
	}
}

// TestS6ArrayMitPaaren hält SPEC-292 Nr. 5 b für Arrays, deren Elemente sich wie die Paare eines Objekts lesen ließen (die
// Klasse „ein Array“ hinter der Form `[]`): jede der 3 Formen ist 400 ohne Aufruf von Create.
func TestS6ArrayMitPaaren(t *testing.T) {
	formen := []string{`["service_id","` + canary() + `"]`, `["service_id","` + canary() + `","id","a"]`, `["id","a","service_id","` + canary() + `"]`}
	for i, body := range formen {
		pruefeUngueltig(t, bForm{achse: "Array", name: strconv.Itoa(i), body: []byte(body)})
	}
	gleichAuf(t, "Zahl der Formen", len(formen), 3)
}

// TestS5Grenze hält die Grenze von SPEC-292 Nr. 5 a an beiden Seiten mit demselben Inhalt: 1048576 Bytes sind gültig, ein
// Byte mehr ist 400, und die Prüfung geht dem Aufruf von Create voraus.
func TestS5Grenze(t *testing.T) {
	for _, n := range []int{bodyGrenze - 1, bodyGrenze, bodyGrenze + 1} {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		rec := u.rufe(postReq(auffuellen(`{"service_id":"svc"}`, n)))
		soll, create := 201, 1
		if n > bodyGrenze {
			soll, create = 400, 0
		}
		if rec.Code != soll || fake.zahlen()[opCreate] != create {
			t.Errorf("%d Bytes: Status %d, %d Aufrufe von Create, erwartet %d und %d", n, rec.Code, fake.zahlen()[opCreate], soll, create)
		}
	}
}

// TestS7KontextfehlerBeimLesen hält SPEC-292 Nr. 5 a: ein Kontextfehler beim Lesen bleibt der Fehler der Route (der Handler
// liefert ihn unverändert, der Server bricht die Verbindung ab), und weder Create noch Append noch Zählen folgen.
func TestS7KontextfehlerBeimLesen(t *testing.T) {
	for _, ce := range []error{context.Canceled, context.DeadlineExceeded, errors.Join(errors.New("x"), context.Canceled)} {
		u, fake, routen := sessionsUmgebung(t)
		f := bForm{body: []byte(`{"service_id":"svc"`), rest: ce}
		_, err := routen[1].Handler(context.Background(), admin.Anfrage{Methode: "POST", Body: f.leser()})
		if err != ce {
			t.Errorf("Handler liefert %v, erwartet %v unverändert", err, ce)
		}
		rec := httptest.NewRecorder()
		if p := abbruchVon(func() { u.srv.ServeHTTP(rec, f.request()) }); p != http.ErrAbortHandler {
			t.Errorf("%v: ServeHTTP endete mit %v, erwartet den Abbruch der Verbindung", ce, p)
		}
		codes, andere := zaehlungenVon(u.tele)
		if fake.gesamt() != 0 || u.audit.Aufrufe() != 0 || len(codes) != 0 || andere != 0 || rec.Body.Len() != 0 {
			t.Errorf("%v: %d Aufrufe des Dienstes, %d von Audit, Zählungen %q, Body %q", ce, fake.gesamt(), u.audit.Aufrufe(), codes, rec.Body.String())
		}
	}
}

// abbruchVon führt f aus und liefert den Wert einer Panik (nil ohne Panik).
func abbruchVon(f func()) (p any) {
	defer func() { p = recover() }()
	f()
	return nil
}
