package baselineimport_test

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type probeErgebnis struct {
	grund   string
	gelesen bool
	art     string
}

// probe lässt eine Session aus einem gültigen $metadata und dem einen Exchange GET pfad?query (mit
// den Änderungen) importieren und liefert den Ausgang des Exchange: gelesen mit der Art des Aufrufs
// des Codecs oder übersprungen mit dem Grund. Ein übersprungener Exchange wird nie geöffnet.
func probe(t *testing.T, pfad, query string, as ...aenderung) probeErgebnis {
	t.Helper()
	w := neu(t)
	b := w.standard()
	e := b.roh(pfad, query, as...)
	b.vorgabe(e, datenVorgabe{seite: seite(produkt(1, "A", "1")), entity: produkt(1, "A", "1")})
	r, err := w.importiere(anfrage())
	if err != nil {
		t.Errorf("%s?%s: %v", pfad, query, err)
		return probeErgebnis{grund: "FEHLER"}
	}
	sk := uebersprungen(t, r)
	if len(sk) > 1 {
		t.Errorf("%s?%s: mehrere Gründe %v", pfad, query, sk)
	}
	coll, entity := w.p.anzahl("ParseCollection"), w.p.anzahl("ParseEntity")
	if len(sk) == 0 {
		if w.p.anzahl("Blobs.Open") != 2 || coll+entity != 1 {
			t.Errorf("%s?%s: gelesen, aber Open %d, Parse %d", pfad, query, w.p.anzahl("Blobs.Open"), coll+entity)
		}
		art := "collection"
		if entity == 1 {
			art = "entity"
		}
		return probeErgebnis{gelesen: true, art: art}
	}
	if w.p.anzahl("Blobs.Open") != 1 || coll+entity != 0 {
		t.Errorf("%s?%s: übersprungen, aber Open %d, Parse %d", pfad, query, w.p.anzahl("Blobs.Open"), coll+entity)
	}
	for g := range sk {
		return probeErgebnis{grund: g}
	}
	return probeErgebnis{}
}

func ct(v ...string) aenderung {
	return func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"Content-Type": v} }
}

// TestZeileE1Spur: die Spur nach dem Pfad (Handmenge 36 und Generator 120).
func TestZeileE1Spur(t *testing.T) {
	coll := []string{"/Products", "/Orders"}
	ent := []string{"/Products(1)", "/Products('a')", "/Products(Id=1)", "/Orders(OrderID=1,LineNo=2)", "/Products('a)b')", "/Products('a/b')", "/Products('x''y')", "/Products('')"}
	pfad := []string{
		"/Products(1)/Category", "/Products('a')/Orders('b')", "/Products(1)/Name", "/Products(1)/Name/$value", "/Products/$count",
		"/Products/$ref", "/Products(1)/$ref", "/Products/Ns.Special", "/Products(1)/Ns.Special", "/Products/", "/products", "/PRODUCTS",
		"/Prod%75cts", "/Unknown", "/", "/$batch", "/Products()", "/Products(1", "/Products1)", "/(1)", "//Products", "/Products(1)(2)",
		"/Products(1)/", "/$metadata/x", "/Me",
	}
	if len(coll)+len(ent)+len(pfad)+1 != 36 {
		t.Fatalf("Handmenge %d, erwartet 36", len(coll)+len(ent)+len(pfad)+1)
	}
	for _, p := range coll {
		if g := probe(t, p, ""); !g.gelesen || g.art != "collection" {
			t.Errorf("%s: %+v", p, g)
		}
	}
	for _, p := range ent {
		if g := probe(t, p, ""); !g.gelesen || g.art != "entity" {
			t.Errorf("%s: %+v", p, g)
		}
	}
	// Ergänzung außerhalb der Handmenge: Namen, die einem Set nur ähneln (Präfix, Anhang, Navigation vor der Klammer).
	ergaenzung := []string{"/ProductsX", "/Product", "/Prod(1)", "/ProductsX(1)", "/Orde(1)", "/Products/Ns.Special(1)", "/Products/Orders('b')", "/Products(1)x", "/Products(1)(", "/Products(1))"}
	for _, p := range slices.Concat(pfad, ergaenzung) {
		if g := probe(t, p, ""); g.grund != "pfad" {
			t.Errorf("%s: %+v", p, g)
		}
	}
	n := 0
	for _, set := range []string{"Products", "Orders", "Unknown", "products"} {
		for _, praed := range []string{"", "(1)", "('a')", "(", "()", "(1)(2)"} {
			for _, suffix := range []string{"", "/", "/x", "/$count", "/$ref"} {
				n++
				got := probe(t, "/"+set+praed+suffix, "")
				// Orakel: eine zweite, von der Implementierung unabhängige Regelfunktion.
				bekannt := set == "Products" || set == "Orders"
				want := probeErgebnis{grund: "pfad"}
				switch {
				case bekannt && praed == "" && suffix == "":
					want = probeErgebnis{gelesen: true, art: "collection"}
				case bekannt && (praed == "(1)" || praed == "('a')") && suffix == "":
					want = probeErgebnis{gelesen: true, art: "entity"}
				}
				if got != want {
					t.Errorf("%s%s%s: %+v, erwartet %+v", set, praed, suffix, got, want)
				}
			}
		}
	}
	if n != 120 {
		t.Errorf("Generator %d, erwartet 120", n)
	}
}

