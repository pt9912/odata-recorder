package telemetry_test

import (
	"bytes"
	"context"
	"log/slog"
	"math"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// slogAttr bildet ein Attribut der Port-Art in das slog.Attr gleichen Werts ab.
func slogAttr(a support.Attr) slog.Attr {
	switch a.Art {
	case support.ArtInt:
		return slog.Int64(a.Key, a.Ganz)
	case support.ArtFloat:
		return slog.Float64(a.Key, a.Gleit)
	}
	return slog.String(a.Key, a.Text)
}

func slogAttrs(attrs []support.Attr) []slog.Attr {
	var aus []slog.Attr
	for _, a := range attrs {
		aus = append(aus, slogAttr(a))
	}
	return aus
}

// ctxe sind die drei ctx der Menge M8: nil, beendet und mit abgelaufener Frist.
func ctxe(t *testing.T) []context.Context {
	t.Helper()
	var kein context.Context
	beendet, abbruch := context.WithCancel(context.Background())
	abbruch()
	frist, ende := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	t.Cleanup(ende)
	return []context.Context{kein, beendet, frist}
}

// rec bildet einen Record mit einer Zeit, die nie die der Uhr ist.
func rec(level slog.Level, msg string, attrs ...slog.Attr) slog.Record {
	r := slog.NewRecord(time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC), level, msg, 0)
	r.AddAttrs(attrs...)
	return r
}

// modAttr ist ein Attribut im Modell: der volle Schlüssel (Präfix davor) und der aufgelöste Wert.
type modAttr struct {
	key string
	v   slog.Value
}

// erwartung ist, was das Modell von SPEC-231 Nr. 5 und SPEC-230 Nr. 4 für eine Folge von Attributen
// verlangt: die Felder der Zeile mit ihrem Text und die Zahl der Verwürfe schluessel und wert.
type erwartung struct {
	felder           map[string]string
	schluessel, wert uint64
}

// modellAttr ist die Abbildung der Arten nach SPEC-231 Nr. 5 aus ihrer Beschreibung.
func modellAttr(key string, v slog.Value) (support.Attr, bool) {
	switch v.Kind() {
	case slog.KindString:
		return support.Str(key, v.String()), true
	case slog.KindInt64:
		return support.Int(key, v.Int64()), true
	case slog.KindUint64:
		return support.Int(key, int64(v.Uint64())), v.Uint64() <= math.MaxInt64
	case slog.KindFloat64:
		return support.Attr{Key: key, Art: support.ArtFloat, Gleit: v.Float64()}, true
	case slog.KindDuration:
		return support.Millis(key, v.Duration()), true
	}
	return support.Attr{}, false
}

func katalogZeile(key string) (zeile, bool) {
	for _, z := range katalog() {
		if z.menge == "log" && z.key == key {
			return z, true
		}
	}
	return zeile{}, false
}

// modell wendet SPEC-230 Nr. 4 an: zuerst der Schlüssel (außerhalb der Menge oder schon belegt:
// schluessel), dann die Art und die Form des Werts (wert); der erste angenommene Wert gilt.
func modell(attrs []modAttr) erwartung {
	e := erwartung{felder: map[string]string{}}
	for _, m := range attrs {
		z, ok := katalogZeile(m.key)
		if _, belegt := e.felder[m.key]; !ok || belegt {
			e.schluessel++
			continue
		}
		a, ok := modellAttr(m.key, m.v)
		if !ok || !orakel(z, a) {
			e.wert++
			continue
		}
		e.felder[m.key] = erwarteterText(a)
	}
	return e
}

// pruefeZeile prüft eine Zeile: Zeitstempel der Uhr, Stufe, Text und genau die Felder von e in der
// Folge des Katalogs mit ihrem Text.
func pruefeZeile(t *testing.T, wo string, zeile []byte, stufe, msg string, e erwartung) {
	t.Helper()
	g := lies(t, zeile)
	want := []string{"timestamp", "level", "msg"}
	for _, k := range orakelFolge() {
		if _, ok := e.felder[k]; ok {
			want = append(want, k)
		}
	}
	if !slices.Equal(g.schluessel, want) {
		t.Errorf("%s: Schlüssel %v, erwartet %v", wo, g.schluessel, want)
		return
	}
	if g.text["timestamp"] != festerStempel || g.text["level"] != stufe || g.text["msg"] != msg {
		t.Errorf("%s: Kopf %v, erwartet %s, %s, %q", wo, g.text, festerStempel, stufe, msg)
	}
	for k, v := range e.felder {
		if g.text[k] != v {
			t.Errorf("%s: %s = %q, erwartet %q", wo, k, g.text[k], v)
		}
	}
}

