package outboundtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen K1 bis K3 prüfen den Schlüssel (SPEC-172 Nr. 5 mit SPEC-126 Nr. 11): die Literalform, die Verletzungen und die
// Gleichheit.

// v2Schluessel ist ein Schlüsselwert eines Typs der sieben mit dem erwarteten Literal.
type v2Schluessel struct {
	json, literal string
}

// v2SchluesselWerte sind die Schlüsselwerte je Typ: Edm.String mit verdoppeltem Anführungszeichen, jeder andere Typ sein
// kanonischer Text (Edm.DateTimeOffset: der Zeitpunkt in UTC ohne Anführungszeichen).
func v2SchluesselWerte(edm string) []v2Schluessel {
	switch edm {
	case "Edm.String":
		return []v2Schluessel{{`"a"`, `'a'`}, {`""`, `''`}, {`"O'Hara"`, `'O''Hara'`}, {`"'"`, `''''`}, {`"''"`, `''''''`}, {`"a b/c"`, `'a b/c'`}}
	case "Edm.Boolean":
		return []v2Schluessel{{`true`, `true`}, {`false`, `false`}}
	case "Edm.Int32":
		return []v2Schluessel{{`1`, `1`}, {`-5`, `-5`}, {`0`, `0`}, {`2147483647`, `2147483647`}}
	case "Edm.Int64":
		return []v2Schluessel{{`"5"`, `5`}, {`5`, `5`}, {`"-9223372036854775808"`, `-9223372036854775808`}, {`0`, `0`}}
	case "Edm.Decimal":
		return []v2Schluessel{{`1.50`, `1.5`}, {`"1.5"`, `1.5`}, {`-0`, `0`}, {`100`, `100`}}
	case "Edm.Guid":
		return []v2Schluessel{{`"` + strings.ToUpper(guidKlein) + `"`, guidKlein}, {`"` + guidKlein + `"`, guidKlein}}
	}
	return []v2Schluessel{{v2Dt("0"), `1970-01-01T00:00:00Z`}, {v2Dt("915148800000"), `1999-01-01T00:00:00Z`}, {v2Dt("-1"), `1969-12-31T23:59:59.999Z`}}
}

func v2FaelleK1() []jfall {
	var out []jfall
	for _, edm := range v2Typen() {
		for _, w := range v2SchluesselWerte(edm) {
			for _, f := range v2BeideGut("K1 "+edm+" "+w.json, "T", "K"+kurz(edm), `"Id":`+w.json) {
				f.schluessel = w.literal
				out = append(out, f)
			}
		}
	}
	for _, p := range permutationen(2) {
		teile := []string{`"A":1`, `"B":"x"`}
		var ms []string
		for _, i := range p {
			ms = append(ms, teile[i])
		}
		for _, f := range v2BeideGut(fmt.Sprintf("K1 K2 Reihenfolge %v", p), "T", "K2", strings.Join(ms, ",")) {
			f.schluessel = "A=1,B='x'"
			out = append(out, f)
		}
	}
	for _, p := range permutationen(3) {
		teile := []string{`"A":1.0`, `"B":"it's"`, `"C":true`}
		var ms []string
		for _, i := range p {
			ms = append(ms, teile[i])
		}
		for _, f := range v2BeideGut(fmt.Sprintf("K1 K3 Reihenfolge %v", p), "T", "K3", strings.Join(ms, ",")+`,"Note":"n"`) {
			f.schluessel = "A=1,B='it''s',C=true"
			out = append(out, f)
		}
	}
	return out
}

func v2K1(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleK1()) }

// v2Verletzungen sind die Verstöße gegen den Schlüssel einer Position: `null` (auch bei Nullable), der Marker, ein String mit
// U+0000, die fehlende Property. ursache nennt die Ursache je Typ der Position.
type v2Verletzung struct {
	name, json string
	ursache    func(edm string) model.DatenUrsache
}

func v2Verletzungen() []v2Verletzung {
	immer := func(string) model.DatenUrsache { return model.DatenSchluessel }
	return []v2Verletzung{
		{"null", `null`, immer},
		{"Marker", `"[REDACTED]"`, func(edm string) model.DatenUrsache {
			// Bei Edm.Int32 und Edm.Boolean ist der Marker ein Fehler des Werts (typ), der dem Schlüssel vorgeht.
			if edm == "Edm.Int32" || edm == "Edm.Boolean" {
				return model.DatenTyp
			}
			return model.DatenSchluessel
		}},
		{"U+0000", `"a` + esc("0000") + `"`, func(edm string) model.DatenUrsache {
			switch edm {
			case "Edm.String":
				return model.DatenSchluessel
			case "Edm.Int32", "Edm.Boolean":
				return model.DatenTyp
			}
			return model.DatenWert
		}},
		{"fehlt", ``, immer},
	}
}