type spurFall struct {
	name string
	pfad string
}

func spuren() []spurFall {
	return []spurFall{{"Collection", "/Products"}, {"Entity", "/Products(1)"}}
}

// TestZeileE2MethodeStatusCaptureBody: Menge 1026 und vier Gutfälle.
func TestZeileE2MethodeStatusCaptureBody(t *testing.T) {
	n := 0
	for _, s := range spuren() {
		var faelle []struct {
			name, grund string
			a           aenderung
		}
		add := func(name, grund string, a aenderung) {
			faelle = append(faelle, struct {
				name, grund string
				a           aenderung
			}{name, grund, a})
		}
		for _, m := range []string{"POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "get", "Get", ""} {
			add("Methode "+m, "methode", func(e *model.Exchange) { e.Method = m })
		}
		for st := 0; st <= 599; st++ {
			if st == 200 || (st > 0 && st < 100) {
				continue
			}
			add(fmt.Sprint("Status ", st), "status", func(e *model.Exchange) { e.ResponseStatus = st })
		}
		for _, c := range []model.CaptureStatus{model.CaptureIncomplete, model.CaptureFailed} {
			add("capture "+string(c), "capture", func(e *model.Exchange) { e.Capture = c })
		}
		for _, b := range []model.BodyState{model.BodyAbsent, model.BodyDiscarded} {
			add("Body "+string(b), "body", func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = b, "" })
		}
		for _, f := range faelle {
			n++
			if g := probe(t, s.pfad, "", f.a); g.grund != f.grund {
				t.Errorf("%s %s: %+v, erwartet %s", s.name, f.name, g, f.grund)
			}
		}
		for _, r := range []model.RedactionStatus{model.RedactionNone, model.RedactionApplied} {
			art := strings.ToLower(s.name)
			if g := probe(t, s.pfad, "", func(e *model.Exchange) { e.ResponseRedact = r }); !g.gelesen || g.art != art {
				t.Errorf("%s redaction %s: %+v", s.name, r, g)
			}
		}
	}
	if n != 1026 {
		t.Errorf("Menge %d, erwartet 1026", n)
	}
}

// TestZeileE3ContentType: 20 Formen mal beide Spuren (Menge 40).
func TestZeileE3ContentType(t *testing.T) {
	type form struct {
		name   string
		a      aenderung
		erwart string // "" gelesen
	}
	json := "application/json"
	formen := []form{
		{"json", ct(json), ""}, {"JSON groß", ct("APPLICATION/JSON"), ""}, {"minimal", ct(json + ";odata.metadata=minimal"), ""},
		{"full", ct(json + ";odata.metadata=full"), ""}, {"charset utf-8", ct(json + ";charset=utf-8"), ""},
		{"charset UTF-8", ct(json + ";charset=UTF-8"), ""}, {"IEEE754", ct(json + ";IEEE754Compatible=true"), ""},
		{"Name klein", func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"content-type": {json}} }, ""},
		{"none", ct(json + ";odata.metadata=none"), "ohne_kontext"}, {"NONE groß", ct(json + "; ODATA.METADATA=NONE"), "ohne_kontext"},
		{"none in Anführungszeichen", ct(json + `;odata.metadata="none"`), "ohne_kontext"},
		{"Leerzeichen vor dem Semikolon", ct(json + " ;odata.metadata=none"), "ohne_kontext"},
		{"fehlt", func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{} }, "kein_json"},
		{"zwei Zeilen", ct(json, json), "kein_json"}, {"iso-8859-1", ct(json + ";charset=iso-8859-1"), "kein_json"},
		{"atom", ct("application/atom+xml"), "kein_json"}, {"text/plain", ct("text/plain"), "kein_json"},
		{"jsonx", ct("application/jsonx"), "kein_json"}, {"leer", ct(""), "kein_json"},
		{"minimal und iso", ct(json + ";odata.metadata=minimal;charset=iso-8859-1"), "kein_json"},
	}
	if len(formen) != 20 {
		t.Fatalf("Formen %d", len(formen))
	}
	for _, s := range spuren() {
		for _, f := range formen {
			g := probe(t, s.pfad, "", f.a)
			if (f.erwart == "" && !g.gelesen) || g.grund != f.erwart {
				t.Errorf("%s %s: %+v, erwartet %q", s.name, f.name, g, f.erwart)
			}
		}
		// Zwei Schreibweisen des Namens zählen zusammen als zwei Werte.
		g := probe(t, s.pfad, "", func(e *model.Exchange) {
			e.ResponseHeaders = map[string][]string{"Content-Type": {json}, "content-type": {json}}
		})
		if g.grund != "kein_json" {
			t.Errorf("%s zwei Schreibweisen: %+v", s.name, g)
		}
	}
}