// pruefeVerwuerfe prüft, dass jeder Grund um genau want[grund] gewachsen ist.
func pruefeVerwuerfe(t *testing.T, wo string, vorher, nachher support.Snapshot, want map[support.Verwurf]uint64) {
	t.Helper()
	for g, n := range nachher.Dropped {
		if n != vorher.Dropped[g]+want[g] {
			t.Errorf("%s: Grund %s zählt %d, erwartet %d", wo, g, n, vorher.Dropped[g]+want[g])
		}
	}
}

func verwuerfeVon(e erwartung) map[support.Verwurf]uint64 {
	return map[support.Verwurf]uint64{support.VerwurfSchluessel: e.schluessel, support.VerwurfWert: e.wert}
}

// kette ist eine Art von Handler (Menge M17) samt ihrem Modell: Präfix und vorgelegte Attribute.
type kette struct {
	name    string
	h       slog.Handler
	praefix string
	vor     []modAttr
}

// mitAttrs und mitGruppe sind die Modelle von WithAttrs und WithGroup; h ist der abgeleitete Handler.
func (k kette) mitAttrs(h slog.Handler, attrs ...slog.Attr) kette {
	vor := slices.Clone(k.vor)
	for _, x := range attrs {
		vor = append(vor, modAttr{k.praefix + x.Key, x.Value})
	}
	k.h, k.vor = h, vor
	return k
}

func (k kette) mitGruppe(h slog.Handler, name string) kette {
	if name != "" {
		k.praefix += name + "."
	}
	k.h = h
	return k
}

func (k kette) erwartet(attrs ...slog.Attr) erwartung {
	m := slices.Clone(k.vor)
	for _, x := range attrs {
		m = append(m, modAttr{k.praefix + x.Key, x.Value})
	}
	return modell(m)
}

func vorAttrs() []slog.Attr {
	return []slog.Attr{slog.String("session_id", "v1"), slog.String("mode", "a b")}
}

// ketten sind die vier Arten von Handlern der Menge M17 über a: die Wurzel, WithAttrs, WithGroup und
// WithAttrs auf WithGroup.
func ketten(a *telemetry.Adapter) []kette {
	w := a.Handler()
	wurzel := kette{name: "Wurzel", h: w}
	mitA := wurzel.mitAttrs(w.WithAttrs(vorAttrs()), vorAttrs()...)
	mitA.name = "WithAttrs"
	g := w.WithGroup("g")
	mitG := wurzel.mitGruppe(g, "g")
	mitG.name = "WithGroup"
	gA := mitG.mitAttrs(g.WithAttrs(vorAttrs()), vorAttrs()...)
	gA.name = "WithAttrs auf WithGroup"
	return []kette{wurzel, mitA, mitG, gA}
}

// schreibe ruft Handle der Kette und prüft: Rückgabe nil, eine neue Zeile gleich dem Modell, die
// Verwürfe gleich dem Modell.
func schreibe(t *testing.T, wo string, a *telemetry.Adapter, senke *FakeSenke, k kette, level slog.Level, msg string, attrs ...slog.Attr) {
	t.Helper()
	vorher, n := a.Snapshot(), len(senke.zeilen())
	if err := k.h.Handle(context.Background(), rec(level, msg, attrs...)); err != nil {
		t.Errorf("%s: Handle liefert %v", wo, err)
	}
	z := senke.zeilen()
	if len(z) != n+1 {
		t.Errorf("%s: %d Zeilen, erwartet %d", wo, len(z), n+1)
		return
	}
	e := k.erwartet(attrs...)
	pruefeZeile(t, wo, z[n], modellStufe(int(level)), msg, e)
	pruefeVerwuerfe(t, wo, vorher, a.Snapshot(), verwuerfeVon(e))
}

// dienstFall ist ein Eintrag der Dienste (Menge M13) mit den Attributen in der Folge der Stelle.
type dienstFall struct {
	name  string
	level slog.Level
	msg   string
	attrs []slog.Attr
}

func s(k, v string) slog.Attr { return slog.String(k, v) }

