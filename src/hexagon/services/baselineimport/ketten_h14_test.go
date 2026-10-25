package baselineimport_test

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// --- H14: Generator S und der naive Beweisprüfer ------------------------------------------------

// seiteS ist die Absicht des Generators für einen Exchange der Session.
type seiteS struct {
	gelesen      bool  // der Codec liest die Antwort als Products (Collection oder Entity)
	lesbar       bool  // die Antwort ist eine gelesene Collection von Products, also eine Seite einer Kette
	start        bool  // die Antwort beginnt eine Kette
	bestand      []int // die Schlüssel, die sie zum Bestand von Products beiträgt
	ids          []int // die Schlüssel der Elemente
	anzahl       int64 // -1: keine
	link         bool  // die Seite trägt einen Link
	redigiert    bool  // der Link ist redigiert
	ziel         string
	ok           bool // der Link wird aufgelöst, ziel ist der Match-Key, auf den er zeigt
	keyRedigiert bool
}

// sollS ist das Urteil des Orakels über eine Session.
type sollS struct {
	ketten []sollKette
	status string
	grund  string
}

type sollKette struct {
	ende       string
	beobachtet int
	anzahl     int64
}

func gleicheIDs(a, b map[int]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// orakelS ist der naive Beweisprüfer: er geht paarweise über die Liste aller Exchanges, ohne Index, und
// bildet den Bestand von Products aus allen gelesenen Antworten (SPEC-127 Nr. 12 g).
func orakelS(exs []model.Exchange, absicht map[string]seiteS) sollS {
	var soll sollS
	bestand := map[int]bool{}
	gelesen := 0
	for _, e := range exs {
		if a := absicht[e.ID]; a.gelesen {
			gelesen++
			for _, id := range a.bestand {
				bestand[id] = true
			}
		}
	}
	var mengen []map[int]bool
	alleBelegt := true
	for _, start := range exs {
		if a, ok := absicht[start.ID]; !ok || !a.start || !a.lesbar {
			continue
		}
		k, keys := laufeOrakel(start, exs, absicht)
		soll.ketten = append(soll.ketten, k)
		alleBelegt = alleBelegt && k.ende == "proven"
		mengen = append(mengen, keys)
	}
	soll.status, soll.grund = urteilS(gelesen, len(soll.ketten) > 0, alleBelegt, mengen, bestand)
	return soll
}

// laufeOrakel geht die Kette von start Seite für Seite und liefert ihr Ende und ihre Schlüssel.
func laufeOrakel(start model.Exchange, exs []model.Exchange, absicht map[string]seiteS) (sollKette, map[int]bool) {
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
	anzahl := absicht[start.ID].anzahl
	if ende == "proven" && anzahl >= 0 && anzahl != int64(len(keys)) {
		ende = "anzahl_abweichend"
	}
	return sollKette{ende, len(keys), anzahl}, keys
}

// urteilS bestimmt Status und Grund der Coverage aus dem Nachweis (Nr. 12 g).
func urteilS(gelesen int, hatKette, alleBelegt bool, mengen []map[int]bool, bestand map[int]bool) (string, string) {
	abweichend := false
	for _, m := range mengen {
		abweichend = abweichend || !gleicheIDs(m, mengen[0])
	}
	switch {
	case gelesen == 0:
		return "unknown", "keine_beobachtung"
	case !hatKette:
		return "partial", "keine_ungefilterte_abfrage"
	case !alleBelegt:
		return "partial", "kette_nicht_belegt"
	case abweichend:
		return "partial", "ketten_abweichend"
	case !gleicheIDs(mengen[0], bestand):
		return "partial", "bestand_ausserhalb_kette"
	}
	return "complete", "kette_belegt"
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

// schluesselMit bildet den Match-Key von GET pfad?query mit den Headern kopf.
func (w *welt) schluesselMit(pfad, query string, kopf map[string][]string) string {
	k, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: dienstID, Path: pfad, RawQuery: query, Header: kopf, RequestRedaction: model.RedactionNone}, w.konfig.Regeln)
	if err != nil {
		w.t.Fatal(err)
	}
	return k.MatchKey
}

