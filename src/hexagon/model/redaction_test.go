package model_test

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const geheimnis = "S3CR3T-7f"

func standardNamen() []string {
	return []string{"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token", "X-CSRF-Token"}
}

func standardParameter() []string { return []string{"token", "api_key"} }

type hmap = map[string][]string

func TestMarker(t *testing.T) {
	if model.Marker != "[REDACTED]" {
		t.Errorf("Marker = %q", model.Marker)
	}
}

func TestRedigiereHeaderWerte(t *testing.T) {
	for name, werte := range map[string][]string{"keiner": nil, "leer": {}, "leerer Wert": {""}, "einer": {"a"}, "mehrere": {"a", "b", ""}} {
		got := model.RedigiereHeader(hmap{"Authorization": werte}, standardNamen(), nil)
		if want := (hmap{"Authorization": {"[REDACTED]"}}); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, erwartet %v", name, got, want)
		}
	}
}

func TestRedigiereHeaderNamen(t *testing.T) {
	for _, n := range standardNamen() {
		for _, schreibweise := range []string{n, strings.ToUpper(n), strings.ToLower(n)} {
			got := model.RedigiereHeader(hmap{schreibweise: {geheimnis}}, standardNamen(), nil)
			if len(got) != 1 || !reflect.DeepEqual(got[http.CanonicalHeaderKey(n)], []string{"[REDACTED]"}) {
				t.Errorf("%s nicht redigiert: %v", schreibweise, got)
			}
		}
	}
	got := model.RedigiereHeader(hmap{"Authorisation": {"x"}, "X-Extra": {"y"}}, append(standardNamen(), "x-extra"), nil)
	if !reflect.DeepEqual(got, hmap{"Authorisation": {"x"}, "X-Extra": {"[REDACTED]"}}) {
		t.Errorf("Ergänzung oder ähnlicher Name: %v", got)
	}
	// Kelvin-Zeichen U+212A ist kein ASCII-k und gleicht dem Namen nicht.
	if got := model.RedigiereHeader(hmap{"CooKie": {"x"}}, []string{"Cooxie"}, nil); got["CooKie"][0] != "x" {
		t.Errorf("Unicode-Faltung: %v", got)
	}
}

func TestRedigiereHeaderUnveraendert(t *testing.T) {
	in := hmap{"Accept": {"b", "a"}, "X-Leer": {}, "X-Leerwert": {""}}
	if got := model.RedigiereHeader(in, standardNamen(), nil); !reflect.DeepEqual(got, in) {
		t.Errorf("%v, erwartet %v", got, in)
	}
}

func TestRedigiereHeaderSchreibweise(t *testing.T) {
	got := model.RedigiereHeader(hmap{"cookie": {"a"}, "Cookie": {"b"}}, standardNamen(), nil)
	if !reflect.DeepEqual(got, hmap{"Cookie": {"[REDACTED]"}}) {
		t.Errorf("redigiert: %v", got)
	}
	got = model.RedigiereHeader(hmap{"x-a": {"2"}, "X-A": {"1", "3"}, "X-a": {"4"}}, nil, nil)
	if !reflect.DeepEqual(got, hmap{"X-A": {"1", "3", "4", "2"}}) {
		t.Errorf("Bytefolge der Namen: %v", got)
	}
	got = model.RedigiereHeader(hmap{"x-api-key": {"a"}, "bad name": {"v"}}, nil, nil)
	if !reflect.DeepEqual(got, hmap{"X-Api-Key": {"a"}, "bad name": {"v"}}) {
		t.Errorf("kanonisch, Nicht-Token: %v", got)
	}
}

