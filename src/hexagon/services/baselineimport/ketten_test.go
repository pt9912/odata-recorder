package baselineimport_test

import (
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