// dienstFaelle sind die 14 Fälle der Menge M13: die zehn Ereignisse von SPEC-121 Nr. 12, SPEC-124
// Nr. 9 und SPEC-123 Nr. 11, dazu upstream je Grund (3 weitere) und capture je Richtung (1 weiterer);
// Text, Stufe und Attribute stehen so, wie die Spezifikation sie nennt.
func dienstFaelle() []dienstFall {
	record := func(ereignis, code string, extra ...slog.Attr) []slog.Attr {
		return append([]slog.Attr{
			s("mode", "record"), s("ereignis", ereignis), s("error_code", code), s("service_id", "svc-1"),
			s("session_id", "ses-1"), s("exchange_id", "ex-1"), s("correlation_id", "cor-1"),
		}, extra...)
	}
	const capture, upstream = "record: Aufzeichnung eines Bodys verworfen", "record: Upstream ohne Antwort"
	blob := "sha256:" + strings.Repeat("c", 64)
	return []dienstFall{
		{"capture request", slog.LevelError, capture, record("capture", "ODR-E2401", s("richtung", "request"), s("grund", "typ"))},
		{"capture response", slog.LevelError, capture, record("capture", "ODR-E2401", s("richtung", "response"), s("grund", "typ"))},
		{"upstream drain", slog.LevelError, upstream, record("upstream", "ODR-E4001", s("grund", "drain"))},
		{"upstream zeit", slog.LevelError, upstream, record("upstream", "ODR-E4001", s("grund", "zeit"))},
		{"upstream nicht_erreichbar", slog.LevelError, upstream, record("upstream", "ODR-E4001", s("grund", "nicht_erreichbar"))},
		{"upstream intern", slog.LevelError, upstream, record("upstream", "ODR-E9001", s("grund", "intern"))},
		{"ablage", slog.LevelError, "record: Exchange nicht angehängt", record("ablage", "ODR-E9001")},
		{"bereinigung_fehlgeschlagen", slog.LevelError, "session: Bereinigung nach dem Löschen fehlgeschlagen",
			[]slog.Attr{s("service_id", "svc-1"), s("session_id", "ses-1"), s("ereignis", "bereinigung_fehlgeschlagen"), s("error_code", "ODR-E3001")}},
		{"bereinigung_zurueckgestellt", slog.LevelInfo, "session: Bereinigung nach dem Löschen zurückgestellt",
			[]slog.Attr{s("service_id", "svc-1"), s("session_id", "ses-1"), s("ereignis", "bereinigung_zurueckgestellt")}},
		{"beenden_fehlgeschlagen", slog.LevelError, "session: Beenden beim Herunterfahren fehlgeschlagen",
			[]slog.Attr{s("service_id", "svc-1"), s("session_id", "ses-1"), s("ereignis", "beenden_fehlgeschlagen"), s("error_code", "ODR-E7003")}},
		{"integritaet", slog.LevelError, "wartung: Integritätsprüfung mit Befund",
			[]slog.Attr{s("ereignis", "integritaet"), s("error_code", "ODR-E3001"), slog.Int("befunde", 3)}},
		{"aufbewahrung_uebersprungen", slog.LevelWarn, "wartung: Session als Quelle einer Baseline nicht gelöscht",
			[]slog.Attr{s("ereignis", "aufbewahrung_uebersprungen"), s("error_code", "ODR-E7004"), s("service_id", "svc-1"), s("session_id", "ses-1")}},
		{"blob_fehlt", slog.LevelError, "wartung: referenzierte Blob-Datei fehlt",
			[]slog.Attr{s("ereignis", "blob_fehlt"), s("error_code", "ODR-E3001"), s("blob", blob)}},
		{"blob_beschaedigt", slog.LevelError, "wartung: referenzierte Blob-Datei beschädigt",
			[]slog.Attr{s("ereignis", "blob_beschaedigt"), s("error_code", "ODR-E3001"), s("blob", blob)}},
	}
}

