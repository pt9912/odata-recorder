package telemetry_test

import (
	"context"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// zaehlFehler, zaehlText und zaehlWert zählen ihre Aufrufe von Error, String und LogValue: "wird nicht
// gerufen" beweist ein Zähler und nicht die Zeile. Der Text beider ist eine gültige Kennung und ein
// gültiges Token, damit ein Aufruf eine Zeile mit dem Text ergäbe.
type zaehlFehler struct{ n *atomic.Int64 }

func (z zaehlFehler) Error() string { z.n.Add(1); return "svc1" }

type zaehlText struct{ n *atomic.Int64 }

func (z zaehlText) String() string { z.n.Add(1); return "svc1" }

type zaehlWert struct {
	n *atomic.Int64
	v slog.Value
}

func (z zaehlWert) LogValue() slog.Value { z.n.Add(1); return z.v }

type panikWert struct{ n *atomic.Int64 }

func (z panikWert) LogValue() slog.Value { z.n.Add(1); panic("LogValue gerät in Panik") }

// endlos löst sich zu sich selbst auf.
type endlos struct{}

func (e endlos) LogValue() slog.Value { return slog.AnyValue(e) }

// artFall ist ein Wert, wie ihn der Aufrufer übergibt (v), und wie das Modell ihn nach dem Auflösen
// erwartet (modell). zaehler zählt die Aufrufe von Error, String oder LogValue; lv sagt, dass der
// Wert ein LogValuer ist, dessen LogValue je Eintrag (Aufruf) oder einmal beim Ableiten (WithAttrs)
// gerufen wird.
type artFall struct {
	name    string
	key     string
	v       slog.Value
	modell  slog.Value
	zaehler *atomic.Int64
	lv      bool
}

// artWerte liefert je Art von slog.Kind einen Wert für den Schlüssel service_id; anyWert ist der Wert
// der Art Any.
func artWerte(anyWert slog.Value) map[slog.Kind]slog.Value {
	n := new(atomic.Int64)
	return map[slog.Kind]slog.Value{
		slog.KindAny:       anyWert,
		slog.KindBool:      slog.BoolValue(true),
		slog.KindDuration:  slog.DurationValue(1500 * time.Microsecond),
		slog.KindFloat64:   slog.Float64Value(1.5),
		slog.KindInt64:     slog.Int64Value(7),
		slog.KindString:    slog.StringValue("svc1"),
		slog.KindTime:      slog.TimeValue(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)),
		slog.KindUint64:    slog.Uint64Value(7),
		slog.KindGroup:     slog.GroupValue(s("a", "b"), s("c", "d"), s("e", "f")),
		slog.KindLogValuer: slog.AnyValue(zaehlWert{n, slog.StringValue("svc1")}),
	}
}

// zaehlprobeArten prüft die Zahl der Arten gegen die Quelle: die Art hinter KindLogValuer gibt es
// nicht, und die Standardbibliothek nennt sie <unknown slog.Kind>.
func zaehlprobeArten(t *testing.T, werte map[slog.Kind]slog.Value) []slog.Kind {
	t.Helper()
	var arten []slog.Kind
	for k := slog.KindAny; k <= slog.KindLogValuer; k++ {
		arten = append(arten, k)
		if _, ok := werte[k]; !ok {
			t.Fatalf("kein Wert für die Art %s", k)
		}
	}
	if (slog.KindLogValuer+1).String() != "<unknown slog.Kind>" || len(werte) != len(arten) {
		t.Fatalf("%d Werte, %d Arten; die Art hinter LogValuer heißt %q", len(werte), len(arten), (slog.KindLogValuer + 1).String())
	}
	return arten
}

