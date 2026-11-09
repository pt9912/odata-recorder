package policy_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

const (
	zielPassend = "https://h"
	pfadRoute   = "/svc/Orders"
)

// bekannte sind die sieben Methoden, die die Matrix oder eine Routenfreigabe nennt; methoden13 ist
// die Menge M des Plans: dazu OPTIONS, TRACE, CONNECT, die Kleinschreibung, die leere Methode und
// PROPFIND.
func bekannte() []string {
	return []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE"}
}

func methoden13() []string {
	return append(bekannte(), "OPTIONS", "TRACE", "CONNECT", "get", "", "PROPFIND")
}

func lesend(m string) bool { return m == "GET" || m == "HEAD" }

func mutierend(m string) bool {
	return slices.Contains([]string{"POST", "PUT", "PATCH", "DELETE", "MERGE"}, m)
}

func neu(t *testing.T, w policy.Werte) *policy.Policy {
	t.Helper()
	p, err := policy.New(w)
	if err != nil {
		t.Fatalf("New(%+v): %v", w, err)
	}
	return p
}

func name(e policy.Entscheidung) string {
	switch e {
	case policy.Bleiben:
		return "Bleiben"
	case policy.Replay:
		return "Replay"
	case policy.Upstream:
		return "Upstream"
	}
	return "?"
}

// codesC ist die Menge C des Plans: jeder Code aus §4 (N = 41), dazu der Nullwert, -1 und N + 1.
func codesC(t *testing.T) []model.Code {
	t.Helper()
	alle := model.AlleCodes()
	if len(alle) != 41 {
		t.Fatalf("model.AlleCodes() liefert %d Codes, erwartet 41", len(alle))
	}
	c := append([]model.Code{0, -1, model.Code(len(alle) + 1)}, alle...)
	if len(c) != len(alle)+3 {
		t.Fatalf("Menge C hat %d Werte", len(c))
	}
	return c
}

const (
	e6001 = model.CodeE6001
	e6002 = model.CodeE6002
	e6003 = model.CodeE6003
	e5001 = model.CodeE5001
)

func mock3() []model.Code {
	return []model.Code{model.CodeE6001, model.CodeE6002, model.CodeE6003}
}

// kategorieDes ist die Kategorie, die ein Mock-Auslöser nach LH-FA-12.b trägt.
func kategorieDes(c model.Code) string {
	if c == model.CodeE6003 {
		return policy.KategorieCoverage
	}
	return policy.KategorieUnsupported
}

const (
	kategorien2 = "beide"
	kategorienU = "unsupported"
	kategorienC = "coverage"
	kategorien0 = "keine"
)

func kategorienListe(k string) []string {
	switch k {
	case kategorien2:
		return []string{policy.KategorieUnsupported, policy.KategorieCoverage}
	case kategorienU:
		return []string{policy.KategorieUnsupported}
	case kategorienC:
		return []string{policy.KategorieCoverage}
	}
	return nil
}

// kategorienHat sagt, ob die Kategorien-Auswahl k die Kategorie kat enthält.
func kategorienHat(k, kat string) bool { return slices.Contains(kategorienListe(k), kat) }

func ent(p *policy.Policy, m, pfad, ziel string, c model.Code) policy.Entscheidung {
	return p.Entscheide(policy.Anfrage{Methode: m, Pfad: pfad, Ziel: ziel, Code: c})
}

// TestAnfrageFelder (P04): Anfrage trägt genau Methode, Pfad, Ziel und Code, keinen Status und
// keinen Header.
func TestAnfrageFelder(t *testing.T) {
	typ := reflect.TypeOf(policy.Anfrage{})
	want := []struct{ name, typ string }{
		{"Methode", "string"}, {"Pfad", "string"}, {"Ziel", "string"},
		{"Code", "github.com/pt9912/odata-recorder/src/hexagon/model.Code"},
	}
	if typ.NumField() != len(want) {
		t.Fatalf("Anfrage hat %d Felder, erwartet %d", typ.NumField(), len(want))
	}
	for i, w := range want {
		f := typ.Field(i)
		got := f.Type.String()
		if f.Type.PkgPath() != "" {
			got = f.Type.PkgPath() + "." + f.Type.Name()
		}
		if f.Name != w.name || got != w.typ {
			t.Errorf("Feld %d: %s %s, erwartet %s %s", i, f.Name, got, w.name, w.typ)
		}
	}
}

