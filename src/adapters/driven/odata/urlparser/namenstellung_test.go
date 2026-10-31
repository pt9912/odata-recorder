package urlparser_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// lage ist die Stellung eines Elements in einer Folge: das idx-te (ab 0) von n.
type lage struct{ n, idx int }

func (l lage) String() string {
	if l.n == 1 {
		return "einziges"
	}
	return fmt.Sprintf("%d. von %d", l.idx+1, l.n)
}

// listenLagen sind die Stellungen eines Elements in `$select`, `$orderby` und in der Query: einziges, erstes, mittleres, letztes.
func listenLagen() []lage { return []lage{{1, 0}, {3, 0}, {3, 1}, {3, 2}} }

// gliedLagen sind die Stellungen eines Glieds in einem Pfad `a/b/c` und in einem qualifizierten Namen `a.b.c`: erstes und letztes
// von zwei, erstes, mittleres und letztes von drei.
func gliedLagen() []lage { return []lage{{2, 0}, {2, 1}, {3, 0}, {3, 1}, {3, 2}} }

// teilLagen sind die Stellungen eines Teils im Körper eines Schlüssels: allein, erster und letzter von zwei, erster, mittlerer und
// letzter von drei.
func teilLagen() []lage { return []lage{{1, 0}, {2, 0}, {2, 1}, {3, 0}, {3, 1}, {3, 2}} }

// folge liefert die n Werte f(1) bis f(n), der an der Stelle idx durch ein ersetzt.
func folge(l lage, f func(int) string, ein string) []string {
	out := make([]string, l.n)
	for i := range out {
		out[i] = f(i + 1)
	}
	out[l.idx] = ein
	return out
}

func nummeriert(vor string) func(int) string {
	return func(i int) string { return vor + strconv.Itoa(i) }
}

// prozent schreibt jedes Byte von s als %XX.
func prozent(s string) string {
	var b strings.Builder
	for i := range len(s) {
		fmt.Fprintf(&b, "%%%02X", s[i])
	}
	return b.String()
}

