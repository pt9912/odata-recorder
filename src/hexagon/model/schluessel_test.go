package model_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func schluesselRegeln() model.SchluesselRegeln {
	return model.SchluesselRegeln{
		VaryHeaders:       []string{"Accept", "Accept-Language", "Authorization", "Date", "X-Tenant"},
		VolatileHeaders:   []string{"Date", "User-Agent"},
		RedactHeaders:     standardNamen(),
		RedactQueryParams: standardParameter(),
		InvalidEncoding:   model.KodierungLiteral,
		BodyMode:          model.BodyBytes,
	}
}

func basisEingabe() model.SchluesselEingabe {
	return model.SchluesselEingabe{Method: "GET", ServiceID: "svc", Path: "/Products", RequestRedaction: model.RedactionNone}
}

func key(t *testing.T, in model.SchluesselEingabe, regeln model.SchluesselRegeln) model.Schluessel {
	t.Helper()
	s, err := model.BildeSchluessel(in, regeln)
	if err != nil {
		t.Fatalf("BildeSchluessel(%+v): %v", in, err)
	}
	return s
}

func mit(f func(*model.SchluesselEingabe)) model.SchluesselEingabe {
	in := basisEingabe()
	f(&in)
	return in
}

// gleichUndVerschieden prüft, dass a und b denselben bzw. verschiedene Keys ergeben.
func gleichUndVerschieden(t *testing.T, name string, regeln model.SchluesselRegeln, a, b model.SchluesselEingabe, gleich bool) {
	t.Helper()
	ka, kb := key(t, a, regeln).MatchKey, key(t, b, regeln).MatchKey
	if (ka == kb) != gleich {
		t.Errorf("%s: gleich=%v, erwartet %v (%s / %s)", name, ka == kb, gleich, ka, kb)
	}
}

type paarFall struct {
	name   string
	a, b   model.SchluesselEingabe
	gleich bool
	regeln func(*model.SchluesselRegeln)
}

func laufe(t *testing.T, faelle []paarFall) {
	t.Helper()
	for _, f := range faelle {
		r := schluesselRegeln()
		if f.regeln != nil {
			f.regeln(&r)
		}
		gleichUndVerschieden(t, f.name, r, f.a, f.b, f.gleich)
	}
}

func q(raw string) model.SchluesselEingabe {
	return mit(func(in *model.SchluesselEingabe) { in.RawQuery = raw })
}

func pfad(p string) model.SchluesselEingabe {
	return mit(func(in *model.SchluesselEingabe) { in.Path = p })
}

func body(b string) model.SchluesselEingabe {
	return mit(func(in *model.SchluesselEingabe) { in.Body = []byte(b) })
}

func vary(h hmap) model.SchluesselEingabe {
	return mit(func(in *model.SchluesselEingabe) { in.Header = h })
}

func kanonisch(r *model.SchluesselRegeln) { r.BodyMode = model.BodyJSONCanonical }

func TestKeyBeispieleLH(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Parameter-Reihenfolge", a: q("$top=5&$skip=10"), b: q("$skip=10&$top=5"), gleich: true},
		{name: "anderer $filter", a: q("$filter=Price%20gt%2010"), b: q("$filter=Price%20gt%2020"), gleich: false},
	})
}

func TestKeyEindeutig(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Method/ServiceID", a: mit(func(in *model.SchluesselEingabe) { in.Method, in.ServiceID = "A", "BC" }),
			b: mit(func(in *model.SchluesselEingabe) { in.Method, in.ServiceID = "AB", "C" }), gleich: false},
		{name: "ServiceID/Pfad", a: mit(func(in *model.SchluesselEingabe) { in.ServiceID, in.Path = "s", "/xa" }),
			b: mit(func(in *model.SchluesselEingabe) { in.ServiceID, in.Path = "sx", "/a" }), gleich: false},
		{name: "Pfad gegen Paar", a: mit(func(in *model.SchluesselEingabe) { in.Path, in.RawQuery = "/a", "b" }),
			b: mit(func(in *model.SchluesselEingabe) { in.Path, in.RawQuery = "/ab", "" }), gleich: false},
		{name: "Schlüssel/Wert", a: q("ab=c"), b: q("a=bc"), gleich: false},
		{name: "Vary-Wertgrenzen", a: vary(hmap{"Accept": {"a", "b"}, "X-Tenant": {"c"}}),
			b: vary(hmap{"Accept": {"a"}, "X-Tenant": {"b", "c"}}), gleich: false},
		{name: "Vary-Name/Wert", a: vary(hmap{"Accept": {"x"}}), b: vary(hmap{"Accept-Language": {"x"}}), gleich: false},
	})
}