// beziehungK sind zwei Schlüsselmengen in einer der acht geordneten Beziehungen (Plan, H10 b).
type beziehungK struct {
	name   string
	k1, k2 []int
}

func h14Beziehungen() []beziehungK {
	return []beziehungK{
		{"gleich", []int{1, 2, 3}, []int{1, 2, 3}}, {"K1 in K2", []int{1, 2}, []int{1, 2, 3}}, {"K2 in K1", []int{1, 2, 3}, []int{1, 2}},
		{"getauscht", []int{1, 2, 3}, []int{1, 2, 4}}, {"ueberlappend", []int{1, 2}, []int{2, 3, 4}}, {"disjunkt", []int{1, 2}, []int{3, 4}},
		{"K1 leer", nil, []int{1, 2, 3}}, {"K2 leer", []int{1, 2, 3}, nil},
	}
}

// Linkformen mit dem Platzhalter {T} für das Token der Folgeseite: aufgelöste und nicht aufgelöste Formen aus den
// Klassen von H6; ob ein Link auflösbar ist und worauf er zeigt, sagt linkOrakel.
func h14Positiv() []string {
	return []string{
		"Products?$skiptoken={T}", "/odata/v4/Products?$skiptoken={T}", "https://up.example/odata/v4/Products?$skiptoken={T}", "./Products?$skiptoken={T}",
		"../v4/Products?$skiptoken={T}", "//up.example/odata/v4/Products?$skiptoken={T}#f", "HTTPS://UP.EXAMPLE:443/odata/v4/Products?$skiptoken={T}",
		"https://u:p@up.example:0443/odata/v4/Products?$skiptoken={T}", "/odata/v4/x/../Products?$skiptoken={T}", "https://up.example/odata/v4/./Products?$skiptoken={T}",
		"Prod%75cts?$skiptoken={T}",
	}
}

func h14Negativ() []string {
	return []string{
		"mailto:x@y", "https://other.example/odata/v4/Products?$skiptoken={T}", "/odata/v42/Products?$skiptoken={T}", "ftp://up.example/odata/v4/Products?$skiptoken={T}",
		"http://up.example/odata/v4/Products?$skiptoken={T}", "https://up.example:8443/odata/v4/Products?$skiptoken={T}", "https://up.example./odata/v4/Products?$skiptoken={T}",
		"https://up.example:+443/odata/v4/Products?$skiptoken={T}", "https://up.example:65536/odata/v4/Products?$skiptoken={T}", "Products?$skiptoken={T} ",
		"https://[::1]x/odata/v4/Products?$skiptoken={T}", "///odata/v4/Products?$skiptoken={T}",
	}
}

// sessionS baut eine Session des Generators S und hält, was das Orakel über sie wissen muss.
type sessionS struct {
	z         *rand.Rand
	w         *welt
	b         *bauer
	absicht   map[string]seiteS
	dim       map[string]bool
	sauber    bool // eine Menge, kein Verstoß
	belegt    bool // acht Beziehungen, kein Verstoß
	frei      bool // alles frei
	kopfModus int
	nKetten   int
	mengen    [][]int
	erstesK   int
}

func startQueriesS() []string { return []string{"", "$format=json", "$count=true"} }