// TestNullwerte (P22, Teil): der Nullwert von Entscheidung ist Bleiben, die nil-Policy und
// Policy{} antworten Bleiben auf jede Anfrage der Mengen M und C.
func TestNullwerte(t *testing.T) {
	var zero policy.Entscheidung
	if zero != policy.Bleiben {
		t.Fatalf("Nullwert %s", name(zero))
	}
	var nilPolicy *policy.Policy
	leer := &policy.Policy{}
	faelle := 0
	for _, p := range []*policy.Policy{nilPolicy, leer} {
		for _, m := range methoden13() {
			for _, c := range codesC(t) {
				faelle++
				if e := ent(p, m, pfadRoute, zielPassend, c); e != policy.Bleiben {
					t.Errorf("%v %q %d: %s", p, m, c, name(e))
				}
			}
		}
	}
	if faelle != 2*13*44 {
		t.Fatalf("%d Fälle", faelle)
	}
}

// voll baut das Werte-Paket "voll offen" des Plans: Upstream an, Host passend, je mutierender
// Methode ein Eintrag mit Muster /svc/Orders, die gewählten Kategorien.
func voll(kat string) policy.Werte {
	return policy.Werte{
		Upstream:         true,
		Hosts:            []string{"h"},
		Schreibfreigaben: []string{"POST " + pfadRoute, "PUT " + pfadRoute, "PATCH " + pfadRoute, "DELETE " + pfadRoute, "MERGE " + pfadRoute},
		ReplayKategorien: kategorienListe(kat),
	}
}

// TestMethodenKlassen (P02): 13 Methoden x 4 Codes im Zustand voll offen. Die sechs Methoden
// außerhalb der Klassen ergeben stets Bleiben; die übrigen nach der Tabelle von Hand.
func TestMethodenKlassen(t *testing.T) {
	p := neu(t, voll(kategorien2))
	codes := []model.Code{e6001, e6002, e6003, e5001}
	faelle := 0
	for _, m := range methoden13() {
		for _, c := range codes {
			faelle++
			want := policy.Bleiben
			switch {
			case !lesend(m) && !mutierend(m):
			case c != e5001:
				want = policy.Replay
			case lesend(m):
				want = policy.Upstream
			}
			if e := ent(p, m, pfadRoute, zielPassend, c); e != want {
				t.Errorf("%q %d: %s, erwartet %s", m, c, name(e), name(want))
			}
		}
	}
	if faelle != 13*4 {
		t.Fatalf("%d Fälle, erwartet 52", faelle)
	}
	// die Klassen einzeln, von Hand: Upstream ohne Replay-Kategorie
	q := neu(t, voll(kategorien0))
	hand := map[string]policy.Entscheidung{
		"GET": policy.Upstream, "HEAD": policy.Upstream,
		"POST": policy.Upstream, "PUT": policy.Upstream, "PATCH": policy.Upstream, "DELETE": policy.Upstream, "MERGE": policy.Upstream,
		"OPTIONS": policy.Bleiben, "TRACE": policy.Bleiben, "CONNECT": policy.Bleiben,
		"get": policy.Bleiben, "": policy.Bleiben, "PROPFIND": policy.Bleiben,
	}
	if len(hand) != 13 {
		t.Fatalf("Tabelle von Hand: %d Methoden", len(hand))
	}
	for m, want := range hand {
		if e := ent(q, m, pfadRoute, zielPassend, e6001); e != want {
			t.Errorf("ohne Kategorie %q: %s, erwartet %s", m, name(e), name(want))
		}
	}
	// die Schreibweise zählt: Klein- und Titelschreibung jeder bekannten Methode (14 Formen) ist
	// nie freigegeben, gegen jeden der vier Codes
	schreibweise := 0
	for _, m := range bekannte() {
		for _, v := range []string{strings.ToLower(m), m[:1] + strings.ToLower(m[1:])} {
			for _, c := range codes {
				schreibweise++
				if e := ent(p, v, pfadRoute, zielPassend, c); e != policy.Bleiben {
					t.Errorf("Schreibweise %q, Code %d: %s", v, c, name(e))
				}
			}
		}
	}
	if schreibweise != 7*2*4 {
		t.Fatalf("%d Fälle der Schreibweise, erwartet 56", schreibweise)
	}
}

