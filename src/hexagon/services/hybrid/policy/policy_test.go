package policy_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// TestKonstanten bindet die Schlüssel und Ursachen an den Wortlaut von SPEC-261.
func TestKonstanten(t *testing.T) {
	paare := [][2]string{
		{policy.SchluesselUpstream, "HYBRID_FALLBACK"}, {policy.SchluesselHosts, "FALLBACK_ALLOW_HOSTS"},
		{policy.SchluesselSchreibfreigaben, "FALLBACK_WRITE_ALLOW"}, {policy.SchluesselReplayKategorien, "HYBRID_REPLAY_FALLBACK"},
		{policy.UrsacheLeer, "leer"}, {policy.UrsacheTrenner, "trenner"}, {policy.UrsacheMethode, "methode"},
		{policy.UrsacheMuster, "muster"}, {policy.UrsacheHost, "host"}, {policy.UrsacheKategorie, "kategorie"},
		{policy.UrsacheOhneHosts, "ohne-hosts"},
		{policy.KategorieUnsupported, "unsupported"}, {policy.KategorieCoverage, "coverage_insufficient"},
	}
	for _, p := range paare {
		if p[0] != p[1] {
			t.Errorf("%q, erwartet %q", p[0], p[1])
		}
	}
}

func TestKategorien(t *testing.T) {
	gut := []struct {
		eintrag    string
		unsup, cov bool
	}{
		{"unsupported", true, false},
		{"coverage_insufficient", false, true},
	}
	schlecht := []eintragUrsache{
		{"", policy.UrsacheLeer},
		{"Unsupported", policy.UrsacheKategorie}, {"UNSUPPORTED", policy.UrsacheKategorie}, {"none", policy.UrsacheKategorie},
		{"coverage-insufficient", policy.UrsacheKategorie}, {"unsupported ", policy.UrsacheKategorie},
		{"ODR-E6003", policy.UrsacheKategorie}, {"all", policy.UrsacheKategorie},
	}
	if len(gut) != 2 || len(schlecht) != 8 {
		t.Fatalf("Menge der Tabelle: %d Gutfälle und %d Verstöße, erwartet 2 und 8", len(gut), len(schlecht))
	}
	for _, g := range gut {
		if u := einzel(t, policy.SchluesselReplayKategorien, g.eintrag); u != "" {
			t.Errorf("Gutfall %q: New meldet %q", g.eintrag, u)
		}
		p, err := policy.New(policy.Werte{ReplayKategorien: []string{g.eintrag}})
		if err != nil {
			t.Fatalf("Gutfall %q: %v", g.eintrag, err)
		}
		if unsup, cov := p.ReplayFlags(); unsup != g.unsup || cov != g.cov {
			t.Errorf("Gutfall %q: unsupported %v coverage %v", g.eintrag, unsup, cov)
		}
	}
	// beide zusammen, in beiden Reihenfolgen
	for _, l := range [][]string{{"unsupported", "coverage_insufficient"}, {"coverage_insufficient", "unsupported"}} {
		p, err := policy.New(policy.Werte{ReplayKategorien: l})
		if err != nil {
			t.Fatalf("Liste %q: %v", l, err)
		}
		if unsup, cov := p.ReplayFlags(); !unsup || !cov {
			t.Errorf("Liste %q: unsupported %v coverage %v", l, unsup, cov)
		}
	}
	for _, s := range schlecht {
		if u := einzel(t, policy.SchluesselReplayKategorien, s.eintrag); u != s.ursache {
			t.Errorf("Verstoß %q: New meldet %q, erwartet %q", s.eintrag, u, s.ursache)
		}
	}
	// jedes Byte an jeder Stelle beider gültigen Einträge
	erwartet, gezaehlt := 0, 0
	for _, b := range []string{"unsupported", "coverage_insufficient"} {
		erwartet += zahlVarianten(b)
		for _, e := range varianten(b) {
			gezaehlt++
			soll := policy.UrsacheKategorie
			if e == "unsupported" || e == "coverage_insufficient" {
				soll = ""
			}
			if u := policy.PruefeKategorie(e); u != soll {
				t.Errorf("%q: Ursache %q, erwartet %q", e, u, soll)
			}
		}
	}
	if gezaehlt != erwartet {
		t.Fatalf("Zahl der Elemente %d, erwartet %d", gezaehlt, erwartet)
	}
}

type erwFehler struct {
	schluessel string
	eintrag    int
	ursache    string
}

