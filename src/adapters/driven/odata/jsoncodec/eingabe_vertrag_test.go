package jsoncodec_test

import (
	"io"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eZaehler ist ein src, der nur die Bytes zählt, die er liefert.
type eZaehler struct {
	r *strings.Reader
	n int64
}

func (z *eZaehler) Read(p []byte) (int, error) {
	n, err := z.r.Read(p)
	z.n += int64(n)
	return n, err
}

// eZuteilungFormen sind die neun Formen der Zeile 29; jede liefert einen Body von etwa der gewünschten Größe.
func eZuteilungFormen() []struct {
	name string
	body func(groesse int) string
} {
	namen := func(praefix string, groesse int) string {
		var b strings.Builder
		b.WriteByte('{')
		for i := 0; b.Len() < groesse; i++ {
			b.WriteString(`"` + praefix + strconv.Itoa(i) + `":1,`)
		}
		b.WriteString(`"z":1}`)
		return b.String()
	}
	return []struct {
		name string
		body func(groesse int) string
	}{
		{"lauter [", func(g int) string { return strings.Repeat("[", g) }},
		{"lauter {\"a\":", func(g int) string { return strings.Repeat(`{"a":`, g/5) }},
		{"[0,-Folge", func(g int) string { return `{"@a":[` + strings.Repeat("0,", g/2) + `0]}` }},
		{"1,-Folge", func(g int) string { return `{"@a":[` + strings.Repeat("1,", g/2) + `1]}` }},
		{"Mitglieder", func(g int) string { return namen("a", g) }},
		{"Escape-Folgen in einem String", func(g int) string { return `{"@s":"` + strings.Repeat(esc("0041"), g/6) + `"}` }},
		{"lange Namen", func(g int) string { return `{"` + strings.Repeat("a", g) + `":1}` }},
		{"lange Strings", func(g int) string { return `{"@s":"` + strings.Repeat("a", g) + `"}` }},
		{"lauter Annotationen", func(g int) string { return namen("@a", g) }},
	}
}

// TestEingabeZuteilung hält Zeile 29: die Gesamtzuteilung eines Aufrufs (`TotalAlloc`) ist höchstens 256 · n + 1 MiB, n die Zahl der Bytes aus `src`.
// Die Zuteilung der Formen mit vielen Knoten hängt von der Wachstumsstufe des Slice `kinder` ab (sägezahnförmig, Maximum 224,3 bei etwa 4,08 MiB
// und 223,8 bei 8 MiB); die `1,`-Folge läuft deshalb über die Stufen bis zur Standardgrenze von 8 MiB, die übrigen Formen über 1 KiB, 64 KiB und
// 1 MiB. Größen über 14 MiB sind ungeprüft (akzeptiertes Negativ, Plan §3). Menge 8 Formen × 3 Größen + `1,`-Folge × 11 Größen = 35; das Verhältnis
// steht im Protokoll (`go test -v`).
func TestEingabeZuteilung(t *testing.T) {
	n := 0
	for _, f := range eZuteilungFormen() {
		groessen := []int{1 << 10, 64 << 10, 1 << 20}
		if f.name == "1,-Folge" {
			groessen = []int{1 << 10, 64 << 10, 1 << 20, 2 << 20, 3 << 20, 4 << 20, 4278618, 5 << 20, 6 << 20, 7 << 20, 8<<20 - 16}
		}
		for _, g := range groessen {
			body := f.body(g)
			z := &eZaehler{r: strings.NewReader(body)}
			a := eA("Ms", model.EingabePatch, "")
			a.src = z
			var vor, nach runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&vor)
			_, _ = a.los()
			runtime.ReadMemStats(&nach)
			zugeteilt := int64(nach.TotalAlloc - vor.TotalAlloc)
			grenze := 256*z.n + 1<<20
			t.Logf("%-32s Größe %7d: %9d Bytes zugeteilt für %7d gelesen (%.1f je Byte), Grenze %d", f.name, g, zugeteilt, z.n, float64(zugeteilt)/float64(max(z.n, 1)), grenze)
			if zugeteilt > grenze {
				t.Fatalf("%s %d: %d Bytes zugeteilt, höchstens %d erlaubt (n = %d)", f.name, g, zugeteilt, grenze, z.n)
			}
			n++
		}
	}
	if n != 35 {
		t.Errorf("Menge %d, erwartet 35", n)
	}
}