// artFaelle sind die 53 Werte der Menge M11: je Art × 3 Schlüssel (service_id String, duration_ms
// Float, befunde Int), 18 Grenzfälle und 5 LogValuer.
func artFaelle(t *testing.T) []artFall {
	t.Helper()
	nAny := new(atomic.Int64)
	werte := artWerte(slog.AnyValue(zaehlText{nAny}))
	var faelle []artFall
	for _, k := range zaehlprobeArten(t, werte) {
		for _, key := range []string{"service_id", "duration_ms", "befunde"} {
			v, m := werte[k], werte[k]
			zaehler := new(atomic.Int64)
			switch k {
			case slog.KindAny:
				v = slog.AnyValue(zaehlText{zaehler})
			case slog.KindLogValuer:
				v, m = slog.AnyValue(zaehlWert{zaehler, slog.StringValue("svc1")}), slog.StringValue("svc1")
			}
			faelle = append(faelle, artFall{name: k.String() + " in " + key, key: key, v: v, modell: m, zaehler: zaehler, lv: k == slog.KindLogValuer})
		}
	}
	grenz := func(name, key string, v slog.Value) {
		faelle = append(faelle, artFall{name: name, key: key, v: v, modell: v})
	}
	for _, u := range []uint64{math.MaxInt64, math.MaxInt64 + 1, math.MaxUint64} {
		grenz("Uint64 "+slog.Uint64Value(u).String(), "befunde", slog.Uint64Value(u))
	}
	for _, i := range []int64{-1, 0, math.MaxInt64} {
		grenz("Int64 "+slog.Int64Value(i).String(), "befunde", slog.Int64Value(i))
	}
	for _, d := range []time.Duration{0, 1, -1} {
		grenz("Duration "+d.String(), "duration_ms", slog.DurationValue(d))
	}
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Copysign(0, -1), 0, 1.5} {
		grenz("Float64 "+slog.Float64Value(f).String(), "duration_ms", slog.Float64Value(f))
	}
	zf, zt := new(atomic.Int64), new(atomic.Int64)
	faelle = append(faelle,
		artFall{name: "Any error", key: "mode", v: slog.AnyValue(zaehlFehler{zf}), modell: slog.AnyValue(zaehlFehler{zf}), zaehler: zf},
		artFall{name: "Any Stringer", key: "mode", v: slog.AnyValue(zaehlText{zt}), modell: slog.AnyValue(zaehlText{zt}), zaehler: zt},
	)
	grenz("Any nil", "mode", slog.AnyValue(nil))
	grenz("Any String", "mode", slog.AnyValue("record"))
	lvFall := func(name, key string, lv slog.LogValuer, n *atomic.Int64, aufgeloest slog.Value) {
		faelle = append(faelle, artFall{name: name, key: key, v: slog.AnyValue(lv), modell: aufgeloest, zaehler: n, lv: true})
	}
	n1, n2, n3, n4 := new(atomic.Int64), new(atomic.Int64), new(atomic.Int64), new(atomic.Int64)
	lvFall("LogValuer zu String", "mode", zaehlWert{n1, slog.StringValue("record")}, n1, slog.StringValue("record"))
	lvFall("LogValuer zu Int64", "befunde", zaehlWert{n2, slog.Int64Value(4)}, n2, slog.Int64Value(4))
	lvFall("LogValuer zu Bool", "mode", zaehlWert{n3, slog.BoolValue(true)}, n3, slog.BoolValue(true))
	lvFall("LogValuer mit Panik", "mode", panikWert{n4}, n4, slog.AnyValue(nil))
	grenz("LogValuer löst endlos zu sich selbst auf", "mode", slog.AnyValue(endlos{}))
	faelle[len(faelle)-1].modell = slog.AnyValue(nil)
	return faelle
}

// stelle schreibt n Einträge mit dem Attribut x, als Attribut des Aufrufs oder über WithAttrs
// vorgelegt, hinter einer Gruppe grp (leer: keine) und prüft Zeile und Verwürfe gegen das Modell mit
// dem aufgelösten Wert mv.
func stelle(t *testing.T, wo, grp string, x slog.Attr, mv slog.Value, vorgelegt bool, n int) {
	t.Helper()
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	h, praefix := a.Handler(), ""
	if grp != "" {
		h, praefix = h.WithGroup(grp), grp+"."
	}
	var aufruf []slog.Attr
	if vorgelegt {
		h = h.WithAttrs([]slog.Attr{x})
	} else {
		aufruf = []slog.Attr{x}
	}
	e := modell([]modAttr{{praefix + x.Key, mv}})
	l := slog.New(h)
	for range n {
		l.LogAttrs(context.Background(), slog.LevelInfo, "m", aufruf...)
	}
	z := senke.zeilen()
	if len(z) != n {
		t.Errorf("%s: %d Zeilen, erwartet %d", wo, len(z), n)
		return
	}
	for _, zeile := range z {
		pruefeZeile(t, wo, zeile, "INFO", "m", e)
	}
	for g, c := range a.Snapshot().Dropped {
		want := map[support.Verwurf]uint64{support.VerwurfSchluessel: uint64(n) * e.schluessel, support.VerwurfWert: uint64(n) * e.wert}[g]
		if c != want {
			t.Errorf("%s: Grund %s zählt %d, erwartet %d", wo, g, c, want)
		}
	}
}

