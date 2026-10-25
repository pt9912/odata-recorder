package baselineimport_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// verstossP4 ist einer der zwanzig Verstöße einer Kette aus vier Seiten (Plan, Zusatztabelle P4).
type verstossP4 struct {
	name  string
	seite int // 1 bis 4; 0 für anzahl_abweichend
	rang  int // Nummer der Prüfung von SPEC-127 Nr. 12 b
}

func verstoesseP4() []verstossP4 {
	var out []verstossP4
	for p := 2; p <= 4; p++ {
		out = append(out, verstossP4{"seite_nicht_lesbar", p, 1})
	}
	for p := 1; p <= 4; p++ {
		out = append(out, verstossP4{"key_redacted", p, 2})
	}
	for p := 1; p <= 3; p++ {
		out = append(out, verstossP4{"link_redigiert", p, 3}, verstossP4{"link_nicht_aufloesbar", p, 5}, verstossP4{"schleife", p, 6}, verstossP4{"seite_fehlt", p, 7})
	}
	return append(out, verstossP4{"anzahl_abweichend", 0, 8})
}

func (v verstossP4) linkVerstoss() bool { return v.rang >= 3 && v.rang <= 7 }

// herstellbar sagt, ob beide Verstöße zugleich bestehen (die 32 Paare des Plans fehlen).
func herstellbar(a, b verstossP4) bool {
	if a.seite == b.seite && a.linkVerstoss() && b.linkVerstoss() {
		return false
	}
	for _, x := range [][2]verstossP4{{a, b}, {b, a}} {
		if x[0].name == "seite_nicht_lesbar" && x[1].linkVerstoss() && x[0].seite == x[1].seite {
			return false
		}
		if x[0].name == "seite_fehlt" && (x[1].name == "seite_nicht_lesbar" || x[1].name == "key_redacted") && x[1].seite == x[0].seite+1 {
			return false
		}
	}
	return true
}

func frueher(a, b verstossP4) bool {
	sa, sb := a.seite, b.seite
	if sa == 0 {
		sa = 99
	}
	if sb == 0 {
		sb = 99
	}
	return sa < sb || sa == sb && a.rang < b.rang
}

// weltP4 baut die Kette aus vier Seiten mit den Verstößen.
func weltP4(t *testing.T, vs ...verstossP4) *welt {
	w := neu(t)
	spec := standardSeiten(4)
	for _, v := range vs {
		i := v.seite - 1
		switch v.name {
		case "seite_nicht_lesbar":
			schluessel := w.matchKey("/Products", queryVon(i))
			spec[i].aend = func(e *model.Exchange) { e.ResponseStatus, e.MatchKey = 404, schluessel }
		case "key_redacted":
			alt := spec[i].aend
			spec[i].aend = func(e *model.Exchange) {
				if alt != nil {
					alt(e)
				}
				e.KeyRedacted = true
			}
		case "link_redigiert":
			spec[i].link, spec[i].redigiert = "", true
		case "link_nicht_aufloesbar":
			spec[i].link = "mailto:x@y"
		case "schleife":
			spec[i].link = linkZu(0)
		case "seite_fehlt":
			spec[i].link = "Products?$skiptoken=77"
		case "anzahl_abweichend":
			spec[0].anzahl = int64p(99)
		}
	}
	w.standard().kette(spec...)
	return w
}

// TestZeileH15RangfolgeDerKette: die 20 einzelnen Verstöße und die 158 herstellbaren Paare; erwartet der
// Verstoß an der früheren Seite, an derselben Seite der früheren Prüfung, anzahl_abweichend zuletzt.
func TestZeileH15RangfolgeDerKette(t *testing.T) {
	alle := verstoesseP4()
	if len(alle) != 20 {
		t.Fatalf("Verstöße %d", len(alle))
	}
	pruefe := func(vs ...verstossP4) {
		t.Helper()
		erwartet := vs[0]
		for _, v := range vs[1:] {
			if frueher(v, erwartet) {
				erwartet = v
			}
		}
		w := weltP4(t, vs...)
		mustImport(t, w, anfrage())
		status, _, ketten := ketteVon(t, w)
		if len(ketten) != 1 || endeVon(ketten[0]) != erwartet.name || status != "partial" {
			t.Errorf("%v: Ende %v, erwartet %s", vs, ketten, erwartet.name)
		}
	}
	for _, v := range alle {
		pruefe(v)
	}
	paare := 0
	for i := range alle {
		for j := i + 1; j < len(alle); j++ {
			if herstellbar(alle[i], alle[j]) {
				paare++
				pruefe(alle[i], alle[j])
			}
		}
	}
	if paare != 158 {
		t.Errorf("Paare %d, erwartet 158", paare)
	}
	// Ergänzung: eine gelesene Seite eines anderen Sets mit key_redacted ist nicht lesbar, bevor ihr Key zählt.
	w := neu(t)
	spec := standardSeiten(4)
	schluessel := w.matchKey("/Products", queryVon(1))
	spec[1].aend = func(e *model.Exchange) { e.Path, e.MatchKey, e.KeyRedacted = "/Orders", schluessel, true }
	w.standard().kette(spec...)
	mustImport(t, w, anfrage())
	if _, _, ketten := ketteVon(t, w); len(ketten) != 1 || endeVon(ketten[0]) != "seite_nicht_lesbar" {
		t.Errorf("Seite eines anderen Sets mit key_redacted: %v", ketten)
	}
}