// H01: Menge M13 (14 Fälle aus 10 Ereignissen) über slog.New(Handler()): Text, Stufe und alle
// Attribute der Stelle stehen in der Zeile, kein Verwurf. Der Eintrag capture steht bytegleich da wie
// das Beispiel in SPEC-231 Nr. 4. Die zehn Namen sind dem Wortlaut der drei Stellen entnommen und
// werden von Hand abgeglichen (die Dienste liegen dem Test nicht als Quelle vor).
func TestDienstEintraege(t *testing.T) {
	faelle := dienstFaelle()
	namen := map[string]bool{}
	for _, f := range faelle {
		namen[f.attrs[slices.IndexFunc(f.attrs, func(x slog.Attr) bool { return x.Key == "ereignis" })].Value.String()] = true
	}
	if len(faelle) != 14 || len(namen) != 10 {
		t.Fatalf("%d Fälle, %d Ereignisse; erwartet 14 und 10", len(faelle), len(namen))
	}
	for _, f := range faelle {
		senke := &FakeSenke{}
		a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
		slog.New(a.Handler()).LogAttrs(context.Background(), f.level, f.msg, f.attrs...)
		z := senke.zeilen()
		if len(z) != 1 {
			t.Errorf("%s: %d Zeilen, erwartet 1", f.name, len(z))
			continue
		}
		e := erwartung{felder: map[string]string{}}
		for _, x := range f.attrs {
			e.felder[x.Key] = x.Value.String()
		}
		pruefeZeile(t, f.name, z[0], modellStufe(int(f.level)), f.msg, e)
		for g, n := range a.Snapshot().Dropped {
			if n != 0 {
				t.Errorf("%s: Grund %s zählt %d, erwartet 0", f.name, g, n)
			}
		}
		if f.name == "capture request" {
			want := `{"timestamp":"` + festerStempel + `","level":"ERROR","msg":"record: Aufzeichnung eines Bodys verworfen",` +
				`"mode":"record","service_id":"svc-1","session_id":"ses-1","exchange_id":"ex-1","correlation_id":"cor-1",` +
				`"error_code":"ODR-E2401","ereignis":"capture","richtung":"request","grund":"typ"}` + "\n"
			if string(z[0]) != want {
				t.Errorf("Beispielzeile %q, erwartet %q", z[0], want)
			}
		}
	}
}

// mengeLaenge ist die Menge M2 (469 Texte nach Länge und Zeichen an der Grenze 256), nach demselben
// Muster gebildet wie TestMsgLaenge.
func mengeLaenge() []string {
	var menge []string
	for n := 0; n <= 300; n++ {
		menge = append(menge, strings.Repeat("a", n))
	}
	folgen := []string{"", "x", string(rune(0xE9)), string(rune(0x20AC)), string(rune(0x1F600))}
	for _, r := range []rune{'x', 0xE9, 0x20AC, 0x1F600} {
		for off := 250; off <= 257; off++ {
			for _, danach := range folgen {
				menge = append(menge, strings.Repeat("a", off)+string(r)+danach)
			}
		}
	}
	for off := 250; off <= 257; off++ {
		menge = append(menge, strings.Repeat("a", off)+string([]byte{0x80}))
	}
	return menge
}