// zeichenfolge, zahl und tupel bauen die Bytefolge aus SPEC-116 Nr. 1 von Hand.
func zeichenfolge(s string) []byte {
	return append(binary.BigEndian.AppendUint32(nil, uint32(len(s))), s...)
}

func zahl(n int) []byte { return binary.BigEndian.AppendUint32(nil, uint32(n)) }

func sha(b []byte) []byte { s := sha256.Sum256(b); return s[:] }

func handKey(teile ...[]byte) string {
	s := sha256.Sum256(bytes.Join(teile, nil))
	return "sha256:" + hex.EncodeToString(s[:])
}

func TestKeyKodierungHandgebaut(t *testing.T) {
	r := schluesselRegeln()
	in := model.SchluesselEingabe{
		Method: "GET", ServiceID: "svc", Path: "/Products(%7e)/%2f%c3%A9", RawQuery: "b&a=1&&c=%zz&token=" + geheimnis, RequestRedaction: model.RedactionNone,
		Header: hmap{"accept": {" x\t", "y"}, "Authorization": {"z"}, "X-Other": {"o"}}, Body: []byte(`{"b" : 1, "a": []}`),
	}
	r.BodyMode = model.BodyJSONCanonical
	want := handKey(
		zeichenfolge("odr-match-v1"), zeichenfolge("GET"), zeichenfolge("svc"), zeichenfolge("/Products(~)/%2F%C3%A9"),
		zahl(4),
		[]byte{0}, zeichenfolge("b"),
		[]byte{1}, zeichenfolge("a"), zeichenfolge("1"),
		[]byte{1}, zeichenfolge("token"), zeichenfolge("[REDACTED]"),
		[]byte{2}, zeichenfolge("c=%zz"),
		zahl(1), zeichenfolge("Accept"), zahl(2), zeichenfolge("x"), zeichenfolge("y"),
		[]byte{3}, sha([]byte(`{"a":[],"b":1}`)),
	)
	s := key(t, in, r)
	if s.MatchKey != want || !strings.HasPrefix(s.MatchKey, "sha256:") || len(s.MatchKey) != 71 || s.MatchKey != strings.ToLower(s.MatchKey) {
		t.Errorf("Key %s, erwartet %s", s.MatchKey, want)
	}
	// Art 1 (Bytes), Art 2 (Rückfall) und Art 0 (leer) des Bodys.
	r.BodyMode = model.BodyBytes
	in2 := model.SchluesselEingabe{Method: "POST", ServiceID: "s", Path: "/", Body: []byte("abc"), RequestRedaction: model.RedactionApplied}
	want2 := handKey(zeichenfolge("odr-match-v1"), zeichenfolge("POST"), zeichenfolge("s"), zeichenfolge("/"), zahl(0), zahl(0), []byte{1}, sha([]byte("abc")))
	if got := key(t, in2, r).MatchKey; got != want2 {
		t.Errorf("bytes: %s, erwartet %s", got, want2)
	}
	r.BodyMode = model.BodyJSONCanonical
	want3 := handKey(zeichenfolge("odr-match-v1"), zeichenfolge("POST"), zeichenfolge("s"), zeichenfolge("/"), zahl(0), zahl(0), []byte{2}, sha([]byte("abc")))
	if got := key(t, in2, r).MatchKey; got != want3 {
		t.Errorf("Rückfall: %s, erwartet %s", got, want3)
	}
	in2.Body = []byte(`"\u001F"`)
	want5 := handKey(zeichenfolge("odr-match-v1"), zeichenfolge("POST"), zeichenfolge("s"), zeichenfolge("/"), zahl(0), zahl(0), []byte{3}, sha([]byte("\"\\u001f\"")))
	if got := key(t, in2, r).MatchKey; got != want5 {
		t.Errorf("Steuerzeichen als \\u00xx klein: %s, erwartet %s", got, want5)
	}
	kopf := func(art byte, nutz []byte) string {
		return handKey(zeichenfolge("odr-match-v1"), zeichenfolge("POST"), zeichenfolge("s"), zeichenfolge("/"), zahl(0), zahl(0), []byte{art}, nutz)
	}
	in2.Body = []byte(`"\n"`)
	if got, want := key(t, in2, r).MatchKey, kopf(3, sha([]byte("\"\\u000a\""))); got != want {
		t.Errorf("LF als \\u000a klein: %s, erwartet %s", got, want)
	}
	in2.Body = []byte("\"\x7f\"")
	if got, want := key(t, in2, r).MatchKey, kopf(3, sha([]byte("\"\x7f\""))); got != want {
		t.Errorf("U+007F unverändert: %s, erwartet %s", got, want)
	}
	in2.Body = nil
	want4 := handKey(zeichenfolge("odr-match-v1"), zeichenfolge("POST"), zeichenfolge("s"), zeichenfolge("/"), zahl(0), zahl(0), []byte{0})
	if got := key(t, in2, r).MatchKey; got != want4 {
		t.Errorf("leer: %s, erwartet %s", got, want4)
	}
}