// neueSessionS wählt die Art der Session, die Header, die Zahl der Ketten und ihre Schlüsselmengen.
func neueSessionS(t *testing.T, z *rand.Rand) *sessionS {
	w := neu(t)
	s := &sessionS{z: z, w: w, b: w.standard(), absicht: map[string]seiteS{}, dim: map[string]bool{}}
	art := z.IntN(100)
	s.sauber, s.belegt = art < 20, art >= 20 && art < 50
	s.frei = !s.sauber && !s.belegt
	s.kopfModus = z.IntN(3)
	s.dim[fmt.Sprint("header:", s.kopfModus)] = true
	s.nKetten = 1
	switch r := z.IntN(100); {
	case r >= 55:
		s.nKetten = 3
	case r >= 15:
		s.nKetten = 2
	}
	if s.frei && z.IntN(5) == 0 {
		s.nKetten = 0
	}
	rel := h14Beziehungen()[0]
	if !s.sauber {
		rel = h14Beziehungen()[z.IntN(len(h14Beziehungen()))]
	}
	s.mengen = make([][]int, s.nKetten)
	for c := range s.mengen {
		s.mengen[c] = rel.k1
		if c == 1 || c == 2 && z.IntN(2) == 1 {
			s.mengen[c] = rel.k2
		}
	}
	if s.nKetten >= 2 {
		s.dim["beziehung:"+rel.name] = true
	}
	s.dim[fmt.Sprint("ketten:", s.nKetten)] = true
	if s.nKetten == 0 {
		s.dim["seiten:0"] = true
	}
	s.erstesK = 1
	if s.nKetten > 0 && len(s.mengen[0]) > 0 {
		s.erstesK = s.mengen[0][0]
	}
	return s
}

// kopf sind die Header der Seite p der Kette c: keine, gleiche oder je Seite verschiedene.
func (s *sessionS) kopf(c, p int) map[string][]string {
	switch s.kopfModus {
	case 1:
		return map[string][]string{"Accept": {"a"}}
	case 2:
		return map[string][]string{"Accept": {fmt.Sprintf("p%d-%d", c, p)}}
	}
	return nil
}

// kettenPlanS ist die Wahl für eine Kette: Seiten, Verteilung der Schlüssel, Verstoß und Anzahl.
type kettenPlanS struct {
	c, n       int
	ids        [][]int
	art        string // "" oder die Art des Verstoßes
	pos        int    // Seite des Verstoßes (null-basiert)
	deklariert int64  // -1: keine Anzahl
}

func (p kettenPlanS) query(i int) string {
	if i == 0 {
		return startQueriesS()[p.c]
	}
	return fmt.Sprintf("$skiptoken=%d", 100*(p.c+1)+i)
}

// planeKette wählt Seitenzahl, Verteilung, Verstoß und `declared` der Kette c.
func (s *sessionS) planeKette(c int) kettenPlanS {
	menge := s.mengen[c]
	p := kettenPlanS{c: c, n: 1 + s.z.IntN(5), deklariert: -1}
	s.dim[fmt.Sprint("seiten:", p.n)] = true
	p.ids = make([][]int, p.n)
	for i, id := range menge {
		p.ids[i%p.n] = append(p.ids[i%p.n], id)
	}
	elemente := len(menge)
	if s.frei && p.n > 1 && len(menge) > 0 && s.z.IntN(4) == 0 {
		p.ids[p.n-1] = append(p.ids[p.n-1], menge[0]) // eine Dublette auf einer anderen Seite
		elemente++
	}
	if s.frei && s.z.IntN(100) < 85 {
		p.art, p.pos = s.waehleVerstoss(p.n)
	}
	for _, a := range []string{"404", "set", "entity"} {
		if p.art == "nicht_lesbar_"+a {
			s.dim["seitenart:"+map[string]string{"404": "404", "set": "anderes Set", "entity": "Entity"}[a]] = true
		}
	}
	p.deklariert = s.waehleDeklariert(int64(len(menge)), int64(elemente))
	return p
}