// kodiereASCII schreibt jedes ASCII-Byte als %XX und lässt die Bytes ab 0x80 roh: ein rohes Strukturzeichen (`'`, `,`, `=`, `&`)
// würde den Aufbau ändern, ein rohes Zeichen jenseits von ASCII ist keines.
func kodiereASCII(s string) string {
	var b strings.Builder
	for i := range len(s) {
		if s[i] < 0x80 {
			fmt.Fprintf(&b, "%%%02X", s[i])
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// praefixOk meldet, ob s ein Typpräfix ist: nichtleer, Buchstaben (L), Ziffern (Nd), `_` und `.` in jeder Folge (SPEC-161 Nr. 6 g).
func praefixOk(s string) bool {
	for _, r := range s {
		if r != '.' && r != '_' && !unicode.Is(unicode.L, r) && !unicode.Is(unicode.Nd, r) {
			return false
		}
	}
	return s != ""
}

// praefixUrsache: das Literal `vor+m'x'` über 512 Bytes ist `zu_lang`, mit gültigem Präfix `literal_typ`, sonst `literal`.
func praefixUrsache(vor, m string) model.URLUrsache {
	switch voll := vor + m; {
	case len(voll)+3 > 512:
		return model.URLZuLang
	case praefixOk(voll):
		return model.URLLiteralTyp
	}
	return model.URLLiteral
}

// praefixFormen: das Präfix des Literals `<vor>m'x'` in jeder Stellung im Körper der Klammer (8 Formen).
func praefixFormen() []achsenOrt {
	bau := func(vor string, l lage, benannt bool) func(string, namenOrakel) fall {
		return func(m string, _ namenOrakel) fall {
			wert := vor + m + "'x'"
			if benannt {
				wert = "Id=" + wert
			}
			teile := folge(l, func(i int) string { return "n" + strconv.Itoa(i) + "=" + strconv.Itoa(i) }, wert)
			return fall{pfad: "/S(" + strings.Join(teile, ",") + ")", u: praefixUrsache(vor, m), ort: "pfad/1"}
		}
	}
	var fs []achsenOrt
	for _, l := range teilLagen() {
		fs = append(fs, achsenOrt{"Typpräfix: benannter Teil " + l.String(), bau("", l, true)})
	}
	return append(fs, achsenOrt{"Typpräfix: unbenannter Teil", bau("", lage{1, 0}, false)},
		achsenOrt{"Typpräfix: hinter einem Punkt", bau("Ns.", lage{1, 0}, true)})
}

// mengeFormen: der Name der Menge ohne Klammer, mit Klammer, mit leerer Klammer und vor `/$count` (4 Formen); ein Name, den die
// Regel verwirft, ist in jeder Form `segment` (SPEC-161 Nr. 4).
func mengeFormen() []achsenOrt {
	menge := func(m string, o namenOrakel) fall {
		gut := &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: m}}
		return nameOderNicht(o.istName(m), gut, model.URLSegment, "pfad/1", "/"+m, "")
	}
	klammer := func(m string, o namenOrakel) fall {
		return nameOderNicht(o.istName(m), entity(m, teil("", zahlLit("1"))), model.URLSegment, "pfad/1", "/"+m+"(1)", "")
	}
	leer := func(m string, o namenOrakel) fall {
		if o.istName(m) {
			return fall{pfad: "/" + m + "()", u: model.URLPfadNichtUnterstuetzt, ort: "pfad/1"}
		}
		return fall{pfad: "/" + m + "()", u: model.URLSegment, ort: "pfad/1"}
	}
	anzahl := func(m string, o namenOrakel) fall {
		gut := &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceAnzahl, EntitySet: m}}
		return nameOderNicht(o.istName(m), gut, model.URLSegment, "pfad/1", "/"+m+"/$count", "")
	}
	return []achsenOrt{
		{"Name der Menge: ohne Klammer", menge}, {"Name der Menge: mit Klammer", klammer},
		{"Name der Menge: mit leerer Klammer", leer}, {"Name der Menge: vor $count", anzahl},
	}
}

// schluesselFormen: der Name eines Schlüsselteils in jeder Stellung im Körper (6 Formen); die anderen Teile sind gültig.
func schluesselFormen() []achsenOrt {
	var fs []achsenOrt
	for _, l := range teilLagen() {
		fs = append(fs, achsenOrt{"Name eines Schlüsselteils: Teil " + l.String(), func(m string, o namenOrakel) fall {
			teile := make([]string, l.n)
			want := make([]model.SchluesselTeil, l.n)
			for i, n := range folge(l, nummeriert("n"), m) {
				wert := strconv.Itoa(i + 1)
				teile[i], want[i] = n+"="+wert, teil(n, zahlLit(wert))
			}
			return nameOderNicht(o.istName(m), entity("S", want...), model.URLKey, "pfad/1", "/S("+strings.Join(teile, ",")+")", "")
		}})
	}
	return fs
}

// optionFormen: der Name eines Parameter-Alias (`@` und ein Name) oder einer herstellerspezifischen Option an jeder Stelle der Query
// (4 Formen); wire schreibt den Namen in die URL, die Erwartung liest ihn dekodiert.
func optionFormen(alias bool, wire func(string) string) []achsenOrt {
	ort, vor := "Name einer herstellerspezifischen Option", ""
	if alias {
		ort, vor = "Alias", "@"
	}
	var fs []achsenOrt
	for _, l := range listenLagen() {
		fs = append(fs, achsenOrt{ort + ": Option " + l.String(), func(m string, o namenOrakel) fall {
			gut := o.istHersteller(m)
			if alias {
				gut = o.istName(m)
			}
			segmente := make([]string, l.n)
			sonstige := make([]model.SonstigeOption, l.n)
			for i := range segmente {
				f := "f" + strconv.Itoa(i+1)
				segmente[i], sonstige[i] = f+"=v", model.SonstigeOption{Name: f, Wert: "v"}
				if i == l.idx {
					segmente[i], sonstige[i] = vor+wire(m)+"=1", model.SonstigeOption{Name: vor + m, Wert: "1"}
				}
			}
			return nameOderNicht(gut, mengeMitOptionen(model.Optionen{Sonstige: sonstige}), model.URLOptionName,
				"query/"+strconv.Itoa(l.idx+1), "/S", strings.Join(segmente, "&"))
		}})
	}
	return fs
}