func TestKeyKopf(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Methode Schreibweise", a: mit(func(in *model.SchluesselEingabe) { in.Method = "get" }), b: basisEingabe(), gleich: false},
		{name: "Service Schreibweise", a: mit(func(in *model.SchluesselEingabe) { in.ServiceID = "SVC" }), b: basisEingabe(), gleich: false},
	})
	for name, in := range map[string]model.SchluesselEingabe{
		"Method leer":    mit(func(in *model.SchluesselEingabe) { in.Method = "" }),
		"ServiceID leer": mit(func(in *model.SchluesselEingabe) { in.ServiceID = "" }),
	} {
		if _, err := model.BildeSchluessel(in, schluesselRegeln()); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestKeyPfad(t *testing.T) {
	laufe(t, []paarFall{
		{name: "%7e/~", a: pfad("/a%7e"), b: pfad("/a~"), gleich: true},
		{name: "%2f/%2F", a: pfad("/a%2f"), b: pfad("/a%2F"), gleich: true},
		{name: "%41/A", a: pfad("/%41"), b: pfad("/A"), gleich: true},
		{name: "%2F/ /", a: pfad("/a%2Fb"), b: pfad("/a/b"), gleich: false},
		{name: "%27/'", a: pfad("/a%27"), b: pfad("/a'"), gleich: false},
		{name: "%20/+", a: pfad("/a%20b"), b: pfad("/a+b"), gleich: false},
		{name: "abschließend", a: pfad("/a"), b: pfad("/a/"), gleich: false},
		{name: "doppelt", a: pfad("/a"), b: pfad("//a"), gleich: false},
		{name: "Schreibweise", a: pfad("/a"), b: pfad("/A"), gleich: false},
		{name: "Punktsegment", a: pfad("/a/./b"), b: pfad("/a/b"), gleich: false},
		{name: "Nicht-ASCII", a: pfad("/é"), b: pfad("/%C3%A9"), gleich: false},
		{name: "Root", a: pfad("/"), b: pfad("//"), gleich: false},
		{name: "%ff/%FF", a: pfad("/a%ff"), b: pfad("/a%FF"), gleich: true},
		{name: "%ff/%FF reject", a: pfad("/a%ff"), b: pfad("/a%FF"), gleich: true, regeln: func(r *model.SchluesselRegeln) { r.InvalidEncoding = model.KodierungReject }},
	})
	for _, p := range []string{"", "a", "/a?b", "/a#b", "/a\x00", "/a\x1f", "/a\x7f", "/a\n"} {
		if _, err := model.BildeSchluessel(pfad(p), schluesselRegeln()); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%q: %v", p, err)
		}
	}
	if _, err := model.BildeSchluessel(pfad("/"), schluesselRegeln()); err != nil {
		t.Errorf("Root: %v", err)
	}
	for _, p := range []string{"/a%zz", "/a%", "/a%2", "/%G0"} {
		lit := schluesselRegeln()
		if _, err := model.BildeSchluessel(pfad(p), lit); err != nil {
			t.Errorf("literal %q: %v", p, err)
		}
		lit.InvalidEncoding = model.KodierungReject
		if _, err := model.BildeSchluessel(pfad(p), lit); !errors.Is(err, model.ErrKodierung) || errors.Is(err, model.ErrUngueltig) {
			t.Errorf("reject %q: %v", p, err)
		}
	}
	laufe(t, []paarFall{{name: "literal %zz", a: pfad("/a%zz"), b: pfad("/a%25zz"), gleich: false}})
}

func TestKeyQueryReihenfolge(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Reihenfolge", a: q("a=1&b=2"), b: q("b=2&a=1"), gleich: true},
		{name: "gleicher Schlüssel", a: q("a=1&a=2"), b: q("a=2&a=1"), gleich: true},
		{name: "Multimenge", a: q("a=1&a=1"), b: q("a=1"), gleich: false},
		{name: "Wiederholung", a: q("a=1&a=1&b"), b: q("a=1&b"), gleich: false},
		{name: "Wert", a: q("a=1"), b: q("a=2"), gleich: false},
	})
}