// pruefeFehler wirft jeden Fehler von New in seine Teile und vergleicht Zahl, Reihenfolge,
// Schlüssel, Nummer und Ursache; die Policy ist nil.
func pruefeFehler(t *testing.T, name string, w policy.Werte, soll []erwFehler) {
	t.Helper()
	p, err := policy.New(w)
	if p != nil {
		t.Errorf("%s: Policy trotz Fehler", name)
	}
	if err == nil {
		t.Errorf("%s: kein Fehler", name)
		return
	}
	var ist []erwFehler
	for _, f := range zerlege(err) {
		var kf *policy.KonfigFehler
		if !errors.As(f, &kf) {
			t.Errorf("%s: %T ist kein *KonfigFehler", name, f)
			return
		}
		ist = append(ist, erwFehler{kf.Schluessel, kf.Eintrag, kf.Ursache})
	}
	if !slices.Equal(ist, soll) {
		t.Errorf("%s: Fehler %v, erwartet %v", name, ist, soll)
	}
	var erster *policy.KonfigFehler
	if !errors.As(err, &erster) || erster.Schluessel != soll[0].schluessel {
		t.Errorf("%s: errors.As findet nicht den ersten Fehler", name)
	}
}

// liste beschreibt einen Schlüssel: ein gültiger Eintrag und ein Verstoß mit seiner Ursache.
type liste struct {
	schluessel, gueltig, erster, ursache string
}

func listen() []liste {
	return []liste{
		{policy.SchluesselHosts, "h", "http://h", policy.UrsacheHost},
		{policy.SchluesselSchreibfreigaben, "POST /x", "GET /x", policy.UrsacheMethode},
		{policy.SchluesselReplayKategorien, "unsupported", "none", policy.UrsacheKategorie},
	}
}

// werteMit belegt die Liste des Schlüssels; ohne Belegung hat Hosts den Eintrag h.
func werteMit(upstream bool, belegung map[string][]string) policy.Werte {
	w := policy.Werte{Upstream: upstream, Hosts: []string{"h"}}
	for k, v := range belegung {
		switch k {
		case policy.SchluesselHosts:
			w.Hosts = v
		case policy.SchluesselSchreibfreigaben:
			w.Schreibfreigaben = v
		case policy.SchluesselReplayKategorien:
			w.ReplayKategorien = v
		}
	}
	return w
}

func TestNewFehler(t *testing.T) {
	n := 0
	// Liste allein x Upstream {aus, an}: 6
	for _, l := range listen() {
		for _, up := range []bool{false, true} {
			n++
			pruefeFehler(t, "allein "+l.schluessel, werteMit(up, map[string][]string{l.schluessel: {l.erster}}),
				[]erwFehler{{l.schluessel, 1, l.ursache}})
		}
	}
	// zwei Verstöße an den Stellen 1 und 3 (der zweite ein leerer Eintrag, also mit anderer Ursache)
	// sowie 2 und 3, und ein gültiger Eintrag vor dem Verstoß: 3 + 3 + 3
	for _, l := range listen() {
		n += 3
		pruefeFehler(t, "1 und 3 "+l.schluessel, werteMit(false, map[string][]string{l.schluessel: {l.erster, l.gueltig, ""}}),
			[]erwFehler{{l.schluessel, 1, l.ursache}})
		pruefeFehler(t, "2 und 3 "+l.schluessel, werteMit(false, map[string][]string{l.schluessel: {l.gueltig, l.erster, ""}}),
			[]erwFehler{{l.schluessel, 2, l.ursache}})
		pruefeFehler(t, "gültig davor "+l.schluessel, werteMit(true, map[string][]string{l.schluessel: {l.gueltig, l.erster}}),
			[]erwFehler{{l.schluessel, 2, l.ursache}})
	}
	// Paare: 3, alle drei: 1; die Reihenfolge der Schlüssel ist Hosts, Schreibfreigaben, Kategorien
	ls := listen()
	for i := 0; i < len(ls); i++ {
		for j := i + 1; j < len(ls); j++ {
			n++
			a, b := ls[i], ls[j]
			pruefeFehler(t, "Paar "+a.schluessel+" "+b.schluessel,
				werteMit(false, map[string][]string{a.schluessel: {a.gueltig, a.erster}, b.schluessel: {b.erster}}),
				[]erwFehler{{a.schluessel, 2, a.ursache}, {b.schluessel, 1, b.ursache}})
		}
	}
	n++
	pruefeFehler(t, "alle drei", werteMit(true, map[string][]string{
		policy.SchluesselHosts: {"http://h"}, policy.SchluesselSchreibfreigaben: {"GET /x"}, policy.SchluesselReplayKategorien: {"none"}}),
		[]erwFehler{
			{policy.SchluesselHosts, 1, policy.UrsacheHost},
			{policy.SchluesselSchreibfreigaben, 1, policy.UrsacheMethode},
			{policy.SchluesselReplayKategorien, 1, policy.UrsacheKategorie},
		})
	if n != 19 {
		t.Fatalf("Zahl der Fälle %d, erwartet 19", n)
	}
	// zwei Verstöße mit verschiedener Ursache an den Stellen 2 und 3 (zweiter Verstoß leer)
	for _, l := range listen() {
		pruefeFehler(t, "leer an 2 "+l.schluessel, werteMit(false, map[string][]string{l.schluessel: {l.gueltig, "", l.erster}}),
			[]erwFehler{{l.schluessel, 2, policy.UrsacheLeer}})
	}
}

