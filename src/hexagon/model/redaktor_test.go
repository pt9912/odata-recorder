package model_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

var (
	errTest = errors.New("testfehler")
	// errVoll steht für den Fehler eines Schreibers, der voll ist (outbound.ErrZuGross); das
	// Modell importiert die Ports nicht, der Test prüft allein, dass die Ursache erreichbar bleibt.
	errVoll = errors.New("schreiber voll")
)

// --- Eingabeströme -------------------------------------------------------------------------

// teilReader liefert vorgegebene Teile, je Read höchstens einen.
type teilReader struct {
	teile    []string
	leer     bool // vor jedem Teil ein leerer Read
	eofMit   bool // das letzte Teil kommt zusammen mit io.EOF
	i        int
	pendLeer bool
	reads    int
	gelesen  int
}

func teile(leer, eofMit bool, t ...string) *teilReader {
	return &teilReader{teile: append([]string(nil), t...), leer: leer, eofMit: eofMit, pendLeer: leer}
}

func einmal(s string) *teilReader { return teile(false, false, s) }

func (r *teilReader) Read(p []byte) (int, error) {
	r.reads++
	if r.pendLeer {
		r.pendLeer = false
		return 0, nil
	}
	if r.i >= len(r.teile) {
		return 0, io.EOF
	}
	n := copy(p, r.teile[r.i])
	r.teile[r.i] = r.teile[r.i][n:]
	if r.teile[r.i] == "" {
		r.i++
	}
	r.pendLeer = r.leer
	r.gelesen += n
	if r.eofMit && r.i == len(r.teile) {
		return n, io.EOF
	}
	return n, nil
}

// zerlegung beschreibt, wie ein Strom in Reads zerfällt.
type zerlegung struct {
	name   string
	teile  []string
	leer   bool
	eofMit bool
}

func (z zerlegung) leser() *teilReader { return teile(z.leer, z.eofMit, z.teile...) }

func einzelneBytes(s string) []string {
	out := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		out[i] = s[i : i+1]
	}
	return out
}

// zerlegungen liefert jede Zweiteilung, einzelne Bytes, einzelne Bytes mit leeren Reads dazwischen
// und das letzte Byte zusammen mit io.EOF.
func zerlegungen(s string) []zerlegung {
	var out []zerlegung
	for i := 1; i < len(s); i++ {
		out = append(out, zerlegung{name: "geteilt bei " + strconv.Itoa(i), teile: []string{s[:i], s[i:]}})
	}
	return append(out,
		zerlegung{name: "einzeln", teile: einzelneBytes(s)},
		zerlegung{name: "einzeln mit leeren Reads", teile: einzelneBytes(s), leer: true},
		zerlegung{name: "einzeln, letztes mit EOF", teile: einzelneBytes(s), eofMit: true},
		zerlegung{name: "einmal, mit EOF", teile: []string{s}, eofMit: true},
	)
}

// abbruchReader liefert die ersten k Bytes von daten und endet dann mit err.
type abbruchReader struct {
	daten    string
	k        int
	err      error
	zusammen bool // die letzten Bytes kommen zusammen mit dem Fehler
	teil     int  // höchstens so viele Bytes je Read; 0 heißt alle
	pos      int
	reads    int
}

func (r *abbruchReader) Read(p []byte) (int, error) {
	r.reads++
	rest := r.daten[r.pos:r.k]
	if rest == "" {
		return 0, r.err
	}
	if r.teil > 0 && len(rest) > r.teil {
		rest = rest[:r.teil]
	}
	n := copy(p, rest)
	r.pos += n
	if r.zusammen && r.pos == r.k {
		return n, r.err
	}
	return n, nil
}

// genReader erzeugt kopf, n Bytes aus muster (wiederholt) und schluss, ohne sie zu halten.
type genReader struct {
	kopf, muster, schluss string
	n                     int
	pos                   int
}

func (r *genReader) Read(p []byte) (int, error) {
	gesamt := len(r.kopf) + r.n + len(r.schluss)
	if r.pos >= gesamt {
		return 0, io.EOF
	}
	i := 0
	for ; i < len(p) && r.pos < gesamt; i++ {
		switch {
		case r.pos < len(r.kopf):
			p[i] = r.kopf[r.pos]
		case r.pos < len(r.kopf)+r.n:
			p[i] = r.muster[(r.pos-len(r.kopf))%len(r.muster)]
		default:
			p[i] = r.schluss[r.pos-len(r.kopf)-r.n]
		}
		r.pos++
	}
	return i, nil
}

// --- Ausgabeströme -------------------------------------------------------------------------

type zaehlWriter struct {
	n      int64
	writes int
}

func (w *zaehlWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	w.writes++
	return len(p), nil
}

// scheiterWriter schlägt beim bei-ten Write fehl (oder schreibt dort ein Byte zu wenig).
type scheiterWriter struct {
	buf            bytes.Buffer
	writes, bei    int
	err            error
	kurz           bool
	leser          *teilReader
	readsBeiFehler int
}

func (w *scheiterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.bei {
		if w.leser != nil {
			w.readsBeiFehler = w.leser.reads
		}
		if w.kurz {
			w.buf.Write(p[:len(p)-1])
			return len(p) - 1, nil
		}
		return 0, w.err
	}
	return w.buf.Write(p)
}

// --- Fälle ---------------------------------------------------------------------------------

type ergebnis struct {
	aus    string
	status model.RedactionStatus
	err    error
}

func standardPfade(t *testing.T) []model.JSONPfad {
	t.Helper()
	return pfadeAus(t, "$..password", "$..passwd", "$..secret", "$..client_secret", "$..access_token", "$..refresh_token", "$..api_key")
}

func pfadeAus(t *testing.T, texte ...string) []model.JSONPfad {
	t.Helper()
	out := []model.JSONPfad{}
	for _, s := range texte {
		p, err := model.ParseJSONPfad(s)
		if err != nil {
			t.Fatalf("ParseJSONPfad(%q): %v", s, err)
		}
		out = append(out, p)
	}
	return out
}

func jsonEingang() model.BodyEingang {
	return model.BodyEingang{ContentType: []string{"application/json"}}
}

func bodyRegeln(t *testing.T) model.BodyRegeln {
	t.Helper()
	return model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 1 << 20}
}

func redigiere(t *testing.T, src io.Reader, ein model.BodyEingang, regeln model.BodyRegeln) ergebnis {
	t.Helper()
	var dst bytes.Buffer
	st, err := model.RedigiereBody(&dst, src, ein, regeln)
	return ergebnis{dst.String(), st, err}
}

// bodyFall ist eine Zeile der Tabelle bodyFaelle: Eingabe, Typ und Erwartung stehen von Hand
// ausgeschrieben da; bei einem Fehler ist aus der redigierte Teil bis zum Verstoß. geheim sagt,
// dass der Fall das Geheimnis unter einem treffenden Namen führt, arten welche Wertarten unter
// treffenden Namen (string, null, true, false, number, object, array).
type bodyFall struct {
	name        string
	ct          []string
	ctGesetzt   bool
	meta        bool
	max         int64
	pfade       []string
	eigenePfade bool
	ein, aus    string
	status      model.RedactionStatus
	grund       model.UnsicherGrund
	geheim      bool
	arten       string
}

func gut(name, ein, aus string, st model.RedactionStatus) bodyFall {
	return bodyFall{name: name, ein: ein, aus: aus, status: st}
}

func schlecht(name, ein, aus string, g model.UnsicherGrund) bodyFall {
	return bodyFall{name: name, ein: ein, aus: aus, status: model.RedactionUnsafe, grund: g}
}

func (f bodyFall) typ(ct ...string) bodyFall { f.ct, f.ctGesetzt = ct, true; return f }
func (f bodyFall) edmx() bodyFall            { return f.typ("application/xml").mitMeta() }
func (f bodyFall) mitMeta() bodyFall         { f.meta = true; return f }
func (f bodyFall) mitMax(n int64) bodyFall   { f.max = n; return f }
func (f bodyFall) mitPfade(p ...string) bodyFall {
	f.pfade, f.eigenePfade = p, true
	return f
}
func (f bodyFall) mitGeheimnis(arten string) bodyFall { f.geheim, f.arten = true, arten; return f }

func (f bodyFall) eingang() model.BodyEingang {
	ein := model.BodyEingang{ContentType: f.ct, Metadata: f.meta}
	if !f.ctGesetzt {
		ein.ContentType = jsonEingang().ContentType
	}
	return ein
}

func (f bodyFall) regeln(t *testing.T) model.BodyRegeln {
	t.Helper()
	r := bodyRegeln(t)
	if f.max != 0 {
		r.MaxBytes = f.max
	}
	if f.eigenePfade {
		r.Pfade = pfadeAus(t, f.pfade...)
	}
	return r
}

func (f bodyFall) fuehre(t *testing.T, src io.Reader) ergebnis {
	t.Helper()
	return redigiere(t, src, f.eingang(), f.regeln(t))
}

// pruefe fährt den Fall mit einem einzigen Read und vergleicht Ausgabe, Status und Grund.
func (f bodyFall) pruefe(t *testing.T) {
	t.Helper()
	f.vergleiche(t, f.fuehre(t, einmal(f.ein)), "ein Read")
}

func kurz(s string) string {
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}

func (f bodyFall) vergleiche(t *testing.T, g ergebnis, wie string) {
	t.Helper()
	if g.aus != f.aus {
		t.Errorf("%s [%s]: Ausgabe %q, erwartet %q", f.name, wie, kurz(g.aus), kurz(f.aus))
	}
	if g.status != f.status {
		t.Errorf("%s [%s]: Status %q, erwartet %q", f.name, wie, g.status, f.status)
	}
	if f.grund == "" {
		if g.err != nil {
			t.Errorf("%s [%s]: Fehler %v", f.name, wie, g.err)
		}
		return
	}
	var uf *model.UnsicherFehler
	switch {
	case !errors.As(g.err, &uf):
		t.Errorf("%s [%s]: Fehler %v ist kein *UnsicherFehler", f.name, wie, g.err)
	case uf.Grund != f.grund:
		t.Errorf("%s [%s]: Grund %q, erwartet %q", f.name, wie, uf.Grund, f.grund)
	case !errors.Is(g.err, model.ErrUnsicher):
		t.Errorf("%s [%s]: errors.Is(err, ErrUnsicher) ist falsch", f.name, wie)
	case uf.Code() != model.CodeE2401 || uf.Code().String() != "ODR-E2401":
		t.Errorf("%s [%s]: Code %v", f.name, wie, uf.Code())
	}
}

func wiederhole(s string, n int) string { return strings.Repeat(s, n) }