// --- H14: Generator S und der naive Beweisprüfer ------------------------------------------------

// seiteS ist die Absicht des Generators für einen Exchange der Session.
type seiteS struct {
	lesbar       bool
	start        bool
	ids          []int
	anzahl       int64 // -1: keine
	link         bool  // die Seite trägt einen Link
	redigiert    bool  // der Link ist redigiert
	ziel         string
	ok           bool // der Link wird aufgelöst, ziel ist der Match-Key, auf den er zeigt
	keyRedigiert bool
}

// orakelS ist der naive Beweisprüfer: er geht paarweise über die Liste aller Exchanges, ohne Index.
func orakelS(exs []model.Exchange, absicht map[string]seiteS) (enden []string, complete bool) {
	var mengen []map[int]bool
	alleBelegt := true
	for _, start := range exs {
		if a, ok := absicht[start.ID]; !ok || !a.start || !a.lesbar {
			continue
		}
		cur := start
		var gesehen []string
		keys := map[int]bool{}
		ende := ""
		for ende == "" {
			a := absicht[cur.ID]
			switch {
			case !a.lesbar:
				ende = "seite_nicht_lesbar"
			case a.keyRedigiert:
				ende = "key_redacted"
			default:
				for _, id := range a.ids {
					keys[id] = true
				}
				gesehen = append(gesehen, cur.MatchKey)
				ende = naechste(&cur, exs, a, gesehen)
			}
		}
		if ende == "proven" {
			if a := absicht[start.ID]; a.anzahl >= 0 && a.anzahl != int64(len(keys)) {
				ende = "anzahl_abweichend"
			}
		}
		enden = append(enden, ende)
		if ende != "proven" {
			alleBelegt = false
		}
		mengen = append(mengen, keys)
	}
	complete = len(enden) > 0 && alleBelegt
	for _, m := range mengen {
		if fmt.Sprint(m) != fmt.Sprint(mengen[0]) {
			complete = false
		}
	}
	return enden, complete
}

// naechste setzt *cur auf die nächste Seite oder liefert das Ende; die Suche nach dem kleinsten Treffer
// geht paarweise über alle Exchanges.
func naechste(cur *model.Exchange, exs []model.Exchange, a seiteS, gesehen []string) string {
	switch {
	case a.redigiert:
		return "link_redigiert"
	case !a.link:
		return "proven"
	case !a.ok:
		return "link_nicht_aufloesbar"
	case slices.Contains(gesehen, a.ziel):
		return "schleife"
	}
	var treffer *model.Exchange
	for i := range exs {
		x := &exs[i]
		if x.MatchKey == a.ziel && x.SequenceNo > cur.SequenceNo && (treffer == nil || x.SequenceNo < treffer.SequenceNo) {
			treffer = x
		}
	}
	if treffer == nil {
		return "seite_fehlt"
	}
	*cur = *treffer
	return ""
}

