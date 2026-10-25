package baselineimport_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// dekodiereOrakel ist die zweite, von der Implementierung unabhängige Regel für den Namen einer Option
// (SPEC-127 Nr. 8 i): einmal `%XX`-dekodiert (Hexziffern groß oder klein); ist die Kodierung irgendwo im Text
// ungültig (`%` ohne zwei Hexziffern), bleibt der ganze Text unverändert.
func dekodiereOrakel(s string) string {
	hex := func(c byte) bool { return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' }
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && (i+2 >= len(s) || !hex(s[i+1]) || !hex(s[i+2])) {
			return s
		}
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			b.WriteByte(s[i])
			continue
		}
		n, _ := strconv.ParseUint(s[i+1:i+3], 16, 8)
		b.WriteByte(byte(n))
		i += 2
	}
	return b.String()
}

// orakelOption sagt, ob eine Option mit dem Namen name lesbar ist (SPEC-127 Nr. 8 i): die zehn Systemoptionen und
// jeder Name ohne `$` am Anfang, nach der Dekodierung.
func orakelOption(name string) bool {
	d := dekodiereOrakel(name)
	return !strings.HasPrefix(d, "$") || slices.Contains([]string{"$format", "$count", "$select", "$orderby", "$filter", "$top", "$skip", "$skiptoken", "$search", "$expand"}, d)
}

// hexForm schreibt das Byte c als `%XX` in großen oder kleinen Hexziffern.
func hexForm(c byte, klein bool) string {
	s := fmt.Sprintf("%%%02X", c)
	if klein {
		s = strings.ToLower(s)
	}
	return s
}

// kodierformen sind die neun Formen, in denen ein Name erscheint (Plan, Zeile E4): unkodiert, das erste Zeichen
// `%XX` groß, `%xx` klein, jedes Zeichen kodiert groß, klein, das erste Zeichen doppelt kodiert, `%zz`, `%`
// und `%4` angehängt.
func kodierformen(name string) []string {
	erstes := func(f func(byte) string) string {
		if name == "" {
			return name
		}
		return f(name[0]) + name[1:]
	}
	alle := func(klein bool) string {
		var b strings.Builder
		for i := 0; i < len(name); i++ {
			b.WriteString(hexForm(name[i], klein))
		}
		return b.String()
	}
	return []string{
		name,
		erstes(func(c byte) string { return hexForm(c, false) }),
		erstes(func(c byte) string { return hexForm(c, true) }),
		alle(false),
		alle(true),
		erstes(func(c byte) string { return "%25" + hexForm(c, false)[1:] }),
		name + "%zz",
		name + "%",
		name + "%4",
	}
}

// optionenE4 sind die 29 Optionen (Name und Wert) der Menge von E4; hatWert falsch heißt ein Segment ohne `=`.
func optionenE4() []struct {
	name, wert string
	hatWert    bool
} {
	var out []struct {
		name, wert string
		hatWert    bool
	}
	for _, o := range []string{"$format=json", "$count=true", "$select=Id", "$orderby=Id", "$filter=Id%20eq%201", "$top=1", "$skip=1", "$skiptoken=a", "$search=x", "$expand=Category",
		"sap-client=100", "@a=1", "x", "=v", "%zz=1", "$apply=x", "$compute=x", "$deltatoken=x", "$levels=1", "$schemaversion=1", "$index=1", "$batch=1", "$Top=1", "$TOP=1",
		"$Select=Id", "$FORMAT=json", "$unknown=1", "%24apply=x", "$apply%zz=1"} {
		name, wert, hat := strings.Cut(o, "=")
		out = append(out, struct {
			name, wert string
			hatWert    bool
		}{name, wert, hat})
	}
	return out
}