// H04: Log und Handler ergeben für dieselben Aufrufe dieselbe Zeile, ein Write je Aufruf. Mengen M1
// (16384 Aufrufe: alle Teilmengen der 13 Felder in zwei Folgen), M2 (469), M3 (8909) und die neun
// Stufen von M4; Orakel: die Zeilen beider Wege bytegleich, die Zahl der Write gleich der Zahl der
// Aufrufe, jedes Argument endet auf den Zeilenvorschub und trägt genau einen.
func TestHandlerGleichLog(t *testing.T) {
	type aufruf struct {
		level slog.Level
		msg   string
		attrs []support.Attr
	}
	folge := logSchluessel(t)
	var aufrufe []aufruf
	for maske := range 1 << len(folge) {
		for _, umgekehrt := range []bool{false, true} {
			var attrs []support.Attr
			for i, k := range folge {
				if maske>>i&1 == 1 {
					attrs = append(attrs, gueltigFuer(k))
				}
			}
			if umgekehrt {
				slices.Reverse(attrs)
			}
			aufrufe = append(aufrufe, aufruf{slog.LevelInfo, "m", attrs})
		}
	}
	m1 := len(aufrufe)
	laenge, zeichen := mengeLaenge(), mengeZeichen(t)
	for _, msg := range slices.Concat(laenge, zeichen) {
		aufrufe = append(aufrufe, aufruf{slog.LevelInfo, msg, nil})
	}
	for _, l := range stufenGrenzen(t) {
		aufrufe = append(aufrufe, aufruf{slog.Level(l), "m", nil})
	}
	if m1 != 2<<len(folge) || len(laenge) != 469 || len(zeichen) != 8909 || len(aufrufe) != m1+469+8909+9 {
		t.Fatalf("%d, %d, %d Elemente, %d Aufrufe; erwartet 16384, 469, 8909 und 9 Stufen", m1, len(laenge), len(zeichen), len(aufrufe))
	}
	ueberLog, ueberHandler := &FakeSenke{}, &FakeSenke{}
	a, b := neuMit(t, ueberLog, festeUhr(), slog.Level(-8)), neuMit(t, ueberHandler, festeUhr(), slog.Level(-8))
	logger := slog.New(b.Handler())
	for _, c := range aufrufe {
		a.Log(context.Background(), c.level, c.msg, c.attrs...)
		logger.LogAttrs(context.Background(), c.level, c.msg, slogAttrs(c.attrs)...)
	}
	zl, zh := ueberLog.zeilen(), ueberHandler.zeilen()
	if len(zl) != len(aufrufe) || len(zh) != len(aufrufe) || ueberLog.anzahl() != len(aufrufe) || ueberHandler.anzahl() != len(aufrufe) {
		t.Fatalf("%d und %d Zeilen, %d und %d Write; erwartet je %d", len(zl), len(zh), ueberLog.anzahl(), ueberHandler.anzahl(), len(aufrufe))
	}
	for i := range zl {
		if !bytes.Equal(zl[i], zh[i]) {
			t.Fatalf("Aufruf %d (%q): Log %q, Handler %q", i, aufrufe[i].msg, zl[i], zh[i])
		}
		if !bytes.HasSuffix(zh[i], []byte("\n")) || bytes.Count(zh[i], []byte("\n")) != 1 {
			t.Fatalf("Aufruf %d: Write ohne genau einen Zeilenvorschub am Ende: %q", i, zh[i])
		}
	}
}

// H05 (M7): doppelter Schlüssel über den Handler. 13 Schlüssel × 14 Positionen des zweiten Werts in
// einer Liste aus den 13 Feldern, beide Werte gültig und verschieden × 3 Schnitte der Liste zwischen
// WithAttrs und Aufruf (alles im Aufruf, alles vorgelegt, vor dem zweiten Wert). Orakel: das Modell,
// der erste Wert gilt, schluessel zählt 1.
func TestHandlerDoppelterSchluessel(t *testing.T) {
	folge, n := logSchluessel(t), 0
	for _, k := range folge {
		for pos := 0; pos <= len(folge); pos++ {
			var attrs []slog.Attr
			for _, f := range folge {
				attrs = append(attrs, slogAttr(gueltigFuer(f)))
			}
			zweiter := slogAttr(zweiterWert(k))
			attrs = slices.Insert(attrs, pos, zweiter)
			for _, schnitt := range []int{0, pos, len(attrs)} {
				n++
				senke := &FakeSenke{}
				a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
				m := ketten(a)[0]
				if schnitt > 0 {
					m = m.mitAttrs(m.h.WithAttrs(slices.Clone(attrs[:schnitt])), attrs[:schnitt]...)
				}
				schreibe(t, k+", Position "+strconv.Itoa(pos)+", Schnitt "+strconv.Itoa(schnitt), a, senke, m, slog.LevelInfo, "m", attrs[schnitt:]...)
				if e := ketten(a)[0].erwartet(attrs...); e.schluessel != 1 || e.wert != 0 {
					t.Fatalf("Modell: %d Verstöße gegen den Schlüssel, %d gegen den Wert; erwartet 1 und 0", e.schluessel, e.wert)
				}
			}
		}
	}
	if n != 13*14*3 {
		t.Fatalf("%d Fälle, erwartet %d", n, 13*14*3)
	}
}

// H07: Handle wendet MinLevel selbst an, und der Zeitstempel kommt von der Clock (minLevelPaare und
// zeitDerUhr).
func TestHandlerMinLevelZeit(t *testing.T) {
	minLevelPaare(t)
	zeitDerUhr(t)
}

