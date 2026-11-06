package jsoncodec_test

import (
	"reflect"
	"runtime"
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

// eZuteilungFormen sind die acht Formen der Zeile 29; jede liefert einen Body von etwa der gewünschten Größe.
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
		{"Mitglieder", func(g int) string { return namen("a", g) }},
		{"Escape-Folgen in einem String", func(g int) string { return `{"@s":"` + strings.Repeat(esc("0041"), g/6) + `"}` }},
		{"lange Namen", func(g int) string { return `{"` + strings.Repeat("a", g) + `":1}` }},
		{"lange Strings", func(g int) string { return `{"@s":"` + strings.Repeat("a", g) + `"}` }},
		{"lauter Annotationen", func(g int) string { return namen("@a", g) }},
	}
}

// TestEingabeZuteilung hält Zeile 29: die Gesamtzuteilung eines Aufrufs (`TotalAlloc`) ist höchstens 256 · n + 1 MiB, n die Zahl der Bytes aus `src`.
// Menge 24 = 8 Formen × 3 Größen; das Verhältnis steht im Protokoll (`go test -v`).
func TestEingabeZuteilung(t *testing.T) {
	n := 0
	for _, f := range eZuteilungFormen() {
		for _, g := range []int{1 << 10, 64 << 10, 1 << 20} {
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
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
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
			r1, err1 := l.LeseEntity(a.ctx, a.src, a.ct, a.m, a.set, a.art, a.grenze)
			if err1 != nil || errFrisch != nil {
				t.Fatalf("%s %s: %v %v", art, body, err1, errFrisch)
			}
			r1.Typ, r1.Schluessel = "X", "X"
			if len(r1.Properties) > 0 {
				r1.Properties[0].Name, r1.Properties[0].Wert.Text = "X", "X"
			}
			r1.Properties = append(r1.Properties, ePW("Y", model.ArtString, "Y"))
			b := eA(set, art, body)
			r2, err2 := l.LeseEntity(b.ctx, b.src, b.ct, b.m, b.set, b.art, b.grenze)
			if err2 != nil || !reflect.DeepEqual(r2, frisch) {
				t.Errorf("%s %s: zweiter Aufruf %+v %v, erwartet %+v", art, body, r2, err2, frisch)
			}
		}
	}
}

// eGleichzeitigFaelle sind die vier Bodys der Zeile 39 (gültig, `fehlend`, `wert`, `doppelter_name`) an Ws.
func eGleichzeitigFaelle() []string {
	return []string{`{"Id":1,"S":"x","Dec":1.50}`, `{"S":"x"}`, `{"Id":1,"I32":1.5}`, `{"Id":1,"Id":2}`}
}

// TestEingabeGleichzeitig hält Zeile 39: gleichzeitige Aufrufe auf einer Instanz beeinflussen sich nicht; jedes Ergebnis gleicht dem eines
// einzelnen Aufrufs. 32 Goroutinen × 3 Arten × 4 Bodys, je 25 Runden. Der Zähler zählt abweichende Ergebnisse; unter `make test-race` meldet der
// Detektor den ungeordneten Zugriff.
func TestEingabeGleichzeitig(t *testing.T) {
	type erg struct {
		e   model.EntityEingabe
		err string
	}
	einzeln := func(l *jsoncodec.EingabeLeser, art model.EingabeArt, body string) erg {
		a := eA("Ws", art, body)
		e, err := l.LeseEntity(a.ctx, a.src, a.ct, a.m, a.set, a.art, a.grenze)
		if err != nil {
			return erg{e, err.Error()}
		}
		return erg{e, ""}
	}
	soll := map[string]erg{}
	for _, art := range eArten() {
		for i, b := range eGleichzeitigFaelle() {
			soll[string(art)+eP(i)] = einzeln(jsoncodec.NewEingabe(), art, b)
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
			for range 25 {
				for _, art := range eArten() {
					for i, b := range eGleichzeitigFaelle() {
						aufrufe.Add(1)
						if got := einzeln(l, art, b); !reflect.DeepEqual(got, soll[string(art)+eP(i)]) {
							abweichend.Add(1)
						}
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if abweichend.Load() != 0 || aufrufe.Load() != 32*25*12 {
		t.Errorf("%d von %d Aufrufen weichen vom Ergebnis des einzelnen Aufrufs ab", abweichend.Load(), aufrufe.Load())
	}
}

// TestEingabeKeinZustand hält Zeile 40: der Leser hat keinen Zustand zwischen Aufrufen; derselbe Leser für die drei Arten nacheinander und in umgekehrter
// Folge ergibt jeweils das Ergebnis eines frischen Lesers. Menge 3 Arten × 2 Folgen × 2 Bodys.
func TestEingabeKeinZustand(t *testing.T) {
	l := jsoncodec.NewEingabe()
	arten := eArten()
	folgen := [][]model.EingabeArt{arten, {arten[2], arten[1], arten[0]}}
	for _, body := range []string{`{"Id":1,"S":"x"}`, `{"S":"x"}`, `{}`} {
		for _, folge := range folgen {
			for _, art := range folge {
				a := eA("Ws", art, body)
				got, err := l.LeseEntity(a.ctx, a.src, a.ct, a.m, a.set, a.art, a.grenze)
				b := eA("Ws", art, body)
				frisch, errFrisch := jsoncodec.NewEingabe().LeseEntity(b.ctx, b.src, b.ct, b.m, b.set, b.art, b.grenze)
				if !reflect.DeepEqual(got, frisch) || (err == nil) != (errFrisch == nil) || (err != nil && err.Error() != errFrisch.Error()) {
					t.Errorf("%s %s: %+v %v, frisch %+v %v", art, body, got, err, frisch, errFrisch)
				}
			}
		}
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