// waehleVerstoss wählt eine der sieben Arten und die Seite; für eine Art, die die Seitenzahl nicht trägt, keinen Verstoß.
func (s *sessionS) waehleVerstoss(n int) (string, int) {
	arten := []string{"nicht_lesbar_404", "nicht_lesbar_set", "nicht_lesbar_entity", "key_redacted", "link_redigiert", "schleife", "seite_fehlt"}
	art := arten[s.z.IntN(len(arten))]
	switch {
	case art == "key_redacted":
		return art, s.z.IntN(n)
	case n < 2:
		return "", 0
	case strings.HasPrefix(art, "nicht_lesbar"):
		return art, 1 + s.z.IntN(n-1)
	}
	return art, s.z.IntN(n - 1)
}

// waehleDeklariert wählt einen der zehn Werte von `declared` (d verschiedene Schlüssel, elemente Elemente).
func (s *sessionS) waehleDeklariert(d, elemente int64) int64 {
	mode := s.z.IntN(10)
	if !s.frei {
		mode = []int{0, 3, 3, 5}[s.z.IntN(4)]
	}
	werte := []int64{-1, 0, d - 1, d, d + 1, elemente, math.MaxInt32, math.MaxInt32 + 1, 1<<32 + d, math.MaxInt64}
	if mode == 2 && d == 0 {
		return -1
	}
	s.dim[fmt.Sprint("declared:", mode)] = true
	return werte[mode]
}

// baueKette legt die Exchanges der Kette an.
func (s *sessionS) baueKette(p kettenPlanS) {
	// Vorläufer mit den Keys der späteren Seiten, vor dem Beginn der Kette: sie sind nie die nächste Seite.
	for i := 1; i < p.n; i++ {
		if s.frei && s.z.IntN(12) == 0 {
			id := 9
			if s.z.IntN(2) == 0 {
				id = s.erstesK
			}
			e := s.b.collection("Products", p.query(i), seite(produkteVon([]int{id})...))
			s.absicht[e.ID] = seiteS{gelesen: true, lesbar: true, anzahl: -1, bestand: []int{id}, ids: []int{id}}
			s.dim["bestand:vorlaeufer"] = true
		}
	}
	for i := 0; i < p.n; i++ {
		s.baueSeite(p, i)
	}
}

// linkVorlage wählt die Form des Links der Seite i (nicht für einen redigierten Link).
func (s *sessionS) linkVorlage(p kettenPlanS, i int) string {
	switch {
	case p.art == "schleife" && i == p.pos:
		q := s.z.IntN(i + 1)
		switch {
		case q == 0 && startQueriesS()[p.c] != "":
			return "Products?" + startQueriesS()[p.c]
		case q > 0:
			return fmt.Sprintf("Products?$skiptoken=%d", 100*(p.c+1)+q)
		}
		return "Products"
	case p.art == "seite_fehlt" && i == p.pos:
		return "Products?$skiptoken=77{T}"
	case s.frei && s.z.IntN(100) < 18:
		return h14Negativ()[s.z.IntN(len(h14Negativ()))]
	}
	return h14Positiv()[s.z.IntN(len(h14Positiv()))]
}