func TestRedigiereHeaderKopie(t *testing.T) {
	for _, in := range []hmap{nil, {}} {
		if got := model.RedigiereHeader(in, standardNamen(), nil); got == nil || len(got) != 0 {
			t.Errorf("%v: %#v", in, got)
		}
	}
	in := hmap{"Authorization": {"a"}, "Accept": {"x", "y"}, "Location": {"/p?token=1"}}
	vorher := hmap{"Authorization": {"a"}, "Accept": {"x", "y"}, "Location": {"/p?token=1"}}
	got := model.RedigiereHeader(in, standardNamen(), standardParameter())
	got["Accept"][0] = "veraendert"
	if !reflect.DeepEqual(in, vorher) {
		t.Errorf("Eingabe verändert: %v", in)
	}
	got2 := model.OhneRedigierte(in)
	got2["Accept"][0] = "veraendert"
	if !reflect.DeepEqual(in, vorher) || got2 == nil {
		t.Errorf("OhneRedigierte verändert die Eingabe")
	}
	if got2 := model.OhneRedigierte(nil); got2 == nil || len(got2) != 0 {
		t.Errorf("OhneRedigierte(nil) = %#v", got2)
	}
}

func TestRedigiereLocation(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"/p?token=1&a=2", "/p?token=[REDACTED]&a=2"},
		{"/p", "/p"},
		{"", ""},
		{"?", "?"},
		{"http://x?", "http://x?"},
		{"http://x/p?a=1&api_key=k#frag?token=1", "http://x/p?a=1&api_key=[REDACTED]#frag?token=1"},
		{"/p?token=1#f", "/p?token=[REDACTED]#f"},
		{"/p#token=1", "/p#token=1"},
	} {
		for _, name := range []string{"Location", "Content-Location", "location"} {
			got := model.RedigiereHeader(hmap{name: {tc.in}}, nil, standardParameter())
			if w := (hmap{http.CanonicalHeaderKey(name): {tc.want}}); !reflect.DeepEqual(got, w) {
				t.Errorf("%s %q: %v, erwartet %v", name, tc.in, got, w)
			}
		}
	}
	got := model.RedigiereHeader(hmap{"Location": {"/a?token=1", "/b?x=1&token=2"}}, nil, standardParameter())
	if !reflect.DeepEqual(got, hmap{"Location": {"/a?token=[REDACTED]", "/b?x=1&token=[REDACTED]"}}) {
		t.Errorf("mehrere Werte: %v", got)
	}
	got = model.RedigiereHeader(hmap{"Location": {"/a?token=1"}}, []string{"location"}, standardParameter())
	if !reflect.DeepEqual(got, hmap{"Location": {"[REDACTED]"}}) {
		t.Errorf("ganz redigiert gewinnt: %v", got)
	}
	if got := model.RedigiereHeader(hmap{"Referer": {"/a?token=1"}}, nil, standardParameter()); got["Referer"][0] != "/a?token=1" {
		t.Errorf("Referer wird nicht untersucht: %v", got)
	}
}

type queryFall struct{ in, want string }

func queryFaelle() []queryFall {
	return []queryFall{
		{"", ""}, {"?", "?"}, {"a", "a"}, {"a=", "a="}, {"=v", "=v"}, {"&&", "&&"},
		{"token=x&token=y", "token=[REDACTED]&token=[REDACTED]"},
		{"TOKEN=x", "TOKEN=[REDACTED]"}, {"tok%65n=x", "tok%65n=[REDACTED]"}, {"to%ken=x", "to%ken=x"},
		{"a=1;token=x", "a=1;token=x"}, {"token", "token"}, {"?token=x", "?token=x"},
		{"token=%zz", "token=[REDACTED]"}, {"token=", "token=[REDACTED]"},
		{"a=1&token=x&b=2&api_key=y&c", "a=1&token=[REDACTED]&b=2&api_key=[REDACTED]&c"},
		{"token=a=b", "token=[REDACTED]"}, {"a+b=1", "a+b=1"}, {"tok%2Bn=1", "tok%2Bn=1"},
	}
}