// TestEingabeOrt hält Zeile 36: der Ort ist ein JSON-Pointer, `~` als `~0`, `/` als `~1`, sonst nichts ersetzt; `Error()` quotet den Ort. Menge 8.
func TestEingabeOrt(t *testing.T) {
	faelle := []struct{ name, ort string }{
		{"Price", "/Price"}, {"a~b", "/a~0b"}, {"a/b", "/a~1b"}, {"a" + bs + "b", "/a" + bs + "b"}, {`a"b`, `/a"b`}, {"a\x01b", "/a\x01b"},
		{"\xc3\xa9", "/\xc3\xa9"}, {"~/", "/~0~1"},
	}
	for _, f := range faelle {
		got, err := eA("Ws", model.EingabePatch, eObj("Id", "1", f.name, "1")).los()
		eFehler(t, "Name "+strconv.Quote(f.name), got, err, model.EingabeUnbekannteProperty, f.ort)
		if want := "ODR-E6007: Request-Body ungültig: unbekannte_property " + strconv.Quote(f.ort); err != nil && err.Error() != want {
			t.Errorf("Error() %q, erwartet %q", err.Error(), want)
		}
	}
	got, err := eA("Ws", model.EingabePatch, "5").los()
	eFehler(t, "Wurzel", got, err, model.EingabeStruktur, "")
	if want := "ODR-E6007: Request-Body ungültig: struktur"; err.Error() != want {
		t.Errorf("Error() %q, erwartet %q", err.Error(), want)
	}
}

// TestEingabeErgebnisGehoertDemAufrufer hält Zeile 38: der Aufrufer ändert das Ergebnis, ein früherer und ein späterer Aufruf bleiben gleich. Menge 6.
func TestEingabeErgebnisGehoertDemAufrufer(t *testing.T) {
	l := jsoncodec.NewEingabe()
	for _, art := range eArten() {
		for _, f := range [][2]string{{"Ws", `{"Id":1,"S":"x"}`}, {"OhneKey", `{}`}} {
			set, body := f[0], f[1]
			frisch, errFrisch := eA(set, art, body).los()
			a := eA(set, art, body)
			r1, err1 := l.LeseEntity(a.kontext(), a.src, a.ct, a.m, a.set, a.art, a.grenze)
			if err1 != nil || errFrisch != nil {
				t.Fatalf("%s %s: %v %v", art, body, err1, errFrisch)
			}
			frisch.Properties = slices.Clone(frisch.Properties)
			r1.Typ, r1.Schluessel = "X", "X"
			if len(r1.Properties) > 0 {
				r1.Properties[0].Name, r1.Properties[0].Wert.Text = "X", "X"
			}
			r1.Properties = append(r1.Properties, ePW("Y", model.ArtString, "Y"))
			b := eA(set, art, body)
			r2, err2 := l.LeseEntity(b.kontext(), b.src, b.ct, b.m, b.set, b.art, b.grenze)
			if err2 != nil || !reflect.DeepEqual(r2, frisch) {
				t.Errorf("%s %s: zweiter Aufruf %+v %v, erwartet %+v", art, body, r2, err2, frisch)
			}
			// der zweite Aufruf ändert das erste Ergebnis nicht: die Änderungen des Aufrufers stehen noch da
			if r1.Typ != "X" || r1.Properties[len(r1.Properties)-1].Name != "Y" || (len(frisch.Properties) > 0 && r1.Properties[0].Name != "X") {
				t.Errorf("%s %s: das erste Ergebnis hat sich durch den zweiten Aufruf geändert: %+v", art, body, r1)
			}
		}
	}
}

// eGleichzeitigFall ist ein Zweig des Lesers: der Aufruf mit Modell m und der Art, mit frischem Strom je Aufruf.
type eGleichzeitigFall struct {
	name string
	bau  func(m model.ServiceModel, art model.EingabeArt) eAufruf
}

