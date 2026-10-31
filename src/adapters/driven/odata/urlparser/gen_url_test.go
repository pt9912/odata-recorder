package urlparser_test

import (
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// genKey erzeugt den Körper einer Klammer: 1 bis maxTeile Teile, benannt (verschiedene Namen) oder, bei einem Teil, auch
// unbenannt; der Wert zuerst, dann die Schreibweise.
func genKey(r *rand.Rand, maxTeile int, arten []model.LiteralArt) (teile []string, want []model.SchluesselTeil) {
	n := 1 + r.IntN(maxTeile)
	benannt := n > 1 || r.IntN(2) == 0
	gesehen := map[string]bool{}
	for range n {
		lit := genLiteral(r, pick(r, arten...))
		name := ""
		if benannt {
			for name == "" || gesehen[name] {
				name = genName(r, 8)
			}
			gesehen[name] = true
		}
		text := lit.roh(r)
		if benannt {
			text = name + "=" + text
		}
		teile = append(teile, text)
		want = append(want, model.SchluesselTeil{Name: name, Literal: model.Literal{Art: lit.art, Text: lit.text}})
	}
	return teile, want
}

// genPfad erzeugt Pfad und erwartetes Ergebnis einer Art; segmente sind die Klartexte, aus denen er geschrieben ist.
func genPfad(r *rand.Rand, art model.RessourcenArt) (pfad string, want model.Ressource, segmente []string) {
	name := genName(r, 12)
	want = model.Ressource{Art: art, EntitySet: name}
	switch art {
	case model.RessourceMetadaten:
		want.EntitySet, segmente = "", []string{"$metadata"}
	case model.RessourceMenge:
		segmente = []string{name}
	case model.RessourceAnzahl:
		segmente = []string{name, "$count"}
	default:
		var teile []string
		teile, want.Schluessel = genKey(r, 5, alleArten())
		segmente = []string{name + "(" + strings.Join(teile, ",") + ")"}
	}
	roh := make([]string, len(segmente))
	for i, s := range segmente {
		roh[i] = kodiere(r, s, pfadZwingend)
	}
	return "/" + strings.Join(roh, "/"), want, segmente
}

func alleRessourcen() []model.RessourcenArt {
	return []model.RessourcenArt{model.RessourceMetadaten, model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl}
}

// optionKlassen sind die sechs Klassen der Systemoptionen des Generators.
func optionKlassen() []string {
	return []string{"select", "orderby", "toskip", "count", "format", "sonstige"}
}

type optionsPaar struct {
	name, wert string // Klartext
	setze      func(*model.Optionen)
}

func zufallsZahl(r *rand.Rand) (int, string) {
	v := pick(r, 0, 1, r.IntN(1000), r.IntN(1<<31), 1<<31-1)
	return v, strings.Repeat("0", r.IntN(3)) + strconv.Itoa(v)
}

// genOptionen erzeugt die Optionen einer Klasse als Paare in zufälliger Reihenfolge.
func genOptionen(r *rand.Rand, klasse string) []optionsPaar {
	switch klasse {
	case "alle": // alle sechs Klassen in einer Query, damit keine Option von der Reihenfolge der anderen abhängt
		var out []optionsPaar
		for _, k := range optionKlassen() {
			out = append(out, genOptionen(r, k)...)
		}
		return out
	case "select":
		return []optionsPaar{genSelect(r)}
	case "orderby":
		return []optionsPaar{genOrderBy(r)}
	case "toskip":
		v1, t1 := zufallsZahl(r)
		v2, t2 := zufallsZahl(r)
		return []optionsPaar{
			{"$top", t1, func(o *model.Optionen) { o.Top = model.ZahlOption{Gesetzt: true, Wert: v1} }},
			{"$skip", t2, func(o *model.Optionen) { o.Skip = model.ZahlOption{Gesetzt: true, Wert: v2} }},
		}
	case "count":
		b := r.IntN(2) == 0
		return []optionsPaar{{"$count", zufallsGross(r, strconv.FormatBool(b)), func(o *model.Optionen) { o.Count = model.WahrOption{Gesetzt: true, Wert: b} }}}
	case "format":
		return []optionsPaar{{"$format", pick(r, "json", "JSON", "application/json", "Application/Json", "application/json;odata.metadata=none", "application/json;"), func(*model.Optionen) {}}}
	}
	var out []optionsPaar
	for range 1 + r.IntN(3) {
		s := model.SonstigeOption{Name: pick(r, "sap-client", "x", "api-version", "a.b", "_y", "@p1", "@alias"), Wert: zeichenText(r, 12)}
		out = append(out, optionsPaar{s.Name, s.Wert, func(o *model.Optionen) { o.Sonstige = append(o.Sonstige, s) }})
	}
	return out
}

func genSelect(r *rand.Rand) optionsPaar {
	var elemente, namen []string
	alle := r.IntN(3) == 0
	if alle {
		elemente = append(elemente, "*")
	}
	gesehen := map[string]bool{}
	for range r.IntN(4) + btoi(!alle) {
		n := genName(r, 10)
		if !gesehen[n] {
			gesehen[n] = true
			elemente, namen = append(elemente, n), append(namen, n)
		}
	}
	if len(elemente) == 0 {
		elemente, namen = []string{"Id"}, []string{"Id"}
	}
	r.Shuffle(len(elemente), func(i, j int) { elemente[i], elemente[j] = elemente[j], elemente[i] })
	// Die Namen stehen im Ergebnis in der Reihenfolge der URL, ohne `*`.
	namen = namen[:0]
	for _, e := range elemente {
		if e != "*" {
			namen = append(namen, e)
		}
	}
	return optionsPaar{"$select", strings.Join(elemente, ","), func(o *model.Optionen) {
		o.Select = &model.SelectOption{Alle: alle, Namen: append([]string(nil), namen...)}
		if len(namen) == 0 {
			o.Select.Namen = nil
		}
	}}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func genOrderBy(r *rand.Rand) optionsPaar {
	var elemente []string
	var sort []model.Sortierung
	gesehen := map[string]bool{}
	for range 1 + r.IntN(4) {
		n := genName(r, 10)
		if gesehen[n] {
			continue
		}
		gesehen[n] = true
		richtung := pick(r, "", " asc", " desc", " ASC", " Desc", " aSc", " DESC")
		elemente = append(elemente, n+richtung)
		sort = append(sort, model.Sortierung{Name: n, Absteigend: strings.EqualFold(strings.TrimSpace(richtung), "desc")})
	}
	return optionsPaar{"$orderby", strings.Join(elemente, ","), func(o *model.Optionen) { o.OrderBy = sort }}
}

// genQuery schreibt die Paare in zufälliger Reihenfolge, mit Escapes und leeren Segmenten, und liefert die erwarteten
// Optionen. plain sind die Klartexte der Segmente in der Reihenfolge der URL (leere Segmente nicht eingerechnet).
func genQuery(r *rand.Rand, paare []optionsPaar) (raw string, want model.Optionen, plain [][2]string) {
	r.Shuffle(len(paare), func(i, j int) { paare[i], paare[j] = paare[j], paare[i] })
	var teile []string
	for _, p := range paare {
		p.setze(&want)
		plain = append(plain, [2]string{p.name, p.wert})
		teile = append(teile, kodiere(r, p.name, nameZwingend)+"="+kodiere(r, p.wert, wertZwingend))
		for r.IntN(4) == 0 {
			teile = append(teile, "")
		}
	}
	return strings.Join(teile, "&"), want, plain
}

// pruefeAbstieg hält die Schreibweise gegen net/url: Pfad je Segment mit PathUnescape, Query mit QueryUnescape nach
// Ersetzen von `+` durch `%2B`.
func pruefeAbstieg(t *testing.T, saat uint64, pfad string, segmente []string, query string, plain [][2]string) {
	t.Helper()
	if pfad != "" {
		roh := strings.Split(pfad[1:], "/")
		if len(roh) != len(segmente) {
			t.Fatalf("Saat %d: %d Segmente, erwartet %d", saat, len(roh), len(segmente))
		}
		for i, s := range roh {
			if got, err := url.PathUnescape(s); err != nil || got != segmente[i] {
				t.Fatalf("Saat %d: Segment %q entkodiert %q (%v), erwartet %q", saat, s, got, err, segmente[i])
			}
		}
	}
	var segs []string
	for _, s := range strings.Split(query, "&") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	if len(segs) != len(plain) {
		t.Fatalf("Saat %d: %d Segmente der Query, erwartet %d", saat, len(segs), len(plain))
	}
	for i, s := range segs {
		n, w, _ := strings.Cut(s, "=")
		gn, e1 := url.QueryUnescape(strings.ReplaceAll(n, "+", "%2B"))
		gw, e2 := url.QueryUnescape(strings.ReplaceAll(w, "+", "%2B"))
		if e1 != nil || e2 != nil || gn != plain[i][0] || gw != plain[i][1] {
			t.Fatalf("Saat %d: Segment %q entkodiert %q=%q (%v %v), erwartet %q", saat, s, gn, gw, e1, e2, plain[i])
		}
	}
}