func TestFehlerText(t *testing.T) {
	const marker = "GEHEIM-7f3a"
	faelle := []struct {
		werte      policy.Werte
		schluessel string
		ursache    string
	}{
		{policy.Werte{Hosts: []string{"h", "h", marker + "/x"}}, policy.SchluesselHosts, policy.UrsacheHost},
		{policy.Werte{Schreibfreigaben: []string{"POST /x", "POST /x", marker}}, policy.SchluesselSchreibfreigaben, policy.UrsacheTrenner},
		{policy.Werte{Schreibfreigaben: []string{"POST /x", "POST /x", marker + " /x"}}, policy.SchluesselSchreibfreigaben, policy.UrsacheMethode},
		{policy.Werte{Schreibfreigaben: []string{"POST /x", "POST /x", "POST " + marker}}, policy.SchluesselSchreibfreigaben, policy.UrsacheMuster},
		{policy.Werte{ReplayKategorien: []string{"unsupported", "unsupported", marker}}, policy.SchluesselReplayKategorien, policy.UrsacheKategorie},
	}
	if len(faelle) != 5 {
		t.Fatalf("Zahl der Fälle %d, erwartet 5", len(faelle))
	}
	for _, f := range faelle {
		_, err := policy.New(f.werte)
		if err == nil {
			t.Fatalf("%s: kein Fehler", f.ursache)
		}
		fehler := zerlege(err)
		var kf *policy.KonfigFehler
		if len(fehler) != 1 || !errors.As(fehler[0], &kf) {
			t.Fatalf("%s: %v", f.ursache, fehler)
		}
		for _, text := range []string{err.Error(), kf.Error()} {
			if strings.Contains(text, marker) || strings.Contains(text, "7f3a") {
				t.Errorf("%s: der Text nennt den Wert: %q", f.ursache, text)
			}
			if !strings.Contains(text, f.schluessel) || !strings.Contains(text, f.ursache) || !strings.Contains(text, "3") {
				t.Errorf("%s: der Text %q nennt nicht Schlüssel, Nummer 3 und Ursache", f.ursache, text)
			}
		}
		if kf.Schluessel != f.schluessel || kf.Eintrag != 3 || kf.Ursache != f.ursache {
			t.Errorf("%s: %+v", f.ursache, *kf)
		}
	}
}

func TestZusammenhaenge(t *testing.T) {
	ohne := []erwFehler{{policy.SchluesselUpstream, 0, policy.UrsacheOhneHosts}}
	host := []erwFehler{{policy.SchluesselHosts, 1, policy.UrsacheHost}}
	faelle := []struct {
		name  string
		werte policy.Werte
		soll  []erwFehler // nil: kein Fehler
	}{
		{"aus, Hosts nil", policy.Werte{}, nil},
		{"aus, Hosts leer", policy.Werte{Hosts: []string{}}, nil},
		{"aus, Hosts gültig", policy.Werte{Hosts: []string{"h"}}, nil},
		{"aus, Hosts ungültig", policy.Werte{Hosts: []string{"http://h"}}, host},
		{"an, Hosts nil", policy.Werte{Upstream: true}, ohne},
		{"an, Hosts leer", policy.Werte{Upstream: true, Hosts: []string{}}, ohne},
		{"an, Hosts gültig", policy.Werte{Upstream: true, Hosts: []string{"h"}}, nil},
		{"an, Hosts ungültig", policy.Werte{Upstream: true, Hosts: []string{"http://h"}}, host},
		{"an, Hosts leerer Eintrag", policy.Werte{Upstream: true, Hosts: []string{""}}, []erwFehler{{policy.SchluesselHosts, 1, policy.UrsacheLeer}}},
		{"aus, Freigaben und Kategorien ohne Hosts", policy.Werte{Schreibfreigaben: []string{"POST /x"}, ReplayKategorien: []string{"unsupported"}}, nil},
		{"an, Freigabe ohne Host im Eintrag", policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"POST /svc/x"}}, nil},
		{"an ohne Hosts mit Verstoß in einer anderen Liste", policy.Werte{Upstream: true, Schreibfreigaben: []string{"GET /x"}},
			[]erwFehler{{policy.SchluesselSchreibfreigaben, 1, policy.UrsacheMethode}, {policy.SchluesselUpstream, 0, policy.UrsacheOhneHosts}}},
		{"an ohne Hosts mit Verstoß in zwei anderen Listen", policy.Werte{Upstream: true, Schreibfreigaben: []string{"GET /x"}, ReplayKategorien: []string{"none"}},
			[]erwFehler{
				{policy.SchluesselSchreibfreigaben, 1, policy.UrsacheMethode},
				{policy.SchluesselReplayKategorien, 1, policy.UrsacheKategorie},
				{policy.SchluesselUpstream, 0, policy.UrsacheOhneHosts},
			}},
	}
	for _, f := range faelle {
		if f.soll != nil {
			pruefeFehler(t, f.name, f.werte, f.soll)
			continue
		}
		if p, err := policy.New(f.werte); err != nil || p == nil {
			t.Errorf("%s: %v %v", f.name, p, err)
		}
	}
}