// TestAusloeser (P03): Schleife über C mit GET im Zustand voll offen. Mit beiden Kategorien ergeben
// genau die drei Mock-Auslöser Replay, ohne Kategorie genau die vier Auslöser Upstream, alles
// andere Bleiben.
func TestAusloeser(t *testing.T) {
	mitKat := neu(t, voll(kategorien2))
	ohneKat := neu(t, voll(kategorien0))
	codes := codesC(t)
	replay, upstream := 0, 0
	for _, c := range codes {
		mock := c == e6001 || c == e6002 || c == e6003
		ausl := mock || c == e5001
		wantMit, wantOhne := policy.Bleiben, policy.Bleiben
		if mock {
			wantMit = policy.Replay
		}
		if ausl {
			wantOhne = policy.Upstream
		}
		if c == e5001 {
			wantMit = policy.Upstream
		}
		if e := ent(mitKat, "GET", pfadRoute, zielPassend, c); e != wantMit {
			t.Errorf("mit Kategorien, Code %d: %s, erwartet %s", c, name(e), name(wantMit))
		}
		if e := ent(ohneKat, "GET", pfadRoute, zielPassend, c); e != wantOhne {
			t.Errorf("ohne Kategorien, Code %d: %s, erwartet %s", c, name(e), name(wantOhne))
		}
		if wantMit == policy.Replay {
			replay++
		}
		if wantOhne == policy.Upstream {
			upstream++
		}
	}
	if len(codes) != 44 || replay != 3 || upstream != 4 {
		t.Fatalf("Menge: %d Codes, %d Replay, %d Upstream; erwartet 44, 3, 4", len(codes), replay, upstream)
	}
}

// zustand ist ein Element der Menge Z des Plans.
type zustand struct {
	upstream bool
	host     string // leer, passend, fremd, port
	frei     string // keine, passend, andere-methode, anderes-muster
	kat      string // keine, unsupported, coverage, beide
}

// zustaendeZ liefert die 112 Zustände: Upstream aus mit 4 Hosts, Upstream an mit den 3 nicht leeren.
func zustaendeZ(t *testing.T) []zustand {
	t.Helper()
	var z []zustand
	for _, up := range []bool{false, true} {
		hosts := []string{"leer", "passend", "fremd", "port"}
		if up {
			hosts = hosts[1:]
		}
		for _, h := range hosts {
			for _, f := range []string{"keine", "passend", "andere-methode", "anderes-muster"} {
				for _, k := range []string{kategorien0, kategorienU, kategorienC, kategorien2} {
					z = append(z, zustand{up, h, f, k})
				}
			}
		}
	}
	if len(z) != 112 {
		t.Fatalf("Menge Z hat %d Zustände", len(z))
	}
	return z
}

// werteFuer baut die Werte eines Zustands für die Methode m: die Freigabe "passend" trägt die
// mutierende Methode m (bei einer anderen Methode POST), "andere-methode" eine andere mutierende.
func werteFuer(z zustand, m string) policy.Werte {
	w := policy.Werte{Upstream: z.upstream, ReplayKategorien: kategorienListe(z.kat)}
	switch z.host {
	case "passend":
		w.Hosts = []string{"h"}
	case "fremd":
		w.Hosts = []string{"x"}
	case "port":
		w.Hosts = []string{"h:8443"}
	}
	passend := "POST"
	if mutierend(m) {
		passend = m
	}
	andere := "DELETE"
	if passend == "DELETE" {
		andere = "PUT"
	}
	switch z.frei {
	case "passend":
		w.Schreibfreigaben = []string{passend + " " + pfadRoute}
	case "andere-methode":
		w.Schreibfreigaben = []string{andere + " " + pfadRoute}
	case "anderes-muster":
		w.Schreibfreigaben = []string{passend + " /svc/Other"}
	}
	return w
}