// TestZeileE4Kodierungen: die 29 Namen der Optionen in neun Formen der Kodierung mal drei Lagen mal beide
// Spuren (1566), dazu `;` im Wert (je Spur 1): 1568 Fälle; Orakel eine zweite Regelfunktion (einmal dekodieren,
// ein ungültiger Rest macht den ganzen Text unverändert).
func TestZeileE4Kodierungen(t *testing.T) {
	n, gelesen, uebersprungen := 0, 0, 0
	for _, s := range spuren() {
		for _, lage := range []string{"allein", "hinter", "vor"} {
			for _, o := range optionenE4() {
				for _, name := range kodierformen(o.name) {
					n++
					seg := name
					if o.hatWert {
						seg += "=" + o.wert
					}
					q := map[string]string{"allein": seg, "hinter": "$top=1&" + seg, "vor": seg + "&$top=1"}[lage]
					g := probe(t, s.pfad, q)
					if orakelOption(name) {
						gelesen++
						if !g.gelesen {
							t.Errorf("%s %s %q: %+v, erwartet gelesen", s.name, lage, q, g)
						}
					} else {
						uebersprungen++
						if g.grund != "query" {
							t.Errorf("%s %s %q: %+v, erwartet query", s.name, lage, q, g)
						}
					}
				}
			}
		}
		n++
		if g := probe(t, s.pfad, "$top=1;$apply"); !g.gelesen {
			t.Errorf("%s: `;` trennt nichts, aber %+v", s.name, g)
		}
	}
	if len(optionenE4()) != 29 || n != 1568 {
		t.Errorf("Namen %d, Menge %d, erwartet 29 und 1568", len(optionenE4()), n)
	}
	if gelesen == 0 || uebersprungen == 0 {
		t.Errorf("das Orakel ist trivial: %d gelesen, %d übersprungen", gelesen, uebersprungen)
	}
}

// orakelContentType urteilt nach SPEC-127 Nr. 8 g über die Parameter charset und odata.metadata: kein_json bei einem
// charset, das nicht utf-8 ist (Wert ohne ein umgebendes Paar von Anführungszeichen, ASCII-klein), danach
// ohne_kontext bei einem odata.metadata gleich none; sonst gelesen (leer).
func orakelContentType(charsets, metadatas []string) string {
	wert := func(v string) string {
		if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
			v = v[1 : len(v)-1]
		}
		return strings.ToLower(v)
	}
	for _, c := range charsets {
		if wert(c) != "utf-8" {
			return "kein_json"
		}
	}
	for _, m := range metadatas {
		if wert(m) == "none" {
			return "ohne_kontext"
		}
	}
	return ""
}

// folgen liefert alle Folgen der Länge 0 bis 3 über die Werte.
func folgen(werte ...string) [][]string {
	out := [][]string{nil}
	letzte := [][]string{nil}
	for l := 1; l <= 3; l++ {
		var naechste [][]string
		for _, f := range letzte {
			for _, w := range werte {
				naechste = append(naechste, append(slices.Clone(f), w))
			}
		}
		out = append(out, naechste...)
		letzte = naechste
	}
	return out
}

// TestZeileE3ParameterFolgen: Folgen der Parameter charset über {utf-8, iso-8859-1} der Länge 0 bis 3 (15) mal Folgen
// von odata.metadata über {none, minimal} der Länge 0 bis 3 (15) in zwei Anordnungen, dazu Anführungszeichen um
// charset (6) und um odata.metadata (4), je Spur; höchstens 920 nach dem Entdoppeln. Orakel: kein_json genau bei einem
// charset ungleich utf-8, ohne_kontext genau bei einem odata.metadata gleich none (SPEC-127 Nr. 8 g, jeder Parameter zählt).
func TestZeileE3ParameterFolgen(t *testing.T) {
	cs, om := folgen("utf-8", "iso-8859-1"), folgen("none", "minimal")
	if len(cs) != 15 || len(om) != 15 {
		t.Fatalf("Folgen %d und %d, erwartet 15 und 15", len(cs), len(om))
	}
	type fall struct {
		kopf     string
		erwartet string
	}
	baue := func(charsets, metadatas []string, charsetZuerst bool) string {
		var c, m string
		for _, v := range charsets {
			c += ";charset=" + v
		}
		for _, v := range metadatas {
			m += ";odata.metadata=" + v
		}
		if charsetZuerst {
			return "application/json" + c + m
		}
		return "application/json" + m + c
	}
	gesehen := map[string]bool{}
	var faelle []fall
	add := func(charsets, metadatas []string, charsetZuerst bool) {
		k := baue(charsets, metadatas, charsetZuerst)
		if !gesehen[k] {
			gesehen[k] = true
			faelle = append(faelle, fall{k, orakelContentType(charsets, metadatas)})
		}
	}
	for _, c := range cs {
		for _, m := range om {
			add(c, m, true)
			add(c, m, false)
		}
	}
	for _, c := range []string{`"utf-8"`, `"utf-8`, `utf-8"`, `"`, `""`, `"UTF-8"`} {
		add([]string{c}, nil, true)
	}
	for _, m := range []string{`"none"`, `"none`, `none"`, `""`} {
		add(nil, []string{m}, true)
	}
	if len(faelle) > 460 {
		t.Fatalf("%d Fälle je Spur, erwartet höchstens 460 (920 für beide Spuren)", len(faelle))
	}
	arten := map[string]int{}
	for _, s := range spuren() {
		for _, f := range faelle {
			arten[f.erwartet]++
			g := probe(t, s.pfad, "", ct(f.kopf))
			if (f.erwartet == "" && !g.gelesen) || g.grund != f.erwartet {
				t.Errorf("%s %q: %+v, erwartet %q", s.name, f.kopf, g, f.erwartet)
			}
		}
	}
	if arten[""] == 0 || arten["kein_json"] == 0 || arten["ohne_kontext"] == 0 {
		t.Errorf("das Orakel ist trivial: %v", arten)
	}
	t.Logf("E3: %d Fälle je Spur nach dem Entdoppeln (höchstens 460), Ausgänge %v", len(faelle), arten)
}