// H02: der Handler bildet die Arten ab (SPEC-231 Nr. 5): Menge M11 (53 Werte × 2 Stellen = 106), drei
// Einträge je Fall. Orakel: die Modelltabelle; der Text eines error fehlt in der Zeile; Error() und
// String() werden nie gerufen, LogValue je Eintrag (Aufruf) oder einmal beim Ableiten (WithAttrs).
func TestSlogArten(t *testing.T) {
	const eintraege = 3
	n := 0
	for _, vorgelegt := range []bool{false, true} {
		faelle := artFaelle(t)
		if len(faelle) != (int(slog.KindLogValuer)+1)*3+18+5 {
			t.Fatalf("%d Werte, erwartet %d", len(faelle), (int(slog.KindLogValuer)+1)*3+18+5)
		}
		for _, f := range faelle {
			n++
			stelle(t, f.name, "", slog.Attr{Key: f.key, Value: f.v}, f.modell, vorgelegt, eintraege)
			if f.zaehler == nil {
				continue
			}
			want := int64(0)
			switch {
			case f.lv && vorgelegt:
				want = 1
			case f.lv:
				want = eintraege
			}
			if got := f.zaehler.Load(); got != want && f.name != "LogValuer löst endlos zu sich selbst auf" {
				t.Errorf("%s, vorgelegt %t: %d Aufrufe, erwartet %d", f.name, vorgelegt, got, want)
			}
		}
	}
	if n != 106 {
		t.Fatalf("%d Fälle, erwartet 106", n)
	}
}

// H03: zuerst der Schlüssel, dann die Art; eine Gruppe ist ein Attribut; ein leerer Schlüssel hat keine
// Sonderregel. Menge M12: 4 Schlüssel (gültig, fremd, leer samt Nullwert von slog.Attr, gültig hinter
// einer Gruppe) × 10 Arten × 2 Stellen = 80. Orakel: genau ein Verwurf je Attribut; die Gruppe mit drei
// Elementen zählt 1.
func TestHandlerRangfolge(t *testing.T) {
	werte := artWerte(slog.Value{})
	arten := zaehlprobeArten(t, werte)
	schluessel := []struct{ name, key, grp string }{
		{"gültig", "service_id", ""}, {"fremd", "fremd", ""}, {"leer", "", ""}, {"hinter Gruppe", "service_id", "g"},
	}
	n := 0
	for _, sk := range schluessel {
		for _, k := range arten {
			for _, vorgelegt := range []bool{false, true} {
				n++
				v, m := werte[k], werte[k]
				if k == slog.KindLogValuer {
					v, m = slog.AnyValue(zaehlWert{new(atomic.Int64), slog.StringValue("svc1")}), slog.StringValue("svc1")
				}
				stelle(t, sk.name+", "+k.String(), sk.grp, slog.Attr{Key: sk.key, Value: v}, m, vorgelegt, 1)
			}
		}
	}
	if n != len(schluessel)*len(arten)*2 || n != 80 {
		t.Fatalf("%d Fälle, erwartet 80", n)
	}
}

// attrFall ist ein Fall der Menge M14 für WithAttrs: Attribute eines Elternhandlers, die abgeleiteten
// Attribute und die des Aufrufs.
type attrFall struct {
	name                  string
	eltern, attrs, aufruf []slog.Attr
}

func zweiterWert(key string) support.Attr {
	z, _ := katalogZeile(key)
	return andererWert(z)
}