// baueSeite legt die Seite i der Kette an und trägt ihre Absicht ein.
func (s *sessionS) baueSeite(p kettenPlanS, i int) {
	ids := p.ids[i]
	a := seiteS{gelesen: true, lesbar: true, start: i == 0, anzahl: -1, bestand: ids, ids: ids}
	sp := seite(produkteVon(ids)...)
	query := p.query(i)
	kp := s.kopf(p.c, i)
	if i < p.n-1 {
		a.link = true
		if p.art == "link_redigiert" && i == p.pos {
			sp, a.redigiert = redigierterLink(sp), true
		} else {
			link := strings.ReplaceAll(s.linkVorlage(p, i), "{T}", fmt.Sprint(100*(p.c+1)+i+1))
			sp = mitLink(sp, link)
			a.ziel, a.ok = linkOrakel(upstream, "/Products", query, kp, link)
			s.dim[fmt.Sprint("linkform:", a.ok)] = true
		}
	}
	if i == 0 && p.deklariert >= 0 {
		a.anzahl = p.deklariert
		sp = mitAnzahl(sp, p.deklariert)
	}
	verstoss := p.art != "" && i == p.pos
	a.keyRedigiert = p.art == "key_redacted" && verstoss
	erwartet := ""
	if i > 0 {
		erwartet = s.w.schluesselMit("/Products", query, s.kopf(p.c, i-1))
	}
	aend := func(e *model.Exchange) {
		e.RequestHeaders = kp
		if i > 0 {
			e.MatchKey = erwartet
		}
		e.KeyRedacted = e.KeyRedacted || a.keyRedigiert
	}
	var e model.Exchange
	switch {
	case verstoss && p.art == "nicht_lesbar_404":
		e, a = s.b.collection("Products", query, sp, aend, status(404)), seiteS{anzahl: -1}
	case verstoss && p.art == "nicht_lesbar_set":
		e, a = s.b.collection("Orders", query, sp, aend), seiteS{anzahl: -1}
	case verstoss && p.art == "nicht_lesbar_entity":
		id := 1
		if len(ids) > 0 {
			id = ids[0]
		}
		e, a = s.b.entity("Products", fmt.Sprint(id), produkteVon([]int{id})[0], aend), seiteS{gelesen: true, anzahl: -1, bestand: []int{id}}
	default:
		e = s.b.collection("Products", query, sp, aend)
	}
	s.absicht[e.ID] = a
}

// bestandsquelle legt, mit der Wahrscheinlichkeit der Art der Session, eine Einzel-Entity oder eine gefilterte Abfrage an,
// deren Schlüssel im Schlüsselbereich der ersten Kette liegt oder außerhalb.
func (s *sessionS) bestandsquelle() {
	anteil := 33
	switch {
	case s.nKetten == 0:
		anteil = 75
	case s.sauber:
		anteil = 50
	}
	if s.z.IntN(100) >= anteil {
		return
	}
	id := s.erstesK
	if s.z.IntN(100) < 60 {
		id = 99
		s.dim["bestand:aussen"] = true
	} else {
		s.dim["bestand:innen"] = true
	}
	if s.z.IntN(2) == 0 {
		e := s.b.entity("Products", fmt.Sprint(id), produkteVon([]int{id})[0])
		s.absicht[e.ID] = seiteS{gelesen: true, anzahl: -1, bestand: []int{id}}
		return
	}
	e := s.b.collection("Products", "$filter=Id%20eq%201", seite(produkteVon([]int{id})...))
	s.absicht[e.ID] = seiteS{gelesen: true, lesbar: true, anzahl: -1, bestand: []int{id}, ids: []int{id}}
}

// vergleiche importiert die Session und vergleicht Ketten, Enden, `count`, Status und Grund mit dem Orakel; es
// liefert das Urteil des Orakels.
func (s *sessionS) vergleiche(t *testing.T, sess int) sollS {
	t.Helper()
	mustImport(t, s.w, anfrage())
	st, grund, ketten := ketteVon(t, s.w)
	soll := orakelS(s.w.exchanges.m[sitzungID], s.absicht)
	if len(ketten) != len(soll.ketten) {
		t.Fatalf("Session %d: %d Ketten, Orakel %d", sess, len(ketten), len(soll.ketten))
	}
	for i, k := range ketten {
		o := soll.ketten[i]
		if got := endeVon(k); got != o.ende {
			t.Fatalf("Session %d Kette %d: Ende %s, Orakel %s", sess, i, got, o.ende)
		}
		s.dim["ende:"+o.ende] = true
		c, hat := k["count"].(map[string]any)
		if hat != (o.anzahl >= 0) {
			t.Fatalf("Session %d Kette %d: count %v, Orakel Anzahl %d", sess, i, k["count"], o.anzahl)
		}
		// declared jenseits von 2^53 verliert als JSON-Zahl in float64 Stellen; observed ist immer genau.
		if hat && (o.anzahl < 1<<53 && int64(c["declared"].(float64)) != o.anzahl || int(c["observed"].(float64)) != o.beobachtet) {
			t.Fatalf("Session %d Kette %d: count %v, Orakel %d von %d", sess, i, c, o.beobachtet, o.anzahl)
		}
	}
	if st != soll.status || grund != soll.grund {
		t.Fatalf("Session %d: %s/%s, Orakel %s/%s (%v)", sess, st, grund, soll.status, soll.grund, soll.ketten)
	}
	s.dim["status:"+soll.status+"/"+soll.grund] = true
	return soll
}