// matrixZeile ist eine Zeile von LH-FA-12.b, soweit sie einen Fallback zulässt: die Codes, die
// Kategorie der Spalte "Optionaler Replay-Fallback" (leer: kein Replay) und der Wortlaut der
// Spalte "Upstream-Fallback".
type matrixZeile struct {
	zeile    string
	codes    []model.Code
	replay   string
	upstream string
}

const (
	upstreamLesendUndRoute = "GET/HEAD + Allowlist; mutierend nur mit FALLBACK_WRITE_ALLOW"
	upstreamNurLesend      = "Nur GET/HEAD und nur für Hosts aus FALLBACK_ALLOW_HOSTS"
)

// matrix: jede Zeile der Tabelle, die in einer Spalte "Nein" nicht nur "Nein" trägt. Erfolg,
// fachliche 404/409/412, ungültige Anfrage und jeder andere Code stehen nicht hier: beide Spalten
// "Nein".
func matrix() []matrixZeile {
	return []matrixZeile{
		{"Coverage unvollständig", []model.Code{model.CodeE6003}, policy.KategorieCoverage, upstreamLesendUndRoute},
		{"Nicht unterstützte Operation", []model.Code{model.CodeE6001, model.CodeE6002}, policy.KategorieUnsupported, upstreamLesendUndRoute},
		{"Replay-Miss", []model.Code{model.CodeE5001}, "", upstreamNurLesend},
	}
}

// orakel liest die Entscheidung aus der Tabelle: erst die Spalte Replay (LH-FA-12.a: Mock, dann
// Replay), dann die Spalte Upstream; ohne Zeile Bleiben.
func orakel(c model.Code, m string, z zustand) policy.Entscheidung {
	var zeile *matrixZeile
	zeilen := matrix()
	for i := range zeilen {
		if slices.Contains(zeilen[i].codes, c) {
			zeile = &zeilen[i]
		}
	}
	if zeile == nil || !slices.Contains(bekannte(), m) {
		return policy.Bleiben
	}
	if zeile.replay != "" && kategorienHat(z.kat, zeile.replay) {
		return policy.Replay
	}
	if !z.upstream || z.host != "passend" {
		return policy.Bleiben
	}
	switch {
	case lesend(m):
		return policy.Upstream
	case zeile.upstream == upstreamLesendUndRoute && z.frei == "passend":
		return policy.Upstream
	}
	return policy.Bleiben
}