func TestKeyQueryZerlegung(t *testing.T) {
	laufe(t, []paarFall{
		{name: "a / a=", a: q("a"), b: q("a="), gleich: false},
		{name: "a=& / a=", a: q("a=&"), b: q("a="), gleich: true},
		{name: "&&", a: q("&&"), b: q(""), gleich: true},
		{name: "?", a: q("?"), b: q(""), gleich: false},
		{name: "leer / fehlt", a: q(""), b: basisEingabe(), gleich: true},
		{name: "=v", a: q("=v"), b: q("v"), gleich: false},
		{name: "=", a: q("="), b: q(""), gleich: false},
		{name: "a=b=c", a: q("a=b=c"), b: q("a=b&c"), gleich: false},
		{name: "a=b=c Wert", a: q("a=b=c"), b: q("a%3Db=c"), gleich: false},
		{name: ";", a: q("a=1;b=2"), b: q("a=1&b=2"), gleich: false},
	})
}

func TestKeyQueryDekodierung(t *testing.T) {
	laufe(t, []paarFall{
		{name: "+ / %20", a: q("a=b+c"), b: q("a=b%20c"), gleich: false},
		{name: "Hex-Schreibweise", a: q("a=%c3%a9"), b: q("a=%C3%A9"), gleich: true},
		{name: "%2B / +", a: q("a=%2B"), b: q("a=+"), gleich: true},
		{name: "%41 / A", a: q("a=%41"), b: q("a=A"), gleich: true},
		{name: "Schlüssel %41", a: q("%41=1"), b: q("A=1"), gleich: true},
		{name: "Schreibweise des Wertes", a: q("a=x"), b: q("a=X"), gleich: false},
		{name: "%2B / + neben %41", a: q("a=%2B%41"), b: q("a=+%41"), gleich: true},
		{name: "+ / %20 neben %41", a: q("a=b+%41"), b: q("a=b%20%41"), gleich: false},
		{name: "+ im Schlüssel neben %41", a: q("a+b%41=1"), b: q("a%20b%41=1"), gleich: false},
	})
}

func TestKeyUngueltigeKodierung(t *testing.T) {
	roh := []string{"a=%zz", "a=%", "a=%2", "a=%FF", "a=\xff", "%zz=1", "\xff=1", "%FF=1", "%"}
	for _, raw := range roh {
		lit := schluesselRegeln()
		if _, err := model.BildeSchluessel(q(raw), lit); err != nil {
			t.Errorf("literal %q: %v", raw, err)
		}
		// byteidentische Segmente sind gleich
		gleichUndVerschieden(t, "literal gleich "+raw, lit, q(raw), q(raw), true)
		rej := schluesselRegeln()
		rej.InvalidEncoding = model.KodierungReject
		_, err := model.BildeSchluessel(q(raw), rej)
		if !errors.Is(err, model.ErrKodierung) || errors.Is(err, model.ErrUngueltig) || strings.Contains(err.Error(), "zz") {
			t.Errorf("reject %q: %v", raw, err)
		}
	}
	laufe(t, []paarFall{
		{name: "%zz / %25zz", a: q("a=%zz"), b: q("a=%25zz"), gleich: false},
		{name: "%zz / %yy", a: q("a=%zz"), b: q("a=%yy"), gleich: false},
		{name: "Art 2 neben gültigen", a: q("a=%zz&b=1"), b: q("b=1&a=%zz"), gleich: true},
		{name: "Art 2 Hex-Schreibweise", a: q("a=%2f%zz"), b: q("a=%2F%zz"), gleich: false},
		{name: "Art 2 ohne gültige", a: q("a=%zz&b=1"), b: q("b=1"), gleich: false},
	})
}

func TestKeyRedigiertVorKodierung(t *testing.T) {
	rej := schluesselRegeln()
	rej.InvalidEncoding = model.KodierungReject
	for _, raw := range []string{"token=%zz", "token=\xff", "api_key=%"} {
		if _, err := model.BildeSchluessel(q(raw), rej); err != nil {
			t.Errorf("%q: %v", raw, err)
		}
	}
	gleichUndVerschieden(t, "unter Redaction", rej, q("token=%zz"), q("token=abc"), true)
	if _, err := model.BildeSchluessel(q("token=%zz&b=%zz"), rej); !errors.Is(err, model.ErrKodierung) {
		t.Errorf("anderes Paar bleibt ungültig: %v", err)
	}
}