// eGleichzeitigFaelle sind die Zweige der Zeile 39: gültig, `fehlend`, `wert`, `doppelter_name`, `bind`, `navigation`, `property_typ`, `offen`, gültiges
// und ungültiges `@odata.type`, `schluessel`, Programmfehler, `content_type`, `charset` gültig und ungültig und der Fehler von `src`; jeder Zweig, den der Leser ausführen kann, hat einen Aufruf.
func eGleichzeitigFaelle() []eGleichzeitigFall {
	mit := func(name, set, body string) eGleichzeitigFall {
		return eGleichzeitigFall{name, func(m model.ServiceModel, art model.EingabeArt) eAufruf {
			return eAm(m, set, art, body)
		}}
	}
	nul := esc("0000")
	return []eGleichzeitigFall{
		mit("gültig", "Ws", `{"Id":1,"S":"x","Dec":1.50}`), mit("fehlend", "Ws", `{"S":"x"}`), mit("wert", "Ws", `{"Id":1,"I32":1.5}`),
		mit("doppelter_name", "Ws", `{"Id":1,"Id":2}`), mit("bind", "Ms", `{"Id":1,"Einzel@odata.bind":"x"}`), mit("navigation", "Ms", `{"Id":1,"Einzel":{}}`),
		mit("property_typ", "Ms", `{"Id":1,"Dbl":1}`), mit("offen", "Os", `{"Id":1,"dyn":1}`),
		mit("odata_typ gültig", "Mitten", `{"@odata.type":"#Ns.Blatt","Id":1,"Extra":"x"}`), mit("odata_typ ungültig", "Mitten", `{"@odata.type":5}`),
		mit("schluessel", "Sts", `{"Id":"a`+nul+`"}`), mit("programmfehler", "Unbekannt", `{}`),
		{"content_type", func(m model.ServiceModel, art model.EingabeArt) eAufruf {
			a := eAm(m, "Ws", art, `{"Id":1}`)
			a.ct = []string{"text/plain"}
			return a
		}},
		{"charset ungültig", func(m model.ServiceModel, art model.EingabeArt) eAufruf {
			a := eAm(m, "Ws", art, `{"Id":1}`)
			a.ct = []string{"application/json;charset=utf-16"}
			return a
		}},
		{"charset gültig", func(m model.ServiceModel, art model.EingabeArt) eAufruf {
			a := eAm(m, "Ws", art, `{"Id":1}`)
			a.ct = []string{"application/json; charset=UTF-8;odata.metadata=minimal"}
			return a
		}},
		{"Fehler von src", func(m model.ServiceModel, art model.EingabeArt) eAufruf {
			a := eAm(m, "Ws", art, "")
			a.src = eNeuerStrom(eSchritt{daten: []byte(`{"Id":`)}, eSchritt{err: io.ErrUnexpectedEOF})
			return a
		}},
	}
}

