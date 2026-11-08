package jsoncodec_test

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// acceptNone ist der Accept der Stufe none.
func acceptNone() []string { return []string{"application/json;odata.metadata=none"} }

// accepts sind die zwei Stufen: Name und Accept.
func accepts() map[string][]string {
	return map[string][]string{"minimal": nil, "none": acceptNone()}
}

// mitEtag ist die Entity mit dem Etag.
func mitEtag(e model.Entity, etag string) model.Entity {
	e.Etag = etag
	return e
}

// jsonEtag ist der Etag als JSON-String, aus Bytewerten: Anführungszeichen im Text mit Backslash (SPEC-169 Nr. 1, AGENTS.md §3.18).
func jsonEtag(etag string) string {
	return `"` + strings.ReplaceAll(etag, `"`, bs+`"`) + `"`
}

// nachID fügt `,"@odata.etag":<Etag>` hinter dem Mitglied `@odata.id` ein; das Orakel kennt weder Schreiber noch Struktur des Codes.
func nachID(t *testing.T, body, etag string) string {
	t.Helper()
	const id = `"@odata.id":"`
	i := strings.Index(body, id)
	if i < 0 {
		t.Fatalf("kein @odata.id in %s", body)
	}
	ende := i + len(id) + strings.Index(body[i+len(id):], `"`) + 1
	return body[:ende] + `,"@odata.etag":` + jsonEtag(etag) + body[ende:]
}

// gueltigeEtags sind die Randwerte und 0 bis 50.
func gueltigeEtags() []string {
	r := []string{}
	for _, n := range []string{"0", "1", "9", "10", "99", "100", "2147483647", "4294967296", "9223372036854775807"} {
		r = append(r, `W/"`+n+`"`)
	}
	for n := 0; n <= 50; n++ {
		r = append(r, `W/"`+strconv.Itoa(n)+`"`)
	}
	return r
}

// TestEtagMitglied (E13): der Schreiber setzt `@odata.etag` hinter `@odata.id` (und hinter `@odata.type`), in der Stufe none nie, für ein
// leeres Etag nie; Entity und jedes Element einer Collection.
func TestEtagMitglied(t *testing.T) {
	n := 0
	sonder := model.Entity{Typ: "Ns.Sonder", Schluessel: "7", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "7"), pw("Name", model.ArtString, "S")}}
	for _, e := range []model.Entity{produkt("1", "A"), sonder} {
		for _, etag := range gueltigeEtags() {
			ea := model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: e}
			ohne := entityBody(t, "ohne", ea)
			ea.Entity = mitEtag(e, etag)
			if got, want := entityBody(t, etag, ea), nachID(t, ohne, etag); got != want {
				t.Errorf("Entity %s %s:\n got %s\nwant %s", e.Typ, etag, got, want)
			}
			ea.Accept = acceptNone()
			if got, want := entityBody(t, "none "+etag, ea), entityBody(t, "none ohne", model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: e, Accept: acceptNone()}); got != want || strings.Contains(got, "etag") {
				t.Errorf("Entity none %s: %s, will %s", etag, got, want)
			}
			n += 2
		}
	}
	// Collection: das Mitglied steht an jedem Element, an keiner anderen Stelle.
	for _, etag := range gueltigeEtags() {
		es := []model.Entity{produkt("1", "A"), produkt("2", "B"), produkt("3", "C")}
		ca := model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: es}
		ohne := collectionBody(t, "ohne", ca)
		for i := range es {
			es[i] = mitEtag(es[i], etag)
		}
		want := ohne
		for _, id := range []string{"Produkte(1)", "Produkte(2)", "Produkte(3)"} {
			marke := `"@odata.id":"` + wurzel + `/` + id + `"`
			want = strings.Replace(want, marke, marke+`,"@odata.etag":`+jsonEtag(etag), 1)
		}
		if got := collectionBody(t, etag, ca); got != want || strings.Count(got, "@odata.etag") != 3 {
			t.Errorf("Collection %s:\n got %s\nwant %s", etag, got, want)
		}
		ca.Accept = acceptNone()
		if got := collectionBody(t, "none "+etag, ca); strings.Contains(got, "etag") {
			t.Errorf("Collection none %s: %s", etag, got)
		}
		n += 2
	}
	if want := 2*2*len(gueltigeEtags()) + 2*len(gueltigeEtags()); n != want || len(gueltigeEtags()) != 60 {
		t.Errorf("%d Fälle, erwartet %d", n, want)
	}
}

// ungueltigeEtags sind die Formen von SPEC-168 Nr. 3 e, die nicht `W/"<Ganzzahl 0 bis 2^63-1 ohne führende Null>"` sind; die Zeichen jenseits
// von ASCII und die Steuerzeichen entstehen aus Bytewerten.
func ungueltigeEtags() []string {
	arabisch := string([]byte{0xd9, 0xa3})
	return []string{
		`7`, `"7"`, `W/7`, `w/"7"`, `W/"7`, `W/7"`, `W/"7" `, ` W/"7"`, `W/ "7"`, `W/" 7"`, `W/"7 "`, `W/"07"`, `W/"00"`, `W/"+7"`, `W/"-7"`, `W/"-0"`,
		`W/"a"`, `W/"7a"`, `W/"0x7"`, `W/"7.0"`, `W/"1e3"`, `W/"12345678901234567890"`, `W/"9223372036854775808"`, `W/"99999999999999999999999"`,
		`W/"7""`, `W/"7"x`, `W/"` + bs + `"`, `W/"7` + bs + `"`, `W/"7` + bs + bs + `"`, `W/"` + "\x01" + `"`, `W/"7` + "\x00" + `"`, `W/"7` + "\n" + `"`,
		`W/""`, `W/"`, `W/`, `W`, `"W/` + bs + `"7` + bs + `""`, `W/"` + arabisch + `"`, "W/\"7\xff\"", `W/'7'`, `V/"7"`, `W\"7"`,
	}
}