func TestKeyVary(t *testing.T) {
	laufe(t, []paarFall{
		{name: "kein Eintrag / leere Liste", a: vary(hmap{}), b: vary(hmap{"Accept": {}}), gleich: true},
		{name: "fehlt / leerer Wert", a: vary(hmap{}), b: vary(hmap{"Accept": {""}}), gleich: false},
		{name: "Wert", a: vary(hmap{"Accept": {"a"}}), b: vary(hmap{"Accept": {"b"}}), gleich: false},
		{name: "Zeilen / Kommaliste", a: vary(hmap{"Accept": {"a", "b"}}), b: vary(hmap{"Accept": {"a, b"}}), gleich: false},
		{name: "Schreibweise des Namens", a: vary(hmap{"accept": {"a"}}), b: vary(hmap{"ACCEPT": {"a"}}), gleich: true},
		{name: "Rand-Leerraum", a: vary(hmap{"Accept": {" a \t"}}), b: vary(hmap{"Accept": {"a"}}), gleich: true},
		{name: "Innen-Leerraum", a: vary(hmap{"Accept": {"a  b"}}), b: vary(hmap{"Accept": {"a b"}}), gleich: false},
		{name: "nicht gelistet", a: vary(hmap{"X-Other": {"a"}}), b: vary(hmap{}), gleich: true},
		{name: "zwei Schreibweisen", a: vary(hmap{"accept": {"a"}, "Accept": {"b"}}), b: vary(hmap{"Accept": {"b", "a"}}), gleich: true},
		{name: "none", a: vary(hmap{"Accept": {"a"}}), b: vary(hmap{}), gleich: true, regeln: func(r *model.SchluesselRegeln) { r.VaryHeaders = nil }},
	})
	for i := 0; i < 20; i++ {
		h := hmap{"Accept": {"a"}, "Accept-Language": {"de"}, "X-Tenant": {"t"}, "accept-language": {"en"}}
		gleichUndVerschieden(t, "Map-Reihenfolge", schluesselRegeln(), vary(h), vary(h), true)
	}
}

func TestKeyVaryAusnahme(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Date", a: vary(hmap{"Date": {"a"}}), b: vary(hmap{"Date": {"b"}}), gleich: true},
		{name: "Authorization", a: vary(hmap{"Authorization": {"a"}}), b: vary(hmap{"authorization": {"b"}}), gleich: true},
		{name: "redigiert, nicht flüchtig", a: vary(hmap{"X-Tenant": {"a"}}), b: vary(hmap{"X-Tenant": {"b"}}), gleich: true,
			regeln: func(r *model.SchluesselRegeln) { r.RedactHeaders = append(r.RedactHeaders, "x-tenant") }},
		{name: "flüchtig, nicht redigiert", a: vary(hmap{"X-Tenant": {"a"}}), b: vary(hmap{"X-Tenant": {"b"}}), gleich: true,
			regeln: func(r *model.SchluesselRegeln) { r.VolatileHeaders = append(r.VolatileHeaders, "X-TENANT") }},
	})
}

func TestKeyBodyBytes(t *testing.T) {
	laufe(t, []paarFall{
		{name: "leer / nil", a: body(""), b: basisEingabe(), gleich: true},
		{name: "Byte", a: body("abc"), b: body("abd"), gleich: false},
		{name: "Leerraum", a: body("{}"), b: body("{} "), gleich: false},
		{name: "leer / Inhalt", a: body(""), b: body("0"), gleich: false},
	})
}