func TestRedigiereQueryRandformen(t *testing.T) {
	for _, tc := range queryFaelle() {
		if got, _ := model.RedigiereQuery(tc.in, standardParameter()); got != tc.want {
			t.Errorf("%q: %q, erwartet %q", tc.in, got, tc.want)
		}
	}
	if got, _ := model.RedigiereQuery("token=x", nil); got != "token=x" {
		t.Errorf("leere Menge: %q", got)
	}
}

func TestRedigiereQueryGeaendert(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{{"token=x", true}, {"token=", true}, {"a=1", false}, {"token", false}, {"token=[REDACTED]", false}, {"", false}, {"a=1&token=[REDACTED]&api_key=z", true}} {
		if _, got := model.RedigiereQuery(tc.in, standardParameter()); got != tc.want {
			t.Errorf("%q: %v, erwartet %v", tc.in, got, tc.want)
		}
	}
	erst, _ := model.RedigiereQuery("token=x", standardParameter())
	if _, zweit := model.RedigiereQuery(erst, standardParameter()); zweit {
		t.Error("zweiter Lauf meldet eine Änderung")
	}
}

func TestRedigiereIdempotent(t *testing.T) {
	for _, tc := range queryFaelle() {
		eins, _ := model.RedigiereQuery(tc.in, standardParameter())
		zwei, geaendert := model.RedigiereQuery(eins, standardParameter())
		if eins != zwei || geaendert {
			t.Errorf("%q: %q dann %q (geändert %v)", tc.in, eins, zwei, geaendert)
		}
	}
	for _, in := range headerFaelle() {
		eins := model.RedigiereHeader(in, standardNamen(), standardParameter())
		if zwei := model.RedigiereHeader(eins, standardNamen(), standardParameter()); !reflect.DeepEqual(eins, zwei) {
			t.Errorf("%v: %v dann %v", in, eins, zwei)
		}
	}
}

func headerFaelle() []hmap {
	return []hmap{
		nil, {}, {"Authorization": {geheimnis}}, {"cookie": {geheimnis, "b"}, "Cookie": {geheimnis}},
		{"Location": {"http://x/p?token=" + geheimnis + "#f"}}, {"Content-Location": {"/p?a=1&api_key=" + geheimnis}},
		{"X-Api-Key": {}, "Accept": {"a", "b"}, "set-cookie": {""}}, {"Location": {"/a?token=1", "/b"}},
	}
}

func TestRedigiereQueryUngueltigeKodierung(t *testing.T) {
	for _, in := range []string{"token=%zz", "token=%", "token=\xff", "a=%zz&token=x", "%=1&token=x", "\xff=1&token=x"} {
		got, _ := model.RedigiereQuery(in, standardParameter())
		if strings.HasPrefix(in, "token=") && got != "token=[REDACTED]" {
			t.Errorf("%q: %q", in, got)
		}
		if strings.HasSuffix(in, "token=x") && !strings.HasSuffix(got, "token=[REDACTED]") {
			t.Errorf("%q: %q", in, got)
		}
	}
}