// attrFaelle sind die 55 Fälle von WithAttrs der Menge M14: 13 Schlüssel × (gültig, ungültiger Wert,
// doppelt zum Aufruf), ein fremder Schlüssel, nil und eine leere Liste, 13 Schlüssel doppelt zwischen
// Elternhandler und Kind.
func attrFaelle(t *testing.T) []attrFall {
	t.Helper()
	var faelle []attrFall
	for _, key := range logSchluessel(t) {
		gut, anders := slogAttr(gueltigFuer(key)), slogAttr(zweiterWert(key))
		faelle = append(faelle,
			attrFall{name: key + " gültig", attrs: []slog.Attr{gut}},
			attrFall{name: key + " ungültiger Wert", attrs: []slog.Attr{s(key, "a b")}},
			attrFall{name: key + " doppelt zum Aufruf", attrs: []slog.Attr{gut}, aufruf: []slog.Attr{anders}},
		)
	}
	faelle = append(faelle, attrFall{name: "fremder Schlüssel", attrs: []slog.Attr{s("fremd", "x")}},
		attrFall{name: "nil"}, attrFall{name: "leere Liste", attrs: []slog.Attr{}})
	for _, key := range logSchluessel(t) {
		faelle = append(faelle, attrFall{name: key + " doppelt zwischen Eltern und Kind",
			eltern: []slog.Attr{slogAttr(gueltigFuer(key))}, attrs: []slog.Attr{slogAttr(zweiterWert(key))}})
	}
	return faelle
}

// H05 (1): WithAttrs legt Attribute vor, die bei jedem Eintrag geprüft und gezählt werden. Menge M14
// (55 Fälle) × Menge M17 (4 Arten von Handlern) × 3 Einträge. Orakel: das Modell der Kette; WithAttrs
// selbst zählt 0; der Aufrufer überschreibt den Slice nach WithAttrs, die Zeile bleibt (Kopie); der
// Empfänger schreibt vor und nach dem Ableiten die Zeile seines Modells.
func TestWithAttrs(t *testing.T) {
	faelle := attrFaelle(t)
	if len(faelle) != 55 {
		t.Fatalf("%d Fälle, erwartet 55", len(faelle))
	}
	n := 0
	for art := range 4 {
		for _, f := range faelle {
			n++
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			k := ketten(a)[art]
			wo := k.name + ", " + f.name
			if len(f.eltern) > 0 {
				k = k.mitAttrs(k.h.WithAttrs(slices.Clone(f.eltern)), f.eltern...)
			}
			schreibe(t, wo+" (Empfänger vorher)", a, senke, k, slog.LevelInfo, "m")
			vorher := a.Snapshot()
			attrs := slices.Clone(f.attrs)
			kind := k.mitAttrs(k.h.WithAttrs(attrs), f.attrs...)
			for i := range attrs {
				attrs[i] = s("mode", "zz")
			}
			pruefeVerwuerfe(t, wo+" (WithAttrs)", vorher, a.Snapshot(), nil)
			for range 3 {
				schreibe(t, wo, a, senke, kind, slog.LevelInfo, "m", f.aufruf...)
			}
			schreibe(t, wo+" (Empfänger nachher)", a, senke, k, slog.LevelInfo, "m")
		}
	}
	if n != 4*len(faelle) {
		t.Fatalf("%d Läufe, erwartet %d", n, 4*len(faelle))
	}
}

// H05 (2): Geschwister mit Kapazität über der Länge beeinflussen sich nicht. Menge: Ketten aus 0 bis 9
// WithAttrs mit je einem Attribut (jede Kapazität, die das Wachstum der Liste ergibt) × 2 Geschwister
// mit verschiedenem neuem Attribut, im Wechsel geschrieben.
func TestWithAttrsGeschwister(t *testing.T) {
	folge, n := logSchluessel(t), 0
	for laenge := 0; laenge <= 9; laenge++ {
		senke := &FakeSenke{}
		a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
		k := kette{name: "Eltern", h: a.Handler()}
		for _, key := range folge[:laenge] {
			x := slogAttr(gueltigFuer(key))
			k = k.mitAttrs(k.h.WithAttrs([]slog.Attr{x}), x)
		}
		eins, zwei := slog.Int("befunde", 1), slog.Int("befunde", 2)
		k1, k2 := k.mitAttrs(k.h.WithAttrs([]slog.Attr{eins}), eins), k.mitAttrs(k.h.WithAttrs([]slog.Attr{zwei}), zwei)
		for _, g := range []kette{k1, k2, k1, k2, k} {
			n++
			schreibe(t, "Länge "+string(rune('0'+laenge)), a, senke, g, slog.LevelInfo, "m")
		}
	}
	if n != 50 {
		t.Fatalf("%d Einträge, erwartet 50", n)
	}
}