// listenWant ist das Ergebnis von `$select` oder `$orderby` mit den einfachen Namen; die Richtung gilt für das Element idx.
func listenWant(option string, namen []string, idx int, richtung string) *model.ODataURL {
	if option == "$select" {
		return mengeMitOptionen(model.Optionen{Select: &model.SelectOption{Namen: namen}})
	}
	sort := make([]model.Sortierung, len(namen))
	for i, n := range namen {
		sort[i] = model.Sortierung{Name: n, Absteigend: i == idx && richtung == "%20desc"}
	}
	return mengeMitOptionen(model.Optionen{OrderBy: sort})
}

// listenFall: der Name m als Element von `option` an der Stelle l der Liste, mit der Richtung (nur `$orderby`) und, wenn g.n > 0,
// als Glied g eines Pfads; die anderen Elemente und Glieder sind gültig. Ein Pfad aus Namen ist `wert_nicht_unterstuetzt`, jede
// andere Form `wert` (SPEC-162 Nr. 6, 7).
func listenFall(option string, l lage, richtung string, g lage) func(string, namenOrakel) fall {
	return func(m string, o namenOrakel) fall {
		element := m
		if g.n > 0 {
			element = strings.Join(folge(g, nummeriert("p"), m), "/")
		}
		f := fall{pfad: "/S", query: option + "=" + strings.Join(folge(l, nummeriert("n"), element+richtung), ",")}
		switch {
		case !o.istName(m):
			f.u, f.ort = model.URLWert, "query/1"
		case g.n > 0:
			f.u, f.ort = model.URLWertNichtUnterstuetzt, "query/1"
		default:
			f.want = listenWant(option, folge(l, nummeriert("n"), m), l.idx, richtung)
		}
		return f
	}
}

// listenFormen: die Elemente und die Pfad-Glieder von `option` in jeder Stellung und Form: 4 Stellungen im Wert × (der Name allein
// oder 5 Stellungen im Pfad) × die Richtungen; dazu das Element hinter einer anderen Option.
func listenFormen(option string, richtungen []string) []achsenOrt {
	var fs []achsenOrt
	for _, r := range richtungen {
		for _, l := range listenLagen() {
			fs = append(fs, achsenOrt{fmt.Sprintf("Element von %s: %s, Richtung %q", option, l, r), listenFall(option, l, r, lage{})})
			for _, g := range gliedLagen() {
				fs = append(fs, achsenOrt{fmt.Sprintf("Pfad-Glied in %s: %s, Glied %s, Richtung %q", option, l, g, r), listenFall(option, l, r, g)})
			}
		}
	}
	return append(fs, achsenOrt{"Element von " + option + ": hinter einer Option", func(m string, o namenOrakel) fall {
		f := fall{pfad: "/S", query: "f1=v&" + option + "=" + m}
		if !o.istName(m) {
			f.u, f.ort = model.URLWert, "query/2"
			return f
		}
		opt := listenWant(option, []string{m}, 0, "").Optionen
		opt.Sonstige = []model.SonstigeOption{{Name: "f1", Wert: "v"}}
		f.want = mengeMitOptionen(opt)
		return f
	}})
}