func TestKeyBodyKanonisch(t *testing.T) {
	laufe(t, []paarFall{
		{name: "Schlüsselreihenfolge", a: body(`{"a":1,"b":2}`), b: body(`{"b":2,"a":1}`), gleich: true, regeln: kanonisch},
		{name: "Leerraum", a: body(" {\n\"a\" : [ 1 , 2 ] } "), b: body(`{"a":[1,2]}`), gleich: true, regeln: kanonisch},
		{name: "Escape", a: body(`"\u0061"`), b: body(`"a"`), gleich: true, regeln: kanonisch},
		{name: "Escape Schrägstrich", a: body(`"\/"`), b: body(`"/"`), gleich: true, regeln: kanonisch},
		{name: "Surrogat-Paar", a: body(`"\ud83d\ude00"`), b: body("\"\U0001F600\""), gleich: true, regeln: kanonisch},
		{name: "Array-Reihenfolge", a: body(`[1,2]`), b: body(`[2,1]`), gleich: false, regeln: kanonisch},
		{name: "1 / 1.0", a: body(`1`), b: body(`1.0`), gleich: false, regeln: kanonisch},
		{name: "1 / 1e0", a: body(`1`), b: body(`1e0`), gleich: false, regeln: kanonisch},
		{name: "0 / -0", a: body(`0`), b: body(`-0`), gleich: false, regeln: kanonisch},
		{name: "doppelter Schlüssel", a: body(`{"a":1,"a":2}`), b: body(`{"a":2,"a":1}`), gleich: false, regeln: kanonisch},
		{name: "doppelter Schlüssel bleibt", a: body(`{"a":1,"a":1}`), b: body(`{"a":1}`), gleich: false, regeln: kanonisch},
		{name: "verschachtelt", a: body(`{"x":{"b":1,"a":2}}`), b: body(`{"x":{"a":2,"b":1}}`), gleich: true, regeln: kanonisch},
		{name: "Steuerzeichen", a: body(`"\n"`), b: body("\"\\u000a\""), gleich: true, regeln: kanonisch},
		{name: "true/null", a: body(`[true,false,null]`), b: body(`[ true , false , null ]`), gleich: true, regeln: kanonisch},
		{name: "512 kanonisch", a: body(strings.Repeat("[", 512) + strings.Repeat("]", 512)), b: body(" " + strings.Repeat("[", 512) + strings.Repeat("]", 512)), gleich: true, regeln: kanonisch},
		{name: "Name nach Bytes", a: body(`{"é":1,"z":2}`), b: body(`{"z":2,"é":1}`), gleich: true, regeln: kanonisch},
	})
}

func TestKeyBodyRueckfall(t *testing.T) {
	r := schluesselRegeln()
	r.BodyMode = model.BodyJSONCanonical
	for name, b := range map[string]string{
		"ungültig": `{"a":`, "BOM": "\xef\xbb\xbf{}", "Nachlauf": `{} x`, "zwei Werte": `{}{}`,
		"einzelnes Surrogat": `"\ud800"`, "Low-Surrogat": `"\udc00"`, "Low-Surrogat mit Low": `"\udc00\udc00"`, "High ohne Low": `"\ud800\u0041"`, "ungültiges UTF-8": "\"\xff\"",
		"Tiefe 513": strings.Repeat("[", 513) + strings.Repeat("]", 513), "führende Null": `01`, "Komma am Ende": `[1,]`,
		"Steuerzeichen roh": "\"\n\"", "unbekanntes Escape": `"\x"`, "Zahl ohne Ziffern": `-`, "Punkt ohne Ziffer": `1.`, "Schlüssel ohne Anführung": `{a:1}`,
		"Surrogat und doppelter Schlüssel": `{"a":"\ud800","a":1}`, "nur Leerraum": ` `,
	} {
		in := body(b)
		want := handKey(zeichenfolge("odr-match-v1"), zeichenfolge("GET"), zeichenfolge("svc"), zeichenfolge("/Products"), zahl(0), zahl(0), []byte{2}, sha([]byte(b)))
		if got := key(t, in, r).MatchKey; got != want {
			t.Errorf("%s: kein Rückfall auf Art 2 (%s)", name, got)
		}
	}
	// Modi ergeben für denselben gültigen Body verschiedene Keys, ungültigen Body ebenfalls (Art 1 gegen Art 2).
	for _, b := range []string{`{"a":1}`, `{"a":`} {
		gleichUndVerschieden(t, "Modi "+b, r, body(b), body(b), true)
		bytesRegeln := schluesselRegeln()
		if key(t, body(b), r).MatchKey == key(t, body(b), bytesRegeln).MatchKey {
			t.Errorf("Modi für %q ergeben denselben Key", b)
		}
	}
}