// H05 (3): ein LogValuer in WithAttrs wird einmal beim Ableiten aufgelöst, nicht bei jedem Eintrag; die
// Zählung je Eintrag folgt dem Ergebnis (hier ein ungültiger Wert: wert 3 nach drei Einträgen).
func TestWithAttrsLogValuerEinmal(t *testing.T) {
	n := new(atomic.Int64)
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	h := a.Handler().WithAttrs([]slog.Attr{slog.Any("mode", zaehlWert{n, slog.StringValue("a b")})})
	if n.Load() != 1 {
		t.Fatalf("%d Aufrufe von LogValue beim Ableiten, erwartet 1", n.Load())
	}
	for range 3 {
		_ = h.Handle(context.Background(), rec(slog.LevelInfo, "m"))
	}
	if n.Load() != 1 || a.Snapshot().Dropped[support.VerwurfWert] != 3 {
		t.Errorf("%d Aufrufe von LogValue, wert %d; erwartet 1 und 3", n.Load(), a.Snapshot().Dropped[support.VerwurfWert])
	}
}

// H06: WithGroup mit nichtleerem Namen setzt das Präfix name. vor jeden folgenden Schlüssel, auch vor
// die eines späteren WithAttrs; der Empfänger bleibt unverändert; ein leerer Name liefert denselben
// Handler. Menge M14 (4 Folgen von Namen × 2 Lagen von WithAttrs + der Fall des leeren Namens = 9) ×
// Menge M17 (4 Arten von Handlern).
func TestWithGroup(t *testing.T) {
	folgen := [][]string{{""}, {"g"}, {"a", "b"}, {"", "g"}}
	vor := slog.String("session_id", "v2")
	n := 0
	for art := range 4 {
		for _, folge := range folgen {
			for _, lage := range []string{"vor", "nach"} {
				n++
				senke := &FakeSenke{}
				a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
				k := ketten(a)[art]
				m := k
				if lage == "vor" {
					m = m.mitAttrs(m.h.WithAttrs([]slog.Attr{vor}), vor)
				}
				for _, name := range folge {
					m = m.mitGruppe(m.h.WithGroup(name), name)
				}
				if lage == "nach" {
					m = m.mitAttrs(m.h.WithAttrs([]slog.Attr{vor}), vor)
				}
				wo := k.name + ", " + lage + ", " + folge[len(folge)-1]
				schreibe(t, wo, a, senke, m, slog.LevelWarn, "m", s("service_id", "svc-9"))
				schreibe(t, wo+" (Empfänger)", a, senke, k, slog.LevelWarn, "m", s("service_id", "svc-9"))
			}
		}
		if k := ketten(neuMit(t, &FakeSenke{}, festeUhr(), slog.LevelInfo))[art]; k.h.WithGroup("") != k.h {
			t.Errorf("%s: WithGroup(\"\") liefert einen anderen Handler", k.name)
		}
		n++
	}
	if n != 4*(len(folgen)*2+1) || n/4 != 9 {
		t.Fatalf("%d Fälle, erwartet 36", n)
	}
}