// h14Erwartet sind die Dimensionen, die jede in mindestens 300 Sessions vorkommen müssen.
func h14Erwartet() []string {
	erwartet := []string{"seiten:0", "seiten:1", "seiten:2", "seiten:3", "seiten:4", "seiten:5", "ketten:1", "ketten:2", "ketten:3", "header:0", "header:1", "header:2",
		"ende:proven", "ende:seite_nicht_lesbar", "ende:key_redacted", "ende:link_redigiert", "ende:link_nicht_aufloesbar", "ende:schleife", "ende:seite_fehlt", "ende:anzahl_abweichend",
		"seitenart:404", "seitenart:anderes Set", "seitenart:Entity", "linkform:true", "linkform:false", "bestand:innen", "bestand:aussen",
		"status:complete/kette_belegt", "status:partial/kette_nicht_belegt", "status:partial/ketten_abweichend", "status:partial/bestand_ausserhalb_kette", "status:partial/keine_ungefilterte_abfrage"}
	for _, r := range h14Beziehungen() {
		erwartet = append(erwartet, "beziehung:"+r.name)
	}
	for m := 0; m < 10; m++ {
		erwartet = append(erwartet, fmt.Sprint("declared:", m))
	}
	return erwartet
}

// TestZeileH14Allaussage: Generator S (5000 Sessions, fester Seed): complete genau dann, wenn ein Beweis
// besteht, in beiden Richtungen; jede Kette endet, wie das Orakel es bestimmt, und Status, Grund und
// `count` stimmen. Der Generator führt jede Dimension des Plans (Zahl der Seiten, Verstöße, Linkformen,
// Seiten eines anderen Sets, eine Entity als Seite, die acht Beziehungen der Schlüsselmengen, zehn Werte
// von `declared`, drei Formen der Header, eins bis drei Ketten, Quellen des Bestands); der Test zählt je
// Dimension die Sessions und verlangt mindestens 300.
func TestZeileH14Allaussage(t *testing.T) {
	zufall := rand.New(rand.NewPCG(42, 4711))
	dimensionen := map[string]int{}
	vollstaendig, unvollstaendig := 0, 0
	for sess := 0; sess < 5000; sess++ {
		s := neueSessionS(t, zufall)
		for c := 0; c < s.nKetten; c++ {
			s.baueKette(s.planeKette(c))
		}
		s.bestandsquelle()
		if s.vergleiche(t, sess).status == "complete" {
			vollstaendig++
		} else {
			unvollstaendig++
		}
		for k := range s.dim {
			dimensionen[k]++
		}
	}
	for _, k := range h14Erwartet() {
		if dimensionen[k] < 300 {
			t.Errorf("Dimension %s: %d Sessions, erwartet mindestens 300", k, dimensionen[k])
		}
	}
	if vollstaendig < 300 || unvollstaendig < 300 {
		t.Errorf("der Generator erreicht nur %d vollständige und %d unvollständige Sessions", vollstaendig, unvollstaendig)
	}
	var zeilen []string
	for _, k := range slices.Sorted(maps.Keys(dimensionen)) {
		zeilen = append(zeilen, fmt.Sprint(k, "=", dimensionen[k]))
	}
	t.Logf("H14: %d vollständige, %d unvollständige Sessions; Sessions je Dimension: %s", vollstaendig, unvollstaendig, strings.Join(zeilen, " "))
}