// restFormen: ein Glied eines qualifizierten Rest-Segments `a.b.c` in jeder Stellung, ohne und mit Klammer, als zweites und als
// drittes Segment des Pfads (30 Formen); ein Name macht es `pfad_nicht_unterstuetzt` (beim zweiten Segment, der erste Verstoß), sonst
// ist es `segment` am Ort des Segments (SPEC-161 Nr. 7).
func restFormen() []achsenOrt {
	var fs []achsenOrt
	for _, klammer := range []string{"", "(1)", "()"} {
		for _, g := range gliedLagen() {
			for _, seg := range []int{2, 3} {
				name := fmt.Sprintf("Glied eines Rest-Segments: Glied %s, Klammer %q, Segment %d", g, klammer, seg)
				fs = append(fs, achsenOrt{name, func(m string, o namenOrakel) fall {
					pfad := "/S/" + strings.Join(folge(g, nummeriert("p"), m), ".") + klammer
					if seg == 3 {
						pfad = "/S/q/" + strings.Join(folge(g, nummeriert("p"), m), ".") + klammer
					}
					if !o.istName(m) {
						return fall{pfad: pfad, u: model.URLSegment, ort: "pfad/" + strconv.Itoa(seg)}
					}
					return fall{pfad: pfad, u: model.URLPfadNichtUnterstuetzt, ort: "pfad/2"}
				}})
			}
		}
	}
	return fs
}

// stellenFormen sind alle Formen, in denen der Parser einen Namen liest: 10 Stellen im Code, 11 Orte, 155 Formen.
func stellenFormen() []achsenOrt {
	var fs []achsenOrt
	for _, gruppe := range [][]achsenOrt{
		praefixFormen(), mengeFormen(), schluesselFormen(), optionFormen(true, func(m string) string { return m }),
		optionFormen(false, func(m string) string { return m }),
		{{"nacktes Wort im Schlüssel: allein", nacktesWortFall}},
		listenFormen("$select", []string{""}), listenFormen("$orderby", []string{"", "%20asc", "%20desc"}), restFormen(),
	} {
		fs = append(fs, gruppe...)
	}
	return fs
}

// stellenMuster sind die Namen, die in jede Form eingesetzt werden: die 62 Muster von TestNamenAchsen (Zeichenanteil × Stellung im
// Namen, Länge in Zeichen) und je Unicode-Kategorie das erste Zeichen ab U+0080 (29): 91 Muster.
func stellenMuster(t *testing.T) []string {
	t.Helper()
	ms := achsenMuster()
	for _, z := range kategorieMenge(t) {
		ms = append(ms, z.bytes)
	}
	return ms
}

func zaehleOrte(formen []achsenOrt) map[string]int {
	z := map[string]int{}
	for _, f := range formen {
		ort, _, _ := strings.Cut(f.name, ": ")
		z[ort]++
	}
	return z
}

// TestNamenStellungen (Zeile 28, SPEC-161 Nr. 4 bis 7, SPEC-162 Nr. 3, 6, 7, SPEC-125 Nr. 8 b): „Name“ gilt an jeder Stelle des Orts,
// an dem der Parser ihn liest: erstes, mittleres, letztes und einziges Element einer Liste, jedes Glied eines Pfads, jeder Teil des
// Schlüssels, mit und ohne Klammer, mit und ohne Richtung. Schleife über 155 Formen (11 Orte) × 91 Muster = 14105 Elemente; die
// Erwartung kommt aus Ausdrücken des Wortlauts, nie aus dem Parser.
func TestNamenStellungen(t *testing.T) {
	o := neuerNamenOrakel()
	formen, muster := stellenFormen(), stellenMuster(t)
	n := 0
	for _, f := range formen {
		for _, m := range muster {
			pruefeMitPanik(t, fmt.Sprintf("%s %.40q", f.name, m), f.bau(m, o))
			n++
		}
	}
	erwartet := map[string]int{
		"Typpräfix": 8, "Name der Menge": 4, "Name eines Schlüsselteils": 6, "Alias": 4, "Name einer herstellerspezifischen Option": 4,
		"nacktes Wort im Schlüssel": 1, "Element von $select": 5, "Pfad-Glied in $select": 20, "Element von $orderby": 13,
		"Pfad-Glied in $orderby": 60, "Glied eines Rest-Segments": 30,
	}
	if !gleich(zaehleOrte(formen), erwartet) || len(formen) != 155 || len(muster) != 91 || n != 155*91 {
		t.Fatalf("Zählprobe: %v, %d Formen, %d Muster, %d Elemente, erwartet %v, 155, 91 und 14105", zaehleOrte(formen), len(formen), len(muster), n, erwartet)
	}
}