func TestKeinGeheimnisImErgebnis(t *testing.T) {
	var eingaben []string
	for _, tc := range queryFaelle() {
		eingaben = append(eingaben, tc.in+"&token="+geheimnis, "api_key="+geheimnis+"&"+tc.in+"&token="+geheimnis, "token="+geheimnis+"&"+tc.in)
	}
	eingaben = append(eingaben, "token="+geheimnis+"&api_key="+geheimnis, "x=%zz&token="+geheimnis, "x=\xff&token="+geheimnis)
	regeln := schluesselRegeln()
	for _, in := range eingaben {
		if got, _ := model.RedigiereQuery(in, standardParameter()); strings.Contains(got, geheimnis) {
			t.Errorf("RedigiereQuery(%q) = %q", in, got)
		}
		for _, modus := range []model.KodierungsModus{model.KodierungLiteral, model.KodierungReject} {
			regeln.InvalidEncoding = modus
			s, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "d", Path: "/p", RawQuery: in,
				Header: hmap{"Authorization": {geheimnis}, "Location": {"/p?token=" + geheimnis}}, RequestRedaction: model.RedactionNone}, regeln)
			if err != nil && strings.Contains(err.Error(), geheimnis) {
				t.Errorf("Fehlertext zu %q: %v", in, err)
			}
			if flach := s.RawQuery + "|" + flachHeader(s.Header); strings.Contains(flach, geheimnis) {
				t.Errorf("BildeSchluessel(%q): %s", in, flach)
			}
		}
	}
	for _, h := range headerFaelle() {
		if flach := flachHeader(model.RedigiereHeader(h, standardNamen(), standardParameter())); strings.Contains(flach, geheimnis) {
			t.Errorf("%v: %s", h, flach)
		}
	}
	_, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "", ServiceID: "d", Path: "/p" + geheimnis, RawQuery: geheimnis}, regeln)
	if err == nil || strings.Contains(err.Error(), geheimnis) {
		t.Errorf("Fehler ohne Wert erwartet: %v", err)
	}
	regeln.InvalidEncoding = model.KodierungReject
	_, err = model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "d", Path: "/p%zz" + geheimnis, RequestRedaction: model.RedactionNone}, regeln)
	if !errors.Is(err, model.ErrKodierung) || strings.Contains(err.Error(), geheimnis) {
		t.Errorf("ErrKodierung ohne Wert erwartet: %v", err)
	}
	if _, err := model.ParseJSONPfad("$.." + geheimnis + " x"); err == nil || strings.Contains(err.Error(), geheimnis) {
		t.Errorf("Pfadmuster-Fehler: %v", err)
	}
}

func flachHeader(h hmap) string {
	var sb strings.Builder
	for n, ws := range h {
		sb.WriteString(n + "=" + strings.Join(ws, ",") + ";")
	}
	return sb.String()
}

func TestOhneRedigierte(t *testing.T) {
	in := hmap{
		"A": {"[REDACTED]"}, "B": {"[REDACTED]", "[REDACTED]"}, "C": {}, "D": {""}, "E": {"x[REDACTED]"},
		"F": {"[REDACTED]", "x"}, "G": {"x"},
	}
	want := hmap{"C": {}, "D": {""}, "E": {"x[REDACTED]"}, "F": {"[REDACTED]", "x"}, "G": {"x"}}
	if got := model.OhneRedigierte(in); !reflect.DeepEqual(got, want) {
		t.Errorf("%v, erwartet %v", got, want)
	}
}

func TestParseJSONPfad(t *testing.T) {
	for _, s := range []string{"$..password", "$..api_key", "$..A-b_1"} {
		if _, err := model.ParseJSONPfad(s); err != nil {
			t.Errorf("%q: %v", s, err)
		}
	}
	for _, s := range []string{"", "$", "$..", "$.a", "$...a", "..a", "$..a.b", "$..a b", " $..a", "$..a ", "$..*", "$..[0]", "$..a[0]", "$..ä", "$..a,b"} {
		if _, err := model.ParseJSONPfad(s); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%q: %v, erwartet ErrUngueltig", s, err)
		}
	}
}

func TestJSONPfadPasst(t *testing.T) {
	p, err := model.ParseJSONPfad("$..Password")
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{"password": true, "PASSWORD": true, "Password": true, "passwort": false, "": false, "password ": false, "passKword": false} {
		if got := p.Passt(name); got != want {
			t.Errorf("Passt(%q) = %v", name, got)
		}
	}
	k, err := model.ParseJSONPfad("$..k")
	if err != nil || k.Passt("\u212A") || !k.Passt("K") {
		t.Errorf("Kelvin-Zeichen oder ASCII-Schreibweise: %v", err)
	}
	if (model.JSONPfad{}).Passt("") {
		t.Error("Nullwert passt")
	}
}