// TestZeileE4Optionen: 29 Namen mal drei Lagen mal beide Spuren (Menge 174).
func TestZeileE4Optionen(t *testing.T) {
	lesbar := []string{"$format=json", "$count=true", "$select=Id", "$orderby=Id", "$filter=Id%20eq%201", "$top=1", "$skip=1", "$skiptoken=a", "$search=x", "$expand=Category",
		"sap-client=100", "@a=1", "x", "=v", "%zz=1"}
	query := []string{"$apply=x", "$compute=x", "$deltatoken=x", "$levels=1", "$schemaversion=1", "$index=1", "$batch=1", "$Top=1", "$TOP=1", "$Select=Id", "$FORMAT=json",
		"$unknown=1", "%24apply=x", "$apply%zz=1"}
	if len(lesbar)+len(query) != 29 {
		t.Fatalf("Namen %d", len(lesbar)+len(query))
	}
	n := 0
	for _, s := range spuren() {
		for _, lage := range []string{"allein", "hinter", "vor"} {
			for i, o := range append(slices.Clone(lesbar), query...) {
				n++
				q := map[string]string{"allein": o, "hinter": "$top=1&" + o, "vor": o + "&$top=1"}[lage]
				g := probe(t, s.pfad, q)
				if i < len(lesbar) {
					if !g.gelesen {
						t.Errorf("%s %s %q: %+v, erwartet gelesen", s.name, lage, q, g)
					}
				} else if g.grund != "query" {
					t.Errorf("%s %s %q: %+v, erwartet query", s.name, lage, q, g)
				}
			}
		}
	}
	if n != 174 {
		t.Errorf("Menge %d, erwartet 174", n)
	}
}

// TestZeileE5SelectOhneSchluessel: 21 Formen mal beide Spuren (Menge 42).
func TestZeileE5SelectOhneSchluessel(t *testing.T) {
	type form struct {
		query   string
		gelesen bool
	}
	einfach := []form{{"", true}, {"$select=Id", true}, {"$select=Id,Name", true}, {"$select=Name,Id", true}, {"$select=*", true}, {"$select=%49d", true}, {"$select=Id,", true},
		{"$select=Name", false}, {"$select=", false}, {"$select=id", false}, {"$select=%20Id", false}, {"$select=Id%20", false}, {"$select=Ns.Product%2FId", false},
		{"$select=Name&$select=Id", true}, {"$select=Name&$select=Name2", false}}
	zusammen := []form{{"$select=OrderID,LineNo", true}, {"$select=LineNo,OrderID", true}, {"$select=*", true},
		{"$select=OrderID", false}, {"$select=LineNo", false}, {"$select=OrderID,Qty", false}}
	if len(einfach)+len(zusammen) != 21 {
		t.Fatalf("Formen %d", len(einfach)+len(zusammen))
	}
	n := 0
	pruefe := func(pfad string, f form) {
		n++
		g := probe(t, pfad, f.query)
		if f.gelesen != g.gelesen || (!f.gelesen && g.grund != "projektion_ohne_schluessel") {
			t.Errorf("%s?%s: %+v, erwartet gelesen %v", pfad, f.query, g, f.gelesen)
		}
	}
	for _, f := range einfach {
		pruefe("/Products", f)
		pruefe("/Products(1)", f)
	}
	for _, f := range zusammen {
		pruefe("/Orders", f)
		pruefe("/Orders(OrderID=1,LineNo=2)", f)
	}
	if n != 42 {
		t.Errorf("Menge %d, erwartet 42", n)
	}
}