func v2FaelleK2() []jfall {
	var out []jfall
	for _, edm := range v2Typen() {
		for _, v := range v2Verletzungen() {
			members := `"Id":` + v.json
			if v.json == "" {
				members = ``
			}
			name := fmt.Sprintf("K2 %s %s", edm, v.name)
			out = append(out, v2Beide(name, "T", "K"+kurz(edm), members, v.ursache(edm), "/Id")...)
		}
	}
	pos := []struct{ name, typ string }{{"A", "Edm.Int32"}, {"B", "Edm.String"}, {"C", "Edm.Boolean"}}
	gut := []string{`"A":1`, `"B":"x"`, `"C":true`}
	for i, p := range pos {
		for _, v := range v2Verletzungen() {
			ms := append([]string(nil), gut...)
			ms[i] = `"` + p.name + `":` + v.json
			if v.json == "" {
				ms = append(ms[:i:i], ms[i+1:]...)
			}
			out = append(out, v2Beide(fmt.Sprintf("K2 K3 Position %s %s", p.name, v.name), "T", "K3", strings.Join(ms, ","), v.ursache(p.typ), "/"+p.name)...)
		}
	}
	out = append(out, v2Beide("K2 nullbarer Schlüssel mit null", "T", "KN", `"Id":null`, model.DatenSchluessel, "/Id")...)
	out = append(out, v2Beide("K2 Fehler des Werts geht dem fehlenden Schlüssel vor", "T", "K2", `"A":"x"`, model.DatenTyp, "/A")...)
	out = append(out, v2Beide("K2 Wert 1.5 geht dem fehlenden Schlüssel vor", "T", "K2", `"A":1.5`, model.DatenWert, "/A")...)
	return out
}

func v2K2(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleK2()) }

// v2Gleiche sind Schreibweisen desselben Schlüssels je Typ der sieben; die Schreibweisen von v2Andere sind verschieden von ihnen.
func v2Gleiche(edm string) (gleich []string, anders string) {
	switch edm {
	case "Edm.String":
		return []string{`"a"`, `"` + esc("0061") + `"`}, `"A"`
	case "Edm.Boolean":
		return []string{`true`}, `false`
	case "Edm.Int32":
		return []string{`1`, `1.0`, `1e0`, `10e-1`, `1E+0`, `0.1e1`}, `2`
	case "Edm.Int64":
		return []string{`1`, `"1"`, `1.0`, `"1.0"`, `"1e0"`, `10e-1`}, `2`
	case "Edm.Decimal":
		return []string{`1`, `"1.0"`, `1e0`, `10e-1`, `"1.00"`, `0.1e1`}, `1.01`
	case "Edm.Guid":
		return []string{`"` + guidKlein + `"`, `"` + strings.ToUpper(guidKlein) + `"`, `"0F8fad5b-D9CB-469f-A165-70867728950e"`}, `"` + strings.Replace(guidKlein, "0f", "1f", 1) + `"`
	}
	return []string{v2Dt("0"), `"/Date(0)/"`, `"` + esc("002F") + `Date(0)/"`, `"/Date(0)` + esc("002f") + `"`}, v2Dt("1")
}

// schluesselVon liest die Entity des Sets mit dem Wert als Schlüssel und liefert ihr Literal.
func (s *jsuite) schluesselVon(t *testing.T, set, json string) string {
	t.Helper()
	r := s.lauf(jfall{m: "T", set: set, einzel: true, body: v2Ein(`"Id":` + json)})
	if r.err != nil {
		t.Errorf("K3 %s %s: %v", set, json, r.err)
	}
	return r.entity.Schluessel
}

// v2K3: verschiedene Schlüssel haben verschiedene Texte, derselbe Schlüssel hat denselben, gleich in welcher Schreibweise; der
// Text trägt den Typ (`1` gegen `'1'`). Schleife über die Schreibweisen je Typ.
func v2K3(t *testing.T, s *jsuite) {
	texte := map[string]string{}
	for _, edm := range v2Typen() {
		gleich, anders := v2Gleiche(edm)
		set := "K" + kurz(edm)
		erster := s.schluesselVon(t, set, gleich[0])
		for _, j := range gleich[1:] {
			if got := s.schluesselVon(t, set, j); got != erster {
				t.Errorf("K3 %s: %s ergibt %q, %s ergibt %q, erwartet denselben Text", edm, j, got, gleich[0], erster)
			}
		}
		if got := s.schluesselVon(t, set, anders); got == erster {
			t.Errorf("K3 %s: %s und %s ergeben beide %q", edm, anders, gleich[0], got)
		}
		texte[edm] = erster
	}
	// Der Text trägt den Typ: der Wert 1 als Edm.Int64 gegen den String "1".
	if a, b := s.schluesselVon(t, "KInt64", `1`), s.schluesselVon(t, "KString", `"1"`); a == b || a != "1" || b != "'1'" {
		t.Errorf("K3: Edm.Int64 1 ergibt %q, Edm.String \"1\" ergibt %q", a, b)
	}
	// Verschiedene Typen mit gleichem Text sind verschieden: der Boolean true gegen den String "true".
	if a, b := s.schluesselVon(t, "KBoolean", `true`), s.schluesselVon(t, "KString", `"true"`); a == b {
		t.Errorf("K3: Edm.Boolean true und Edm.String \"true\" ergeben beide %q", a)
	}
	// Die Reihenfolge des Bodys ändert den zusammengesetzten Schlüssel nicht (K1 prüft sie je Permutation).
	r1 := s.lauf(jfall{m: "T", set: "K2", einzel: true, body: v2Ein(`"A":1,"B":"x"`)})
	r2 := s.lauf(jfall{m: "T", set: "K2", einzel: true, body: v2Ein(`"B":"x","A":1.0`)})
	r3 := s.lauf(jfall{m: "T", set: "K2", einzel: true, body: v2Ein(`"A":1,"B":"y"`)})
	if r1.entity.Schluessel != r2.entity.Schluessel || r1.entity.Schluessel == r3.entity.Schluessel {
		t.Errorf("K3 zusammengesetzt: %q %q %q", r1.entity.Schluessel, r2.entity.Schluessel, r3.entity.Schluessel)
	}
}

func v2FaelleK() []jfall { return verbinde(v2FaelleK1(), v2FaelleK2()) }