// TestZeileD1QueryFormat: der Name der Option in der Query des $metadata zählt nach der Dekodierung und byteweise
// (SPEC-127 Nr. 7 a, 8 i): acht Namen {$format, %24format, $Format, $FORMAT, $formatx, $forma, format, $format%zz} mal
// drei Lagen {allein, hinter $format=xml, vor $format=xml} (24); Kandidat genau bei $format und %24format.
func TestZeileD1QueryFormat(t *testing.T) {
	n := 0
	for _, name := range []string{"$format", "%24format", "$Format", "$FORMAT", "$formatx", "$forma", "format", "$format%zz"} {
		for _, lage := range []string{"allein", "hinter", "vor"} {
			n++
			seg := name + "=xml"
			q := map[string]string{"allein": seg, "hinter": "$format=xml&" + seg, "vor": seg + "&$format=xml"}[lage]
			w := neu(t)
			w.sitzung().metadaten("<edmx/>", func(e *model.Exchange) { e.RawQuery = q })
			_, err := w.importiere(anfrage())
			kandidat := dekodiereOrakel(name) == "$format"
			if kandidat != (err == nil) || (!kandidat && !errors.Is(err, baselineimport.ErrKeineMetadaten)) {
				t.Errorf("Query %q: Kandidat erwartet %v, Fehler %v", q, kandidat, err)
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// TestZeileE5SelectKodierung: der Wert von $select wird einmal %XX-dekodiert und an `,` geteilt; ist die Kodierung irgendwo
// ungültig, gilt der Text unverändert (SPEC-127 Nr. 8 h, i). Fünf Werte {Id, Id,Name, Name,Id, *, Name} in den neun Formen
// der Kodierung mal beide Spuren (90), außerhalb der 42 Formen; Orakel eine zweite Regelfunktion (dekodiereOrakel).
func TestZeileE5SelectKodierung(t *testing.T) {
	n, gelesen := 0, 0
	for _, s := range spuren() {
		for _, wert := range []string{"Id", "Id,Name", "Name,Id", "*", "Name"} {
			for _, form := range kodierformen(wert) {
				n++
				auswahl := strings.Split(dekodiereOrakel(form), ",")
				lesbar := slices.Contains(auswahl, "Id") || slices.Contains(auswahl, "*")
				if lesbar {
					gelesen++
				}
				g := probe(t, s.pfad, "$select="+form)
				if lesbar != g.gelesen || (!lesbar && g.grund != "projektion_ohne_schluessel") {
					t.Errorf("%s $select=%q: %+v, erwartet gelesen %v", s.name, form, g, lesbar)
				}
			}
		}
	}
	if n != 90 || gelesen == 0 || gelesen == n {
		t.Errorf("Menge %d, gelesen %d", n, gelesen)
	}
}