// minLevelPaare: Menge M4 (81 Paare Stufe × MinLevel) auf der Wurzel und auf einem WithAttrs-Handler
// mit ungültigem vorgelegtem Attribut, direkter Aufruf: eine Zeile genau dann, wenn Stufe >= Minimum,
// Now einmal oder nie, die Verwürfe zählen immer; Enabled liefert dasselbe Urteil.
func minLevelPaare(t *testing.T) {
	t.Helper()
	werte := stufenGrenzen(t)
	paare := 0
	for _, vorgelegt := range []bool{false, true} {
		for _, stufe := range werte {
			for _, mindest := range werte {
				paare++
				senke, uhr := &FakeSenke{}, festeUhr()
				a := neuMit(t, senke, uhr, slog.Level(mindest))
				h, wert := a.Handler(), uint64(1)
				if vorgelegt {
					h, wert = h.WithAttrs([]slog.Attr{s("mode", "a b")}), 2
				}
				if err := h.Handle(context.Background(), rec(slog.Level(stufe), "m", s("session_id", "s1"), s("mode", "a b"))); err != nil {
					t.Errorf("Handle liefert %v", err)
				}
				geschrieben := int64(0)
				if stufe >= mindest {
					geschrieben = 1
				}
				if got := h.Enabled(context.Background(), slog.Level(stufe)); got != (geschrieben == 1) {
					t.Errorf("Stufe %d, MinLevel %d: Enabled %t, erwartet %t", stufe, mindest, got, geschrieben == 1)
				}
				if int64(len(senke.zeilen())) != geschrieben || uhr.n.Load() != geschrieben {
					t.Errorf("Stufe %d, MinLevel %d, vorgelegt %t: %d Zeilen, %d Aufrufe von Now; erwartet je %d", stufe, mindest, vorgelegt, len(senke.zeilen()), uhr.n.Load(), geschrieben)
				}
				if w := a.Snapshot().Dropped[support.VerwurfWert]; w != wert {
					t.Errorf("Stufe %d, MinLevel %d, vorgelegt %t: wert %d, erwartet %d", stufe, mindest, vorgelegt, w, wert)
				}
			}
		}
	}
	if paare != 2*len(werte)*len(werte) {
		t.Fatalf("%d Paare, erwartet %d", paare, 2*len(werte)*len(werte))
	}
}

// zeitDerUhr: der Zeitstempel kommt von der Clock, nie von Record.Time: Menge M5 (8 Zustände der Uhr)
// × 2 Zeiten des Records (eine feste und der Nullwert); Now einmal je geschriebener Zeile.
func zeitDerUhr(t *testing.T) {
	t.Helper()
	faelle := zeitFaelle()
	n := 0
	for _, f := range faelle {
		for _, rt := range []time.Time{time.Date(1999, 1, 2, 3, 4, 5, 6, time.UTC), {}} {
			n++
			senke, uhr := &FakeSenke{}, &zaehlUhr{t: f.t}
			a := neuMit(t, senke, uhr, slog.LevelInfo)
			r := slog.NewRecord(rt, slog.LevelInfo, "m", 0)
			if err := a.Handler().Handle(context.Background(), r); err != nil {
				t.Errorf("Handle liefert %v", err)
			}
			z := senke.zeilen()
			if len(z) != 1 || uhr.n.Load() != 1 || lies(t, z[0]).text["timestamp"] != f.want {
				t.Errorf("Uhr %v, Record %v: %q, %d Aufrufe von Now; erwartet timestamp %s und 1", f.t, rt, z, uhr.n.Load(), f.want)
			}
		}
	}
	if n != 2*len(faelle) || len(faelle) != 8 {
		t.Fatalf("%d Fälle, erwartet 16", n)
	}
	senke, uhr := &FakeSenke{}, festeUhr()
	h := neuMit(t, senke, uhr, slog.LevelInfo).Handler()
	for _, l := range []slog.Level{slog.LevelInfo, slog.LevelDebug, slog.LevelError} {
		_ = h.Handle(context.Background(), rec(l, "m"))
	}
	if uhr.n.Load() != 2 || len(senke.zeilen()) != 2 {
		t.Errorf("%d Aufrufe von Now, %d Zeilen; erwartet je 2 (der Eintrag DEBUG fällt unter MinLevel)", uhr.n.Load(), len(senke.zeilen()))
	}
}