// bodyFaelle ist die eine Tabelle der Fälle; die Erwartungen sind gegen SPEC-118 geschrieben,
// nicht aus dem Code berechnet, und zitieren den Marker als Text.
func bodyFaelle() []bodyFall {
	zahl400 := wiederhole("7", 400)
	return []bodyFall{
		gut("string-ersetzt", `{"password":"S3CR3T-7f"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("layout-bleibt",
			"{ \"password\" :\t\"S3CR3T-7f\" ,\r\n\"b\":1.50E+2,\"a\":-0,\"c\":[1, 2.0],\"d\":\"\\u00e9\\/\"}",
			"{ \"password\" :\t\"[REDACTED]\" ,\r\n\"b\":1.50E+2,\"a\":-0,\"c\":[1, 2.0],\"d\":\"\\u00e9\\/\"}",
			model.RedactionApplied).mitGeheimnis("string"),
		gut("leerer-string-ersetzt", `{"password":""}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("escape-im-wert", `{"password":"S3CR3T-7f\"x\\","k":1}`, `{"password":"[REDACTED]","k":1}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("unicode-escape-im-wert", `{"password":"\u0022S3CR3T-7f"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("verschachtelt", `{"a":{"b":[{"c":{"password":"S3CR3T-7f"}}]}}`, `{"a":{"b":[{"c":{"password":"[REDACTED]"}}]}}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("doppelte-namen", `{"password":"S3CR3T-7f","x":0,"password":"S3CR3T-7f"}`, `{"password":"[REDACTED]","x":0,"password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("schreibweisen", `{"password":"a","PASSWORD":null,"Password":"b"}`, `{"password":"[REDACTED]","PASSWORD":null,"Password":"[REDACTED]"}`, model.RedactionApplied),
		gut("name-escape", `{"pass\u0077ord":"S3CR3T-7f"}`, `{"pass\u0077ord":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("array-element-kein-name", `["password","x"]`, `["password","x"]`, model.RedactionNone),
		gut("wert-kein-name", `{"a":"password"}`, `{"a":"password"}`, model.RedactionNone),
		gut("wurzel-string", `"password"`, `"password"`, model.RedactionNone),
		gut("null-treffer", `{"password":null}`, `{"password":null}`, model.RedactionNone),
		gut("true-treffer", `{"password":true}`, `{"password":true}`, model.RedactionNone),
		gut("false-treffer", `{"password":false}`, `{"password":false}`, model.RedactionNone),
		gut("null-mit-geheimnis", `{"password":null,"secret":"S3CR3T-7f"}`, `{"password":null,"secret":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string null"),
		gut("true-mit-geheimnis", `{"password":true,"secret":"S3CR3T-7f"}`, `{"password":true,"secret":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string true"),
		gut("false-mit-geheimnis", `{"password":false,"secret":"S3CR3T-7f"}`, `{"password":false,"secret":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string false"),
		gut("null-tief", `{"a":{"password":null}}`, `{"a":{"password":null}}`, model.RedactionNone),
		gut("marker-schon-da", `{"password":"[REDACTED]"}`, `{"password":"[REDACTED]"}`, model.RedactionNone),
		gut("marker-beide", `{"password":"[REDACTED]","secret":"[REDACTED]"}`, `{"password":"[REDACTED]","secret":"[REDACTED]"}`, model.RedactionNone),
		gut("marker-escape", `{"password":"\u005bREDACTED]"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-mit-rest", `{"password":"[REDACTED]x"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-mit-leerzeichen", `{"password":"[REDACTED] "}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-abgeschnitten", `{"password":"[REDACTE"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-ohne-klammer", `{"password":"[REDACTED"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-mit-vorsatz", `{"password":"x[REDACTED]"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("marker-gleiche-laenge", `{"password":"S3CR3T-7fX"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("marker-gleiche-laenge-letztes-byte", `{"password":"[REDACTEX]"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("name-ueber-musterlaenge", `{"AAAAAAAAAAAAAAAAAAAApassword":"x","password":"S3CR3T-7f"}`,
			`{"AAAAAAAAAAAAAAAAAAAApassword":"x","password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("ohne-treffer", `{"a":1,"b":[true,false,null,"x"]}`, `{"a":1,"b":[true,false,null,"x"]}`, model.RedactionNone),
		gut("ohne-pfade", `{"password":"S3CR3T-7f"}`, `{"password":"S3CR3T-7f"}`, model.RedactionNone).mitPfade(),
		gut("eigenes-muster", `{"foo":"x","password":"S3CR3T-7f"}`, `{"foo":"[REDACTED]","password":"S3CR3T-7f"}`, model.RedactionApplied).mitPfade("$..foo"),
		gut("skalar-zahl", `12`, `12`, model.RedactionNone),
		gut("skalar-string", `"abc"`, `"abc"`, model.RedactionNone),
		gut("skalar-true", `true`, `true`, model.RedactionNone),
		gut("skalar-minus-null", `-0`, `-0`, model.RedactionNone),
		gut("skalar-exponent", `1e5`, `1e5`, model.RedactionNone),
		gut("leerraum-umhuellt", " \n\t\r{} \n\t\r", " \n\t\r{} \n\t\r", model.RedactionNone),
		gut("leeres-objekt", `{}`, `{}`, model.RedactionNone),
		gut("leeres-array", `[]`, `[]`, model.RedactionNone),
		gut("zahlen-durchgereicht", `{"a":1.50E+2,"b":-0,"c":1E5,"d":0,"e":-12.5e+3}`, `{"a":1.50E+2,"b":-0,"c":1E5,"d":0,"e":-12.5e+3}`, model.RedactionNone),
		gut("zahl-400-ziffern", `{"a":`+zahl400+`}`, `{"a":`+zahl400+`}`, model.RedactionNone),
		gut("mehrbyte-ueberall", `{"é€😀":"é€😀","password":"é€😀S3CR3T-7f","x":"é€😀"}`, `{"é€😀":"é€😀","password":"[REDACTED]","x":"é€😀"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("escape-und-surrogat", `{"a\ud83d\ude00\n":"\u00e9\ud800","password":"\u00e9\ud83d\ude00S3CR3T-7f"}`, `{"a\ud83d\ude00\n":"\u00e9\ud800","password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("literale-und-zahlen", `{"n":-12.5e+3,"t":true,"f":false,"z":null,"password":"S3CR3T-7f"}`, `{"n":-12.5e+3,"t":true,"f":false,"z":null,"password":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string"),
		gut("treffer-bei-tiefe-512", wiederhole("[", 511)+`{"password":"S3CR3T-7f"}`+wiederhole("]", 511),
			wiederhole("[", 511)+`{"password":"[REDACTED]"}`+wiederhole("]", 511), model.RedactionApplied).mitGeheimnis("string"),
		gut("wert-10000-bytes", `{"password":"`+wiederhole("a", 10000)+`"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("edmx", `<?xml version="1.0"?><edmx:Edmx Version="4.0"></edmx:Edmx>`, `<?xml version="1.0"?><edmx:Edmx Version="4.0"></edmx:Edmx>`, model.RedactionNone).edmx(),
		gut("edmx-bom-leerraum", "\xEF\xBB\xBF \r\n\t<a/>", "\xEF\xBB\xBF \r\n\t<a/>", model.RedactionNone).edmx(),
		gut("edmx-element-passwort", `<password>S3CR3T-7f</password>`, `<password>S3CR3T-7f</password>`, model.RedactionNone).edmx(),
		gut("json-bei-metadata", `{"password":"S3CR3T-7f"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).mitMeta().mitGeheimnis("string"),

		schlecht("typ-text", `{"password":"S3CR3T-7f"}`, ``, model.GrundTyp).typ("text/plain").mitGeheimnis("string"),
		schlecht("typ-fehlt", `{"password":"S3CR3T-7f"}`, ``, model.GrundTyp).typ().mitGeheimnis("string"),
		schlecht("typ-multipart", `{"password":"S3CR3T-7f"}`, ``, model.GrundTyp).typ("multipart/mixed; boundary=b").mitGeheimnis("string"),
		schlecht("typ-xml-ohne-metadata", `<a/>`, ``, model.GrundTyp).typ("application/xml"),
		schlecht("edmx-json-body", `{"password":"S3CR3T-7f"}`, ``, model.GrundMismatch).edmx().mitGeheimnis("string"),
		schlecht("edmx-x-vor-klammer", `x<`, ``, model.GrundMismatch).edmx(),
		schlecht("edmx-bom-allein", "\xEF\xBB\xBF", "\xEF\xBB\xBF", model.GrundMismatch).edmx(),
		schlecht("edmx-leerraum-allein", "  ", "  ", model.GrundMismatch).edmx(),
		schlecht("edmx-bom-rest", "\xEF\xBB", "\xEF\xBB", model.GrundMismatch).edmx(),
		schlecht("edmx-bom-falsch", "\xEF\xBB<", "\xEF\xBB", model.GrundMismatch).edmx(),
		schlecht("edmx-bom-nach-leerraum", " \xEF\xBB\xBF<a/>", " ", model.GrundMismatch).edmx(),
		schlecht("edmx-bom-doppelt", "\xEF\xBB\xBF\xEF\xBB\xBF<a/>", "\xEF\xBB\xBF", model.GrundMismatch).edmx(),
		schlecht("edmx-bom-zweites-byte", "\xEF\x41\xBF<a/>", "\xEF", model.GrundMismatch).edmx(),
		schlecht("edmx-utf16", "\xFF\xFE<\x00", ``, model.GrundMismatch).edmx(),
		schlecht("syntax-komma-hinten", `{"a":1,}`, `{"a":1,`, model.GrundSyntax),
		schlecht("syntax-wert-fehlt-treffer", `{"password":}`, `{"password":`, model.GrundSyntax),
		schlecht("syntax-x-treffer", `{"password":x}`, `{"password":`, model.GrundSyntax),
		schlecht("syntax-zweiter-wert", `{"a":1}{"b":2}`, `{"a":1}`, model.GrundSyntax),
		schlecht("syntax-nachlauf", `{"a":1} x`, `{"a":1} `, model.GrundSyntax),
		schlecht("syntax-escape-im-ersetzten", `{"password":"S3CR3T-7f\x"}`, `{"password":`, model.GrundSyntax).mitGeheimnis("string"),
		schlecht("syntax-hex-im-ersetzten", `{"password":"S3CR3T-7f\u12G4"}`, `{"password":`, model.GrundSyntax).mitGeheimnis("string"),
		schlecht("syntax-bom", "\xEF\xBB\xBF{}", ``, model.GrundSyntax),
		schlecht("syntax-roher-umbruch", "{\"a\":\"x\ny\"}", "{\"a\":\"x", model.GrundSyntax),
		schlecht("syntax-fuehrende-null", `{"a":01}`, `{"a":0`, model.GrundSyntax),
		schlecht("syntax-xml-unter-json", `<a/>`, ``, model.GrundSyntax),
		schlecht("syntax-erster-verstoss", `{"a":01,"password":1}`, `{"a":0`, model.GrundSyntax),
		schlecht("utf8-name", "{\"\xC0\x80\":1}", `{"`, model.GrundUTF8),
		schlecht("utf8-ersetzter-wert", "{\"password\":\"S3CR3T-7f\xFF\"}", `{"password":`, model.GrundUTF8).mitGeheimnis("string"),
		schlecht("utf8-anderer-wert", "{\"p\":\"\xE2\x82\"}", "{\"p\":\"\xE2\x82", model.GrundUTF8),
		schlecht("tiefe-513-array", wiederhole("[", 513), wiederhole("[", 512), model.GrundTiefe),
		schlecht("tiefe-513-treffer", wiederhole("[", 512)+`{"password":"S3CR3T-7f"}`, wiederhole("[", 512), model.GrundTiefe).mitGeheimnis("string"),
		schlecht("nichtstring-zahl", `{"password":12345}`, `{"password":`, model.GrundNichtstring),
		schlecht("nichtstring-zahl-geheim", `{"a":1,"password":0,"secret":"S3CR3T-7f"}`, `{"a":1,"password":`, model.GrundNichtstring).mitGeheimnis("number"),
		schlecht("nichtstring-objekt", `{"password":{"v":"S3CR3T-7f"}}`, `{"password":`, model.GrundNichtstring).mitGeheimnis("object"),
		schlecht("nichtstring-array", `{"password":["S3CR3T-7f"]}`, `{"password":`, model.GrundNichtstring).mitGeheimnis("array"),
		schlecht("nichtstring-aussen", `{"secret":{"password":"S3CR3T-7f"}}`, `{"secret":`, model.GrundNichtstring).mitGeheimnis("object"),
		schlecht("zu-gross-gueltig", `{"a":"bbbbbbbbb"}`, `{"a":"bbbbbbbbb"`, model.GrundZuGross).mitMax(16),
		schlecht("zu-gross-geheimnis", `{"password":"S3CR3T-7f"}`, `{"password":`, model.GrundZuGross).mitMax(20).mitGeheimnis("string"),
		schlecht("unvollstaendig-wert-geheim", `{"password":"S3CR3T-7f`, `{"password":`, model.GrundUnvollstaendig).mitGeheimnis("string"),
		schlecht("unvollstaendig-array", `{"a":[1,2`, `{"a":[1,2`, model.GrundUnvollstaendig),
		schlecht("unvollstaendig-name", `{"passw`, `{"passw`, model.GrundUnvollstaendig),
		schlecht("unvollstaendig-nur-leerraum", " ", " ", model.GrundUnvollstaendig),
		schlecht("unvollstaendig-literal", `tru`, `tru`, model.GrundUnvollstaendig),
		schlecht("unvollstaendig-zahl", `[-`, `[-`, model.GrundUnvollstaendig),
	}
}

func fall(t *testing.T, name string) bodyFall {
	t.Helper()
	for _, f := range bodyFaelle() {
		if f.name == name {
			return f
		}
	}
	t.Fatalf("kein Fall %q in bodyFaelle", name)
	return bodyFall{}
}

// --- Tests ---------------------------------------------------------------------------------

func TestBodyArgumente(t *testing.T) {
	doc := `{"password":"x"}`
	for _, grenze := range []int64{0, -1} {
		for _, ct := range []string{"application/json", "text/plain"} {
			r := einmal(doc)
			w := &zaehlWriter{}
			st, err := model.RedigiereBody(w, r, model.BodyEingang{ContentType: []string{ct}}, model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: grenze})
			if !errors.Is(err, model.ErrUngueltig) || errors.Is(err, model.ErrUnsicher) || st != "" || r.reads != 0 || w.writes != 0 {
				t.Errorf("MaxBytes %d, %s: Status %q, Fehler %v, Reads %d, Writes %d", grenze, ct, st, err, r.reads, w.writes)
			}
		}
	}
	nullwert := []model.JSONPfad{{}}
	gemischt := append(pfadeAus(t, "$..password"), model.JSONPfad{})
	for name, pfade := range map[string][]model.JSONPfad{"nur Nullwert": nullwert, "Nullwert hinter gültigem": gemischt} {
		r := einmal(doc)
		w := &zaehlWriter{}
		st, err := model.RedigiereBody(w, r, jsonEingang(), model.BodyRegeln{Pfade: pfade, MaxBytes: 100})
		if !errors.Is(err, model.ErrUngueltig) || st != "" || r.reads != 0 || w.writes != 0 {
			t.Errorf("%s: Status %q, Fehler %v, Reads %d, Writes %d", name, st, err, r.reads, w.writes)
		}
	}
	regeln := model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 100}
	if st, err := model.RedigiereBody(nil, einmal(doc), jsonEingang(), regeln); !errors.Is(err, model.ErrUngueltig) || st != "" {
		t.Errorf("dst nil: %q, %v", st, err)
	}
	if st, err := model.RedigiereBody(&bytes.Buffer{}, nil, jsonEingang(), regeln); !errors.Is(err, model.ErrUngueltig) || st != "" {
		t.Errorf("src nil: %q, %v", st, err)
	}
	for name, pfade := range map[string][]model.JSONPfad{"nil": nil, "leer": {}} {
		g := redigiere(t, einmal(doc), jsonEingang(), model.BodyRegeln{Pfade: pfade, MaxBytes: 100})
		if g.err != nil || g.status != model.RedactionNone || g.aus != doc {
			t.Errorf("Pfade %s: %+v", name, g)
		}
	}
	// Die Prüfung der Argumente kommt vor der Typprüfung.
	g := redigiere(t, einmal(doc), model.BodyEingang{ContentType: []string{"text/plain"}}, model.BodyRegeln{MaxBytes: 0})
	if !errors.Is(g.err, model.ErrUngueltig) || errors.Is(g.err, model.ErrUnsicher) {
		t.Errorf("MaxBytes 0 mit text/plain: %v", g.err)
	}
}

func TestBodyLeer(t *testing.T) {
	typen := map[string]model.BodyEingang{
		"nil":      {},
		"text":     {ContentType: []string{"text/plain"}},
		"json":     jsonEingang(),
		"xml":      {ContentType: []string{"application/xml"}, Metadata: true},
		"mehrfach": {ContentType: []string{"a", "b"}},
	}
	for name, ein := range typen {
		for lname, src := range map[string]io.Reader{"keine Bytes": einmal(""), "nur leere Reads": teile(true, false)} {
			if g := redigiere(t, src, ein, bodyRegeln(t)); g.err != nil || g.status != model.RedactionNone || g.aus != "" {
				t.Errorf("%s, %s: %+v", name, lname, g)
			}
		}
	}
	for _, ein := range []string{" ", "\n"} {
		bodyFall{name: "nur-leerraum", ein: ein, aus: ein, status: model.RedactionUnsafe, grund: model.GrundUnvollstaendig}.pruefe(t)
	}
	edmx := bodyFall{name: "edmx-nur-leerraum", ein: " ", aus: " ", status: model.RedactionUnsafe, grund: model.GrundMismatch}.edmx()
	edmx.pruefe(t)
	g := redigiere(t, &abbruchReader{err: errTest}, jsonEingang(), bodyRegeln(t))
	var uf *model.UnsicherFehler
	if !errors.As(g.err, &uf) || uf.Grund != model.GrundEingabe || !errors.Is(g.err, errTest) || g.status != model.RedactionUnsafe || g.aus != "" {
		t.Errorf("Fehler des ersten Reads: %+v", g)
	}
}

func TestBodyTyp(t *testing.T) {
	gueltig := []string{
		"application/json", "APPLICATION/JSON", " application/json ", "\tapplication/json\t",
		"application/json; charset=utf-8", "application/json;odata.metadata=minimal;odata.streaming=true",
		"application/vnd.api+json", "application/odata+json", "Application/Vnd.Api+JSON",
	}
	for _, ct := range gueltig {
		gut("typ "+ct, `{}`, `{}`, model.RedactionNone).typ(ct).pruefe(t)
	}
	ungueltig := [][]string{
		nil, {}, {""}, {";charset=utf-8"}, {"application/json", "application/json"}, {"application/json,application/json"},
		{"text/json"}, {"text/plain"}, {"text/xml"}, {"multipart/mixed; boundary=b"}, {"application/xjson"},
		{"application/jsonx"}, {"application/json-seq"}, {"application/+json"}, {"application/x-www-form-urlencoded"},
		{"application/octet-stream"}, {"*/*"}, {"application/xml"}, {"json"}, {"application/ json"},
		// Am Rand fallen nur Leerzeichen und Tabulator weg, kein anderer Leerraum.
		{"application/json\r"}, {"application/json\n"}, {"\rapplication/json"}, {"\napplication/json"}, {"application/json\r\n"},
		{"application/json\r;charset=utf-8"}, {"\fapplication/json"}, {"application/json\v"}, {"\u00a0application/json"}, {"application/json\u00a0"},
	}
	for _, ct := range ungueltig {
		schlecht("typ "+strings.Join(ct, "|"), `{}`, ``, model.GrundTyp).typ(ct...).pruefe(t)
	}
	// Metadata ändert den Typ JSON nicht und macht text/xml nicht bekannt.
	schlecht("text/xml mit Metadata", `<a/>`, ``, model.GrundTyp).typ("text/xml").mitMeta().pruefe(t)
	gut("json mit Metadata", `{"password":"x"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).mitMeta().pruefe(t)
	gut("xml mit Parameter", `<a/>`, `<a/>`, model.RedactionNone).typ("Application/XML; charset=utf-8").mitMeta().pruefe(t)
}

func TestBodyEdmx(t *testing.T) {
	for _, n := range []string{
		"edmx", "edmx-bom-leerraum", "edmx-element-passwort", "json-bei-metadata", "typ-xml-ohne-metadata", "edmx-json-body",
		"edmx-x-vor-klammer", "edmx-bom-allein", "edmx-leerraum-allein", "edmx-bom-rest", "edmx-bom-falsch",
		"edmx-bom-nach-leerraum", "edmx-bom-doppelt", "edmx-bom-zweites-byte", "edmx-utf16",
	} {
		fall(t, n).pruefe(t)
	}
	gut("edmx-mit-json-im-element", `<a>{"password":"S3CR3T-7f"}</a>`, `<a>{"password":"S3CR3T-7f"}</a>`, model.RedactionNone).edmx().pruefe(t)
}

func TestBodyMarker(t *testing.T) {
	for _, n := range []string{
		"string-ersetzt", "layout-bleibt", "leerer-string-ersetzt", "escape-im-wert", "unicode-escape-im-wert",
		"wert-10000-bytes", "literale-und-zahlen", "zahlen-durchgereicht", "zahl-400-ziffern", "escape-und-surrogat",
	} {
		fall(t, n).pruefe(t)
	}
	gut("wert-1-byte", `{"password":"x","k":1}`, `{"password":"[REDACTED]","k":1}`, model.RedactionApplied).pruefe(t)
	gut("escape-backslash", `{"password":"a\\","k":1}`, `{"password":"[REDACTED]","k":1}`, model.RedactionApplied).pruefe(t)
	gut("escape-anfuehrungszeichen", `{"password":"a\"b","k":1}`, `{"password":"[REDACTED]","k":1}`, model.RedactionApplied).pruefe(t)
}

func TestBodyNamen(t *testing.T) {
	treffer := []string{
		`password`, `PassWord`, `PASSWORD`, `pass\u0077ord`, `\u0050ASSWORD`, `\u0070assword`, `pass\u0077\u006fr\u0064`,
		`api_\u006b\u0065y`, `api_\u004B\u0045Y`, `Api_Key`,
	}
	for _, n := range treffer {
		gut("treffer "+n, `{"`+n+`":"S3CR3T-7f"}`, `{"`+n+`":"[REDACTED]"}`, model.RedactionApplied).mitGeheimnis("string").pruefe(t)
	}
	keine := []string{
		`passwords`, `passwor`, `xpassword`, "password ", "passw\u00f6rd", "api_\u212Aey", `api_\u212Aey`, `pass\\u0077ord`,
		``, `password@odata.type`, `password\u0000`, `pass\nword`, `pass\/word`, `ssap`, `p`,
	}
	for _, n := range keine {
		doc := `{"` + n + `":"S3CR3T-7f"}`
		gut("kein treffer "+n, doc, doc, model.RedactionNone).pruefe(t)
	}
	gut("muster-mit-grossbuchstaben", `{"foo-bar_1":"a","FOO-BAR_1":"b","Foo-Bar_2":"c"}`, `{"foo-bar_1":"[REDACTED]","FOO-BAR_1":"[REDACTED]","Foo-Bar_2":"c"}`,
		model.RedactionApplied).mitPfade("$..Foo-Bar_1").pruefe(t)
}

func TestBodyDefaults(t *testing.T) {
	namen := []string{"password", "passwd", "secret", "client_secret", "access_token", "refresh_token", "api_key"}
	alle, alleErwartet := []string{}, []string{}
	for _, n := range namen {
		alle = append(alle, `"`+n+`":"S3CR3T-7f"`)
		alleErwartet = append(alleErwartet, `"`+n+`":"[REDACTED]"`)
	}
	rueckwaerts := func(p []model.JSONPfad) []model.JSONPfad {
		out := make([]model.JSONPfad, len(p))
		for i := range p {
			out[len(p)-1-i] = p[i]
		}
		return out
	}
	for _, richtung := range []string{"vorwärts", "rückwärts"} {
		pfade := standardPfade(t)
		if richtung == "rückwärts" {
			pfade = rueckwaerts(pfade)
		}
		regeln := model.BodyRegeln{Pfade: pfade, MaxBytes: 1 << 20}
		for _, n := range namen {
			g := redigiere(t, einmal(`{"`+n+`":"S3CR3T-7f"}`), jsonEingang(), regeln)
			if g.aus != `{"`+n+`":"[REDACTED]"}` || g.status != model.RedactionApplied || g.err != nil {
				t.Errorf("%s, %s: %+v", richtung, n, g)
			}
		}
		g := redigiere(t, einmal("{"+strings.Join(alle, ",")+"}"), jsonEingang(), regeln)
		if g.aus != "{"+strings.Join(alleErwartet, ",")+"}" || g.err != nil {
			t.Errorf("%s, alle sieben: %+v", richtung, g)
		}
	}
}

func TestBodyMitglieder(t *testing.T) {
	for _, n := range []string{
		"array-element-kein-name", "wert-kein-name", "wurzel-string", "verschachtelt", "doppelte-namen", "schreibweisen", "null-tief",
	} {
		fall(t, n).pruefe(t)
	}
	gut("array-von-objekten", `[{"password":"a"},{"password":"b"}]`, `[{"password":"[REDACTED]"},{"password":"[REDACTED]"}]`, model.RedactionApplied).pruefe(t)
	gut("tief-unter-anderem-namen", `{"x":{"y":{"secret":"S3CR3T-7f"}}}`, `{"x":{"y":{"secret":"[REDACTED]"}}}`, model.RedactionApplied).pruefe(t)
	gut("treffer-zweimal-gleiche-ebene-und-tief", `{"password":"a","n":{"password":"b"},"password":"c"}`,
		`{"password":"[REDACTED]","n":{"password":"[REDACTED]"},"password":"[REDACTED]"}`, model.RedactionApplied).pruefe(t)
	schlecht("treffer-mal-duplikat", `{"password":"S3CR3T-7f","password":1}`, `{"password":"[REDACTED]","password":`, model.GrundNichtstring).pruefe(t)
}

func TestBodyWertarten(t *testing.T) {
	for _, n := range []string{
		"null-treffer", "true-treffer", "false-treffer", "nichtstring-zahl", "nichtstring-objekt", "nichtstring-array",
		"nichtstring-aussen", "nichtstring-zahl-geheim", "syntax-wert-fehlt-treffer", "syntax-x-treffer",
	} {
		fall(t, n).pruefe(t)
	}
	schlecht("nichtstring-negativ-exponent", `{"password":-1.5e3}`, `{"password":`, model.GrundNichtstring).pruefe(t)
	schlecht("nichtstring-nach-zahl", `{"a":1,"password":0}`, `{"a":1,"password":`, model.GrundNichtstring).pruefe(t)
	schlecht("nichtstring-fuehrende-null", `{"password":01}`, `{"password":`, model.GrundNichtstring).pruefe(t)
	schlecht("nichtstring-mit-leerraum", `{"password": 1}`, `{"password": `, model.GrundNichtstring).pruefe(t)
	schlecht("syntax-fehlender-wert-vor-komma", `{"password":,}`, `{"password":`, model.GrundSyntax).pruefe(t)
	schlecht("syntax-literal-gross", `{"password":True}`, `{"password":`, model.GrundSyntax).pruefe(t)
	schlecht("syntax-literal-unvollstaendig", `{"password":nul}`, `{"password":nul`, model.GrundSyntax).pruefe(t)
}

func TestBodyMarkerIdempotent(t *testing.T) {
	for _, n := range []string{
		"marker-schon-da", "marker-beide", "marker-escape", "marker-mit-rest", "marker-mit-leerzeichen",
		"marker-abgeschnitten", "marker-ohne-klammer", "marker-mit-vorsatz", "leerer-string-ersetzt",
		"marker-gleiche-laenge", "marker-gleiche-laenge-letztes-byte",
	} {
		fall(t, n).pruefe(t)
	}
}

func TestBodyStatus(t *testing.T) {
	for _, f := range bodyFaelle() {
		g := f.fuehre(t, einmal(f.ein))
		if g.status != f.status {
			t.Errorf("%s: Status %q, erwartet %q", f.name, g.status, f.status)
		}
	}
	for _, n := range []string{"null-treffer", "ohne-treffer", "ohne-pfade", "marker-schon-da"} {
		if f := fall(t, n); f.status != model.RedactionNone || f.aus != f.ein {
			t.Errorf("%s: Erwartung der Tabelle ist nicht none mit gleicher Ausgabe", n)
		}
	}
}

// syntaxWert ist ein Wert V in `{"p":V}` und `{"password":V}`: vorher ist, was von V unter einem
// nicht treffenden Namen in der Ausgabe steht; unter einem treffenden Namen steht davon nichts
// (ein String wird nicht geschrieben), außer den Bytes eines Literals, die durchgehen.
type syntaxWert struct {
	v          string
	vorher     string
	istString  bool // V ist der Inhalt eines Strings (in Anführungszeichen)
	treffer    model.UnsicherGrund
	trefferAus string
}

func TestBodySyntax(t *testing.T) {
	werte := []syntaxWert{
		{v: `x\y`, vorher: `x\`, istString: true},
		{v: `\u12G4`, vorher: `\u12`, istString: true},
		{v: `\u00`, vorher: `\u00`, istString: true},
		{v: "x\ny", vorher: "x", istString: true},
		{v: "x\ty", vorher: "x", istString: true},
		{v: "x\x1fy", vorher: "x", istString: true},
		{v: `01`, vorher: `0`, treffer: model.GrundNichtstring},
		{v: `1.`, vorher: `1.`, treffer: model.GrundNichtstring},
		{v: `.5`, vorher: ``, treffer: model.GrundSyntax},
		{v: `+1`, vorher: ``, treffer: model.GrundSyntax},
		{v: `1e`, vorher: `1e`, treffer: model.GrundNichtstring},
		{v: `1e+`, vorher: `1e+`, treffer: model.GrundNichtstring},
		{v: `-`, vorher: `-`, treffer: model.GrundNichtstring},
		{v: `-x`, vorher: `-`, treffer: model.GrundNichtstring},
		{v: `NaN`, vorher: ``, treffer: model.GrundSyntax},
		{v: `Infinity`, vorher: ``, treffer: model.GrundSyntax},
		{v: `True`, vorher: ``, treffer: model.GrundSyntax},
		{v: `nul`, vorher: `nul`, treffer: model.GrundSyntax, trefferAus: `nul`},
		{v: `nulx`, vorher: `nul`, treffer: model.GrundSyntax, trefferAus: `nul`},
		{v: `truee`, vorher: `true`, treffer: model.GrundSyntax, trefferAus: `true`},
	}
	for _, w := range werte {
		v, vorher, grund := w.v, w.vorher, w.treffer
		if w.istString {
			v, vorher, grund = `"`+w.v+`"`, `"`+w.vorher, model.GrundSyntax
		}
		schlecht("syntax "+kurz(w.v)+" unter p", `{"p":`+v+`}`, `{"p":`+vorher, model.GrundSyntax).pruefe(t)
		schlecht("syntax "+kurz(w.v)+" unter password", `{"password":`+v+`}`, `{"password":`+w.trefferAus, grund).pruefe(t)
	}
	dokumente := []struct{ ein, aus string }{
		{`{"a":1,}`, `{"a":1,`}, {`[1,]`, `[1,`}, {`{,}`, `{`}, {`{"a" 1}`, `{"a" `}, {`{a:1}`, `{`}, {`{'a':1}`, `{`},
		{`{"a":1}{"b":2}`, `{"a":1}`}, {`{"a":1} x`, `{"a":1} `}, {"{}\n{}", "{}\n"}, {`"a" "b"`, `"a" `}, {`1 2`, `1 `},
		{"\xEF\xBB\xBF{}", ""}, {"{\"a\":1}\x00", `{"a":1}`}, {"{\f}", "{"}, {"{\xC2\xA0}", "{"},
		{"\xFF\xFE{\x00}\x00", ""}, {`{"a":1 "b":2}`, `{"a":1 `}, {`[1 2]`, `[1 `}, {`{"a":1]`, `{"a":1`}, {`[1}`, `[1`},
		{`{"a":1,,}`, `{"a":1,`}, {`]`, ``}, {`}`, ``}, {`:`, ``}, {`,`, ``}, {`{"a":}`, `{"a":`}, {`{"a"}`, `{"a"`},
		{`[,1]`, `[`}, {`{"a":1,"b"}`, `{"a":1,"b"`}, {`{"a":{}}x`, `{"a":{}}`},
		{`{"a":-01}`, `{"a":-0`}, {`{"a":1.5.}`, `{"a":1.5`}, {`{"a":1e5e5}`, `{"a":1e5`}, {`[}`, `[`}, {`{]`, `{`},
	}
	for _, d := range dokumente {
		schlecht("syntax-dokument "+kurz(d.ein), d.ein, d.aus, model.GrundSyntax).pruefe(t)
	}
	gutFaelle := []bodyFall{
		gut("leerraum-nachlauf", "{} \n\t\r", "{} \n\t\r", model.RedactionNone),
		gut("del-im-string", "{\"a\":\"\x7f\"}", "{\"a\":\"\x7f\"}", model.RedactionNone),
		gut("nul-escape", `{"a":"\u0000"}`, `{"a":"\u0000"}`, model.RedactionNone),
		gut("einzelnes-surrogat", `{"\ud800":"\ud800"}`, `{"\ud800":"\ud800"}`, model.RedactionNone),
		gut("einzelnes-surrogat-ersetzt", `{"password":"\ud800"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied),
		gut("null-punkt-fuenf", `{"a":0.5}`, `{"a":0.5}`, model.RedactionNone),
		gut("minus-null-punkt-fuenf", `{"a":-0.5}`, `{"a":-0.5}`, model.RedactionNone),
		gut("null-exponent", `{"a":0e5}`, `{"a":0e5}`, model.RedactionNone),
		gut("exponent-minus", `{"a":1e-5}`, `{"a":1e-5}`, model.RedactionNone),
		gut("alle-escapes", `["\"","\\","\/","\b","\f","\n","\r","\t","\u00E9"]`, `["\"","\\","\/","\b","\f","\n","\r","\t","\u00E9"]`, model.RedactionNone),
	}
	for _, f := range gutFaelle {
		f.pruefe(t)
	}
}

func TestBodyUTF8(t *testing.T) {
	gueltig := map[string]string{"é": "\xC3\xA9", "€": "\xE2\x82\xAC", "😀": "\xF0\x9F\x98\x80", "DEL": "\x7F", "U+FFFF": "\xEF\xBF\xBF", "U+10FFFF": "\xF4\x8F\xBF\xBF",
		"U+80": "\xC2\x80", "U+7FF": "\xDF\xBF", "U+800": "\xE0\xA0\x80", "U+D7FF": "\xED\x9F\xBF", "U+E000": "\xEE\x80\x80", "U+10000": "\xF0\x90\x80\x80"}
	for n, v := range gueltig {
		gut("name "+n, `{"`+v+`":1}`, `{"`+v+`":1}`, model.RedactionNone).pruefe(t)
		gut("ersetzter wert "+n, `{"password":"`+v+`"}`, `{"password":"[REDACTED]"}`, model.RedactionApplied).pruefe(t)
		gut("anderer wert "+n, `{"p":"`+v+`"}`, `{"p":"`+v+`"}`, model.RedactionNone).pruefe(t)
	}
	ungueltig := map[string]struct{ v, bis string }{
		"C0 80": {"\xC0\x80", ""}, "C1 BF": {"\xC1\xBF", ""}, "E0 80 80": {"\xE0\x80\x80", "\xE0"},
		"ED A0 80": {"\xED\xA0\x80", "\xED"}, "F4 90 80 80": {"\xF4\x90\x80\x80", "\xF4"}, "80": {"\x80", ""},
		"FF": {"\xFF", ""}, "E2 82 vor Anführungszeichen": {"\xE2\x82", "\xE2\x82"}, "F5": {"\xF5\x80\x80\x80", ""},
		"F0 80 80 80": {"\xF0\x80\x80\x80", "\xF0"}, "F0 8F BF BF": {"\xF0\x8F\xBF\xBF", "\xF0"}, "E2 28 A1": {"\xE2\x28\xA1", "\xE2"},
		"C2 vor ASCII": {"\xC2x", "\xC2"}, "DF 7F": {"\xDF\x7F", "\xDF"}, "EF BF 7F": {"\xEF\xBF\x7F", "\xEF\xBF"},
		"C2 C0": {"\xC2\xC0", "\xC2"}, "C2 FF": {"\xC2\xFF", "\xC2"}, "E1 C0": {"\xE1\xC0", "\xE1"},
		"E1 80 C0": {"\xE1\x80\xC0", "\xE1\x80"}, "F1 80 80 C0": {"\xF1\x80\x80\xC0", "\xF1\x80\x80"},
	}
	for n, u := range ungueltig {
		schlecht("name "+n, `{"`+u.v+`":1}`, `{"`+u.bis, model.GrundUTF8).pruefe(t)
		schlecht("ersetzter wert "+n, `{"password":"`+u.v+`"}`, `{"password":`, model.GrundUTF8).pruefe(t)
		schlecht("anderer wert "+n, `{"p":"`+u.v+`"}`, `{"p":"`+u.bis, model.GrundUTF8).pruefe(t)
	}
}

func TestBodyTiefe(t *testing.T) {
	objekte := func(n int) string { return wiederhole(`{"a":`, n) + `1` + wiederhole(`}`, n) }
	gemischt := func(n int) (string, string) {
		auf, zu := "", ""
		for i := 0; i < n; i++ {
			if i%2 == 0 {
				auf, zu = auf+"[", "]"+zu
			} else {
				auf, zu = auf+`{"a":`, "}"+zu
			}
		}
		return auf, zu
	}
	arrays := wiederhole("[", 512) + wiederhole("]", 512)
	gut("arrays 512", arrays, arrays, model.RedactionNone).pruefe(t)
	gut("objekte 512", objekte(512), objekte(512), model.RedactionNone).pruefe(t)
	auf512, zu512 := gemischt(512)
	gut("gemischt 512", auf512+"0"+zu512, auf512+"0"+zu512, model.RedactionNone).pruefe(t)
	schlecht("arrays 513", wiederhole("[", 513)+wiederhole("]", 513), wiederhole("[", 512), model.GrundTiefe).pruefe(t)
	schlecht("objekte 513", objekte(513), wiederhole(`{"a":`, 512), model.GrundTiefe).pruefe(t)
	auf, zu := gemischt(513)
	schlecht("gemischt 513", auf+"0"+zu, wiederhole("[{\"a\":", 256), model.GrundTiefe).pruefe(t)
	schlecht("513 offene ohne Schluss", wiederhole("[", 513), wiederhole("[", 512), model.GrundTiefe).pruefe(t)
	schlecht("513 offene, danach x", wiederhole("[", 513)+"x", wiederhole("[", 512), model.GrundTiefe).pruefe(t)
	fall(t, "treffer-bei-tiefe-512").pruefe(t)
	fall(t, "tiefe-513-treffer").pruefe(t)
}

func TestBodyPraefixe(t *testing.T) {
	zaehler := 0
	for _, f := range bodyFaelle() {
		wurzel := strings.TrimLeft(f.ein, " \t\r\n")
		if f.grund != "" || f.ctGesetzt || wurzel == "" || (wurzel[0] != '{' && wurzel[0] != '[') {
			continue
		}
		ende := len(strings.TrimRight(f.ein, " \t\r\n"))
		for i := 1; i < ende; i++ {
			g := f.fuehre(t, einmal(f.ein[:i]))
			var uf *model.UnsicherFehler
			if !errors.As(g.err, &uf) || uf.Grund != model.GrundUnvollstaendig || g.status != model.RedactionUnsafe {
				t.Fatalf("%s: Präfix von %d Bytes ergibt Status %q, Fehler %v", f.name, i, g.status, g.err)
			}
			zaehler++
		}
	}
	if zaehler < 500 {
		t.Errorf("nur %d Präfixe geprüft", zaehler)
	}
	unvollstaendig := []string{`"abc`, `tru`, `-`, `1.`, `1e`, `{"a"`, `{"a":`, `[1,`, `[1`, `{"a":1`, `{"a":"b"`, `[`, `{`, `{"a":1,`, `-1e+`, `fals`, `nu`}
	for _, s := range unvollstaendig {
		schlecht("praefix "+s, s, s, model.GrundUnvollstaendig).pruefe(t)
	}
	for _, s := range []string{`12`, `"abc"`, `true`, `-0`, `1e5`, `null`, `0`, `1.5`} {
		gut("ganzer skalar "+s, s, s, model.RedactionNone).pruefe(t)
	}
}

func TestBodyZerlegung(t *testing.T) {
	geprueft := map[string]bool{}
	for _, f := range bodyFaelle() {
		if len(f.ein) > 200 || f.ein == "" {
			continue
		}
		geprueft[f.name] = true
		for _, z := range zerlegungen(f.ein) {
			f.vergleiche(t, f.fuehre(t, z.leser()), z.name)
		}
	}
	// Fälle der Zustände, die an einer Chunk-Grenze enden können.
	for _, n := range []string{
		"name-escape", "escape-im-wert", "unicode-escape-im-wert", "mehrbyte-ueberall", "escape-und-surrogat", "literale-und-zahlen",
		"marker-schon-da", "marker-escape", "name-ueber-musterlaenge", "layout-bleibt", "utf8-name", "utf8-anderer-wert", "edmx-bom-leerraum", "zahlen-durchgereicht",
		"nichtstring-objekt", "syntax-hex-im-ersetzten", "typ-text", "unvollstaendig-wert-geheim",
	} {
		if !geprueft[n] {
			t.Errorf("Fall %q fehlt in der Zerlegung", n)
		}
	}
}

func TestBodyGrenze(t *testing.T) {
	voll16 := `{"a":"bbbbbbbb"}`
	if len(voll16) != 16 {
		t.Fatalf("Fixture hat %d Bytes", len(voll16))
	}
	erweitert := `{"password": ""}`
	if len(erweitert) != 16 {
		t.Fatalf("Fixture hat %d Bytes", len(erweitert))
	}
	fenster := "[" + wiederhole("1,", 7) + "1" // 16 Bytes
	faelle := []bodyFall{
		gut("16 Bytes ohne Ersatz", voll16, voll16, model.RedactionNone),
		gut("16 Bytes mit Verlängerung", erweitert, `{"password": "[REDACTED]"}`, model.RedactionApplied),
		schlecht("17 Bytes gültig", `{"a":"bbbbbbbbb"}`, `{"a":"bbbbbbbbb"`, model.GrundZuGross),
		schlecht("40 Bytes gültiger Anfang", "["+wiederhole("1,", 20), fenster, model.GrundZuGross),
		schlecht("16 Bytes unvollständig", fenster, fenster, model.GrundUnvollstaendig),
		schlecht("Syntaxfehler an Byte 17", fenster+"x", fenster, model.GrundZuGross),
		schlecht("Syntaxfehler an Byte 16", fenster[:15]+"x", fenster[:15], model.GrundSyntax),
		gut("EDMX 16 Bytes", "<a>"+wiederhole("x", 9)+"</a>", "<a>"+wiederhole("x", 9)+"</a>", model.RedactionNone).edmx(),
		schlecht("EDMX 17 Bytes", "<a>"+wiederhole("x", 10)+"</a>", "<a>"+wiederhole("x", 10)+"</a", model.GrundZuGross).edmx(),
		schlecht("EDMX nur Leerraum über der Grenze", wiederhole(" ", 17), wiederhole(" ", 16), model.GrundZuGross).edmx(),
		// Ein Verstoß vor Byte MaxBytes gewinnt gegen zu_gross, auch wenn derselbe Read die Grenze überschreitet.
		schlecht("Syntaxfehler vor der Grenze, Strom darüber", `{"a":x,"b":"cccccccccccc"}`, `{"a":`, model.GrundSyntax),
		schlecht("utf8 vor der Grenze, Strom darüber", "{\"a\":\"\xFF\",\"b\":\"cccccccc\"}", `{"a":"`, model.GrundUTF8),
		schlecht("nichtstring vor der Grenze, Strom darüber", `{"password":12,"b":"cc"}`, `{"password":`, model.GrundNichtstring),
		schlecht("mismatch vor der Grenze, Strom darüber", "x"+wiederhole("y", 30), ``, model.GrundMismatch).edmx(),
	}
	for _, f := range faelle {
		f = f.mitMax(16)
		f.pruefe(t)
		for _, z := range zerlegungen(f.ein) {
			f.vergleiche(t, f.fuehre(t, z.leser()), z.name)
		}
	}
	auf := wiederhole("[", 513) + wiederhole(" ", 200)
	schlecht("Tiefe vor der Grenze, Strom darüber", auf, wiederhole("[", 512), model.GrundTiefe).mitMax(600).pruefe(t)
	gut("MaxBytes 1, ein Byte", `1`, `1`, model.RedactionNone).mitMax(1).pruefe(t)
	schlecht("MaxBytes 1, unvollständig", `{`, `{`, model.GrundUnvollstaendig).mitMax(1).pruefe(t)
	schlecht("MaxBytes 1, zwei Bytes", `{}`, `{`, model.GrundZuGross).mitMax(1).pruefe(t)
	// Die größte Grenze läuft nicht über: MaxBytes + 1 wäre kein int64.
	if g := redigiere(t, einmal(voll16), jsonEingang(), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: math.MaxInt64}); g.err != nil || g.aus != voll16 || g.status != model.RedactionNone {
		t.Errorf("MaxBytes math.MaxInt64: %+v", g)
	}
	// Der Zähler: bei 1 MiB Angebot liest die Funktion höchstens MaxBytes + 1 Bytes.
	r := einmal(wiederhole("[", 1<<20))
	g := redigiere(t, r, jsonEingang(), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 16})
	if r.gelesen > 17 {
		t.Errorf("%d Bytes gelesen, erwartet höchstens 17", r.gelesen)
	}
	if g.err == nil {
		t.Errorf("kein Fehler bei 1 MiB über der Grenze")
	}
	r = teile(false, false, einzelneBytes(wiederhole("[", 100))...)
	redigiere(t, r, jsonEingang(), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 16})
	if r.gelesen > 17 {
		t.Errorf("einzelne Bytes: %d gelesen, erwartet höchstens 17", r.gelesen)
	}
}

func messe(t *testing.T, name string, src io.Reader, status model.RedactionStatus, aus int64) {
	t.Helper()
	w := &zaehlWriter{}
	var vor, nach runtime.MemStats
	runtime.ReadMemStats(&vor)
	st, err := model.RedigiereBody(w, src, jsonEingang(), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 16 << 20})
	runtime.ReadMemStats(&nach)
	if err != nil || st != status {
		t.Errorf("%s: Status %q, Fehler %v", name, st, err)
	}
	if w.n != aus {
		t.Errorf("%s: %d Bytes geschrieben, erwartet %d", name, w.n, aus)
	}
	if zuwachs := nach.TotalAlloc - vor.TotalAlloc; zuwachs >= 1<<20 {
		t.Errorf("%s: Zuwachs %d Bytes bei 9 MiB Strom", name, zuwachs)
	}
}

func TestBodySpeicher(t *testing.T) {
	const neun = 9 << 20
	marker := int64(len(`{"password":"[REDACTED]"}`))
	messe(t, "String unter nicht treffendem Namen", &genReader{kopf: `{"p":"`, muster: "a", n: neun, schluss: `"}`}, model.RedactionNone, int64(len(`{"p":""}`))+neun)
	messe(t, "String unter treffendem Namen", &genReader{kopf: `{"password":"`, muster: "a", n: neun, schluss: `"}`}, model.RedactionApplied, marker)
	messe(t, "Zahl", &genReader{kopf: `{"n":`, muster: "1", n: neun, schluss: `}`}, model.RedactionNone, int64(len(`{"n":}`))+neun)
	messe(t, "Name", &genReader{kopf: `{"`, muster: "a", n: neun, schluss: `":1}`}, model.RedactionNone, int64(len(`{"":1}`))+neun)
	messe(t, "Leerraum", &genReader{kopf: `{`, muster: " ", n: neun, schluss: `}`}, model.RedactionNone, int64(len(`{}`))+neun)
	messe(t, "Array", &genReader{kopf: `[`, muster: "0,", n: neun, schluss: `0]`}, model.RedactionNone, int64(len(`[0]`))+neun)
}

func TestBodyNamenLaenge(t *testing.T) {
	regeln := func(muster ...string) model.BodyRegeln {
		return model.BodyRegeln{Pfade: pfadeAus(t, muster...), MaxBytes: 1 << 20}
	}
	for _, c := range []struct {
		name, aus string
		status    model.RedactionStatus
	}{
		{"client_secret", `{"client_secret":"[REDACTED]"}`, model.RedactionApplied},
		{"CLIENT_SECRET", `{"CLIENT_SECRET":"[REDACTED]"}`, model.RedactionApplied},
		{"client_secre", `{"client_secre":"S3CR3T-7f"}`, model.RedactionNone},
		{"client_secrett", `{"client_secrett":"S3CR3T-7f"}`, model.RedactionNone},
		{"client_secrex", `{"client_secrex":"S3CR3T-7f"}`, model.RedactionNone},
		{wiederhole("c", 100000), `{"` + wiederhole("c", 100000) + `":"S3CR3T-7f"}`, model.RedactionNone},
		{"client_secret" + wiederhole("c", 100000), `{"client_secret` + wiederhole("c", 100000) + `":"S3CR3T-7f"}`, model.RedactionNone},
	} {
		for _, r := range []model.BodyRegeln{regeln("$..client_secret"), regeln("$..a", "$..client_secret"), regeln("$..client_secret", "$..a")} {
			g := redigiere(t, einmal(`{"`+c.name+`":"S3CR3T-7f"}`), jsonEingang(), r)
			if g.aus != c.aus || g.status != c.status || g.err != nil {
				t.Errorf("Name %.20q: %+v", c.name, ergebnis{kurz(g.aus), g.status, g.err})
			}
		}
	}
	// Ein Name wird nach einem überlangen Namen wieder verglichen.
	g := redigiere(t, einmal(`{"`+wiederhole("c", 50)+`":1,"client_secret":"x"}`), jsonEingang(), regeln("$..client_secret"))
	if g.aus != `{"`+wiederhole("c", 50)+`":1,"client_secret":"[REDACTED]"}` {
		t.Errorf("Name nach überlangem Namen: %q", g.aus)
	}
}

func TestKeinGeheimnisInAusgabe(t *testing.T) {
	geprueft := 0
	for _, f := range bodyFaelle() {
		if !f.geheim {
			continue
		}
		pruefe := func(wie string, g ergebnis) {
			if strings.Contains(g.aus, "S3CR3T-7f") {
				t.Errorf("%s [%s]: Ausgabe %q enthält das Geheimnis", f.name, wie, g.aus)
			}
			if g.err != nil && strings.Contains(g.err.Error(), "S3CR3T") {
				t.Errorf("%s [%s]: Fehlertext %q enthält das Geheimnis", f.name, wie, g.err)
			}
		}
		for _, z := range zerlegungen(f.ein) {
			pruefe(z.name, f.fuehre(t, z.leser()))
			geprueft++
		}
		for k := 0; k <= len(f.ein); k++ {
			pruefe("Eingabefehler", f.fuehre(t, &abbruchReader{daten: f.ein, k: k, err: errTest}))
			pruefe("Eingabefehler mit Daten", f.fuehre(t, &abbruchReader{daten: f.ein, k: k, err: errTest, zusammen: true}))
		}
		for bei := 1; bei <= 3; bei++ {
			w := &scheiterWriter{bei: bei, err: errTest}
			_, err := model.RedigiereBody(w, teile(false, false, einzelneBytes(f.ein)...), f.eingang(), f.regeln(t))
			pruefe("Ausgabefehler", ergebnis{w.buf.String(), "", err})
		}
	}
	if geprueft == 0 {
		t.Errorf("kein Fall mit Geheimnis")
	}
}

func TestBodyFaelleDecken(t *testing.T) {
	gruende := map[model.UnsicherGrund]bool{}
	stati := map[model.RedactionStatus]bool{}
	arten := map[string]bool{}
	namen := map[string]bool{}
	for _, f := range bodyFaelle() {
		if namen[f.name] {
			t.Errorf("Fall %q kommt zweimal vor", f.name)
		}
		namen[f.name] = true
		stati[f.status] = true
		if f.grund != "" {
			gruende[f.grund] = true
		}
		if f.geheim {
			if !strings.Contains(f.ein, "S3CR3T-7f") {
				t.Errorf("Fall %q führt das Geheimnis nicht", f.name)
			}
			for _, a := range strings.Fields(f.arten) {
				arten[a] = true
			}
		}
	}
	for _, g := range []model.UnsicherGrund{
		model.GrundTyp, model.GrundMismatch, model.GrundSyntax, model.GrundUTF8, model.GrundTiefe, model.GrundNichtstring,
		model.GrundZuGross, model.GrundUnvollstaendig,
	} {
		if !gruende[g] {
			t.Errorf("kein Fall mit Grund %q", g)
		}
	}
	for _, s := range []model.RedactionStatus{model.RedactionNone, model.RedactionApplied, model.RedactionUnsafe} {
		if !stati[s] {
			t.Errorf("kein Fall mit Status %q", s)
		}
	}
	for _, a := range []string{"string", "null", "true", "false", "number", "object", "array"} {
		if !arten[a] {
			t.Errorf("kein Fall mit Geheimnis und Wertart %q", a)
		}
	}
}

func TestBodyGrund(t *testing.T) {
	for _, f := range bodyFaelle() {
		f.vergleiche(t, f.fuehre(t, einmal(f.ein)), "ein Read")
	}
}

func TestBodyTabelle(t *testing.T) {
	for _, f := range bodyFaelle() {
		f.pruefe(t)
	}
}

func TestBodyReihenfolge(t *testing.T) {
	schlecht("nichtstring vor Komma", `{"password":1,}`, `{"password":`, model.GrundNichtstring).pruefe(t)
	fall(t, "syntax-erster-verstoss").pruefe(t)
	schlecht("utf8 vor nichtstring", "{\"b\":\"\xFF\",\"password\":1}", `{"b":"`, model.GrundUTF8).pruefe(t)
	fall(t, "tiefe-513-array").pruefe(t)
	schlecht("513 und x", wiederhole("[", 513)+"x", wiederhole("[", 512), model.GrundTiefe).pruefe(t)
	g := redigiere(t, einmal("x"), model.BodyEingang{ContentType: []string{"text/plain"}}, model.BodyRegeln{MaxBytes: 0})
	if !errors.Is(g.err, model.ErrUngueltig) {
		t.Errorf("text/plain mit MaxBytes 0: %v", g.err)
	}
	gut("leer", ``, ``, model.RedactionNone).typ("text/plain").pruefe(t)
}

func TestBodyEingabeFehler(t *testing.T) {
	dokumente := []string{`{"password":"S3CR3T-7f","a":[1,2]}`, `{"a":1}`, `{"a":`, `[true]`}
	for _, daten := range dokumente {
		for _, ursache := range []error{errTest, context.Canceled, context.DeadlineExceeded} {
			for k := 0; k <= len(daten); k++ {
				for _, zusammen := range []bool{false, true} {
					g := redigiere(t, &abbruchReader{daten: daten, k: k, err: ursache, zusammen: zusammen}, jsonEingang(), bodyRegeln(t))
					var uf *model.UnsicherFehler
					if !errors.As(g.err, &uf) || uf.Grund != model.GrundEingabe || !errors.Is(g.err, ursache) || !errors.Is(g.err, model.ErrUnsicher) || g.status != model.RedactionUnsafe {
						t.Fatalf("%q nach %d Bytes (%v, zusammen %v): Status %q, Fehler %v", daten, k, ursache, zusammen, g.status, g.err)
					}
				}
			}
		}
	}
	// Daten, die zusammen mit dem Fehler kommen, werden zuerst verarbeitet.
	g := redigiere(t, &abbruchReader{daten: `{"a":x}`, k: 7, err: errTest, zusammen: true}, jsonEingang(), bodyRegeln(t))
	var uf *model.UnsicherFehler
	if !errors.As(g.err, &uf) || uf.Grund != model.GrundSyntax {
		t.Errorf("Syntaxfehler in den Daten mit Fehler: %v", g.err)
	}
	// Auch ein EDMX-Strom endet mit eingabe.
	g = redigiere(t, &abbruchReader{daten: `<a/>`, k: 4, err: errTest}, model.BodyEingang{ContentType: []string{"application/xml"}, Metadata: true}, bodyRegeln(t))
	if !errors.As(g.err, &uf) || uf.Grund != model.GrundEingabe {
		t.Errorf("EDMX: %v", g.err)
	}
}

func TestBodyAusgabeFehler(t *testing.T) {
	dok := `{"password":"S3CR3T-7f","a":[1,2,3]}`
	ref := &zaehlWriter{}
	if _, err := model.RedigiereBody(ref, teile(false, false, einzelneBytes(dok)...), jsonEingang(), bodyRegeln(t)); err != nil || ref.writes < 20 {
		t.Fatalf("Referenzlauf: %v, %d Writes", err, ref.writes)
	}
	for _, ursache := range []error{errTest, errVoll} {
		for bei := 1; bei <= ref.writes; bei++ {
			r := teile(false, false, einzelneBytes(dok)...)
			w := &scheiterWriter{bei: bei, err: ursache, leser: r}
			st, err := model.RedigiereBody(w, r, jsonEingang(), bodyRegeln(t))
			var uf *model.UnsicherFehler
			if !errors.As(err, &uf) || uf.Grund != model.GrundAusgabe || !errors.Is(err, ursache) || st != model.RedactionUnsafe {
				t.Fatalf("Write %d: Status %q, Fehler %v", bei, st, err)
			}
			if w.writes != bei || r.reads != w.readsBeiFehler {
				t.Errorf("Write %d: %d Writes, %d Reads nach dem Fehler", bei, w.writes, r.reads-w.readsBeiFehler)
			}
		}
	}
	for bei := 1; bei <= 5; bei++ {
		r := teile(false, false, einzelneBytes(dok)...)
		w := &scheiterWriter{bei: bei, kurz: true, leser: r}
		_, err := model.RedigiereBody(w, r, jsonEingang(), bodyRegeln(t))
		var uf *model.UnsicherFehler
		if !errors.As(err, &uf) || uf.Grund != model.GrundAusgabe || !errors.Is(err, io.ErrShortWrite) || w.writes != bei || r.reads != w.readsBeiFehler {
			t.Errorf("kurzer Write %d: %v, %d Writes", bei, err, w.writes)
		}
	}
	// Ein Body über der Größe eines Chunks wird in mehreren Reads gelesen; der Fehler des zweiten
	// Writes (am Ende des zweiten Chunks) beendet den Lauf, bevor der Rest gelesen oder geschrieben
	// wird. Den Fehler bei vollem Puffer mitten im Chunk fängt TestBodyAusgabeVoll.
	gross := "[" + wiederhole(`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",`, 3000) + `"z"]`
	ref = &zaehlWriter{}
	if _, err := model.RedigiereBody(ref, einmal(gross), jsonEingang(), bodyRegeln(t)); err != nil || ref.writes < 3 {
		t.Fatalf("Referenzlauf: %v, %d Writes", err, ref.writes)
	}
	r := einmal(gross)
	w := &scheiterWriter{bei: 2, err: errTest, leser: r}
	_, err := model.RedigiereBody(w, r, jsonEingang(), bodyRegeln(t))
	var uf *model.UnsicherFehler
	if !errors.As(err, &uf) || uf.Grund != model.GrundAusgabe || !errors.Is(err, errTest) || w.writes != 2 || r.reads != w.readsBeiFehler {
		t.Errorf("großer Body: %v, %d Writes, %d Reads nach dem Fehler", err, w.writes, r.reads-w.readsBeiFehler)
	}
}

func TestBodyOhneZustand(t *testing.T) {
	regeln := bodyRegeln(t)
	docs := []string{`{"password":"S3CR3T-7f","x":1,"secret":"abc"}`, `{"x":1,"y":[true,"z"]}`, `{"pass\u0077ord":"é"}`}
	frisch, vorher := make([]ergebnis, len(docs)), make([]ergebnis, len(docs))
	for i, d := range docs {
		frisch[i] = redigiere(t, einmal(d), jsonEingang(), bodyRegeln(t))
		vorher[i] = redigiere(t, einmal(d), jsonEingang(), regeln)
	}
	abbrueche := []string{
		`{"password":"S3CR3T-`, `{"passw`, `{"a":[1,{"secret":"S3CR3`, `{"x":"\u00`, "{\"x\":\"\xE2\x82", `{"password":"a","y`,
		`{"password":"[REDACTED]","k":tru`, `{"a":[[[[[1`, `{"a":-1.5e`, `{"a":"\`,
	}
	for _, abbruch := range abbrueche {
		a := redigiere(t, &abbruchReader{daten: abbruch, k: len(abbruch), err: errTest}, jsonEingang(), regeln)
		if a.err == nil {
			t.Fatalf("Abbruch %q ohne Fehler", abbruch)
		}
		for i, d := range docs {
			if b := redigiere(t, einmal(d), jsonEingang(), regeln); b != vorher[i] || b != frisch[i] {
				t.Errorf("nach Abbruch %q, Dokument %q: %+v, vorher %+v", abbruch, d, b, vorher[i])
			}
		}
	}
	if frisch[0].aus != `{"password":"[REDACTED]","x":1,"secret":"[REDACTED]"}` || frisch[0].status != model.RedactionApplied ||
		frisch[1].status != model.RedactionNone || frisch[2].status != model.RedactionApplied {
		t.Errorf("Referenz: %+v", frisch)
	}
}

func TestBodyIdempotent(t *testing.T) {
	zaehler := 0
	for _, f := range bodyFaelle() {
		if f.grund != "" {
			continue
		}
		g := f.fuehre(t, einmal(f.aus))
		if g.aus != f.aus || g.status != model.RedactionNone || g.err != nil {
			t.Errorf("%s: erneute Anwendung ergibt %+v", f.name, g)
		}
		zaehler++
	}
	if zaehler < 20 {
		t.Errorf("nur %d Fälle geprüft", zaehler)
	}
}

func TestBodyFixturesSPEC053(t *testing.T) {
	fixtures := map[string][]string{
		"Token über Chunk-Grenzen":   {"string-ersetzt", "name-escape", "mehrbyte-ueberall", "escape-im-wert"},
		"MIME-Mismatch":              {"edmx-json-body", "syntax-xml-unter-json", "typ-text", "typ-multipart"},
		"Abbruch mitten im Dokument": {"unvollstaendig-wert-geheim", "unvollstaendig-name", "unvollstaendig-array"},
		"Redaction-Fehler":           {"nichtstring-objekt", "nichtstring-array", "nichtstring-zahl-geheim", "utf8-ersetzter-wert"},
	}
	for name, faelle := range fixtures {
		for _, n := range faelle {
			f := fall(t, n)
			if f.geheim && strings.Contains(f.aus, "S3CR3T-7f") {
				t.Errorf("%s: Fall %q erwartet das Geheimnis in der Ausgabe", name, n)
			}
			f.pruefe(t)
		}
	}
}

// --- Fänger zu den Findings des Reviews --------------------------------------------------------

func grundVon(err error) model.UnsicherGrund {
	var uf *model.UnsicherFehler
	if errors.As(err, &uf) {
		return uf.Grund
	}
	return ""
}

// TestBodyZahlAnfang: unter einem treffenden Namen ist jede Zahl `nichtstring`, entschieden am
// ersten Byte (SPEC-118 Nr. 5), für jede Ziffer und für `-`; unter einem nicht treffenden Namen
// geht dieselbe Zahl durch.
func TestBodyZahlAnfang(t *testing.T) {
	for _, v := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "-1"} {
		schlecht("Zahl "+v+" unter password", `{"password":`+v+`}`, `{"password":`, model.GrundNichtstring).pruefe(t)
		schlecht("Zahl "+v+"46 unter password", `{"password":`+v+`46}`, `{"password":`, model.GrundNichtstring).pruefe(t)
		gut("Zahl "+v+" unter p", `{"p":`+v+`}`, `{"p":`+v+`}`, model.RedactionNone).pruefe(t)
	}
}

// TestBodyNamenEscapes: der Name wird vor dem Vergleich dekodiert. Ein kurzer Escape (`\b \f \n
// \r \t`) ist ein Steuerzeichen und nicht der Buchstabe hinter dem Backslash; ein `\u` ist der
// ganze Codepunkt und nicht eines seiner Bytes. Die Muster stehen so, dass der falsch
// dekodierte Name träfe.
func TestBodyNamenEscapes(t *testing.T) {
	pfade := []string{"$..ab", "$..af", "$..an", "$..ar", "$..at", "$..password"}
	keine := []string{`a\b`, `a\f`, `a\n`, `a\r`, `a\t`, `pass\"word`, `pass\\word`, `pass\/word`, `\u0170assword`, `\u7000assword`, `\u0170\u0061ssword`}
	for _, n := range keine {
		doc := `{"` + n + `":"x"}`
		gut("kein treffer "+n, doc, doc, model.RedactionNone).mitPfade(pfade...).pruefe(t)
	}
	treffer := []string{`a\u0062`, `a\u0066`, `a\u006e`, `a\u0072`, `a\u0074`, `\u0070assword`}
	for _, n := range treffer {
		gut("treffer "+n, `{"`+n+`":"x"}`, `{"`+n+`":"[REDACTED]"}`, model.RedactionApplied).mitPfade(pfade...).pruefe(t)
	}
}

// TestBodyFehlertext: Error() nennt den Code und den Grund, nie den Text der Ursache und nie
// ein Byte der Eingabe (SPEC-118 Nr. 7).
func TestBodyFehlertext(t *testing.T) {
	pruefe := func(wie string, err error, grund model.UnsicherGrund, verboten ...string) {
		t.Helper()
		if grundVon(err) != grund {
			t.Errorf("%s: Grund %q, erwartet %q", wie, grundVon(err), grund)
			return
		}
		txt := err.Error()
		for _, soll := range []string{"ODR-E2401", string(grund)} {
			if !strings.Contains(txt, soll) {
				t.Errorf("%s: Text %q nennt %q nicht", wie, txt, soll)
			}
		}
		for _, v := range append([]string{"S3CR3T", "QXZJ"}, verboten...) {
			if strings.Contains(txt, v) {
				t.Errorf("%s: Text %q enthält %q", wie, txt, v)
			}
		}
	}
	for _, f := range bodyFaelle() {
		if f.grund != "" {
			pruefe(f.name, f.fuehre(t, einmal(f.ein)).err, f.grund)
		}
	}
	for doc, grund := range map[string]model.UnsicherGrund{
		`{"QXZJ":x}`:          model.GrundSyntax,
		"{\"QXZJ\":\"\xFF\"}": model.GrundUTF8,
		`{"password":QXZJ}`:   model.GrundSyntax,
		`{"password":1QXZJ}`:  model.GrundNichtstring,
		`{"password":"QXZJ`:   model.GrundUnvollstaendig,
	} {
		pruefe(kurz(doc), redigiere(t, einmal(doc), jsonEingang(), bodyRegeln(t)).err, grund)
	}
	quelle, senke := errors.New("ursache-quelle-kaputt"), errors.New("ursache-senke-kaputt")
	g := redigiere(t, &abbruchReader{daten: `{"a":`, k: 5, err: quelle}, jsonEingang(), bodyRegeln(t))
	pruefe("eingabe", g.err, model.GrundEingabe, "ursache", "kaputt")
	if !errors.Is(g.err, quelle) {
		t.Errorf("eingabe: Ursache nicht erreichbar")
	}
	_, err := model.RedigiereBody(&scheiterWriter{bei: 1, err: senke}, einmal(`{"a":1}`), jsonEingang(), bodyRegeln(t))
	pruefe("ausgabe", err, model.GrundAusgabe, "ursache", "kaputt")
	if !errors.Is(err, senke) {
		t.Errorf("ausgabe: Ursache nicht erreichbar")
	}
}

// TestBodyRangfolge: entschieden ist, was der Ablauf zuerst trifft (SPEC-118 Nr. 7, Rangfolge
// der Ursachen): ein Read, seine Daten, das Leeren nach dst, erst danach ein mit den Daten
// gelieferter Lesefehler.
func TestBodyRangfolge(t *testing.T) {
	quelle, senke := errors.New("quelle kaputt"), errors.New("senke kaputt")
	for _, zusammen := range []bool{false, true} {
		// dst scheitert beim Leeren der Daten: ausgabe, der Lesefehler wird nicht ausgewertet.
		w := &scheiterWriter{bei: 1, err: senke}
		_, err := model.RedigiereBody(w, &abbruchReader{daten: `{"a":1,`, k: 7, err: quelle, zusammen: zusammen}, jsonEingang(), bodyRegeln(t))
		if grundVon(err) != model.GrundAusgabe || !errors.Is(err, senke) || errors.Is(err, quelle) || w.writes != 1 {
			t.Errorf("Lesefehler und dst-Fehler (zusammen %v): %v, %d Writes", zusammen, err, w.writes)
		}
		// Ohne Fehler von dst gilt der Lesefehler.
		g := redigiere(t, &abbruchReader{daten: `{"a":1,`, k: 7, err: quelle, zusammen: zusammen}, jsonEingang(), bodyRegeln(t))
		if grundVon(g.err) != model.GrundEingabe || !errors.Is(g.err, quelle) || g.aus != `{"a":1,` {
			t.Errorf("Lesefehler allein (zusammen %v): %+v", zusammen, g)
		}
		// Ein Verstoß in den Daten schlägt den Lesefehler und den Fehler von dst beim Leeren;
		// der gepufferte Teil geht ein letztes Mal an dst (Nr. 9).
		w = &scheiterWriter{bei: 1, err: senke}
		_, err = model.RedigiereBody(w, &abbruchReader{daten: `{"a":x}`, k: 7, err: quelle, zusammen: zusammen}, jsonEingang(), bodyRegeln(t))
		if grundVon(err) != model.GrundSyntax || errors.Is(err, senke) || errors.Is(err, quelle) || w.writes != 1 {
			t.Errorf("Verstoß, Lesefehler und dst-Fehler (zusammen %v): %v, %d Writes", zusammen, err, w.writes)
		}
	}
	// zu_gross kommt nach den MaxBytes Bytes und vor dem Leeren.
	zuLang := "[" + wiederhole("1,", 20)
	w := &scheiterWriter{bei: 1, err: senke}
	_, err := model.RedigiereBody(w, einmal(zuLang), jsonEingang(), model.BodyRegeln{Pfade: standardPfade(t), MaxBytes: 16})
	if grundVon(err) != model.GrundZuGross || errors.Is(err, senke) || w.writes != 1 {
		t.Errorf("zu_gross und dst-Fehler: %v, %d Writes", err, w.writes)
	}
	// Ist der Puffer voll, bevor das Verstoßbyte kommt, ist der Fehler von dst der erste Verstoß.
	regeln := model.BodyRegeln{Pfade: pfadeAus(t, "$..a"), MaxBytes: 1 << 20}
	r := einmal("{" + wiederhole(`"a":"",`, 3000) + "x}")
	w = &scheiterWriter{bei: 1, err: senke, leser: r}
	_, err = model.RedigiereBody(w, r, jsonEingang(), regeln)
	if grundVon(err) != model.GrundAusgabe || !errors.Is(err, senke) || w.writes != 1 || r.reads != 1 {
		t.Errorf("voller Puffer vor dem Verstoßbyte: %v, %d Writes, %d Reads", err, w.writes, r.reads)
	}
	// Mit einem dst ohne Fehler gewinnt derselbe Verstoß.
	g := redigiere(t, einmal("{"+wiederhole(`"a":"",`, 3000)+"x}"), jsonEingang(), regeln)
	if grundVon(g.err) != model.GrundSyntax || !strings.HasPrefix(g.aus, `{"a":"[REDACTED]",`) || !strings.HasSuffix(g.aus, `"a":"[REDACTED]",`) {
		t.Errorf("voller Puffer, dst ohne Fehler: Grund %q, Ausgabe %q", grundVon(g.err), kurz(g.aus))
	}
}

// TestBodyPraefixBeiFehler: bei einem Fehler von src steht in dst die Ausgabe für alle gelieferten
// Bytes, auch für die, die zusammen mit dem Fehler kamen; ein offener ersetzter String trägt
// nichts, ein Literal oder eine Mehrbyte-Folge kann halb stehen; vor dem ersten Byte steht nichts
// (SPEC-118 Nr. 9).
func TestBodyPraefixBeiFehler(t *testing.T) {
	faelle := []struct{ daten, aus string }{
		{`{"password":"S3CR3T-7f","a":1,"b"`, `{"password":"[REDACTED]","a":1,"b"`},
		{`{"password":"S3CR3T-7`, `{"password":`},
		{`{"password":"S3CR3T-7f"`, `{"password":"[REDACTED]"`},
		{"{\"a\":\"\xE2\x82", "{\"a\":\"\xE2\x82"},
		{`{"a":nu`, `{"a":nu`},
		{`{"a":-1.5e`, `{"a":-1.5e`},
		{"", ""},
	}
	for _, f := range faelle {
		for _, zusammen := range []bool{false, true} {
			// teil 3: die letzten Bytes kommen in einem späteren Read, nicht im ersten.
			for _, teil := range []int{0, 3} {
				g := redigiere(t, &abbruchReader{daten: f.daten, k: len(f.daten), err: errTest, zusammen: zusammen, teil: teil}, jsonEingang(), bodyRegeln(t))
				if g.aus != f.aus || grundVon(g.err) != model.GrundEingabe || !errors.Is(g.err, errTest) {
					t.Errorf("%q (zusammen %v, teil %d): Ausgabe %q, Grund %q, erwartet %q", f.daten, zusammen, teil, g.aus, grundVon(g.err), f.aus)
				}
			}
		}
	}
}

// langWriter nimmt alles an und merkt sich die Länge des längsten Writes.
type langWriter struct {
	bytes.Buffer
	writes, max int
}

func (w *langWriter) Write(p []byte) (int, error) {
	w.writes++
	w.max = max(w.max, len(p))
	return w.Buffer.Write(p)
}

// messLeser merkt sich die größte Länge, die ein Read anbietet.
type messLeser struct {
	r   io.Reader
	max int
}

func (m *messLeser) Read(p []byte) (int, error) {
	m.max = max(m.max, len(p))
	return m.r.Read(p)
}

func erweiterungEin(n int) string { return "{" + wiederhole(`"a":"",`, n) + `"z":1}` }
func erweiterungAus(n int) string { return "{" + wiederhole(`"a":"[REDACTED]",`, n) + `"z":1}` }

// TestBodyPuffer hält die heutigen Größen der festen Puffer aus SPEC-118 Nr. 6 für JSON: ein Read
// bietet höchstens 32 KiB an, ein Write an dst ist höchstens 32 KiB plus die Reserve eines
// Eingabebytes lang. Der Ausgabepuffer wird auch mitten in einem Chunk geleert: ein Chunk von 28 KB,
// dessen Ausgabe 68 KB lang ist, kommt in mehreren Writes an. Die Zusage der Spezifikation, die
// Unabhängigkeit von der Länge des Stroms, fängt TestBodySpeicher.
func TestBodyPuffer(t *testing.T) {
	const grenze = 32<<10 + 16
	regeln := model.BodyRegeln{Pfade: pfadeAus(t, "$..a"), MaxBytes: 1 << 20}
	w := &langWriter{}
	st, err := model.RedigiereBody(w, einmal(erweiterungEin(4000)), jsonEingang(), regeln)
	if err != nil || st != model.RedactionApplied || w.String() != erweiterungAus(4000) {
		t.Fatalf("Erweiterung: Status %q, Fehler %v, Ausgabe gleich Erwartung: %v", st, err, w.String() == erweiterungAus(4000))
	}
	if w.max > grenze || w.writes < 3 {
		t.Errorf("Erweiterung: längster Write %d Bytes, %d Writes; erwartet höchstens %d und mindestens 3 Writes", w.max, w.writes, grenze)
	}
	for _, grenzeBytes := range []int64{1 << 20, 1 << 30} {
		m := &messLeser{r: &genReader{kopf: `{"p":"`, muster: "a", n: 256 << 10, schluss: `"}`}}
		w = &langWriter{}
		st, err = model.RedigiereBody(w, m, jsonEingang(), model.BodyRegeln{Pfade: regeln.Pfade, MaxBytes: grenzeBytes})
		if err != nil || st != model.RedactionNone {
			t.Fatalf("Durchreichen: Status %q, Fehler %v", st, err)
		}
		if m.max < 1 || m.max > 32<<10 || w.max > grenze {
			t.Errorf("MaxBytes %d: größter Read %d, größter Write %d; erwartet höchstens %d und %d", grenzeBytes, m.max, w.max, 32<<10, grenze)
		}
	}
}

// TestBodyAusgabeVoll: scheitert dst, wenn der Ausgabepuffer mitten in einem Chunk voll ist,
// endet der Lauf in diesem Chunk: ein Read, zwei Writes, und dst bekommt nach dem fehlgeschlagenen
// Write nichts mehr (SPEC-118 Nr. 7, 9).
func TestBodyAusgabeVoll(t *testing.T) {
	regeln := model.BodyRegeln{Pfade: pfadeAus(t, "$..a"), MaxBytes: 1 << 20}
	r := einmal(erweiterungEin(4000))
	w := &scheiterWriter{bei: 2, err: errVoll, leser: r}
	st, err := model.RedigiereBody(w, r, jsonEingang(), regeln)
	if st != model.RedactionUnsafe || grundVon(err) != model.GrundAusgabe || !errors.Is(err, errVoll) {
		t.Fatalf("Status %q, Fehler %v", st, err)
	}
	if w.writes != 2 || r.reads != 1 || w.readsBeiFehler != 1 {
		t.Errorf("%d Writes, %d Reads (bei dem Fehler %d)", w.writes, r.reads, w.readsBeiFehler)
	}
	if got := w.buf.String(); len(got) == 0 || len(got) > 32<<10+16 || got != erweiterungAus(4000)[:len(got)] {
		t.Errorf("dst trägt %d Bytes, die kein redigiertes Präfix von höchstens 32 KiB sind", len(got))
	}
}