// TestEingabeGleichzeitig hält Zeile 39: gleichzeitige Aufrufe auf einer Instanz beeinflussen sich nicht; jedes Ergebnis gleicht dem eines
// einzelnen Aufrufs. 32 Goroutinen × 3 Arten × 16 Zweige, je 8 Runden. Der Zähler zählt abweichende Ergebnisse; unter `make test-race` meldet der
// Detektor den ungeordneten Zugriff, in einem Zweig ohne Wirkung auf das Ergebnis allein er.
func TestEingabeGleichzeitig(t *testing.T) {
	type erg struct {
		e   model.EntityEingabe
		err string
	}
	m := eModell()
	einzeln := func(l *jsoncodec.EingabeLeser, f eGleichzeitigFall, art model.EingabeArt) erg {
		a := f.bau(m, art)
		e, err := l.LeseEntity(a.kontext(), a.src, a.ct, a.m, a.set, a.art, a.grenze)
		if err != nil {
			return erg{e, err.Error()}
		}
		return erg{e, ""}
	}
	soll := map[string]erg{}
	for _, art := range eArten() {
		for _, f := range eGleichzeitigFaelle() {
			soll[string(art)+f.name] = einzeln(jsoncodec.NewEingabe(), f, art)
		}
	}
	l := jsoncodec.NewEingabe()
	var abweichend, aufrufe atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 8 {
				for _, art := range eArten() {
					for _, f := range eGleichzeitigFaelle() {
						aufrufe.Add(1)
						if got := einzeln(l, f, art); !reflect.DeepEqual(got, soll[string(art)+f.name]) {
							abweichend.Add(1)
						}
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if abweichend.Load() != 0 || aufrufe.Load() != 32*8*3*16 {
		t.Errorf("%d von %d Aufrufen weichen vom Ergebnis des einzelnen Aufrufs ab", abweichend.Load(), aufrufe.Load())
	}
}

// TestEingabeKeinZustand hält Zeile 40: der Leser hat keinen Zustand zwischen Aufrufen; derselbe Leser für die drei Arten nacheinander und in umgekehrter
// Folge ergibt jeweils das Ergebnis, das eine Tabelle des Tests nennt (nicht das eines frischen Lesers: ein Zustand im Paket träfe ihn ebenso). Das Ergebnis
// zählt die Wirkung: der Fänger wartet nicht auf ein Datenrennen. Menge 3 Bodys × 2 Folgen × 3 Arten, dazu 4 Fälle × 2 Reihen × 3 Arten × 8 Wechsel zweier
// Modelle mit demselben Set.
func TestEingabeKeinZustand(t *testing.T) {
	l := jsoncodec.NewEingabe()
	arten := eArten()
	folgen := [][]model.EingabeArt{arten, {arten[2], arten[1], arten[0]}}
	idS := []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("S", model.ArtString, "x")}
	faelle := []struct {
		body      string
		ohnePatch bool // bei POST und PUT `fehlend`, bei PATCH das Ergebnis
		patch     model.EntityEingabe
		voll      model.EntityEingabe
	}{
		{`{"Id":1,"S":"x"}`, false, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: idS}, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: idS}},
		{`{"S":"x"}`, true, model.EntityEingabe{Typ: "Ns.W", Properties: []model.PropertyWert{ePW("S", model.ArtString, "x")}}, model.EntityEingabe{}},
		{`{}`, true, model.EntityEingabe{Typ: "Ns.W", Properties: []model.PropertyWert{}}, model.EntityEingabe{}},
	}
	n := 0
	for _, f := range faelle {
		for _, folge := range folgen {
			for _, art := range folge {
				a := eA("Ws", art, f.body)
				got, err := l.LeseEntity(a.kontext(), a.src, a.ct, a.m, a.set, a.art, a.grenze)
				switch {
				case f.ohnePatch && art != model.EingabePatch:
					eFehler(t, f.body+" "+string(art), got, err, model.EingabeFehlend, "/Id")
				case art == model.EingabePatch:
					eOK(t, f.body+" "+string(art), got, err, f.patch)
				default:
					eOK(t, f.body+" "+string(art), got, err, f.voll)
				}
				n++
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
	// Achse Modell: dasselbe Set `Ws` in zwei Modellen mit verschiedenem Typ von `S`; ein Cache der Deuter nach Name des Sets lieferte das Ergebnis des anderen.
	m1 := eModell()
	m2 := model.NeuesServiceModel("4.0", []string{"Ns"}, []model.EntityType{{Name: "Ns.W", Keys: []string{"Id"},
		Properties: []model.Property{prop("Id", "Edm.Int32", false), prop("S", "Edm.Int32", true)}}}, []model.EntitySet{{Name: "Ws", EntityType: "Ns.W"}}, nil)
	modelle := []struct {
		name   string
		m      model.ServiceModel
		body   string
		ursach model.EingabeUrsache
		props  []model.PropertyWert
	}{
		{"M1 String", m1, `{"Id":1,"S":"x"}`, "", idS},
		{"M2 String", m2, `{"Id":1,"S":"x"}`, model.EingabeTyp, nil},
		{"M2 Zahl", m2, `{"Id":1,"S":5}`, "", []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("S", model.ArtInt32, "5")}},
		{"M1 Zahl", m1, `{"Id":1,"S":5}`, model.EingabeTyp, nil},
	}
	k := 0
	for _, reihe := range [][]int{{0, 1, 2, 3, 0, 1, 2, 3}, {3, 2, 1, 0, 3, 2, 1, 0}} {
		for _, art := range arten {
			for _, i := range reihe {
				c := modelle[i]
				a := eA("Ws", art, c.body)
				a.m = c.m
				got, err := l.LeseEntity(a.kontext(), a.src, a.ct, a.m, a.set, a.art, a.grenze)
				if c.ursach != "" {
					eFehler(t, c.name+" "+string(art), got, err, c.ursach, "/S")
				} else {
					eOK(t, c.name+" "+string(art), got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: c.props})
				}
				k++
			}
		}
	}
	if k != 2*3*8 {
		t.Errorf("Menge Modelle %d, erwartet 48", k)
	}
}

// TestEingabeOhneZustand hält Zeile 41: der Leser hat kein Feld und erfüllt den Port (Prüfung zur Übersetzungszeit in eingabe_helfer_test.go).
func TestEingabeOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[jsoncodec.EingabeLeser]().NumField(); n != 0 {
		t.Errorf("EingabeLeser hat %d Felder", n)
	}
	if jsoncodec.NewEingabe() == nil {
		t.Error("NewEingabe liefert nil")
	}
}