// H08: Handle liefert immer nil und gerät nie in Panik: Menge M9 (7 Formen der Senke) × Menge M8
// (3 ctx) × Menge M17 (4 Arten von Handlern) = 84 Fälle, je zwei Einträge. Orakel: nil, senke je
// Eintrag gemäß Modell, die zweite Zeile ganz in der Senke.
func TestHandleOhneFehler(t *testing.T) {
	formen, ctxs := formenWrite(), ctxe(t)
	faelle := 0
	for art := range len(ketten(neuMit(t, &FakeSenke{}, festeUhr(), slog.LevelInfo))) {
		for _, f := range formen {
			for ci, ctx := range ctxs {
				faelle++
				senke := &FakeSenke{antwort: f.f}
				a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
				k := ketten(a)[art]
				wo := k.name + ", " + f.name + ", ctx " + string(rune('0'+ci))
				for range 2 {
					var err error
					if !binnen(t, 2*time.Second, wo, func() { err = k.h.Handle(ctx, rec(slog.LevelWarn, "m", s("service_id", "svc-9"))) }) {
						t.Fatalf("%s: Handle kehrt nicht zurück", wo)
					}
					if err != nil {
						t.Errorf("%s: Handle liefert %v", wo, err)
					}
				}
				z := senke.zeilen()
				if len(z) != 2 || senke.anzahl() != 2 {
					t.Errorf("%s: %d Write, erwartet 2", wo, senke.anzahl())
					continue
				}
				pruefeZeile(t, wo, z[1], "WARN", "m", k.erwartet(s("service_id", "svc-9")))
				if got := a.Snapshot().Dropped[support.VerwurfSenke]; got != 2*f.senke {
					t.Errorf("%s: senke %d, erwartet %d", wo, got, 2*f.senke)
				}
			}
		}
	}
	if want := 4 * len(formen) * len(ctxs); faelle != want || want != 84 {
		t.Fatalf("%d Fälle, erwartet 84", faelle)
	}
}

// H12: Handle ignoriert Record.PC, und für den Nullwert von slog.Record steht eine Zeile da.
func TestHandlePCNull(t *testing.T) {
	pcs := make([]uintptr, 1)
	if runtime.Callers(1, pcs) != 1 || pcs[0] == 0 {
		t.Fatal("kein PC")
	}
	var zeilen [][]byte
	for _, pc := range []uintptr{pcs[0], 0} {
		senke := &FakeSenke{}
		r := slog.NewRecord(time.Now(), slog.LevelWarn, "m", pc)
		r.AddAttrs(s("session_id", "s1"))
		if err := neuMit(t, senke, festeUhr(), slog.LevelInfo).Handler().Handle(context.Background(), r); err != nil {
			t.Errorf("Handle liefert %v", err)
		}
		zeilen = append(zeilen, senke.bytesAlle())
	}
	if len(zeilen[0]) == 0 || !bytes.Equal(zeilen[0], zeilen[1]) {
		t.Errorf("PC gesetzt %q, PC 0 %q; erwartet gleiche, nichtleere Zeilen", zeilen[0], zeilen[1])
	}
	senke := &FakeSenke{}
	var kein context.Context
	if err := neuMit(t, senke, festeUhr(), slog.LevelInfo).Handler().Handle(kein, slog.Record{}); err != nil {
		t.Errorf("Handle(Record{}) liefert %v", err)
	}
	z := senke.zeilen()
	if len(z) != 1 {
		t.Fatalf("%d Zeilen für slog.Record{}, erwartet 1", len(z))
	}
	pruefeZeile(t, "slog.Record{}", z[0], "INFO", "", erwartung{felder: map[string]string{}})
}

// H14: Enabled folgt MinLevel, liest ctx nicht, ruft Now nicht und zählt nichts: Menge M17 (4 Arten)
// × Menge M8 (3 ctx) × 9 Stufen von M4 bei MinLevel INFO = 108 Aufrufe.
func TestEnabled(t *testing.T) {
	ctxs, stufen := ctxe(t), stufenGrenzen(t)
	faelle := 0
	senke, uhr := &FakeSenke{}, festeUhr()
	a := neuMit(t, senke, uhr, slog.LevelInfo)
	arten := ketten(a)
	for _, k := range arten {
		for _, ctx := range ctxs {
			for _, l := range stufen {
				faelle++
				if got, want := k.h.Enabled(ctx, slog.Level(l)), l >= int(slog.LevelInfo); got != want {
					t.Errorf("%s, Stufe %d: Enabled %t, erwartet %t", k.name, l, got, want)
				}
			}
		}
	}
	vorher := a.Snapshot()
	if faelle != len(arten)*len(ctxs)*len(stufen) || faelle != 108 {
		t.Fatalf("%d Aufrufe, erwartet 108", faelle)
	}
	if uhr.n.Load() != 0 || len(senke.zeilen()) != 0 {
		t.Errorf("%d Aufrufe von Now, %d Zeilen; erwartet 0", uhr.n.Load(), len(senke.zeilen()))
	}
	for g, n := range vorher.Dropped {
		if n != 0 {
			t.Errorf("Grund %s zählt %d, erwartet 0", g, n)
		}
	}
}