// TestMatrix (P06): C x M x Z = 44 x 13 x 112 Fälle gegen die Tabelle von LH-FA-12.b. Die Zählung
// der Ergebnisse steht als Zahl von Hand (1176 Replay, 110 Upstream): nur ODR-E6001, ODR-E6002, ODR-E6003 und ODR-E5001 sind Auslöser, jeder
// andere Code, auch ODR-E6010 (die fachliche 412 des Mock), liefert Bleiben, so dass der 44. Wert von C keine Zahl der Ergebnisse ändert.
func TestMatrix(t *testing.T) {
	codes := codesC(t)
	zust := zustaendeZ(t)
	faelle, replay, upstream, orakelReplay, orakelUpstream := 0, 0, 0, 0, 0
	for _, m := range methoden13() {
		for _, z := range zust {
			p := neu(t, werteFuer(z, m))
			for _, c := range codes {
				faelle++
				got := ent(p, m, pfadRoute, zielPassend, c)
				want := orakel(c, m, z)
				if got != want {
					t.Errorf("Methode %q, Code %d, Zustand %+v: %s, erwartet %s", m, c, z, name(got), name(want))
				}
				replay += b2i(got == policy.Replay)
				upstream += b2i(got == policy.Upstream)
				orakelReplay += b2i(want == policy.Replay)
				orakelUpstream += b2i(want == policy.Upstream)
			}
		}
	}
	if faelle != 64064 {
		t.Fatalf("%d Fälle, erwartet 64064", faelle)
	}
	if orakelReplay != 1176 || orakelUpstream != 110 {
		t.Fatalf("Orakel zählt %d Replay und %d Upstream, erwartet 1176 und 110", orakelReplay, orakelUpstream)
	}
	if replay != 1176 || upstream != 110 {
		t.Errorf("Code liefert %d Replay und %d Upstream, erwartet 1176 und 110", replay, upstream)
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestReplayFallback (P05): 4 Kategorien x 3 Mock-Auslöser x 7 Methoden x Upstream {aus, an} x Ziel
// {passend, leer} = 336 Fälle: Replay genau bei gesetzter Kategorie des Codes. Dazu 28 Fälle mit
// ODR-E5001: nie Replay.
func TestReplayFallback(t *testing.T) {
	faelle, replay := 0, 0
	for _, k := range []string{kategorien0, kategorienU, kategorienC, kategorien2} {
		for _, up := range []bool{false, true} {
			w := policy.Werte{Upstream: up, ReplayKategorien: kategorienListe(k)}
			if up {
				w.Hosts = []string{"h"}
			}
			p := neu(t, w)
			for _, ziel := range []string{zielPassend, ""} {
				for _, c := range mock3() {
					for _, m := range bekannte() {
						faelle++
						want := kategorienHat(k, kategorieDes(c))
						got := ent(p, m, pfadRoute, ziel, c) == policy.Replay
						if got != want {
							t.Errorf("Kategorien %s, Upstream %v, Ziel %q, Code %d, %s: Replay %v, erwartet %v", k, up, ziel, c, m, got, want)
						}
						replay += b2i(got)
					}
				}
			}
		}
	}
	if faelle != 336 || replay != 7*3*2*2*2 {
		t.Fatalf("%d Fälle, %d Replay; erwartet 336 und 168", faelle, replay)
	}
	miss := 0
	for _, k := range []string{kategorien0, kategorienU, kategorienC, kategorien2} {
		p := neu(t, policy.Werte{ReplayKategorien: kategorienListe(k)})
		for _, m := range bekannte() {
			miss++
			if e := ent(p, m, pfadRoute, zielPassend, e5001); e == policy.Replay {
				t.Errorf("ODR-E5001, Kategorien %s, %s: Replay", k, m)
			}
		}
	}
	if miss != 28 {
		t.Fatalf("%d Fälle ODR-E5001, erwartet 28", miss)
	}
}

// TestRoutenfreigabeMethodeUndMuster (P07): 5 mutierende Methoden x 9 Freigaben x 3 Mock-Codes =
// 135 Fälle; Upstream genau bei der Freigabe, die Methode und Pfad deckt.
func TestRoutenfreigabeMethodeUndMuster(t *testing.T) {
	faelle, upstream := 0, 0
	for _, m := range []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"} {
		var freigaben [][]string
		freigaben = append(freigaben, nil, []string{m + " " + pfadRoute})
		for _, a := range []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"} {
			if a != m {
				freigaben = append(freigaben, []string{a + " " + pfadRoute})
			}
		}
		freigaben = append(freigaben, []string{m + " /svc/Other"}, []string{m + " /svc/Orders/x"}, []string{m + " /svc"})
		if len(freigaben) != 9 {
			t.Fatalf("%d Freigaben", len(freigaben))
		}
		for i, f := range freigaben {
			p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: f})
			for _, c := range mock3() {
				faelle++
				want := policy.Bleiben
				if i == 1 {
					want = policy.Upstream
					upstream++
				}
				if e := ent(p, m, pfadRoute, zielPassend, c); e != want {
					t.Errorf("%s, Freigabe %v, Code %d: %s, erwartet %s", m, f, c, name(e), name(want))
				}
			}
		}
	}
	if faelle != 135 || upstream != 15 {
		t.Fatalf("%d Fälle, %d Upstream; erwartet 135 und 15", faelle, upstream)
	}
}