// H06 (Punkt): der Präfix ist name. und nichts anderes. Menge: jeder Schlüssel K der Menge Log, in
// einen oder zwei Gruppennamen und den Rest zerlegt an jeder Stelle (jeder Name nichtleer, darunter
// Namen, die selbst Buchstaben, Ziffern und _ tragen, und Reste wie _id), ohne Zeichen und mit je einem
// ausgelassenen Zeichen an jeder Stelle (der Platz eines Trenners): ohne Punkt, mit einem anderen
// Trenner oder mit dem Punkt an anderer Stelle ergäbe die Verkettung den gültigen Schlüssel K.
// Orakel: das Modell mit Punkt, genau ein Verwurf schluessel, kein Feld. Die Zahl der Zerlegungen
// kommt aus der Länge der Schlüssel: zerlegungen(m) = (m-1) + (m-1)(m-2)/2 für eine Länge m, einmal
// für K und je einmal für K ohne das Zeichen d (len Stellen).
func TestWithGroupPunkt(t *testing.T) {
	zahl := func(m int) int { return (m - 1) + (m-1)*(m-2)/2 }
	n, want := 0, 0
	for _, key := range logSchluessel(t) {
		l := len(key)
		want += zahl(l) + l*zahl(l-1)
		x := slogAttr(gueltigFuer(key))
		zerlegungen := [][]string{}
		for d := -1; d < l; d++ {
			rest := key
			if d >= 0 {
				rest = key[:d] + key[d+1:]
			}
			for i := 1; i < len(rest); i++ {
				zerlegungen = append(zerlegungen, []string{rest[:i], rest[i:]})
				for j := i + 1; j < len(rest); j++ {
					zerlegungen = append(zerlegungen, []string{rest[:i], rest[i:j], rest[j:]})
				}
			}
		}
		for _, z := range zerlegungen {
			n++
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			m := ketten(a)[0]
			for _, name := range z[:len(z)-1] {
				m = m.mitGruppe(m.h.WithGroup(name), name)
			}
			y := x
			y.Key = z[len(z)-1]
			schreibe(t, key+" zerlegt "+strings.Join(z, "|"), a, senke, m, slog.LevelInfo, "m", y)
			if e := m.erwartet(y); e.schluessel != 1 || len(e.felder) != 0 {
				t.Fatalf("Modell: %d Verwürfe, %d Felder; erwartet 1 und 0", e.schluessel, len(e.felder))
			}
		}
	}
	if n != want {
		t.Fatalf("%d Zerlegungen, erwartet %d", n, want)
	}
}

// wiederEintritt ist ein LogValuer, dessen LogValue auf demselben Handler Handle ruft (Wiedereintritt
// zwischen dem Aufbau zweier Teile der Liste des äußeren Eintrags).
type wiederEintritt struct{ h slog.Handler }

func (w wiederEintritt) LogValue() slog.Value {
	_ = w.h.Handle(context.Background(), rec(slog.LevelInfo, "innen", s("exchange_id", "ex-2")))
	return slog.StringValue("x")
}

// H05 (Wiedereintritt): Handle baut die Liste je Aufruf neu und hängt nie an vorgelegt an. Menge:
// Ketten aus 0 bis 9 WithAttrs mit je einem Attribut (jede Kapazität, die das Wachstum ergibt); der
// äußere Eintrag trägt exchange_id=ex-1 und einen LogValuer, der denselben Handler mit exchange_id=ex-2
// aufruft. Orakel: das Modell; die innere Zeile (zuerst geschrieben) trägt ex-2, die äußere ex-1.
// Fängt den Anhang an eine geteilte Liste sequentiell, ohne Detektor.
func TestHandleWiedereintritt(t *testing.T) {
	var keys []string
	for _, k := range logSchluessel(t) {
		if k != "exchange_id" && k != "mode" {
			keys = append(keys, k)
		}
	}
	n := 0
	for laenge := 0; laenge <= 9; laenge++ {
		n++
		senke := &FakeSenke{}
		a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
		m := kette{h: a.Handler()}
		for _, key := range keys[:laenge] {
			x := slogAttr(gueltigFuer(key))
			m = m.mitAttrs(m.h.WithAttrs([]slog.Attr{x}), x)
		}
		aussen := []slog.Attr{s("exchange_id", "ex-1"), slog.Any("mode", wiederEintritt{m.h})}
		if err := m.h.Handle(context.Background(), rec(slog.LevelInfo, "aussen", aussen...)); err != nil {
			t.Errorf("Länge %d: Handle liefert %v", laenge, err)
		}
		z := senke.zeilen()
		if len(z) != 2 {
			t.Fatalf("Länge %d: %d Zeilen, erwartet 2", laenge, len(z))
		}
		pruefeZeile(t, "Länge "+strconv.Itoa(laenge)+" innen", z[0], "INFO", "innen", m.erwartet(s("exchange_id", "ex-2")))
		pruefeZeile(t, "Länge "+strconv.Itoa(laenge)+" außen", z[1], "INFO", "aussen", m.erwartet(s("exchange_id", "ex-1"), s("mode", "x")))
	}
	if n != 10 {
		t.Fatalf("%d Ketten, erwartet 10", n)
	}
}