// H15: jede Operation von slog.Handler wirkt auf jeder Art von Handler wie das Modell der Kette:
// Menge M17 (4 Arten × 4 Operationen = 16 Aufrufstellen). Enabled wie H14; Handle schreibt die Zeile
// aus Präfix und vorgelegten Attributen der Kette; WithAttrs und WithGroup liefern einen Handler, der
// in der Kette weiterwirkt, und lassen den Empfänger unverändert.
func TestAufrufstellen(t *testing.T) {
	ops := []string{"Enabled", "Handle", "WithAttrs", "WithGroup"}
	if want := reflect.TypeFor[slog.Handler]().NumMethod(); len(ops) != want {
		t.Fatalf("%d Operationen im Test, slog.Handler hat %d", len(ops), want)
	}
	aufruf := []slog.Attr{s("service_id", "svc-9"), s("fremd", "x")}
	neu := []slog.Attr{s("workspace_id", "ws-1"), s("mode", "a b")}
	stellen := 0
	for art := range 4 {
		for _, op := range ops {
			stellen++
			senke, uhr := &FakeSenke{}, festeUhr()
			a := neuMit(t, senke, uhr, slog.LevelInfo)
			k := ketten(a)[art]
			wo := k.name + "/" + op
			kind := k
			switch op {
			case "Enabled":
				if !k.h.Enabled(context.Background(), slog.LevelInfo) || k.h.Enabled(context.Background(), slog.LevelDebug) {
					t.Errorf("%s: Enabled folgt MinLevel nicht", wo)
				}
			case "WithAttrs":
				kind = k.mitAttrs(k.h.WithAttrs(neu), neu...)
			case "WithGroup":
				kind = k.mitGruppe(k.h.WithGroup("n"), "n")
			}
			schreibe(t, wo, a, senke, kind, slog.LevelWarn, "m", aufruf...)
			schreibe(t, wo+" (Empfänger)", a, senke, k, slog.LevelWarn, "m", aufruf...)
			if uhr.n.Load() != 2 {
				t.Errorf("%s: %d Aufrufe von Now, erwartet 2", wo, uhr.n.Load())
			}
		}
	}
	if stellen != 4*len(ops) || stellen != 16 {
		t.Fatalf("%d Aufrufstellen, erwartet 16", stellen)
	}
}

// H16: über slog.New(Handler()) erreicht ein Eintrag unter MinLevel Handle nicht und zählt nicht; der
// Eintrag auf MinLevel zählt (ein Logger ohne und mit With).
func TestLoggerUnterMinLevel(t *testing.T) {
	for _, mitWith := range []bool{false, true} {
		senke, uhr := &FakeSenke{}, festeUhr()
		a := neuMit(t, senke, uhr, slog.LevelInfo)
		l, wert := slog.New(a.Handler()), uint64(1)
		if mitWith {
			l, wert = l.With("mode", "a b"), 2
		}
		l.Debug("m", "mode", "a b")
		if len(senke.zeilen()) != 0 || uhr.n.Load() != 0 {
			t.Errorf("With %t: %d Zeilen, %d Aufrufe von Now unter MinLevel", mitWith, len(senke.zeilen()), uhr.n.Load())
		}
		for g, n := range a.Snapshot().Dropped {
			if n != 0 {
				t.Errorf("With %t: Grund %s zählt %d unter MinLevel, erwartet 0", mitWith, g, n)
			}
		}
		l.Info("m", "mode", "a b")
		if w := a.Snapshot().Dropped[support.VerwurfWert]; len(senke.zeilen()) != 1 || w != wert {
			t.Errorf("With %t: %d Zeilen, wert %d; erwartet 1 und %d", mitWith, len(senke.zeilen()), w, wert)
		}
	}
}