// TestNamenAsciiKodiert (Zeile 46, SPEC-162 Nr. 3 d, e): jedes ASCII-Zeichen von U+0001 bis U+007F, als %XX geschrieben, an jeder
// Stelle des Namens eines Alias und einer herstellerspezifischen Option (allein, erstes, inneres, letztes Zeichen) und in jeder
// Stellung der Option in der Query: 127 Zeichen × 4 Stellungen × 4 Stellungen × 2 Orte = 4064 Elemente.
func TestNamenAsciiKodiert(t *testing.T) {
	o := neuerNamenOrakel()
	formen := append(optionFormen(true, prozent), optionFormen(false, prozent)...)
	n := 0
	for c := rune(1); c < 0x80; c++ {
		for _, st := range []stellung{allein, erste, innere, letzte} {
			m := st.muster(string(c))
			for _, f := range formen {
				pruefeMitPanik(t, fmt.Sprintf("%s %q", f.name, m), f.bau(m, o))
				n++
			}
		}
	}
	if len(formen) != 8 || n != 127*4*8 {
		t.Fatalf("Zählprobe: %d Formen, %d Elemente, erwartet 8 und 4064", len(formen), n)
	}
}

// codepunkte ruft f für jeden Codepunkt von U+0001 bis U+10FFFF außer den Surrogaten (2048) und zählt sie.
func codepunkte(f func(rune)) int {
	n := 0
	for r := rune(1); r <= unicode.MaxRune; r++ {
		if r < 0xD800 || r > 0xDFFF {
			f(r)
			n++
		}
	}
	return n
}

// ursacheVon liest eine URL und liefert die Ursache des Verstoßes ("" ohne Verstoß) und das Ergebnis.
func ursacheVon(p *urlparser.Parser, pfad, query string) (model.URLUrsache, model.ODataURL) {
	u, err := p.ParseURL(pfad, query)
	var uf *model.URLFehler
	switch {
	case err == nil:
		return "", u
	case errors.As(err, &uf):
		return uf.Ursache, u
	}
	return "?", u
}

// mengeErwartung: der Name der Menge m (Klammer inbegriffen); ein Name, an den eine `(` ohne `)` grenzt, ist `key`, sonst `segment`.
func mengeErwartung(o namenOrakel, m string) model.URLUrsache {
	if o.istName(m) {
		return ""
	}
	if kopf, ok := strings.CutSuffix(m, "("); ok && o.istName(kopf) {
		return model.URLKey
	}
	return model.URLSegment
}