// TestZeileH14Allaussage: Generator S (5000 Sessions, fester Seed): complete genau dann, wenn ein Beweis
// besteht, in beiden Richtungen, und jede Kette endet wie das Orakel es bestimmt.
func TestZeileH14Allaussage(t *testing.T) {
	zufall := rand.New(rand.NewPCG(42, 4711))
	positiv := []string{"Products?$skiptoken=%d", "/odata/v4/Products?$skiptoken=%d", "https://up.example/odata/v4/Products?$skiptoken=%d", "./Products?$skiptoken=%d", "../v4/Products?$skiptoken=%d", "//up.example/odata/v4/Products?$skiptoken=%d#f"}
	negativ := []string{"mailto:x@y", "https://other.example/odata/v4/Products?$skiptoken=%d", "/odata/v42/Products?$skiptoken=%d", "ftp://up.example/odata/v4/Products?$skiptoken=%d"}
	queries := []string{"", "$format=json", "$count=true"}
	vollstaendig, unvollstaendig := 0, 0
	for sess := 0; sess < 5000; sess++ {
		w := neu(t)
		b := w.standard()
		absicht := map[string]seiteS{}
		nStarts := zufall.IntN(4)
		for c := 0; c < nStarts; c++ {
			ids := []int{1, 2, 3}
			if zufall.IntN(5) == 0 {
				ids = []int{1, 2, 4}
			}
			seiten := 1 + zufall.IntN(5)
			defekt := zufall.IntN(10) // 1 bis 5 je ein Defekt, sonst keiner
			pos := zufall.IntN(seiten)
			var tokens []int
			for p := 0; p < seiten; p++ {
				tokens = append(tokens, 10*(c+1)+p)
			}
			// Vorläufer mit den Keys der späteren Seiten, vor dem Beginn der Kette: sie sind nie die nächste Seite.
			for p := 1; p < seiten; p++ {
				if zufall.IntN(3) == 0 {
					b.collection("Products", fmt.Sprintf("$skiptoken=%d", tokens[p]), seite(produkt(9, "x", "1")))
				}
			}
			for p := 0; p < seiten; p++ {
				a := seiteS{lesbar: true, start: p == 0, anzahl: -1}
				if p == 0 {
					a.ids = slices.Clone(ids)
				} else {
					a.ids = []int{ids[p%len(ids)]}
				}
				query := queries[c]
				if p > 0 {
					query = fmt.Sprintf("$skiptoken=%d", tokens[p])
				}
				sp := seite(produkteVon(a.ids)...)
				var aend aenderung
				if p < seiten-1 {
					a.link, a.ok = true, true
					sp = mitLink(sp, fmt.Sprintf(positiv[zufall.IntN(len(positiv))], tokens[p+1]))
					a.ziel = w.matchKey("/Products", fmt.Sprintf("$skiptoken=%d", tokens[p+1]))
					if defekt == 1 && p == pos {
						sp = mitLink(seite(produkteVon(a.ids)...), fmt.Sprintf(negativ[zufall.IntN(len(negativ))], tokens[p+1]))
						a.ok = false
					}
					if defekt == 2 && p == pos {
						sp, a.redigiert = redigierterLink(seite(produkteVon(a.ids)...)), true
					}
					if defekt == 3 && p == pos { // Schleife auf den Beginn der Kette
						a.ziel = w.matchKey("/Products", queries[c])
						sp = mitLink(seite(produkteVon(a.ids)...), "Products")
						if c > 0 {
							sp = mitLink(seite(produkteVon(a.ids)...), "Products?"+queries[c])
						}
					}
				}
				if defekt == 4 && p == pos && p > 0 {
					a.lesbar = false
					aend = status(404)
				}
				if defekt == 5 && p == pos {
					a.keyRedigiert = true
					aend = func(e *model.Exchange) { e.KeyRedacted = true }
				}
				if p == 0 && zufall.IntN(4) == 0 {
					a.anzahl = int64(len(ids))
					if zufall.IntN(3) == 0 {
						a.anzahl++
					}
					sp = mitAnzahl(sp, a.anzahl)
				}
				e := b.collection("Products", query, sp, aend)
				absicht[e.ID] = a
			}
		}
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		enden, complete := orakelS(w.exchanges.m[sitzungID], absicht)
		var got []string
		for _, k := range ketten {
			got = append(got, endeVon(k))
		}
		if !slices.Equal(got, enden) {
			t.Fatalf("Session %d: Enden %v, Orakel %v", sess, got, enden)
		}
		if (status == "complete") != complete {
			t.Fatalf("Session %d: Status %s/%s, Orakel complete %v (%v)", sess, status, grund, complete, enden)
		}
		if complete {
			vollstaendig++
		} else {
			unvollstaendig++
		}
	}
	if vollstaendig < 300 || unvollstaendig < 300 {
		t.Errorf("der Generator erreicht nur %d vollständige und %d unvollständige Sessions", vollstaendig, unvollstaendig)
	}
}