// TestEtagUngueltig (E13): jede ungültige Form ist ErrUngueltig ohne Schreiben, als Entity, als jedes Element einer Collection von drei
// (erste, mittlere und letzte Stelle) und in beiden Stufen.
func TestEtagUngueltig(t *testing.T) {
	n := 0
	for stufe, accept := range accepts() {
		for _, etag := range ungueltigeEtags() {
			was := stufe + " " + strconv.Quote(etag)
			entityScheitert(t, "Entity "+was, model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: mitEtag(produkt("1", "A"), etag), Accept: accept})
			for i := 0; i < 3; i++ {
				es := []model.Entity{produkt("1", "A"), produkt("2", "B"), produkt("3", "C")}
				es[i] = mitEtag(es[i], etag)
				collectionScheitert(t, "Collection "+was+" Stelle "+strconv.Itoa(i), model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: es, Accept: accept})
				n++
			}
			n++
		}
	}
	if want := 2 * len(ungueltigeEtags()) * 4; n != want {
		t.Errorf("%d Fälle, erwartet %d", n, want)
	}
}

// etagGrammatik ist das Orakel der Form (SPEC-168 Nr. 3 e), unabhängig vom Schreiber: `W/"`, Ziffern ohne führende Null, `"`, und ein Wert bis 2^63-1.
var etagGrammatik = regexp.MustCompile(`^W/"(0|[1-9][0-9]*)"$`)

func etagOrakelGueltig(s string) bool {
	m := etagGrammatik.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	n, ok := new(big.Int).SetString(m[1], 10)
	return ok && n.Cmp(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(1))) <= 0
}

// etagErzeugte erzeugt aus fünf gültigen Ausgangsformen alle Formen, die ein Zeichen löschen, einfügen, ersetzen oder verdoppeln oder an einer
// Stelle abschneiden, mit 16 Zeichen (darunter Ziffern, Anführungszeichen, Schrägstrich, Vorzeichen, Leerraum, Backslash, Steuer- und
// Nicht-ASCII-Zeichen aus Bytewerten), dazu die Zehnerpotenzen und Neunerfolgen von 1 bis 25 Stellen; jede Form ohne Duplikat.
func etagErzeugte() []string {
	basen := []string{`W/"0"`, `W/"7"`, `W/"10"`, `W/"2147483647"`, `W/"9223372036854775807"`}
	zeichen := []string{"0", "7", "9", `"`, "/", "W", "w", "a", "+", "-", " ", bs, "\x00", "\n", string([]byte{0xd9, 0xa3}), string([]byte{0xff})}
	gesehen := map[string]bool{}
	var alle []string
	add := func(s string) {
		if s != "" && !gesehen[s] { // leer heißt kein ETag und ist keine Form des Etag
			gesehen[s] = true
			alle = append(alle, s)
		}
	}
	for _, b := range basen {
		add(b)
		for i := 0; i <= len(b); i++ {
			add(b[:i])
			for _, c := range zeichen {
				add(b[:i] + c + b[i:])
				if i < len(b) {
					add(b[:i] + c + b[i+1:])
				}
			}
			if i < len(b) {
				add(b[:i] + b[i+1:])
				add(b[:i+1] + b[i:])
			}
		}
	}
	for n := 1; n <= 25; n++ {
		add(`W/"1` + strings.Repeat("0", n-1) + `"`)
		add(`W/"` + strings.Repeat("9", n) + `"`)
	}
	return alle
}

// TestEtagGenerator (E13, AGENTS.md §3.17): die Klasse „jede Form, die aus einer gültigen durch eine Änderung entsteht“ gegen das Orakel der Grammatik: gültig
// laut Orakel heißt, der Schreiber schreibt das Mitglied; sonst ErrUngueltig ohne Schreiben. Die Zählprüfung nennt die Menge.
func TestEtagGenerator(t *testing.T) {
	alle := etagErzeugte()
	gueltig, ungueltig := 0, 0
	for _, etag := range alle {
		e := mitEtag(produkt("1", "A"), etag)
		ea := model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: e}
		if etagOrakelGueltig(etag) {
			gueltig++
			ohne := entityBody(t, "ohne", model.EntityAntwort{Wurzel: wurzel, Set: "Produkte", Entity: produkt("1", "A")})
			if got, want := entityBody(t, etag, ea), nachID(t, ohne, etag); got != want {
				t.Errorf("gültig %q:\n got %s\nwant %s", etag, got, want)
			}
			continue
		}
		ungueltig++
		entityScheitert(t, "ungültig "+strconv.Quote(etag), ea)
		collectionScheitert(t, "Collection ungültig "+strconv.Quote(etag), model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: []model.Entity{produkt("2", "B"), e}})
	}
	if len(alle) < 1500 || gueltig < 40 || ungueltig < 1400 {
		t.Errorf("Menge: %d Formen, %d gültig, %d ungültig", len(alle), gueltig, ungueltig)
	}
	t.Logf("Menge: %d Formen, %d gültig, %d ungültig", len(alle), gueltig, ungueltig)
}

// TestEtagRundlauf (E13): der Body mit `@odata.etag` liest ODataCodec in dieselbe Entity wie den ohne (der Leser setzt Etag nie).
func TestEtagRundlauf(t *testing.T) {
	for _, etag := range gueltigeEtags() {
		e := model.Entity{Typ: "Ns.Alle", Schluessel: "'k'", Properties: []model.PropertyWert{pw("S", model.ArtString, "k")}, Etag: etag}
		rundlauf(t, "Etag "+etag, "Alles", nil, e, model.Anzahl{}, nil)
	}
}