// TestReplayMiss (P08): ODR-E5001 erreicht den Upstream nur mit GET und HEAD, auch wenn für jede
// mutierende Methode ein passender Eintrag steht.
func TestReplayMiss(t *testing.T) {
	p := neu(t, voll(kategorien0))
	faelle := 0
	for _, m := range bekannte() {
		faelle++
		want := policy.Bleiben
		if lesend(m) {
			want = policy.Upstream
		}
		if e := ent(p, m, pfadRoute, zielPassend, e5001); e != want {
			t.Errorf("%s: %s, erwartet %s", m, name(e), name(want))
		}
	}
	if faelle != 7 {
		t.Fatalf("%d Fälle", faelle)
	}
}

// TestReplayVorUpstream (P09): bei offenem Replay und Upstream gilt der Replay; ODR-E5001 führt zu
// Upstream. 3 Mock-Auslöser x {Kategorie gesetzt, Kategorie leer, ODR-E5001 im Zustand mit Kategorie}.
func TestReplayVorUpstream(t *testing.T) {
	faelle := 0
	for _, c := range mock3() {
		mit := neu(t, voll(kategorien2))
		ohne := neu(t, voll(kategorien0))
		faelle += 3
		if e := ent(mit, "GET", pfadRoute, zielPassend, c); e != policy.Replay {
			t.Errorf("Code %d mit Kategorie: %s", c, name(e))
		}
		if e := ent(ohne, "GET", pfadRoute, zielPassend, c); e != policy.Upstream {
			t.Errorf("Code %d ohne Kategorie: %s", c, name(e))
		}
		if e := ent(mit, "GET", pfadRoute, zielPassend, e5001); e != policy.Upstream {
			t.Errorf("ODR-E5001 im Zustand mit Kategorie: %s", name(e))
		}
	}
	if faelle != 9 {
		t.Fatalf("%d Fälle", faelle)
	}
}

// TestHauptschalterAus (P10): bei Upstream aus erreicht keine Anfrage den Upstream, gleich welche
// Hosts und Schreibfreigaben stehen (168 Fälle); die Gegenprobe mit Upstream an liefert für GET und
// HEAD Upstream (8 Fälle).
func TestHauptschalterAus(t *testing.T) {
	faelle := 0
	for _, h := range [][]string{nil, {"h"}, {"x"}} {
		for _, frei := range []bool{false, true} {
			for _, m := range bekannte() {
				w := policy.Werte{Hosts: h}
				if frei && mutierend(m) {
					w.Schreibfreigaben = []string{m + " " + pfadRoute}
				}
				if frei && !mutierend(m) {
					w.Schreibfreigaben = []string{"POST " + pfadRoute}
				}
				p := neu(t, w)
				for _, c := range []model.Code{e6001, e6002, e6003, e5001} {
					faelle++
					if e := ent(p, m, pfadRoute, zielPassend, c); e != policy.Bleiben {
						t.Errorf("Hosts %v, Freigabe %v, %s, Code %d: %s", h, frei, m, c, name(e))
					}
				}
			}
		}
	}
	if faelle != 168 {
		t.Fatalf("%d Fälle, erwartet 168", faelle)
	}
	gegen := 0
	p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}})
	for _, m := range []string{"GET", "HEAD"} {
		for _, c := range []model.Code{e6001, e6002, e6003, e5001} {
			gegen++
			if e := ent(p, m, pfadRoute, zielPassend, c); e != policy.Upstream {
				t.Errorf("Gegenprobe %s, Code %d: %s", m, c, name(e))
			}
		}
	}
	if gegen != 8 {
		t.Fatalf("%d Gegenproben", gegen)
	}
}