// TestNamenAlleCodepunkte (Zeile 28, SPEC-161 Nr. 2, 4, 6 g, SPEC-162 Nr. 2): jeder Codepunkt von U+0001 bis U+10FFFF außer den
// Surrogaten (1112063), nicht eine Auswahl, an den Funktionen, die ein Zeichen klassifizieren: als Name der Menge allein und hinter
// `a` (Buchstabe, Ziffer, `_`, sonst `segment`), als Typpräfix (ohne die fünf Strukturzeichen `'` `,` `=` `(` `)` des Körpers) und
// im Wert einer Option (gültig). ASCII steht als %XX, alles andere roh.
func TestNamenAlleCodepunkte(t *testing.T) {
	o, p := neuerNamenOrakel(), parserNeu()
	fehl := 0
	melde := func(ort string, r rune, ist, soll model.URLUrsache) {
		if fehl++; fehl <= 20 {
			t.Errorf("%s U+%04X: Ursache %q, erwartet %q", ort, r, ist, soll)
		}
	}
	var menge, praefix, wert int
	codepunkte(func(r rune) {
		z := string(r)
		for _, m := range []string{z, "a" + z} {
			menge++
			ist, _ := ursacheVon(p, "/"+kodiereASCII(m), "")
			if soll := mengeErwartung(o, m); ist != soll {
				melde("Name der Menge "+strconv.Quote(m), r, ist, soll)
			}
		}
		wert++
		if u, got := ursacheVon(p, "/S", "x="+kodiereASCII(z)); u != "" || len(got.Optionen.Sonstige) != 1 || got.Optionen.Sonstige[0].Wert != z {
			melde("Wert", r, u, "")
		}
		if strings.ContainsRune("',=()", r) {
			return
		}
		praefix++
		soll := model.URLLiteral
		if praefixOk(z) {
			soll = model.URLLiteralTyp
		}
		if ist, _ := ursacheVon(p, "/S(Id="+kodiereASCII(z)+"'x')", ""); ist != soll {
			melde("Typpräfix", r, ist, soll)
		}
	})
	if menge != 2*1112063 || praefix != 1112063-5 || wert != 1112063 {
		t.Fatalf("Zählprobe: %d Namen der Menge, %d Präfixe, %d Werte, erwartet 2224126, 1112058 und 1112063", menge, praefix, wert)
	}
}

// TestOptionsnamenSchreibweise (Zeile 46, SPEC-162 Nr. 3 a bis c: die Schreibweise zählt): jeder der 17 Namen mit `$` (7 behandelte,
// 10 nicht unterstützte) in der Schreibweise der Spezifikation (gelesen oder `option_nicht_unterstuetzt`, `$filter` ohne Leser
// `filter`) und in vier anderen (nur Großbuchstaben, nur der erste Buchstabe groß, nur der letzte groß, abwechselnd): `option_name`.
// 85 Elemente.
func TestOptionsnamenSchreibweise(t *testing.T) {
	werte := map[string]string{"$select": "Id", "$orderby": "Id", "$top": "1", "$skip": "1", "$count": "true", "$filter": "x", "$format": "json"}
	behandelt := namenBehandelt()
	n := 0
	for _, name := range append(namenBehandelt(), namenNichtUnterstuetzt()...) {
		buchstaben := []byte(name[1:])
		erste, letzte, abwechselnd := append([]byte(nil), buchstaben...), append([]byte(nil), buchstaben...), append([]byte(nil), buchstaben...)
		erste[0] -= 'a' - 'A'
		letzte[len(letzte)-1] -= 'a' - 'A'
		for i := 0; i < len(abwechselnd); i += 2 {
			abwechselnd[i] -= 'a' - 'A'
		}
		for _, schreibweise := range []string{
			"$" + string(buchstaben), "$" + strings.ToUpper(string(buchstaben)), "$" + string(erste), "$" + string(letzte), "$" + string(abwechselnd),
		} {
			f := fall{pfad: "/S", query: schreibweise + "=" + werte[name], u: model.URLOptionName, ort: "query/1"}
			if schreibweise == name {
				f.u, f.ort = "", ""
				switch {
				case name == "$filter":
					f.u, f.ort = model.URLFilter, "query/1"
				case !slices.Contains(behandelt, name):
					f.u, f.ort = model.URLOptionNichtUnterstuetzt, "query/1"
					f.query = schreibweise + "=v"
				}
			}
			pruefeFall(t, parserNeu(), f)
			n++
		}
	}
	if n != 85 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 85", n)
	}
}