// grundBaustein ist ein Baustein einer Session des Generators S: die Änderungen an einem
// gültigen Exchange und der Grund des Überspringens ("" gelesen).
type grundBaustein struct {
	pfad, query string
	a           aenderung
	grund       string
}

func bausteine() []grundBaustein {
	return []grundBaustein{
		{"/Products", "", nil, ""}, {"/Products(1)", "", nil, ""}, {"/Orders", "$top=2", nil, ""},
		{"/Products", "", func(e *model.Exchange) { e.Method = "POST" }, "methode"},
		{"/Products/$count", "", nil, "pfad"}, {"/Nothing", "", nil, "pfad"},
		{"/Products", "", func(e *model.Exchange) { e.Capture = model.CaptureIncomplete }, "capture"},
		{"/Products", "", func(e *model.Exchange) { e.ResponseStatus = 404 }, "status"},
		{"/Products", "", func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyAbsent, "" }, "body"},
		{"/Products", "$apply=x", nil, "query"},
		{"/Products", "", ct("text/plain"), "kein_json"},
		{"/Products", "", ct("application/json;odata.metadata=none"), "ohne_kontext"},
		{"/Products", "$select=Name", nil, "projektion_ohne_schluessel"},
		{"/$metadata", "", func(e *model.Exchange) {
			e.ResponseHeaders = map[string][]string{"Content-Type": {"application/json"}}
		}, "kein_xml"},
	}
}

// TestZeileE6JederExchangeZaehltEinmal: jeder Grund mal zwei Formen wird nie geöffnet (22), und der
// Generator S (200 Sessions, 2000 Exchanges) zählt jeden Exchange genau einmal.
func TestZeileE6JederExchangeZaehltEinmal(t *testing.T) {
	// Die elf Gründe in je zwei Formen.
	type fall struct {
		grund string
		pfad  string
		q     string
		a     aenderung
	}
	var faelle []fall
	for _, s := range spuren() {
		faelle = append(faelle,
			fall{"methode", s.pfad, "", func(e *model.Exchange) { e.Method = "POST" }},
			fall{"capture", s.pfad, "", func(e *model.Exchange) { e.Capture = model.CaptureFailed }},
			fall{"status", s.pfad, "", status(500)},
			fall{"body", s.pfad, "", func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyDiscarded, "" }},
			fall{"query", s.pfad, "$compute=x", nil},
			fall{"kein_json", s.pfad, "", ct("text/plain")},
			fall{"ohne_kontext", s.pfad, "", ct("application/json;odata.metadata=none")},
			fall{"projektion_ohne_schluessel", s.pfad, "$select=Name", nil},
		)
	}
	faelle = append(faelle,
		fall{"pfad", "/Unknown", "", nil}, fall{"pfad", "/Products(1)/Category", "", nil},
		fall{"kein_xml", "/$metadata", "", ct("application/json")}, fall{"kein_xml", "/$metadata", "", func(e *model.Exchange) { e.ResponseHeaders = nil }},
	)
	if len(faelle) != 20 {
		t.Fatalf("Formen %d", len(faelle))
	}
	for _, f := range faelle {
		if f.grund == "kein_xml" {
			w := neu(t)
			b := w.standard()
			e := b.roh(f.pfad, f.q, f.a)
			r := mustImport(t, w, anfrage())
			if uebersprungen(t, r)["kein_xml"] != 1 || slices.ContainsFunc(w.p.argsVon("Blobs.Open"), func(a []any) bool { return a[0] == e.ResponseBlob }) {
				t.Errorf("kein_xml: %v", uebersprungen(t, r))
			}
			continue
		}
		if g := probe(t, f.pfad, f.q, f.a); g.grund != f.grund {
			t.Errorf("%s %s?%s: %+v", f.grund, f.pfad, f.q, g)
		}
	}
	// metadata_doppelt in zwei Formen (n gleich 2 und 3); das zweite und dritte Dokument wird nicht geöffnet.
	for _, n := range []int{2, 3} {
		w := neu(t)
		b := w.sitzung()
		var es []model.Exchange
		for i := 0; i < n; i++ {
			es = append(es, b.metadaten("<edmx/>"))
		}
		r := mustImport(t, w, anfrage())
		if uebersprungen(t, r)["metadata_doppelt"] != n-1 || w.p.anzahl("Blobs.Open") != 1 {
			t.Errorf("metadata_doppelt n=%d: %v, Open %d", n, uebersprungen(t, r), w.p.anzahl("Blobs.Open"))
		}
		_ = es
	}
	// Generator S.
	zufall := rand.New(rand.NewPCG(7, 11))
	bs := bausteine()
	gesamt := 0
	for sess := 0; sess < 200; sess++ {
		w := neu(t)
		b := w.standard()
		soll := map[string]int{}
		gelesen := 1
		for i := 0; i < 10; i++ {
			bs1 := bs[zufall.IntN(len(bs))]
			e := b.roh(bs1.pfad, bs1.query, bs1.a)
			b.vorgabe(e, datenVorgabe{seite: seite(produkt(1, "A", "1")), entity: produkt(1, "A", "1")})
			if bs1.grund == "" {
				gelesen++
			} else {
				soll[bs1.grund]++
			}
		}
		gesamt += 11
		r := mustImport(t, w, anfrage())
		ex := bericht(t, r)["report"].(map[string]any)["exchanges"].(map[string]any)
		if int(ex["read"].(float64)) != gelesen || int(ex["total"].(float64)) != 11 || !maps.Equal(uebersprungen(t, r), soll) {
			t.Fatalf("Session %d: read %v total %v skipped %v, erwartet %d und %v", sess, ex["read"], ex["total"], uebersprungen(t, r), gelesen, soll)
		}
		sum := 0
		var gruende []string
		for _, g := range ex["skipped"].([]any) {
			z := g.(map[string]any)
			sum += int(z["count"].(float64))
			gruende = append(gruende, z["reason"].(string))
		}
		if gelesen+sum != 11 || !slices.IsSorted(gruende) {
			t.Fatalf("Session %d: Summe %d, Gründe %v", sess, gelesen+sum, gruende)
		}
		if w.p.anzahl("Blobs.Open") != gelesen {
			t.Fatalf("Session %d: Open %d, gelesen %d", sess, w.p.anzahl("Blobs.Open"), gelesen)
		}
	}
	if gesamt != 2200 {
		t.Errorf("Exchanges %d", gesamt)
	}
}