func TestKeyRedacted(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   model.SchluesselEingabe
		want bool
	}{
		{"none", basisEingabe(), false},
		{"applied", mit(func(in *model.SchluesselEingabe) { in.RequestRedaction = model.RedactionApplied }), true},
		{"Query geändert", q("token=x"), true},
		{"Query ohne Treffer", q("a=1"), false},
		{"Marker als Wert", q("token=[REDACTED]"), false},
		{"Header allein", vary(hmap{"Authorization": {"x"}, "Cookie": {"y"}}), false},
		{"Location allein", vary(hmap{"Location": {"/p?token=x"}}), false},
	} {
		if got := key(t, tc.in, schluesselRegeln()).KeyRedacted; got != tc.want {
			t.Errorf("%s: %v", tc.name, got)
		}
	}
	for _, rr := range []model.RedactionStatus{model.RedactionUnsafe, "", "kaputt"} {
		s, err := model.BildeSchluessel(mit(func(in *model.SchluesselEingabe) { in.RequestRedaction = rr }), schluesselRegeln())
		if !errors.Is(err, model.ErrUngueltig) || s.MatchKey != "" {
			t.Errorf("%q: %v %q", rr, err, s.MatchKey)
		}
	}
}

func TestKeyIdempotent(t *testing.T) {
	r := schluesselRegeln()
	in := model.SchluesselEingabe{Method: "GET", ServiceID: "svc", Path: "/Products", RawQuery: "a=1&token=" + geheimnis + "&b",
		Header: hmap{"Authorization": {geheimnis}, "Accept": {"x"}, "Location": {"/p?token=" + geheimnis}}, Body: []byte("b"), RequestRedaction: model.RedactionNone}
	s := key(t, in, r)
	if s.RawQuery != "a=1&token=[REDACTED]&b" || !reflect.DeepEqual(s.Header, hmap{"Authorization": {"[REDACTED]"}, "Accept": {"x"}, "Location": {"/p?token=[REDACTED]"}}) {
		t.Errorf("redigierte Darstellung: %q %v", s.RawQuery, s.Header)
	}
	zwei := in
	zwei.RawQuery, zwei.Header = s.RawQuery, s.Header
	if s2 := key(t, zwei, r); s2.MatchKey != s.MatchKey || s2.RawQuery != s.RawQuery || !reflect.DeepEqual(s2.Header, s.Header) {
		t.Errorf("zweiter Lauf: %v", s2)
	}
	if !s.KeyRedacted {
		t.Error("KeyRedacted")
	}
}

func TestKeyGeheimnisUnsichtbar(t *testing.T) {
	r := schluesselRegeln()
	a := mit(func(in *model.SchluesselEingabe) {
		in.RawQuery = "token=eins&api_key=a&x=1"
		in.Header = hmap{"Authorization": {"eins"}, "Accept": {"j"}}
	})
	b := mit(func(in *model.SchluesselEingabe) {
		in.RawQuery = "token=zwei&api_key=b&x=1"
		in.Header = hmap{"Authorization": {"zwei"}, "Accept": {"j"}}
	})
	gleichUndVerschieden(t, "anderes Geheimnis", r, a, b, true)
	c := mit(func(in *model.SchluesselEingabe) { in.RawQuery = "token=eins&api_key=a&x=2" })
	gleichUndVerschieden(t, "anderer Parameter", r, a, c, false)
}

func TestKeyRegelnUngueltig(t *testing.T) {
	for name, f := range map[string]func(*model.SchluesselRegeln){
		"Kodierung leer":    func(r *model.SchluesselRegeln) { r.InvalidEncoding = "" },
		"Kodierung fremd":   func(r *model.SchluesselRegeln) { r.InvalidEncoding = "lenient" },
		"Kodierung Schreib": func(r *model.SchluesselRegeln) { r.InvalidEncoding = "Literal" },
		"Body leer":         func(r *model.SchluesselRegeln) { r.BodyMode = "" },
		"Body fremd":        func(r *model.SchluesselRegeln) { r.BodyMode = "text" },
	} {
		r := schluesselRegeln()
		f(&r)
		if _, err := model.BildeSchluessel(basisEingabe(), r); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestKeyReihenfolgeVerstoesse(t *testing.T) {
	reject := schluesselRegeln()
	reject.InvalidEncoding = model.KodierungReject
	_, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "", ServiceID: "s", Path: "/a%zz", RequestRedaction: model.RedactionNone}, reject)
	if !errors.Is(err, model.ErrUngueltig) || errors.Is(err, model.ErrKodierung) {
		t.Errorf("leere Method und Pfad-%%: %v", err)
	}
	_, err = model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "s", Path: "/a%zz", RawQuery: "b=%zz", RequestRedaction: model.RedactionNone}, reject)
	if !errors.Is(err, model.ErrKodierung) {
		t.Errorf("Pfad und Query: %v", err)
	}
	_, err = model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "s", Path: "/a", RawQuery: "b=%zz", RequestRedaction: model.RedactionUnsafe}, reject)
	if !errors.Is(err, model.ErrUngueltig) || errors.Is(err, model.ErrKodierung) {
		t.Errorf("unsafe und Query: %v", err)
	}
	_, err = model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "", Path: "", RequestRedaction: model.RedactionNone}, schluesselRegelnMit(""))
	if !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Regeln vor allem: %v", err)
	}
	_, err = model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "s", Path: "/a", RawQuery: "b=%zz", RequestRedaction: model.RedactionNone}, reject)
	if !errors.Is(err, model.ErrKodierung) {
		t.Errorf("nur Query: %v", err)
	}
}