// TestZeileE7AufrufJeSpur: ParseCollection oder ParseEntity mit den Argumenten unverändert (Menge 12).
func TestZeileE7AufrufJeSpur(t *testing.T) {
	n := 0
	for _, s := range []struct{ name, pfad, set string }{
		{"Collection", "/Products", "Products"}, {"Collection", "/Orders", "Orders"},
		{"Entity", "/Products(1)", "Products"}, {"Entity", "/Orders(OrderID=1,LineNo=2)", "Orders"},
	} {
		for _, grenze := range []int64{1, 8 << 20, 1<<63 - 1} {
			n++
			w := neu(t)
			w.konfig.MaxCaptureBody = grenze
			b := w.standard()
			e := b.roh(s.pfad, "")
			b.vorgabe(e, datenVorgabe{seite: seite(produkt(1, "A", "1")), entity: produkt(1, "A", "1")})
			mustImport(t, w, anfrage())
			op, anders := "ParseCollection", "ParseEntity"
			if s.name == "Entity" {
				op, anders = anders, op
			}
			as := w.daten.aufrufe[op]
			if len(as) != 1 || len(w.daten.aufrufe[anders]) != 0 {
				t.Fatalf("%s %s: %d Aufrufe, %d der anderen Methode", s.name, s.pfad, len(as), len(w.daten.aufrufe[anders]))
			}
			strome := w.blobs.strome()
			a := as[0]
			if a.src != any(strome[1]) || a.entitySet != s.set || a.grenze != grenze || !modellGleich(a.m, w.meta.vorgabe.modell) {
				t.Errorf("%s %s Grenze %d: Aufruf %+v", s.name, s.pfad, grenze, a)
			}
			if strome[1].geschlossen() != 1 || w.p.ctxVon(op)[0] == nil {
				t.Errorf("%s %s: Close %d", s.name, s.pfad, strome[1].geschlossen())
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge %d", n)
	}
}

func modellGleich(a, b model.ServiceModel) bool {
	return fmt.Sprintf("%#v", a.EntitySets()) == fmt.Sprintf("%#v", b.EntitySets()) &&
		fmt.Sprintf("%#v", a.EntityTypes()) == fmt.Sprintf("%#v", b.EntityTypes())
}

// bedingung ist eine Bedingung des Überspringens: ihr Grund und die Änderung, die sie erfüllt.
type bedingung struct {
	grund string
	a     aenderung
	query string
	kj    bool
	ok    bool
}

// TestZeileE8Rangfolge: für jede Teilmenge der Bedingungen einer Spur (512, 256, 64) gilt der
// Grund der ersten Bedingung der Reihenfolge; die 79 Paare sind die Teilmengen der Größe 2.
func TestZeileE8Rangfolge(t *testing.T) {
	coll := []bedingung{
		{grund: "methode", a: func(e *model.Exchange) { e.Method = "POST" }},
		{grund: "pfad"},
		{grund: "capture", a: func(e *model.Exchange) { e.Capture = model.CaptureIncomplete }},
		{grund: "status", a: status(404)},
		{grund: "body", a: func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyAbsent, "" }},
		{grund: "query", query: "$apply=x"},
		{grund: "kein_json", kj: true},
		{grund: "ohne_kontext", ok: true},
		{grund: "projektion_ohne_schluessel", query: "$select=Name"},
	}
	ohnePfad := slices.Concat(coll[:1], coll[2:])
	meta := []bedingung{
		{grund: "methode", a: func(e *model.Exchange) { e.Method = "POST" }},
		{grund: "capture", a: func(e *model.Exchange) { e.Capture = model.CaptureIncomplete }},
		{grund: "status", a: status(404)},
		{grund: "body", a: func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyAbsent, "" }},
		{grund: "query", query: "a=b"},
		{grund: "kein_xml", kj: true},
	}
	zaehler := map[int]int{}
	lauf := func(art, basis string, bs []bedingung) {
		for maske := 1; maske < 1<<len(bs); maske++ {
			pfad := basis
			var as []aenderung
			var qs []string
			kj, ok := false, false
			erwartet := ""
			for i, b := range bs {
				if maske&(1<<i) == 0 {
					continue
				}
				if erwartet == "" {
					erwartet = b.grund
				}
				switch {
				case b.grund == "pfad":
					pfad = "/Unknown"
				case b.a != nil:
					as = append(as, b.a)
				}
				if b.query != "" {
					qs = append(qs, b.query)
				}
				kj, ok = kj || b.kj, ok || b.ok
			}
			switch {
			case art == "Metadaten" && kj:
				as = append(as, ct("application/json"))
			case kj && ok:
				as = append(as, ct("text/plain;odata.metadata=none"))
			case kj:
				as = append(as, ct("text/plain"))
			case ok:
				as = append(as, ct("application/json;odata.metadata=none"))
			}
			if bits := popcount(maske); bits == 2 {
				zaehler[len(bs)]++
			}
			var g probeErgebnis
			if art == "Metadaten" {
				g = probeMeta(t, strings.Join(qs, "&"), as)
			} else {
				g = probe(t, pfad, strings.Join(qs, "&"), as...)
			}
			if g.grund != erwartet {
				t.Errorf("%s Maske %b: Grund %q, erwartet %q", art, maske, g.grund, erwartet)
			}
		}
	}
	lauf("Collection", "/Products", coll)
	lauf("Entity", "/Products(1)", ohnePfad)
	lauf("Metadaten", "/$metadata", meta)
	if zaehler[9]+zaehler[8]+zaehler[6] != 36+28+15 {
		t.Errorf("Paare: %d, erwartet 79", zaehler[9]+zaehler[8]+zaehler[6])
	}
}

func popcount(m int) int {
	n := 0
	for ; m > 0; m >>= 1 {
		n += m & 1
	}
	return n
}

// probeMeta importiert eine Session mit einem gültigen $metadata und einem zweiten, das die
// Änderungen trägt; es liefert den Grund, aus dem das zweite übersprungen wurde.
func probeMeta(t *testing.T, query string, as []aenderung) probeErgebnis {
	t.Helper()
	w := neu(t)
	b := w.sitzung()
	b.metadaten("<edmx/>")
	b.metadaten("<edmx 2/>", append([]aenderung{func(e *model.Exchange) { e.RawQuery = query }}, as...)...)
	// Das zweite Dokument hat eine andere Referenz; ist es ein Kandidat, ist der Import abgewiesen.
	r, err := w.importiere(anfrage())
	if err != nil {
		return probeErgebnis{grund: "KANDIDAT"}
	}
	for g := range uebersprungen(t, r) {
		return probeErgebnis{grund: g}
	}
	return probeErgebnis{}
}