func schluesselRegelnMit(enc model.KodierungsModus) model.SchluesselRegeln {
	r := schluesselRegeln()
	r.InvalidEncoding = enc
	return r
}

func TestKeyGleichwertigkeit(t *testing.T) {
	laufe(t, []paarFall{
		{name: "gleich Query-Reihenfolge", a: q("a=1&b=2"), b: q("b=2&a=1"), gleich: true},
		{name: "gleich Hex Query", a: q("a=%2f"), b: q("a=%2F"), gleich: true},
		{name: "gleich Hex Pfad", a: pfad("/a%2f"), b: pfad("/a%2F"), gleich: true},
		{name: "gleich Header-Name", a: vary(hmap{"accept": {"a"}}), b: vary(hmap{"Accept": {"a"}}), gleich: true},
		{name: "verschieden Methode", a: mit(func(in *model.SchluesselEingabe) { in.Method = "POST" }), b: basisEingabe(), gleich: false},
		{name: "verschieden Service", a: mit(func(in *model.SchluesselEingabe) { in.ServiceID = "x" }), b: basisEingabe(), gleich: false},
		{name: "verschieden Pfad-Schreibweise", a: pfad("/products"), b: pfad("/Products"), gleich: false},
		{name: "verschieden Segmentgrenze", a: pfad("/a%2Fb"), b: pfad("/a/b"), gleich: false},
		{name: "verschieden + / %20", a: q("a=+"), b: q("a=%20"), gleich: false},
		{name: "verschieden Wiederholung", a: q("a=1&a=1"), b: q("a=1"), gleich: false},
		{name: "verschieden Vary", a: vary(hmap{"Accept": {"a"}}), b: vary(hmap{"Accept": {"b"}}), gleich: false},
		{name: "verschieden Body", a: body("a"), b: body("b"), gleich: false},
	})
}

func TestKeyEingabeUnveraendert(t *testing.T) {
	h := hmap{"Authorization": {geheimnis}, "Accept": {" a "}}
	b := []byte(`{"b":1,"a":2}`)
	in := model.SchluesselEingabe{Method: "GET", ServiceID: "s", Path: "/a%2f", RawQuery: "token=x&b=1", Header: h, Body: b, RequestRedaction: model.RedactionNone}
	r := schluesselRegeln()
	r.BodyMode = model.BodyJSONCanonical
	r.VaryHeaders = append([]string(nil), r.VaryHeaders...)
	vorher := schluesselRegeln()
	key(t, in, r)
	if !reflect.DeepEqual(h, hmap{"Authorization": {geheimnis}, "Accept": {" a "}}) || string(b) != `{"b":1,"a":2}` ||
		in.Path != "/a%2f" || in.RawQuery != "token=x&b=1" || !reflect.DeepEqual(r.VaryHeaders, vorher.VaryHeaders) ||
		!reflect.DeepEqual(r.RedactHeaders, vorher.RedactHeaders) || !reflect.DeepEqual(r.RedactQueryParams, vorher.RedactQueryParams) {
		t.Error("Eingabe verändert")
	}
}

func TestKeyDeterministisch(t *testing.T) {
	h := hmap{"Accept": {"a"}, "Accept-Language": {"de"}, "X-Tenant": {"t"}, "x-tenant": {"u"}, "Authorization": {"s"}, "Cookie": {"c"}}
	in := model.SchluesselEingabe{Method: "GET", ServiceID: "s", Path: "/a", RawQuery: "z=1&y=2&x=3&w&v=&u=%41", Header: h, Body: []byte(`{"b":1,"a":2}`), RequestRedaction: model.RedactionNone}
	r := schluesselRegeln()
	r.BodyMode = model.BodyJSONCanonical
	erste := key(t, in, r)
	for i := 0; i < 100; i++ {
		if s := key(t, in, r); s.MatchKey != erste.MatchKey || !reflect.DeepEqual(s.Header, erste.Header) {
			t.Fatalf("Lauf %d: %v", i, s)
		}
	}
}
